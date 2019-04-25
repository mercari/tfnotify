package gitlab

import (
	"fmt"
	"regexp"

	gitlab "github.com/xanzy/go-gitlab"
)

// CommentService handles communication with the comment related
// methods of GitLab API
type CommentService service

// PostOptions specifies the optional parameters to post comments to a pull request
type PostOptions struct {
	Number   int
	Revision string
}

// Post posts comment
func (g *CommentService) Post(body string, opt PostOptions) error {
	if opt.Number != 0 {
		_, _, err := g.client.API.CreateMergeRequestNote(
			opt.Number,
			&gitlab.CreateMergeRequestNoteOptions{Body: gitlab.String(body)},
		)
		return err
	}
	if opt.Revision != "" {
		_, _, err := g.client.API.PostCommitComment(
			opt.Revision,
			&gitlab.PostCommitCommentOptions{Note: gitlab.String(body)},
		)
		return err
	}
	return fmt.Errorf("gitlab.comment.post: Number or Revision is required")
}

// List lists comments on GitLab merge requests
func (g *CommentService) List(number int) ([]*gitlab.Note, error) {
	comments, _, err := g.client.API.ListMergeRequestNotes(
		number,
		&gitlab.ListMergeRequestNotesOptions{},
	)
	return comments, err
}

// Delete deletes comment on GitLab merge requests
func (g *CommentService) Delete(note int) error {
	_, err := g.client.API.DeleteMergeRequestNote(
		g.client.Config.MR.Number,
		note,
	)
	return err
}

// DeleteDuplicates deletes duplicate comments containing arbitrary character strings
func (g *CommentService) DeleteDuplicates(title string) {
	var ids []int
	comments := g.getDuplicates(title)
	for _, comment := range comments {
		ids = append(ids, comment.ID)
	}
	for _, id := range ids {
		// don't handle error
		g.client.Comment.Delete(id)
	}
}

func (g *CommentService) getDuplicates(title string) []*gitlab.Note {
	var dup []*gitlab.Note
	re := regexp.MustCompile(`(?m)^(\n+)?` + title + `( +.*)?\n+` + g.client.Config.MR.Message + `\n+`)

	comments, _ := g.client.Comment.List(g.client.Config.MR.Number)
	for _, comment := range comments {
		if re.MatchString(comment.Body) {
			dup = append(dup, comment)
		}
	}

	return dup
}
