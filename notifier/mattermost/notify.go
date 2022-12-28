package mattermost

import (
	"errors"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/mercari/tfnotify/terraform"
)

// NotifyService handles communication with the notification related
// methods of Slack API
type NotifyService service

// mmString handles converting string to pointer
func mmString(s string) *string {
	return &s
}

// Notify posts comment optimized for notifications
func (m *NotifyService) Notify(body string) (exit int, err error) {
	cfg := m.client.Config
	parser := m.client.Config.Parser
	template := m.client.Config.Template

	if cfg.Webhook == "" {
		return terraform.ExitFail, errors.New("webhook is required")
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
		Link:    cfg.CI,
	})
	text, err := template.Execute()
	if err != nil {
		return result.ExitCode, err
	}

	var attachments []slack.Attachment
	attachment := slack.Attachment{
		Color:    mmString(color),
		Fallback: mmString(text),
		Footer:   mmString(cfg.CI),
		Text:     mmString(text),
		Title:    mmString(template.GetValue().Title),
	}

	attachments = append(attachments, attachment)

	err = m.client.API.ChatPostMessage(attachments)
	return result.ExitCode, err
}
