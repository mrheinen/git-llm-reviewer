# Changelog

All notable changes to git-llm-reviewer will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of git-llm-reviewer
- Git repository detection and validation
- Support for analyzing staged files with `git-llm-reviewer`
- Support for analyzing all changed files with `--all` flag
- Configuration management via YAML files
- Support for multiple LLM providers (OpenAI, Anthropic)
- Concurrent file processing with configurable limits
- Terminal output formatting with color support
- Markdown report generation
- Progress tracking during LLM reviews
- Installation script for easy deployment
- Cross-platform builds (Linux, macOS, Windows)
- Comprehensive documentation and user guide
- Initial project structure
- Basic CLI framework with cobra
- Logging setup

### Fixed
- Fixed deadlock in concurrent file processor
- Fixed handling of files with both staged and unstaged changes
- Fixed error handling in Git commands
