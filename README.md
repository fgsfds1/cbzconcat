# cbzconcat

cbzconcat is a Go-based command-line tool to merge multiple `.cbz` files into a single archive.

It preserves image order, uses natural sorting to determine chapter order, extracts metadata from ComicInfo.xml if available, and generates a sanitized output filename.

The tool should be especially useful for manga or comic series split across multiple CBZ files.

Tested only on MangaDex archives (for now).

---

## TODO

[ ] Modify the chapter info struct, include volumes
[ ] Volume search in name
[ ] Compare using the volumes
[ ] Mixed comparison logic
[ ] Figure out the command-action stuff
[ ] Output version in help
[ ] Builds that get passed the version
[ ] Prune action
[ ] Resize action
[ ] Meta-edit action

---

## Features

- Merge multiple CBZ archives into one.
- Natural chapter sorting (`Ch0015`, `Ch0015.5`, `Ch0015.5.5`, etc.).
- Preserves only image files (`.jpg`, `.jpeg`, `.png`, `.gif`) from source CBZs.
- Generates a new `ComicInfo.xml` in the merged archive.
- Sanitizes output filenames for cross-platform compatibility.
- ASCII transliteration of filenames.

---

## Installation

1. Clone the repository:

```
git clone https://github.com/yourusername/cbzconcat.git
cd cbzconcat
```


2. Initialize modules and install dependencies:

```
go mod tidy
```


3. Build the binary:

```
go build -o cbzconcat
```


---

## Usage

```
cbzconcat [flags] <input_dir> <output_dir>
```

- `<input_dir>`: Directory containing CBZ files to merge.
- `<output_dir>`: Directory where the merged CBZ will be created.

### Flags

- `-v` : Verbose output (overrides silent mode).
- `-s` : Silent mode; suppress stdout output except for errors.
- `-r` : Print the order of input CBZ files before merging.
- `-x` : Print the resulting `ComicInfo.xml` content.

---

## Example

Merge a folder of chapters into one CBZ in the current directory, with the name extracted from the first chapter and sanitized; with verbose output:

```
cbzconcat -v ./Elf-san\ wa\ Yaserarenai .
```

This will produce a `./Elf-san_wa_Yaserarenai_Ch_0000-0047_6.cbz` file, with 0000-0047.6 being the chapters read from ComicInfo.xml from the first and the last chapters.

---

## Filename Sanitization

The tool replaces spaces and dots with underscores, removes filesystem-invalid characters, and can transliterate non-Latin characters to ASCII using [mozillazg/go-unidecode](https://github.com/mozillazg/go-unidecode).

Example:

"Vol.01 Ch.0001 - あなたはどうですか?.cbz"
→ "Vol_01_Ch_0001_-_anataha_doudesuka.cbz"

---

## Chapter Sorting

### Chapter Detection
The tool uses sophisticated regex patterns to extract chapter numbers:

1. **Primary pattern**: Matches `Ch`, `Chap`, or `Chapter` followed by optional separators and numbers
   - Examples: `Ch0015`, `Ch-0015.5`, `Ch_0015.5.5`, `chapter 0015`
   - Supports various separators: space, dash, underscore, dot, colon, semicolon, comma, exclamation, question mark, tab, newline

2. **Fallback pattern**: If no chapter prefix is found, looks for any 3+ digit number
   - This helps with files that have chapter numbers without explicit "Ch" prefixes
   - 3+ digits are used to avoid matching volume numbers (which are typically 1-2 digits)

### Natural Sorting
- Splits chapter numbers by decimal points and compares each part numerically
- Ensures `Ch0015 < Ch0015.5 < Ch0016` and `Ch0015.5 < Ch0015.6`
- Files without detectable chapters are placed at the end
- Falls back to string comparison when no chapters are found

### Current Limitations
- **Volume handling**: Currently only extracts chapter numbers, not volume information
- **Mixed formats**: Files with both volume and chapter numbers may not sort optimally
- **Error handling**: Uses panic() for critical errors (will exit the program)

---

## License

This project is licensed under the GNU General Public License v3 (GPLv3). See the [LICENSE](LICENSE) file for the full license text.

The GPLv3 is a free software license that ensures the software remains free and open source. You are free to use, modify, and distribute this software under the terms of the GPLv3, with the requirement that any derivative works must also be licensed under the GPLv3.