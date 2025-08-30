package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ProgressUpdate represents a progress update message
type ProgressUpdate struct {
	value  float64
	status string
}

// GUIApp represents the main GUI application
type GUIApp struct {
	app             fyne.App
	window          fyne.Window
	inputDirEntry   *widget.Entry
	outputDirEntry  *widget.Entry
	showXMLCheck    *widget.Check
	printOrderCheck *widget.Check
	silentCheck     *widget.Check
	verboseCheck    *widget.Check
	statusLabel     *widget.Label
	progressBar     *widget.ProgressBar
	progressChan    chan ProgressUpdate
}

// NewGUIApp creates a new GUI application instance
func NewGUIApp() *GUIApp {
	app := app.New()
	app.SetIcon(nil) // You can set an icon here if you have one

	gui := &GUIApp{
		app:          app,
		progressChan: make(chan ProgressUpdate, 100),
	}

	gui.createWindow()
	gui.startProgressListener()
	return gui
}

// createWindow sets up the main window and its contents
func (gui *GUIApp) createWindow() {
	gui.window = gui.app.NewWindow("CBZ Concat - GUI")
	gui.window.Resize(fyne.NewSize(600, 500))

	// Create input directory selection
	inputLabel := widget.NewLabel("Input Directory:")
	gui.inputDirEntry = widget.NewEntry()
	gui.inputDirEntry.SetPlaceHolder("Select directory containing CBZ files...")
	inputBrowseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				gui.inputDirEntry.SetText(uri.Path())
			}
		}, gui.window)
	})
	inputContainer := container.NewBorder(nil, nil, nil, inputBrowseBtn, gui.inputDirEntry)

	// Create output directory selection
	outputLabel := widget.NewLabel("Output Directory:")
	gui.outputDirEntry = widget.NewEntry()
	gui.outputDirEntry.SetPlaceHolder("Select output directory...")
	outputBrowseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				gui.outputDirEntry.SetText(uri.Path())
			}
		}, gui.window)
	})
	outputContainer := container.NewBorder(nil, nil, nil, outputBrowseBtn, gui.outputDirEntry)

	// Create options checkboxes
	optionsLabel := widget.NewLabel("Options:")
	gui.showXMLCheck = widget.NewCheck("Show XML output", nil)
	gui.printOrderCheck = widget.NewCheck("Print file order", nil)
	gui.silentCheck = widget.NewCheck("Silent mode", nil)
	gui.verboseCheck = widget.NewCheck("Verbose output", nil)

	// Create status and progress
	gui.statusLabel = widget.NewLabel("Ready to concatenate CBZ files")
	gui.progressBar = widget.NewProgressBar()
	gui.progressBar.Hide()

	// Create concatenate button
	concatBtn := widget.NewButton("Concatenate CBZ Files", func() {
		gui.runConcatenation()
	})

	// Create clear button
	clearBtn := widget.NewButton("Clear Fields", func() {
		gui.clearFields()
	})

	// Layout the window
	optionsContainer := container.NewVBox(
		gui.showXMLCheck,
		gui.printOrderCheck,
		gui.silentCheck,
		gui.verboseCheck,
	)

	buttonContainer := container.NewHBox(concatBtn, clearBtn)

	content := container.NewVBox(
		inputLabel,
		inputContainer,
		widget.NewSeparator(),
		outputLabel,
		outputContainer,
		widget.NewSeparator(),
		optionsLabel,
		optionsContainer,
		widget.NewSeparator(),
		gui.statusLabel,
		gui.progressBar,
		buttonContainer,
	)

	gui.window.SetContent(content)
}

// clearFields resets all input fields to their default state
func (gui *GUIApp) clearFields() {
	gui.inputDirEntry.SetText("")
	gui.outputDirEntry.SetText("")
	gui.showXMLCheck.SetChecked(false)
	gui.printOrderCheck.SetChecked(false)
	gui.silentCheck.SetChecked(false)
	gui.verboseCheck.SetChecked(false)
	gui.statusLabel.SetText("Ready to concatenate CBZ files")
	gui.progressBar.Hide()
}

// runConcatenation executes the CBZ concatenation process
func (gui *GUIApp) runConcatenation() {
	inputDir := gui.inputDirEntry.Text
	outputDir := gui.outputDirEntry.Text

	// Validate inputs
	if inputDir == "" || outputDir == "" {
		dialog.ShowError(fmt.Errorf("Please select both input and output directories"), gui.window)
		return
	}

	// Check if input directory exists and contains CBZ files
	if !gui.validateInputDirectory(inputDir) {
		return
	}

	// Check if output directory exists and is writable
	if !gui.validateOutputDirectory(outputDir) {
		return
	}

	// Show progress
	gui.progressBar.Show()
	gui.progressBar.SetValue(0)
	gui.statusLabel.SetText("Starting concatenation...")

	// Run concatenation in a goroutine to avoid blocking the UI
	go func() {
		gui.executeConcatenation(inputDir, outputDir)
	}()
}

// validateInputDirectory checks if the input directory is valid
func (gui *GUIApp) validateInputDirectory(inputDir string) bool {
	// Check if directory exists
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("Input directory does not exist: %s", inputDir), gui.window)
		return false
	}

	// Check if it contains CBZ files
	cbzFiles, err := gui.findCBZFiles(inputDir)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error reading input directory: %v", err), gui.window)
		return false
	}

	if len(cbzFiles) == 0 {
		dialog.ShowError(fmt.Errorf("No CBZ files found in input directory: %s", inputDir), gui.window)
		return false
	}

	if len(cbzFiles) == 1 {
		dialog.ShowError(fmt.Errorf("Only one CBZ file found - no concatenation needed"), gui.window)
		return false
	}

	return true
}

