/*
	Copyright (C) 2016 - 2017, Lefteris Zafiris <zaf@fastmail.com>

	This program is free software, distributed under the terms of
	the BSD 3-Clause License. See the LICENSE file
	at the top of the source tree.
*/

package resample

import (
	"io/ioutil"
	"testing"
)

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
