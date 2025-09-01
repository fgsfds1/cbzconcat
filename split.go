package main

import (
	"fmt"
	"os"
)

// Splits a cbz file in half (for now)
func cmdSplit(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: cbztools split <file.cbz>")
		os.Exit(1)
	}

	cbzFile := args[0]

}
