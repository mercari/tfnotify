package platform

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mercari/tfnotify/v1/pkg/config"
	"github.com/suzuki-shunsuke/go-ci-env/v3/cienv"
)

func Complement(cfg *config.Config) error {
	if cfg.RepoOwner != "" {
		cfg.CI.Owner = cfg.RepoOwner
	}
	if cfg.RepoName != "" {
		cfg.CI.Repo = cfg.RepoName
	}
	if err := complementWithCIEnv(&cfg.CI); err != nil {
		return fmt.Errorf("complement parameters with CI specific environment variables: %w", err)
	}

	if err := complementCIInfo(&cfg.CI); err != nil {
		return fmt.Errorf("complement parameters with ci-info's environment variables: %w", err)
	}

	return nil
}

func complementCIInfo(ci *config.CI) error {
	if ci.PRNumber <= 0 {
		// support suzuki-shunsuke/ci-info
		if prS := os.Getenv("CI_INFO_PR_NUMBER"); prS != "" {
			a, err := strconv.Atoi(prS)
			if err != nil {
				return fmt.Errorf("parse CI_INFO_PR_NUMBER %s: %w", prS, err)
			}
			ci.PRNumber = a
		}
	}
	return nil
}

func getLink(ciname string) string {
	switch ciname {
	case "circleci", "circle-ci":
		return os.Getenv("CIRCLE_BUILD_URL")
	case "codebuild":
		return os.Getenv("CODEBUILD_BUILD_URL")
	case "github-actions":
		return fmt.Sprintf(
			"%s/%s/actions/runs/%s",
			os.Getenv("GITHUB_SERVER_URL"),
			os.Getenv("GITHUB_REPOSITORY"),
			os.Getenv("GITHUB_RUN_ID"),
		)
	case "google-cloud-build":
		region := os.Getenv("_REGION")
		if region == "" {
			region = "global"
		}
		return fmt.Sprintf(
			"https://console.cloud.google.com/cloud-build/builds;region=%s/%s?project=%s",
			region,
			os.Getenv("BUILD_ID"),
			os.Getenv("PROJECT_ID"),
		)
	}
	return ""
}

func complementWithCIEnv(ci *config.CI) error {
	cienv.Add(func(param *cienv.Param) cienv.Platform {
		return NewGoogleCloudBuild(param)
	})
	if pt := cienv.Get(nil); pt != nil {
		ci.Name = pt.ID()

		if ci.Owner == "" {
			ci.Owner = pt.RepoOwner()
		}

		if ci.Repo == "" {
			ci.Repo = pt.RepoName()
		}

		if ci.SHA == "" {
			ci.SHA = pt.SHA()
		}

		if ci.PRNumber <= 0 {
			n, err := pt.PRNumber()
			if err != nil {
				return err
			}
			ci.PRNumber = n
		}

		if ci.Link == "" {
			ci.Link = getLink(ci.Name)
		}
	}
	return nil
}
