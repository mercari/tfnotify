package github

import (
	"testing"
)

func TestCommentPost(t *testing.T) { //nolint:tparallel
	t.Setenv("GITHUB_TOKEN", "xxx")
	testCases := []struct {
		name   string
		config Config
		body   string
		opt    PostOptions
		ok     bool
	}{
		{
			name:   "1",
			config: newFakeConfig(),
			body:   "",
			opt: PostOptions{
				Number:   1,
				Revision: "abcd",
			},
			ok: true,
		},
		{
			name:   "2",
			config: newFakeConfig(),
			body:   "",
			opt: PostOptions{
				Number:   2,
				Revision: "",
			},
			ok: true,
		},
		{
			name:   "3",
			config: newFakeConfig(),
			body:   "",
			opt: PostOptions{
				Number:   0,
				Revision: "",
			},
			ok: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			cfg := testCase.config
			client, err := NewClient(t.Context(), &cfg)
			if err != nil {
				t.Fatal(err)
			}
			api := newFakeAPI()
			client.API = &api
			opt := testCase.opt
			err = client.Comment.Post(t.Context(), testCase.body, &opt)
			if (err == nil) != testCase.ok {
				t.Errorf("got error %q", err)
			}
		})
	}
}
