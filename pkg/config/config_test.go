package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoadFile(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		file string
		cfg  Config
		ok   bool
	}{
		{
			file: "../../example.tfnotify.yaml",
			cfg: Config{
				Terraform: Terraform{
					Plan: Plan{
						Template:    "## Plan Result\n{{if .Result}}\n<pre><code>{{ .Result }}\n</pre></code>\n{{end}}\n<details><summary>Details (Click me)</summary>\n\n<pre><code>{{ .CombinedOutput }}\n</pre></code></details>\n",
						WhenDestroy: WhenDestroy{},
					},
					Apply: Apply{
						Template: "",
					},
					UseRawOutput: false,
				},
			},
			ok: true,
		},
		{
			file: "../../example-with-destroy-and-result-labels.tfnotify.yaml",
			cfg: Config{
				Terraform: Terraform{
					Plan: Plan{
						Template: `{{if .HasDestroy}}
## :warning: WARNING: Resource Deletion will happen

This plan contains **resource deletion**. Please check the plan result very carefully!
{{else}}
## Plan Result
{{if .Result}}
<pre><code>{{ .Result }}
</pre></code>
{{end}}
<details><summary>Details (Click me)</summary>

<pre><code>{{ .CombinedOutput }}
</pre></code></details>
{{end}}
`,
						WhenAddOrUpdateOnly: WhenAddOrUpdateOnly{
							Label: "add-or-update",
						},
						WhenDestroy: WhenDestroy{
							Label: "destroy",
						},
						WhenPlanError: WhenPlanError{
							Label: "error",
						},
						WhenNoChanges: WhenNoChanges{
							Label: "no-changes",
						},
					},
					Apply: Apply{
						Template: "",
					},
					UseRawOutput: false,
				},
			},
			ok: true,
		},
		{
			file: "no-such-config.yaml",
			cfg: Config{
				Terraform: Terraform{
					Plan: Plan{
						Template:    "## Plan Result\n{{if .Result}}\n<pre><code>{{ .Result }}\n</pre></code>\n{{end}}\n<details><summary>Details (Click me)</summary>\n\n<pre><code>{{ .CombinedOutput }}\n</pre></code></details>\n",
						WhenDestroy: WhenDestroy{},
					},
					Apply: Apply{
						Template: "",
					},
				},
			},
			ok: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.file, func(t *testing.T) {
			t.Parallel()
			var cfg Config

			if err := cfg.LoadFile(testCase.file); err == nil {
				if !testCase.ok {
					t.Error("got no error but want error")
				} else if diff := cmp.Diff(cfg, testCase.cfg); diff != "" {
					t.Error(diff)
				}
			} else {
				if testCase.ok {
					t.Errorf("got error %v but want no error", err)
				}
			}
		})
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
	f, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, 0o666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
}

func removeDummy(file string) {
	os.Remove(file)
}

func TestFind(t *testing.T) { //nolint:paralleltest
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
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
			expect: filepath.Join(wd, "tfnotify.yaml"),
			ok:     true,
		},
	}
	var cfg Config
	for _, testCase := range testCases { //nolint:paralleltest
		t.Run(testCase.file, func(t *testing.T) {
			createDummy(testCase.file)
			actual, err := cfg.Find(testCase.file)
			if (err == nil) != testCase.ok {
				t.Errorf("got error %q", err)
			}
			if actual != testCase.expect {
				t.Errorf("got %q but want %q", actual, testCase.expect)
			}
		})
		defer removeDummy(testCase.file)
	}
}
