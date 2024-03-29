/*
	Copyright (C) 2016 - 2024, Lefteris Zafiris <zaf@fastmail.com>

	This program is free software, distributed under the terms of
	the BSD 3-Clause License. See the LICENSE file
	at the top of the source tree.
*/

// The program takes as input a WAV or RAW PCM sound file
// and resamples it to the desired sampling rate.
// The output is RAW PCM data.
// Usage: goresample [flags] input_file output_file
//
// Example: go run main.go -ir 16000 -or 8000 ../../testing/piano-16k-16-2.wav 8k.raw

package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zaf/resample"
)

const wavHeader = 44

var (
	format = flag.String("format", "i16", "PCM format")
	ch     = flag.Int("ch", 2, "Number of channels")
	ir     = flag.Int("ir", 44100, "Input sample rate")
	or     = flag.Int("or", 0, "Output sample rate")
)

func main() {
	flag.Parse()
	var frmt int
	switch *format {
	case "i16":
		frmt = resample.I16
	case "i32":
		frmt = resample.I32
	case "f32":
		frmt = resample.F32
	case "f64":
		frmt = resample.F64
	default:
		log.Fatalln("Invalid Format")
	}
	if *ch < 1 {
		log.Fatalln("Invalid channel number")
	}
	if *ir <= 0 || *or <= 0 {
		log.Fatalln("Invalid input or output sample rate")
	}
	if flag.NArg() < 2 {
		log.Fatalln("No input or output files given")
	}
	inputFile := flag.Arg(0)
	outputFile := flag.Arg(1)
	var err error

	// Open input file (WAV or RAW PCM)
	input, err := os.Open(inputFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer input.Close()
	output, err := os.Create(outputFile)
	if err != nil {
		log.Fatalln(err)
	}
	// Create a Resampler
	res, err := resample.New(output, float64(*ir), float64(*or), *ch, frmt, resample.HighQ)
	if err != nil {
		output.Close()
		os.Remove(outputFile)
		log.Fatalln(err)
	}
	// Skip WAV file header in order to pass only the PCM data to the Resampler
	if strings.ToLower(filepath.Ext(inputFile)) == ".wav" {
		input.Seek(wavHeader, 0)
	}

	// Read input and pass it to the Resampler in chunks
	_, err = io.Copy(res, input)
	// Close the Resampler and the output file. Clsoing the Resampler will flush any remaining data to the output file.
	// If the Resampler is not closed before the output file, any remaining data will be lost.
	res.Close()
	output.Close()
	if err != nil {
		os.Remove(outputFile)
		log.Fatalln(err)
	}
}
