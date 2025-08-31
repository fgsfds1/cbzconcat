package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Helper test functions to capture stdout, used in tests that test over stdout
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
	testCases := []struct {
		silentFlag  bool
		verboseFlag bool
		message     string
		shouldPrint bool
		description string
	}{
		// Basic functionality tests
		{false, false, "1_should_print", true, "Should print when not silent and not verbose"},
		{true, false, "2_shouldnt_print", false, "Should not print when silent and not verbose"},
		{true, true, "3_should_print", true, "Should print when silent but verbose overrides"},
		{false, true, "4_should_print", true, "Should print when not silent and verbose"},

		// Edge cases
		{false, false, "", true, "Empty string should print when not silent"},
		{false, false, "test\nwith\nnewlines", true, "Newlines should print when not silent"},
		{false, false, "test\twith\ttabs", true, "Tabs should print when not silent"},
		{false, false, "test with spaces", true, "Spaces should print when not silent"},
		{false, false, strings.Repeat("a", 1000), true, "Long string should print when not silent"},
		{false, false, "test with unicode: 漫画", true, "Unicode should print when not silent"},
	}

	for _, tc := range testCases {
		originalStdout, r, w := setupStdout(t)

		silentFlag := &tc.silentFlag
		verboseFlag := &tc.verboseFlag

		printIfNotSilent(tc.message, silentFlag, verboseFlag)

		output := getStdoutAndClose(originalStdout, w, r)

		if tc.shouldPrint {
			if !strings.Contains(output, tc.message) {
				t.Errorf("Test '%s': Expected message '%s' to be printed, but it wasn't", tc.description, tc.message)
			}
		} else {
			if strings.Contains(output, tc.message) {
				t.Errorf("Test '%s': Expected message '%s' NOT to be printed, but it was", tc.description, tc.message)
			}
		}
	}
}

func TestPrintIfVerbose(t *testing.T) {
	testCases := []struct {
		verboseFlag bool
		message     string
		shouldPrint bool
		description string
	}{
		// Basic functionality tests
		{false, "1_shouldnt_print", false, "Should not print when not verbose"},
		{true, "2_should_print", true, "Should print when verbose"},

		// Edge cases
		{true, "", true, "Empty string should print when verbose"},
		{true, "test\nwith\nnewlines", true, "Newlines should print when verbose"},
		{true, "test\twith\ttabs", true, "Tabs should print when verbose"},
		{true, "test with spaces", true, "Spaces should print when verbose"},
		{true, strings.Repeat("a", 1000), true, "Long string should print when verbose"},
		{true, "test with unicode: 漫画", true, "Unicode should print when verbose"},
	}

	for _, tc := range testCases {
		originalStdout, r, w := setupStdout(t)

		verboseFlag := &tc.verboseFlag

		printIfVerbose(tc.message, verboseFlag)

		output := getStdoutAndClose(originalStdout, w, r)

		if tc.shouldPrint {
			if !strings.Contains(output, tc.message) {
				t.Errorf("Test '%s': Expected message '%s' to be printed, but it wasn't", tc.description, tc.message)
			}
		} else {
			if strings.Contains(output, tc.message) {
				t.Errorf("Test '%s': Expected message '%s' NOT to be printed, but it was", tc.description, tc.message)
			}
		}
	}
}

