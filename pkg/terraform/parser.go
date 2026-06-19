package terraform

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
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
	// ModuleResults is populated only by TerragruntParser when Consolidated=true
	// and the parsed body contains 2+ modules (or at least one named module).
	// Consumer templates can use this to render a per-module Create/Update/etc
	// summary instead of the flat global lists. Nil otherwise; templates should
	// fall back to the flat lists when nil.
	ModuleResults []*ModuleResult
}

// ModuleResult is the per-module breakdown of resource changes in a
// consolidated Terragrunt run. The Module field carries the module path as
// seen in the run-all output (e.g. `cluster-citadel-2g/regions/tokyo/shared-vpc`).
// Root-module changes are labeled "Root module".
type ModuleResult struct {
	Module             string
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

// TerragruntParser is a parser for terragrunt run-all commands
type TerragruntParser struct {
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
	ModuleHeader   *regexp.Regexp
	LogModule      *regexp.Regexp
	PlanSummary    *regexp.Regexp
	ApplySummary   *regexp.Regexp
	ActionHeader   *regexp.Regexp
	Consolidated   bool
}

// NewPlanParser is PlanParser initialized with its Regexp
func NewPlanParser() *PlanParser {
	return &PlanParser{
		Pass:           regexp.MustCompile(`(?m)^(Plan: \d|No changes.)`),
		Fail:           regexp.MustCompile(`(?m)^([│|╵] )?(Error: )`),
		Warning:        regexp.MustCompile(`(?m)^([│|╵] )?(Warning: )`),
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
		Fail: regexp.MustCompile(`(?m)^([│|╵] )?(Error: )`),
	}
}

// NewTerragruntParser is TerragruntParser initialized with its Regexp
// Pass consolidated=true to enable consolidated output mode for terragrunt run-all
func NewTerragruntParser(consolidated bool) *TerragruntParser {
	// Prefix pattern handles: HH:MM:SS.mmm (STDOUT|STDERR|INFO|ERROR) [optional-module] (tfwrapper.sh|terraform|tf):
	prefix := `(?:\d{2}:\d{2}:\d{2}\.\d{3} (?:STDOUT|STDERR|INFO|ERROR)\s+(?:\[.*?\]\s+)?(?:tfwrapper\.sh|terraform|tf):\s*)?`

	return &TerragruntParser{
		Pass:           regexp.MustCompile(`(?m)^` + prefix + `(Plan: \d|No changes\.|Apply complete!)`),
		Fail:           regexp.MustCompile(`(?m)^` + prefix + `([│|╵] )?(Error: )`),
		Warning:        regexp.MustCompile(`(?m)^` + prefix + `([│|╵] )?(Warning: )`),
		OutputsChanges: regexp.MustCompile(`(?m)^` + prefix + `Changes to Outputs:`),
		HasDestroy:     regexp.MustCompile(`(?m)([1-9][0-9]* to destroy\.)`),
		HasNoChanges:   regexp.MustCompile(`(?m)^` + prefix + `(No changes\.|Plan: 0 to add, 0 to change, 0 to destroy\.)`),
		Create:         regexp.MustCompile(`^` + prefix + ` *# (.*) will be created$`),
		Update:         regexp.MustCompile(`^` + prefix + ` *# (.*) will be updated in-place$`),
		Delete:         regexp.MustCompile(`^` + prefix + ` *# (.*) will be destroyed$`),
		Replace:        regexp.MustCompile(`^` + prefix + ` *# (.*?)(?: is tainted, so)? must be replaced$`),
		ReplaceOption:  regexp.MustCompile(`^` + prefix + ` *# (.*?) will be replaced, as requested$`),
		Move:           regexp.MustCompile(`^` + prefix + ` *# (.*?) has moved to (.*?)$`),
		Import:         regexp.MustCompile(`^` + prefix + ` *# (.*?) will be imported$`),
		ImportedFrom:   regexp.MustCompile(`^` + prefix + ` *# \(imported from (.*?)\)$`),
		MovedFrom:      regexp.MustCompile(`^` + prefix + ` *# \(moved from (.*?)\)$`),
		ModuleHeader:   regexp.MustCompile(`^(?:(?:Group \d+)|(?:- )?Module) (.+?)(?:\s+\[run-all\])?$`),
		LogModule:      regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3} (?:STDOUT|STDERR|INFO|ERROR)\s+\[(.+?)\]\s`),
		PlanSummary:    regexp.MustCompile(`^Plan: (?:(\d+) to import, )?(\d+) to add, (\d+) to change, (\d+) to destroy\.`),
		ApplySummary:   regexp.MustCompile(`^Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed\.`),
		ActionHeader:   regexp.MustCompile(`^(?:Terraform|OpenTofu) will perform the following actions:$`),
		Consolidated:   consolidated,
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

// Parse returns ParseResult related with terragrunt run-all
func (p *TerragruntParser) Parse(body string) ParseResult {
	return p.ParseWithConsolidation(body, p.Consolidated)
}

// moduleSection holds the captured plan/apply output of a single terragrunt module
type moduleSection struct {
	name      string
	lines     []string
	capturing bool
}

// ParseWithConsolidation returns ParseResult with optional consolidation
// If consolidated=true, module names are included as section headers in ChangedResult
func (p *TerragruntParser) ParseWithConsolidation(body string, consolidated bool) ParseResult { //nolint:cyclop,maintidx,gocognit,funlen
	switch {
	case p.Fail.MatchString(body):
	case p.Pass.MatchString(body) || p.OutputsChanges.MatchString(body):
	default:
		return ParseResult{
			Result:        "",
			HasParseError: true,
			Error:         errors.New("cannot parse terragrunt result"),
		}
	}

	lines := strings.Split(body, "\n")

	var allCreatedResources, allUpdatedResources, allDeletedResources, allReplacedResources, allImportedResources []string
	var allMovedResources []*MovedResource
	var warnings []string
	var errorLines []string
	hasError := false

	var totalImport, totalAdd, totalChange, totalDestroy int
	planSummaryCount := 0
	applySummaryCount := 0
	noChangesCount := 0
	hasOutputsChanges := false

	// Per-module captured content, in order of first appearance
	var sections []*moduleSection
	sectionIndex := map[string]int{}
	getSection := func(name string) *moduleSection {
		if idx, ok := sectionIndex[name]; ok {
			return sections[idx]
		}
		sec := &moduleSection{name: name}
		sectionIndex[name] = len(sections)
		sections = append(sections, sec)
		return sec
	}

	// Per-module Create/Update/etc buckets, in order of first appearance.
	// Mirrors the flat all*Resources slices but keyed by module so consumer
	// templates can render a per-module summary in consolidated mode.
	var moduleResults []*ModuleResult
	moduleResultIndex := map[string]int{}
	getModuleResult := func(name string) *ModuleResult {
		if idx, ok := moduleResultIndex[name]; ok {
			return moduleResults[idx]
		}
		mr := &ModuleResult{Module: name}
		moduleResultIndex[name] = len(moduleResults)
		moduleResults = append(moduleResults, mr)
		return mr
	}

	currentModule := ""

	for i, line := range lines {
		stripped := stripTerragruntPrefix(line)

		// Determine module context. Prefer the [module/path] tag embedded in the
		// log prefix (terragrunt run-all interleaves output of modules, and the
		// "Group N / - Module <path>" queue listing is printed before execution
		// starts, so headers alone cannot attribute output lines correctly).
		if m := p.LogModule.FindStringSubmatch(line); len(m) == 2 { //nolint:mnd
			currentModule = m[1]
		} else if m := p.ModuleHeader.FindStringSubmatch(line); len(m) > 1 {
			name := m[1]
			if name == "." {
				name = ""
			}
			currentModule = name
			continue
		}

		// Collect error output from the first error line to the end
		if hasError {
			errorLines = append(errorLines, stripped)
		} else if p.Fail.MatchString(line) {
			hasError = true
			errorLines = append(errorLines, stripped)
		}

		// Aggregate plan/apply summaries across all modules
		summaryLine := false
		if m := p.PlanSummary.FindStringSubmatch(stripped); len(m) == 5 { //nolint:mnd
			imported, _ := strconv.Atoi(m[1]) // empty when the plan has no imports
			add, _ := strconv.Atoi(m[2])
			change, _ := strconv.Atoi(m[3])
			destroy, _ := strconv.Atoi(m[4])
			totalImport += imported
			totalAdd += add
			totalChange += change
			totalDestroy += destroy
			planSummaryCount++
			summaryLine = true
		} else if m := p.ApplySummary.FindStringSubmatch(stripped); len(m) == 4 { //nolint:mnd
			add, _ := strconv.Atoi(m[1])
			change, _ := strconv.Atoi(m[2])
			destroy, _ := strconv.Atoi(m[3])
			totalAdd += add
			totalChange += change
			totalDestroy += destroy
			applySummaryCount++
			summaryLine = true
		} else if p.HasNoChanges.MatchString(line) {
			noChangesCount++
			summaryLine = true
		}
		if p.OutputsChanges.MatchString(line) {
			hasOutputsChanges = true
		}

		// Capture per-module plan/apply content (everything between the action
		// header or first resource action line and the summary line, inclusive)
		if !hasError {
			sec := getSection(currentModule)
			if sec.capturing { //nolint:gocritic
				sec.lines = append(sec.lines, stripped)
				if summaryLine {
					sec.capturing = false
				}
			} else if p.ActionHeader.MatchString(stripped) ||
				strings.HasPrefix(stripped, "Changes to Outputs:") ||
				p.Create.MatchString(line) || p.Update.MatchString(line) ||
				p.Delete.MatchString(line) || p.Replace.MatchString(line) ||
				p.ReplaceOption.MatchString(line) || p.Import.MatchString(line) ||
				p.Move.MatchString(line) {
				sec.capturing = true
				sec.lines = append(sec.lines, stripped)
			} else if summaryLine {
				sec.lines = append(sec.lines, stripped)
			}
		}

		// Extract resources. Each match goes to both the flat global slice and
		// the per-module bucket (the latter drives per-module rendering in
		// consolidated mode; flat slices stay for templates that haven't migrated).
		mr := getModuleResult(currentModule)
		if rsc := extractResource(p.Create, line); rsc != "" {
			allCreatedResources = append(allCreatedResources, rsc)
			mr.CreatedResources = append(mr.CreatedResources, rsc)
		} else if rsc := extractResource(p.Update, line); rsc != "" {
			allUpdatedResources = append(allUpdatedResources, rsc)
			mr.UpdatedResources = append(mr.UpdatedResources, rsc)
		} else if rsc := extractResource(p.Delete, line); rsc != "" {
			allDeletedResources = append(allDeletedResources, rsc)
			mr.DeletedResources = append(mr.DeletedResources, rsc)
		} else if rsc := extractResource(p.Replace, line); rsc != "" {
			allReplacedResources = append(allReplacedResources, rsc)
			mr.ReplacedResources = append(mr.ReplacedResources, rsc)
		} else if rsc := extractResource(p.ReplaceOption, line); rsc != "" {
			allReplacedResources = append(allReplacedResources, rsc)
			mr.ReplacedResources = append(mr.ReplacedResources, rsc)
		} else if rsc := extractResource(p.Import, line); rsc != "" {
			allImportedResources = append(allImportedResources, rsc)
			mr.ImportedResources = append(mr.ImportedResources, rsc)
		} else if rsc := extractResource(p.ImportedFrom, line); rsc != "" {
			if i > 0 {
				if toRsc := p.changedResources(lines[i-1]); toRsc != "" {
					allImportedResources = append(allImportedResources, toRsc)
					mr.ImportedResources = append(mr.ImportedResources, toRsc)
				}
			}
		} else if rsc := extractMovedResource(p.Move, line); rsc != nil {
			allMovedResources = append(allMovedResources, rsc)
			mr.MovedResources = append(mr.MovedResources, rsc)
		} else if fromRsc := extractResource(p.MovedFrom, line); fromRsc != "" {
			if i > 0 {
				if toRsc := p.changedResources(lines[i-1]); toRsc != "" {
					mvd := &MovedResource{Before: fromRsc, After: toRsc}
					allMovedResources = append(allMovedResources, mvd)
					mr.MovedResources = append(mr.MovedResources, mvd)
				}
			}
		}

		// Collect warnings
		if p.Warning.MatchString(line) {
			warnings = append(warnings, stripped)
		}
	}

	// A run has no changes only when no module reported any change at all
	hasDestroy := totalDestroy > 0
	hasNoChanges := !hasError && applySummaryCount == 0 &&
		totalImport == 0 && totalAdd == 0 && totalChange == 0 && totalDestroy == 0 &&
		(noChangesCount > 0 || planSummaryCount > 0)

	var result string
	switch {
	case hasError:
		result = strings.Join(trimBars(trimLastNewline(errorLines)), "\n")
	case applySummaryCount > 0:
		result = fmt.Sprintf("Apply complete! Resources: %d added, %d changed, %d destroyed.", totalAdd, totalChange, totalDestroy)
	case hasNoChanges:
		result = "No changes. Your infrastructure matches the configuration."
	case planSummaryCount > 0:
		if totalImport > 0 {
			result = fmt.Sprintf("Plan: %d to import, %d to add, %d to change, %d to destroy.", totalImport, totalAdd, totalChange, totalDestroy)
		} else {
			result = fmt.Sprintf("Plan: %d to add, %d to change, %d to destroy.", totalAdd, totalChange, totalDestroy)
		}
	case hasOutputsChanges:
		result = "Only Outputs will be changed."
	}

	// Build the changed result, one section per module
	var changeResults []string
	for _, sec := range sections {
		if len(sec.lines) == 0 {
			continue
		}
		content := strings.TrimRight(strings.Join(sec.lines, "\n"), "\n ")
		if content == "" {
			continue
		}
		if consolidated && sec.name != "" {
			content = "Module: " + sec.name + "\n\n" + content
		}
		changeResults = append(changeResults, content)
	}

	hasAddOrUpdateOnly := !hasNoChanges && !hasDestroy && !hasError

	// Build ModuleResults only in consolidated mode and only when we have
	// something useful to render. "Useful" = at least one named module with
	// changes, OR multiple modules with changes (so the per-module split adds
	// signal even if all are unnamed). Single-unnamed-module input falls
	// through to the flat lists.
	var emitModuleResults []*ModuleResult
	if consolidated {
		namedWithChanges := 0
		withChanges := make([]*ModuleResult, 0, len(moduleResults))
		for _, mr := range moduleResults {
			if len(mr.CreatedResources)+len(mr.UpdatedResources)+len(mr.DeletedResources)+
				len(mr.ReplacedResources)+len(mr.MovedResources)+len(mr.ImportedResources) == 0 {
				continue
			}
			withChanges = append(withChanges, mr)
			if mr.Module != "" {
				namedWithChanges++
			}
		}
		if namedWithChanges > 0 || len(withChanges) > 1 {
			for _, mr := range withChanges {
				if mr.Module == "" {
					mr.Module = "Root module"
				}
			}
			emitModuleResults = withChanges
		}
	}

	return ParseResult{
		Result:             strings.TrimSpace(result),
		ChangedResult:      strings.Join(changeResults, "\n\n"),
		Warning:            strings.TrimSpace(strings.Join(warnings, "\n")),
		HasAddOrUpdateOnly: hasAddOrUpdateOnly,
		HasDestroy:         hasDestroy,
		HasNoChanges:       hasNoChanges,
		HasError:           hasError,
		Error:              nil,
		CreatedResources:   allCreatedResources,
		UpdatedResources:   allUpdatedResources,
		DeletedResources:   allDeletedResources,
		ReplacedResources:  allReplacedResources,
		MovedResources:     allMovedResources,
		ImportedResources:  allImportedResources,
		ModuleResults:      emitModuleResults,
	}
}

func (p *TerragruntParser) changedResources(line string) string {
	if rsc := extractResource(p.Update, line); rsc != "" {
		return rsc
	} else if rsc := extractResource(p.Replace, line); rsc != "" {
		return rsc
	} else if rsc := extractResource(p.ReplaceOption, line); rsc != "" {
		return rsc
	}
	return ""
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

// terragruntPrefixRe matches Terragrunt timestamp prefixes.
// The trailing `: ?` (optional single space) eats exactly the one separator
// space terragrunt inserts between the colon and the wrapped command's output,
// but preserves the wrapped command's own indentation. With the previous
// `:\s*`, the 2/4/6-space indentation of `terraform plan` was eaten too,
// causing diff lines like `+ field = ...` to render at column 0 in the
// rendered Change Result block.
var terragruntPrefixRe = regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3} (?:STDOUT|STDERR|INFO|ERROR)\s+(?:\[.*?\]\s+)?(?:tfwrapper\.sh|terraform|tf): ?`)

// stripTerragruntPrefix removes Terragrunt timestamp prefixes
func stripTerragruntPrefix(line string) string {
	return terragruntPrefixRe.ReplaceAllString(line, "")
}
