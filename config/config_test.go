package config

import (
	"os"
	"reflect"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func helperLoadConfig(contents []byte) (*Config, error) {
	cfg := &Config{}
	err := yaml.Unmarshal(contents, cfg)
	return cfg, err
}

func TestLoadFile(t *testing.T) {
	testCases := []struct {
		file string
		cfg  Config
		ok   bool
	}{
		{
			file: "../example.tfnotify.yaml",
			cfg: Config{
				CI: "circleci",
				Notifier: Notifier{
					Github: GithubNotifier{
						Token: "$GITHUB_TOKEN",
						Repository: Repository{
							Owner: "mercari",
							Name:  "tfnotify",
						},
					},
					Slack: SlackNotifier{
						Token:   "",
						Channel: "",
						Bot:     "",
					},
					Typetalk: TypetalkNotifier{
						Token:   "",
						TopicID: "",
					},
				},
				Terraform: Terraform{
					Default: Default{
						Template: "",
					},
					Fmt: Fmt{
						Template: "",
					},
					Plan: Plan{
						Template: "{{ .Title }}\n{{ .Message }}\n{{if .Result}}\n<pre><code> {{ .Result }}\n</pre></code>\n{{end}}\n<details><summary>Details (Click me)</summary>\n<pre><code> {{ .Body }}\n</pre></code></details>\n",
					},
					Apply: Apply{
						Template: "",
					},
				},
				path: "../example.tfnotify.yaml",
			},
			ok: true,
		},
		{
			file: "no-such-config.yaml",
			cfg: Config{
				CI: "circleci",
				Notifier: Notifier{
					Github: GithubNotifier{
						Token: "$GITHUB_TOKEN",
						Repository: Repository{
							Owner: "mercari",
							Name:  "tfnotify",
						},
					},
					Slack: SlackNotifier{
						Token:   "",
						Channel: "",
						Bot:     "",
					},
					Typetalk: TypetalkNotifier{
						Token:   "",
						TopicID: "",
					},
				},
				Terraform: Terraform{
					Default: Default{
						Template: "",
					},
					Fmt: Fmt{
						Template: "",
					},
					Plan: Plan{
						Template: "{{ .Title }}\n{{ .Message }}\n{{if .Result}}\n<pre><code> {{ .Result }}\n</pre></code>\n{{end}}\n<details><summary>Details (Click me)</summary>\n<pre><code> {{ .Body }}\n</pre></code></details>\n",
					},
					Apply: Apply{
						Template: "",
					},
				},
				path: "no-such-config.yaml",
			},
			ok: false,
		},
	}

	var cfg Config
	for _, testCase := range testCases {
		err := cfg.LoadFile(testCase.file)
		if !reflect.DeepEqual(cfg, testCase.cfg) {
			t.Errorf("got %q but want %q", cfg, testCase.cfg)
		}
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
	}
}

func TestValidation(t *testing.T) {
	testCases := []struct {
		contents []byte
		expected string
	}{
		{
			contents: []byte(""),
			expected: "ci: need to be set",
		},
		{
			contents: []byte("ci: rare-ci\n"),
			expected: "rare-ci: not supported yet",
		},
		{
			contents: []byte("ci: circleci\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: travisci\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: codebuild\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: teamcity\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: drone\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: jenkins\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: gitlabci\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: circleci\nnotifier:\n  github:\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: circleci\nnotifier:\n  github:\n    token: token\n"),
			expected: "repository owner is missing",
		},
		{
			contents: []byte(`
ci: circleci
notifier:
  github:
    token: token
    repository:
      owner: owner
`),
			expected: "repository name is missing",
		},
		{
			contents: []byte(`
ci: circleci
notifier:
  github:
    token: token
    repository:
      owner: owner
      name: name
`),
			expected: "",
		},
		{
			contents: []byte(`
ci: circleci
notifier:
  slack:
`),
			expected: "notifier is missing",
		},
		{
			contents: []byte(`
ci: circleci
notifier:
  slack:
    token: token
`),
			expected: "slack channel id is missing",
		},
		{
			contents: []byte(`
ci: circleci
notifier:
  slack:
    token: token
    channel: channel
`),
			expected: "",
		},
		{
			contents: []byte(`
ci: circleci
notifier:
  typetalk:
`),
			expected: "notifier is missing",
		},
		{
			contents: []byte(`
ci: circleci
notifier:
  typetalk:
    token: token
`),
			expected: "Typetalk topic id is missing",
		},
		{
			contents: []byte(`
ci: circleci
notifier:
  typetalk:
    token: token
    topic_id: 12345
`),
			expected: "",
		},
	}
	for _, testCase := range testCases {
		cfg, err := helperLoadConfig(testCase.contents)
		if err != nil {
			t.Fatal(err)
		}
		err = cfg.Validation()
		if err == nil {
			if testCase.expected != "" {
				t.Errorf("got no error but want %q", testCase.expected)
			}
		} else {
			if err.Error() != testCase.expected {
				t.Errorf("got %q but want %q", err.Error(), testCase.expected)
			}
		}
	}
}

func TestGetNotifierType(t *testing.T) {
	testCases := []struct {
		contents []byte
		expected string
	}{
		{
			contents: []byte("repository:\n  owner: a\n  name: b\nci: circleci\nnotifier:\n  github:\n    token: token\n"),
			expected: "github",
		},
		{
			contents: []byte("repository:\n  owner: a\n  name: b\nci: circleci\nnotifier:\n  slack:\n    token: token\n"),
			expected: "slack",
		},
		{
			contents: []byte("repository:\n  owner: a\n  name: b\nci: circleci\nnotifier:\n  typetalk:\n    token: token\n"),
			expected: "typetalk",
		},
	}
	for _, testCase := range testCases {
		cfg, err := helperLoadConfig(testCase.contents)
		if err != nil {
			t.Fatal(err)
		}
		actual := cfg.GetNotifierType()
		if actual != testCase.expected {
			t.Errorf("got %q but want %q", actual, testCase.expected)
		}
	}
}

func createDummy(file string) {
	validConfig := func(file string) bool {
		for _, c := range []string{
			"tfnotify.yaml",
			"tfnotify.yml",
			".tfnotify.yaml",
			".tfnotify.yml",
		} {
			if file == c {
				return true
			}
		}
		return false
	}
	if !validConfig(file) {
		return
	}
	if _, err := os.Stat(file); err == nil {
		return
	}
	f, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
}

func removeDummy(file string) {
	os.Remove(file)
}

func TestFind(t *testing.T) {
	testCases := []struct {
		file   string
		expect string
		ok     bool
	}{
		{
			// valid config
			file:   ".tfnotify.yaml",
			expect: ".tfnotify.yaml",
			ok:     true,
		},
		{
			// valid config
			file:   "tfnotify.yaml",
			expect: "tfnotify.yaml",
			ok:     true,
		},
		{
			// valid config
			file:   ".tfnotify.yml",
			expect: ".tfnotify.yml",
			ok:     true,
		},
		{
			// valid config
			file:   "tfnotify.yml",
			expect: "tfnotify.yml",
			ok:     true,
		},
		{
			// invalid config
			file:   "codecov.yml",
			expect: "",
			ok:     false,
		},
		{
			// in case of no args passed
			file:   "",
			expect: "tfnotify.yaml",
			ok:     true,
		},
	}
	var cfg Config
	for _, testCase := range testCases {
		createDummy(testCase.file)
		defer removeDummy(testCase.file)
		actual, err := cfg.Find(testCase.file)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
		if actual != testCase.expect {
			t.Errorf("got %q but want %q", actual, testCase.expect)
		}
	}
}