func TestGetChapter(t *testing.T) {
	testCases := []struct {
		title           string
		expectedChapter string
		description     string
	}{
		// Basic chapter extraction tests
		{"", "", "Empty title should return empty chapter"},
		{"Ch.0000", "0", "Basic Ch. prefix with 4 digits - zero"},
		{"Ch.0001", "1", "Basic Ch. prefix with 4 digits"},
		{"Ch.0001.5", "1.5", "Ch. prefix with decimal"},
		{"Ch 0001.5", "1.5", "Ch prefix with space separator"},
		{"Ch  0001.5", "1.5", "Ch prefix with multiple spaces"},
		{"Ch 0001.5.5.5", "1.5.5.5", "Ch prefix with multiple decimal parts"},
		{"ch 0001.5.5.5", "1.5.5.5", "Lowercase ch prefix"},
		{"ch. 0001.5.5.5", "1.5.5.5", "Lowercase ch. prefix"},
		{"chapter 0001.5.5.5", "1.5.5.5", "Full 'chapter' prefix"},
		{"chapter0001.5.5.5", "1.5.5.5", "No separator after 'chapter'"},
		{"chapter #0001.5.5.5", "1.5.5.5", "Hash separator after 'chapter'"},
		{"chapter №0001.5.5.5", "1.5.5.5", "No. separator after 'chapter'"},
		{"chapter№0001.5.5.5", "1.5.5.5", "No separator after 'chapter' with No."},
		{"chapter#0001.5.5.5", "1.5.5.5", "No separator after 'chapter' with hash"},
		{"ch #0001.5.5.5", "1.5.5.5", "Hash separator after 'ch'"},

		// Fallback regex tests (3+ digits without ch prefix)
		{"My Manga 001", "1", "3-digit number without ch prefix"},
		{"My Manga 001.5", "1.5", "3-digit decimal without ch prefix"},
		{"My Manga 001.5.5", "1.5.5", "3-digit multi-decimal without ch prefix"},
		{"My Manga 0001", "1", "4-digit number without ch prefix"},
		{"My Manga 0001.5", "1.5", "4-digit decimal without ch prefix"},
		{"My Manga 0001.5.5", "1.5.5", "4-digit multi-decimal without ch prefix"},

		// Edge cases for chapter extraction
		{"Ch001", "1", "Ch prefix with no separator"},
		{"Ch-001", "1", "Ch prefix with dash separator"},
		{"Ch_001", "1", "Ch prefix with underscore separator"},
		{"Ch.001", "1", "Ch prefix with dot separator"},
		{"Ch:001", "1", "Ch prefix with colon separator"},
		{"Ch;001", "1", "Ch prefix with semicolon separator"},
		{"Ch,001", "1", "Ch prefix with comma separator"},
		{"Ch!001", "1", "Ch prefix with exclamation separator"},
		{"Ch?001", "1", "Ch prefix with question separator"},
		{"Ch 001", "1", "Ch prefix with space separator"},
		{"Ch  001", "1", "Ch prefix with multiple spaces"},
		{"Ch\t001", "1", "Ch prefix with tab separator"},
		{"Ch\n001", "1", "Ch prefix with newline separator"},

		// Case variations
		{"CH001", "1", "Uppercase CH"},
		{"ch001", "1", "Lowercase ch"},
		{"Ch001", "1", "Mixed case Ch"},
		{"cH001", "1", "Mixed case cH"},

		// Chapter variations
		{"Chapter001", "1", "Full 'Chapter' prefix"},
		{"CHAPTER001", "1", "Uppercase 'CHAPTER' prefix"},
		{"chapter001", "1", "Lowercase 'chapter' prefix"},
		{"Chap001", "1", "Abbreviated 'Chap' prefix"},

		// Numbers that shouldn't match (less than 3 digits)
		{"My Manga 12", "", "2-digit number should not match fallback"},
		{"My Manga 1", "", "1-digit number should not match fallback"},
		{"My Manga 0", "", "0 should not match fallback"},
		// Numbers that should match (3+ digits)
		{"123", "123", "3-digit number should match fallback"},
		{"012", "12", "3-digit number should match fallback"},
		{"12", "", "2-digit number should not match fallback"},
		{"1", "", "1-digit number should not match fallback"},
		{"0", "", "0 should not match fallback"},

		// Text after numbers
		{"Ch001 [END]", "1", "Chapter with text after"},
		{"Ch001.5 [END]", "1.5", "Decimal chapter with text after"},
		{"Ch001.5.5 [END]", "1.5.5", "Multi-decimal chapter with text after"},

		// Text before numbers
		{"[START] Ch001", "1", "Chapter with text before"},
		{"[START] Ch001.5", "1.5", "Decimal chapter with text before"},

		// Multiple numbers (should pick the first chapter match)
		{"Ch001 Vol002", "1", "Chapter should take precedence over volume"},
		{"Vol002 Ch001", "1", "Chapter should take precedence over volume"},

		// Edge cases for decimal numbers
		{"Ch001.", "1", "Chapter ending with dot"},
		{"Ch001.5.", "1.5", "Decimal chapter ending with dot"},
		{"Ch001..5", "1", "Chapter with double dot (should stop at first dot)"},
		{"Ch001.5..5", "1.5", "Decimal chapter with double dot"},

		// Very long chapter numbers
		{"Ch123456789", "123456789", "Very long chapter number"},
		{"Ch123456789.987654321", "123456789.987654321", "Very long decimal chapter"},

		// Zero values
		{"Ch000", "0", "Chapter with all zeros"},
		{"Ch000.0", "0.0", "Decimal chapter with zeros"},
		{"Ch000.0.0", "0.0.0", "Multi-decimal chapter with zeros"},

		// Negative numbers (should still extract number)
		{"Ch-001", "1", "Negative chapter should still extract number"},
		{"Ch-001.5", "1.5", "Negative decimal chapter should still extract number"},

		// Special characters in chapter numbers
		{"Ch001_5", "1", "Underscore in chapter should not be part of number"},
		{"Ch001-5", "1", "Dash in chapter should not be part of number"},

		// No valid chapter
		{"My Manga Title", "", "No chapter information"},
		{"Ch", "", "Just 'Ch' with no number"},
		{"Chapter", "", "Just 'Chapter' with no number"},
		{"123", "123", "3-digit number should match fallback"},
		{"12", "", "2-digit number should not match fallback"},
		{"1", "", "1-digit number should not match fallback"},
		{"0", "", "0 should not match fallback"},
	}

	for _, tc := range testCases {
		result := getChapter(tc.title)
		if result != tc.expectedChapter {
			t.Errorf("Test '%s': Expected chapter '%s' from '%s', got '%s'",
				tc.description, tc.expectedChapter, tc.title, result)
		}
	}
}

