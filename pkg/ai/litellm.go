package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

// LiteLLM API structures (uses OpenAI-compatible format)
type liteLLMRequest struct {
	Model     string           `json:"model"`
	Messages  []liteLLMMessage `json:"messages"`
	MaxTokens int              `json:"max_tokens,omitempty"`
}

type liteLLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type liteLLMResponse struct {
	Choices []struct {
		Message liteLLMMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *Summarizer) callLiteLLM(ctx context.Context, prompt string) (string, error) {
	s.logger.Info("calling LiteLLM API...")

	apiKey := s.config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("LITELLM_API_KEY")
	}
	if apiKey == "" {
		s.logger.Error("LITELLM_API_KEY not set")
		return "", fmt.Errorf("LITELLM_API_KEY not set")
	}
	s.logger.Debug("API key found")

	// LiteLLM base URL (default to localhost, can be overridden via env var)
	baseURL := os.Getenv("LITELLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://litellm.mercari.in"
	}
	s.logger.WithField("base_url", baseURL).Debug("LiteLLM base URL")

	model := s.config.Model
	if model == "" {
		model = "gpt-4o" // Default model
	}

	maxTokens := s.config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 500
	}

	s.logger.WithFields(logrus.Fields{
		"model":      model,
		"max_tokens": maxTokens,
		"base_url":   baseURL,
	}).Info("LiteLLM request configuration")

	reqBody := liteLLMRequest{
		Model: model,
		Messages: []liteLLMMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// LiteLLM endpoint format: /openai/deployments/{model}/chat/completions
	endpoint := fmt.Sprintf("%s/openai/deployments/%s/chat/completions", baseURL, model)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	s.logger.WithField("endpoint", endpoint).Info("sending request to LiteLLM...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.WithError(err).Error("failed to send request to LiteLLM")
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	s.logger.WithField("status_code", resp.StatusCode).Info("received response from LiteLLM")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"body":        string(body),
		}).Error("LiteLLM API returned non-200 status")
		return "", fmt.Errorf("LiteLLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp liteLLMResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		s.logger.WithError(err).Error("failed to unmarshal LiteLLM response")
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if apiResp.Error != nil {
		s.logger.WithField("error", apiResp.Error.Message).Error("LiteLLM API returned error")
		return "", fmt.Errorf("LiteLLM API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		s.logger.Error("LiteLLM response has no choices")
		return "", fmt.Errorf("no choices in response")
	}

	s.logger.WithField("content_length", len(apiResp.Choices[0].Message.Content)).Info("LiteLLM summary generated successfully")
	return apiResp.Choices[0].Message.Content, nil
}
