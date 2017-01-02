/*
	Copyright (C) 2016 - 2017, Lefteris Zafiris <zaf@fastmail.com>

	This program is free software, distributed under the terms of
	the BSD 3-Clause License. See the LICENSE file
	at the top of the source tree.
*/

package resample

import (
	"io"
	"io/ioutil"
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
	{writer: ioutil.Discard, inputRate: 8000.0, outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, err: ""},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, err: ""},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I32, quality: MediumQ, err: ""},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: F32, quality: MediumQ, err: ""},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: F64, quality: MediumQ, err: ""},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: Quick, err: ""},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: LowQ, err: ""},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: HighQ, err: ""},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: VeryHighQ, err: ""},
	{writer: nil, inputRate: 8000.0, outputRate: 8000.0, channels: 2, format: I16, quality: MediumQ, err: "io.Writer is nil"},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 0, format: I16, quality: MediumQ, err: "Invalid channels number"},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 0.0, channels: 0, format: I16, quality: MediumQ, err: "Invalid input or output sampling rates"},
	{writer: ioutil.Discard, inputRate: 0.0, outputRate: 8000.0, channels: 0, format: I16, quality: MediumQ, err: "Invalid input or output sampling rates"},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: 10, quality: MediumQ, err: "Invalid format setting"},
	{writer: ioutil.Discard, inputRate: 16000.0, outputRate: 8000.0, channels: 2, format: I16, quality: 10, err: "Invalid quality setting"},
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

var WriterTest = []struct {
	data     []byte
	expected int
	err      string
}{
	{[]byte{}, 0, ""},
	{[]byte{0x01}, 0, "Incomplete input frame data"},
	{[]byte{0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde}, 12, ""},
	{[]byte{0x01, 0x00, 0x7c, 0x7f, 0xd1, 0xd0, 0xd3, 0xd2, 0xdd, 0xdc, 0xdf, 0xde, 0xd9}, 13, "Fragmented last frame in input data"},
}

func TestWrite(t *testing.T) {
	res, err := New(ioutil.Discard, 8000.0, 8000.0, 1, I16, MediumQ)
	if err != nil {
		t.Fatal("Failed to create a 1-1 Resampler: ", err)
	}
	for _, tc := range WriterTest {
		i, err := res.Write(tc.data)
		if err != nil && err.Error() != tc.err {
			t.Errorf("Resampler 1-1 writer error: %s , expecting: %s", tc.err, err.Error())
		}
		if i != tc.expected {
			t.Errorf("Resampler 1-1 writer returned: %d , actual: %d", tc.expected, i)
		}
	}
	res.Close()
}

func TestClose(t *testing.T) {
	res, err := New(ioutil.Discard, 16000.0, 8000.0, 1, I16, MediumQ)
	if err != nil {
		t.Fatal("Failed to create a Resampler: ", err)
	}
	err = res.Close()
	if err != nil {
		t.Fatal("Failed to Close the Resampler: ", err)
	}
	err = res.Close()
	if err == nil {
		t.Fatal("Running Close on a closed Resampler didn't return an error.")
	}
}

func TestReset(t *testing.T) {
	res, err := New(ioutil.Discard, 16000.0, 8000.0, 1, I16, MediumQ)
	if err != nil {
		t.Fatal("Failed to create a Resampler: ", err)
	}
	err = res.Reset(ioutil.Discard)
	if err != nil {
		t.Fatal("Failed to Reset the Resampler: ", err)
	}
	err = res.Close()
	if err != nil {
		t.Fatal("Failed to Close the Resampler: ", err)
	}
	err = res.Reset(ioutil.Discard)
	if err == nil {
		t.Fatal("Running Reset on a closed Resampler didn't return an error.")
	}
}

// Benchmark Downsampling 16b 44.1k->16k
func BenchmarkDownsample16b44k(b *testing.B) {
	rawData, err := ioutil.ReadFile("testing/piano-44.1k-16-2.wav")
	if err != nil {
		b.Fatalf("Failed to read test data: %s\n", err)
	}
	b.SetBytes(int64(len(rawData[44:])))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := New(ioutil.Discard, 44100.0, 16000.0, 2, I16, MediumQ)
		if err != nil {
			b.Fatalf("Failed to create Writer: %s\n", err)
		}
		_, err = res.Write(rawData[44:]) // Skip wav header
		res.Close()
		if err != nil {
			b.Fatalf("Encoding failed: %s\n", err)
		}
	}
}

// Benchmark Downsampling 16b 16k->8k
func BenchmarkDownsample16b16k(b *testing.B) {
	rawData, err := ioutil.ReadFile("testing/piano-16k-16-1.wav")
	if err != nil {
		b.Fatalf("Failed to read test data: %s\n", err)
	}
	b.SetBytes(int64(len(rawData[44:])))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := New(ioutil.Discard, 16000.0, 8000.0, 1, I16, MediumQ)
		if err != nil {
			b.Fatalf("Failed to create Writer: %s\n", err)
		}
		_, err = res.Write(rawData[44:])
		res.Close()
		if err != nil {
			b.Fatalf("Encoding failed: %s\n", err)
		}
	}
}

