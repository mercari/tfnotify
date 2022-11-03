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
func (g *CommentService) Post(body string, opt PostOptions) error {
	if opt.Number != 0 {
		_, _, err := g.client.API.IssuesCreateComment(
			context.Background(),
			opt.Number,
			&github.IssueComment{Body: &body},
		)
		return err
	}
	if opt.Revision != "" {
		_, _, err := g.client.API.RepositoriesCreateComment(
			context.Background(),
			opt.Revision,
			&github.RepositoryComment{Body: &body},
		)
		return err
	}
	return fmt.Errorf("github.comment.post: Number or Revision is required")
}

// List lists comments on GitHub issues/pull requests
func (g *CommentService) List(number int) ([]*github.IssueComment, error) {
	comments, _, err := g.client.API.IssuesListComments(
		context.Background(),
		number,
		&github.IssueListCommentsOptions{},
	)
	return comments, err
}

// Delete deletes comment on GitHub issues/pull requests
func (g *CommentService) Delete(id int) error {
	_, err := g.client.API.IssuesDeleteComment(
		context.Background(),
		int64(id),
	)
	return err
}

// DeleteDuplicates deletes duplicate comments containing arbitrary character strings
func (g *CommentService) DeleteDuplicates(title string) {
	var ids []int64
	comments := g.getDuplicates(title)
	for _, comment := range comments {
		ids = append(ids, *comment.ID)
	}
	for _, id := range ids {
		// don't handle error
		g.client.Comment.Delete(int(id))
	}
}

func (g *CommentService) getDuplicates(title string) []*github.IssueComment {
	var dup []*github.IssueComment
	re := regexp.MustCompile(`(?m)^(\n+)?` + title + `( +.*)?\n+` + g.client.Config.PR.Message + `\n+`)

	comments, _ := g.client.Comment.List(g.client.Config.PR.Number)
	for _, comment := range comments {
		if re.MatchString(*comment.Body) {
			dup = append(dup, comment)
		}
	}

	return dup
}
