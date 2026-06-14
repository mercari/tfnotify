package terraform

import (
	"strings"
	"testing"
)

func TestTerragruntParser_Parse(t *testing.T) {
	parser := NewTerragruntParser(false)

	tests := []struct {
		name              string
		input             string
		wantHasError      bool
		wantHasDestroy    bool
		wantHasNoChanges  bool
		wantCreatedCount  int
		wantUpdatedCount  int
		wantDeletedCount  int
		wantReplacedCount int
	}{
		{
			name: "single module with timestamp prefix",
			input: `09:32:46.963 STDOUT terraform:
09:32:46.963 STDOUT terraform: Terraform will perform the following actions:
09:32:46.963 STDOUT terraform:
09:32:46.963 STDOUT terraform:   # null_resource.test will be created
09:32:46.963 STDOUT terraform:   + resource "null_resource" "test" {
09:32:46.963 STDOUT terraform:       + id = (known after apply)
09:32:46.963 STDOUT terraform:     }
09:32:46.963 STDOUT terraform:
09:32:46.963 STDOUT terraform: Plan: 1 to add, 0 to change, 0 to destroy.`,
			wantHasError:     false,
			wantHasDestroy:   false,
			wantHasNoChanges: false,
			wantCreatedCount: 1,
		},
		{
			name: "multiple modules with run-all",
			input: `Group 1
- Module /path/to/app1

09:32:46.963 STDOUT terraform:   # null_resource.app1 will be created
09:32:46.963 STDOUT terraform:   + resource "null_resource" "app1" {
09:32:46.963 STDOUT terraform:       + id = (known after apply)
09:32:46.963 STDOUT terraform:     }
09:32:46.963 STDOUT terraform: Plan: 1 to add, 0 to change, 0 to destroy.

Group 2
- Module /path/to/app2

09:33:12.145 STDOUT terraform:   # null_resource.app2 will be created
09:33:12.145 STDOUT terraform:   + resource "null_resource" "app2" {
09:33:12.145 STDOUT terraform:       + id = (known after apply)
09:33:12.145 STDOUT terraform:     }
09:33:12.145 STDOUT terraform: Plan: 1 to add, 0 to change, 0 to destroy.`,
			wantHasError:     false,
			wantHasDestroy:   false,
			wantHasNoChanges: false,
			wantCreatedCount: 2,
		},
		{
			name: "no changes",
			input: `09:32:46.963 STDOUT terraform:
09:32:46.963 STDOUT terraform: No changes. Your infrastructure matches the configuration.`,
			wantHasError:     false,
			wantHasDestroy:   false,
			wantHasNoChanges: true,
			wantCreatedCount: 0,
		},
		{
			name: "with destroy",
			input: `09:32:46.963 STDOUT terraform:
09:32:46.963 STDOUT terraform:   # null_resource.test will be destroyed
09:32:46.963 STDOUT terraform:   - resource "null_resource" "test" {
09:32:46.963 STDOUT terraform:       - id = "123" -> null
09:32:46.963 STDOUT terraform:     }
09:32:46.963 STDOUT terraform: Plan: 0 to add, 0 to change, 1 to destroy.`,
			wantHasError:     false,
			wantHasDestroy:   true,
			wantHasNoChanges: false,
			wantDeletedCount: 1,
		},
		{
			name: "with update",
			input: `09:32:46.963 STDOUT terraform:
09:32:46.963 STDOUT terraform:   # null_resource.test will be updated in-place
09:32:46.963 STDOUT terraform:   ~ resource "null_resource" "test" {
09:32:46.963 STDOUT terraform:       id = "123"
09:32:46.963 STDOUT terraform:     }
09:32:46.963 STDOUT terraform: Plan: 0 to add, 1 to change, 0 to destroy.`,
			wantHasError:     false,
			wantHasDestroy:   false,
			wantHasNoChanges: false,
			wantUpdatedCount: 1,
		},
		{
			name: "with replace",
			input: `09:32:46.963 STDOUT terraform:
09:32:46.963 STDOUT terraform:   # null_resource.test must be replaced
09:32:46.963 STDOUT terraform: -/+ resource "null_resource" "test" {
09:32:46.963 STDOUT terraform:       id = "123"
09:32:46.963 STDOUT terraform:     }
09:32:46.963 STDOUT terraform: Plan: 1 to add, 0 to change, 1 to destroy.`,
			wantHasError:      false,
			wantHasDestroy:    true,
			wantHasNoChanges:  false,
			wantReplacedCount: 1,
		},
		{
			name: "with error",
			input: `09:32:46.963 STDOUT terraform:
09:32:46.963 STDERR terraform: Error: Invalid configuration
09:32:46.963 STDERR terraform:
09:32:46.963 STDERR terraform: Something went wrong`,
			wantHasError: true,
		},
		{
			name: "without timestamp prefix (TERRAGRUNT_LOG_DISABLE=true)",
			input: `Terraform will perform the following actions:

  # null_resource.test will be created
  + resource "null_resource" "test" {
      + id = (known after apply)
    }

Plan: 1 to add, 0 to change, 0 to destroy.`,
			wantHasError:     false,
			wantHasDestroy:   false,
			wantHasNoChanges: false,
			wantCreatedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			if result.HasError != tt.wantHasError {
				t.Errorf("HasError = %v, want %v", result.HasError, tt.wantHasError)
			}
			if result.HasDestroy != tt.wantHasDestroy {
				t.Errorf("HasDestroy = %v, want %v", result.HasDestroy, tt.wantHasDestroy)
			}
			if result.HasNoChanges != tt.wantHasNoChanges {
				t.Errorf("HasNoChanges = %v, want %v", result.HasNoChanges, tt.wantHasNoChanges)
			}
			if len(result.CreatedResources) != tt.wantCreatedCount {
				t.Errorf("CreatedResources count = %d, want %d. Resources: %v",
					len(result.CreatedResources), tt.wantCreatedCount, result.CreatedResources)
			}
			if len(result.UpdatedResources) != tt.wantUpdatedCount {
				t.Errorf("UpdatedResources count = %d, want %d",
					len(result.UpdatedResources), tt.wantUpdatedCount)
			}
			if len(result.DeletedResources) != tt.wantDeletedCount {
				t.Errorf("DeletedResources count = %d, want %d",
					len(result.DeletedResources), tt.wantDeletedCount)
			}
			if len(result.ReplacedResources) != tt.wantReplacedCount {
				t.Errorf("ReplacedResources count = %d, want %d",
					len(result.ReplacedResources), tt.wantReplacedCount)
			}
		})
	}
}

