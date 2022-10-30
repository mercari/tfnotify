package mattermost

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	mattermostWebhook := os.Getenv(EnvWebhook)
	defer func() {
		os.Setenv(EnvWebhook, mattermostWebhook)
	}()
	os.Setenv(EnvWebhook, "")

	testCases := []struct {
		config     Config
		EnvWebhook string
		expect     string
	}{
		{
			// specify directly
			config:     Config{Webhook: "abcdefg"},
			EnvWebhook: "",
			expect:     "",
		},
		{
			// specify via env but not to be set env (part 1)
			config:     Config{Webhook: "MATTERMOST_WEBHOOK"},
			EnvWebhook: "",
			expect:     "mattermost webhook is missing",
		},
		{
			// specify via env (part 1)
			config:     Config{Webhook: "MATTERMOST_WEBHOOK"},
			EnvWebhook: "abcdefg",
			expect:     "",
		},
		{
			// specify via env but not to be set env (part 2)
			config:     Config{Webhook: "$MATTERMOST_WEBHOOK"},
			EnvWebhook: "",
			expect:     "mattermost webhook is missing",
		},
		{
			// specify via env (part 2)
			config:     Config{Webhook: "$MATTERMOST_WEBHOOK"},
			EnvWebhook: "abcdefg",
			expect:     "",
		},
		{
			// no specification (part 1)
			config:     Config{},
			EnvWebhook: "",
			expect:     "mattermost webhook is missing",
		},
		{
			// no specification (part 2)
			config:     Config{},
			EnvWebhook: "abcdefg",
			expect:     "mattermost webhook is missing",
		},
	}
	for _, testCase := range testCases {
		os.Setenv(EnvWebhook, testCase.EnvWebhook)
		_, err := NewClient(testCase.config)
		if err == nil {
			continue
		}
		if err.Error() != testCase.expect {
			t.Errorf("got %q but want %q", err.Error(), testCase.expect)
		}
	}
}
