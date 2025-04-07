# git-llm-review

A tool that uses LLMs to review code changes in Git repositories.

## Overview

`git-llm-review` is a command-line tool that leverages Large Language Models (LLMs) to automatically review code changes in Git repositories. It helps developers identify potential issues, suggest improvements, and ensure code quality.

## Features

- Automatically reviews Git staged changes or all changes
- Supports multiple LLM providers (OpenAI, Anthropic)
- Provides detailed feedback on code quality, bugs, and improvements
- Generates markdown reports for easy sharing and reference
- Concurrent processing for faster reviews of multiple files

## Installation

### Option 1: Using the installation script

```bash
# Download and run the installation script
curl -sL https://raw.githubusercontent.com/niels/git-llm-review/main/install.sh | bash
```

### Option 2: Pre-built binaries

1. Download the latest release for your platform from the [Releases](https://github.com/niels/git-llm-review/releases) page.
2. Extract the archive: `unzip git-llm-review-*.zip`
3. Move the binary to a directory in your PATH: `mv git-llm-review-*/git-llm-review /usr/local/bin/`
4. Make it executable: `chmod +x /usr/local/bin/git-llm-review`

### Option 3: Using Go

```bash
go install github.com/niels/git-llm-review/cmd/git-llm-review@latest
```

## Configuration

Create a configuration file at `.git-llm-review.yaml` in your home directory or project root:

```yaml
extensions:
  - .go
  - .js
  - .py
  - .java
llm:
  provider: openai  # or anthropic
  apiKey: your-api-key
  model: gpt-4
  timeout: 30
concurrency:
  maxTasks: 4
```

## Usage

```bash
# Review staged changes
git-llm-review

# Review all changes (both staged and unstaged)
git-llm-review --all

# Generate a markdown report
git-llm-review --output-dir reports

# Use verbose output
git-llm-review --verbose

# Override LLM provider
git-llm-review --provider anthropic

# Show version information
git-llm-review --version

# Log prompts for debugging
git-llm-review --log-prompts
```

### Flags

- `-v, --version`: Display version information
- `-h, --help`: Show help information
- `-c, --config string`: Specify configuration file path (default: ".git-llm-review.yaml")
- `-o, --output-dir string`: Directory to write review reports
- `-d, --debug`: Enable debug mode
- `-a, --all`: Review all changed files (both staged and unstaged)
- `-p, --provider string`: LLM provider to use (overrides config)
- `--verbose`: Enable verbose output
- `-x, --log-prompts`: Log prompts to prompt.log for debugging

## Building from Source

### Prerequisites

- Go 1.20 or later
- Git
- Make (for using the Makefile)

### Building

```bash
# Clone the repository
git clone https://github.com/niels/git-llm-review.git
cd git-llm-review

# Build for your current platform
make build

# Run tests
make test

# Build for all supported platforms
make build-all

# Create release packages
make release

# Install locally
make install
```

### Cross-Compilation

The Makefile supports cross-compilation for multiple platforms:

```bash
# Build for specific platform
make build-linux-amd64
make build-darwin-arm64
make build-windows-amd64
```

## Versioning

`git-llm-review` uses semantic versioning and Git tags for versioning. The build process automatically embeds version information from Git tags into the binary.

## Development

### Project Structure

```
.
├── cmd/git-llm-review/    # Main application entry point
├── internal/              # Internal packages
├── pkg/                   # Public packages
│   ├── config/            # Configuration handling
│   ├── git/               # Git repository interactions
│   ├── llm/               # LLM provider implementations
│   ├── output/            # Output formatting
│   ├── parse/             # Response parsing
│   ├── processor/         # File processing
│   ├── progress/          # Progress tracking
│   ├── prompt/            # Prompt generation
│   ├── version/           # Version information
│   └── workflow/          # Main application workflow
├── Makefile               # Build automation
├── install.sh             # Installation script
└── README.md              # Documentation
```

### Running tests

```bash
# Run all tests
go test ./...

# Run tests with coverage report
make test-coverage
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
