package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// Helper functions to capture stdout, used in tests that test over stdout
// (printIfVerbose, printIfNotSilent)
func setupStdout(t *testing.T) (*os.File, *os.File, *os.File) {
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	return originalStdout, r, w
}

func getStdoutAndClose(originalStdout *os.File, w *os.File, r *os.File) string {
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
	output := getStdoutAndClose(originalStdout, w, r)

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
	output := getStdoutAndClose(originalStdout, w, r)

	expectedOutput := "2_should_print\n"
	if output != expectedOutput {
		t.Errorf("Expected: \"%s\", Got: \"%s\"", expectedOutput, output)
	}
}

func TestGetChapter(t *testing.T) {
	title, expectedChapter, resultChapter := "", "", ""

	title, expectedChapter = "", ""
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "Ch.0001", "0001"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "Ch.0001.5", "0001.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "Ch 0001.5", "0001.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "Ch  0001.5", "0001.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "Ch 0001.5.5.5", "0001.5.5.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "ch 0001.5.5.5", "0001.5.5.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "ch. 0001.5.5.5", "0001.5.5.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "chapter 0001.5.5.5", "0001.5.5.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "chapter0001.5.5.5", "0001.5.5.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "chapter #0001.5.5.5", "0001.5.5.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "chapter №0001.5.5.5", "0001.5.5.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "chapter№0001.5.5.5", "0001.5.5.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "chapter#0001.5.5.5", "0001.5.5.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
	title, expectedChapter = "ch #0001.5.5.5", "0001.5.5.5"
	resultChapter = getChapter(title)
	if resultChapter != expectedChapter {
		t.Errorf("Expected to get chapter \"%s\" from \"%s\", got \"%s\"", expectedChapter, title, resultChapter)
	}
}

func TestCompareChapters(t *testing.T) {
	chapter1, chapter2 := "", ""

	// These tests are for alphabetical sort, since we don't have "ch." before them
	chapter1, chapter2 = "", ""
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "1", "2"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "2", "2.5"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "2.4", "2.5"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "2.4.5", "2.5"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "2.4.5", "2.4.6"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "0000000014", "015"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "0000000014", "015.5.5.5.5.5.5.5.5.5.5"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("\"%s\" should be a previous chapter to \"%s\", and it isn't.", chapter1, chapter2)
	}

	// These tests are for mixed sort
	chapter1, chapter2 = "My Code Can't Be That Bad! 123456", "My Code Can't Be That Bad! Ch. 123457"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! 123456.5", "My Code Can't Be That Bad! Ch. 123457"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! 123456.5.5", "My Code Can't Be That Bad! Ch. 123457"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! Vol. 123456.5.5", "My Code Can't Be That Bad! Ch. 123457"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! 0123", "My Code Can't Be That Bad! Ch. 123457"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! 00000123", "My Code Can't Be That Bad! Ch. 123457"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}

	// something something volumes should go here

	// These tests are for natural sort
	chapter1, chapter2 = "My Code Can't Be That Bad! Ch. 123456", "My Code Can't Be That Bad! Ch. 123457"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! Ch. 123456.5", "My Code Can't Be That Bad! Ch. 123457"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! Ch. 123456.5 [superScans]", "My Code Can't Be That Bad! Ch. 123457 [superScans]"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! Ch 123456.5 [superScans]", "My Code Can't Be That Bad! Ch. 123457 [superScans]"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! chapter 123456.5 [superScans]", "My Code Can't Be That Bad! Ch. 123457 [superScans]"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! ch  123456.5 [superScans]", "My Code Can't Be That Bad! Ch. 123457 [superScans]"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
	chapter1, chapter2 = "My Code Can't Be That Bad! ch.123456.5 [superScans]", "My Code Can't Be That Bad! Ch. 123457 [superScans]"
	if !compareChaptersLess(chapter1, chapter2) {
		t.Errorf("%s should be a previous chapter to %s, and it isn't.", chapter1, chapter2)
	}
}
