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

// Devin API structures for creating sessions
type CreateSessionRequest struct {
	Prompt       string   `json:"prompt"`                  // Required: The task description for Devin
	SnapshotID   *string  `json:"snapshot_id,omitempty"`   // Optional: ID of a machine snapshot to use
	Unlisted     *bool    `json:"unlisted,omitempty"`      // Optional: Whether the session should be unlisted
	Idempotent   *bool    `json:"idempotent,omitempty"`    // Optional: Enable idempotent session creation
	MaxACULimit  *int     `json:"max_acu_limit,omitempty"` // Optional: Maximum ACU limit for the session
	SecretIDs    []string `json:"secret_ids,omitempty"`    // Optional: List of secret IDs to use
	KnowledgeIDs []string `json:"knowledge_ids,omitempty"` // Optional: List of knowledge IDs to use
	Tags         []string `json:"tags,omitempty"`          // Optional: List of tags to add to the session
	Title        *string  `json:"title,omitempty"`         // Optional: Custom title for the session
}

type CreateSessionResponse struct {
	SessionID    string `json:"session_id"`     // Unique identifier for the session
	URL          string `json:"url"`            // URL to view the session in the web interface
	IsNewSession *bool  `json:"is_new_session"` // Indicates if a new session was created (only present if idempotent=true)
}

// Devin API structures for sending messages
type SendMessageRequest struct {
	Message string `json:"message"` // Required: The message to send to Devin
}

// SessionStatusResponse represents the session status from Devin API
type SessionStatusResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"` // e.g., "idle", "busy", "completed", "error"
	URL       string `json:"url"`
}

// SessionMessagesResponse represents messages from a Devin session
type SessionMessagesResponse struct {
	Messages []DevinMessage `json:"messages"`
}

// DevinMessage represents a message in a Devin session
type DevinMessage struct {
	ID        string `json:"id"`
	Role      string `json:"role"` // "user" or "assistant"
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// callDevin handles Devin AI provider with session management
func (s *Summarizer) callDevin(ctx context.Context, sessionID string, prompt string) (string, error) {
	// If no session ID provided, create a new session
	prompt = "Comment on the Pull Request if there is a specified PR number.\n\n" + prompt
	if sessionID == "" {
		s.logger.Info("no session ID provided, creating new Devin session...")
		return s.createDevinSessionAndSendMessage(ctx, prompt)
	}

	// Check if session exists and is available
	status, err := s.GetSessionStatus(ctx, sessionID)
	if err != nil {
		// Session doesn't exist or error checking status - create new session
		s.logger.WithError(err).Warn("session not found or error checking status, creating new session...")
		return s.createDevinSessionAndSendMessage(ctx, prompt)
	}

	// Check if session is busy
	if status.Status == "busy" {
		s.logger.WithField("session_id", sessionID).Warn("session is busy, creating new session...")
		return s.createDevinSessionAndSendMessage(ctx, prompt)
	}

	// Session exists and is available, send message to it
	s.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"status":     status.Status,
	}).Info("using existing Devin session")

	return s.SendMessageToSession(ctx, sessionID, prompt)
}

// createDevinSessionAndSendMessage creates a new session and sends the initial message
func (s *Summarizer) createDevinSessionAndSendMessage(ctx context.Context, prompt string) (string, error) {
	session, err := s.CreateSession(ctx, prompt, nil)
	if err != nil {
		return "", fmt.Errorf("create Devin session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": session.SessionID,
		"url":        session.URL,
	}).Info("new Devin session created")

	// Return information about the session creation
	// Note: Devin works asynchronously, so we can't wait for a response immediately
	return fmt.Sprintf("**Devin Session Created**\n\nSession ID: `%s`\n\nDevin is now analyzing the Terraform plan failure. View progress at: %s\n\nNote: Devin works asynchronously, please wait for it's comment.", session.SessionID, session.URL), nil
}

