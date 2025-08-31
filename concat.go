package main

import (
	"archive/zip"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// cmdConcat handles the concatenation functionality
func cmdConcat(args []string) {
	// Parse flags for concat command
	concatFlags := flag.NewFlagSet("concat", flag.ExitOnError)
	showXML := concatFlags.Bool("xml", false, "Print resulting XML (in the resulting cbz archive)")
	printOrder := concatFlags.Bool("order", false, "Print the order of the input cbz files")
	runSilent := concatFlags.Bool("silent", false, "Whether to produce any stdout output at all; errors will still be output; overrides other output flags")
	runVerbose := concatFlags.Bool("verbose", false, "Verbose output, overrides -silent (silent) flag")
	concatFlags.Usage = func() {
		fmt.Println("Usage: cbztools concat [flags] <input_dir> <output_dir>")
		fmt.Println("Flags:")
		concatFlags.PrintDefaults()
	}

	concatFlags.Parse(args)

	// We should have only two args left - the input dir and the output name
	if concatFlags.NArg() != 2 {
		concatFlags.Usage()
		os.Exit(1)
	}
	inputDir, outputDir := concatFlags.Arg(0), concatFlags.Arg(1)

	// Find CBZ files
	cbzFiles, err := findCBZFiles(inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding CBZ files: %v\n", err)
		os.Exit(1)
	}

	if len(cbzFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No CBZ files found")
		os.Exit(1)
	}

	if len(cbzFiles) == 1 {
		fmt.Fprintln(os.Stderr, "Only one CBZ file found - no concatenation needed")
		os.Exit(1)
	}

	// Print the original order of the files, for debugging
	if *printOrder || *runVerbose {
		printIfVerbose("Original order:", runVerbose)
		for _, name := range cbzFiles {
			printIfVerbose(name, runVerbose)
		}
	}

	// Sort files using the helper functions
	sort.Slice(cbzFiles, func(i, j int) bool {
		return compareChaptersLess(cbzFiles[i], cbzFiles[j])
	})

	// Print the order of the files
	if *printOrder || *runVerbose {
		printIfNotSilent("The files will be concatenated in the following order:", runSilent, runVerbose)
		for _, name := range cbzFiles {
			printIfNotSilent(name, runSilent, runVerbose)
		}
	}

	// Get basic book info from the first file, and the last chapter number from the last file
	firstComicInfo, err := readXmlFromZip(cbzFiles[0])
	if err != nil {
		panic(err)
	}
	firstXMLBytes, err := xml.MarshalIndent(firstComicInfo, "", "  ")
	if err != nil {
		panic(err)
	}
	printIfVerbose("XML read from first chapter:", runVerbose)
	printIfVerbose(string(firstXMLBytes[:]), runVerbose)

	lastComicInfo, err := readXmlFromZip(cbzFiles[len(cbzFiles)-1])
	if err != nil {
		panic(err)
	}
	lastXMLBytes, err := xml.MarshalIndent(lastComicInfo, "", "  ")
	if err != nil {
		panic(err)
	}
	printIfVerbose("XML read from last chapter:", runVerbose)
	printIfVerbose(string(lastXMLBytes[:]), runVerbose)

	seriesName := firstComicInfo.Series
	firstChapter := getChapter(firstComicInfo.Title)
	lastChapter := getChapter(lastComicInfo.Title)
	title := fmt.Sprintf("%s Ch.%s-%s", seriesName, firstChapter, lastChapter)
	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s.cbz", sanitizeFilenameASCII(title)))

	// Create output CBZ
	out, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	outZipFile := zip.NewWriter(out)
	defer outZipFile.Close()

	// Starting with the first page, for each archive, read it, get all images inside (opened in the order they were added to the zip file (!))
	// and write them to the `outZipFile` one-by-one, with the filename `pageIndex`
	pageIndex := 1
	for _, cbz := range cbzFiles {
		r, err := zip.OpenReader(cbz)
		if err != nil {
			panic(err)
		}
		for _, f := range r.File {
			// Copy only image files
			ext := strings.ToLower(filepath.Ext(f.Name))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
				rc, _ := f.Open()
				filename := fmt.Sprintf("%05d%s", pageIndex, ext)
				pageIndex++
				w, _ := outZipFile.Create(filename)
				io.Copy(w, rc)
				rc.Close()
			}
		}
		r.Close()
	}

	// Add ComicInfo.xml
	info := ComicInfo{
		Title:     title,
		Series:    seriesName,
		PageCount: pageIndex - 1,
	}
	xmlBytes, _ := xml.MarshalIndent(info, "", "  ")

	if *showXML || *runVerbose {
		printIfNotSilent(fmt.Sprintf("Resulting XML written to %s:", outputFile), runSilent, runVerbose)
		printIfNotSilent(string(xmlBytes[:]), runSilent, runVerbose)
	}

	w, _ := outZipFile.Create("ComicInfo.xml")
	w.Write([]byte(xml.Header))
	w.Write(xmlBytes)

	printIfNotSilent(fmt.Sprintf("Merged %d files into %s with %d pages\n", len(cbzFiles), outputFile, pageIndex-1), runSilent, runVerbose)
}
