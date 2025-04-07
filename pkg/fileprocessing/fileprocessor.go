package fileprocessing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode/utf16"
	"unicode/utf8"
)

// Common errors
var (
	ErrInvalidEncoding = errors.New("invalid or unsupported file encoding")
	ErrEmptyFile       = errors.New("file is empty")
)

// ReadFileContent reads the content of a file and returns it as a string.
// It handles different encodings (UTF-8, UTF-16LE, UTF-16BE) and returns
// the content as a UTF-8 string suitable for LLM processing.
func ReadFileContent(filePath string) (string, error) {
	// Convert to absolute path if it's a relative path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if file exists and is readable
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to access file: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", absPath)
	}

	// Check if file is empty
	if info.Size() == 0 {
		return "", ErrEmptyFile
	}

	// Read file content
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Detect and handle encoding
	return detectAndHandleEncoding(data)
}

// detectAndHandleEncoding detects the encoding of the file content
// and converts it to a UTF-8 string.
func detectAndHandleEncoding(data []byte) (string, error) {
	// Check for empty data
	if len(data) == 0 {
		return "", ErrEmptyFile
	}

	// Check for BOM (Byte Order Mark) to detect encoding
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		// UTF-8 with BOM
		return string(data[3:]), nil
	} else if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		// UTF-16LE with BOM
		return decodeUTF16LE(data[2:]), nil
	} else if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		// UTF-16BE with BOM
		return decodeUTF16BE(data[2:]), nil
	}

	// No BOM detected, try to determine encoding by content
	if isUTF8(data) {
		// UTF-8 without BOM
		return string(data), nil
	}

	// Try UTF-16LE without BOM
	if len(data)%2 == 0 && looksLikeUTF16LE(data) {
		return decodeUTF16LE(data), nil
	}

	// Try UTF-16BE without BOM
	if len(data)%2 == 0 && looksLikeUTF16BE(data) {
		return decodeUTF16BE(data), nil
	}

	// Default to UTF-8 if we can't determine the encoding
	// This is a reasonable default for most text files
	return string(data), nil
}

// isUTF8 checks if the data is valid UTF-8
func isUTF8(data []byte) bool {
	return utf8.Valid(data)
}

// looksLikeUTF16LE checks if the data looks like UTF-16LE
// by checking for null bytes in even positions
func looksLikeUTF16LE(data []byte) bool {
	// Simple heuristic: in UTF-16LE, ASCII characters have a null byte
	// after each character byte
	nullByteCount := 0
	for i := 1; i < len(data); i += 2 {
		if data[i] == 0 {
			nullByteCount++
		}
	}
	// If more than 30% of even positions have null bytes, it's likely UTF-16LE
	return nullByteCount > len(data)/2/3
}

// looksLikeUTF16BE checks if the data looks like UTF-16BE
// by checking for null bytes in odd positions
func looksLikeUTF16BE(data []byte) bool {
	// Simple heuristic: in UTF-16BE, ASCII characters have a null byte
	// before each character byte
	nullByteCount := 0
	for i := 0; i < len(data); i += 2 {
		if data[i] == 0 {
			nullByteCount++
		}
	}
	// If more than 30% of odd positions have null bytes, it's likely UTF-16BE
	return nullByteCount > len(data)/2/3
}

// decodeUTF16LE decodes UTF-16LE data to a UTF-8 string
func decodeUTF16LE(data []byte) string {
	// Ensure we have an even number of bytes
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}

	// Convert bytes to uint16 code points
	u16s := make([]uint16, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		u16s = append(u16s, uint16(data[i])|uint16(data[i+1])<<8)
	}

	// Convert UTF-16 code points to UTF-8 string
	return string(utf16.Decode(u16s))
}

// decodeUTF16BE decodes UTF-16BE data to a UTF-8 string
func decodeUTF16BE(data []byte) string {
	// Ensure we have an even number of bytes
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}

	// Convert bytes to uint16 code points
	u16s := make([]uint16, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		u16s = append(u16s, uint16(data[i])<<8|uint16(data[i+1]))
	}

	// Convert UTF-16 code points to UTF-8 string
	return string(utf16.Decode(u16s))
}

// ReadFileContentWithLimit reads the content of a file up to a specified limit
// and returns it as a string. This is useful for large files where you only
// need the first N bytes.
func ReadFileContentWithLimit(filePath string, maxBytes int64) (string, error) {
	// Convert to absolute path if it's a relative path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Open the file
	file, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", absPath)
	}

	// Check if file is empty
	if info.Size() == 0 {
		return "", ErrEmptyFile
	}

	// Determine how many bytes to read
	bytesToRead := info.Size()
	if maxBytes > 0 && bytesToRead > maxBytes {
		bytesToRead = maxBytes
	}

	// Read file content with limit
	data := make([]byte, bytesToRead)
	_, err = io.ReadFull(file, data)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Detect and handle encoding
	return detectAndHandleEncoding(data)
}

// IsTextFile checks if a file is likely to be a text file
// by examining its content for binary data
func IsTextFile(filePath string) (bool, error) {
	// Read a sample of the file (first 8KB should be enough)
	data, err := ReadFileContentWithLimit(filePath, 8*1024)
	if err != nil {
		if errors.Is(err, ErrEmptyFile) {
			// Empty files are considered text files
			return true, nil
		}
		return false, err
	}

	// Check for null bytes which are common in binary files
	// but rare in text files
	if bytes.Contains([]byte(data), []byte{0}) {
		return false, nil
	}

	// Check if the content is valid UTF-8
	return utf8.ValidString(data), nil
}
