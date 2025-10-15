package terraform

import (
	"errors"
	"regexp"
	"strings"
)

// Parser is an interface for parsing terraform execution result
type Parser interface {
	Parse(body string) ParseResult
}

// ParseResult represents the result of parsed terraform execution
type ParseResult struct {
	Result             string
	OutsideTerraform   string
	ChangedResult      string
	Warning            string
	HasAddOrUpdateOnly bool
	HasDestroy         bool
	HasNoChanges       bool
	HasError           bool
	HasParseError      bool
	Error              error
	CreatedResources   []string
	UpdatedResources   []string
	DeletedResources   []string
	ReplacedResources  []string
	MovedResources     []*MovedResource
	ImportedResources  []string
}

// PlanParser is a parser for terraform plan
type PlanParser struct {
	Pass           *regexp.Regexp
	Fail           *regexp.Regexp
	Warning        *regexp.Regexp
	OutputsChanges *regexp.Regexp
	HasDestroy     *regexp.Regexp
	HasNoChanges   *regexp.Regexp
	Create         *regexp.Regexp
	Update         *regexp.Regexp
	Delete         *regexp.Regexp
	Replace        *regexp.Regexp
	ReplaceOption  *regexp.Regexp
	Move           *regexp.Regexp
	Import         *regexp.Regexp
	ImportedFrom   *regexp.Regexp
	MovedFrom      *regexp.Regexp
}

// ApplyParser is a parser for terraform apply
type ApplyParser struct {
	Pass *regexp.Regexp
	Fail *regexp.Regexp
}

// NewPlanParser is PlanParser initialized with its Regexp
func NewPlanParser() *PlanParser {
	return &PlanParser{
		Pass:           regexp.MustCompile(`(?m)^(Plan: \d|No changes.)`),
		Fail:           regexp.MustCompile(`(?m)^([│|] )?(Error: )`),
		Warning:        regexp.MustCompile(`(?m)^([│|] )?(Warning: )`),
		OutputsChanges: regexp.MustCompile(`(?m)^Changes to Outputs:`),
		// "0 to destroy" should be treated as "no destroy"
		HasDestroy: regexp.MustCompile(`(?m)([1-9][0-9]* to destroy.)`),
		// "0 to add, 0 to change, 0 to destroy" should be treated as "no change" (issue#358)
		HasNoChanges:  regexp.MustCompile(`(?m)^(No changes\.|Plan: 0 to add, 0 to change, 0 to destroy\.)`),
		Create:        regexp.MustCompile(`^ *# (.*) will be created$`),
		Update:        regexp.MustCompile(`^ *# (.*) will be updated in-place$`),
		Delete:        regexp.MustCompile(`^ *# (.*) will be destroyed$`),
		Replace:       regexp.MustCompile(`^ *# (.*?)(?: is tainted, so)? must be replaced$`),
		ReplaceOption: regexp.MustCompile(`^ *# (.*?) will be replaced, as requested$`),
		Move:          regexp.MustCompile(`^ *# (.*?) has moved to (.*?)$`),
		Import:        regexp.MustCompile(`^ *# (.*?) will be imported$`),
		ImportedFrom:  regexp.MustCompile(`^ *# \(imported from (.*?)\)$`),
		MovedFrom:     regexp.MustCompile(`^ *# \(moved from (.*?)\)$`),
	}
}

// NewApplyParser is ApplyParser initialized with its Regexp
func NewApplyParser() *ApplyParser {
	return &ApplyParser{
		Pass: regexp.MustCompile(`(?m)^(Apply complete!)`),
		Fail: regexp.MustCompile(`(?m)^(Error: )`),
	}
}

func extractResource(pattern *regexp.Regexp, line string) string {
	if arr := pattern.FindStringSubmatch(line); len(arr) == 2 { //nolint:mnd
		return arr[1]
	}
	return ""
}

func extractMovedResource(pattern *regexp.Regexp, line string) *MovedResource {
	if arr := pattern.FindStringSubmatch(line); len(arr) == 3 { //nolint:mnd
		return &MovedResource{
			Before: arr[1],
			After:  arr[2],
		}
	}
	return nil
}

