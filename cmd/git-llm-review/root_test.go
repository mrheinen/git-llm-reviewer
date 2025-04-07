package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

// createMockDetector creates a mock repository detector for testing
func createMockDetector() *MockRepositoryDetector {
	return &MockRepositoryDetector{
		IsGitRepositoryFunc: func(dir string) (bool, error) {
			return true, nil
		},
		GetRepositoryRootFunc: func(dir string) (string, error) {
			return "/mock/repo/root", nil
		},
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := NewRootCmd()
	output, err := executeCommand(cmd, "--version")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !strings.Contains(output, "git-llm-review version 0.1.0") {
		t.Errorf("Expected version information, got: %s", output)
	}
}

func TestHelpFlag(t *testing.T) {
	cmd := NewRootCmd()
	output, err := executeCommand(cmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Check for required help content
	requiredContent := []string{
		"git-llm-review",
		"--config", 
		"--output", 
		"--debug", 
		"--all",
	}
	
	for _, content := range requiredContent {
		if !strings.Contains(output, content) {
			t.Errorf("Help output missing: %s", content)
		}
	}
}

func TestConfigFlag(t *testing.T) {
	mockDetector := createMockDetector()
	cmd := NewRootCmdWithDetector(mockDetector)
	configPath := "/custom/config/path.yaml"
	_, err := executeCommand(cmd, "--config", configPath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Check if the config path was set correctly
	if cmd.Flag("config").Value.String() != configPath {
		t.Errorf("Expected config path to be %s, got %s", configPath, cmd.Flag("config").Value.String())
	}
}

func TestOutputFlag(t *testing.T) {
	mockDetector := createMockDetector()
	cmd := NewRootCmdWithDetector(mockDetector)
	outputPath := "/custom/output/path.txt"
	_, err := executeCommand(cmd, "--output", outputPath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Check if the output path was set correctly
	if cmd.Flag("output").Value.String() != outputPath {
		t.Errorf("Expected output path to be %s, got %s", outputPath, cmd.Flag("output").Value.String())
	}
}

func TestDebugFlag(t *testing.T) {
	mockDetector := createMockDetector()
	cmd := NewRootCmdWithDetector(mockDetector)
	_, err := executeCommand(cmd, "--debug")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Check if debug flag was set
	if cmd.Flag("debug").Value.String() != "true" {
		t.Errorf("Expected debug to be true, got %s", cmd.Flag("debug").Value.String())
	}
}

func TestAllFlag(t *testing.T) {
	mockDetector := createMockDetector()
	cmd := NewRootCmdWithDetector(mockDetector)
	_, err := executeCommand(cmd, "--all")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Check if all flag was set
	if cmd.Flag("all").Value.String() != "true" {
		t.Errorf("Expected all to be true, got %s", cmd.Flag("all").Value.String())
	}
}

func TestShortFlagAliases(t *testing.T) {
	// Test short flags
	tests := []struct {
		shortFlag string
		longFlag  string
		value     string
	}{
		{"-c", "config", "/custom/config.yaml"},
		{"-o", "output", "/custom/output.txt"},
		{"-d", "debug", "true"},
		{"-a", "all", "true"},
	}
	
	for _, test := range tests {
		t.Run(test.shortFlag, func(t *testing.T) {
			mockDetector := createMockDetector()
			cmdInstance := NewRootCmdWithDetector(mockDetector)
			args := []string{test.shortFlag}
			if test.value != "true" {
				args = append(args, test.value)
			}
			
			_, err := executeCommand(cmdInstance, args...)
			if err != nil {
				t.Errorf("Unexpected error with %s: %v", test.shortFlag, err)
			}
			
			expected := test.value
			actual := cmdInstance.Flag(test.longFlag).Value.String()
			if actual != expected {
				t.Errorf("Expected %s to be %s, got %s", test.longFlag, expected, actual)
			}
		})
	}
}
