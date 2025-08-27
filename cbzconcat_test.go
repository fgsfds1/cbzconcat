package main

import (
	"bytes"
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

func TestPrintIfNotSilent(t *testing.T) {

	originalStdout, r, w := setupStdout(t)

	silentFlag := new(bool)
	*silentFlag = false
	verboseFlag := new(bool)
	*verboseFlag = false
	printIfNotSilent("1_should_print", silentFlag, verboseFlag)
	*silentFlag = true
	printIfNotSilent("2_shouldnt_print", silentFlag, verboseFlag)
	*verboseFlag = true
	printIfNotSilent("3_should_print", silentFlag, verboseFlag)
	*silentFlag = false
	printIfNotSilent("4_should_print", silentFlag, verboseFlag)
	output := unsetStdout(originalStdout, w, r)

	expectedOutput := "1_should_print\n3_should_print\n4_should_print\n"
	if output != expectedOutput {
		t.Errorf("Expected: \"%s\", Got: \"%s\"", expectedOutput, output)
	}
}

func TestPrintIfVerbose(t *testing.T) {
	originalStdout, r, w := setupStdout(t)

	verboseFlag := new(bool)
	*verboseFlag = false
	printIfVerbose("1_shouldnt_print", verboseFlag)
	*verboseFlag = true
	printIfVerbose("2_should_print", verboseFlag)
	output := unsetStdout(originalStdout, w, r)

	expectedOutput := "2_should_print\n"
	if output != expectedOutput {
		t.Errorf("Expected: \"%s\", Got: \"%s\"", expectedOutput, output)
	}
}