// Benchmark Downsampling 32b 44k->8k
func BenchmarkDownsample23b8k(b *testing.B) {
	rawData, err := ioutil.ReadFile("testing/organ44.1k-32f-2.wav")
	if err != nil {
		b.Fatalf("Failed to read test data: %s\n", err)
	}
	b.SetBytes(int64(len(rawData[44:])))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := New(ioutil.Discard, 44100.0, 8000.0, 1, F32, MediumQ)
		if err != nil {
			b.Fatalf("Failed to create Writer: %s\n", err)
		}
		_, err = res.Write(rawData[44:])
		res.Close()
		if err != nil {
			b.Fatalf("Encoding failed: %s\n", err)
		}
	}
}

// Benchmark Upsampling 44.1k->48k
func BenchmarkUpsample16b44k(b *testing.B) {
	rawData, err := ioutil.ReadFile("testing/piano-44.1k-16-2.wav")
	if err != nil {
		b.Fatalf("Failed to read test data: %s\n", err)
	}
	b.SetBytes(int64(len(rawData[44:])))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := New(ioutil.Discard, 44100.0, 48000.0, 2, I16, MediumQ)
		if err != nil {
			b.Fatalf("Failed to create Writer: %s\n", err)
		}
		_, err = res.Write(rawData[44:])
		res.Close()
		if err != nil {
			b.Fatalf("Encoding failed: %s\n", err)
		}
	}
}

// Benchmark Downsampling Quick
func BenchmarkQuick(b *testing.B) {
	rawData, err := ioutil.ReadFile("testing/piano-16k-16-1.wav")
	if err != nil {
		b.Fatalf("Failed to read test data: %s\n", err)
	}
	b.SetBytes(int64(len(rawData[44:])))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := New(ioutil.Discard, 16000.0, 8000.0, 1, I16, Quick)
		if err != nil {
			b.Fatalf("Failed to create Writer: %s\n", err)
		}
		_, err = res.Write(rawData[44:])
		res.Close()
		if err != nil {
			b.Fatalf("Encoding failed: %s\n", err)
		}
	}
}

// Benchmark Downsampling LowQ
func BenchmarkLowQ(b *testing.B) {
	rawData, err := ioutil.ReadFile("testing/piano-16k-16-1.wav")
	if err != nil {
		b.Fatalf("Failed to read test data: %s\n", err)
	}
	b.SetBytes(int64(len(rawData[44:])))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := New(ioutil.Discard, 16000.0, 8000.0, 1, I16, LowQ)
		if err != nil {
			b.Fatalf("Failed to create Writer: %s\n", err)
		}
		_, err = res.Write(rawData[44:])
		res.Close()
		if err != nil {
			b.Fatalf("Encoding failed: %s\n", err)
		}
	}
}

// Benchmark Downsampling MediumQ
func BenchmarkMediumQ(b *testing.B) {
	rawData, err := ioutil.ReadFile("testing/piano-16k-16-1.wav")
	if err != nil {
		b.Fatalf("Failed to read test data: %s\n", err)
	}
	b.SetBytes(int64(len(rawData[44:])))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := New(ioutil.Discard, 16000.0, 8000.0, 1, I16, MediumQ)
		if err != nil {
			b.Fatalf("Failed to create Writer: %s\n", err)
		}
		_, err = res.Write(rawData[44:])
		res.Close()
		if err != nil {
			b.Fatalf("Encoding failed: %s\n", err)
		}
	}
}

// Benchmark Downsampling HighQ
func BenchmarkHighQ(b *testing.B) {
	rawData, err := ioutil.ReadFile("testing/piano-16k-16-1.wav")
	if err != nil {
		b.Fatalf("Failed to read test data: %s\n", err)
	}
	b.SetBytes(int64(len(rawData[44:])))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := New(ioutil.Discard, 16000.0, 8000.0, 1, I16, HighQ)
		if err != nil {
			b.Fatalf("Failed to create Writer: %s\n", err)
		}
		_, err = res.Write(rawData[44:])
		res.Close()
		if err != nil {
			b.Fatalf("Encoding failed: %s\n", err)
		}
	}
}

// Benchmark Downsampling VeryHighQ
func BenchmarkVeryHighQ(b *testing.B) {
	rawData, err := ioutil.ReadFile("testing/piano-16k-16-1.wav")
	if err != nil {
		b.Fatalf("Failed to read test data: %s\n", err)
	}
	b.SetBytes(int64(len(rawData[44:])))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := New(ioutil.Discard, 16000.0, 8000.0, 1, I16, VeryHighQ)
		if err != nil {
			b.Fatalf("Failed to create Writer: %s\n", err)
		}
		_, err = res.Write(rawData[44:])
		res.Close()
		if err != nil {
			b.Fatalf("Encoding failed: %s\n", err)
		}
	}
}
