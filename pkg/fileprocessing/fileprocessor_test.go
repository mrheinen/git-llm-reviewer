package fileprocessing

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileContent(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "fileprocessing-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Successfully read a UTF-8 encoded file
	utf8Content := "This is a test file with UTF-8 content.\nLine 2\nLine 3"
	utf8FilePath := filepath.Join(tempDir, "test-utf8.txt")
	err = os.WriteFile(utf8FilePath, []byte(utf8Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test UTF-8 file: %v", err)
	}

	// Test with absolute path
	content, err := ReadFileContent(utf8FilePath)
	if err != nil {
		t.Errorf("Expected no error reading UTF-8 file, got: %v", err)
	}
	if content != utf8Content {
		t.Errorf("Expected content '%s', got: '%s'", utf8Content, content)
	}

	// Test case 2: Successfully read a file with relative path
	// Change to the temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Read using relative path
	relativeContent, err := ReadFileContent(filepath.Base(utf8FilePath))
	if err != nil {
		t.Errorf("Expected no error reading file with relative path, got: %v", err)
	}
	if relativeContent != utf8Content {
		t.Errorf("Expected content '%s', got: '%s'", utf8Content, relativeContent)
	}

	// Test case 3: File with UTF-16 encoding (simulated)
	// For simplicity, we'll create a UTF-8 file but handle it as if it were UTF-16
	utf16FilePath := filepath.Join(tempDir, "test-utf16.txt")
	// UTF-16LE BOM (0xFF 0xFE) followed by "Hello" in UTF-16LE
	utf16Data := []byte{0xFF, 0xFE, 'H', 0, 'e', 0, 'l', 0, 'l', 0, 'o', 0}
	err = os.WriteFile(utf16FilePath, utf16Data, 0644)
	if err != nil {
		t.Fatalf("Failed to create test UTF-16 file: %v", err)
	}

	utf16Content, err := ReadFileContent(utf16FilePath)
	if err != nil {
		t.Errorf("Expected no error reading UTF-16 file, got: %v", err)
	}
	// The expected result should be "Hello" after encoding detection and conversion
	if utf16Content != "Hello" {
		t.Errorf("Expected UTF-16 content to be converted to 'Hello', got: '%s'", utf16Content)
	}

	// Test case 4: Error handling - file not found
	_, err = ReadFileContent(filepath.Join(tempDir, "nonexistent-file.txt"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Expected 'file not found' error, got: %v", err)
	}

	// Test case 5: Error handling - permission denied
	// Create a file with no read permissions
	noPermFilePath := filepath.Join(tempDir, "no-permission.txt")
	err = os.WriteFile(noPermFilePath, []byte("Can't read this"), 0000)
	if err != nil {
		t.Fatalf("Failed to create no-permission file: %v", err)
	}

	_, err = ReadFileContent(noPermFilePath)
	if err == nil {
		t.Error("Expected error when reading file with no permissions")
	}

	// Test case 6: Large file handling
	largeFilePath := filepath.Join(tempDir, "large-file.txt")
	largeContent := make([]byte, 1024*1024) // 1MB of data
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	err = os.WriteFile(largeFilePath, largeContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	_, err = ReadFileContent(largeFilePath)
	if err != nil {
		t.Errorf("Expected no error reading large file, got: %v", err)
	}
}

func TestDetectAndHandleEncoding(t *testing.T) {
	// Test various byte order marks (BOMs) and encoding detection
	testCases := []struct {
		name           string
		input          []byte
		expectedOutput string
	}{
		{
			name:           "UTF-8 without BOM",
			input:          []byte("Hello, world!"),
			expectedOutput: "Hello, world!",
		},
		{
			name:           "UTF-8 with BOM",
			input:          []byte{0xEF, 0xBB, 0xBF, 'H', 'e', 'l', 'l', 'o'},
			expectedOutput: "Hello",
		},
		{
			name:           "UTF-16LE with BOM",
			input:          []byte{0xFF, 0xFE, 'H', 0, 'e', 0, 'l', 0, 'l', 0, 'o', 0},
			expectedOutput: "Hello",
		},
		{
			name:           "UTF-16BE with BOM",
			input:          []byte{0xFE, 0xFF, 0, 'H', 0, 'e', 0, 'l', 0, 'l', 0, 'o'},
			expectedOutput: "Hello",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := detectAndHandleEncoding(tc.input)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if output != tc.expectedOutput {
				t.Errorf("Expected '%s', got: '%s'", tc.expectedOutput, output)
			}
		})
	}
}
