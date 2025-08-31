package main

import (
	"flag"
	"fmt"
	"os"
)

// cmdResize handles the image resizing functionality
func cmdResize(args []string) {
	// Parse flags for resize command
	resizeFlags := flag.NewFlagSet("resize", flag.ExitOnError)
	runSilent := resizeFlags.Bool("silent", false, "Whether to produce any stdout output at all; errors will still be output; overrides other output flags")
	runVerbose := resizeFlags.Bool("verbose", false, "Verbose output, overrides -silent (silent) flag")
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
	printIfNotSilent("Resize functionality not yet implemented", runSilent, runVerbose)

	panic("Not implemented yet")
}
