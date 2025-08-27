package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func setupStdout(t *testing.T) (*os.File, *os.File, *os.File) {
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	return originalStdout, r, w
}

func unsetStdout(originalStdout *os.File, w *os.File, r *os.File) string {
	var buf bytes.Buffer

	w.Close()
	os.Stdout = originalStdout
	io.Copy(&buf, r)
	output := buf.String()

	return output
}

func TestNoSilentFlagPrint(t *testing.T) {

	originalStdout, r, w := setupStdout(t)

	testString := "test123"
	silentFlag := new(bool)
	*silentFlag = false
	verboseFlag := new(bool)
	*verboseFlag = false
	printIfNotSilent(testString, silentFlag, verboseFlag)

	output := unsetStdout(originalStdout, w, r)

	if output != fmt.Sprintf("%s\n", testString) {
		t.Errorf("Expected: \"%s\", Got: \"%s\"", testString, output)
	}
}

func TestSilentFlagNoPrint(t *testing.T) {

	originalStdout, r, w := setupStdout(t)

	testString := "test123"
	silentFlag := new(bool)
	*silentFlag = true
	verboseFlag := new(bool)
	*verboseFlag = false
	printIfNotSilent(testString, silentFlag, verboseFlag)

	output := unsetStdout(originalStdout, w, r)

	if output != "" {
		t.Errorf("Expected: \"%s\", Got: \"%s\"", testString, output)
	}
}

func TestSilentFlagVerboseFlagPrint(t *testing.T) {

	originalStdout, r, w := setupStdout(t)

	testString := "test123"
	silentFlag := new(bool)
	*silentFlag = true
	verboseFlag := new(bool)
	*verboseFlag = true
	printIfNotSilent(testString, silentFlag, verboseFlag)

	output := unsetStdout(originalStdout, w, r)

	if output != fmt.Sprintf("%s\n", testString) {
		t.Errorf("Expected: \"%s\", Got: \"%s\"", testString, output)
	}
}
