package main

import (
	"os"
	"reflect"
	"testing"
)

func TestCircleci(t *testing.T) {
	envs := []string{
		"CIRCLE_SHA1",
		"CIRCLE_BUILD_URL",
		"CIRCLE_PULL_REQUEST",
		"CI_PULL_REQUEST",
		"CIRCLE_PR_NUMBER",
	}
	saveEnvs := make(map[string]string)
	for _, key := range envs {
		saveEnvs[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, value := range saveEnvs {
			os.Setenv(key, value)
		}
	}()

	testCases := []struct {
		fn func()
		ci CI
		ok bool
	}{
		{
			fn: func() {
				os.Setenv("CIRCLE_SHA1", "abcdefg")
				os.Setenv("CIRCLE_BUILD_URL", "https://circleci.com/gh/owner/repo/1234")
				os.Setenv("CIRCLE_PULL_REQUEST", "")
				os.Setenv("CI_PULL_REQUEST", "")
				os.Setenv("CIRCLE_PR_NUMBER", "")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "abcdefg",
					Number:   0,
				},
				URL: "https://circleci.com/gh/owner/repo/1234",
			},
			ok: true,
		},
		{
			fn: func() {
				os.Setenv("CIRCLE_SHA1", "abcdefg")
				os.Setenv("CIRCLE_BUILD_URL", "https://circleci.com/gh/owner/repo/1234")
				os.Setenv("CIRCLE_PULL_REQUEST", "https://github.com/owner/repo/pull/1")
				os.Setenv("CI_PULL_REQUEST", "")
				os.Setenv("CIRCLE_PR_NUMBER", "")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "abcdefg",
					Number:   1,
				},
				URL: "https://circleci.com/gh/owner/repo/1234",
			},
			ok: true,
		},
		{
			fn: func() {
				os.Setenv("CIRCLE_SHA1", "abcdefg")
				os.Setenv("CIRCLE_BUILD_URL", "https://circleci.com/gh/owner/repo/1234")
				os.Setenv("CIRCLE_PULL_REQUEST", "")
				os.Setenv("CI_PULL_REQUEST", "2")
				os.Setenv("CIRCLE_PR_NUMBER", "")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "abcdefg",
					Number:   2,
				},
				URL: "https://circleci.com/gh/owner/repo/1234",
			},
			ok: true,
		},
		{
			fn: func() {
				os.Setenv("CIRCLE_SHA1", "abcdefg")
				os.Setenv("CIRCLE_BUILD_URL", "https://circleci.com/gh/owner/repo/1234")
				os.Setenv("CIRCLE_PULL_REQUEST", "")
				os.Setenv("CI_PULL_REQUEST", "")
				os.Setenv("CIRCLE_PR_NUMBER", "3")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "abcdefg",
					Number:   3,
				},
				URL: "https://circleci.com/gh/owner/repo/1234",
			},
			ok: true,
		},
		{
			fn: func() {
				os.Setenv("CIRCLE_SHA1", "")
				os.Setenv("CIRCLE_BUILD_URL", "https://circleci.com/gh/owner/repo/1234")
				os.Setenv("CIRCLE_PULL_REQUEST", "")
				os.Setenv("CI_PULL_REQUEST", "")
				os.Setenv("CIRCLE_PR_NUMBER", "")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "",
					Number:   0,
				},
				URL: "https://circleci.com/gh/owner/repo/1234",
			},
			ok: true,
		},
		{
			fn: func() {
				os.Setenv("CIRCLE_SHA1", "")
				os.Setenv("CIRCLE_BUILD_URL", "https://circleci.com/gh/owner/repo/1234")
				os.Setenv("CIRCLE_PULL_REQUEST", "abcdefg")
				os.Setenv("CI_PULL_REQUEST", "")
				os.Setenv("CIRCLE_PR_NUMBER", "")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "",
					Number:   0,
				},
				URL: "https://circleci.com/gh/owner/repo/1234",
			},
			ok: false,
		},
	}

	for _, testCase := range testCases {
		testCase.fn()
		ci, err := circleci()
		if !reflect.DeepEqual(ci, testCase.ci) {
			t.Errorf("got %q but want %q", ci, testCase.ci)
		}
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
	}
}

func TestTravisCI(t *testing.T) {
	envs := []string{
		"TRAVIS_PULL_REQUEST_SHA",
		"TRAVIS_PULL_REQUEST",
	}
	saveEnvs := make(map[string]string)
	for _, key := range envs {
		saveEnvs[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, value := range saveEnvs {
			os.Setenv(key, value)
		}
	}()

	// https://docs.travis-ci.com/user/environment-variables/
	testCases := []struct {
		fn func()
		ci CI
		ok bool
	}{
		{
			fn: func() {
				os.Setenv("TRAVIS_PULL_REQUEST_SHA", "abcdefg")
				os.Setenv("TRAVIS_PULL_REQUEST", "1")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "abcdefg",
					Number:   1,
				},
				URL: "",
			},
			ok: true,
		},
		{
			fn: func() {
				os.Setenv("TRAVIS_PULL_REQUEST_SHA", "abcdefg")
				os.Setenv("TRAVIS_PULL_REQUEST", "false")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "abcdefg",
					Number:   0,
				},
				URL: "",
			},
			ok: false,
		},
	}

	for _, testCase := range testCases {
		testCase.fn()
		ci, err := travisci()
		if !reflect.DeepEqual(ci, testCase.ci) {
			t.Errorf("got %q but want %q", ci, testCase.ci)
		}
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
	}
}

