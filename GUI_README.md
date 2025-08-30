# CBZ Concat GUI

This project now includes a simple graphical user interface built with Fyne for the CBZ concatenation tool.

## Features

- **Directory Selection**: Browse and select input and output directories
- **Progress Tracking**: Real-time progress bar and status updates
- **Error Handling**: User-friendly error dialogs
- **Options**: Checkboxes for various concatenation options
- **Cross-platform**: Works on Windows, macOS, and Linux

## Usage

### GUI Mode
```bash
./cbzconcat -gui
```

### Command Line Mode (existing)
```bash
./cbzconcat [flags] <input_dir> <output_dir>
```

## GUI Interface

The GUI provides a simple interface with:

1. **Input Directory**: Select the folder containing your CBZ files
2. **Output Directory**: Choose where to save the concatenated file
3. **Options**: 
   - Show XML output
   - Print file order
   - Silent mode
   - Verbose output
4. **Progress**: Real-time progress bar and status updates
5. **Actions**: 
   - Concatenate CBZ Files
   - Clear Fields

## Building

The GUI requires the Fyne toolkit. Dependencies are automatically managed via Go modules:

```bash
go mod tidy
go build .
```

## Requirements

- Go 1.16 or later
- Fyne v2 toolkit
- OpenGL support (usually included with graphics drivers)

## Notes

- The GUI runs the same core logic as the command-line version
- Progress updates are shown in real-time
- Error messages are displayed in user-friendly dialogs
- The interface is responsive and won't freeze during processing 