package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "git-llm-review-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Valid configuration file
	validConfigPath := filepath.Join(tempDir, "valid-config.yaml")
	validConfigContent := `
extensions:
  - .go
  - .js
  - .py
llm:
  provider: openai
  api_url: https://custom-api.example.com
  api_key: test-api-key
  model: gpt-4-turbo
  timeout: 600
concurrency:
  max_tasks: 10
`
	err = os.WriteFile(validConfigPath, []byte(validConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write valid config file: %v", err)
	}

	cfg, err := Load(validConfigPath)
	if err != nil {
		t.Fatalf("Failed to load valid config: %v", err)
	}

	// Verify the loaded configuration matches expected values
	expectedExtensions := []string{".go", ".js", ".py"}
	if !reflect.DeepEqual(cfg.Extensions, expectedExtensions) {
		t.Errorf("Expected extensions %v, got %v", expectedExtensions, cfg.Extensions)
	}

	if cfg.LLM.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", cfg.LLM.Provider)
	}

	if cfg.LLM.APIURL != "https://custom-api.example.com" {
		t.Errorf("Expected API URL 'https://custom-api.example.com', got '%s'", cfg.LLM.APIURL)
	}

	if cfg.LLM.APIKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", cfg.LLM.APIKey)
	}

	if cfg.LLM.Model != "gpt-4-turbo" {
		t.Errorf("Expected model 'gpt-4-turbo', got '%s'", cfg.LLM.Model)
	}

	if cfg.LLM.Timeout != 600 {
		t.Errorf("Expected timeout 600, got %d", cfg.LLM.Timeout)
	}

	if cfg.Concurrency.MaxTasks != 10 {
		t.Errorf("Expected max tasks 10, got %d", cfg.Concurrency.MaxTasks)
	}

	// Test case 2: Default values when settings are omitted
	minimalConfigPath := filepath.Join(tempDir, "minimal-config.yaml")
	minimalConfigContent := `
llm:
  provider: anthropic
  api_key: minimal-api-key
`
	err = os.WriteFile(minimalConfigPath, []byte(minimalConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write minimal config file: %v", err)
	}

	minimalCfg, err := Load(minimalConfigPath)
	if err != nil {
		t.Fatalf("Failed to load minimal config: %v", err)
	}

	// Check default extensions
	defaultExtensions := []string{".go", ".c", ".cc", ".proto", ".vue"}
	if !reflect.DeepEqual(minimalCfg.Extensions, defaultExtensions) {
		t.Errorf("Expected default extensions %v, got %v", defaultExtensions, minimalCfg.Extensions)
	}

	// Check provider was set correctly
	if minimalCfg.LLM.Provider != "anthropic" {
		t.Errorf("Expected provider 'anthropic', got '%s'", minimalCfg.LLM.Provider)
	}

	// Check API key was set correctly
	if minimalCfg.LLM.APIKey != "minimal-api-key" {
		t.Errorf("Expected API key 'minimal-api-key', got '%s'", minimalCfg.LLM.APIKey)
	}

	// Check default API URL
	if minimalCfg.LLM.APIURL != "https://api.openai.com/v1" {
		t.Errorf("Expected default API URL 'https://api.openai.com/v1', got '%s'", minimalCfg.LLM.APIURL)
	}

	// Check default model
	if minimalCfg.LLM.Model != "gpt-4" {
		t.Errorf("Expected default model 'gpt-4', got '%s'", minimalCfg.LLM.Model)
	}

	// Check default timeout
	if minimalCfg.LLM.Timeout != 300 {
		t.Errorf("Expected default timeout 300, got %d", minimalCfg.LLM.Timeout)
	}

	// Check default concurrency
	if minimalCfg.Concurrency.MaxTasks != 5 {
		t.Errorf("Expected default max tasks 5, got %d", minimalCfg.Concurrency.MaxTasks)
	}

	// Test case 3: Invalid configuration file
	invalidConfigPath := filepath.Join(tempDir, "invalid-config.yaml")
	invalidConfigContent := `
extensions:
  - .go
  - 
invalid yaml format
`
	err = os.WriteFile(invalidConfigPath, []byte(invalidConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	_, err = Load(invalidConfigPath)
	if err == nil {
		t.Errorf("Expected error when loading invalid config, got nil")
	}

	// Test case 4: Non-existent file
	nonExistentPath := filepath.Join(tempDir, "non-existent.yaml")
	_, err = Load(nonExistentPath)
	if err == nil {
		t.Errorf("Expected error when loading non-existent file, got nil")
	}
}

func TestLoadDefaultConfig(t *testing.T) {
	cfg := LoadDefault()

	// Check default extensions
	defaultExtensions := []string{".go", ".c", ".cc", ".proto", ".vue"}
	if !reflect.DeepEqual(cfg.Extensions, defaultExtensions) {
		t.Errorf("Expected default extensions %v, got %v", defaultExtensions, cfg.Extensions)
	}

	// Check default provider
	if cfg.LLM.Provider != "openai" {
		t.Errorf("Expected default provider 'openai', got '%s'", cfg.LLM.Provider)
	}

	// Check default API URL
	if cfg.LLM.APIURL != "https://api.openai.com/v1" {
		t.Errorf("Expected default API URL 'https://api.openai.com/v1', got '%s'", cfg.LLM.APIURL)
	}

	// Check default model
	if cfg.LLM.Model != "gpt-4" {
		t.Errorf("Expected default model 'gpt-4', got '%s'", cfg.LLM.Model)
	}

	// Check default timeout
	if cfg.LLM.Timeout != 300 {
		t.Errorf("Expected default timeout 300, got %d", cfg.LLM.Timeout)
	}

	// Check default concurrency
	if cfg.Concurrency.MaxTasks != 5 {
		t.Errorf("Expected default max tasks 5, got %d", cfg.Concurrency.MaxTasks)
	}
}