func TestTerragruntParser_ParseWithConsolidation(t *testing.T) {
	parser := NewTerragruntParser(false)

	multiModuleInput := `Group 1
Module /path/to/app1

09:32:46.963 STDOUT terraform:   # null_resource.app1_resource will be created
09:32:46.963 STDOUT terraform:   + resource "null_resource" "app1_resource" {
09:32:46.963 STDOUT terraform:       + id = (known after apply)
09:32:46.963 STDOUT terraform:     }
09:32:46.963 STDOUT terraform:   # null_resource.app1_data will be created
09:32:46.963 STDOUT terraform:   + resource "null_resource" "app1_data" {
09:32:46.963 STDOUT terraform:       + id = (known after apply)
09:32:46.963 STDOUT terraform:     }
09:32:46.963 STDOUT terraform: Plan: 2 to add, 0 to change, 0 to destroy.

Group 2
Module /path/to/app2

09:33:12.145 STDOUT terraform:   # null_resource.app2_resource will be created
09:33:12.145 STDOUT terraform:   + resource "null_resource" "app2_resource" {
09:33:12.145 STDOUT terraform:       + id = (known after apply)
09:33:12.145 STDOUT terraform:     }
09:33:12.145 STDOUT terraform: Plan: 1 to add, 0 to change, 0 to destroy.`

	t.Run("consolidated mode", func(t *testing.T) {
		result := parser.ParseWithConsolidation(multiModuleInput, true)

		// Should aggregate resources from all modules
		if len(result.CreatedResources) != 3 {
			t.Errorf("CreatedResources count = %d, want 3. Resources: %v",
				len(result.CreatedResources), result.CreatedResources)
		}

		// Check that all resources are captured
		expectedResources := map[string]bool{
			"null_resource.app1_resource": false,
			"null_resource.app1_data":     false,
			"null_resource.app2_resource": false,
		}

		for _, rsc := range result.CreatedResources {
			if _, exists := expectedResources[rsc]; exists {
				expectedResources[rsc] = true
			}
		}

		for rsc, found := range expectedResources {
			if !found {
				t.Errorf("Expected resource %s not found in results", rsc)
			}
		}

		// Check that the output contains the plan details
		if result.ChangedResult == "" {
			t.Error("ChangedResult is empty, expected aggregated plan output")
		} else {
			t.Logf("ChangedResult content:\n%s", result.ChangedResult)
		}

		if !strings.Contains(result.ChangedResult, "/path/to/app1") {
			t.Error("ChangedResult missing /path/to/app1 summary")
		}
		if !strings.Contains(result.ChangedResult, "null_resource.app1_resource will be created") {
			t.Error("ChangedResult missing app1_resource details")
		}
	})

	t.Run("non-consolidated mode", func(t *testing.T) {
		result := parser.ParseWithConsolidation(multiModuleInput, false)

		// Should still parse all resources
		if len(result.CreatedResources) != 3 {
			t.Errorf("CreatedResources count = %d, want 3", len(result.CreatedResources))
		}
	})
}

