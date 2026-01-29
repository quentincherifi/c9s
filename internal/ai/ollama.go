// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// DefaultOllamaModel is the default Ollama model.
	DefaultOllamaModel = "llama3.2"
	// DefaultOllamaURL is the default Ollama API URL.
	DefaultOllamaURL = "http://localhost:11434"
)

// OllamaProvider implements the Provider interface for local Ollama models.
type OllamaProvider struct {
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewOllamaProvider creates a new Ollama provider.
func NewOllamaProvider(model, baseURL string) *OllamaProvider {
	if model == "" {
		model = DefaultOllamaModel
	}
	if baseURL == "" {
		baseURL = DefaultOllamaURL
	}
	return &OllamaProvider{
		model:   model,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout * 2, // Longer timeout for local models
		},
	}
}

// Name returns the provider name.
func (*OllamaProvider) Name() string {
	return "Ollama"
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaResponse struct {
	Model   string `json:"model"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done              bool  `json:"done"`
	TotalDuration     int64 `json:"total_duration"`
	PromptEvalCount   int   `json:"prompt_eval_count"`
	EvalCount         int   `json:"eval_count"`
}

// Send sends a message to Ollama and returns the response.
func (o *OllamaProvider) Send(system string, messages []Message) (*Response, error) {
	// Build messages array with system message first
	ollamaMessages := make([]ollamaMessage, 0, len(messages)+1)

	if system != "" {
		ollamaMessages = append(ollamaMessages, ollamaMessage{
			Role:    "system",
			Content: system,
		})
	}

	for _, msg := range messages {
		ollamaMessages = append(ollamaMessages, ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	req := ollamaRequest{
		Model:    o.model,
		Messages: ollamaMessages,
		Stream:   false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := o.baseURL + "/api/chat"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &Response{
		Content:      ollamaResp.Message.Content,
		Model:        ollamaResp.Model,
		InputTokens:  ollamaResp.PromptEvalCount,
		OutputTokens: ollamaResp.EvalCount,
	}, nil
}
