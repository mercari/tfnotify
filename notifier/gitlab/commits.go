package gitlab

import (
	"errors"

	gitlab "github.com/xanzy/go-gitlab"
)

// CommitsService handles communication with the commits related
// methods of GitLab API
type CommitsService service

// List lists commits on a repository
func (g *CommitsService) List(revision string) ([]string, error) {
	if revision == "" {
		return []string{}, errors.New("no revision specified")
	}
	var s []string
	commits, _, err := g.client.API.ListCommits(
		&gitlab.ListCommitsOptions{},
	)
	if err != nil {
		return s, err
	}
	for _, commit := range commits {
		s = append(s, commit.ID)
	}
	return s, nil
}

// lastOne returns the hash of the previous commit of the given commit
func (g *CommitsService) lastOne(commits []string, revision string) (string, error) {
	if revision == "" {
		return "", errors.New("no revision specified")
	}
	if len(commits) == 0 {
		return "", errors.New("no commits")
	}

	return commits[1], nil
}
