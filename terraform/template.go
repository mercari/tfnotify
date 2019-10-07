package terraform

import (
	"bytes"
	htmltemplate "html/template"
	texttemplate "text/template"
)

const (
	// DefaultDefaultTitle is a default title for terraform commands
	DefaultDefaultTitle = "## Terraform result"
	// DefaultFmtTitle is a default title for terraform fmt
	DefaultFmtTitle = "## Fmt result"
	// DefaultPlanTitle is a default title for terraform plan
	DefaultPlanTitle = "## Plan result"
	// DefaultDestroyWarningTitle is a default title of destroy warning
	DefaultDestroyWarningTitle = "## WARNING: Resource Deletion will happen"
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

	// DefaultDestroyWarningTemplate is a default template for terraform plan
	DefaultDestroyWarningTemplate = `
{{ .Title }}

This plan contains resource delete operation. Please check the plan result very carefully!

{{if .Result}}
<pre><code>{{ .Result }}
</code></pre>
{{end}}
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
	Title        string
	Message      string
	Action       string
	Result       string
	Body         string
	Link         string
	UseRawOutput bool
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

// DestroyWarningTemplate is a default template for warning of destroy operation in plan
type DestroyWarningTemplate struct {
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

// NewDestroyWarningTemplate is DestroyWarningTemplate initializer
func NewDestroyWarningTemplate(template string) *DestroyWarningTemplate {
	if template == "" {
		template = DefaultDestroyWarningTemplate
	}
	return &DestroyWarningTemplate{
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

func generateOutput(kind, template string, data map[string]interface{}, useRawOutput bool) (string, error) {
	var b bytes.Buffer

	if useRawOutput {
		tpl, err := texttemplate.New(kind).Parse(template)
		if err != nil {
			return "", err
		}
		if err := tpl.Execute(&b, data); err != nil {
			return "", err
		}
	} else {
		tpl, err := htmltemplate.New(kind).Parse(template)
		if err != nil {
			return "", err
		}
		if err := tpl.Execute(&b, data); err != nil {
			return "", err
		}
	}

	return b.String(), nil
}

// Execute binds the execution result of terraform command into tepmlate
func (t *DefaultTemplate) Execute() (string, error) {
	data := map[string]interface{}{
		"Title":   t.Title,
		"Message": t.Message,
		"Result":  "",
		"Body":    t.Result,
		"Link":    t.Link,
	}

	resp, err := generateOutput("default", t.Template, data, t.UseRawOutput)
	if err != nil {
		return "", err
	}

	return resp, nil
}

// Execute binds the execution result of terraform fmt into tepmlate
func (t *FmtTemplate) Execute() (string, error) {
	data := map[string]interface{}{
		"Title":   t.Title,
		"Message": t.Message,
		"Result":  "",
		"Body":    t.Result,
		"Link":    t.Link,
	}

	resp, err := generateOutput("fmt", t.Template, data, t.UseRawOutput)
	if err != nil {
		return "", err
	}

	return resp, nil
}

// Execute binds the execution result of terraform plan into tepmlate
func (t *PlanTemplate) Execute() (string, error) {
	data := map[string]interface{}{
		"Title":   t.Title,
		"Message": t.Message,
		"Result":  t.Result,
		"Action":  t.Action,
		"Body":    t.Body,
		"Link":    t.Link,
	}

	resp, err := generateOutput("plan", t.Template, data, t.UseRawOutput)
	if err != nil {
		return "", err
	}

	return resp, nil
}

// Execute binds the execution result of terraform plan into template
func (t *DestroyWarningTemplate) Execute() (string, error) {
	data := map[string]interface{}{
		"Title":   t.Title,
		"Message": t.Message,
		"Result":  t.Result,
		"Action":  t.Action,
		"Body":    t.Body,
		"Link":    t.Link,
	}

	resp, err := generateOutput("destroy_warning", t.Template, data, t.UseRawOutput)
	if err != nil {
		return "", err
	}

	return resp, nil
}

// Execute binds the execution result of terraform apply into tepmlate
func (t *ApplyTemplate) Execute() (string, error) {
	data := map[string]interface{}{
		"Title":   t.Title,
		"Message": t.Message,
		"Result":  t.Result,
		"Body":    t.Body,
		"Link":    t.Link,
	}

	resp, err := generateOutput("apply", t.Template, data, t.UseRawOutput)
	if err != nil {
		return "", err
	}

	return resp, nil
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

// SetValue sets template entities about destroy warning to CommonTemplate
func (t *DestroyWarningTemplate) SetValue(ct CommonTemplate) {
	if ct.Title == "" {
		ct.Title = DefaultDestroyWarningTitle
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
func (t *DestroyWarningTemplate) GetValue() CommonTemplate {
	return t.CommonTemplate
}

// GetValue gets template entities
func (t *ApplyTemplate) GetValue() CommonTemplate {
	return t.CommonTemplate
}
