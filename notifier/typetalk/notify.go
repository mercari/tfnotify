package typetalk

import (
	"context"
	"errors"

	"github.com/mercari/tfnotify/terraform"
)

// NotifyService handles notification process.
type NotifyService service

// Notify posts message to Typetalk.
func (s *NotifyService) Notify(body string) (exit int, err error) {
	cfg := s.client.Config
	parser := s.client.Config.Parser
	template := s.client.Config.Template

	if cfg.TopicID == "" {
		return terraform.ExitFail, errors.New("topic id is required")
	}

	result := parser.Parse(body)
	if result.Error != nil {
		return result.ExitCode, result.Error
	}
	if result.Result == "" {
		return result.ExitCode, result.Error
	}

	template.SetValue(terraform.CommonTemplate{
		Title:   cfg.Title,
		Message: cfg.Message,
		Result:  result.Result,
		Body:    body,
	})
	text, err := template.Execute()
	if err != nil {
		return result.ExitCode, err
	}

	_, _, err = s.client.API.ChatPostMessage(context.Background(), text)
	return result.ExitCode, err
}
