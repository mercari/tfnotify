package slack

import (
	"context"
	"testing"

	"github.com/mercari/tfnotify/terraform"
	"github.com/lestrrat-go/slack/objects"
)

func TestNotify(t *testing.T) {
	testCases := []struct {
		config   Config
		body     string
		exitCode int
		ok       bool
	}{
		{
			config: Config{
				Token:    "token",
				Channel:  "channel",
				Botname:  "botname",
				Message:  "",
				Parser:   terraform.NewPlanParser(),
				Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
			},
			body:     "Plan: 1 to add",
			exitCode: 0,
			ok:       true,
		},
		{
			config: Config{
				Token:    "token",
				Channel:  "",
				Botname:  "botname",
				Message:  "",
				Parser:   terraform.NewPlanParser(),
				Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
			},
			body:     "Plan: 1 to add",
			exitCode: 1,
			ok:       false,
		},
	}
	fake := fakeAPI{
		FakeChatPostMessage: func(ctx context.Context, attachments []*objects.Attachment) (*objects.ChatResponse, error) {
			return nil, nil
		},
	}

	for _, testCase := range testCases {
		client, err := NewClient(testCase.config)
		if err != nil {
			t.Fatal(err)
		}
		client.API = &fake
		exitCode, err := client.Notify.Notify(testCase.body)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
		if exitCode != testCase.exitCode {
			t.Errorf("got %q but want %q", exitCode, testCase.exitCode)
		}
	}
}