// Parse returns ParseResult related with terraform plan
func (p *PlanParser) Parse(body string) ParseResult { //nolint:cyclop,maintidx
	switch {
	case p.Fail.MatchString(body):
	case p.Pass.MatchString(body) || p.OutputsChanges.MatchString(body):
	default:
		return ParseResult{
			Result:        "",
			HasParseError: true,
			Error:         errors.New("cannot parse plan result"),
		}
	}
	lines := strings.Split(body, "\n")
	firstMatchLineIndex := -1
	var result, firstMatchLine string
	var createdResources, updatedResources, deletedResources, replacedResources, importedResources []string
	var movedResources []*MovedResource
	startOutsideTerraform := -1
	endOutsideTerraform := -1
	startChangeOutput := -1
	endChangeOutput := -1
	startWarning := -1
	endWarning := -1
	startErrorIndex := -1
	for i, line := range lines {
		if line == "Note: Objects have changed outside of Terraform" || line == "Note: Objects have changed outside of OpenTofu" { // https://github.com/hashicorp/terraform/blob/332045a4e4b1d256c45f98aac74e31102ace7af7/internal/command/views/plan.go#L403
			startOutsideTerraform = i + 1
		}
		if startOutsideTerraform != -1 && endOutsideTerraform == -1 && strings.HasPrefix(line, "Unless you have made equivalent changes to your configuration") { // https://github.com/hashicorp/terraform/blob/332045a4e4b1d256c45f98aac74e31102ace7af7/internal/command/views/plan.go#L110
			endOutsideTerraform = i + 1
		}
		if line == "Terraform will perform the following actions:" || line == "OpenTofu will perform the following actions:" { // https://github.com/hashicorp/terraform/blob/332045a4e4b1d256c45f98aac74e31102ace7af7/internal/command/views/plan.go#L252
			startChangeOutput = i + 1
		}
		// If we have output changes but not resource changes, Terraform
		// does not output `Terraform will perform the following actions:`.
		if line == "Changes to Outputs:" && startChangeOutput == -1 {
			startChangeOutput = i
		}
		if p.Warning.MatchString(line) && startWarning == -1 {
			startWarning = i
		}
		// Terraform uses two types of rules.
		if strings.HasPrefix(line, "─────") || strings.HasPrefix(line, "-----") {
			if startWarning != -1 && endWarning == -1 {
				endWarning = i
			}
			if startChangeOutput != -1 && endChangeOutput == -1 {
				endChangeOutput = i - 1
			}
		}
		if startErrorIndex == -1 {
			if p.Fail.MatchString(line) {
				startErrorIndex = i
				firstMatchLineIndex = i
				firstMatchLine = line
			}
		}
		if firstMatchLineIndex == -1 {
			if p.Pass.MatchString(line) || p.OutputsChanges.MatchString(line) {
				firstMatchLineIndex = i
				firstMatchLine = line
			}
		}
		if rsc := extractResource(p.Create, line); rsc != "" {
			createdResources = append(createdResources, rsc)
		} else if rsc := extractResource(p.Update, line); rsc != "" {
			updatedResources = append(updatedResources, rsc)
		} else if rsc := extractResource(p.Delete, line); rsc != "" {
			deletedResources = append(deletedResources, rsc)
		} else if rsc := extractResource(p.Replace, line); rsc != "" {
			replacedResources = append(replacedResources, rsc)
		} else if rsc := extractResource(p.ReplaceOption, line); rsc != "" {
			replacedResources = append(replacedResources, rsc)
		} else if rsc := extractResource(p.Import, line); rsc != "" {
			importedResources = append(importedResources, rsc)
		} else if rsc := extractResource(p.ImportedFrom, line); rsc != "" {
			if i == 0 {
				continue
			}
			if rsc := p.changedResources(lines[i-1]); rsc != "" {
				importedResources = append(importedResources, rsc)
			}
		} else if rsc := extractMovedResource(p.Move, line); rsc != nil {
			movedResources = append(movedResources, rsc)
		} else if fromRsc := extractResource(p.MovedFrom, line); fromRsc != "" {
			if i == 0 {
				continue
			}
			if toRsc := p.changedResources(lines[i-1]); toRsc != "" {
				movedResources = append(movedResources, &MovedResource{
					Before: fromRsc,
					After:  toRsc,
				})
			}
		}
	}
	var hasPlanError bool
	switch {
	case p.Fail.MatchString(firstMatchLine):
		// Fail should be checked before Pass
		hasPlanError = true
		result = strings.Join(trimBars(trimLastNewline(lines[firstMatchLineIndex:])), "\n")
	case p.Pass.MatchString(firstMatchLine):
		result = lines[firstMatchLineIndex]
	case p.OutputsChanges.MatchString(firstMatchLine):
		result = "Only Outputs will be changed."
	}

	hasDestroy := p.HasDestroy.MatchString(firstMatchLine)
	hasNoChanges := p.HasNoChanges.MatchString(firstMatchLine)
	HasAddOrUpdateOnly := !hasNoChanges && !hasDestroy && !hasPlanError

	outsideTerraform := ""
	if startOutsideTerraform != -1 {
		outsideTerraform = strings.Join(lines[startOutsideTerraform:endOutsideTerraform], "\n")
	}

	changeResult := ""
	if startChangeOutput != -1 {
		// if we get here before finding a horizontal rule, output all remaining.
		if endChangeOutput == -1 {
			endChangeOutput = len(lines) - 1
		}
		changeResult = strings.Join(lines[startChangeOutput:endChangeOutput], "\n")
	}

	warnings := ""
	if startWarning != -1 {
		if endWarning == -1 {
			warnings = strings.Join(trimBars(lines[startWarning:]), "\n")
		} else {
			warnings = strings.Join(trimBars(lines[startWarning:endWarning]), "\n")
		}
	}

	return ParseResult{
		Result:             strings.TrimSpace(result),
		ChangedResult:      changeResult,
		OutsideTerraform:   outsideTerraform,
		Warning:            strings.TrimSpace(warnings),
		HasAddOrUpdateOnly: HasAddOrUpdateOnly,
		HasDestroy:         hasDestroy,
		HasNoChanges:       hasNoChanges,
		HasError:           hasPlanError,
		Error:              nil,
		CreatedResources:   createdResources,
		UpdatedResources:   updatedResources,
		DeletedResources:   deletedResources,
		ReplacedResources:  replacedResources,
		MovedResources:     movedResources,
		ImportedResources:  importedResources,
	}
}

