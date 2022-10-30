package mattermost

import (
	"fmt"

	"github.com/ashwanthkumar/slack-go-webhook"
)

// API is Mattermost API interface
type API interface {
	ChatPostMessage(attachments []slack.Attachment) error
}

// Mattermost represents the attribute information necessary for requesting Mattermost Webhook API
type Mattermost struct {
	Webhook string
	Channel string
	Botname string
}

// ChatPostMessage is a wrapper of https://pkg.go.dev/github.com/ashwanthkumar/slack-go-webhook#Send
func (m *Mattermost) ChatPostMessage(attachments []slack.Attachment) error {

	payload := slack.Payload{
		Username: func() string {
			if m.Botname != "" {
				return m.Botname
			} else {
				return "tfnotify"
			}
		}(),
		Channel:     m.Channel,
		IconUrl:     "https://docs.mattermost.com/_images/icon-76x76.png",
		Attachments: attachments,
		Markdown:    true,
	}

	errs := slack.Send(m.Webhook, "", payload)
	if len(errs) > 0 {
		_, err := fmt.Printf("error: %s\n", errs)
		return err
	}

	return nil
}
