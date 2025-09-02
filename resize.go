package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

// cmdResize handles the image resizing functionality
func cmdResize(args []string) {
	// Parse flags for resize command
	resizeFlags := flag.NewFlagSet("resize", flag.ExitOnError)
	runSilent := resizeFlags.Bool("silent", false, "Whether to produce any stdout output at all; errors will still be output; overrides other output flags")
	runVerbose := resizeFlags.Bool("verbose", false, "Verbose output, overrides -silent (silent) flag")
	targetWidth := resizeFlags.Int("width", 1024, "Target width in pixels")
	resizeFlags.Usage = func() {
		fmt.Println("Usage: cbztools resize [flags] <input_file> <output_file>")
		fmt.Println("Flags:")
		resizeFlags.PrintDefaults()
	}

	resizeFlags.Parse(args)

	// Parse the input and output files
	if resizeFlags.NArg() != 2 {
		resizeFlags.Usage()
		os.Exit(1)
	}
	inputFile := resizeFlags.Arg(0)
	outputFile := resizeFlags.Arg(1)

	printIfVerbose(fmt.Sprintf("Input file: %s", inputFile), runVerbose)
	printIfVerbose(fmt.Sprintf("Output file: %s", outputFile), runVerbose)
	printIfVerbose(fmt.Sprintf("Target width: %d", *targetWidth), runVerbose)

	// Read the input file (a zip archive with images, though with an cbz extension)
	inputCbz, err := zip.OpenReader(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %s", err)
		os.Exit(1)
	}
	defer inputCbz.Close()

	// Extract the files from the zip archive to a temporary directory
	tempDir, err := os.MkdirTemp("", "cbztools-resize")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temporary directory: %s", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	printIfVerbose("Extracting files to temporary directory...", runVerbose)
	for _, file := range inputCbz.File {
		// Only extract image files
		ext := strings.ToLower(filepath.Ext(file.Name))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
			rc, err := file.Open()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening file %s: %s", file.Name, err)
				os.Exit(1)
			}

			extractPath := filepath.Join(tempDir, file.Name)
			err = os.MkdirAll(filepath.Dir(extractPath), 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory for %s: %s", extractPath, err)
				rc.Close()
				os.Exit(1)
			}

			outFile, err := os.Create(extractPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file %s: %s", extractPath, err)
				rc.Close()
				os.Exit(1)
			}

			_, err = io.Copy(outFile, rc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error extracting file %s: %s", file.Name, err)
				rc.Close()
				outFile.Close()
				os.Exit(1)
			}

			rc.Close()
			outFile.Close()
		}
	}

	// For each image in the zip archive, resize it to the target width
	printIfVerbose("Resizing images...", runVerbose)
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			printIfVerbose(fmt.Sprintf("Processing: %s", filepath.Base(path)), runVerbose)

			// Open and decode the image
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening image %s: %w", path, err)
			}
			defer file.Close()

			var img image.Image
			if ext == ".jpg" || ext == ".jpeg" {
				img, err = jpeg.Decode(file)
			} else if ext == ".png" {
				img, err = png.Decode(file)
			}
			if err != nil {
				return fmt.Errorf("error decoding image %s: %w", path, err)
			}
			file.Close()

			// Get original dimensions
			bounds := img.Bounds()
			originalWidth := bounds.Dx()
			originalHeight := bounds.Dy()

			// Only resize if the image is wider than target width
			if originalWidth > *targetWidth {
				// Calculate new height maintaining aspect ratio
				newHeight := uint(originalHeight * (*targetWidth) / originalWidth)

				printIfVerbose(fmt.Sprintf("Resizing %s from %dx%d to %dx%d",
					filepath.Base(path), originalWidth, originalHeight, *targetWidth, newHeight), runVerbose)

				// Resize the image
				resizedImg := resize.Resize(uint(*targetWidth), newHeight, img, resize.Lanczos3)

				// Save the resized image
				outFile, err := os.Create(path)
				if err != nil {
					return fmt.Errorf("error creating resized image file %s: %w", path, err)
				}
				defer outFile.Close()

				if ext == ".jpg" || ext == ".jpeg" {
					err = jpeg.Encode(outFile, resizedImg, &jpeg.Options{Quality: 90})
				} else if ext == ".png" {
					err = png.Encode(outFile, resizedImg)
				}
				if err != nil {
					return fmt.Errorf("error encoding resized image %s: %w", path, err)
				}
				outFile.Close()
			} else {
				printIfVerbose(fmt.Sprintf("Skipping %s (already smaller than target width)",
					filepath.Base(path)), runVerbose)
			}
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing images: %s", err)
		os.Exit(1)
	}

	// Re-zip the files into the output file
	printIfVerbose("Creating output CBZ file...", runVerbose)
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %s", err)
		os.Exit(1)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	pageIndex := 1
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
			// Read the processed file
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening processed file %s: %w", path, err)
			}
			defer file.Close()

			// Create entry in zip with sequential naming
			filename := fmt.Sprintf("%05d%s", pageIndex, ext)
			zipEntry, err := zipWriter.Create(filename)
			if err != nil {
				return fmt.Errorf("error creating zip entry %s: %w", filename, err)
			}

			// Copy file content to zip
			_, err = io.Copy(zipEntry, file)
			if err != nil {
				return fmt.Errorf("error writing to zip entry %s: %w", filename, err)
			}
			file.Close()

			pageIndex++
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output zip: %s", err)
		os.Exit(1)
	}

	printIfNotSilent(fmt.Sprintf("Successfully resized CBZ file: %s -> %s with %d pages",
		inputFile, outputFile, pageIndex-1), runSilent, runVerbose)
}
