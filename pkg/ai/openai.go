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

// OpenAI API structures
type openAIRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// callOpenAI calls OpenAI API
func (s *Summarizer) callOpenAI(ctx context.Context, prompt string) (string, error) {
	s.logger.Info("calling OpenAI API...")

	apiKey := s.config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		s.logger.Error("OPENAI_API_KEY not set")
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}
	s.logger.Debug("API key found")

	model := s.config.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	maxTokens := s.config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 500
	}

	s.logger.WithFields(logrus.Fields{
		"model":      model,
		"max_tokens": maxTokens,
	}).Info("OpenAI request configuration")

	reqBody := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
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

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	s.logger.Info("sending request to OpenAI...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.WithError(err).Error("failed to send request to OpenAI")
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	s.logger.WithField("status_code", resp.StatusCode).Info("received response from OpenAI")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"body":        string(body),
		}).Error("OpenAI API returned non-200 status")
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp openAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		s.logger.WithError(err).Error("failed to unmarshal OpenAI response")
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if apiResp.Error != nil {
		s.logger.WithField("error", apiResp.Error.Message).Error("OpenAI API returned error")
		return "", fmt.Errorf("OpenAI API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		s.logger.Error("OpenAI response has no choices")
		return "", fmt.Errorf("no choices in response")
	}

	s.logger.WithField("content_length", len(apiResp.Choices[0].Message.Content)).Info("OpenAI summary generated successfully")
	return apiResp.Choices[0].Message.Content, nil
}