// GetSessionStatus checks the status of a Devin session
func (s *Summarizer) GetSessionStatus(ctx context.Context, sessionID string) (*SessionStatusResponse, error) {
	s.logger.WithField("session_id", sessionID).Debug("checking Devin session status...")

	apiKey := s.config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("DEVIN_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("DEVIN_API_KEY not set")
	}

	baseURL := os.Getenv("DEVIN_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.devin.ai"
	}

	endpoint := fmt.Sprintf("%s/v1/sessions/%s", baseURL, sessionID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		s.logger.WithField("session_id", sessionID).Warn("session not found")
		return nil, fmt.Errorf("session not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Devin API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var statusResp SessionStatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &statusResp, nil
}

// SendMessageToSession sends a message to an existing Devin session and returns the summary
func (s *Summarizer) SendMessageToSession(ctx context.Context, sessionID, message string) (string, error) {
	s.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
	}).Info("sending message to Devin session...")

	apiKey := s.config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("DEVIN_API_KEY")
	}
	if apiKey == "" {
		s.logger.Error("DEVIN_API_KEY not set")
		return "", fmt.Errorf("DEVIN_API_KEY not set")
	}

	baseURL := os.Getenv("DEVIN_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.devin.ai"
	}

	reqBody := SendMessageRequest{
		Message: message,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// Note: endpoint is /message (singular), not /messages
	endpoint := fmt.Sprintf("%s/v1/sessions/%s/message", baseURL, sessionID)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	s.logger.WithField("endpoint", endpoint).Info("sending request to Devin...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.WithError(err).Error("failed to send request to Devin")
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	s.logger.WithField("status_code", resp.StatusCode).Info("received response from Devin")

	// API returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"body":        string(body),
		}).Error("Devin API returned error status")
		return "", fmt.Errorf("Devin API error (status %d): %s", resp.StatusCode, string(body))
	}

	s.logger.Info("message sent to Devin session successfully")

	// Return formatted message about the session
	return fmt.Sprintf("**Message Sent to Devin Session**\n\nSession ID: `%s`\n\nDevin is analyzing your request. View progress at: %s/sessions/%s\n", sessionID, baseURL, sessionID), nil
}

// CreateSession creates a new Devin session (optional - if you also need this)
func (s *Summarizer) CreateSession(ctx context.Context, prompt string, options *CreateSessionRequest) (*CreateSessionResponse, error) {
	s.logger.Info("creating new Devin session...")

	apiKey := s.config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("DEVIN_API_KEY")
	}
	if apiKey == "" {
		s.logger.Error("DEVIN_API_KEY not set")
		return nil, fmt.Errorf("DEVIN_API_KEY not set")
	}

	baseURL := os.Getenv("DEVIN_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.devin.ai"
	}

	reqBody := CreateSessionRequest{
		Prompt: prompt,
	}

	// Merge optional parameters if provided
	if options != nil {
		if options.SnapshotID != nil {
			reqBody.SnapshotID = options.SnapshotID
		}
		if options.Unlisted != nil {
			reqBody.Unlisted = options.Unlisted
		}
		if options.Idempotent != nil {
			reqBody.Idempotent = options.Idempotent
		}
		if options.MaxACULimit != nil {
			reqBody.MaxACULimit = options.MaxACULimit
		}
		if options.SecretIDs != nil {
			reqBody.SecretIDs = options.SecretIDs
		}
		if options.KnowledgeIDs != nil {
			reqBody.KnowledgeIDs = options.KnowledgeIDs
		}
		if options.Tags != nil {
			reqBody.Tags = options.Tags
		}
		if options.Title != nil {
			reqBody.Title = options.Title
		}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := baseURL + "/v1/sessions"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	s.logger.WithField("endpoint", endpoint).Info("sending request to Devin...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.WithError(err).Error("failed to send request to Devin")
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	s.logger.WithField("status_code", resp.StatusCode).Info("received response from Devin")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		s.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"body":        string(body),
		}).Error("Devin API returned error status")
		return nil, fmt.Errorf("Devin API error (status %d): %s", resp.StatusCode, string(body))
	}

	var sessionResp CreateSessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		s.logger.WithError(err).Error("failed to unmarshal Devin response")
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": sessionResp.SessionID,
		"url":        sessionResp.URL,
	}).Info("Devin session created successfully")

	return &sessionResp, nil
}