func TestTeamCityCI(t *testing.T) {
	envs := []string{
		"BUILD_VCS_NUMBER",
		"BUILD_NUMBER",
	}
	saveEnvs := make(map[string]string)
	for _, key := range envs {
		saveEnvs[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, value := range saveEnvs {
			os.Setenv(key, value)
		}
	}()

	// https://confluence.jetbrains.com/display/TCD18/Predefined+Build+Parameters
	testCases := []struct {
		fn func()
		ci CI
		ok bool
	}{
		{
			fn: func() {
				os.Setenv("BUILD_NUMBER", "123")
				os.Setenv("BUILD_VCS_NUMBER", "fafef5adb5b9c39244027c8f16f7c3aa7e352b2e")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "fafef5adb5b9c39244027c8f16f7c3aa7e352b2e",
					Number:   123,
				},
				URL: "",
			},
			ok: true,
		},
		{
			fn: func() {
				os.Setenv("BUILD_NUMBER", "abcdefg")
				os.Setenv("BUILD_VCS_NUMBER", "false")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "abcdefg",
					Number:   0,
				},
				URL: "",
			},
			ok: false,
		},
	}

	for _, testCase := range testCases {
		testCase.fn()
		ci, err := travisci()
		if !reflect.DeepEqual(ci, testCase.ci) {
			t.Errorf("got %q but want %q", ci, testCase.ci)
		}
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
	}
}

func TestCodeBuild(t *testing.T) {
	envs := []string{
		"CODEBUILD_RESOLVED_SOURCE_VERSION",
		"CODEBUILD_SOURCE_VERSION",
		"CODEBUILD_BUILD_URL",
	}
	saveEnvs := make(map[string]string)
	for _, key := range envs {
		saveEnvs[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, value := range saveEnvs {
			os.Setenv(key, value)
		}
	}()

	// https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref.html
	testCases := []struct {
		fn func()
		ci CI
		ok bool
	}{
		{
			fn: func() {
				os.Setenv("CODEBUILD_RESOLVED_SOURCE_VERSION", "abcdefg")
				os.Setenv("CODEBUILD_SOURCE_VERSION", "pr/123")
				os.Setenv("CODEBUILD_BUILD_URL", "https://ap-northeast-1.console.aws.amazon.com/codebuild/home?region=ap-northeast-1#/builds/test:f2ae4314-c2d6-4db6-83c2-eacbab1517b7/view/new")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "abcdefg",
					Number:   123,
				},
				URL: "https://ap-northeast-1.console.aws.amazon.com/codebuild/home?region=ap-northeast-1#/builds/test:f2ae4314-c2d6-4db6-83c2-eacbab1517b7/view/new",
			},
			ok: true,
		},
		{
			fn: func() {
				os.Setenv("CODEBUILD_RESOLVED_SOURCE_VERSION", "abcdefg")
				os.Setenv("CODEBUILD_SOURCE_VERSION", "pr/1")
				os.Setenv("CODEBUILD_BUILD_URL", "https://ap-northeast-1.console.aws.amazon.com/codebuild/home?region=ap-northeast-1#/builds/test:f2ae4314-c2d6-4db6-83c2-eacbab1517b7/view/new")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "abcdefg",
					Number:   1,
				},
				URL: "https://ap-northeast-1.console.aws.amazon.com/codebuild/home?region=ap-northeast-1#/builds/test:f2ae4314-c2d6-4db6-83c2-eacbab1517b7/view/new",
			},
			ok: true,
		},
		{
			fn: func() {
				os.Setenv("CODEBUILD_RESOLVED_SOURCE_VERSION", "")
				os.Setenv("CODEBUILD_SOURCE_VERSION", "")
				os.Setenv("CODEBUILD_BUILD_URL", "https://ap-northeast-1.console.aws.amazon.com/codebuild/home?region=ap-northeast-1#/builds/test:f2ae4314-c2d6-4db6-83c2-eacbab1517b7/view/new")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "",
					Number:   0,
				},
				URL: "https://ap-northeast-1.console.aws.amazon.com/codebuild/home?region=ap-northeast-1#/builds/test:f2ae4314-c2d6-4db6-83c2-eacbab1517b7/view/new",
			},
			ok: true,
		},
		{
			fn: func() {
				os.Setenv("CODEBUILD_RESOLVED_SOURCE_VERSION", "")
				os.Setenv("CODEBUILD_SOURCE_VERSION", "pr/abc")
				os.Setenv("CODEBUILD_BUILD_URL", "https://ap-northeast-1.console.aws.amazon.com/codebuild/home?region=ap-northeast-1#/builds/test:f2ae4314-c2d6-4db6-83c2-eacbab1517b7/view/new")
			},
			ci: CI{
				PR: PullRequest{
					Revision: "",
					Number:   0,
				},
				URL: "https://ap-northeast-1.console.aws.amazon.com/codebuild/home?region=ap-northeast-1#/builds/test:f2ae4314-c2d6-4db6-83c2-eacbab1517b7/view/new",
			},
			ok: false,
		},
	}

	for _, testCase := range testCases {
		testCase.fn()
		ci, err := codebuild()
		if !reflect.DeepEqual(ci, testCase.ci) {
			t.Errorf("got %q but want %q", ci, testCase.ci)
		}
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
	}
}
