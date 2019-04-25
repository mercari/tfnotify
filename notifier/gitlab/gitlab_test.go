package gitlab

import (
	"github.com/mercari/tfnotify/terraform"
	gitlab "github.com/xanzy/go-gitlab"
)

type fakeAPI struct {
	API
	FakeCreateMergeRequestNote func(mergeRequest int, opt *gitlab.CreateMergeRequestNoteOptions, options ...gitlab.OptionFunc) (*gitlab.Note, *gitlab.Response, error)
	FakeDeleteMergeRequestNote func(mergeRequest, note int, options ...gitlab.OptionFunc) (*gitlab.Response, error)
	FakeListMergeRequestNotes  func(mergeRequest int, opt *gitlab.ListMergeRequestNotesOptions, options ...gitlab.OptionFunc) ([]*gitlab.Note, *gitlab.Response, error)
	FakePostCommitComment      func(sha string, opt *gitlab.PostCommitCommentOptions, options ...gitlab.OptionFunc) (*gitlab.CommitComment, *gitlab.Response, error)
	FakeListCommits            func(opt *gitlab.ListCommitsOptions, options ...gitlab.OptionFunc) ([]*gitlab.Commit, *gitlab.Response, error)
}

func (g *fakeAPI) CreateMergeRequestNote(mergeRequest int, opt *gitlab.CreateMergeRequestNoteOptions, options ...gitlab.OptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return g.FakeCreateMergeRequestNote(mergeRequest, opt, options...)
}

func (g *fakeAPI) DeleteMergeRequestNote(mergeRequest, note int, options ...gitlab.OptionFunc) (*gitlab.Response, error) {
	return g.FakeDeleteMergeRequestNote(mergeRequest, note, options...)
}

func (g *fakeAPI) ListMergeRequestNotes(mergeRequest int, opt *gitlab.ListMergeRequestNotesOptions, options ...gitlab.OptionFunc) ([]*gitlab.Note, *gitlab.Response, error) {
	return g.FakeListMergeRequestNotes(mergeRequest, opt, options...)
}

func (g *fakeAPI) PostCommitComment(sha string, opt *gitlab.PostCommitCommentOptions, options ...gitlab.OptionFunc) (*gitlab.CommitComment, *gitlab.Response, error) {
	return g.FakePostCommitComment(sha, opt, options...)
}

func (g *fakeAPI) ListCommits(opt *gitlab.ListCommitsOptions, options ...gitlab.OptionFunc) ([]*gitlab.Commit, *gitlab.Response, error) {
	return g.FakeListCommits(opt, options...)
}

func newFakeAPI() fakeAPI {
	return fakeAPI{
		FakeCreateMergeRequestNote: func(mergeRequest int, opt *gitlab.CreateMergeRequestNoteOptions, options ...gitlab.OptionFunc) (*gitlab.Note, *gitlab.Response, error) {
			return &gitlab.Note{
				ID:   371748792,
				Body: "comment 1",
			}, nil, nil
		},
		FakeDeleteMergeRequestNote: func(mergeRequest, note int, options ...gitlab.OptionFunc) (*gitlab.Response, error) {
			return nil, nil
		},
		FakeListMergeRequestNotes: func(mergeRequest int, opt *gitlab.ListMergeRequestNotesOptions, options ...gitlab.OptionFunc) ([]*gitlab.Note, *gitlab.Response, error) {
			var comments []*gitlab.Note
			comments = []*gitlab.Note{
				&gitlab.Note{
					ID:   371748792,
					Body: "comment 1",
				},
				&gitlab.Note{
					ID:   371765743,
					Body: "comment 2",
				},
			}
			return comments, nil, nil
		},
		FakePostCommitComment: func(sha string, opt *gitlab.PostCommitCommentOptions, options ...gitlab.OptionFunc) (*gitlab.CommitComment, *gitlab.Response, error) {
			return &gitlab.CommitComment{
				Note: "comment 1",
			}, nil, nil
		},
		FakeListCommits: func(opt *gitlab.ListCommitsOptions, options ...gitlab.OptionFunc) ([]*gitlab.Commit, *gitlab.Response, error) {
			var commits []*gitlab.Commit
			commits = []*gitlab.Commit{
				&gitlab.Commit{
					ID: "04e0917e448b662c2b16330fad50e97af16ff27a",
				},
				&gitlab.Commit{
					ID: "04e0917e448b662c2b16330fad50e97af16ff27b",
				},
				&gitlab.Commit{
					ID: "04e0917e448b662c2b16330fad50e97af16ff27c",
				},
			}
			return commits, nil, nil
		},
	}
}

func newFakeConfig() Config {
	return Config{
		Token:     "token",
		NameSpace: "owner",
		Project:   "repo",
		MR: MergeRequest{
			Revision: "abcd",
			Number:   1,
			Message:  "message",
		},
		Parser:   terraform.NewPlanParser(),
		Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
	}
}
