/*
	Copyright (C) 2016 - 2024, Lefteris Zafiris <zaf@fastmail.com>

	This program is free software, distributed under the terms of
	the BSD 3-Clause License. See the LICENSE file
	at the top of the source tree.
*/

package resample

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os"
	"testing"
)

var NewTest = []struct {
	writer     io.Writer
	inputRate  float64
	outputRate float64
	channels   int
	format     int
	quality    int
	err        string
}{
	{writer: io.Discard, inputRate: 8000.0, outputRate: 8000.0, channels: 1, format: I16, quality: MediumQ, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I32, quality: MediumQ, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: F32, quality: MediumQ, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: F64, quality: MediumQ, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: Quick, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: LowQ, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: HighQ, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: VeryHighQ, err: ""},
	{writer: nil, inputRate: 8000.0, outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, err: "io.Writer is nil"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 0, format: I16, quality: MediumQ, err: "invalid channels number"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: -1, format: I16, quality: MediumQ, err: "invalid channels number"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 256, format: I16, quality: MediumQ, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 257, format: I16, quality: MediumQ, err: "invalid channels number"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 1000, format: I16, quality: MediumQ, err: "invalid channels number"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 0.0, channels: 0, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 0.0, outputRate: 8000.0, channels: 0, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: math.NaN(), outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 8000.0, outputRate: math.NaN(), channels: 2, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: math.Inf(1), outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 8000.0, outputRate: math.Inf(-1), channels: 2, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 1e-300, outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 1e300, outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 1.0, outputRate: 1e300, channels: 2, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 1.0, outputRate: 67108864.0, channels: 2, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 67108864.0, outputRate: 1.0, channels: 2, format: I16, quality: MediumQ, err: "invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: 10, quality: MediumQ, err: "invalid format setting"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: 10, err: "invalid quality setting"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: -10, err: "invalid quality setting"},
}

func TestNew(t *testing.T) {
	for _, tc := range NewTest {
		res, err := New(tc.writer, tc.inputRate, tc.outputRate, tc.channels, tc.format, tc.quality)
		if err != nil && tc.err != err.Error() {
			t.Fatalf("Expecting: %s got: %v", tc.err, err)
		}
		if err == nil && tc.err != "" {
			t.Fatalf("No error for: %s", tc.err)
		}
		if res != nil {
			res.Close()
		}
	}
}

var WriteTest = []struct {
	name       string
	inputRate  float64
	outputRate float64
	channels   int
	testData   []struct {
		data     []byte
		expected int
		err      string
	}
}{
	{"1-1 Resampler mono", 8000.0, 8000.0, 1, []struct {
		data     []byte
		expected int
		err      string
	}{
		{[]byte{}, 0, ""},
		{[]byte{0x01}, 0, "partial frame buffered, send more data or Close to discard"},
		{[]byte{0x01, 0x00}, 2, ""},
		{[]byte{0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde, 0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde}, 24, ""},
		{[]byte{0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde, 0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde, 0xd9}, 25, ""},
	}},

	{"1-2 Resampler stereo", 8000.0, 16000.0, 2, []struct {
		data     []byte
		expected int
		err      string
	}{
		{[]byte{}, 0, ""},
		{[]byte{0x01}, 0, "partial frame buffered, send more data or Close to discard"},
	}},

	{"2-1 Resampler stereo", 8000.0, 4000.0, 2, []struct {
		data     []byte
		expected int
		err      string
	}{
		{[]byte{}, 0, ""},
		{[]byte{0x01}, 0, "partial frame buffered, send more data or Close to discard"},
		{[]byte{0x01, 0x00, 0x7c, 0x7f}, 4, ""},
	}},

	{"65536-1 Resampler mono", 1.0, 65536.0, 1, []struct {
		data     []byte
		expected int
		err      string
	}{
		{[]byte{}, 0, ""},
		{[]byte{0x01, 0x00, 0x7c, 0x7f}, 4, ""},
	}},
}

func TestWrite(t *testing.T) {
	for _, tc := range WriteTest {
		res, err := New(io.Discard, tc.inputRate, tc.outputRate, tc.channels, I16, MediumQ)
		if err != nil {
			t.Fatal("Failed to create a", tc.name, "Resampler:", err)
		}
		for _, td := range tc.testData {
			i, err := res.Write(td.data)
			res.Reset(io.Discard)
			if err != nil && err.Error() != td.err {
				t.Errorf("Resampler %s writer, error: %s, expecting: %s", tc.name, err.Error(), td.err)
			}
			if err == nil && td.err != "" {
				t.Errorf("Resampler %s writer, expecting: %s", tc.name, td.err)
			}
			if err != nil {
				continue
			}
			if i != td.expected {
				t.Errorf("Resampler %s writer, returned: %d, expecting: %d", tc.name, i, td.expected)
			}
		}
	}
}

var FileTest = []struct {
	file       string
	inputRate  float64
	outputRate float64
	channels   int
	format     int
	quality    int
}{
	{"testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, MediumQ},
	{"testing/piano-16k-16-2.wav", 16000.0, 4000.0, 2, I16, MediumQ},
	{"testing/piano-44.1k-16-2.wav", 44100.0, 22050.0, 2, I16, MediumQ},
	{"testing/piano-44.1k-32f-2.wav", 44100.0, 48000.0, 2, F32, MediumQ},
	{"testing/piano-48k-16-2.wav", 48000.0, 44100.0, 2, I16, MediumQ},
}

func TestFile(t *testing.T) {
	for _, td := range FileTest {
		input, err := os.ReadFile(td.file)
		if err != nil {
			t.Fatal("Failed to read test data:", err)
		}
		var out bytes.Buffer
		res, err := New(&out, td.inputRate, td.outputRate, td.channels, td.format, td.quality)
		if err != nil {
			t.Fatal("Failed to create a Resampler:", err)
		}
		_, err = res.Write(input[44:])
		if err != nil {
			t.Errorf("Write failed: %s", err)
		}
		err = res.Close()
		if err != nil {
			t.Fatal("Failed to close Resampler:", err)
		}
		inFrames := float64(len(input[44:]) / (res.frameSize * td.channels))
		outFrames := float64(out.Len()) / float64(res.frameSize*td.channels)
		// soxr rounds the total output to whole frames; allow a mismatch
		// of one frame, which also covers non-integer resampling ratios.
		if math.Abs(outFrames-inFrames*td.outputRate/td.inputRate) > 1 {
			t.Error("Resampled file size mismatch, in:", int(inFrames), "out:", int(outFrames))
		}
	}
}

func TestWriteChunked(t *testing.T) {
	// 2000 frames of 32bit float PCM, written in chunks that are not a
	// multiple of the frame size, must give the same output as a single
	// write of the same data. soxr, like any floating point pipeline,
	// can differ by a few LSBs depending on how the input is split into
	// soxr_process calls, so samples are compared with a tolerance.
	data := make([]byte, 2000*8)
	for i := 0; i < len(data); i += 8 {
		for ch := 0; ch < 2; ch++ {
			v := float32((i/8*2+ch)%97) / 97
			binary.LittleEndian.PutUint32(data[i+ch*4:i+ch*4+4], math.Float32bits(v))
		}
	}
	var single, chunked bytes.Buffer
	for _, tc := range []struct {
		out  io.Writer
		size int
	}{{&single, len(data)}, {&chunked, 1001}} {
		res, err := New(tc.out, 44100.0, 16000.0, 2, F32, MediumQ)
		if err != nil {
			t.Fatal("Failed to create a Resampler:", err)
		}
		for off := 0; off < len(data); off += tc.size {
			end := off + tc.size
			if end > len(data) {
				end = len(data)
			}
			if _, err := res.Write(data[off:end]); err != nil {
				t.Fatalf("Write failed: %s", err)
			}
		}
		if err := res.Close(); err != nil {
			t.Fatal("Failed to close Resampler:", err)
		}
	}
	if single.Len() != chunked.Len() {
		t.Fatalf("Chunked write mismatch, single: %d bytes, chunked: %d bytes",
			single.Len(), chunked.Len())
	}
	a, b := single.Bytes(), chunked.Bytes()
	for i := 0; i+4 <= len(a); i += 4 {
		x := math.Float32frombits(binary.LittleEndian.Uint32(a[i:]))
		y := math.Float32frombits(binary.LittleEndian.Uint32(b[i:]))
		if math.Abs(float64(x-y)) > 0.001 {
			t.Errorf("Output sample %d differs: %g vs %g", i/4, x, y)
		}
	}
}

func TestWritePartialFrameRetention(t *testing.T) {
	// A Write that ends mid-frame returns an error but must retain the
	// bytes: completing the frame with the next call has to give the
	// same output as writing everything in one call. The stream is long
	// enough that a dropped or duplicated byte would shift it entirely;
	// samples are compared with a tolerance because soxr, like any
	// floating point pipeline, can differ by a few LSBs depending on
	// how input is split into soxr_process calls.
	data := make([]byte, 200) // 100 frames, mono I16
	for i := range data {
		data[i] = byte(i * 7)
	}
	var whole, split bytes.Buffer
	res, err := New(&whole, 8000.0, 8000.0, 1, I16, Quick)
	if err != nil {
		t.Fatal("Failed to create a Resampler:", err)
	}
	if _, err := res.Write(data); err != nil {
		t.Fatal("Failed to write full frames:", err)
	}
	res.Close()
	res, err = New(&split, 8000.0, 8000.0, 1, I16, Quick)
	if err != nil {
		t.Fatal("Failed to create a Resampler:", err)
	}
	if n, err := res.Write(data[:1]); err == nil || n != 0 {
		t.Fatalf("Expecting an error for a partial frame, got n=%d err=%v", n, err)
	}
	if n, err := res.Write(data[1:2]); err != nil || n != 1 {
		t.Fatalf("Failed to complete the frame exactly: n=%d err=%v", n, err)
	}
	if n, err := res.Write(data[2:]); err != nil || n != len(data[2:]) {
		t.Fatalf("Failed to write the rest: n=%d err=%v", n, err)
	}
	res.Close()
	if whole.Len() != split.Len() {
		t.Fatalf("Partial frame not retained across calls: %d bytes vs %d bytes",
			whole.Len(), split.Len())
	}
	a, b := whole.Bytes(), split.Bytes()
	for i := 0; i+1 < len(a); i += 2 {
		x := int(int16(uint16(a[i]) | uint16(a[i+1])<<8))
		y := int(int16(uint16(b[i]) | uint16(b[i+1])<<8))
		if x-y > 8 || y-x > 8 {
			t.Errorf("Sample %d differs by %d: partial frame not retained", i/2, x-y)
		}
	}
}

func TestClose(t *testing.T) {
	res, err := New(io.Discard, 16000.0, 8000.0, 1, I16, MediumQ)
	if err != nil {
		t.Fatal("Failed to create a Resampler:", err)
	}
	err = res.Close()
	if err != nil {
		t.Fatal("Failed to Close the Resampler:", err)
	}
	// A channel count that needs a flush buffer bigger than the buffer cap
	// must still close cleanly.
	res, err = New(io.Discard, 48000.0, 44100.0, 256, F64, MediumQ)
	if err != nil {
		t.Fatal("Failed to create a Resampler:", err)
	}
	err = res.Close()
	if err != nil {
		t.Fatal("Failed to Close a 256 channel F64 Resampler:", err)
	}
	_, err = res.Write(WriteTest[0].testData[3].data)
	if err == nil {
		t.Fatal("Running Write on a closed Resampler didn't return an error.")
	}
	err = res.Close()
	if err == nil {
		t.Fatal("Running Close on a closed Resampler didn't return an error.")
	}
}

func TestHighRatioDrain(t *testing.T) {
	// Upsampling at a large ratio leaves a lot of output pending inside
	// the resampler. Close must drain all of it, over multiple growing
	// passes, without the drain safety cap cutting it short.
	var out bytes.Buffer
	res, err := New(&out, 1000.0, 6553600.0, 1, I16, MediumQ)
	if err != nil {
		t.Fatal("Failed to create a Resampler:", err)
	}
	data := make([]byte, 1000*2) // 1000 frames of silence
	if _, err := res.Write(data); err != nil {
		t.Fatal("Failed to write data:", err)
	}
	if err := res.Close(); err != nil {
		t.Fatal("Failed to close Resampler:", err)
	}
	// 1000 frames in at a 6553.6 ratio: expect 6553600 output frames.
	got := out.Len() / 2
	if got < 6553600-1 || got > 6553600+1 {
		t.Errorf("Drained %d frames, expecting 6553600", got)
	}
}

func TestReset(t *testing.T) {
	res, err := New(io.Discard, 16000.0, 8000.0, 1, I16, MediumQ)
	if err != nil {
		t.Fatal("Failed to create a Resampler:", err)
	}
	err = res.Reset(io.Discard)
	if err != nil {
		t.Fatal("Failed to Reset the Resampler:", err)
	}
	err = res.Reset(nil)
	if err == nil {
		t.Fatal("Running Reset with a nil writer didn't return an error.")
	}
	err = res.Close()
	if err != nil {
		t.Fatal("Failed to Close the Resampler:", err)
	}
	err = res.Reset(io.Discard)
	if err == nil {
		t.Fatal("Running Reset on a closed Resampler didn't return an error.")
	}
}

func TestResetFlushesToOldDestination(t *testing.T) {
	// Pending output must be flushed to the destination that was active
	// while the data was written, not to the new one.
	data := make([]byte, 8192) // 4096 frames of mono I16 silence
	var before, after bytes.Buffer
	res, err := New(&before, 16000.0, 8000.0, 1, I16, MediumQ)
	if err != nil {
		t.Fatal("Failed to create a Resampler:", err)
	}
	if _, err := res.Write(data); err != nil {
		t.Fatal("Failed to write data:", err)
	}
	flushed := before.Len()
	if flushed == 0 {
		t.Fatal("Write produced no output")
	}
	if err := res.Reset(&after); err != nil {
		t.Fatal("Failed to Reset the Resampler:", err)
	}
	if before.Len() <= flushed {
		t.Error("Reset did not flush pending output to the old destination")
	}
	res.Close()
	if after.Len() != 0 {
		t.Errorf("Reset leaked output to the new destination: %d bytes", after.Len())
	}
}

// Benchmarking data
var BenchData = []struct {
	name     string
	file     string
	inRate   float64
	outRate  float64
	channels int
	format   int
	quality  int
}{
	{"16bit 2 ch 44,1->16 Medium", "testing/piano-44.1k-16-2.wav", 44100.0, 16000.0, 2, I16, MediumQ},
	{"16bit 2 ch 16->8    Medium", "testing/piano-16k-16-2.wav", 16000.0, 8000.0, 2, I16, MediumQ},
	{"32fl  2 ch 44.1->8  Medium", "testing/piano-44.1k-32f-2.wav", 44100.0, 8000.0, 2, F32, MediumQ},
	{"16bit 2 ch 44.1->48 Medium", "testing/piano-44.1k-16-2.wav", 44100.0, 48000.0, 2, I16, MediumQ},
	{"16bit 2 ch 48->44.1 Medium", "testing/piano-48k-16-2.wav", 48000.0, 44100.0, 2, I16, MediumQ},
	{"16bit 1 ch 16->8     Quick", "testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, Quick},
	{"16bit 1 ch 16->8       Low", "testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, LowQ},
	{"16bit 1 ch 16->8    Medium", "testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, MediumQ},
	{"16bit 1 ch 16->8      High", "testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, HighQ},
	{"16bit 1 ch 16->8  VeryHigh", "testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, VeryHighQ},
}

func BenchmarkResampling(b *testing.B) {
	for _, bd := range BenchData {
		b.Run(bd.name, func(b *testing.B) {
			rawData, err := os.ReadFile(bd.file)
			if err != nil {
				b.Fatalf("Failed to read test data: %s\n", err)
			}
			b.SetBytes(int64(len(rawData[44:])))
			res, err := New(io.Discard, bd.inRate, bd.outRate, bd.channels, bd.format, bd.quality)
			if err != nil {
				b.Fatalf("Failed to create Writer: %s\n", err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err = res.Write(rawData[44:])
				if err != nil {
					b.Fatalf("Encoding failed: %s\n", err)
				}
			}
			res.Close()
		})
	}
}
