package main

import (
	"flag"
	"fmt"
	"os"
)

// cmdMetadata handles the metadata editing functionality
func cmdMetadata(args []string) {
	// Parse flags for metadata command
	metadataFlags := flag.NewFlagSet("metadata", flag.ExitOnError)
	runSilent := metadataFlags.Bool("silent", false, "Whether to produce any stdout output at all; errors will still be output; overrides other output flags")
	runVerbose := metadataFlags.Bool("verbose", false, "Verbose output, overrides -silent (silent) flag")
	metadataFlags.Usage = func() {
		fmt.Println("Usage: cbztools metadata [flags] <input_file>")
		fmt.Println("Flags:")
		metadataFlags.PrintDefaults()
	}

	metadataFlags.Parse(args)

	// Parse the input file
	if metadataFlags.NArg() != 1 {
		metadataFlags.Usage()
		os.Exit(1)
	}
	inputFile := metadataFlags.Arg(0)

	printIfVerbose(fmt.Sprintf("Input file: %s", inputFile), runVerbose)
	printIfNotSilent("Metadata functionality not yet implemented", runSilent, runVerbose)

	panic("Not implemented yet")
}
