package mattermost

import (
	"github.com/ashwanthkumar/slack-go-webhook"
)

type fakeAPI struct {
	API
	ChatPostMessageError error
}

func (f *fakeAPI) ChatPostMessage(attachments []slack.Attachment) error {
	return f.ChatPostMessageError
}
