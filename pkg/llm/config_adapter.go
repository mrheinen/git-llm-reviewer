package llm

import (
	"github.com/niels/git-llm-review/pkg/config"
)

// CreateProviderFromConfig creates a provider from the configuration
func CreateProviderFromConfig(cfg *config.Config) (Provider, error) {
	// Convert config to map for provider factory
	configMap := map[string]interface{}{
		"api_key":  cfg.LLM.APIKey,
		"model":    cfg.LLM.Model,
		"api_url":  cfg.LLM.APIURL,
		"timeout":  cfg.LLM.Timeout,
		"config":   cfg, // Pass the whole config for additional settings
	}

	// Create provider using the factory
	return CreateProvider(cfg.LLM.Provider, configMap)
}
