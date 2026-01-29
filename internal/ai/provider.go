// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ai

import "fmt"

// ProviderType represents the AI provider type.
type ProviderType string

const (
	// ProviderClaude represents Anthropic's Claude.
	ProviderClaude ProviderType = "claude"
	// ProviderOpenAI represents OpenAI's GPT models.
	ProviderOpenAI ProviderType = "openai"
	// ProviderOllama represents local Ollama models.
	ProviderOllama ProviderType = "ollama"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response represents an AI response.
type Response struct {
	Content      string
	Model        string
	InputTokens  int
	OutputTokens int
}

// Provider is the interface for AI providers.
type Provider interface {
	// Name returns the provider name.
	Name() string
	// Send sends messages to the AI and returns a response.
	Send(system string, messages []Message) (*Response, error)
}

// ProviderConfig holds configuration for creating a provider.
type ProviderConfig struct {
	Type      ProviderType
	APIKey    string
	Model     string
	MaxTokens int
	BaseURL   string // For custom endpoints (Ollama, Azure, etc.)
}

// NewProvider creates a new AI provider based on the config.
func NewProvider(cfg ProviderConfig) (Provider, error) {
	switch cfg.Type {
	case ProviderClaude:
		return NewClaudeProvider(cfg.APIKey, cfg.Model, cfg.MaxTokens), nil
	case ProviderOpenAI:
		return NewOpenAIProvider(cfg.APIKey, cfg.Model, cfg.MaxTokens, cfg.BaseURL), nil
	case ProviderOllama:
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return NewOllamaProvider(cfg.Model, baseURL), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", cfg.Type)
	}
}