func TestCompareChapters(t *testing.T) {
	testCases := []struct {
		chapter1       string
		chapter2       string
		expectedResult bool
		description    string
	}{
		// Basic alphabetical sort (no ch. prefix)
		{"", "", false, "Empty strings should be equal (both have no chapters, so use string comparison)"},
		{"1", "2", true, "1 should be less than 2"},
		{"2", "2.5", true, "2 should be less than 2.5"},
		{"2.4", "2.5", true, "2.4 should be less than 2.5"},
		{"2.4.5", "2.5", true, "2.4.5 should be less than 2.5"},
		{"2.4.5", "2.4.6", true, "2.4.5 should be less than 2.4.6"},
		{"0000000014", "015", true, "0000000014 should be less than 015"},
		{"0000000014", "015.5.5.5.5.5.5.5.5.5.5", true, "0000000014 should be less than 015.5.5.5.5.5.5.5.5.5.5"},

		// Mixed sort (some with ch. prefix, some without)
		{"My Code Can't Be That Bad! 123456", "My Code Can't Be That Bad! Ch. 123457", true, "123456 should be less than Ch. 123457"},
		{"My Code Can't Be That Bad! 123456.5", "My Code Can't Be That Bad! Ch. 123457", true, "123456.5 should be less than Ch. 123457"},
		{"My Code Can't Be That Bad! 123456.5.5", "My Code Can't Be That Bad! Ch. 123457", true, "123456.5.5 should be less than Ch. 123457"},
		{"My Code Can't Be That Bad! Vol. 123456.5.5", "My Code Can't Be That Bad! Ch. 123457", true, "Vol. 123456.5.5 should be less than Ch. 123457"},
		{"My Code Can't Be That Bad! 0123", "My Code Can't Be That Bad! Ch. 123457", true, "0123 should be less than Ch. 123457"},
		{"My Code Can't Be That Bad! 00000123", "My Code Can't Be That Bad! Ch. 123457", true, "00000123 should be less than Ch. 123457"},
		{"My Code Can't Be That Bad! Vol. 016 010", "My Code Can't Be That Bad! Vol. 006 Ch. 015", false, "Vol. 016 010 has no chapter, should go to end (be greater)"},

		// Natural sort (with ch. prefix)
		{"My Code Can't Be That Bad! Ch. 0002", "My Code Can't Be That Bad! Ch. 0010", true, "Ch. 0002 should be less than Ch. 0010"},
		{"My Code Can't Be That Bad! Ch. 123456", "My Code Can't Be That Bad! Ch. 123457", true, "Ch. 123456 should be less than Ch. 123457"},
		{"My Code Can't Be That Bad! Ch. 123456.5", "My Code Can't Be That Bad! Ch. 123457", true, "Ch. 123456.5 should be less than Ch. 123457"},
		{"My Code Can't Be That Bad! Ch. 123456.5 [superScans]", "My Code Can't Be That Bad! Ch. 123457 [superScans]", true, "Ch. 123456.5 [superScans] should be less than Ch. 123457 [superScans]"},
		{"My Code Can't Be That Bad! Ch 123456.5 [superScans]", "My Code Can't Be That Bad! Ch. 123457 [superScans]", true, "Ch 123456.5 [superScans] should be less than Ch. 123457 [superScans]"},
		{"My Code Can't Be That Bad! chapter 123456.5 [superScans]", "My Code Can't Be That Bad! Ch. 123457 [superScans]", true, "chapter 123456.5 [superScans] should be less than Ch. 123457 [superScans]"},
		{"My Code Can't Be That Bad! ch  123456.5 [superScans]", "My Code Can't Be That Bad! Ch. 123457 [superScans]", true, "ch  123456.5 [superScans] should be less than Ch. 123457 [superScans]"},
		{"My Code Can't Be That Bad! ch.123456.5 [superScans]", "My Code Can't Be That Bad! Ch. 123457 [superScans]", true, "ch.123456.5 [superScans] should be less than Ch. 123457 [superScans]"},

		// Volume designators
		{"My Manga Vol.1 Ch.001", "My Manga Vol.1 Ch.002", true, "Same volume, different chapters"},
		{"My Manga Vol.1 Ch.001", "My Manga Vol.2 Ch.001", false, "Different volumes, same chapter - should compare by chapter only"},
		{"My Manga Vol.1 Ch.001", "My Manga Vol.2 Ch.002", true, "Different volumes and chapters - should compare by chapter"},
		{"My Manga Volume 1 Ch.001", "My Manga Volume 2 Ch.001", false, "Full 'Volume' prefix - should compare by chapter only"},
		{"My Manga V1 Ch.001", "My Manga V2 Ch.001", false, "Abbreviated 'V' prefix - should compare by chapter only"},
		{"My Manga v1 Ch.001", "My Manga v2 Ch.001", false, "Lowercase 'v' prefix - should compare by chapter only"},
		{"My Manga Vol1 Ch.001", "My Manga Vol2 Ch.001", false, "No separator after 'Vol' - should compare by chapter only"},
		{"My Manga Volume1 Ch.001", "My Manga Volume2 Ch.001", false, "No separator after 'Volume' - should compare by chapter only"},
		{"My Manga V.1 Ch.001", "My Manga V.2 Ch.001", false, "V. prefix with dot separator - should compare by chapter only"},
		{"My Manga Vol-1 Ch.001", "My Manga Vol-2 Ch.001", false, "Vol- prefix with dash separator - should compare by chapter only"},
		{"My Manga Vol_1 Ch.001", "My Manga Vol_2 Ch.001", false, "Vol_ prefix with underscore separator - should compare by chapter only"},
		{"My Manga Vol 1 Ch.001", "My Manga Vol 2 Ch.001", false, "Vol prefix with space separator - should compare by chapter only"},
		{"My Manga Vol  1 Ch.001", "My Manga Vol  2 Ch.001", false, "Vol prefix with multiple spaces - should compare by chapter only"},
		{"My Manga Vol.001 Ch.001", "My Manga Vol.002 Ch.001", false, "Vol with leading zeros - should compare by chapter only"},
		{"My Manga Vol.1.5 Ch.001", "My Manga Vol.2.0 Ch.001", false, "Vol with decimal numbers - should compare by chapter only"},
		{"My Manga Vol.1 Ch.001", "My Manga Vol.1 Ch.001.5", true, "Same volume, decimal chapter"},
		{"My Manga Vol.1 Ch.001 [END]", "My Manga Vol.1 Ch.002 [END]", true, "Volume with chapter and brackets"},

		// Equal chapters
		{"Ch001", "Ch001", false, "Equal chapters should return false"},
		{"Ch001.5", "Ch001.5", false, "Equal decimal chapters should return false"},
		{"Ch001.5.5", "Ch001.5.5", false, "Equal multi-decimal chapters should return false"},
		{"My Manga 001", "My Manga 001", false, "Equal chapters without ch prefix should return false"},
		{"My Manga Vol.1 Ch.001", "My Manga Vol.1 Ch.001", false, "Equal volume and chapter should return false"},

		// Boundary conditions for decimal parts
		{"Ch001.9", "Ch002.0", true, "0.9 should be less than 1.0"},
		{"Ch001.99", "Ch002.00", true, "0.99 should be less than 1.00"},
		{"Ch001.999", "Ch002.000", true, "0.999 should be less than 1.000"},
		{"Vol.001.9", "Vol.002.0", true, "Volume 0.9 should be less than 1.0"},
		{"Vol.001.99", "Vol.002.00", true, "Volume 0.99 should be less than 1.00"},

		// Leading zeros
		{"Ch001", "Ch0001", false, "001 should equal 0001"},
		{"Ch0001", "Ch001", false, "0001 should equal 001"},
		{"Ch001.5", "Ch0001.5", false, "001.5 should equal 0001.5"},
		{"Ch0001.5", "Ch001.5", false, "0001.5 should equal 001.5"},
		{"Vol001", "Vol0001", false, "Volume 001 should equal 0001"},
		{"Vol0001", "Vol001", false, "Volume 0001 should equal 001"},

		// Different number of decimal parts
		{"Ch001", "Ch001.0", true, "001 should be less than 001.0"},
		{"Ch001.0", "Ch001", false, "001.0 should be greater than 001"},
		{"Ch001.5", "Ch001.5.0", true, "001.5 should be less than 001.5.0"},
		{"Ch001.5.0", "Ch001.5", false, "001.5.0 should be greater than 001.5"},
		{"Vol001", "Vol001.0", true, "Volume 001 should be less than 001.0"},
		{"Vol001.0", "Vol001", false, "Volume 001.0 should be greater than 001"},

		// Very large numbers
		{"Ch999999", "Ch1000000", true, "999999 should be less than 1000000"},
		{"Ch999999.999", "Ch1000000.000", true, "999999.999 should be less than 1000000.000"},
		{"Vol999999", "Vol1000000", true, "Volume 999999 should be less than 1000000"},
		{"Vol999999.999", "Vol1000000.000", true, "Volume 999999.999 should be less than 1000000.000"},

		// Zero values
		{"Ch000", "Ch001", true, "000 should be less than 001"},
		{"Ch000.0", "Ch001.0", true, "000.0 should be less than 001.0"},
		{"Ch000.0.0", "Ch001.0.0", true, "000.0.0 should be less than 001.0.0"},
		{"Vol000", "Vol001", true, "Volume 000 should be less than 001"},
		{"Vol000.0", "Vol001.0", true, "Volume 000.0 should be less than 001.0"},

		// Single vs multi-part chapters
		{"Ch001", "Ch001.1", true, "Single part should be less than multi-part"},
		{"Ch001.1", "Ch001", false, "Multi-part should be greater than single part"},
		{"Vol001", "Vol001.1", true, "Volume single part should be less than multi-part"},
		{"Vol001.1", "Vol001", false, "Volume multi-part should be greater than single part"},

		// Mixed chapter formats
		{"Ch001", "My Manga 002", true, "Ch001 should be less than 002 (fallback)"},
		{"My Manga 001", "Ch002", true, "001 (fallback) should be less than Ch002"},
		{"Vol001", "My Manga 002", true, "Vol001 should be less than 002 (fallback)"},
		{"My Manga 001", "Vol002", true, "001 (fallback) should be less than Vol002"},

		// String comparison fallback
		{"", "Ch001", false, "Empty string should be greater than any chapter (empty has no chapter, so goes to end)"},
		{"Ch001", "", true, "Any chapter should be less than empty string (empty has no chapter, so goes to end)"},
		{"A", "B", true, "A should be less than B in string comparison"},
		{"B", "A", false, "B should be greater than A in string comparison"},
		{"", "Vol001", false, "Empty string should be greater than any volume (empty has no chapter, so goes to end)"},
		{"Vol001", "", true, "Any volume should be less than empty string (empty has no chapter, so goes to end)"},

		// Special characters in filenames
		{"Ch001 [END]", "Ch002 [END]", true, "Chapters with brackets should compare correctly"},
		{"Ch001.5 [END]", "Ch002.5 [END]", true, "Decimal chapters with brackets should compare correctly"},
		{"Ch001_v2", "Ch002_v1", true, "Chapters with version suffixes should compare correctly"},
		{"Vol001 [END]", "Vol002 [END]", true, "Volumes with brackets should compare correctly"},
		{"Vol001.5 [END]", "Vol002.5 [END]", true, "Decimal volumes with brackets should compare correctly"},

		// Very long filenames
		{longFilename("Ch001"), longFilename("Ch002"), true, "Very long filenames should compare correctly"},
		{longFilename("Ch002"), longFilename("Ch001"), false, "Very long filenames should compare correctly (reverse)"},
		{longFilename("Vol001"), longFilename("Vol002"), true, "Very long volume filenames should compare correctly"},

		// Unicode characters
		{"Ch001 漫画", "Ch002 漫画", true, "Chapters with unicode should compare correctly"},
		{"Ch001.5 漫画", "Ch002.5 漫画", true, "Decimal chapters with unicode should compare correctly"},
		{"Vol001 漫画", "Vol002 漫画", true, "Volumes with unicode should compare correctly"},

		// Numbers that are close but different
		{"Ch001.999999", "Ch002.000001", true, "Very close decimal chapters should compare correctly"},
		{"Ch001.000001", "Ch001.000002", true, "Very close decimal chapters should compare correctly"},
		{"Vol001.999999", "Vol002.000001", true, "Very close decimal volumes should compare correctly"},
		{"Vol001.000001", "Vol001.000002", true, "Very close decimal volumes should compare correctly"},

		// Edge case decimal parts
		{"Ch001.0", "Ch001.1", true, "0.0 should be less than 0.1"},
		{"Ch001.1", "Ch001.0", false, "0.1 should be greater than 0.0"},
		{"Ch001.00", "Ch001.01", true, "0.00 should be less than 0.01"},
		{"Ch001.01", "Ch001.00", false, "0.01 should be greater than 0.00"},
		{"Vol001.0", "Vol001.1", true, "Volume 0.0 should be less than 0.1"},
		{"Vol001.1", "Vol001.0", false, "Volume 0.1 should be greater than 0.0"},
	}

	for _, tc := range testCases {
		result := compareChaptersLess(tc.chapter1, tc.chapter2)
		if result != tc.expectedResult {
			t.Errorf("Test '%s': Expected %s < %s to be %v, got %v",
				tc.description, tc.chapter1, tc.chapter2, tc.expectedResult, result)
		}
	}
}

