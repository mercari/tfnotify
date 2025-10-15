package terraform

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const planSuccessResult = `
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

data.terraform_remote_state.teams_platform_development: Refreshing state...
google_project.my_project: Refreshing state...
aws_iam_policy.datadog_aws_integration: Refreshing state...
aws_iam_user.teams_terraform: Refreshing state...
aws_iam_role.datadog_aws_integration: Refreshing state...
google_project_services.my_project: Refreshing state...
google_bigquery_dataset.gateway_access_log: Refreshing state...
aws_iam_role_policy_attachment.datadog_aws_integration: Refreshing state...
google_logging_project_sink.gateway_access_log_bigquery_sink: Refreshing state...
google_project_iam_member.gateway_access_log_bigquery_sink_writer_is_bigquery_data_editor: Refreshing state...
google_dns_managed_zone.tfnotifyapps_com: Refreshing state...
google_dns_record_set.dev_tfnotifyapps_com: Refreshing state...

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  + google_compute_global_address.my_another_project
      id:         <computed>
      address:    <computed>
      ip_version: "IPV4"
      name:       "my-another-project"
      project:    "my-project"
      self_link:  <computed>


Plan: 1 to add, 0 to change, 0 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.
`

const planOnlyOutputChangesSuccessResult0_12 = `
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

data.terraform_remote_state.teams_platform_development: Refreshing state...
google_project.my_project: Refreshing state...
aws_iam_policy.datadog_aws_integration: Refreshing state...
aws_iam_user.teams_terraform: Refreshing state...
aws_iam_role.datadog_aws_integration: Refreshing state...
google_project_services.my_project: Refreshing state...
google_bigquery_dataset.gateway_access_log: Refreshing state...
aws_iam_role_policy_attachment.datadog_aws_integration: Refreshing state...
google_logging_project_sink.gateway_access_log_bigquery_sink: Refreshing state...
google_project_iam_member.gateway_access_log_bigquery_sink_writer_is_bigquery_data_editor: Refreshing state...
google_dns_managed_zone.tfnotifyapps_com: Refreshing state...
google_dns_record_set.dev_tfnotifyapps_com: Refreshing state...

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

Plan: 0 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + aws_instance_name = "my-instance"

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.
`

const planOnlyOutputChangesSuccessResult0_15 = `
null_resource.this: Refreshing state... [id=6068603774747257119]

Changes to Outputs:
  + test = 42

You can apply this plan to save these new output values to the Terraform
state, without changing any real infrastructure.

─────────────────────────────────────────────────────────────────────────────

Note: You didn't use the -out option to save this plan, so Terraform can't
guarantee to take exactly these actions if you run "terraform apply" now.
`

const planOnlyOutputChangesSuccessInAutomationResult = `
null_resource.this: Refreshing state... [id=6068603774747257119]

Changes to Outputs:
  + test = 42

You can apply this plan to save these new output values to the Terraform
state, without changing any real infrastructure.
`

const planFailureResult = `
xxxxxxxxx
xxxxxxxxx
xxxxxxxxx

Error: Error refreshing state: 4 error(s) occurred:

* google_sql_database.main: 1 error(s) occurred:

* google_sql_database.main: google_sql_database.main: Error reading SQL Database "main" in instance "main-master-instance": googleapi: Error 409: The instance or operation is not in an appropriate state to handle the request., invalidState
* google_sql_user.proxyuser_main: 1 error(s) occurred:
`

