package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	"github.com/niels/git-llm-review/pkg/parse"
)

// ANSI color codes
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"

	// Bold colors
	ColorBoldRed     = "\033[1;31m"
	ColorBoldGreen   = "\033[1;32m"
	ColorBoldYellow  = "\033[1;33m"
	ColorBoldBlue    = "\033[1;34m"
	ColorBoldMagenta = "\033[1;35m"
	ColorBoldCyan    = "\033[1;36m"
	ColorBoldWhite   = "\033[1;37m"

	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// Default width for all terminal output elements
const DefaultWidth = 90

// TerminalFormatter formats review results for terminal output
type TerminalFormatter struct {
	useColor bool
	width    int // Consistent width for all elements
}

// NewTerminalFormatter creates a new terminal formatter
func NewTerminalFormatter(useColor bool) *TerminalFormatter {
	// Get terminal width if possible, otherwise use default
	width := DefaultWidth

	return &TerminalFormatter{
		useColor: useColor,
		width:    width,
	}
}

// FormatReview formats a review result for terminal output
func (f *TerminalFormatter) FormatReview(result *parse.ReviewResult) string {
	if result == nil || (len(result.Issues) == 0 && len(result.Diffs) == 0) {
		return f.colorizeText("No issues found.", ColorBoldGreen)
	}

	var sb strings.Builder

	// Simple header
	headerText := fmt.Sprintf("Found %d issues", len(result.Issues))
	sb.WriteString(f.colorizeText(headerText, ColorBoldBlue))
	sb.WriteString("\n\n")

	// Group issues by file for better organization
	issuesByFile := make(map[string][]parse.Issue)

	// Group issues
	for _, issue := range result.Issues {
		file := issue.File
		if file == "" {
			file = "General" // For issues not associated with any specific file
		}
		issuesByFile[file] = append(issuesByFile[file], issue)
	}

	// Display issues grouped by file
	for file, issues := range issuesByFile {
		// File header
		fileHeader := fmt.Sprintf("File: %s", file)
		sb.WriteString(f.colorizeText(fileHeader, ColorBoldMagenta))
		sb.WriteString("\n")
		sb.WriteString(f.colorizeText("----------------------------------------", ColorMagenta))
		sb.WriteString("\n\n")

		// Display each issue
		for i, issue := range issues {
			// Issue title
			issueTitle := fmt.Sprintf("Issue %d: %s", i+1, issue.Title)
			sb.WriteString(f.colorizeText(issueTitle, ColorBoldCyan))
			sb.WriteString("\n\n")

			// Explanation
			//sb.WriteString(f.colorizeText("Explanation:", ColorBoldYellow))
			//sb.WriteString("\n")
			sb.WriteString(issue.Explanation)
			sb.WriteString("\n\n")
		}

		sb.WriteString("\n")
	}

	// Now add the consolidated diffs if any
	if len(result.Diffs) > 0 {
		diffHeader := "Consolidated Changes"
		sb.WriteString(f.colorizeText(diffHeader, ColorBoldBlue))
		sb.WriteString("\n")
		sb.WriteString(f.colorizeText("========================================", ColorBlue))
		sb.WriteString("\n\n")

		for _, diff := range result.Diffs {
			// File header for the diff
			fileHeader := fmt.Sprintf("File: %s", diff.File)
			sb.WriteString(f.colorizeText(fileHeader, ColorBoldGreen))
			sb.WriteString("\n")
			sb.WriteString(f.colorizeText("----------------------------------------", ColorGreen))
			sb.WriteString("\n\n")

			if diff.Diff != "" {
				sb.WriteString(f.colorizeText("Suggested changes:", ColorBoldBlue))
				sb.WriteString("\n")

				// The diff will be highlighted directly to the terminal when printed
				// We use a special marker that will be replaced with the highlighted diff
				sb.WriteString("<<HIGHLIGHTED_DIFF:" + diff.File + ">>\n\n")
			}
		}
	}

	// Handle backward compatibility for issues with individual diffs
	for _, issue := range result.Issues {
		if issue.Diff != "" {
			// This handles old-style issues with individual diffs
			// We'll only show these if there are no consolidated diffs for the same file

			// Check if we already have a consolidated diff for this file
			hasDiff := false
			for _, diff := range result.Diffs {
				if diff.File == issue.File {
					hasDiff = true
					break
				}
			}

			if !hasDiff {
				fileHeader := fmt.Sprintf("Legacy diff for issue: %s", issue.Title)
				sb.WriteString(f.colorizeText(fileHeader, ColorBoldYellow))
				sb.WriteString("\n")
				sb.WriteString(f.colorizeText("----------------------------------------", ColorYellow))
				sb.WriteString("\n\n")

				sb.WriteString(f.colorizeText("Suggested changes:", ColorBoldBlue))
				sb.WriteString("\n")
				sb.WriteString("<<HIGHLIGHTED_DIFF_LEGACY>>\n\n")
			}
		}
	}

	return sb.String()
}

