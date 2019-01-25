package typetalk

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	typetalkToken := os.Getenv(EnvToken)
	defer func() {
		os.Setenv(EnvToken, typetalkToken)
	}()
	os.Setenv(EnvToken, "")

	testCases := []struct {
		config   Config
		envToken string
		expect   string
	}{
		{
			// specify directly
			config:   Config{Token: "abcdefg", TopicID: "12345"},
			envToken: "",
			expect:   "",
		},
		{
			// specify via env but not to be set env (part 1)
			config:   Config{Token: "TYPETALK_TOKEN", TopicID: "12345"},
			envToken: "",
			expect:   "Typetalk token is missing",
		},
		{
			// specify via env (part 1)
			config:   Config{Token: "TYPETALK_TOKEN", TopicID: "12345"},
			envToken: "abcdefg",
			expect:   "",
		},
		{
			// specify via env but not to be set env (part 2)
			config:   Config{Token: "$TYPETALK_TOKEN", TopicID: "12345"},
			envToken: "",
			expect:   "Typetalk token is missing",
		},
		{
			// specify via env (part 2)
			config:   Config{Token: "$TYPETALK_TOKEN", TopicID: "12345"},
			envToken: "abcdefg",
			expect:   "",
		},
		{
			// no specification (part 1)
			config:   Config{TopicID: "12345"},
			envToken: "",
			expect:   "Typetalk token is missing",
		},
		{
			// no specification (part 2)
			config:   Config{TopicID: "12345"},
			envToken: "abcdefg",
			expect:   "Typetalk token is missing",
		},
	}
	for _, testCase := range testCases {
		os.Setenv(EnvToken, testCase.envToken)
		_, err := NewClient(testCase.config)
		if err == nil {
			continue
		}
		if err.Error() != testCase.expect {
			t.Errorf("got %q but want %q", err.Error(), testCase.expect)
		}
	}
}
