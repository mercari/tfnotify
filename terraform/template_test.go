package terraform

import (
	"reflect"
	"testing"
)

func TestDefaultTemplateExecute(t *testing.T) {
	testCases := []struct {
		template string
		value    CommonTemplate
		resp     string
	}{
		{
			template: DefaultDefaultTemplate,
			value:    CommonTemplate{},
			resp:     "\n## Terraform result\n\n\n\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>\n</code></pre></details>\n",
		},
		{
			template: DefaultDefaultTemplate,
			value: CommonTemplate{
				Message: "message",
			},
			resp: "\n## Terraform result\n\nmessage\n\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>\n</code></pre></details>\n",
		},
		{
			template: DefaultDefaultTemplate,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Detail:  "d",
				Body:    "e",
			},
			resp: "\na\n\nb\n\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>c\n</code></pre></details>\n",
		},

		{
			template: "",
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Detail:  "d",
				Body:    "e",
			},
			resp: "\na\n\nb\n\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>c\n</code></pre></details>\n",
		},
		{
			template: `{{ .Title }}-{{ .Message }}-{{ .Result }}-{{ .Body }}`,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "should be used as body",
				Detail:  "d",
				Body:    "should be empty",
			},
			resp: "a-b--should be used as body",
		},
	}
	for _, testCase := range testCases {
		template := NewDefaultTemplate(testCase.template)
		template.SetValue(testCase.value)
		resp, err := template.Execute()
		if err != nil {
			t.Fatal(err)
		}
		if resp != testCase.resp {
			t.Errorf("got %q but want %q", resp, testCase.resp)
		}
	}
}

func TestFmtTemplateExecute(t *testing.T) {
	testCases := []struct {
		template string
		value    CommonTemplate
		resp     string
	}{
		{
			template: DefaultFmtTemplate,
			value:    CommonTemplate{},
			resp:     "\n## Fmt result\n\n\n\n\n\n\n",
		},
		{
			template: DefaultFmtTemplate,
			value: CommonTemplate{
				Message: "message",
			},
			resp: "\n## Fmt result\n\nmessage\n\n\n\n\n",
		},
		{
			template: DefaultFmtTemplate,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Detail:  "d",
				Body:    "e",
			},
			resp: "\na\n\nb\n\n\n\nc\n",
		},

		{
			template: "",
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Detail:  "d",
				Body:    "e",
			},
			resp: "\na\n\nb\n\n\n\nc\n",
		},
		{
			template: `{{ .Title }}-{{ .Message }}-{{ .Result }}-{{ .Body }}`,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "should be used as body",
				Detail:  "d",
				Body:    "should be empty",
			},
			resp: "a-b--should be used as body",
		},
	}
	for _, testCase := range testCases {
		template := NewFmtTemplate(testCase.template)
		template.SetValue(testCase.value)
		resp, err := template.Execute()
		if err != nil {
			t.Fatal(err)
		}
		if resp != testCase.resp {
			t.Errorf("got %q but want %q", resp, testCase.resp)
		}
	}
}

func TestPlanTemplateExecute(t *testing.T) {
	testCases := []struct {
		template string
		value    CommonTemplate
		resp     string
	}{
		{
			template: DefaultPlanTemplate,
			value:    CommonTemplate{},
			resp:     "\n## Plan result\n\n\n\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>\n</code></pre></details>\n",
		},
		{
			template: DefaultPlanTemplate,
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "result",
				Detail:  "detail",
				Body:    "body",
			},
			resp: "\ntitle\n\nmessage\n\n\n<pre><code>result\n</code></pre>\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>body\n</code></pre></details>\n",
		},
		{
			template: DefaultPlanTemplate,
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "",
				Detail:  "",
				Body:    "body",
			},
			resp: "\ntitle\n\nmessage\n\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>body\n</code></pre></details>\n",
		},
		{
			template: "",
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "",
				Detail:  "",
				Body:    "body",
			},
			resp: "\ntitle\n\nmessage\n\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>body\n</code></pre></details>\n",
		},
		{
			template: `{{ .Title }}-{{ .Message }}-{{ .Result }}-{{ .Detail }}-{{ .Body }}`,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Detail:  "d",
				Body:    "e",
			},
			resp: "a-b-c-d-e",
		},
	}
	for _, testCase := range testCases {
		template := NewPlanTemplate(testCase.template)
		template.SetValue(testCase.value)
		resp, err := template.Execute()
		if err != nil {
			t.Fatal(err)
		}
		if resp != testCase.resp {
			t.Errorf("got %q but want %q", resp, testCase.resp)
		}
	}
}

func TestApplyTemplateExecute(t *testing.T) {
	testCases := []struct {
		template string
		value    CommonTemplate
		resp     string
	}{
		{
			template: DefaultApplyTemplate,
			value:    CommonTemplate{},
			resp:     "\n## Apply result\n\n\n\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>\n</code></pre></details>\n",
		},
		{
			template: DefaultApplyTemplate,
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "result",
				Body:    "body",
			},
			resp: "\ntitle\n\nmessage\n\n\n<pre><code>result\n</code></pre>\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>body\n</code></pre></details>\n",
		},
		{
			template: DefaultApplyTemplate,
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "",
				Body:    "body",
			},
			resp: "\ntitle\n\nmessage\n\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>body\n</code></pre></details>\n",
		},
		{
			template: "",
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "",
				Body:    "body",
			},
			resp: "\ntitle\n\nmessage\n\n\n\n<details><summary>Details (Click me)</summary>\n<pre><code>body\n</code></pre></details>\n",
		},
		{
			template: `{{ .Title }}-{{ .Message }}-{{ .Result }}-{{ .Body }}`,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Body:    "d",
			},
			resp: "a-b-c-d",
		},
	}
	for _, testCase := range testCases {
		template := NewApplyTemplate(testCase.template)
		template.SetValue(testCase.value)
		resp, err := template.Execute()
		if err != nil {
			t.Error(err)
		}
		if resp != testCase.resp {
			t.Errorf("got %q but want %q", resp, testCase.resp)
		}
	}
}

func TestGetValue(t *testing.T) {
	testCases := []struct {
		template Template
		expected CommonTemplate
	}{
		{
			template: NewDefaultTemplate(""),
			expected: CommonTemplate{},
		},
		{
			template: NewFmtTemplate(""),
			expected: CommonTemplate{},
		},
		{
			template: NewPlanTemplate(""),
			expected: CommonTemplate{},
		},
		{
			template: NewApplyTemplate(""),
			expected: CommonTemplate{},
		},
	}
	for _, testCase := range testCases {
		template := testCase.template
		value := template.GetValue()
		if !reflect.DeepEqual(value, testCase.expected) {
			t.Errorf("got %q but want %q", value, testCase.expected)
		}
	}
}
