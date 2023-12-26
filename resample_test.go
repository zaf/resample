/*
	Copyright (C) 2016 - 2018, Lefteris Zafiris <zaf@fastmail.com>

	This program is free software, distributed under the terms of
	the BSD 3-Clause License. See the LICENSE file
	at the top of the source tree.
*/

package resample

import (
	"io"
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
	stream     bool
	err        string
}{
	{writer: io.Discard, inputRate: 8000.0, outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, stream: false, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, stream: true, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I32, quality: MediumQ, stream: false, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: F32, quality: MediumQ, stream: true, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: F64, quality: MediumQ, stream: false, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: Quick, stream: true, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: LowQ, stream: false, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: HighQ, stream: false, err: ""},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: VeryHighQ, stream: false, err: ""},
	{writer: nil, inputRate: 8000.0, outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, stream: false, err: "io.Writer is nil"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 0, format: I16, quality: MediumQ, stream: false, err: "Invalid channels number"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 0.0, channels: 0, format: I16, quality: MediumQ, stream: false, err: "Invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 0.0, outputRate: 8000.0, channels: 0, format: I16, quality: MediumQ, stream: false, err: "Invalid input or output sampling rates"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: 10, quality: MediumQ, stream: false, err: "Invalid format setting"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: 10, stream: false, err: "Invalid quality setting"},
	{writer: io.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: -10, stream: false, err: "Invalid quality setting"},
}

func TestNew(t *testing.T) {
	for _, tc := range NewTest {
		res, err := New(tc.writer, tc.inputRate, tc.outputRate, tc.channels, tc.format, tc.quality, tc.stream)
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

var WriterTest1 = []struct {
	data     []byte
	expected int
	err      string
}{
	{[]byte{}, 0, ""},
	{[]byte{0x01}, 0, "Incomplete input frame data"},
	{[]byte{0x01, 0x00}, 2, ""},
	{[]byte{0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde, 0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde}, 24, ""},
	{[]byte{0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde, 0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde, 0xd9}, 24, ""},
}

func TestWriter1(t *testing.T) {
	for _, stream := range []bool{true, false} {
		res, err := New(io.Discard, 8000.0, 8000.0, 1, I16, MediumQ, stream)
		if err != nil {
			t.Fatal("Failed to create a 1-1 Resampler: Stream:", stream, "Error:", err)
		}
		for _, tc := range WriterTest1 {
			i, err := res.Write(tc.data)
			res.Reset(io.Discard)
			if err != nil && err.Error() != tc.err {
				t.Errorf("Resampler 1-1 writer: stream: %t, error: %s, expecting: %s", stream, err.Error(), tc.err)
			}
			if err != nil {
				continue
			}
			if i != tc.expected {
				t.Errorf("Resampler 1-1 writer: tream: %t, returned: %d, expecting: %d", stream, i, tc.expected)
			}
		}
		res.Close()
	}
}

var WriterTest2 = []struct {
	data     []byte
	expected int
	err      string
}{
	{[]byte{}, 0, ""},
	{[]byte{0x01}, 0, "Incomplete input frame data"},
	{[]byte{0x01, 0x00, 0x7c, 0x7f}, 0, "Not enough input to generate output"},
	{[]byte{0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde, 0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde}, 24, ""},
	{[]byte{0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde, 0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde, 0xd9}, 24, ""},
}

func TestWriter2(t *testing.T) {
	for _, stream := range []bool{true, false} {
		res, err := New(io.Discard, 8000.0, 4000.0, 2, I16, MediumQ, stream)
		if err != nil {
			t.Fatal("Failed to create a 1-2 Resampler: Stream:", stream, "Error:", err)
		}
		for _, tc := range WriterTest2 {
			i, err := res.Write(tc.data)
			res.Reset(io.Discard)
			if err != nil && err.Error() != tc.err {
				t.Errorf("Resampler 1-2 writer: stream %t, error: %s, expecting: %s", stream, err.Error(), tc.err)
			}
			if err != nil {
				continue
			}
			if i != tc.expected {
				t.Errorf("Resampler 1-2 writer: stream: %t, returned: %d , expecting: %d", stream, i, tc.expected)
			}
		}
		res.Close()
	}
}

func TestClose(t *testing.T) {
	for _, stream := range []bool{true, false} {
		res, err := New(io.Discard, 16000.0, 8000.0, 1, I16, MediumQ, stream)
		if err != nil {
			t.Fatal("Failed to create a Resampler: stream:", stream, err)
		}
		err = res.Close()
		if err != nil {
			t.Fatal("Failed to Close the Resampler: stream:", stream, err)
		}
		_, err = res.Write(WriterTest1[3].data)
		if err == nil {
			t.Fatal("Running Write on a closed Resampler didn't return an error.")
		}
		err = res.Close()
		if err == nil {
			t.Fatal("Running Close on a closed Resampler didn't return an error.")
		}
	}
}

func TestReset(t *testing.T) {
	for _, stream := range []bool{true, false} {
		res, err := New(io.Discard, 16000.0, 8000.0, 1, I16, MediumQ, stream)
		if err != nil {
			t.Fatal("Failed to create a Resampler: stream:", stream, err)
		}
		err = res.Reset(io.Discard)
		if err != nil {
			t.Fatal("Failed to Reset the Resampler: stream:", stream, err)
		}
		err = res.Close()
		if err != nil {
			t.Fatal("Failed to Close the Resampler: stream:", stream, err)
		}
		err = res.Reset(io.Discard)
		if err == nil {
			t.Fatal("Running Reset on a closed Resampler didn't return an error.")
		}
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
	stream   bool
}{
	{"16bit 2 ch 44,1->16 Medium", "testing/piano-44.1k-16-2.wav", 44100.0, 16000.0, 2, I16, MediumQ, true},
	{"16bit 2 ch 16->8    Medium", "testing/piano-16k-16-2.wav", 16000.0, 8000.0, 2, I16, MediumQ, true},
	{"32fl  2 ch 44.1->8  Medium", "testing/piano-44.1k-32f-2.wav", 44100.0, 8000.0, 2, F32, MediumQ, true},
	{"16bit 2 ch 44.1->48 Medium", "testing/piano-44.1k-16-2.wav", 44100.0, 48000.0, 2, I16, MediumQ, true},
	{"16bit 2 ch 48->44.1 Medium", "testing/piano-48k-16-2.wav", 48000.0, 44100.0, 2, I16, MediumQ, true},
	{"16bit 1 ch 16->8     Quick", "testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, Quick, true},
	{"16bit 1 ch 16->8       Low", "testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, LowQ, true},
	{"16bit 1 ch 16->8    Medium", "testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, MediumQ, true},
	{"16bit 1 ch 16->8      High", "testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, HighQ, true},
	{"16bit 1 ch 16->8  VeryHigh", "testing/piano-16k-16-1.wav", 16000.0, 8000.0, 1, I16, VeryHighQ, true},
}

func BenchmarkResampling(b *testing.B) {
	for _, bd := range BenchData {
		b.Run(bd.name, func(b *testing.B) {
			rawData, err := os.ReadFile(bd.file)
			if err != nil {
				b.Fatalf("Failed to read test data: %s\n", err)
			}
			b.SetBytes(int64(len(rawData[44:])))
			res, err := New(io.Discard, bd.inRate, bd.outRate, bd.channels, bd.format, bd.quality, bd.stream)
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
