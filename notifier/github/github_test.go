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
	FakePullRequestsCreateComment func(ctx context.Context, number int, comment *github.PullRequestComment) (*github.PullRequestComment, *github.Response, error)
	FakePullRequestsDeleteComment func(ctx context.Context, commentID int64) (*github.Response, error)
	FakePullRequestsListComments  func(ctx context.Context, number int, opt *github.PullRequestListCommentsOptions) ([]*github.PullRequestComment, *github.Response, error)
	FakeRepositoriesCreateComment func(ctx context.Context, sha string, comment *github.RepositoryComment) (*github.RepositoryComment, *github.Response, error)
	FakeRepositoriesListCommits   func(ctx context.Context, opt *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error)
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

func (g *fakeAPI) PullRequestsCreateComment(ctx context.Context, number int, comment *github.PullRequestComment) (*github.PullRequestComment, *github.Response, error) {
	return g.FakePullRequestsCreateComment(ctx, number, comment)
}

func (g *fakeAPI) PullRequestsDeleteComment(ctx context.Context, commentID int64) (*github.Response, error) {
	return g.FakePullRequestsDeleteComment(ctx, commentID)
}

func (g *fakeAPI) PullRequestsListComments(ctx context.Context, number int, opt *github.PullRequestListCommentsOptions) ([]*github.PullRequestsComment, *github.Response, error) {
	return g.FakePullRequestsListComments(ctx, number, opt)
}

func (g *fakeAPI) RepositoriesCreateComment(ctx context.Context, sha string, comment *github.RepositoryComment) (*github.RepositoryComment, *github.Response, error) {
	return g.FakeRepositoriesCreateComment(ctx, sha, comment)
}

func (g *fakeAPI) RepositoriesListCommits(ctx context.Context, opt *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
	return g.FakeRepositoriesListCommits(ctx, opt)
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
				&github.IssueComment{
					ID:   github.Int64(371748792),
					Body: github.String("comment 1"),
				},
				&github.IssueComment{
					ID:   github.Int64(371765743),
					Body: github.String("comment 2"),
				},
			}
			return comments, nil, nil
		},
		FakePullRequestsCreateComment: func(ctx context.Context, number int, comment *github.PullRequestComment) (*github.PullRequestComment, *github.Response, error) {
			return &github.PullRequestComment{
				ID:   github.Int64(471748792),
				Body: github.String("comment 1"),
			}, nil, nil
		},
		FakePullRequestsDeleteComment: func(ctx context.Context, commentID int64) (*github.Response, error) {
			return nil, nil
		},
		FakePullRequestsListComments: func(ctx context.Context, number int, opt *github.PullRequestListCommentsOptions) ([]*github.PullRequestComment, *github.Response, error) {
			var comments []*github.PullRequestComment
			comments = []*github.PullRequestComment{
				&github.PullRequestComment{
					ID:   github.Int64(471748792),
					Body: github.String("comment 1"),
				},
				&github.PullRequestComment{
					ID:   github.Int64(471765743),
					Body: github.String("comment 2"),
				},
			}
			return comments, nil, nil
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
				&github.RepositoryCommit{
					SHA: github.String("04e0917e448b662c2b16330fad50e97af16ff27a"),
				},
				&github.RepositoryCommit{
					SHA: github.String("04e0917e448b662c2b16330fad50e97af16ff27b"),
				},
				&github.RepositoryCommit{
					SHA: github.String("04e0917e448b662c2b16330fad50e97af16ff27c"),
				},
			}
			return commits, nil, nil
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
