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
			resp: `
## Terraform result





<details><summary>Details (Click me)</summary>

<pre><code>
</code></pre></details>
`,
		},
		{
			template: DefaultDefaultTemplate,
			value: CommonTemplate{
				Message: "message",
			},
			resp: `
## Terraform result

message



<details><summary>Details (Click me)</summary>

<pre><code>
</code></pre></details>
`,
		},
		{
			template: DefaultDefaultTemplate,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Body:    "d",
			},
			resp: `
a

b



<details><summary>Details (Click me)</summary>

<pre><code>c
</code></pre></details>
`,
		},

		{
			template: "",
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Body:    "d",
			},
			resp: `
a

b



<details><summary>Details (Click me)</summary>

<pre><code>c
</code></pre></details>
`,
		},
		{
			template: "",
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  `This is a "result".`,
				Body:    "d",
			},
			resp: `
a

b



<details><summary>Details (Click me)</summary>

<pre><code>This is a &#34;result&#34;.
</code></pre></details>
`,
		},
		{
			template: "",
			value: CommonTemplate{
				Title:        "a",
				Message:      "b",
				Result:       `This is a "result".`,
				Body:         "d",
				UseRawOutput: true,
			},
			resp: `
a

b



<details><summary>Details (Click me)</summary>

<pre><code>This is a "result".
</code></pre></details>
`,
		},
		{
			template: "",
			value: CommonTemplate{
				Title:        "a",
				Message:      "b",
				Result:       `This is a "result".`,
				Body:         "d",
				UseRawOutput: true,
			},
			resp: `
a

b



<details><summary>Details (Click me)</summary>

<pre><code>This is a "result".
</code></pre></details>
`,
		},
		{
			template: `{{ .Title }}-{{ .Message }}-{{ .Result }}-{{ .Body }}`,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "should be used as body",
				Body:    "should be empty",
			},
			resp: `a-b--should be used as body`,
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
			resp: `
## Fmt result






`,
		},
		{
			template: DefaultFmtTemplate,
			value: CommonTemplate{
				Message: "message",
			},
			resp: `
## Fmt result

message




`,
		},
		{
			template: DefaultFmtTemplate,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Body:    "d",
			},
			resp: `
a

b



c
`,
		},

		{
			template: "",
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Body:    "d",
			},
			resp: `
a

b



c
`,
		},
		{
			template: "",
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  `This is a "result".`,
				Body:    "d",
			},
			resp: `
a

b



This is a &#34;result&#34;.
`,
		},
		{
			template: "",
			value: CommonTemplate{
				Title:        "a",
				Message:      "b",
				Result:       `This is a "result".`,
				Body:         "d",
				UseRawOutput: true,
			},
			resp: `
a

b



This is a "result".
`,
		},
		{
			template: `{{ .Title }}-{{ .Message }}-{{ .Result }}-{{ .Body }}`,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "should be used as body",
				Body:    "should be empty",
			},
			resp: `a-b--should be used as body`,
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
			resp: `
## Plan result





<details><summary>Details (Click me)</summary>

<pre><code>
</code></pre></details>
`,
		},
		{
			template: DefaultPlanTemplate,
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "result",
				Body:    "body",
			},
			resp: `
title

message


<pre><code>result
</code></pre>


<details><summary>Details (Click me)</summary>

<pre><code>body
</code></pre></details>
`,
		},
		{
			template: DefaultPlanTemplate,
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "",
				Body:    "body",
			},
			resp: `
title

message



<details><summary>Details (Click me)</summary>

<pre><code>body
</code></pre></details>
`,
		},
		{
			template: DefaultPlanTemplate,
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "",
				Body:    `This is a "body".`,
			},
			resp: `
title

message



<details><summary>Details (Click me)</summary>

<pre><code>This is a &#34;body&#34;.
</code></pre></details>
`,
		},
		{
			template: DefaultPlanTemplate,
			value: CommonTemplate{
				Title:        "title",
				Message:      "message",
				Result:       "",
				Body:         `This is a "body".`,
				UseRawOutput: true,
			},
			resp: `
title

message



<details><summary>Details (Click me)</summary>

<pre><code>This is a "body".
</code></pre></details>
`,
		},
		{
			template: "",
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "",
				Body:    "body",
			},
			resp: `
title

message



<details><summary>Details (Click me)</summary>

<pre><code>body
</code></pre></details>
`,
		},
		{
			template: `{{ .Title }}-{{ .Message }}-{{ .Result }}-{{ .Body }}`,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Body:    "d",
			},
			resp: `a-b-c-d`,
		},
		{
			template: `{{ .Title }}-{{ .Message }}-{{ .Action }}-{{ .Result }}-{{ .Body }}`,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Action:  "action",
				Result:  "c",
				Body:    "d",
			},
			resp: `a-b-action-c-d`,
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

