package typetalk

import (
	"context"

	typetalkShared "github.com/nulab/go-typetalk/typetalk/shared"
	typetalk "github.com/nulab/go-typetalk/typetalk/v1"
)

// API is Typetalk API interface
type API interface {
	ChatPostMessage(ctx context.Context, message string) (*typetalk.PostedMessageResult, *typetalkShared.Response, error)
}

// Typetalk represents the attribute information necessary for requesting Typetalk API
type Typetalk struct {
	*typetalk.Client
	TopicID int
}

// ChatPostMessage is wrapper for https://godoc.org/github.com/nulab/go-typetalk/typetalk/v1#MessagesService.PostMessage
func (t *Typetalk) ChatPostMessage(ctx context.Context, message string) (*typetalk.PostedMessageResult, *typetalkShared.Response, error) {
	return t.Client.Messages.PostMessage(ctx, t.TopicID, message, nil)
}
