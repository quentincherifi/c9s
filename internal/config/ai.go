// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"

	"github.com/derailed/k9s/internal/ai"
)

// AI tracks AI configuration options.
type AI struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Provider  string `json:"provider,omitempty" yaml:"provider,omitempty"`
	APIKey    string `json:"apiKey,omitempty" yaml:"apiKey,omitempty"`
	APIKeyEnv string `json:"apiKeyEnv,omitempty" yaml:"apiKeyEnv,omitempty"`
	Model     string `json:"model,omitempty" yaml:"model,omitempty"`
	MaxTokens int    `json:"maxTokens,omitempty" yaml:"maxTokens,omitempty"`
	BaseURL   string `json:"baseURL,omitempty" yaml:"baseURL,omitempty"`
}

// NewAI creates a new AI configuration with defaults.
func NewAI() AI {
	return AI{
		Enabled: false,
	}
}

// GetProvider returns the provider type, defaulting to Claude.
func (a *AI) GetProvider() ai.ProviderType {
	switch a.Provider {
	case "openai", "gpt", "chatgpt":
		return ai.ProviderOpenAI
	case "ollama", "local":
		return ai.ProviderOllama
	default:
		return ai.ProviderClaude
	}
}

// GetAPIKey returns the API key, checking config first, then env var.
func (a *AI) GetAPIKey() string {
	if a.APIKey != "" {
		return a.APIKey
	}
	if a.APIKeyEnv != "" {
		return os.Getenv(a.APIKeyEnv)
	}

	// Check provider-specific env vars
	switch a.GetProvider() {
	case ai.ProviderOpenAI:
		return os.Getenv("OPENAI_API_KEY")
	case ai.ProviderClaude:
		return os.Getenv("ANTHROPIC_API_KEY")
	case ai.ProviderOllama:
		return "" // No API key needed for Ollama
	}

	return ""
}

// GetModel returns the model to use, defaulting based on provider.
func (a *AI) GetModel() string {
	if a.Model != "" {
		return a.Model
	}

	switch a.GetProvider() {
	case ai.ProviderOpenAI:
		return ai.DefaultOpenAIModel
	case ai.ProviderOllama:
		return ai.DefaultOllamaModel
	default:
		return ai.DefaultClaudeModel
	}
}

// GetMaxTokens returns the max tokens, defaulting if not set.
func (a *AI) GetMaxTokens() int {
	if a.MaxTokens > 0 {
		return a.MaxTokens
	}
	return ai.DefaultMaxTokens
}

// GetBaseURL returns the base URL for custom endpoints.
func (a *AI) GetBaseURL() string {
	return a.BaseURL
}

// CreateProvider creates an AI provider based on the configuration.
func (a *AI) CreateProvider() (ai.Provider, error) {
	return ai.NewProvider(ai.ProviderConfig{
		Type:      a.GetProvider(),
		APIKey:    a.GetAPIKey(),
		Model:     a.GetModel(),
		MaxTokens: a.GetMaxTokens(),
		BaseURL:   a.GetBaseURL(),
	})
}
