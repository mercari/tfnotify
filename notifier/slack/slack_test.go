package slack

import (
	"context"

	"github.com/lestrrat-go/slack/objects"
)

type fakeAPI struct {
	API
	FakeChatPostMessage func(ctx context.Context, attachments []*objects.Attachment) (*objects.ChatResponse, error)
}

func (g *fakeAPI) ChatPostMessage(ctx context.Context, attachments []*objects.Attachment) (*objects.ChatResponse, error) {
	return g.FakeChatPostMessage(ctx, attachments)
}
