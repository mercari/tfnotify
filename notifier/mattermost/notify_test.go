package mattermost

import (
	"fmt"
	"testing"

	"github.com/mercari/tfnotify/terraform"
	"github.com/stretchr/testify/assert"
)

func TestNotify(t *testing.T) {
	testCases := []struct {
		config                 Config
		body                   string
		exitCode               int
		ok                     bool
		fakeAPI                *fakeAPI
		expectedApiErrorString string
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
			fakeAPI: &fakeAPI{
				ChatPostMessageError: nil,
			},
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
			fakeAPI: &fakeAPI{
				ChatPostMessageError: nil,
			},
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
			ok:       false,
			fakeAPI: &fakeAPI{
				ChatPostMessageError: fmt.Errorf("500 Internal Server Error"),
			},
			expectedApiErrorString: "500 Internal Server Error",
		},
	}

	for _, testCase := range testCases {
		client, err := NewClient(testCase.config)
		if err != nil {
			t.Fatal(err)
		}
		client.API = testCase.fakeAPI
		exitCode, err := client.Notify.Notify(testCase.body)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
		if exitCode != testCase.exitCode {
			t.Errorf("got %q but want %q", exitCode, testCase.exitCode)
		}
		if err != nil {
			assert.EqualError(t, err, testCase.expectedApiErrorString)
		}
	}
}
