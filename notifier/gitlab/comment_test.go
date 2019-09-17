package gitlab

import (
	"reflect"
	"testing"

	gitlab "github.com/xanzy/go-gitlab"
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
	comments := []*gitlab.Note{
		&gitlab.Note{
			ID:   371748792,
			Body: "comment 1",
		},
		&gitlab.Note{
			ID:   371765743,
			Body: "comment 2",
		},
	}
	testCases := []struct {
		config   Config
		number   int
		ok       bool
		comments []*gitlab.Note
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
