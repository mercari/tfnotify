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