const planNoChanges = `
google_bigquery_dataset.tfnotify_echo: Refreshing state...
google_project.team: Refreshing state...
pagerduty_team.team: Refreshing state...
data.pagerduty_vendor.datadog: Refreshing state...
data.pagerduty_user.service_owner[1]: Refreshing state...
data.pagerduty_user.service_owner[2]: Refreshing state...
data.pagerduty_user.service_owner[0]: Refreshing state...
google_project_services.team: Refreshing state...
google_project_iam_member.team[1]: Refreshing state...
google_project_iam_member.team[2]: Refreshing state...
google_project_iam_member.team[0]: Refreshing state...
google_project_iam_member.team_platform[1]: Refreshing state...
google_project_iam_member.team_platform[2]: Refreshing state...
google_project_iam_member.team_platform[0]: Refreshing state...
pagerduty_team_membership.team[2]: Refreshing state...
pagerduty_schedule.secondary: Refreshing state...
pagerduty_schedule.primary: Refreshing state...
pagerduty_team_membership.team[0]: Refreshing state...
pagerduty_team_membership.team[1]: Refreshing state...
pagerduty_escalation_policy.team: Refreshing state...
pagerduty_service.team: Refreshing state...
pagerduty_service_integration.datadog: Refreshing state...

------------------------------------------------------------------------

No changes. Infrastructure is up-to-date.

This means that Terraform did not detect any differences between your
configuration and real physical resources that exist. As a result, no
actions need to be performed.
`

const planHasDestroy = `
google_bigquery_dataset.tfnotify_echo: Refreshing state...
google_project.team: Refreshing state...
pagerduty_team.team: Refreshing state...
data.pagerduty_vendor.datadog: Refreshing state...
data.pagerduty_user.service_owner[1]: Refreshing state...
data.pagerduty_user.service_owner[2]: Refreshing state...
data.pagerduty_user.service_owner[0]: Refreshing state...
google_project_services.team: Refreshing state...
google_project_iam_member.team[1]: Refreshing state...
google_project_iam_member.team[2]: Refreshing state...
google_project_iam_member.team[0]: Refreshing state...
google_project_iam_member.team_platform[1]: Refreshing state...
google_project_iam_member.team_platform[2]: Refreshing state...
google_project_iam_member.team_platform[0]: Refreshing state...
pagerduty_team_membership.team[2]: Refreshing state...
pagerduty_schedule.secondary: Refreshing state...
pagerduty_schedule.primary: Refreshing state...
pagerduty_team_membership.team[0]: Refreshing state...
pagerduty_team_membership.team[1]: Refreshing state...
pagerduty_escalation_policy.team: Refreshing state...
pagerduty_service.team: Refreshing state...
pagerduty_service_integration.datadog: Refreshing state...

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  - google_project_iam_member.team_platform[2]


Plan: 0 to add, 0 to change, 1 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.
`

const planHasAddAndDestroy = `
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

data.terraform_remote_state.teams_platform_development: Refreshing state...
google_project.my_project: Refreshing state...
aws_iam_policy.datadog_aws_integration: Refreshing state...
aws_iam_user.teams_terraform: Refreshing state...
aws_iam_role.datadog_aws_integration: Refreshing state...
google_project_services.my_project: Refreshing state...
google_bigquery_dataset.gateway_access_log: Refreshing state...
aws_iam_role_policy_attachment.datadog_aws_integration: Refreshing state...
google_logging_project_sink.gateway_access_log_bigquery_sink: Refreshing state...
google_project_iam_member.gateway_access_log_bigquery_sink_writer_is_bigquery_data_editor: Refreshing state...
google_dns_managed_zone.tfnotifyapps_com: Refreshing state...
google_dns_record_set.dev_tfnotifyapps_com: Refreshing state...
google_project_iam_member.team_platform[1]: Refreshing state...
google_project_iam_member.team_platform[2]: Refreshing state...
google_project_iam_member.team_platform[0]: Refreshing state...

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create
  - destroy

Terraform will perform the following actions:

  + google_compute_global_address.my_another_project
      id:         <computed>
      address:    <computed>
      ip_version: "IPV4"
      name:       "my-another-project"
      project:    "my-project"
      self_link:  <computed>

  - google_project_iam_member.team_platform[2]

Plan: 1 to add, 0 to change, 1 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.
`

