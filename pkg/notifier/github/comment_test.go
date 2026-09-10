package github

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/suzuki-shunsuke/github-comment-metadata/metadata"
)

func TestTruncateComment(t *testing.T) {
	t.Parallel()

	meta := "\n<!-- github-comment: {\"Program\":\"tfnotify\",\"Command\":\"plan\",\"Target\":\"prod\"} -->"

	testCases := []struct {
		name    string
		body    string
		wantSub string // substring that must survive
	}{
		{
			name: "short body is unchanged",
			body: "hello world",
		},
		{
			name: "body at the limit is unchanged",
			body: strings.Repeat("a", githubMaxCommentLength),
		},
		{
			name: "oversized body is truncated",
			body: strings.Repeat("a", githubMaxCommentLength+10000),
		},
		{
			name: "oversized body preserves trailing metadata",
			body: strings.Repeat("a", githubMaxCommentLength+10000) + meta,
		},
		{
			name: "oversized multibyte body stays valid utf8 within the limit",
			body: strings.Repeat("日", githubMaxCommentLength+10000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := truncateComment(tc.body)

			if utf8.RuneCountInString(tc.body) <= githubMaxCommentLength {
				if got != tc.body {
					t.Fatalf("body within limit must be unchanged")
				}
				return
			}

			if n := utf8.RuneCountInString(got); n > githubMaxCommentLength {
				t.Errorf("truncated body length = %d runes, exceeds limit %d", n, githubMaxCommentLength)
			}
			if !utf8.ValidString(got) {
				t.Error("truncated body is not valid UTF-8")
			}
			if !strings.Contains(got, "truncated") {
				t.Error("truncated body is missing the truncation notice")
			}

			// A trailing tfnotify metadata line must survive so the comment can
			// still be found and patched on the next run.
			if strings.Contains(tc.body, "github-comment:") {
				data := &Metadata{}
				ok, err := metadata.Extract(got, data)
				if err != nil {
					t.Fatalf("metadata.Extract: %v", err)
				}
				if !ok || data.Program != "tfnotify" {
					t.Error("truncated body lost its github-comment metadata line")
				}
			}
		})
	}
}

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
