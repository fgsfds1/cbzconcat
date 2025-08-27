package main

import (
	"archive/zip"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ComicInfo structure for metadata
type ComicInfo struct {
	XMLName   xml.Name `xml:"ComicInfo"`
	Title     string   `xml:"Title"`
	Series    string   `xml:"Series"`
	PageCount int      `xml:"PageCount"`
}

func printIfNotSilent(msg string, silentFlag *bool) {
	if !*silentFlag {
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
			break
		}
		return *result, fmt.Errorf("No XMLs found in %s", filepath)
	}
	r.Close()

	return *result, nil
}

// getChapter extracts the chapter string like "0015", "0015.5", "0015.5.5" from a filename.
// Returns "" if nothing is found.
func getChapter(name string) string {
	// Regex: match "Ch" + optional separator + digits + optional (.digits)* pattern
	// Example matches: Ch0015, Ch-0015.5, Ch_0015.5.5
	regex := regexp.MustCompile(`(?i)Ch[^0-9]*(\d+(?:\.\d+)*)`)

	matches := regex.FindStringSubmatch(name)
	if len(matches) > 1 {
		return matches[1] // first capturing group is the number string
	}
	return ""
}

// compareChapters does a "natural" comparison based on chapter numbers.
// It splits chapter strings into number slices, then compares piece by piece.
func compareChapters(name1 string, name2 string) bool {
	ch1 := getChapter(name1)
	ch2 := getChapter(name2)

	if ch1 == "" && ch2 == "" {
		// fallback: plain string comparison if no chapters found
		return name1 < name2
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

/*
// A function to compare two strings with chapter numbers inside them
// This function should call the getChapter function to extract the chapter numbers
// Something close to natural sort comparison
// Don't change based on leading zeroes (000015 < 016)
// Chapters can include dots (0015 < 0015.5 < 0016)
// Possibly even multiple dots (0015 < 0015.5 < 0015.5.5 < 0016)
// Returns true on name1 < name2
func compareChapters(name1 string, name2 string) bool {
	return true
}

func getChapter(name string) string {
	result := ""
	fmt.Println(name)
	// For now, only extract Ch<any symbol possibly><four digits><dot and more digits, possibly multiple times>
	// The chapter number should be matched in a capturing group
	regex := regexp.MustCompile("Ch.?(\\d{4}(\\.\\d+)*)")

	matches := regex.FindAllStringSubmatchIndex(name, -1)
	// Print all matches, just for debugging
	for _, match := range matches {
		fmt.Printf("%s\n", match)
	}
	// print all groups?
	println("%s\n", regex.SubexpNames())
	// get the captured group here, and return the chapter number
	return result
} */

func main() {
	// Parse flags
	showXML := flag.Bool("x", false, "Print resulting XML (in the resulting cbz archive)")
	printOrder := flag.Bool("r", false, "Print the order of the input cbz files")
	runSilent := flag.Bool("s", false, "Whether to produce any stdout output at all; errors will still be output; overrides other output flags")
	flag.Parse()

	// We should have only two args left - the input dir and the output name
	if flag.NArg() != 2 {
		fmt.Println("Usage: cbzconcat [flags] <input_dir> <output.cbz>")
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	inputDir, outputFile := flag.Arg(0), flag.Arg(1)

	// Find CBZ files
	var cbzFiles []string
	filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".cbz") {
			cbzFiles = append(cbzFiles, path)
		}
		return nil
	})

	if len(cbzFiles) == 0 {
		fmt.Println("No CBZ files found")
		os.Exit(1)
	}

	// Print the original order of the files, for debugging
	if *printOrder {
		printIfNotSilent("Original order:", runSilent)
		for _, name := range cbzFiles {
			printIfNotSilent(name, runSilent)
		}
	}

	// Sort files using the helper functions
	sort.Slice(cbzFiles, func(i, j int) bool {
		return compareChapters(cbzFiles[i], cbzFiles[j])
	})

	// Print the order of the files
	if *printOrder {
		printIfNotSilent("The files will be concatenated in the following order:", runSilent)
		for _, name := range cbzFiles {
			printIfNotSilent(name, runSilent)
		}
	}

	// Get basic book info from the first file, and the last chapter number from the last file
	firstComicInfo, err := readXmlFromZip(cbzFiles[0])
	if err != nil {
		panic(err)
	}
	firstXMLBytes, err := xml.MarshalIndent(firstComicInfo, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(firstXMLBytes[:]))

	lastComicInfo, err := readXmlFromZip(cbzFiles[len(cbzFiles)-1])
	if err != nil {
		panic(err)
	}
	lastXMLBytes, err := xml.MarshalIndent(lastComicInfo, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(lastXMLBytes[:]))

	seriesName := firstComicInfo.Series
	firstChapter := firstComicInfo.Title
	lastChapter := lastComicInfo.Title
	title := fmt.Sprintf("%s %s - %s", seriesName, firstChapter, lastChapter)
	getChapter(firstChapter)
	getChapter(lastChapter)

	// Create output CBZ
	out, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	outZipFile := zip.NewWriter(out)
	defer outZipFile.Close()

	// Starting with the first page, for each archive, read it, get all images inside (opened in the order they were added to the zip file (!))
	// and write them to the `outZipFile` one-by-one, with the filename `pageIndex`
	pageIndex := 1
	for _, cbz := range cbzFiles {
		r, err := zip.OpenReader(cbz)
		if err != nil {
			panic(err)
		}
		for _, f := range r.File {
			// Copy only image files
			ext := strings.ToLower(filepath.Ext(f.Name))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
				rc, _ := f.Open()
				filename := fmt.Sprintf("%05d%s", pageIndex, ext)
				pageIndex++
				w, _ := outZipFile.Create(filename)
				io.Copy(w, rc)
				rc.Close()
			}
		}
		r.Close()
	}

	// Add ComicInfo.xml
	info := ComicInfo{
		Title:     title,
		Series:    filepath.Base(inputDir),
		PageCount: pageIndex - 1,
	}
	xmlBytes, _ := xml.MarshalIndent(info, "", "  ")

	if *showXML {
		printIfNotSilent(fmt.Sprintf("Resulting XML written to %s:", outputFile), runSilent)
		printIfNotSilent(string(xmlBytes[:]), runSilent)
	}

	w, _ := outZipFile.Create("ComicInfo.xml")
	w.Write([]byte(xml.Header))
	w.Write(xmlBytes)

	printIfNotSilent(fmt.Sprintf("Merged %d files into %s with %d pages\n", len(cbzFiles), outputFile, pageIndex-1), runSilent)
}