const planHasAddAndUpdateInPlace = `
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

data.terraform_remote_state.teams_platform_development: Refreshing state...
google_project.my_project: Refreshing state...
aws_iam_policy.datadog_aws_integration: Refreshing state...
aws_iam_user.teams_terraform: Refreshing state...
aws_iam_role.datadog_aws_integration: Refreshing state...
google_project_services.my_project: Refreshing state...
google_bigquery_dataset.gateway_access_log: Refreshing state...
aws_iam_role_policy_attachment.datadog_aws_integration: Refreshing state...
google_logging_project_sink.gateway_access_log_bigquery_sink: Refreshing state...
google_project_iam_member.gateway_access_log_bigquery_sink_writer_is_bigquery_data_editor: Refreshing state...
google_dns_managed_zone.tfnotifyapps_com: Refreshing state...
google_dns_record_set.dev_tfnotifyapps_com: Refreshing state...
google_project_iam_member.team_platform[1]: Refreshing state...
google_project_iam_member.team_platform[2]: Refreshing state...
google_project_iam_member.team_platform[0]: Refreshing state...

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create
  ~ update in-place

Terraform will perform the following actions:

  + google_compute_global_address.my_another_project
      id:         <computed>
      address:    <computed>
      ip_version: "IPV4"
      name:       "my-another-project"
      project:    "my-project"
      self_link:  <computed>

  ~ google_project_iam_member.team_platform[2]

Plan: 1 to add, 1 to change, 0 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.
`

const planImportedMovedResourceChanged = `
null_resource.bar: Refreshing state... [id=7822522400686116714]
github_repository.tfnotify: Preparing import... [id=tfnotify]
github_repository.tfnotify: Refreshing state... [id=tfnotify]
github_issue.test-2: Refreshing state... [id=tfaction:902]
github_repository.tfaction-2: Refreshing state... [id=tfaction]

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following
symbols:
  + create
  ~ update in-place
-/+ destroy and then create replacement

Terraform will perform the following actions:

  # github_issue.test-2 must be replaced
  # (moved from github_issue.test)
-/+ resource "github_issue" "test-2" {
      - assignees        = [] -> null
      ~ etag             = "W/\"e14116be2014dddd5d766ba8d69e59524491648d626917f4dbe8d5422ef32ba3\"" -> (known after apply)
      ~ id               = "tfaction:902" -> (known after apply)
      ~ issue_id         = 1685922946 -> (known after apply)
      - labels           = [
          - "enhancement",
        ] -> null
      - milestone_number = 0 -> null
      ~ number           = 902 -> (known after apply)
      ~ repository       = "tfaction" -> "tfnotify" # forces replacement
        # (1 unchanged attribute hidden)
    }

  # github_repository.tfaction-2 will be updated in-place
  # (moved from github_repository.tfaction)
  ~ resource "github_repository" "tfaction-2" {
      ~ allow_auto_merge            = true -> false
      - allow_update_branch         = true -> null
      ~ delete_branch_on_merge      = true -> false
      - description                 = "Framework for Monorepo to build high level Terraform Workflows by GitHub Actions" -> null
      ~ full_name                   = "mercari/tfaction" -> (known after apply)
      - has_discussions             = true -> null
      - has_downloads               = true -> null
      - has_issues                  = true -> null
      - has_projects                = true -> null
      - homepage_url                = "https://mercari.github.io/tfaction/docs/" -> null
        id                          = "tfaction"
      ~ name                        = "tfaction" -> "action"
      - vulnerability_alerts        = true -> null
        # (24 unchanged attributes hidden)

      - pages {
          - build_type = "legacy" -> null
          - custom_404 = false -> null
          - html_url   = "https://mercari.github.io/tfaction/" -> null
          - status     = "built" -> null
          - url        = "https://api.github.com/repos/mercari/tfaction/pages" -> null

          - source {
              - branch = "gh-pages" -> null
              - path   = "/" -> null
            }
        }

        # (1 unchanged block hidden)
    }

  # github_repository.tfnotify will be updated in-place
  # (imported from "tfnotify")
  ~ resource "github_repository" "tfnotify" {
        allow_auto_merge            = true
        allow_merge_commit          = true
        allow_rebase_merge          = true
        allow_squash_merge          = true
        allow_update_branch         = true
        archived                    = false
        auto_init                   = false
        default_branch              = "main"
        delete_branch_on_merge      = true
      ~ description                 = "Fork of mercari/tfnotify. tfnotify enhances tfnotify in many ways, including Terraform >= v0.15 support and advanced formatting options" -> "Fork of mercari/tfnotify. tfnotify enhances tfnotify in many ways, including Terraform >= v0.15 support and advanced formatting"
        etag                        = "W/\"21702d2afca222d0defb0331feed3ed1fabd9c2626844a280a27c4375c4b0903\""
        full_name                   = "mercari/tfnotify"
        git_clone_url               = "git://github.com/mercari/tfnotify.git"
        has_discussions             = true
        has_downloads               = true
        has_issues                  = true
        has_projects                = false
        has_wiki                    = false
        homepage_url                = "https://mercari.github.io/tfnotify/"
        html_url                    = "https://github.com/mercari/tfnotify"
        http_clone_url              = "https://github.com/mercari/tfnotify.git"
        id                          = "tfnotify"
        is_template                 = false
        merge_commit_message        = "PR_TITLE"
        merge_commit_title          = "MERGE_MESSAGE"
        name                        = "tfnotify"
        node_id                     = "MDEwOlJlcG9zaXRvcnkzMjYzNDcyNzc="
        primary_language            = "Go"
        private                     = false
        repo_id                     = 326347277
        squash_merge_commit_message = "COMMIT_MESSAGES"
        squash_merge_commit_title   = "COMMIT_OR_PR_TITLE"
        ssh_clone_url               = "git@github.com:mercari/tfnotify.git"
        svn_url                     = "https://github.com/mercari/tfnotify"
        topics                      = []
        visibility                  = "public"
        vulnerability_alerts        = false

        pages {
            build_type = "legacy"
            custom_404 = false
            html_url   = "https://mercari.github.io/tfnotify/"
            status     = "built"
            url        = "https://api.github.com/repos/mercari/tfnotify/pages"

            source {
                branch = "gh-pages"
                path   = "/docs"
            }
        }

        security_and_analysis {
            secret_scanning {
                status = "disabled"
            }
            secret_scanning_push_protection {
                status = "disabled"
            }
        }
    }

  # null_resource.foo has moved to null_resource.bar
    resource "null_resource" "bar" {
        id = "7822522400686116714"
    }

  # null_resource.zoo will be created
  + resource "null_resource" "zoo" {
      + id = (known after apply)
    }

Plan: 1 to import, 2 to add, 2 to change, 1 to destroy.

─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

Note: You didn't use the -out option to save this plan, so Terraform can't guarantee to take exactly these actions if you run
"terraform apply" now.
`

