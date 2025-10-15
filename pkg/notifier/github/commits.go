package github

import (
	"context"
	"errors"

	"github.com/google/go-github/v74/github"
)

// CommitsService handles communication with the commits related
// methods of GitHub API
type CommitsService service

func (g *CommitsService) PRNumber(ctx context.Context, sha string) (int, error) {
	prs, _, err := g.client.API.PullRequestsListPullRequestsWithCommit(ctx, sha, &github.ListOptions{})
	if err != nil {
		return 0, err
	}
	if len(prs) == 0 {
		return 0, errors.New("associated pull request isn't found")
	}
	return prs[0].GetNumber(), nil
}