func (p *PlanParser) changedResources(line string) string {
	if rsc := extractResource(p.Update, line); rsc != "" {
		return rsc
	} else if rsc := extractResource(p.Replace, line); rsc != "" {
		return rsc
	} else if rsc := extractResource(p.ReplaceOption, line); rsc != "" {
		return rsc
	}
	return ""
}

type MovedResource struct {
	Before string
	After  string
}

// Parse returns ParseResult related with terraform apply
func (p *ApplyParser) Parse(body string) ParseResult {
	var hasError bool
	switch {
	case p.Fail.MatchString(body):
		hasError = true
	case p.Pass.MatchString(body):
	default:
		return ParseResult{
			Result:        "",
			HasParseError: true,
			Error:         errors.New("cannot parse apply result"),
		}
	}
	lines := strings.Split(body, "\n")
	var i int
	var result, line string
	for i, line = range lines {
		if p.Pass.MatchString(line) || p.Fail.MatchString(line) {
			break
		}
	}
	switch {
	case p.Fail.MatchString(line):
		// Fail should be checked before Pass
		result = strings.Join(trimBars(trimLastNewline(lines[i:])), "\n")
	case p.Pass.MatchString(line):
		result = lines[i]
	}
	return ParseResult{
		Result:   strings.TrimSpace(result),
		HasError: hasError,
		Error:    nil,
	}
}

func trimLastNewline(s []string) []string {
	if len(s) == 0 {
		return s
	}
	last := len(s) - 1
	if s[last] == "" {
		return s[:last]
	}
	return s
}

func trimBars(list []string) []string {
	ret := make([]string, len(list))
	for i, elem := range list {
		ret[i] = strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(elem, "|"), "│"), "╵")
	}
	return ret
}
