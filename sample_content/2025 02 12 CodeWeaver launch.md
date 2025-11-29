---
title: Launching CodeWeaver
description: |
    Continuing to open-source internal tools, I'm releasing CodeWeaver, a CLI to generate a single markdown file from a directory of source files. The idea is to facilitate sharing and feeding the codebase information to AI tools.
    Check out https://github.com/tesserato/CodeWeaver for more.
created: 2025 02 12
cover_image: codeweaver.gif
tags: projects, CodeWeaver, launch
---

from https://github.com/tesserato/CodeWeaver:

# CodeWeaver: Generate a Markdown Document of Your Codebase Structure and Content

CodeWeaver is a command-line tool designed to weave your codebase into a single, easy-to-navigate Markdown document. It recursively scans a directory, generating a structured representation of your project's file hierarchy and embedding the content of each file within code blocks. This tool simplifies codebase sharing, documentation, and integration with AI/ML code analysis tools by providing a consolidated and readable Markdown output.
The output for the current repository can be found [here](https://github.com/tesserato/CodeWeaver/blob/main/codebase.md).

# Key Features

* **Comprehensive Codebase Documentation:** Generates a Markdown file that meticulously outlines your project's directory and file structure in a clear, tree-like format.
* **Code Content Inclusion:** Embeds the complete content of each file directly within the Markdown document, enclosed in syntax-highlighted code blocks based on file extensions.
* **Flexible Path Filtering:**  Utilize regular expressions to define ignore patterns, allowing you to exclude specific files and directories from the generated documentation (e.g., `.git`, build artifacts, specific file types).
* **Optional Path Logging:** Choose to save lists of included and excluded file paths to separate files for detailed tracking and debugging of your ignore rules.
* **Simple Command-Line Interface:**  Offers an intuitive command-line interface with straightforward options for customization.

# Installation

If you have Go installed, run `go install github.com/tesserato/CodeWeaver@latest`to install the latest version of CodeWeaver or `go install github.com/tesserato/CodeWeaver@vX.Y.Z` to install a specific version.

Alternatively, download the appropriate pre built executable from the [releases page](https://github.com/tesserato/CodeWeaver/releases).

If necessary, make the `codeweaver` executable by using the `chmod` command:

```bash
chmod +x codeweaver
```

# Usage

## For help, run
```bash
codeweaver -h
```

## For actual usage, run
```bash
codeweaver [options]
```

**Options:**

| Option                            | Description                                                               | Default Value           |
| --------------------------------- | ------------------------------------------------------------------------- | ----------------------- |
| `-dir <directory>`                | The root directory to scan and document.                                  | Current directory (`.`) |
| `-output <filename>`              | The name of the output Markdown file.                                     | `codebase.md`           |
| `-ignore "<regex patterns>"`      | Comma-separated list of regular expression patterns for paths to exclude. | `\.git.*`               |
| `-included-paths-file <filename>` | File to save the list of paths that were included in the documentation.   | None                    |
| `-excluded-paths-file <filename>` | File to save the list of paths that were excluded from the documentation. | None                    |
| `-help`                           | Display this help message and exit.                                       |                         |

# Examples

## **Generate documentation for the current directory:**

   ```bash
   ./codeweaver
   ```
   This will create a file named `codebase.md` in the current directory, documenting the structure and content of the current directory and its subdirectories (excluding paths matching the default ignore pattern `\.git.*`).

## **Specify a different input directory and output file:**

   ```bash
   ./codeweaver -dir=my_project -output=project_docs.md
   ```
   This command will process the `my_project` directory and save the documentation to `project_docs.md`.

## **Ignore specific file types and directories:**

   ```bash
   ./codeweaver -ignore="\.log,temp,build" -output=detailed_docs.md
   ```
   This example will generate `detailed_docs.md`, excluding any files or directories with names containing `.log`, `temp`, or `build`. Regular expression patterns are comma-separated.

## **Save lists of included and excluded paths:**

   ```bash
   ./codeweaver -ignore="node_modules" -included-paths-file=included.txt -excluded-paths-file=excluded.txt -output=code_overview.md
   ```
   This command will create `code_overview.md` while also saving the list of included paths to `included.txt` and the list of excluded paths (due to the `node_modules` ignore pattern) to `excluded.txt`.

# Contributing

Contributions are welcome! If you encounter any issues, have suggestions for new features, or want to improve CodeWeaver, please feel free to open an issue or submit a pull request on the project's GitHub repository.

# License

CodeWeaver is released under the [MIT License](LICENSE). See the `LICENSE` file for complete license details.