func TestTerragruntParser_ConsolidationMixedNoChanges(t *testing.T) {
	// A module reporting "No changes." before a module with changes must not
	// mark the whole run as no-changes (regression test)
	parser := NewTerragruntParser(true)

	input := `18:13:47.247 STDOUT [cluster/shared-vpc] tfwrapper.sh: No changes. Your infrastructure matches the configuration.
18:14:11.466 STDOUT [cluster/cluster] tfwrapper.sh: Terraform will perform the following actions:
18:14:11.466 STDOUT [cluster/cluster] tfwrapper.sh:
18:14:11.466 STDOUT [cluster/cluster] tfwrapper.sh:   # null_resource.test will be created
18:14:11.466 STDOUT [cluster/cluster] tfwrapper.sh:   + resource "null_resource" "test" {
18:14:11.466 STDOUT [cluster/cluster] tfwrapper.sh:       + id = (known after apply)
18:14:11.466 STDOUT [cluster/cluster] tfwrapper.sh:     }
18:14:11.466 STDOUT [cluster/cluster] tfwrapper.sh:
18:14:11.466 STDOUT [cluster/cluster] tfwrapper.sh: Plan: 1 to add, 0 to change, 0 to destroy.`

	result := parser.Parse(input)

	if result.HasParseError {
		t.Fatalf("unexpected parse error: %v", result.Error)
	}
	if result.HasNoChanges {
		t.Error("HasNoChanges = true, want false (one module has changes)")
	}
	if result.Result != "Plan: 1 to add, 0 to change, 0 to destroy." {
		t.Errorf("Result = %q, want aggregated plan summary", result.Result)
	}
	if len(result.CreatedResources) != 1 {
		t.Errorf("CreatedResources count = %d, want 1", len(result.CreatedResources))
	}
	// Module attribution must come from the [module/path] log prefix
	if !strings.Contains(result.ChangedResult, "cluster/cluster") {
		t.Errorf("ChangedResult missing module name from log prefix:\n%s", result.ChangedResult)
	}
	// Nested diff lines must be preserved
	if !strings.Contains(result.ChangedResult, "+ id = (known after apply)") {
		t.Errorf("ChangedResult missing nested diff line:\n%s", result.ChangedResult)
	}
	if !strings.Contains(result.ChangedResult, "Plan: 1 to add, 0 to change, 0 to destroy.") {
		t.Errorf("ChangedResult missing plan summary:\n%s", result.ChangedResult)
	}
}

func TestTerragruntParser_ConsolidationAggregatesTotals(t *testing.T) {
	parser := NewTerragruntParser(true)

	input := `Group 1
Module /path/to/app1

09:32:46.963 STDOUT terraform:   # null_resource.app1 will be created
09:32:46.963 STDOUT terraform: Plan: 2 to add, 1 to change, 0 to destroy.

Group 2
Module /path/to/app2

09:33:12.145 STDOUT terraform:   # null_resource.app2 will be destroyed
09:33:12.145 STDOUT terraform: Plan: 1 to add, 0 to change, 3 to destroy.`

	result := parser.Parse(input)

	if result.Result != "Plan: 3 to add, 1 to change, 3 to destroy." {
		t.Errorf("Result = %q, want aggregated totals", result.Result)
	}
	if !result.HasDestroy {
		t.Error("HasDestroy = false, want true")
	}
	if result.HasNoChanges {
		t.Error("HasNoChanges = true, want false")
	}
}