const applySuccessResult = `
data.terraform_remote_state.teams_platform_development: Refreshing state...
google_project.my_service: Refreshing state...
google_storage_bucket.chartmuseum: Refreshing state...
google_storage_bucket.ark_tfnotify_prod: Refreshing state...
google_bigquery_dataset.gateway_access_log: Refreshing state...
google_compute_global_address.chartmuseum_tfnotifyapps_com: Refreshing state...
google_compute_global_address.reviews_web_tfnotify_in: Refreshing state...
google_compute_global_address.reviews_api_tfnotify_in: Refreshing state...
google_compute_global_address.teams_web_tfnotify_in: Refreshing state...
google_project_services.my_service: Refreshing state...
google_logging_project_sink.gateway_access_log_bigquery_sink: Refreshing state...
google_project_iam_member.gateway_access_log_bigquery_sink_writer_is_bigquery_data_editor: Refreshing state...
aws_s3_bucket.teams_terraform_private_modules: Refreshing state...
aws_iam_role.datadog_aws_integration: Refreshing state...
aws_s3_bucket.terraform_backend: Refreshing state...
aws_iam_user.teams_terraform: Refreshing state...
aws_iam_policy.datadog_aws_integration: Refreshing state...
aws_iam_user_policy.teams_terraform: Refreshing state...
aws_iam_role_policy_attachment.datadog_aws_integration: Refreshing state...
google_dns_managed_zone.tfnotifyapps_com: Refreshing state...
google_dns_record_set.dev_tfnotifyapps_com: Refreshing state...

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.
`