// validateOutputDirectory checks if the output directory is valid
func (gui *GUIApp) validateOutputDirectory(outputDir string) bool {
	// Check if directory exists, create if not
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Cannot create output directory: %v", err), gui.window)
			return false
		}
	}

	// Check if directory is writable
	if _, err := os.Stat(outputDir); err == nil {
		testFile := filepath.Join(outputDir, ".test_write")
		file, err := os.Create(testFile)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Output directory is not writable: %v", err), gui.window)
			return false
		}
		file.Close()
		os.Remove(testFile)
	}

	return true
}

// findCBZFiles finds all CBZ files in the given directory
func (gui *GUIApp) findCBZFiles(inputDir string) ([]string, error) {
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

// executeConcatenation runs the actual concatenation process
func (gui *GUIApp) executeConcatenation(inputDir, outputDir string) {
	// Update UI to show progress
	gui.updateProgress(0.1, "Finding CBZ files...")

	// Find CBZ files
	cbzFiles, err := gui.findCBZFiles(inputDir)
	if err != nil {
		gui.showError(fmt.Sprintf("Error finding CBZ files: %v", err))
		return
	}

	gui.updateProgress(0.2, fmt.Sprintf("Found %d CBZ files, sorting...", len(cbzFiles)))

	// Sort files
	sort.Slice(cbzFiles, func(i, j int) bool {
		return compareChaptersLess(cbzFiles[i], cbzFiles[j])
	})

	gui.updateProgress(0.3, "Reading metadata...")

	// Read metadata from first and last files
	firstComicInfo, err := readXmlFromZip(cbzFiles[0])
	if err != nil {
		gui.showError(fmt.Sprintf("Error reading first file metadata: %v", err))
		return
	}

	lastComicInfo, err := readXmlFromZip(cbzFiles[len(cbzFiles)-1])
	if err != nil {
		gui.showError(fmt.Sprintf("Error reading last file metadata: %v", err))
		return
	}

	gui.updateProgress(0.4, "Creating output file...")

	// Generate output filename
	seriesName := firstComicInfo.Series
	firstChapter := getChapter(firstComicInfo.Title)
	lastChapter := getChapter(lastComicInfo.Title)
	title := fmt.Sprintf("%s Ch.%s-%s", seriesName, firstChapter, lastChapter)
	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s.cbz", sanitizeFilenameASCII(title)))

	// Create output CBZ
	out, err := os.Create(outputFile)
	if err != nil {
		gui.showError(fmt.Sprintf("Error creating output file: %v", err))
		return
	}
	defer out.Close()

	outZipFile := zip.NewWriter(out)
	defer outZipFile.Close()

	gui.updateProgress(0.5, "Processing CBZ files...")

	// Process each CBZ file
	pageIndex := 1
	totalFiles := len(cbzFiles)

	for i, cbz := range cbzFiles {
		progress := 0.5 + (float64(i)/float64(totalFiles))*0.4
		gui.updateProgress(progress, fmt.Sprintf("Processing file %d of %d: %s", i+1, totalFiles, filepath.Base(cbz)))

		r, err := zip.OpenReader(cbz)
		if err != nil {
			gui.showError(fmt.Sprintf("Error reading CBZ file %s: %v", cbz, err))
			return
		}

		for _, f := range r.File {
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

	gui.updateProgress(0.9, "Finalizing output...")

	// Add ComicInfo.xml
	info := ComicInfo{
		Title:     title,
		Series:    seriesName,
		PageCount: pageIndex - 1,
	}
	xmlBytes, _ := xml.MarshalIndent(info, "", "  ")

	w, _ := outZipFile.Create("ComicInfo.xml")
	w.Write([]byte(xml.Header))
	w.Write(xmlBytes)

	gui.updateProgress(1.0, fmt.Sprintf("Success! Created %s with %d pages", filepath.Base(outputFile), pageIndex-1))

	// Show success dialog
	dialog.ShowInformation("Success",
		fmt.Sprintf("Successfully concatenated %d CBZ files into:\n%s\n\nTotal pages: %d",
			len(cbzFiles), outputFile, pageIndex-1),
		gui.window)
}

// startProgressListener starts a goroutine to listen for progress updates
func (gui *GUIApp) startProgressListener() {
	go func() {
		for update := range gui.progressChan {
			// Use a timer to defer UI updates to the main thread
			time.AfterFunc(1*time.Millisecond, func() {
				gui.progressBar.SetValue(update.value)
				gui.statusLabel.SetText(update.status)
			})
		}
	}()
}

// updateProgress safely updates the progress bar and status label
func (gui *GUIApp) updateProgress(value float64, status string) {
	gui.progressChan <- ProgressUpdate{value: value, status: status}
}

// showError displays an error message in the UI
func (gui *GUIApp) showError(message string) {
	gui.statusLabel.SetText("Error: " + message)
	gui.progressBar.Hide()

	// Show error dialog
	dialog.ShowError(fmt.Errorf(message), gui.window)
}

// Run starts the GUI application
func (gui *GUIApp) Run() {
	gui.window.ShowAndRun()
}

// runGUI starts the GUI version of the application
func runGUI() {
	gui := NewGUIApp()
	gui.Run()
}
