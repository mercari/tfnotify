package mattermost

import (
	"github.com/ashwanthkumar/slack-go-webhook"
)

type fakeAPI struct {
	API
	FakeChatPostMessage func(attachments []slack.Attachment) error
}

func (f *fakeAPI) ChatPostMessage(attachments []slack.Attachment) error {
	return f.FakeChatPostMessage(attachments)
}