const applyFailureResult = `
data.terraform_remote_state.teams_platform_development: Refreshing state...
google_project.tfnotify_jp_tfnotify_prod: Refreshing state...
google_project_services.tfnotify_jp_tfnotify_prod: Refreshing state...
google_bigquery_dataset.gateway_access_log: Refreshing state...
google_compute_global_address.reviews_web_tfnotify_in: Refreshing state...
google_compute_global_address.chartmuseum_tfnotifyapps_com: Refreshing state...
google_storage_bucket.chartmuseum: Refreshing state...
google_storage_bucket.ark_tfnotify_prod: Refreshing state...
google_compute_global_address.reviews_api_tfnotify_in: Refreshing state...
google_logging_project_sink.gateway_access_log_bigquery_sink: Refreshing state...
google_project_iam_member.gateway_access_log_bigquery_sink_writer_is_bigquery_data_editor: Refreshing state...
aws_s3_bucket.terraform_backend: Refreshing state...
aws_s3_bucket.teams_terraform_private_modules: Refreshing state...
aws_iam_policy.datadog_aws_integration: Refreshing state...
aws_iam_role.datadog_aws_integration: Refreshing state...
aws_iam_user.teams_terraform: Refreshing state...
aws_iam_user_policy.teams_terraform: Refreshing state...
aws_iam_role_policy_attachment.datadog_aws_integration: Refreshing state...
google_dns_managed_zone.tfnotifyapps_com: Refreshing state...
google_dns_record_set.dev_tfnotifyapps_com: Refreshing state...


Error: Batch "project/tfnotify-jp-tfnotify-prod/services:batchEnable" for request "Enable Project Services tfnotify-jp-tfnotify-prod: map[logging.googleapis.com:{}]" returned error: failed to send enable services request: googleapi: Error 403: The caller does not have permission, forbidden

  on .terraform/modules/tfnotify-jp-tfnotify-prod/google_project_service.tf line 6, in resource "google_project_service" "gcp_api_service":
   6: resource "google_project_service" "gcp_api_service" {


`

