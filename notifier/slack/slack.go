package slack

import (
	"context"

	"github.com/lestrrat-go/slack"
	"github.com/lestrrat-go/slack/objects"
)

// API is Slack API interface
type API interface {
	ChatPostMessage(ctx context.Context, attachments []*objects.Attachment) (*objects.ChatResponse, error)
}

// Slack represents the attribute information necessary for requesting Slack API
type Slack struct {
	*slack.Client
	Channel string
	Botname string
}

// ChatPostMessage is a wrapper of https://godoc.org/github.com/lestrrat-go/slack#ChatPostMessageCall
func (s *Slack) ChatPostMessage(ctx context.Context, attachments []*objects.Attachment) (*objects.ChatResponse, error) {
	return s.Client.Chat().PostMessage(s.Channel).Username(s.Botname).SetAttachments(attachments).Do(ctx)
}
