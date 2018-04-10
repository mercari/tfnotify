package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// CI represents a common information obtained from all CI platforms
type CI struct {
	PR  PullRequest
	URL string
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Revision string
	Number   int
}

var ciRegexp = regexp.MustCompile(`[1-9]\d*$`)

func circleci() (ci CI, err error) {
	ci.PR.Number = 0
	ci.PR.Revision = os.Getenv("CIRCLE_SHA1")
	ci.URL = os.Getenv("CIRCLE_BUILD_URL")
	pr := os.Getenv("CIRCLE_PULL_REQUEST")
	if pr == "" {
		pr = os.Getenv("CI_PULL_REQUEST")
	}
	if pr == "" {
		pr = os.Getenv("CIRCLE_PR_NUMBER")
	}
	if pr == "" {
		return ci, nil
	}
	ci.PR.Number, err = strconv.Atoi(ciRegexp.FindString(pr))
	if err != nil {
		return ci, fmt.Errorf("%v: cannot get env", pr)
	}
	return ci, nil
}

func travisci() (ci CI, err error) {
	ci.PR.Revision = os.Getenv("TRAVIS_PULL_REQUEST_SHA")
	ci.PR.Number, err = strconv.Atoi(os.Getenv("TRAVIS_PULL_REQUEST"))
	return ci, err
}
