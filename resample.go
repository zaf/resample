/*
	Copyright (C) 2016 - 2024, Lefteris Zafiris <zaf@fastmail.com>

	This program is free software, distributed under the terms of
	the BSD 3-Clause License. See the LICENSE file
	at the top of the source tree.
*/

/*
Package resample implements resampling of PCM-encoded audio.
It uses the SoX Resampler library `libsoxr'.

To install make sure you have libsoxr and pkg-config installed, then run:

go install github.com/zaf/resample@latest

The package wraps an io.Writer in a Resampler: data written to the
Resampler is resampled and written to the wrapped Writer.
Input should be RAW PCM encoded audio samples.

For usage details please see the code snippet in the cmd folder.
*/
package resample

/*
// Link soxr using pkg-config.
#cgo pkg-config: soxr
#include <soxr.h>
*/
import "C"
import (
	"errors"
	"io"
	"math"
	"runtime"
	"unsafe"
)

const (
	// Quality settings
	Quick     = 0 // Quick cubic interpolation
	LowQ      = 1 // LowQ 16-bit with larger rolloff
	MediumQ   = 2 // MediumQ 16-bit with medium rolloff
	HighQ     = 4 // High quality
	VeryHighQ = 6 // Very high quality

	// Input formats
	F32 = 0 // 32-bit floating point PCM
	F64 = 1 // 64-bit floating point PCM
	I32 = 2 // 32-bit signed linear PCM
	I16 = 3 // 16-bit signed linear PCM

	byteLen = 8

	// maxBufSize caps the output buffer used when resampling and draining.
	maxBufSize = 256 << 20

	// maxChannels caps the number of channels. soxr allocates a resampler per
	// channel, so an unbounded value would allow trivially exhausting memory.
	maxChannels = 256

	// Sampling rate limits, in Hz.
	minRate = 1.0
	maxRate = 1.0e9

	// maxRatio bounds the ratio between the input and the output rate.
	// libsoxr does not cope with extreme ratios: it hangs when flushing
	// resamplers with ratios around 2^20 and crashes on even larger ones,
	// so reject them up front.
	maxRatio = 1 << 16

	// maxDrainPasses caps how many times flush() may receive a completely
	// full maximum-size output buffer before giving up, so that a
	// resampler that never drains cannot hang Close forever.
	maxDrainPasses = 64
)

// Resampler resamples PCM sound data.
type Resampler struct {
	resampler   C.soxr_t
	inRate      float64   // input sample rate
	outRate     float64   // output sample rate
	channels    int       // number of input channels
	frameSize   int       // frame size in bytes
	leftover    []byte    // bytes of an incomplete frame, kept for the next Write
	destination io.Writer // output data
}

var threads int

func init() {
	threads = runtime.NumCPU()
}

// New returns a pointer to a Resampler that implements an io.WriteCloser.
// It takes as parameters the destination data Writer, the input and output
// sampling rates, the number of channels of the input data, the input format
// and the quality setting.
func New(writer io.Writer, inputRate, outputRate float64, channels, format, quality int) (*Resampler, error) {
	var size int
	if writer == nil {
		return nil, errors.New("io.Writer is nil")
	}
	// NaN rates fail every comparison, hence the negated form.
	if !(inputRate >= minRate && inputRate <= maxRate) || !(outputRate >= minRate && outputRate <= maxRate) {
		return nil, errors.New("invalid input or output sampling rates")
	}
	if ratio := inputRate / outputRate; ratio > maxRatio || ratio < 1.0/maxRatio {
		return nil, errors.New("invalid input or output sampling rates")
	}
	if channels <= 0 || channels > maxChannels {
		return nil, errors.New("invalid channels number")
	}
	if quality < 0 || quality > 6 {
		return nil, errors.New("invalid quality setting")
	}
	switch format {
	case F64:
		size = 64 / byteLen
	case F32, I32:
		size = 32 / byteLen
	case I16:
		size = 16 / byteLen
	default:
		return nil, errors.New("invalid format setting")
	}
	var soxErr C.soxr_error_t
	// Setup soxr and create a stream resampler
	ioSpec := C.soxr_io_spec(C.soxr_datatype_t(format), C.soxr_datatype_t(format))
	qSpec := C.soxr_quality_spec(C.ulong(quality), 0)
	runtimeSpec := C.soxr_runtime_spec(C.uint(threads))
	soxr := C.soxr_create(C.double(inputRate), C.double(outputRate), C.uint(channels), &soxErr, &ioSpec, &qSpec, &runtimeSpec)
	if soxErr != nil {
		return nil, errors.New(C.GoString(soxErr))
	}

	r := &Resampler{
		resampler:   soxr,
		inRate:      inputRate,
		outRate:     outputRate,
		channels:    channels,
		frameSize:   size,
		destination: writer,
	}
	return r, nil
}

// Reset permits reusing a Resampler rather than allocating a new one.
func (r *Resampler) Reset(writer io.Writer) error {
	if r.resampler == nil {
		return errors.New("soxr resampler is nil")
	}
	if writer == nil {
		return errors.New("io.Writer is nil")
	}
	err := r.flush()
	r.leftover = nil
	r.destination = writer
	C.soxr_clear(r.resampler)
	return err
}

