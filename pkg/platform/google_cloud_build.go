package platform

import (
	"fmt"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/go-ci-env/v3/cienv"
)

type GoogleCloudBuild struct {
	getenv func(string) string
}

func NewGoogleCloudBuild(param *cienv.Param) *GoogleCloudBuild {
	if param == nil || param.Getenv == nil {
		return &GoogleCloudBuild{
			getenv: os.Getenv,
		}
	}
	return &GoogleCloudBuild{
		getenv: param.Getenv,
	}
}

func (cb *GoogleCloudBuild) ID() string {
	return "google-cloud-build"
}

func (cb *GoogleCloudBuild) Match() bool {
	return cb.getenv("GOOGLE_CLOUD_BUILD") != ""
}

func (cb *GoogleCloudBuild) RepoOwner() string {
	return ""
}

func (cb *GoogleCloudBuild) RepoName() string {
	return ""
}

func (cb *GoogleCloudBuild) Ref() string {
	return ""
}

func (cb *GoogleCloudBuild) Tag() string {
	return ""
}

func (cb *GoogleCloudBuild) Branch() string {
	return ""
}

func (cb *GoogleCloudBuild) PRBaseBranch() string {
	return ""
}

func (cb *GoogleCloudBuild) SHA() string {
	return cb.getenv("COMMIT_SHA")
}

func (cb *GoogleCloudBuild) IsPR() bool {
	return cb.getenv("_PR_NUMBER") != ""
}

func (cb *GoogleCloudBuild) PRNumber() (int, error) {
	pr := cb.getenv("_PR_NUMBER")
	if pr == "" {
		return 0, nil
	}
	b, err := strconv.Atoi(pr)
	if err == nil {
		return b, nil
	}
	return 0, fmt.Errorf("_PR_NUMBER is invalid. It failed to parse _PR_NUMBER as an integer: %w", err)
}

func (cb *GoogleCloudBuild) JobURL() string {
	region := cb.getenv("_REGION")
	if region == "" {
		region = "global"
	}
	return fmt.Sprintf(
		"https://console.cloud.google.com/cloud-build/builds;region=%s/%s?project=%s",
		region,
		cb.getenv("BUILD_ID"),
		cb.getenv("PROJECT_ID"),
	)
}
