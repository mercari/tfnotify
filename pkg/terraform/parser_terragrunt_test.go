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
