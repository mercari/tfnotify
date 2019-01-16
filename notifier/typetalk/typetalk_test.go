package typetalk

import (
	"context"

	typetalkShared "github.com/nulab/go-typetalk/typetalk/shared"
	typetalk "github.com/nulab/go-typetalk/typetalk/v1"
)

type fakeAPI struct {
	API
	FakeChatPostMessage func(ctx context.Context, message string) (*typetalk.PostedMessageResult, *typetalkShared.Response, error)
}

func (g *fakeAPI) ChatPostMessage(ctx context.Context, message string) (*typetalk.PostedMessageResult, *typetalkShared.Response, error) {
	return g.FakeChatPostMessage(ctx, message)
}
