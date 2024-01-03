# resample
--
    import "github.com/zaf/resample"

Package resample implements resampling of PCM-encoded audio. It uses the SoX
Resampler library `libsoxr'.

To install make sure you have libsoxr and pkg-config installed, then run:

go install github.com/zaf/resample@latest

The package warps an io.Reader in a Resampler that resamples and writes all
input data. Input should be RAW PCM encoded audio samples.

For usage details please see the code snippet in the cmd folder.

## Usage

```go
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

)
```

#### type Resampler

```go
type Resampler struct {
}
```

Resampler resamples PCM sound data.

#### func  New

```go
func New(writer io.Writer, inputRate, outputRate float64, channels, format, quality int) (*Resampler, error)
```
New returns a pointer to a Resampler that implements an io.WriteCloser. It takes
as parameters the destination data Writer, the input and output sampling rates,
the number of channels of the input data, the input format and the quality setting.

#### func (*Resampler) Close

```go
func (r *Resampler) Close() (error)
```
Close flushes, clean-ups and frees memory. Should always be called when finished using
the resampler, and before we can use its output.

#### func (*Resampler) Reset

```go
func (r *Resampler) Reset(writer io.Writer) (error)
```
Reset permits reusing a Resampler rather than allocating a new one.

#### func (*Resampler) Write

```go
func (r *Resampler) Write(p []byte) (int, error)
```
Write resamples PCM sound data. Writes len(p) bytes from p to the underlying
data stream, returns the number of bytes written from p (0 <= n <= len(p)) and
any error encountered that caused the write to stop early.