func TestPlanParserParse(t *testing.T) { //nolint:maintidx
	t.Parallel()
	testCases := []struct {
		name   string
		body   string
		result ParseResult
	}{
		{
			name: "plan ok pattern",
			body: planSuccessResult,
			result: ParseResult{
				Result:             "Plan: 1 to add, 0 to change, 0 to destroy.",
				HasAddOrUpdateOnly: true,
				HasDestroy:         false,
				HasNoChanges:       false,
				HasError:           false,
				Error:              nil,
				ChangedResult: `
  + google_compute_global_address.my_another_project
      id:         <computed>
      address:    <computed>
      ip_version: "IPV4"
      name:       "my-another-project"
      project:    "my-project"
      self_link:  <computed>


Plan: 1 to add, 0 to change, 0 to destroy.`,
			},
		},
		{
			name: "plan output changes only pattern 0.12",
			body: planOnlyOutputChangesSuccessResult0_12,
			result: ParseResult{
				Result:             "Plan: 0 to add, 0 to change, 0 to destroy.",
				HasAddOrUpdateOnly: false,
				HasDestroy:         false,
				HasNoChanges:       true,
				HasError:           false,
				Error:              nil,
				ChangedResult: `
Plan: 0 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + aws_instance_name = "my-instance"`,
			},
		},
		{
			name: "plan output changes only pattern 0.15",
			body: planOnlyOutputChangesSuccessResult0_15,
			result: ParseResult{
				Result:             "Only Outputs will be changed.",
				HasAddOrUpdateOnly: true,
				HasDestroy:         false,
				HasNoChanges:       false,
				HasError:           false,
				Error:              nil,
				ChangedResult: `Changes to Outputs:
  + test = 42

You can apply this plan to save these new output values to the Terraform
state, without changing any real infrastructure.`,
			},
		},
		{
			name: "plan output changes only pattern with TF_IN_AUTOMATION",
			body: planOnlyOutputChangesSuccessInAutomationResult,
			result: ParseResult{
				Result:             "Only Outputs will be changed.",
				HasAddOrUpdateOnly: true,
				HasDestroy:         false,
				HasNoChanges:       false,
				HasError:           false,
				Error:              nil,
				ChangedResult: `Changes to Outputs:
  + test = 42

You can apply this plan to save these new output values to the Terraform
state, without changing any real infrastructure.`,
			},
		},
		{
			name: "no stdin",
			body: "",
			result: ParseResult{
				Result:             "",
				HasAddOrUpdateOnly: false,
				HasDestroy:         false,
				HasNoChanges:       false,
				HasError:           false,
				HasParseError:      true,
				Error:              errors.New("cannot parse plan result"),
			},
		},
		{
			name: "plan ng pattern",
			body: planFailureResult,
			result: ParseResult{
				Result: `Error: Error refreshing state: 4 error(s) occurred:

* google_sql_database.main: 1 error(s) occurred:

* google_sql_database.main: google_sql_database.main: Error reading SQL Database "main" in instance "main-master-instance": googleapi: Error 409: The instance or operation is not in an appropriate state to handle the request., invalidState
* google_sql_user.proxyuser_main: 1 error(s) occurred:`,
				HasAddOrUpdateOnly: false,
				HasDestroy:         false,
				HasNoChanges:       false,
				HasError:           true,
				Error:              nil,
			},
		},
		{
			name: "plan no changes",
			body: planNoChanges,
			result: ParseResult{
				Result:             "No changes. Infrastructure is up-to-date.",
				HasAddOrUpdateOnly: false,
				HasDestroy:         false,
				HasNoChanges:       true,
				HasError:           false,
				Error:              nil,
			},
		},
		{
			name: "plan has destroy",
			body: planHasDestroy,
			result: ParseResult{
				Result:             "Plan: 0 to add, 0 to change, 1 to destroy.",
				HasAddOrUpdateOnly: false,
				HasDestroy:         true,
				HasNoChanges:       false,
				HasError:           false,
				Error:              nil,
				ChangedResult: `
  - google_project_iam_member.team_platform[2]


Plan: 0 to add, 0 to change, 1 to destroy.`,
			},
		},
		{
			name: "plan has add and destroy",
			body: planHasAddAndDestroy,
			result: ParseResult{
				Result:             "Plan: 1 to add, 0 to change, 1 to destroy.",
				HasAddOrUpdateOnly: false,
				HasDestroy:         true,
				HasNoChanges:       false,
				HasError:           false,
				Error:              nil,
				ChangedResult: `
  + google_compute_global_address.my_another_project
      id:         <computed>
      address:    <computed>
      ip_version: "IPV4"
      name:       "my-another-project"
      project:    "my-project"
      self_link:  <computed>

  - google_project_iam_member.team_platform[2]

Plan: 1 to add, 0 to change, 1 to destroy.`,
			},
		},
		{
			name: "plan has add and update in place",
			body: planHasAddAndUpdateInPlace,
			result: ParseResult{
				Result:             "Plan: 1 to add, 1 to change, 0 to destroy.",
				HasAddOrUpdateOnly: true,
				ChangedResult: `
  + google_compute_global_address.my_another_project
      id:         <computed>
      address:    <computed>
      ip_version: "IPV4"
      name:       "my-another-project"
      project:    "my-project"
      self_link:  <computed>

  ~ google_project_iam_member.team_platform[2]

Plan: 1 to add, 1 to change, 0 to destroy.`,
			},
		},
		{
			name: "imported and moved resources are changed",
			body: planImportedMovedResourceChanged,
			result: ParseResult{
				Result:             "Plan: 1 to import, 2 to add, 2 to change, 1 to destroy.",
				HasAddOrUpdateOnly: false,
				HasDestroy:         true,
				HasNoChanges:       false,
				HasError:           false,
				Error:              nil,
				CreatedResources: []string{
					"null_resource.zoo",
				},
				UpdatedResources: []string{
					"github_repository.tfaction-2",
					"github_repository.tfnotify",
				},
				ReplacedResources: []string{
					"github_issue.test-2",
				},
				MovedResources: []*MovedResource{
					{
						Before: "github_issue.test",
						After:  "github_issue.test-2",
					},
					{
						Before: "github_repository.tfaction",
						After:  "github_repository.tfaction-2",
					},
					{
						Before: "null_resource.foo",
						After:  "null_resource.bar",
					},
				},
				ImportedResources: []string{
					"github_repository.tfnotify",
				},
				ChangedResult: `
  # github_issue.test-2 must be replaced
  # (moved from github_issue.test)
-/+ resource "github_issue" "test-2" {
      - assignees        = [] -> null
      ~ etag             = "W/\"e14116be2014dddd5d766ba8d69e59524491648d626917f4dbe8d5422ef32ba3\"" -> (known after apply)
      ~ id               = "tfaction:902" -> (known after apply)
      ~ issue_id         = 1685922946 -> (known after apply)
      - labels           = [
          - "enhancement",
        ] -> null
      - milestone_number = 0 -> null
      ~ number           = 902 -> (known after apply)
      ~ repository       = "tfaction" -> "tfnotify" # forces replacement
        # (1 unchanged attribute hidden)
    }

  # github_repository.tfaction-2 will be updated in-place
  # (moved from github_repository.tfaction)
  ~ resource "github_repository" "tfaction-2" {
      ~ allow_auto_merge            = true -> false
      - allow_update_branch         = true -> null
      ~ delete_branch_on_merge      = true -> false
      - description                 = "Framework for Monorepo to build high level Terraform Workflows by GitHub Actions" -> null
      ~ full_name                   = "mercari/tfaction" -> (known after apply)
      - has_discussions             = true -> null
      - has_downloads               = true -> null
      - has_issues                  = true -> null
      - has_projects                = true -> null
      - homepage_url                = "https://mercari.github.io/tfaction/docs/" -> null
        id                          = "tfaction"
      ~ name                        = "tfaction" -> "action"
      - vulnerability_alerts        = true -> null
        # (24 unchanged attributes hidden)

      - pages {
          - build_type = "legacy" -> null
          - custom_404 = false -> null
          - html_url   = "https://mercari.github.io/tfaction/" -> null
          - status     = "built" -> null
          - url        = "https://api.github.com/repos/mercari/tfaction/pages" -> null

          - source {
              - branch = "gh-pages" -> null
              - path   = "/" -> null
            }
        }

        # (1 unchanged block hidden)
    }

  # github_repository.tfnotify will be updated in-place
  # (imported from "tfnotify")
  ~ resource "github_repository" "tfnotify" {
        allow_auto_merge            = true
        allow_merge_commit          = true
        allow_rebase_merge          = true
        allow_squash_merge          = true
        allow_update_branch         = true
        archived                    = false
        auto_init                   = false
        default_branch              = "main"
        delete_branch_on_merge      = true
      ~ description                 = "Fork of mercari/tfnotify. tfnotify enhances tfnotify in many ways, including Terraform >= v0.15 support and advanced formatting options" -> "Fork of mercari/tfnotify. tfnotify enhances tfnotify in many ways, including Terraform >= v0.15 support and advanced formatting"
        etag                        = "W/\"21702d2afca222d0defb0331feed3ed1fabd9c2626844a280a27c4375c4b0903\""
        full_name                   = "mercari/tfnotify"
        git_clone_url               = "git://github.com/mercari/tfnotify.git"
        has_discussions             = true
        has_downloads               = true
        has_issues                  = true
        has_projects                = false
        has_wiki                    = false
        homepage_url                = "https://mercari.github.io/tfnotify/"
        html_url                    = "https://github.com/mercari/tfnotify"
        http_clone_url              = "https://github.com/mercari/tfnotify.git"
        id                          = "tfnotify"
        is_template                 = false
        merge_commit_message        = "PR_TITLE"
        merge_commit_title          = "MERGE_MESSAGE"
        name                        = "tfnotify"
        node_id                     = "MDEwOlJlcG9zaXRvcnkzMjYzNDcyNzc="
        primary_language            = "Go"
        private                     = false
        repo_id                     = 326347277
        squash_merge_commit_message = "COMMIT_MESSAGES"
        squash_merge_commit_title   = "COMMIT_OR_PR_TITLE"
        ssh_clone_url               = "git@github.com:mercari/tfnotify.git"
        svn_url                     = "https://github.com/mercari/tfnotify"
        topics                      = []
        visibility                  = "public"
        vulnerability_alerts        = false

        pages {
            build_type = "legacy"
            custom_404 = false
            html_url   = "https://mercari.github.io/tfnotify/"
            status     = "built"
            url        = "https://api.github.com/repos/mercari/tfnotify/pages"

            source {
                branch = "gh-pages"
                path   = "/docs"
            }
        }

        security_and_analysis {
            secret_scanning {
                status = "disabled"
            }
            secret_scanning_push_protection {
                status = "disabled"
            }
        }
    }

  # null_resource.foo has moved to null_resource.bar
    resource "null_resource" "bar" {
        id = "7822522400686116714"
    }

  # null_resource.zoo will be created
  + resource "null_resource" "zoo" {
      + id = (known after apply)
    }

Plan: 1 to import, 2 to add, 2 to change, 1 to destroy.`,
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := NewPlanParser().Parse(testCase.body)
			if diff := cmp.Diff(result, testCase.result, cmpopts.IgnoreUnexported(ParseResult{}), cmpopts.IgnoreFields(ParseResult{}, "Error")); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestApplyParserParse(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		body   string
		result ParseResult
	}{
		{
			name: "no stdin",
			body: "",
			result: ParseResult{
				Result:        "",
				HasParseError: true,
				Error:         errors.New("cannot parse apply result"),
			},
		},
		{
			name: "apply ok pattern",
			body: applySuccessResult,
			result: ParseResult{
				Result: "Apply complete! Resources: 0 added, 0 changed, 0 destroyed.",
				Error:  nil,
			},
		},
		{
			name: "apply ng pattern",
			body: applyFailureResult,
			result: ParseResult{
				Result: `Error: Batch "project/tfnotify-jp-tfnotify-prod/services:batchEnable" for request "Enable Project Services tfnotify-jp-tfnotify-prod: map[logging.googleapis.com:{}]" returned error: failed to send enable services request: googleapi: Error 403: The caller does not have permission, forbidden

  on .terraform/modules/tfnotify-jp-tfnotify-prod/google_project_service.tf line 6, in resource "google_project_service" "gcp_api_service":
   6: resource "google_project_service" "gcp_api_service" {`,
				Error:    nil,
				HasError: true,
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := NewApplyParser().Parse(testCase.body)
			if diff := cmp.Diff(result, testCase.result, cmpopts.IgnoreUnexported(ParseResult{}), cmpopts.IgnoreFields(ParseResult{}, "Error")); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestTrimLastNewline(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		data     []string
		expected []string
	}{
		{
			data:     []string{},
			expected: []string{},
		},
		{
			data:     []string{"a", "b", "c", ""},
			expected: []string{"a", "b", "c"},
		},
		{
			data:     []string{"a", ""},
			expected: []string{"a"},
		},
		{
			data:     []string{""},
			expected: []string{},
		},
		{
			data:     []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			data:     []string{"a"},
			expected: []string{"a"},
		},
	}
	for _, testCase := range testCases {
		actual := trimLastNewline(testCase.data)
		if diff := cmp.Diff(actual, testCase.expected); diff != "" {
			t.Error(diff)
		}
	}
}
