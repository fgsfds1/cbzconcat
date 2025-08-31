package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mozillazg/go-unidecode"
)

// ComicInfo structure for metadata
type ComicInfo struct {
	XMLName   xml.Name `xml:"ComicInfo"`
	Title     string   `xml:"Title"`
	Series    string   `xml:"Series"`
	PageCount int      `xml:"PageCount"`
}

// Print if silent flag is not set, or if the verbose flag is set (overrides silent flag)
func printIfNotSilent(msg string, silentFlag *bool, verboseFlag *bool) {
	if !*silentFlag || *verboseFlag {
		fmt.Println(msg)
	}
}

func printIfVerbose(msg string, verboseFlag *bool) {
	if *verboseFlag {
		fmt.Println(msg)
	}
}

func readXmlFromZip(filepath string) (ComicInfo, error) {
	result := new(ComicInfo)
	r, err := zip.OpenReader(filepath)
	if err != nil {
		return *result, err
	}
	for _, file := range r.File {
		if strings.Contains(file.Name, ".xml") {
			rc, _ := file.Open()
			data, err := io.ReadAll(rc)
			if err != nil {
				return *result, err
			}
			err = xml.Unmarshal(data, result)
			if err != nil {
				return *result, err
			}
			rc.Close()
			return *result, nil
		}
	}
	r.Close()

	return *result, fmt.Errorf("no XMLs found in %s", filepath)
}

// getChapter extracts the chapter string like "0015", "0015.5", "0015.5.5" from a filename.
// Returns "" if nothing is found.
func getChapter(name string) string {
	result := ""
	// Regex: match "Ch" + optional separator + digits + optional (.digits)* pattern
	// Example matches: Ch0015, Ch-0015.5, Ch_0015.5.5
	regex := regexp.MustCompile(`(?i)ch(?:|ap|apter)[^0-9]{0,2}(\d+(?:\.\d+)*)`)
	// This is a fallback regex, it tries to match any 3+ digit number. 3 and more digits so we don't match volumes
	// Maybe try to match all numbers, but choose the latter? Should be the volume number, probably.
	fallbackRegex := regexp.MustCompile(`(?i)(\d{3,}(?:\.\d+)*)`)

	matches := regex.FindStringSubmatch(name)
	if len(matches) > 1 {
		result = matches[1] // first capturing group is the number string
	} else {
		matches = fallbackRegex.FindStringSubmatch(name)
		if len(matches) > 1 {
			result = matches[1]
		}
	}
	// Trim leading zeros but preserve zero chapters
	// e.g. "0015" -> "15", but "0000" -> "0" and "0000.0" -> "0.0"
	if result != "" {
		parts := strings.Split(result, ".")
		parts[0] = strings.TrimLeft(parts[0], "0")
		if parts[0] == "" {
			parts[0] = "0"
		}
		result = strings.Join(parts, ".")
	}

	return result
}

// compareChaptersLess does a "natural" comparison based on chapter numbers.
// It splits chapter strings into number slices, then compares piece by piece.
func compareChaptersLess(name1 string, name2 string) bool {
	ch1 := getChapter(name1)
	ch2 := getChapter(name2)

	if ch1 == "" && ch2 == "" {
		// fallback: plain natural string comparison if no chapters found
		return stringNatCmpLess(name1, name2)
	}
	if ch1 == "" {
		return false // put ones without chapter at the end
	}
	if ch2 == "" {
		return true
	}

	// Split into parts (e.g. "15.5.5" -> ["15","5","5"])
	parts1 := strings.Split(ch1, ".")
	parts2 := strings.Split(ch2, ".")

	// Compare each numeric part
	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		n1, _ := strconv.Atoi(parts1[i])
		n2, _ := strconv.Atoi(parts2[i])
		if n1 != n2 {
			return n1 < n2
		}
	}

	// If all compared parts equal, shorter one comes first
	return len(parts1) < len(parts2)
}

func sanitizeFilename(name string) string {
	// Replace spaces and dots with underscores
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, ".", "_")

	// Remove illegal characters (Windows reserved: <>:"/\|?*)
	illegal := regexp.MustCompile(`[<>:"/\\|?*]+`)
	name = illegal.ReplaceAllString(name, "_")

	// Trim leading/trailing underscores and dots
	name = strings.Trim(name, "._ ")

	if name == "" {
		return "untitled"
	}
	return name
}

func sanitizeFilenameASCII(name string) string {
	return sanitizeFilename(unidecode.Unidecode(name))
}

// findCBZFiles recursively searches for CBZ files in the given directory
func findCBZFiles(inputDir string) ([]string, error) {
	var cbzFiles []string
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".cbz") {
			cbzFiles = append(cbzFiles, path)
		}
		return nil
	})
	return cbzFiles, err
}

// compareStringsNaturally performs natural string sorting by comparing strings
// character by character, treating consecutive digits as numbers for proper numerical ordering.
// This is useful for sorting filenames that contain numbers.
func stringNatCmpLess(s1, s2 string) bool {
	i, j := 0, 0

	for i < len(s1) && j < len(s2) {
		// If both characters are digits, compare as numbers
		if isDigit(s1[i]) && isDigit(s2[j]) {
			// Extract numbers from both strings
			num1, len1 := extractNumber(s1[i:])
			num2, len2 := extractNumber(s2[j:])

			if num1 != num2 {
				return num1 < num2
			}

			i += len1
			j += len2
		} else {
			// Compare characters normally
			if s1[i] != s2[j] {
				return s1[i] < s2[j]
			}
			i++
			j++
		}
	}

	// If we reach here, one string is a prefix of the other
	return len(s1) < len(s2)
}

// isDigit checks if a byte represents a digit
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// extractNumber extracts a number from the beginning of a string
// Returns the number and the length of the number in the string
func extractNumber(s string) (int, int) {
	num := 0
	length := 0

	for length < len(s) && isDigit(s[length]) {
		num = num*10 + int(s[length]-'0')
		length++
	}

	return num, length
}
