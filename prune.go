package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

// cmdPrune handles the pruning functionality for removing duplicate CBZ files
func cmdPrune(args []string) {
	// Parse flags for prune command
	pruneFlags := flag.NewFlagSet("prune", flag.ExitOnError)
	runSilent := pruneFlags.Bool("silent", false, "Whether to produce any stdout output at all; errors will still be output; overrides other output flags")
	runVerbose := pruneFlags.Bool("verbose", false, "Verbose output, overrides -silent (silent) flag")
	// askBeforePrune := pruneFlags.Bool("y", false, "Ask before pruning each file")
	pruneFlags.Usage = func() {
		fmt.Println("Usage: cbztools prune [flags] <input_dir>")
		fmt.Println("Flags:")
		pruneFlags.PrintDefaults()
	}

	pruneFlags.Parse(args)

	// Parse the input directory
	if pruneFlags.NArg() != 1 {
		pruneFlags.Usage()
		os.Exit(1)
	}
	inputDir := pruneFlags.Arg(0)

	printIfVerbose(fmt.Sprintf("Input directory: %s", inputDir), runVerbose)

	// Find CBZ files
	cbzFiles, err := findCBZFiles(inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding CBZ files: %v\n", err)
		os.Exit(1)
	}
	printIfNotSilent(fmt.Sprintf("Found %d CBZ files", len(cbzFiles)), runSilent, runVerbose)

	// Sort the files by chapter number, just to make the output more readable
	sort.Slice(cbzFiles, func(i, j int) bool {
		return compareChaptersLess(cbzFiles[i], cbzFiles[j])
	})
	printIfVerbose("Sorted files:", runVerbose)
	for _, file := range cbzFiles {
		printIfVerbose(file, runVerbose)
	}

	// Iterating over the files, create a map of the chapter numbers to the files
	// For every chapter, there's a list of files with the same chapter number
	chapterFilesMap := make(map[string][]string)
	for _, file := range cbzFiles {
		chapter := getChapter(file)
		chapterFilesMap[chapter] = append(chapterFilesMap[chapter], file)
	}

	// Sort chapters for consistent output
	var chapters []string
	for chapter := range chapterFilesMap {
		chapters = append(chapters, chapter)
	}
	sort.Slice(chapters, func(i, j int) bool {
		return stringNatCmpLess(chapters[i], chapters[j])
	})

	// Print the chapter files map
	printIfVerbose("Files by chapter:", runVerbose)
	for _, chapter := range chapters {
		files := chapterFilesMap[chapter]
		printIfVerbose(fmt.Sprintf("  Chapter %s:", chapter), runVerbose)
		for _, file := range files {
			printIfVerbose(fmt.Sprintf("    %s", file), runVerbose)
		}
	}

	// Check if there are any chapters with more than one file
	// If there are not, print a stderr message and exit
	hasMultipleFiles := false
	for _, chapter := range chapters {
		if len(chapterFilesMap[chapter]) > 1 {
			hasMultipleFiles = true
			break
		}
	}
	if !hasMultipleFiles {
		fmt.Fprintln(os.Stderr, "No chapters with more than one file found, nothing to prune")
		os.Exit(1)
	}

	panic("Not implemented yet")
}
