package mattermost

import (
	"testing"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/mercari/tfnotify/terraform"
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
				Webhook:  "webhook",
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
				Webhook:  "webhook",
				Channel:  "",
				Botname:  "botname",
				Message:  "",
				Parser:   terraform.NewPlanParser(),
				Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
			},
			body:     "Plan: 1 to add",
			exitCode: 0,
			ok:       true,
		},
	}
	fake := fakeAPI{
		FakeChatPostMessage: func(attachments []slack.Attachment) error {
			return nil
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
