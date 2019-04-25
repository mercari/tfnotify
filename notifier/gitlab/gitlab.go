package gitlab

import (
	"fmt"

	gitlab "github.com/xanzy/go-gitlab"
)

// API is GitLab API interface
type API interface {
	CreateMergeRequestNote(mergeRequest int, opt *gitlab.CreateMergeRequestNoteOptions, options ...gitlab.OptionFunc) (*gitlab.Note, *gitlab.Response, error)
	DeleteMergeRequestNote(mergeRequest, note int, options ...gitlab.OptionFunc) (*gitlab.Response, error)
	ListMergeRequestNotes(mergeRequest int, opt *gitlab.ListMergeRequestNotesOptions, options ...gitlab.OptionFunc) ([]*gitlab.Note, *gitlab.Response, error)
	PostCommitComment(sha string, opt *gitlab.PostCommitCommentOptions, options ...gitlab.OptionFunc) (*gitlab.CommitComment, *gitlab.Response, error)
	ListCommits(opt *gitlab.ListCommitsOptions, options ...gitlab.OptionFunc) ([]*gitlab.Commit, *gitlab.Response, error)
}

// GitLab represents the attribute information necessary for requesting GitLab API
type GitLab struct {
	*gitlab.Client
	namespace, project string
}

// CreateMergeRequestNote is a wrapper of https://godoc.org/github.com/xanzy/go-gitlab#NotesService.CreateMergeRequestNote
func (g *GitLab) CreateMergeRequestNote(mergeRequest int, opt *gitlab.CreateMergeRequestNoteOptions, options ...gitlab.OptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return g.Client.Notes.CreateMergeRequestNote(fmt.Sprintf("%s/%s", g.namespace, g.project), mergeRequest, opt, options...)
}

// DeleteMergeRequestNote is a wrapper of https://godoc.org/github.com/xanzy/go-gitlab#NotesService.DeleteMergeRequestNote
func (g *GitLab) DeleteMergeRequestNote(mergeRequest, note int, options ...gitlab.OptionFunc) (*gitlab.Response, error) {
	return g.Client.Notes.DeleteMergeRequestNote(fmt.Sprintf("%s/%s", g.namespace, g.project), mergeRequest, note, options...)
}

// ListMergeRequestNotes is a wrapper of https://godoc.org/github.com/xanzy/go-gitlab#NotesService.ListMergeRequestNotes
func (g *GitLab) ListMergeRequestNotes(mergeRequest int, opt *gitlab.ListMergeRequestNotesOptions, options ...gitlab.OptionFunc) ([]*gitlab.Note, *gitlab.Response, error) {
	return g.Client.Notes.ListMergeRequestNotes(fmt.Sprintf("%s/%s", g.namespace, g.project), mergeRequest, opt, options...)
}

// PostCommitComment is a wrapper of https://godoc.org/github.com/xanzy/go-gitlab#CommitsService.PostCommitComment
func (g *GitLab) PostCommitComment(sha string, opt *gitlab.PostCommitCommentOptions, options ...gitlab.OptionFunc) (*gitlab.CommitComment, *gitlab.Response, error) {
	return g.Client.Commits.PostCommitComment(fmt.Sprintf("%s/%s", g.namespace, g.project), sha, opt, options...)
}

// ListCommits is a wrapper of https://godoc.org/github.com/xanzy/go-gitlab#CommitsService.ListCommits
func (g *GitLab) ListCommits(opt *gitlab.ListCommitsOptions, options ...gitlab.OptionFunc) ([]*gitlab.Commit, *gitlab.Response, error) {
	return g.Client.Commits.ListCommits(fmt.Sprintf("%s/%s", g.namespace, g.project), opt, options...)
}
