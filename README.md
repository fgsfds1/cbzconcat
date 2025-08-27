# CBZ Concatenator

CBZ Concatenator is a Go-based command-line tool to merge multiple `.cbz` files into a single archive. It preserves image order, tries to use naturalsort to determine chapter order, extracts metadata from ComicInfo.xml if available, and generates a sanitized output filename. The tool is especially useful for manga or comic series split across multiple CBZ files.
Tested only on MangaDex archives (for now).

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
→ "Vol_01_Ch_0001_-anataha_doudesuka.cbz"


---

## Chapter Sorting

- Detects chapter numbers using patterns like `Ch0015`, `Ch0015.5`, `Ch0015.5.5`.
- Natural sorting ensures `Ch0015 < Ch0015.5 < Ch0016`.
- Filenames without detectable chapters are placed at the end.

---

## License

No formal license yet.