# git-llm-reviewer User Guide

## Introduction

`git-llm-reviewer` is a tool that uses Large Language Models (LLMs) to automatically review code changes in Git repositories. It helps identify potential issues, bugs, and improvements in your code before you commit it.

## Installation

### Quick Install

```bash
curl -sL https://raw.githubusercontent.com/mrheinen/git-llm-reviewer/main/install.sh | bash
```

### Manual Install

1. Download the binary for your platform from the [releases page](https://github.com/mrheinen/git-llm-reviewer/releases) (note: no releases published yet)
2. Extract the archive
3. Move the binary to a directory in your PATH (e.g., `/usr/local/bin`)
4. Make it executable

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
  # API key (required)
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

This is especially useful for:
- Debugging issues with LLM reviews
- Understanding what information is being sent to the LLM
- Manually fine-tuning prompts
- Copying prompts to test directly with LLM providers

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
      
      - name: Install git-llm-reviewer
        run: |
          curl -sL https://raw.githubusercontent.com/mrheinen/git-llm-reviewer/main/install.sh | bash
      
      - name: Run code review
        run: |
          git-llm-reviewer --all --output-dir review
        env:
          LLM_API_KEY: ${{ secrets.LLM_API_KEY }}
      
      - name: Upload review
        uses: actions/upload-artifact@v3
        with:
          name: code-review
          path: review
```

## Troubleshooting

### API Key Not Found

Make sure you have the API key correctly set in your configuration file, or set the appropriate environment variable:

- For OpenAI: `OPENAI_API_KEY`
- For Anthropic: `ANTHROPIC_API_KEY`

### Git Repository Not Found

`git-llm-reviewer` must be run from within a Git repository. Make sure you're in a Git repository directory.

### No Files to Review

If `git-llm-reviewer` reports "No files to review", make sure:

1. You have modified files in your repository
2. The files match the extensions defined in your configuration
3. If you're not using `--all`, make sure you've staged files with `git add`

## Support

For issues and feature requests, please file an issue on the [GitHub repository](https://github.com/mrheinen/git-llm-reviewer/issues).
