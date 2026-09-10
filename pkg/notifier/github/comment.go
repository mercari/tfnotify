package github

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/go-github/v74/github"
	"github.com/shurcooL/githubv4"
)

// CommentService handles communication with the comment related
// methods of GitHub API
type CommentService service

// githubMaxCommentLength is GitHub's hard limit on a comment body. Posting a
// longer body fails with a 422, which fails CI even though the underlying
// plan/apply succeeded.
const githubMaxCommentLength = 65536

const truncationNotice = "\n\n:warning: **This comment was truncated by tfnotify because it exceeded GitHub's " +
	"65536-character limit.** :warning:\n"

// truncateComment keeps a comment body within githubMaxCommentLength. The
// template already does a best-effort truncation per code block, but a
// consolidated comment can assemble many blocks and still overflow, so this is
// a final safety net on every GitHub write. It keeps the head of the body (the
// summary, links and warnings) and preserves the trailing github-comment
// metadata line, which is what tfnotify uses to find and patch its own comment.
func truncateComment(body string) string {
	if utf8.RuneCountInString(body) <= githubMaxCommentLength {
		return body
	}

	content, suffix := body, ""
	if i := strings.LastIndex(body, "\n<!-- github-comment: "); i != -1 && strings.HasSuffix(body, " -->") {
		content, suffix = body[:i], body[i:]
	}

	budget := githubMaxCommentLength - utf8.RuneCountInString(suffix) - utf8.RuneCountInString(truncationNotice)
	if budget < 0 {
		budget = 0
	}
	if utf8.RuneCountInString(content) > budget {
		content = string([]rune(content)[:budget])
	}

	return content + truncationNotice + suffix
}

// PostOptions specifies the optional parameters to post comments to a pull request
type PostOptions struct {
	Number   int
	Revision string
}

// Post posts comment
func (g *CommentService) Post(ctx context.Context, body string, opt *PostOptions) error {
	body = truncateComment(body)
	if opt.Number != 0 {
		_, _, err := g.client.API.IssuesCreateComment(
			ctx,
			opt.Number,
			&github.IssueComment{Body: &body},
		)
		return err
	}
	if opt.Revision != "" {
		_, _, err := g.client.API.RepositoriesCreateComment(
			ctx,
			opt.Revision,
			&github.RepositoryComment{Body: &body},
		)
		return err
	}
	return errors.New("github.comment.post: Number or Revision is required")
}

func (g *CommentService) Patch(ctx context.Context, body string, commentID int64) error {
	body = truncateComment(body)
	_, _, err := g.client.API.IssuesEditComment(
		ctx,
		commentID,
		&github.IssueComment{Body: &body},
	)
	return err
}

type IssueComment struct {
	DatabaseID  int
	Body        string
	IsMinimized bool
}

func (g *CommentService) List(ctx context.Context, owner, repo string, number int) ([]*IssueComment, error) {
	cmts, prErr := g.listPRComment(ctx, owner, repo, number)
	if prErr == nil {
		return cmts, nil
	}
	cmts, err := g.listIssueComment(ctx, owner, repo, number)
	if err == nil {
		return cmts, nil
	}
	return nil, fmt.Errorf("get pull request or issue comments: %w, %v", prErr, err) //nolint:errorlint
}

func (g *CommentService) listIssueComment(ctx context.Context, owner, repo string, number int) ([]*IssueComment, error) { //nolint:dupl
	// https://github.com/shurcooL/githubv4#pagination
	var q struct {
		Repository struct {
			Issue struct {
				Comments struct {
					Nodes    []*IssueComment
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"comments(first: 100, after: $commentsCursor)"` // 100 per page.
			} `graphql:"issue(number: $issueNumber)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}
	variables := map[string]any{
		"repositoryOwner": githubv4.String(owner),
		"repositoryName":  githubv4.String(repo),
		"issueNumber":     githubv4.Int(number),    //nolint:gosec
		"commentsCursor":  (*githubv4.String)(nil), // Null after argument to get first page.
	}

	var allComments []*IssueComment
	for {
		if err := g.client.v4Client.Query(ctx, &q, variables); err != nil {
			return nil, fmt.Errorf("list issue comments by GitHub API: %w", err)
		}
		allComments = append(allComments, q.Repository.Issue.Comments.Nodes...)
		if !q.Repository.Issue.Comments.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(q.Repository.Issue.Comments.PageInfo.EndCursor)
	}
	return allComments, nil
}

func (g *CommentService) listPRComment(ctx context.Context, owner, repo string, number int) ([]*IssueComment, error) { //nolint:dupl
	// https://github.com/shurcooL/githubv4#pagination
	var q struct {
		Repository struct {
			PullRequest struct {
				Comments struct {
					Nodes    []*IssueComment
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"comments(first: 100, after: $commentsCursor)"` // 100 per page.
			} `graphql:"pullRequest(number: $issueNumber)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}
	variables := map[string]any{
		"repositoryOwner": githubv4.String(owner),
		"repositoryName":  githubv4.String(repo),
		"issueNumber":     githubv4.Int(number),    //nolint:gosec
		"commentsCursor":  (*githubv4.String)(nil), // Null after argument to get first page.
	}

	var allComments []*IssueComment
	for {
		if err := g.client.v4Client.Query(ctx, &q, variables); err != nil {
			return nil, fmt.Errorf("list issue comments by GitHub API: %w", err)
		}
		allComments = append(allComments, q.Repository.PullRequest.Comments.Nodes...)
		if !q.Repository.PullRequest.Comments.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(q.Repository.PullRequest.Comments.PageInfo.EndCursor)
	}
	return allComments, nil
}
