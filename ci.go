package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultCloudBuildRegion = "global"
)

var (
	// https://help.github.com/en/actions/reference/events-that-trigger-workflows#pull-request-event-pull_request
	githubActionsPRRefRegexp = regexp.MustCompile(`refs/pull/\d+/merge`)
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
	re := regexp.MustCompile(`[1-9]\d*$`)
	ci.PR.Number, err = strconv.Atoi(re.FindString(pr))
	if err != nil {
		return ci, fmt.Errorf("%v: cannot get env", pr)
	}
	return ci, nil
}

func travisci() (ci CI, err error) {
	ci.URL = os.Getenv("TRAVIS_BUILD_WEB_URL")
	prNumber := os.Getenv("TRAVIS_PULL_REQUEST")
	if prNumber == "false" {
		ci.PR.Number = 0
		ci.PR.Revision = os.Getenv("TRAVIS_COMMIT")
		return ci, nil
	}
	ci.PR.Revision = os.Getenv("TRAVIS_PULL_REQUEST_SHA")
	ci.PR.Number, err = strconv.Atoi(prNumber)
	return ci, err
}

func codebuild() (ci CI, err error) {
	ci.PR.Number = 0
	ci.PR.Revision = os.Getenv("CODEBUILD_RESOLVED_SOURCE_VERSION")
	ci.URL = os.Getenv("CODEBUILD_BUILD_URL")
	sourceVersion := os.Getenv("CODEBUILD_SOURCE_VERSION")
	if sourceVersion == "" {
		return ci, nil
	}
	if !strings.HasPrefix(sourceVersion, "pr/") {
		return ci, nil
	}
	pr := strings.Replace(sourceVersion, "pr/", "", 1)
	if pr == "" {
		return ci, nil
	}
	ci.PR.Number, err = strconv.Atoi(pr)
	return ci, err
}

func teamcity() (ci CI, err error) {
	ci.PR.Revision = os.Getenv("BUILD_VCS_NUMBER")
	ci.PR.Number, err = strconv.Atoi(os.Getenv("BUILD_NUMBER"))
	return ci, err
}

func drone() (ci CI, err error) {
	ci.PR.Number = 0
	ci.PR.Revision = os.Getenv("DRONE_COMMIT_SHA")
	ci.URL = os.Getenv("DRONE_BUILD_LINK")
	pr := os.Getenv("DRONE_PULL_REQUEST")
	if pr == "" {
		return ci, nil
	}
	ci.PR.Number, err = strconv.Atoi(pr)
	return ci, err
}

func jenkins() (ci CI, err error) {
	ci.PR.Number = 0
	ci.PR.Revision = os.Getenv("GIT_COMMIT")
	if ci.PR.Revision == "" {
		ci.PR.Revision = os.Getenv("gitlabBefore")
	}
	ci.URL = os.Getenv("BUILD_URL")
	pr := os.Getenv("PULL_REQUEST_NUMBER")
	if pr == "" {
		pr = os.Getenv("gitlabMergeRequestIid")
	}
	if pr == "" {
		pr = os.Getenv("PULL_REQUEST_URL")
	}
	if pr == "" {
		return ci, nil
	}
	re := regexp.MustCompile(`[1-9]\d*$`)
	ci.PR.Number, err = strconv.Atoi(re.FindString(pr))
	if err != nil {
		return ci, fmt.Errorf("%v: Invalid PullRequest number or MergeRequest ID", pr)
	}
	return ci, err
}

func gitlabci() (ci CI, err error) {
	ci.PR.Number = 0
	ci.PR.Revision = os.Getenv("CI_COMMIT_SHA")
	ci.URL = os.Getenv("CI_JOB_URL")
	pr := os.Getenv("CI_MERGE_REQUEST_IID")
	if pr == "" {
		refPath := os.Getenv("CI_MERGE_REQUEST_REF_PATH")
		rep := regexp.MustCompile(`refs/merge-requests/\d*/head`)
		if rep.MatchString(refPath) {
			strLen := strings.Split(refPath, "/")
			pr = strLen[2]
		}
	}
	if pr == "" {
		return ci, nil
	}
	ci.PR.Number, err = strconv.Atoi(pr)
	return ci, err
}

func githubActions() (ci CI, err error) {
	ci.URL = fmt.Sprintf(
		"%s/%s/actions/runs/%s",
		os.Getenv("GITHUB_SERVER_URL"),
		os.Getenv("GITHUB_REPOSITORY"),
		os.Getenv("GITHUB_RUN_ID"),
	)
	ci.PR.Revision = os.Getenv("GITHUB_SHA")
	ci.PR.Number = 0

	if githubActionsPRRefRegexp.MatchString(os.Getenv("GITHUB_REF")) {
		s := strings.Split(os.Getenv("GITHUB_REF"), "/")[2]
		pr, err := strconv.Atoi(s)
		if err != nil {
			return ci, err
		}

		ci.PR.Number = pr
	}

	return ci, err
}

func cloudbuild() (ci CI, err error) {
	ci.PR.Number = 0
	ci.PR.Revision = os.Getenv("COMMIT_SHA")

	region := os.Getenv("_REGION")
	if region == "" {
		region = defaultCloudBuildRegion
	}

	ci.URL = fmt.Sprintf(
		"https://console.cloud.google.com/cloud-build/builds;region=%s/%s?project=%s",
		region,
		os.Getenv("BUILD_ID"),
		os.Getenv("PROJECT_ID"),
	)
	pr := os.Getenv("_PR_NUMBER")
	if pr == "" {
		return ci, nil
	}
	ci.PR.Number, err = strconv.Atoi(pr)
	return ci, err
}