// HighlightDiff highlights a diff using Chroma with the terminal16m formatter
func (f *TerminalFormatter) HighlightDiff(diff string, filePath string) {
	if diff == "" {
		return
	}

	// Detect language from the file extension
	language := "diff" // Default to diff lexer
	if filePath != "" {
		ext := filepath.Ext(filePath)
		if ext != "" {
			// Try to match a lexer for this extension
			lexer := lexers.Match(filePath)
			if lexer != nil {
				language = lexer.Config().Name
			}
		}
	}

	// Highlight the diff and output to stdout without any framing
	err := quick.Highlight(os.Stdout, diff, language, "terminal16m", "monokai")
	if err != nil {
		// Fallback to simple ANSI coloring if Chroma highlighting fails
		fmt.Println(f.simpleColorizeDiff(diff))
	}
}

// formatIssueType extracts and formats the issue type from the title
func (f *TerminalFormatter) formatIssueType(title string) string {
	// Check if the title contains a prefix like "Security:", "Performance:", etc.
	possibleTypes := map[string]string{
		"bug":           ColorBoldRed,
		"security":      ColorBoldRed,
		"perf":          ColorBoldYellow,
		"performance":   ColorBoldYellow,
		"style":         ColorBoldCyan,
		"unused":        ColorBoldCyan,
		"doc":           ColorBoldGreen,
		"documentation": ColorBoldGreen,
		"maintain":      ColorBoldBlue,
		"maintenance":   ColorBoldBlue,
		"refactor":      ColorBoldMagenta,
	}

	titleLower := strings.ToLower(title)
	for typeName, color := range possibleTypes {
		if strings.HasPrefix(titleLower, typeName+":") || strings.HasPrefix(titleLower, "["+typeName+"]") {
			return color
		}
	}

	// If no specific type is found, try to determine by keywords
	for typeName, color := range possibleTypes {
		if strings.Contains(titleLower, typeName) {
			return color
		}
	}

	// Default color for unclassified issues
	return ColorBoldWhite
}

// colorizeText adds color to text if color is enabled
func (f *TerminalFormatter) colorizeText(text string, colorCode string) string {
	if !f.useColor || colorCode == "" {
		return text
	}
	return fmt.Sprintf("%s%s%s", colorCode, text, ColorReset)
}

// simpleColorizeDiff adds basic color to diff lines as a fallback
func (f *TerminalFormatter) simpleColorizeDiff(diff string) string {
	if diff == "" {
		return ""
	}

	lines := strings.Split(diff, "\n")
	var coloredLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "+") {
			coloredLines = append(coloredLines, f.colorizeText(line, ColorGreen))
		} else if strings.HasPrefix(line, "-") {
			coloredLines = append(coloredLines, f.colorizeText(line, ColorRed))
		} else {
			coloredLines = append(coloredLines, line)
		}
	}

	return strings.Join(coloredLines, "\n")
}

// indentText indents each line of text by the specified number of spaces
func (f *TerminalFormatter) indentText(text string, spaces int) string {
	if text == "" {
		return ""
	}

	indent := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		lines[i] = indent + line
	}

	return strings.Join(lines, "\n")
}
