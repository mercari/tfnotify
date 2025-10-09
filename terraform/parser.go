package terraform

import (
	"errors"
	"fmt"
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
	ExitCode           int
	Error              error
	CreatedResources   []string
	UpdatedResources   []string
	DeletedResources   []string
	ReplacedResources  []string
	MovedResources     []*MovedResource
	ImportedResources  []string
}

// MovedResource represents a moved resource
type MovedResource struct {
	Before string
	After  string
}

// DefaultParser is a parser for terraform commands
type DefaultParser struct {
}

// FmtParser is a parser for terraform fmt
type FmtParser struct {
	Pass *regexp.Regexp
	Fail *regexp.Regexp
}

// ValidateParser is a parser for terraform Validate
type ValidateParser struct {
	Pass *regexp.Regexp
	Fail *regexp.Regexp
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

// NewDefaultParser is DefaultParser initializer
func NewDefaultParser() *DefaultParser {
	return &DefaultParser{}
}

// NewFmtParser is FmtParser initialized with its Regexp
func NewFmtParser() *FmtParser {
	return &FmtParser{
		Fail: regexp.MustCompile(`(?m)^@@[^@]+@@`),
	}
}

// NewValidateParser is ValidateParser initialized with its Regexp
func NewValidateParser() *ValidateParser {
	return &ValidateParser{
		Fail: regexp.MustCompile(`(?m)^(│\s{1})?(Error: )`),
	}
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
		// "0 to add, 0 to change, 0 to destroy" should be treated as "no change"
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

// Parse returns ParseResult related with terraform commands
func (p *DefaultParser) Parse(body string) ParseResult {
	return ParseResult{
		Result:   body,
		ExitCode: ExitPass,
		Error:    nil,
	}
}

// Parse returns ParseResult related with terraform fmt
func (p *FmtParser) Parse(body string) ParseResult {
	result := ParseResult{}
	if p.Fail.MatchString(body) {
		result.Result = "There is diff in your .tf file (need to be formatted)"
		result.ExitCode = ExitFail
	}
	return result
}

// Parse returns ParseResult related with terraform validate
func (p *ValidateParser) Parse(body string) ParseResult {
	result := ParseResult{}
	if p.Fail.MatchString(body) {
		result.Result = "There is a validation error in your Terraform code"
		result.ExitCode = ExitFail
	}
	return result
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
	var firstMatchLine string
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
		if line == "Note: Objects have changed outside of Terraform" || line == "Note: Objects have changed outside of OpenTofu" {
			startOutsideTerraform = i + 1
		}
		if startOutsideTerraform != -1 && endOutsideTerraform == -1 && strings.HasPrefix(line, "Unless you have made equivalent changes to your configuration") {
			endOutsideTerraform = i + 1
		}
		if line == "Terraform will perform the following actions:" || line == "OpenTofu will perform the following actions:" {
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
			if startChangeOutput != -1 && endChangeOutput == -1 {
				endChangeOutput = i
			}
			if startWarning != -1 && endWarning == -1 {
				endWarning = i
			}
		}
		if p.Fail.MatchString(line) && startErrorIndex == -1 {
			startErrorIndex = i
		}
		if p.Pass.MatchString(line) || p.OutputsChanges.MatchString(line) {
			firstMatchLine = line
			break
		}
	}

	// Extract resources
	for i, line := range lines {
		if startChangeOutput != -1 && endChangeOutput != -1 && i >= startChangeOutput && i < endChangeOutput {
			if resource := extractResource(p.Create, line); resource != "" {
				createdResources = append(createdResources, resource)
			}
			if resource := extractResource(p.Update, line); resource != "" {
				updatedResources = append(updatedResources, resource)
			}
			if resource := extractResource(p.Delete, line); resource != "" {
				deletedResources = append(deletedResources, resource)
			}
			if resource := extractResource(p.Replace, line); resource != "" {
				replacedResources = append(replacedResources, resource)
			}
			if resource := extractResource(p.ReplaceOption, line); resource != "" {
				replacedResources = append(replacedResources, resource)
			}
			if resource := extractMovedResource(p.Move, line); resource != nil {
				movedResources = append(movedResources, resource)
			}
			if resource := extractResource(p.Import, line); resource != "" {
				importedResources = append(importedResources, resource)
			}
		}
	}

	// Extract outside terraform changes
	var outsideTerraform string
	if startOutsideTerraform != -1 && endOutsideTerraform != -1 {
		outsideTerraform = strings.Join(lines[startOutsideTerraform:endOutsideTerraform], "\n")
	}

	// Extract changed result
	var changedResult string
	if startChangeOutput != -1 && endChangeOutput != -1 {
		changedResult = strings.Join(lines[startChangeOutput:endChangeOutput], "\n")
	}

	// Extract warning
	var warning string
	if startWarning != -1 && endWarning != -1 {
		warning = strings.Join(lines[startWarning:endWarning], "\n")
	}

	hasDestroy := p.HasDestroy.MatchString(firstMatchLine)
	hasNoChanges := p.HasNoChanges.MatchString(firstMatchLine)
	hasError := p.Fail.MatchString(body)
	hasAddOrUpdateOnly := !hasNoChanges && !hasDestroy && !hasError

	return ParseResult{
		Result:             firstMatchLine,
		OutsideTerraform:   outsideTerraform,
		ChangedResult:      changedResult,
		Warning:            warning,
		HasAddOrUpdateOnly: hasAddOrUpdateOnly,
		HasDestroy:         hasDestroy,
		HasNoChanges:       hasNoChanges,
		HasError:           hasError,
		HasParseError:      false,
		ExitCode:           ExitPass,
		Error:              nil,
		CreatedResources:   createdResources,
		UpdatedResources:   updatedResources,
		DeletedResources:   deletedResources,
		ReplacedResources:  replacedResources,
		MovedResources:     movedResources,
		ImportedResources:  importedResources,
	}
}

// Parse returns ParseResult related with terraform apply
func (p *ApplyParser) Parse(body string) ParseResult {
	var exitCode int
	switch {
	case p.Pass.MatchString(body):
		exitCode = ExitPass
	case p.Fail.MatchString(body):
		exitCode = ExitFail
	default:
		return ParseResult{
			Result:   "",
			ExitCode: ExitFail,
			Error:    fmt.Errorf("cannot parse apply result"),
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
	case p.Pass.MatchString(line):
		result = lines[i]
	case p.Fail.MatchString(line):
		result = strings.Join(trimLastNewline(lines[i:]), "\n")
	}
	return ParseResult{
		Result:   result,
		ExitCode: exitCode,
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
