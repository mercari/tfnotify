package typetalk

import (
	"context"
	"testing"

	"github.com/mercari/tfnotify/terraform"
	typetalkShared "github.com/nulab/go-typetalk/typetalk/shared"
	typetalk "github.com/nulab/go-typetalk/typetalk/v1"
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
				TopicID:  "12345",
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
				TopicID:  "12345",
				Message:  "",
				Parser:   terraform.NewPlanParser(),
				Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
			},
			body:     "BLUR BLUR BLUR",
			exitCode: 1,
			ok:       false,
		},
	}
	fake := fakeAPI{
		FakeChatPostMessage: func(ctx context.Context, message string) (*typetalk.PostedMessageResult, *typetalkShared.Response, error) {
			return nil, nil, nil
		},
	}

	for _, testCase := range testCases {
		client, err := NewClient(testCase.config)
		if err != nil {
			t.Fatal(err)
		}
		client.API = &fake
		exitCode, err := client.Notify.Notify(context.Background(), testCase.body)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
		if exitCode != testCase.exitCode {
			t.Errorf("got %q but want %q", exitCode, testCase.exitCode)
		}
	}
}
