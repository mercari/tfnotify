package gitlab

import (
	"testing"

	"github.com/mercari/tfnotify/terraform"
)

func TestNotifyNotify(t *testing.T) {
	testCases := []struct {
		config   Config
		body     string
		ok       bool
		exitCode int
	}{
		{
			// invalid body (cannot parse)
			config: Config{
				Token:     "token",
				NameSpace: "namespace",
				Project:   "project",
				MR: MergeRequest{
					Revision: "abcd",
					Number:   1,
					Message:  "message",
				},
				Parser:   terraform.NewPlanParser(),
				Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
			},
			body:     "body",
			ok:       false,
			exitCode: 1,
		},
		{
			// invalid pr
			config: Config{
				Token:     "token",
				NameSpace: "owner",
				Project:   "repo",
				MR: MergeRequest{
					Revision: "",
					Number:   0,
					Message:  "message",
				},
				Parser:   terraform.NewPlanParser(),
				Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
			},
			body:     "Plan: 1 to add",
			ok:       false,
			exitCode: 0,
		},
		{
			// valid, error
			config: Config{
				Token:     "token",
				NameSpace: "owner",
				Project:   "repo",
				MR: MergeRequest{
					Revision: "",
					Number:   1,
					Message:  "message",
				},
				Parser:   terraform.NewPlanParser(),
				Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
			},
			body:     "Error: hoge",
			ok:       true,
			exitCode: 1,
		},
		{
			// valid, and isPR
			config: Config{
				Token:     "token",
				NameSpace: "owner",
				Project:   "repo",
				MR: MergeRequest{
					Revision: "",
					Number:   1,
					Message:  "message",
				},
				Parser:   terraform.NewPlanParser(),
				Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
			},
			body:     "Plan: 1 to add",
			ok:       true,
			exitCode: 0,
		},
		{
			// valid, and isRevision
			config: Config{
				Token:     "token",
				NameSpace: "owner",
				Project:   "repo",
				MR: MergeRequest{
					Revision: "revision-revision",
					Number:   0,
					Message:  "message",
				},
				Parser:   terraform.NewPlanParser(),
				Template: terraform.NewPlanTemplate(terraform.DefaultPlanTemplate),
			},
			body:     "Plan: 1 to add",
			ok:       true,
			exitCode: 0,
		},
		{
			// apply case
			config: Config{
				Token:     "token",
				NameSpace: "owner",
				Project:   "repo",
				MR: MergeRequest{
					Revision: "revision",
					Number:   0, // For apply, it is always 0
					Message:  "message",
				},
				Parser:   terraform.NewApplyParser(),
				Template: terraform.NewApplyTemplate(terraform.DefaultApplyTemplate),
			},
			body:     "Apply complete!",
			ok:       true,
			exitCode: 0,
		},
	}

	for _, testCase := range testCases {
		client, err := NewClient(testCase.config)
		if err != nil {
			t.Fatal(err)
		}
		api := newFakeAPI()
		client.API = &api
		exitCode, err := client.Notify.Notify(testCase.body)
		if (err == nil) != testCase.ok {
			t.Errorf("got error %q", err)
		}
		if exitCode != testCase.exitCode {
			t.Errorf("got %q but want %q", exitCode, testCase.exitCode)
		}
	}
}
