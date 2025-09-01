package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Splits a cbz file in half
func cmdSplit(args []string) {
	// Parse flags for split command
	splitFlags := flag.NewFlagSet("split", flag.ExitOnError)
	runSilent := splitFlags.Bool("silent", false, "Whether to produce any stdout output at all; errors will still be output")
	runVerbose := splitFlags.Bool("verbose", false, "Verbose output, overrides -silent flag")
	splitFlags.Usage = func() {
		fmt.Println("Usage: cbztools split [flags] <input.cbz> <output_dir>")
		fmt.Println("Flags:")
		splitFlags.PrintDefaults()
	}

	splitFlags.Parse(args)

	if splitFlags.NArg() != 2 {
		splitFlags.Usage()
		os.Exit(1)
	}
	cbzFile, outputDir := splitFlags.Arg(0), splitFlags.Arg(1)

	// Open input CBZ
	r, err := zip.OpenReader(cbzFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening CBZ file: %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

	// Collect image files
	var imageFiles []*zip.File
	for _, file := range r.File {
		ext := strings.ToLower(filepath.Ext(file.Name))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
			imageFiles = append(imageFiles, file)
		}
	}

	if len(imageFiles) < 2 {
		fmt.Fprintln(os.Stderr, "Need at least 2 images to split")
		os.Exit(1)
	}

	splitPoint := len(imageFiles) / 2
	baseName := strings.TrimSuffix(filepath.Base(cbzFile), filepath.Ext(cbzFile))

	// Create output files
	outputFile1 := filepath.Join(outputDir, fmt.Sprintf("%s_part1.cbz", baseName))
	outputFile2 := filepath.Join(outputDir, fmt.Sprintf("%s_part2.cbz", baseName))

	createCBZ(outputFile1, imageFiles[:splitPoint])
	createCBZ(outputFile2, imageFiles[splitPoint:])

	printIfNotSilent(fmt.Sprintf("Split %s into %s (%d pages) and %s (%d pages)",
		cbzFile, outputFile1, splitPoint, outputFile2, len(imageFiles)-splitPoint), runSilent, runVerbose)
}

// createCBZ creates a CBZ file with the given image files
func createCBZ(outputFile string, imageFiles []*zip.File) {
	out, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	zipWriter := zip.NewWriter(out)
	defer zipWriter.Close()

	pageIndex := 1
	for _, file := range imageFiles {
		rc, _ := file.Open()
		ext := strings.ToLower(filepath.Ext(file.Name))
		filename := fmt.Sprintf("%05d%s", pageIndex, ext)
		pageIndex++
		w, _ := zipWriter.Create(filename)
		io.Copy(w, rc)
		rc.Close()
	}
}