func TestDestroyWarningTemplateExecute(t *testing.T) {
	testCases := []struct {
		template string
		value    CommonTemplate
		resp     string
	}{
		{
			template: DefaultDestroyWarningTemplate,
			value:    CommonTemplate{},
			resp: `
## WARNING: Resource Deletion will happen

This plan contains resource delete operation. Please check the plan result very carefully!


`,
		},
		{
			template: DefaultDestroyWarningTemplate,
			value: CommonTemplate{
				Title:  "title",
				Result: `This is a "result".`,
			},
			resp: `
title

This plan contains resource delete operation. Please check the plan result very carefully!


<pre><code>This is a &#34;result&#34;.
</code></pre>

`,
		},
		{
			template: DefaultDestroyWarningTemplate,
			value: CommonTemplate{
				Title:        "title",
				Result:       `This is a "result".`,
				UseRawOutput: true,
			},
			resp: `
title

This plan contains resource delete operation. Please check the plan result very carefully!


<pre><code>This is a "result".
</code></pre>

`,
		},
		{
			template: DefaultDestroyWarningTemplate,
			value: CommonTemplate{
				Title:  "title",
				Result: "",
			},
			resp: `
title

This plan contains resource delete operation. Please check the plan result very carefully!


`,
		},
		{
			template: "",
			value: CommonTemplate{
				Title:  "title",
				Result: "",
			},
			resp: `
title

This plan contains resource delete operation. Please check the plan result very carefully!


`,
		},
		{
			template: `{{ .Title }}-{{ .Message }}-{{ .Result }}-{{ .Body }}`,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Body:    "d",
			},
			resp: `a-b-c-d`,
		},
	}
	for _, testCase := range testCases {
		template := NewDestroyWarningTemplate(testCase.template)
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
			resp: `
## Apply result





<details><summary>Details (Click me)</summary>

<pre><code>
</code></pre></details>
`,
		},
		{
			template: DefaultApplyTemplate,
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "result",
				Body:    "body",
			},
			resp: `
title

message


<pre><code>result
</code></pre>


<details><summary>Details (Click me)</summary>

<pre><code>body
</code></pre></details>
`,
		},
		{
			template: DefaultApplyTemplate,
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "",
				Body:    "body",
			},
			resp: `
title

message



<details><summary>Details (Click me)</summary>

<pre><code>body
</code></pre></details>
`,
		},
		{
			template: "",
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "",
				Body:    "body",
			},
			resp: `
title

message



<details><summary>Details (Click me)</summary>

<pre><code>body
</code></pre></details>
`,
		},
		{
			template: "",
			value: CommonTemplate{
				Title:   "title",
				Message: "message",
				Result:  "",
				Body:    `This is a "body".`,
			},
			resp: `
title

message



<details><summary>Details (Click me)</summary>

<pre><code>This is a &#34;body&#34;.
</code></pre></details>
`,
		},
		{
			template: "",
			value: CommonTemplate{
				Title:        "title",
				Message:      "message",
				Result:       "",
				Body:         `This is a "body".`,
				UseRawOutput: true,
			},
			resp: `
title

message



<details><summary>Details (Click me)</summary>

<pre><code>This is a "body".
</code></pre></details>
`,
		},
		{
			template: `{{ .Title }}-{{ .Message }}-{{ .Result }}-{{ .Body }}`,
			value: CommonTemplate{
				Title:   "a",
				Message: "b",
				Result:  "c",
				Body:    "d",
			},
			resp: `a-b-c-d`,
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
			t.Errorf("got %#v but want %#v", value, testCase.expected)
		}
	}
}
