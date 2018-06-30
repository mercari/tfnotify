package github

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

func TestCommentPost(t *testing.T) {
	testCases := []struct {
		config Config
		body   string
		opt    PostOptions
		ok     bool
	}{
		{
			config: newFakeConfig(),
			body:   "",
			opt: PostOptions{
				Number:   1,
				Revision: "abcd",
			},
			ok: true,
		},
		{
			config: newFakeConfig(),
			body:   "",
			opt: PostOptions{
				Number:   0,
				Revision: "abcd",
			},
			ok: true,
		},
		{
			config: newFakeConfig(),
			body:   "",
			opt: PostOptions{
				Number:   2,
				Revision: "",
			},
			ok: true,
		},
		{
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
		client, err := NewClient(testCase.config)
		if err != nil {
			t.Fatal(err)
		}
		api := newFakeAPI()
		client.API = &api
		err = client.Comment.Post(testCase.body, testCase.opt)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
	}
}

func TestCommentList(t *testing.T) {
	comments := []*github.IssueComment{
		&github.IssueComment{
			ID:   github.Int64(371748792),
			Body: github.String("comment 1"),
		},
		&github.IssueComment{
			ID:   github.Int64(371765743),
			Body: github.String("comment 2"),
		},
	}
	testCases := []struct {
		config   Config
		number   int
		ok       bool
		comments []*github.IssueComment
	}{
		{
			config:   newFakeConfig(),
			number:   1,
			ok:       true,
			comments: comments,
		},
		{
			config:   newFakeConfig(),
			number:   12,
			ok:       true,
			comments: comments,
		},
		{
			config:   newFakeConfig(),
			number:   123,
			ok:       true,
			comments: comments,
		},
	}

	for _, testCase := range testCases {
		client, err := NewClient(testCase.config)
		if err != nil {
			t.Fatal(err)
		}
		api := newFakeAPI()
		client.API = &api
		comments, err := client.Comment.List(testCase.number)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
		if !reflect.DeepEqual(comments, testCase.comments) {
			t.Errorf("got %v but want %v", comments, testCase.comments)
		}
	}
}

func TestCommentDelete(t *testing.T) {
	testCases := []struct {
		config Config
		id     int
		ok     bool
	}{
		{
			config: newFakeConfig(),
			id:     1,
			ok:     true,
		},
		{
			config: newFakeConfig(),
			id:     12,
			ok:     true,
		},
		{
			config: newFakeConfig(),
			id:     123,
			ok:     true,
		},
	}

	for _, testCase := range testCases {
		client, err := NewClient(testCase.config)
		if err != nil {
			t.Fatal(err)
		}
		api := newFakeAPI()
		client.API = &api
		err = client.Comment.Delete(testCase.id)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
	}
}

func TestCommentGetDuplicates(t *testing.T) {
	api := newFakeAPI()
	api.FakeIssuesListComments = func(ctx context.Context, number int, opt *github.IssueListCommentsOptions) ([]*github.IssueComment, *github.Response, error) {
		var comments []*github.IssueComment
		comments = []*github.IssueComment{
			&github.IssueComment{
				ID:   github.Int64(371748792),
				Body: github.String("## Plan result\nfoo message\n"),
			},
			&github.IssueComment{
				ID:   github.Int64(371765743),
				Body: github.String("## Plan result\nbar message\n"),
			},
			&github.IssueComment{
				ID:   github.Int64(371765744),
				Body: github.String("## Plan result\nbaz message\n"),
			},
			&github.IssueComment{
				ID:   github.Int64(371765745),
				Body: github.String("## Plan result <build URL>\nbaz message\n"),
			},
		}
		return comments, nil, nil
	}

	testCases := []struct {
		title    string
		message  string
		comments []*github.IssueComment
	}{
		{
			title:   "## Plan result",
			message: "foo message",
			comments: []*github.IssueComment{
				&github.IssueComment{
					ID:   github.Int64(371748792),
					Body: github.String("## Plan result\nfoo message\n"),
				},
			},
		},
		{
			title:    "## Plan result",
			message:  "hoge message",
			comments: nil,
		},
		{
			title:   "## Plan result",
			message: "baz message",
			comments: []*github.IssueComment{
				&github.IssueComment{
					ID:   github.Int64(371765744),
					Body: github.String("## Plan result\nbaz message\n"),
				},
				&github.IssueComment{
					ID:   github.Int64(371765745),
					Body: github.String("## Plan result <build URL>\nbaz message\n"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		cfg := newFakeConfig()
		cfg.PR.Message = testCase.message
		client, err := NewClient(cfg)
		if err != nil {
			t.Fatal(err)
		}
		client.API = &api
		comments := client.Comment.getDuplicates(testCase.title)
		if !reflect.DeepEqual(comments, testCase.comments) {
			t.Errorf("got %q but want %q", comments, testCase.comments)
		}
	}
}