// Close flushes, cleans up and frees memory.Should always be called
// when finished using the resampler, and before we can use its output.
func (r *Resampler) Close() error {
	if r.resampler == nil {
		return errors.New("soxr resampler is nil")
	}
	err := r.flush()
	C.soxr_delete(r.resampler)
	r.resampler = nil
	return err
}

// Write resamples PCM sound data. Writes len(p) bytes from p to
// the underlying data stream, returns the number of bytes written
// from p (0 <= n <= len(p)) and any error encountered that caused
// the write to stop early.
//
// Input may be written in chunks of any size: bytes of an incomplete
// frame are retained and joined with the next call.
// When all input has been written, Close must be called to flush any remaining output.
func (r *Resampler) Write(p []byte) (int, error) {
	if r.resampler == nil {
		return 0, errors.New("soxr resampler is nil")
	}
	n := len(p)
	if n == 0 {
		return 0, nil
	}
	frameBytes := r.frameSize * r.channels
	// Bytes of p needed to complete the frame left over by the previous
	// call, or a whole frame when there is nothing left over.
	need := frameBytes - len(r.leftover)
	if n < need {
		// Not even a whole frame; keep the bytes for the next call.
		r.leftover = append(r.leftover, p...)
		return 0, errors.New("partial frame buffered, send more data or Close to discard")
	}
	// Complete the leftover frame with the head of p and feed it to the
	// resampler on its own, so that the rest of p is processed in place,
	// without copying it.
	if len(r.leftover) > 0 {
		head := append(r.leftover, p[:need]...)
		p = p[need:]
		if err := r.processFrames(head); err != nil {
			return 0, err
		}
	}
	if err := r.processFrames(p); err != nil {
		return 0, err
	}
	return n, nil
}

// processFrames resamples the whole frames of p and writes the output to
// the destination. A trailing partial frame is kept for the next call.
func (r *Resampler) processFrames(p []byte) error {
	frameBytes := r.frameSize * r.channels
	framesIn := len(p) / frameBytes
	r.leftover = append([]byte(nil), p[framesIn*frameBytes:]...)
	if framesIn == 0 {
		return nil
	}
	// Output frames that fit in a buffer of maxBufSize bytes, and the input
	// frames that generate them. Input is fed to the resampler in slices of at
	// most maxIn frames, so that no single call needs an oversized output
	// buffer, however big the write is.
	ratio := r.outRate / r.inRate
	maxOut := maxBufSize / frameBytes
	maxIn := int(math.Min(float64(framesIn), float64(maxOut)/ratio))
	if maxIn < 1 {
		maxIn = 1
	}
	var dataOut []byte
	for off := 0; off < framesIn; off += maxIn {
		end := off + maxIn
		if end > framesIn {
			end = framesIn
		}
		// Ask for the expected number of output frames of this slice, but at
		// least one, so that the resampler is fed and pending output drained
		// even when downsampling small writes.
		olen := int(math.Min(float64(end-off)*ratio, float64(maxOut)))
		if olen < 1 {
			olen = 1
		}
		if len(dataOut) < olen*frameBytes {
			dataOut = make([]byte, olen*frameBytes)
		}
		var done C.size_t
		soxErr := C.soxr_process(r.resampler,
			C.soxr_in_t(unsafe.Pointer(&p[off*frameBytes])), C.size_t(end-off), nil,
			C.soxr_out_t(unsafe.Pointer(&dataOut[0])), C.size_t(olen), &done)
		if soxErr != nil {
			return errors.New(C.GoString(soxErr))
		}
		// In many cases the resampler will not return the full data unless we flush it,
		// especially if the input chunk is small. As long as we close the resampler
		// (Close() flushes all data) we don't need to worry about short writes, unless r.destination.Write() fails
		if _, err := r.destination.Write(dataOut[:int(done)*frameBytes]); err != nil {
			return err
		}
	}
	return nil
}

// flush any pending output from the resampler. After that no more input can be passed.
func (r *Resampler) flush() error {
	frameBytes := r.frameSize * r.channels
	maxFrames := maxBufSize / frameBytes
	// Start with a small buffer and grow it only while there is more output
	// pending, so that flush works for any channel count without ever
	// needing an oversized buffer.
	frames := 4096
	if frames > maxFrames {
		frames = maxFrames
	}
	full := 0
	for {
		dataOut := make([]byte, frames*frameBytes)
		var done C.size_t
		// Flush any pending output by calling soxr_process with no input data.
		soxErr := C.soxr_process(r.resampler, nil, 0, nil, C.soxr_out_t(unsafe.Pointer(&dataOut[0])), C.size_t(frames), &done)
		if soxErr != nil {
			return errors.New(C.GoString(soxErr))
		}
		_, err := r.destination.Write(dataOut[:int(done)*frameBytes])
		if err != nil {
			return err
		}
		// Fewer frames than requested means the resampler is drained.
		if done < C.size_t(frames) {
			return nil
		}
		if frames < maxFrames {
			frames *= 4
			if frames > maxFrames {
				frames = maxFrames
			}
			continue
		}
		// The buffer is already as large as it can get and it came back
		// full again. A healthy resampler drains eventually; give up
		// after many full passes instead of spinning forever.
		full++
		if full > maxDrainPasses {
			return errors.New("resampler failed to drain")
		}
	}
}
