package github

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/go-github/github"
)

// CommentService handles communication with the comment related
// methods of GitHub API
type CommentService service

// PostOptions specifies the optional parameters to post comments to a pull request
type PostOptions struct {
	Number   int
	Revision string
}

// Post posts comment
func (g *CommentService) Post(ctx context.Context, body string, opt PostOptions) error {
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
	return fmt.Errorf("github.comment.post: Number or Revision is required")
}

// List lists comments on GitHub issues/pull requests
func (g *CommentService) List(ctx context.Context, number int) ([]*github.IssueComment, error) {
	comments, _, err := g.client.API.IssuesListComments(
		ctx,
		number,
		&github.IssueListCommentsOptions{},
	)
	return comments, err
}

// Delete deletes comment on GitHub issues/pull requests
func (g *CommentService) Delete(ctx context.Context, id int) error {
	_, err := g.client.API.IssuesDeleteComment(
		ctx,
		int64(id),
	)
	return err
}

// DeleteDuplicates deletes duplicate comments containing arbitrary character strings
func (g *CommentService) DeleteDuplicates(ctx context.Context, title string) {
	var ids []int64
	comments := g.getDuplicates(ctx, title)
	for _, comment := range comments {
		ids = append(ids, *comment.ID)
	}
	for _, id := range ids {
		// don't handle error
		g.client.Comment.Delete(ctx, int(id))
	}
}

func (g *CommentService) getDuplicates(ctx context.Context, title string) []*github.IssueComment {
	var dup []*github.IssueComment
	re := regexp.MustCompile(`(?m)^(\n+)?` + title + `( +.*)?\n+` + g.client.Config.PR.Message + `\n+`)

	comments, _ := g.client.Comment.List(ctx, g.client.Config.PR.Number)
	for _, comment := range comments {
		if re.MatchString(*comment.Body) {
			dup = append(dup, comment)
		}
	}

	return dup
}