func TestTerragruntParser_ErrorKeepsDetails(t *testing.T) {
	parser := NewTerragruntParser(true)

	input := `09:32:46.963 STDOUT terraform: Initializing...
09:32:46.963 STDERR terraform: Error: Invalid configuration
09:32:46.963 STDERR terraform:
09:32:46.963 STDERR terraform: Something went wrong
09:32:46.963 STDERR terraform: on main.tf line 5`

	result := parser.Parse(input)

	if !result.HasError {
		t.Fatal("HasError = false, want true")
	}
	// The full error details (prefix-stripped) must be preserved, not just the first line
	if !strings.Contains(result.Result, "Error: Invalid configuration") {
		t.Errorf("Result missing error headline: %q", result.Result)
	}
	if !strings.Contains(result.Result, "Something went wrong") {
		t.Errorf("Result missing error details: %q", result.Result)
	}
	if strings.Contains(result.Result, "STDERR") {
		t.Errorf("Result should not contain log prefixes: %q", result.Result)
	}
}

func TestTerragruntParser_PlanSummaryWithImports(t *testing.T) {
	// Terraform >= 1.5 prints "Plan: N to import, ..." when import blocks are used
	parser := NewTerragruntParser(true)

	input := `09:32:46.963 STDOUT terraform:   # null_resource.test will be imported
09:32:46.963 STDOUT terraform: Plan: 1 to import, 2 to add, 0 to change, 0 to destroy.`

	result := parser.Parse(input)

	if result.HasParseError {
		t.Fatalf("unexpected parse error: %v", result.Error)
	}
	if result.HasNoChanges {
		t.Error("HasNoChanges = true, want false")
	}
	if result.Result != "Plan: 1 to import, 2 to add, 0 to change, 0 to destroy." {
		t.Errorf("Result = %q, want import-aware plan summary", result.Result)
	}
	if len(result.ImportedResources) != 1 {
		t.Errorf("ImportedResources count = %d, want 1", len(result.ImportedResources))
	}
}

func TestTerragruntParser_ApplyConsolidated(t *testing.T) {
	parser := NewTerragruntParser(true)

	input := `18:13:47.247 STDOUT [app1] tfwrapper.sh: Apply complete! Resources: 2 added, 1 changed, 0 destroyed.
18:14:11.466 STDOUT [app2] tfwrapper.sh: Apply complete! Resources: 1 added, 0 changed, 1 destroyed.`

	result := parser.Parse(input)

	if result.HasParseError {
		t.Fatalf("unexpected parse error: %v", result.Error)
	}
	if result.HasError {
		t.Error("HasError = true, want false")
	}
	if result.Result != "Apply complete! Resources: 3 added, 1 changed, 1 destroyed." {
		t.Errorf("Result = %q, want aggregated apply summary", result.Result)
	}
}

func TestStripTerragruntPrefix(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "with STDOUT prefix",
			input: "09:32:46.963 STDOUT terraform: Plan: 1 to add",
			want:  "Plan: 1 to add",
		},
		{
			name:  "with STDERR prefix",
			input: "09:32:46.963 STDERR terraform: Error: something failed",
			want:  "Error: something failed",
		},
		{
			name:  "with INFO prefix",
			input: "18:13:47.247 INFO tfwrapper.sh: Terraform has been successfully initialized!",
			want:  "Terraform has been successfully initialized!",
		},
		{
			name:  "with ERROR prefix",
			input: "18:13:47.247 ERROR tfwrapper.sh: Error: something failed",
			want:  "Error: something failed",
		},
		{
			name:  "with module path",
			input: "18:13:47.247 ERROR  [cluster-citadel-2g/regions/tokyo/shared-vpc] tfwrapper.sh: No changes.",
			want:  "No changes.",
		},
		{
			name:  "with STDOUT and module path",
			input: "18:14:11.466 STDOUT [cluster-citadel-2g/regions/tokyo/cluster] tfwrapper.sh: Plan: 1 to add, 0 to change, 0 to destroy.",
			want:  "Plan: 1 to add, 0 to change, 0 to destroy.",
		},
		{
			name:  "without prefix",
			input: "Plan: 1 to add",
			want:  "Plan: 1 to add",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripTerragruntPrefix(tt.input)
			if got != tt.want {
				t.Errorf("stripTerragruntPrefix() = %q, want %q", got, tt.want)
			}
		})
	}
}
