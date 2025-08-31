package main

import (
	"fmt"
	"os"
)

// Version information - these will be set at build time via ldflags
var (
	Version   = "unknown"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// cmdHelp displays help information
func cmdHelp(args []string) {
	fmt.Printf("cbztools v%s (%s)\n", Version, GitCommit)
	fmt.Println("A utility for working with CBZ comic archives.")
	fmt.Println()
	fmt.Println("Usage: cbztools <command> [flags] [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  concat    Concatenate multiple CBZ files into a single archive")
	fmt.Println("  prune     Intelligently prune duplicate CBZ files, mostly useful for removing scans of the same chapter by different groups (not implemented yet)")
	fmt.Println("  resize    Resize all images in a CBZ file to a given size (not implemented yet)")
	fmt.Println("  metadata  Edit the metadata of a CBZ file (not implemented yet)")
	fmt.Println("  version   Show the version of the program and exit")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("For help on a specific command:")
	fmt.Println("  cbztools <command> -h")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cbztools concat ./chapters ./output")
	fmt.Println("  cbztools concat -v -r ./chapters ./output")
}

// cmdVersion shows the version of the program and exit
func cmdVersion(args []string) {
	fmt.Printf("cbztools %s\n", Version)
	fmt.Printf("Build time: %s\n", BuildTime)
	fmt.Printf("Git commit: %s\n", GitCommit)
	os.Exit(0)
}

func main() {
	// Check if we have any arguments
	if len(os.Args) < 2 {
		// No subcommand provided, show help
		cmdHelp(nil)
		os.Exit(1)
	}

	args := os.Args[1:]

	// Get subcommand
	subcommand := args[0]
	subcommandArgs := args[1:]

	// Handle subcommands
	switch subcommand {
	case "concat":
		cmdConcat(subcommandArgs)
	case "prune":
		cmdPrune(subcommandArgs)
	case "resize":
		cmdResize(subcommandArgs)
	case "metadata":
		cmdMetadata(subcommandArgs)
	case "help", "h", "-h", "--help":
		cmdHelp(subcommandArgs)
	case "version", "v", "-v", "--version":
		cmdVersion(subcommandArgs)
	default:
		fmt.Printf("Unknown command: %s\n", subcommand)
		fmt.Println("Run 'cbztools help' for usage information.")
		os.Exit(1)
	}
}
