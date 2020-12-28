package github

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/mercari/tfnotify/terraform"
)

type fakeAPI struct {
	API
	FakeIssuesCreateComment       func(ctx context.Context, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
	FakeIssuesDeleteComment       func(ctx context.Context, commentID int64) (*github.Response, error)
	FakeIssuesListComments        func(ctx context.Context, number int, opt *github.IssueListCommentsOptions) ([]*github.IssueComment, *github.Response, error)
	FakeIssuesListLabels          func(ctx context.Context, number int, opts *github.ListOptions) ([]*github.Label, *github.Response, error)
	FakeIssuesAddLabels           func(ctx context.Context, number int, labels []string) ([]*github.Label, *github.Response, error)
	FakeIssuesRemoveLabel         func(ctx context.Context, number int, label string) (*github.Response, error)
	FakeRepositoriesCreateComment func(ctx context.Context, sha string, comment *github.RepositoryComment) (*github.RepositoryComment, *github.Response, error)
	FakeRepositoriesListCommits   func(ctx context.Context, opt *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error)
	FakeRepositoriesGetCommit     func(ctx context.Context, sha string) (*github.RepositoryCommit, *github.Response, error)
}

func (g *fakeAPI) IssuesCreateComment(ctx context.Context, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	return g.FakeIssuesCreateComment(ctx, number, comment)
}

func (g *fakeAPI) IssuesDeleteComment(ctx context.Context, commentID int64) (*github.Response, error) {
	return g.FakeIssuesDeleteComment(ctx, commentID)
}

func (g *fakeAPI) IssuesListComments(ctx context.Context, number int, opt *github.IssueListCommentsOptions) ([]*github.IssueComment, *github.Response, error) {
	return g.FakeIssuesListComments(ctx, number, opt)
}

func (g *fakeAPI) IssuesListLabels(ctx context.Context, number int, opt *github.ListOptions) ([]*github.Label, *github.Response, error) {
	return g.FakeIssuesListLabels(ctx, number, opt)
}

func (g *fakeAPI) IssuesAddLabels(ctx context.Context, number int, labels []string) ([]*github.Label, *github.Response, error) {
	return g.FakeIssuesAddLabels(ctx, number, labels)
}

func (g *fakeAPI) IssuesRemoveLabel(ctx context.Context, number int, label string) (*github.Response, error) {
	return g.FakeIssuesRemoveLabel(ctx, number, label)
}

func (g *fakeAPI) RepositoriesCreateComment(ctx context.Context, sha string, comment *github.RepositoryComment) (*github.RepositoryComment, *github.Response, error) {
	return g.FakeRepositoriesCreateComment(ctx, sha, comment)
}

func (g *fakeAPI) RepositoriesListCommits(ctx context.Context, opt *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
	return g.FakeRepositoriesListCommits(ctx, opt)
}

func (g *fakeAPI) RepositoriesGetCommit(ctx context.Context, sha string) (*github.RepositoryCommit, *github.Response, error) {
	return g.FakeRepositoriesGetCommit(ctx, sha)
}

func newFakeAPI() fakeAPI {
	return fakeAPI{
		FakeIssuesCreateComment: func(ctx context.Context, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
			return &github.IssueComment{
				ID:   github.Int64(371748792),
				Body: github.String("comment 1"),
			}, nil, nil
		},
		FakeIssuesDeleteComment: func(ctx context.Context, commentID int64) (*github.Response, error) {
			return nil, nil
		},
		FakeIssuesListComments: func(ctx context.Context, number int, opt *github.IssueListCommentsOptions) ([]*github.IssueComment, *github.Response, error) {
			var comments []*github.IssueComment
			comments = []*github.IssueComment{
				{
					ID:   github.Int64(371748792),
					Body: github.String("comment 1"),
				},
				{
					ID:   github.Int64(371765743),
					Body: github.String("comment 2"),
				},
			}
			return comments, nil, nil
		},
		FakeIssuesListLabels: func(ctx context.Context, number int, opts *github.ListOptions) ([]*github.Label, *github.Response, error) {
			labels := []*github.Label{
				{
					ID:   github.Int64(371748792),
					Name: github.String("label 1"),
				},
				{
					ID:   github.Int64(371765743),
					Name: github.String("label 2"),
				},
			}
			return labels, nil, nil
		},
		FakeIssuesAddLabels: func(ctx context.Context, number int, labels []string) ([]*github.Label, *github.Response, error) {
			return nil, nil, nil
		},
		FakeIssuesRemoveLabel: func(ctx context.Context, number int, label string) (*github.Response, error) {
			return nil, nil
		},
		FakeRepositoriesCreateComment: func(ctx context.Context, sha string, comment *github.RepositoryComment) (*github.RepositoryComment, *github.Response, error) {
			return &github.RepositoryComment{
				ID:       github.Int64(28427394),
				CommitID: github.String("04e0917e448b662c2b16330fad50e97af16ff27a"),
				Body:     github.String("comment 1"),
			}, nil, nil
		},
		FakeRepositoriesListCommits: func(ctx context.Context, opt *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
			var commits []*github.RepositoryCommit
			commits = []*github.RepositoryCommit{
				{
					SHA: github.String("04e0917e448b662c2b16330fad50e97af16ff27a"),
				},
				{
					SHA: github.String("04e0917e448b662c2b16330fad50e97af16ff27b"),
				},
				{
					SHA: github.String("04e0917e448b662c2b16330fad50e97af16ff27c"),
				},
			}
			return commits, nil, nil
		},
		FakeRepositoriesGetCommit: func(ctx context.Context, sha string) (*github.RepositoryCommit, *github.Response, error) {
			return &github.RepositoryCommit{
				SHA: github.String(sha),
				Commit: &github.Commit{
					Message: github.String(sha),
				},
			}, nil, nil
		},
	}
}

func newFakeConfig() Config {
	return Config{
		Token: "token",
		Owner: "owner",
		Repo:  "repo",
		PR: PullRequest{
			Revision: "abcd",
			Number:   1,
			Message:  "message",
		},
		Parser:   terraform.NewPlanParser(),
		Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
	}
}
