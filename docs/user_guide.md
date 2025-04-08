# git-llm-reviewer User Guide

## Introduction

`git-llm-reviewer` is a tool that uses Large Language Models (LLMs) to automatically review code changes in Git repositories. It helps identify potential issues, bugs, and improvements in your code before you commit it.

## Installation

### Building from Source

#### Prerequisites

- Go 1.20 or later
- Git
- Make (for using the Makefile)

#### Steps

```bash
# Clone the repository
git clone https://github.com/mrheinen/git-llm-reviewer.git
cd git-llm-reviewer

# Build for your current platform
make build

# Install locally (moves binary to /usr/local/bin)
make install
```

After installation, verify that it's working:

```bash
git-llm-reviewer --version
```

#### Optional: Build for Different Platforms

```bash
# Build for specific platforms
make build-linux-amd64
make build-darwin-arm64
make build-windows-amd64

# Build for all supported platforms
make build-all
```

## Configuration

`git-llm-reviewer` looks for a configuration file in the following locations (in order):

1. Path specified by the `--config` flag
2. `.git-llm-reviewer.yaml` in the current directory
3. `.git-llm-reviewer.yaml` in your home directory

Here's a sample configuration file:

```yaml
# File extensions to analyze
extensions:
  - .go
  - .js
  - .py
  - .java
  - .c
  - .cpp
  - .h
  - .ts
  - .tsx
  - .jsx

# LLM provider settings
llm:
  # Provider (openai or anthropic)
  provider: openai
  # API URL (optional, defaults to the provider's standard API URL)
  apiURL: https://api.openai.com/v1
  # API key (required, can be overridden by LLM_API_KEY environment variable)
  apiKey: your-api-key-here
  # Model to use
  model: gpt-4
  # Timeout in seconds
  timeout: 30

# Concurrency settings
concurrency:
  # Maximum number of files to process concurrently
  maxTasks: 4
```

## Basic Usage

### Review Staged Changes

By default, `git-llm-reviewer` analyzes files that have been staged in git:

```bash
git add file.go
git-llm-reviewer
```

### Review All Changes

To review all changed files (both staged and unstaged):

```bash
git-llm-reviewer --all
```

### Generate a Markdown Report

To save the review results to a directory containing individual markdown files for each reviewed file, along with a summary report:

```bash
git-llm-reviewer --output-dir reports
```

This will create a directory called `reports` and generate individual markdown files for each reviewed file, along with a summary report.

## Advanced Usage

### Override LLM Provider

```bash
git-llm-reviewer --provider anthropic
```

### Enable Debug Mode

```bash
git-llm-reviewer --debug
```

### Verbose Output

```bash
git-llm-reviewer --verbose
```

### Debug Prompt Generation

If you need to debug how git-llm-reviewer is generating prompts for the LLM, you can use the `--log-prompts` flag:

```bash
git-llm-reviewer --log-prompts
```

This will save all prompts sent to the LLM provider to a file called `prompt.log`. Each prompt entry includes:
- Timestamp
- LLM provider name
- File path being reviewed
- Full prompt content

### Debug Full Exchange

For more comprehensive debugging, you can use the `--log-full-exchange` flag to log both prompts and raw LLM responses:

```bash
git-llm-reviewer --log-full-exchange
```

This will save the complete exchange (both prompts and raw responses) to a file called `exchange.log`. Each entry includes:
- Timestamp
- LLM provider name
- File path being reviewed
- Full prompt content
- Complete raw response from the LLM

These logging options are especially useful for:
- Debugging issues with LLM reviews
- Understanding what information is being sent to and received from the LLM
- Manually fine-tuning prompts
- Copying prompts to test directly with LLM providers
- Analyzing raw LLM responses to improve parsing

## Workflow Integration

### Pre-commit Hook

You can set up `git-llm-reviewer` as a pre-commit hook to automatically review your code before committing:

1. Create `.git/hooks/pre-commit` in your repository:

```bash
#!/bin/bash
set -e

# Run git-llm-reviewer
git-llm-reviewer

# Ask for confirmation if issues were found
read -p "Proceed with commit? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi
```

2. Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

### CI Integration

Add `git-llm-reviewer` to your CI pipeline to automatically review changes in pull requests.

Example GitHub Actions workflow:

```yaml
name: Code Review

on:
  pull_request:
    branches: [ main ]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.20'
      
      - name: Build git-llm-reviewer from source
        run: |
          make build
          sudo make install
      
      - name: Run code review
        run: |
          git-llm-reviewer --all --output-dir review
        env:
          LLM_API_KEY: ${{ secrets.LLM_API_KEY }} # This overrides any API key in the config file
      
      - name: Upload review
        uses: actions/upload-artifact@v3
        with:
          name: code-review
          path: review
```

## Troubleshooting

### API Key Not Found

Make sure you have the API key correctly set in your configuration file, or set the `LLM_API_KEY` environment variable.

Using the `LLM_API_KEY` environment variable will override any API key set in the configuration file, regardless of the provider you're using.

### Git Repository Not Found

`git-llm-reviewer` must be run from within a Git repository. Make sure you're in a Git repository directory.

### No Files to Review

If `git-llm-reviewer` reports "No files to review", make sure:

1. You have modified files in your repository
2. The files match the extensions defined in your configuration
3. If you're not using `--all`, make sure you've staged files with `git add`

## Support

For issues and feature requests, please file an issue on the [GitHub repository](https://github.com/mrheinen/git-llm-reviewer/issues).
