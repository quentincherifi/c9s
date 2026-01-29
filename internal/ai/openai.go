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
	openAIAPIURL = "https://api.openai.com/v1/chat/completions"

	// DefaultOpenAIModel is the default OpenAI model.
	DefaultOpenAIModel = "gpt-4o"
)

// OpenAIProvider implements the Provider interface for OpenAI.
type OpenAIProvider struct {
	apiKey     string
	model      string
	maxTokens  int
	baseURL    string
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(apiKey, model string, maxTokens int, baseURL string) *OpenAIProvider {
	if model == "" {
		model = DefaultOpenAIModel
	}
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}
	if baseURL == "" {
		baseURL = openAIAPIURL
	}
	return &OpenAIProvider{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Name returns the provider name.
func (*OpenAIProvider) Name() string {
	return "OpenAI"
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens,omitempty"`
}

type openAIChoice struct {
	Index   int `json:"index"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type openAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Send sends a message to OpenAI and returns the response.
func (o *OpenAIProvider) Send(system string, messages []Message) (*Response, error) {
	// Build messages array with system message first
	openAIMessages := make([]openAIMessage, 0, len(messages)+1)

	if system != "" {
		openAIMessages = append(openAIMessages, openAIMessage{
			Role:    "system",
			Content: system,
		})
	}

	for _, msg := range messages {
		openAIMessages = append(openAIMessages, openAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	req := openAIRequest{
		Model:     o.model,
		Messages:  openAIMessages,
		MaxTokens: o.maxTokens,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, o.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp openAIErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return nil, fmt.Errorf("API error: %s - %s", errResp.Error.Type, errResp.Error.Message)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var content string
	if len(openAIResp.Choices) > 0 {
		content = openAIResp.Choices[0].Message.Content
	}

	return &Response{
		Content:      content,
		Model:        openAIResp.Model,
		InputTokens:  openAIResp.Usage.PromptTokens,
		OutputTokens: openAIResp.Usage.CompletionTokens,
	}, nil
}
