package terraform

import (
	"bytes"
	"html/template"
)

const (
	// DefaultDefaultTitle is a default title for terraform commands
	DefaultDefaultTitle = "## Terraform result"
	// DefaultFmtTitle is a default title for terraform fmt
	DefaultFmtTitle = "## Fmt result"
	// DefaultPlanTitle is a default title for terraform plan
	DefaultPlanTitle = "## Plan result"
	// DefaultApplyTitle is a default title for terraform apply
	DefaultApplyTitle = "## Apply result"

	// DefaultDefaultTemplate is a default template for terraform commands
	DefaultDefaultTemplate = `
{{ .Title }}

{{ .Message }}

{{if .Result}}
<pre><code>{{ .Result }}
</code></pre>
{{end}}

<details><summary>Details (Click me)</summary>

<pre><code>{{ .Body }}
</code></pre></details>
`

	// DefaultFmtTemplate is a default template for terraform fmt
	DefaultFmtTemplate = `
{{ .Title }}

{{ .Message }}

{{ .Result }}

{{ .Body }}
`

	// DefaultPlanTemplate is a default template for terraform plan
	DefaultPlanTemplate = `
{{ .Title }}

{{ .Message }}

{{if .Result}}
<pre><code>{{ .Result }}
</code></pre>
{{end}}

<details><summary>Details (Click me)</summary>

<pre><code>{{ .Body }}
</code></pre></details>
`

	// DefaultApplyTemplate is a default template for terraform apply
	DefaultApplyTemplate = `
{{ .Title }}

{{ .Message }}

{{if .Result}}
<pre><code>{{ .Result }}
</code></pre>
{{end}}

<details><summary>Details (Click me)</summary>

<pre><code>{{ .Body }}
</code></pre></details>
`
)

// Template is an template interface for parsed terraform execution result
type Template interface {
	Execute() (resp string, err error)
	SetValue(template CommonTemplate)
	GetValue() CommonTemplate
}

// CommonTemplate represents template entities
type CommonTemplate struct {
	Title   string
	Message string
	Result  string
	Body    string
	Link    string
}

// DefaultTemplate is a default template for terraform commands
type DefaultTemplate struct {
	Template string

	CommonTemplate
}

// FmtTemplate is a default template for terraform fmt
type FmtTemplate struct {
	Template string

	CommonTemplate
}

// PlanTemplate is a default template for terraform plan
type PlanTemplate struct {
	Template string

	CommonTemplate
}

// ApplyTemplate is a default template for terraform apply
type ApplyTemplate struct {
	Template string

	CommonTemplate
}

// NewDefaultTemplate is DefaultTemplate initializer
func NewDefaultTemplate(template string) *DefaultTemplate {
	if template == "" {
		template = DefaultDefaultTemplate
	}
	return &DefaultTemplate{
		Template: template,
	}
}

// NewFmtTemplate is FmtTemplate initializer
func NewFmtTemplate(template string) *FmtTemplate {
	if template == "" {
		template = DefaultFmtTemplate
	}
	return &FmtTemplate{
		Template: template,
	}
}

// NewPlanTemplate is PlanTemplate initializer
func NewPlanTemplate(template string) *PlanTemplate {
	if template == "" {
		template = DefaultPlanTemplate
	}
	return &PlanTemplate{
		Template: template,
	}
}

// NewApplyTemplate is ApplyTemplate initializer
func NewApplyTemplate(template string) *ApplyTemplate {
	if template == "" {
		template = DefaultApplyTemplate
	}
	return &ApplyTemplate{
		Template: template,
	}
}

// Execute binds the execution result of terraform command into tepmlate
func (t *DefaultTemplate) Execute() (resp string, err error) {
	tpl, err := template.New("default").Parse(t.Template)
	if err != nil {
		return resp, err
	}
	var b bytes.Buffer
	if err := tpl.Execute(&b, map[string]interface{}{
		"Title":   t.Title,
		"Message": t.Message,
		"Result":  "",
		"Body":    t.Result,
		"Link":    t.Link,
	}); err != nil {
		return resp, err
	}
	resp = b.String()
	return resp, err
}

// Execute binds the execution result of terraform fmt into tepmlate
func (t *FmtTemplate) Execute() (resp string, err error) {
	tpl, err := template.New("fmt").Parse(t.Template)
	if err != nil {
		return resp, err
	}
	var b bytes.Buffer
	if err := tpl.Execute(&b, map[string]interface{}{
		"Title":   t.Title,
		"Message": t.Message,
		"Result":  "",
		"Body":    t.Result,
		"Link":    t.Link,
	}); err != nil {
		return resp, err
	}
	resp = b.String()
	return resp, err
}

// Execute binds the execution result of terraform plan into tepmlate
func (t *PlanTemplate) Execute() (resp string, err error) {
	tpl, err := template.New("plan").Parse(t.Template)
	if err != nil {
		return resp, err
	}
	var b bytes.Buffer
	if err := tpl.Execute(&b, map[string]interface{}{
		"Title":   t.Title,
		"Message": t.Message,
		"Result":  t.Result,
		"Body":    t.Body,
		"Link":    t.Link,
	}); err != nil {
		return resp, err
	}
	resp = b.String()
	return resp, err
}

// Execute binds the execution result of terraform apply into tepmlate
func (t *ApplyTemplate) Execute() (resp string, err error) {
	tpl, err := template.New("apply").Parse(t.Template)
	if err != nil {
		return resp, err
	}
	var b bytes.Buffer
	if err := tpl.Execute(&b, map[string]interface{}{
		"Title":   t.Title,
		"Message": t.Message,
		"Result":  t.Result,
		"Body":    t.Body,
		"Link":    t.Link,
	}); err != nil {
		return resp, err
	}
	resp = b.String()
	return resp, err
}

// SetValue sets template entities to CommonTemplate
func (t *DefaultTemplate) SetValue(ct CommonTemplate) {
	if ct.Title == "" {
		ct.Title = DefaultDefaultTitle
	}
	t.CommonTemplate = ct
}

// SetValue sets template entities about terraform fmt to CommonTemplate
func (t *FmtTemplate) SetValue(ct CommonTemplate) {
	if ct.Title == "" {
		ct.Title = DefaultFmtTitle
	}
	t.CommonTemplate = ct
}

// SetValue sets template entities about terraform plan to CommonTemplate
func (t *PlanTemplate) SetValue(ct CommonTemplate) {
	if ct.Title == "" {
		ct.Title = DefaultPlanTitle
	}
	t.CommonTemplate = ct
}

// SetValue sets template entities about terraform apply to CommonTemplate
func (t *ApplyTemplate) SetValue(ct CommonTemplate) {
	if ct.Title == "" {
		ct.Title = DefaultApplyTitle
	}
	t.CommonTemplate = ct
}

// GetValue gets template entities
func (t *DefaultTemplate) GetValue() CommonTemplate {
	return t.CommonTemplate
}

// GetValue gets template entities
func (t *FmtTemplate) GetValue() CommonTemplate {
	return t.CommonTemplate
}

// GetValue gets template entities
func (t *PlanTemplate) GetValue() CommonTemplate {
	return t.CommonTemplate
}

// GetValue gets template entities
func (t *ApplyTemplate) GetValue() CommonTemplate {
	return t.CommonTemplate
}
