package slack

import (
	"context"
	"errors"

	"github.com/lestrrat-go/slack/objects"
	"github.com/mercari/tfnotify/terraform"
)

// NotifyService handles communication with the notification related
// methods of Slack API
type NotifyService service

// Notify posts comment optimized for notifications
func (s *NotifyService) Notify(body string) (exit int, err error) {
	cfg := s.client.Config
	parser := s.client.Config.Parser
	template := s.client.Config.Template

	if cfg.Channel == "" {
		return terraform.ExitFail, errors.New("channel id is required")
	}

	result := parser.Parse(body)
	if result.Error != nil {
		return result.ExitCode, result.Error
	}
	if result.Result == "" {
		return result.ExitCode, result.Error
	}

	color := "warning"
	switch result.ExitCode {
	case terraform.ExitPass:
		color = "good"
	case terraform.ExitFail:
		color = "danger"
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

	var attachments objects.AttachmentList
	attachment := &objects.Attachment{
		Color:    color,
		Fallback: text,
		Footer:   cfg.CI,
		Text:     text,
		Title:    template.GetValue().Title,
	}

	attachments.Append(attachment)
	// _, err = s.client.Chat().PostMessage(cfg.Channel).Username(cfg.Botname).SetAttachments(attachments).Do(cfg.Context)
	_, err = s.client.API.ChatPostMessage(context.Background(), attachments)
	return result.ExitCode, err
}