// Helper function to create very long filenames for testing
func longFilename(chapter string) string {
	prefix := "My Very Long Manga Title That Has Many Words And Characters "
	suffix := " With Additional Information And Metadata That Makes The Filename Very Long"
	return prefix + chapter + suffix
}

func TestFindCBZFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cbzconcat_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test subdirectories
	subDir1 := filepath.Join(tempDir, "subdir1")
	subDir2 := filepath.Join(tempDir, "subdir2")
	os.MkdirAll(subDir1, 0755)
	os.MkdirAll(subDir2, 0755)

	// Create test files
	testFiles := []struct {
		path    string
		content string
		isDir   bool
	}{
		{filepath.Join(tempDir, "file1.cbz"), "content1", false},
		{filepath.Join(tempDir, "file2.CBZ"), "content2", false},
		{filepath.Join(tempDir, "file3.txt"), "content3", false},
		{filepath.Join(tempDir, "file4.cbz"), "content4", false},
		{filepath.Join(subDir1, "nested1.cbz"), "nested1", false},
		{filepath.Join(subDir1, "nested2.txt"), "nested2", false},
		{filepath.Join(subDir2, "deep.cbz"), "deep", false},
		{filepath.Join(tempDir, "empty_dir"), "", true},
	}

	// Create the test files and directories
	for _, tf := range testFiles {
		if tf.isDir {
			os.MkdirAll(tf.path, 0755)
		} else {
			err := os.WriteFile(tf.path, []byte(tf.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", tf.path, err)
			}
		}
	}

	// Test cases
	testCases := []struct {
		name        string
		inputDir    string
		expected    []string
		expectError bool
	}{
		{
			name:     "find all cbz files in root and subdirectories",
			inputDir: tempDir,
			expected: []string{
				filepath.Join(tempDir, "file1.cbz"),
				filepath.Join(tempDir, "file2.CBZ"),
				filepath.Join(tempDir, "file4.cbz"),
				filepath.Join(subDir1, "nested1.cbz"),
				filepath.Join(subDir2, "deep.cbz"),
			},
			expectError: false,
		},
		{
			name:        "non-existent directory",
			inputDir:    filepath.Join(tempDir, "nonexistent"),
			expected:    nil,
			expectError: true,
		},
		{
			name:     "subdirectory with one cbz file",
			inputDir: subDir1,
			expected: []string{
				filepath.Join(subDir1, "nested1.cbz"),
			},
			expectError: false,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := findCBZFiles(tc.inputDir)

			// Check error expectations
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check results
			if !tc.expectError {
				// Sort both slices for comparison since filepath.Walk order is not guaranteed
				sort.Strings(result)
				sort.Strings(tc.expected)

				if !reflect.DeepEqual(result, tc.expected) {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			}
		})
	}
}

func TestFindCBZFilesCaseInsensitive(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cbzconcat_case_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files with different case extensions
	testFiles := []string{
		"file1.cbz",
		"file2.CBZ",
		"file3.Cbz",
		"file4.cBz",
		"file5.txt",
	}

	for _, filename := range testFiles {
		path := filepath.Join(tempDir, filename)
		err := os.WriteFile(path, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	// Test that all case variations of .cbz are found
	result, err := findCBZFiles(tempDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedCount := 4 // Should find all .cbz files regardless of case
	if len(result) != expectedCount {
		t.Errorf("Expected %d CBZ files, got %d", expectedCount, len(result))
	}

	// Verify that .txt file is not included
	for _, file := range result {
		if filepath.Ext(strings.ToLower(file)) != ".cbz" {
			t.Errorf("Found non-CBZ file: %s", file)
		}
	}
}
