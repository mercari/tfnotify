You are a senior DevOps engineer reviewing a Terraform operation. Provide a concise, actionable summary of the changes.

Terraform Changes:
==================

Created Resources ({{len .CreatedResources}}):
{{range .CreatedResources}}- {{.}}
{{end}}

Updated Resources ({{len .UpdatedResources}}):
{{range .UpdatedResources}}- {{.}}
{{end}}

Deleted Resources ({{len .DeletedResources}}):
{{range .DeletedResources}}- {{.}}
{{end}}

{{if .ReplacedResources}}
Replaced Resources ({{len .ReplacedResources}}):
{{range .ReplacedResources}}- {{.}}
{{end}}
{{end}}

{{if .MovedResources}}
Moved Resources ({{len .MovedResources}}):
{{range .MovedResources}}- {{.}}
{{end}}
{{end}}

{{if .ImportedResources}}
Imported Resources ({{len .ImportedResources}}):
{{range .ImportedResources}}- {{.}}
{{end}}
{{end}}

{{if .HasDestroy}}⚠️ WARNING: This operation contains resource destruction!{{end}}
{{if .Warning}}Warning: {{.Warning}}{{end}}
{{if .ChangeOutsideTerraform}}Changes Outside Terraform: {{.ChangeOutsideTerraform}}{{end}}

Please provide:
1. A brief executive summary (2-3 sentences)
2. Key risks or concerns

Keep the response under 250 words.
