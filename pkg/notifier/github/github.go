package github

import (
	"context"

	"github.com/google/go-github/v74/github"
)

// API is GitHub API interface
type API interface {
	IssuesCreateComment(ctx context.Context, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
	IssuesEditComment(ctx context.Context, commentID int64, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
	IssuesListLabels(ctx context.Context, number int, opt *github.ListOptions) ([]*github.Label, *github.Response, error)
	IssuesAddLabels(ctx context.Context, number int, labels []string) ([]*github.Label, *github.Response, error)
	IssuesRemoveLabel(ctx context.Context, number int, label string) (*github.Response, error)
	IssuesUpdateLabel(ctx context.Context, label, color string) (*github.Label, *github.Response, error)
	RepositoriesCreateComment(ctx context.Context, sha string, comment *github.RepositoryComment) (*github.RepositoryComment, *github.Response, error)
	PullRequestsListPullRequestsWithCommit(ctx context.Context, sha string, opt *github.ListOptions) ([]*github.PullRequest, *github.Response, error)
}

// GitHub represents the attribute information necessary for requesting GitHub API
type GitHub struct {
	*github.Client

	owner string
	repo  string
}

// IssuesCreateComment is a wrapper of https://godoc.org/github.com/google/go-github/github#IssuesService.CreateComment
func (g *GitHub) IssuesCreateComment(ctx context.Context, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	return g.Issues.CreateComment(ctx, g.owner, g.repo, number, comment)
}

func (g *GitHub) IssuesEditComment(ctx context.Context, commentID int64, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	return g.Issues.EditComment(ctx, g.owner, g.repo, commentID, comment)
}

// IssuesAddLabels is a wrapper of https://godoc.org/github.com/google/go-github/github#IssuesService.AddLabelsToIssue
func (g *GitHub) IssuesAddLabels(ctx context.Context, number int, labels []string) ([]*github.Label, *github.Response, error) {
	return g.Issues.AddLabelsToIssue(ctx, g.owner, g.repo, number, labels)
}

// IssuesListLabels is a wrapper of https://godoc.org/github.com/google/go-github/github#IssuesService.ListLabelsByIssue
func (g *GitHub) IssuesListLabels(ctx context.Context, number int, opt *github.ListOptions) ([]*github.Label, *github.Response, error) {
	return g.Issues.ListLabelsByIssue(ctx, g.owner, g.repo, number, opt)
}

// IssuesRemoveLabel is a wrapper of https://godoc.org/github.com/google/go-github/github#IssuesService.RemoveLabelForIssue
func (g *GitHub) IssuesRemoveLabel(ctx context.Context, number int, label string) (*github.Response, error) {
	return g.Issues.RemoveLabelForIssue(ctx, g.owner, g.repo, number, label)
}

// IssuesUpdateLabel is a wrapper of https://pkg.go.dev/github.com/google/go-github/github#IssuesService.EditLabel
func (g *GitHub) IssuesUpdateLabel(ctx context.Context, label, color string) (*github.Label, *github.Response, error) {
	return g.Issues.EditLabel(ctx, g.owner, g.repo, label, &github.Label{
		Color: &color,
	})
}

// RepositoriesCreateComment is a wrapper of https://godoc.org/github.com/google/go-github/github#RepositoriesService.CreateComment
func (g *GitHub) RepositoriesCreateComment(ctx context.Context, sha string, comment *github.RepositoryComment) (*github.RepositoryComment, *github.Response, error) {
	return g.Repositories.CreateComment(ctx, g.owner, g.repo, sha, comment)
}

func (g *GitHub) PullRequestsListPullRequestsWithCommit(ctx context.Context, sha string, opt *github.ListOptions) ([]*github.PullRequest, *github.Response, error) {
	return g.PullRequests.ListPullRequestsWithCommit(ctx, g.owner, g.repo, sha, opt)
}
