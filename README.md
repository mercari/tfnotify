tfnotify
========

[![][release-svg]][release] [![][test-svg]][test] [![][goreportcard-svg]][goreportcard]

[release]: https://github.com/mercari/tfnotify/actions?query=workflow%3Arelease
[release-svg]: https://github.com/mercari/tfnotify/workflows/release/badge.svg
[test]: https://github.com/mercari/tfnotify/actions?query=workflow%3Atest
[test-svg]: https://github.com/mercari/tfnotify/workflows/test/badge.svg
[goreportcard]: https://goreportcard.com/report/github.com/mercari/tfnotify
[goreportcard-svg]: https://goreportcard.com/badge/github.com/mercari/tfnotify

Tfnotify is a Go-based template engine that creates comments on Terraform operations. It pipes commands directly and captures operation outputs, including exit codes and error details and offers to post them as a Github Comment, with AI summaries and analysis based on custom templates.

## Motivation

There are commands such as `plan` and `apply` on Terraform command, but many developers think they would like to check if the execution of those commands succeeded.
Terraform commands are often executed via CI like Cloudbuild CI/Github Actions, but in that case you need to go to the CI page to check it.
This is very troublesome. It is very efficient if you can check it with the supported CI.

### Key Features:

- **Flexible Templating**: Go template system with custom functions for formatting notifications
- **Intelligent Parsing**: Regex-based engine that extracts meaningful information from Terraform plan and apply outputs.
- **Multi-Platform Support**: Integrates with GitHub, CircleCI, Cloudbuild.
- **CI/CD Awareness**: Automatically detects and integrates with 9+ CI/CD platforms
- **Label Management**: Automatically creates and manages GitHub labels based on plan results
- **AI-Powered Summaries**: Generates optional AI summaries of infrastructure changes using multiple providers
- **Security Features**: Built-in masking system prevents sensitive data exposure
- **Comment Management**: Updates existing comments instead of creating duplicates

## Installation

Grab the binary from GitHub Releases (Recommended)

or

```console
$ go get -u github.com/mercari/tfnotify
```

### Supported Platforms

Tfnotify supports the following platforms:

- CI
    - CircleCI
    - CodeBuild
    - CloudBuild
    - GitHub Actions
- Notifier
    - GitHub


### Basic

tfnotify is just CLI command. So you can run it from your local after grabbing the binary.

Basically tfnotify waits for the input from Stdin. So tfnotify needs to pipe the output of Terraform command like the following:

```console
$ tfnotify plan
```

For `plan` command, you also need to specify `plan` as the argument of tfnotify. In the case of `apply`, you need to do `apply`. Currently supported commands can be checked with `tfnotify --help`.

### Configurations

When running tfnotify, you can specify the configuration path via `--config` option (if it's omitted, it defaults to `{.,}tfnotify.y{,a}ml`).

The example settings of GitHub and GitHub Enterprise. Incidentally, there is no need to replace TOKEN string such as `$GITHUB_TOKEN` with the actual token. Instead, it must be defined as environment variables in CI settings.

[template](https://golang.org/pkg/text/template/) of Go can be used for `template`. The templates can be used in `tfnotify.yaml` are as follows:

Placeholder | Usage
---|---
`{{ .Title }}` | Like `## Plan result`
`{{ .Message }}` | A string that can be set from CLI with `--message` option
`{{ .Result }}` | Matched result by parsing like `Plan: 1 to add` or `No changes`
`{{ .Body }}` | The entire of Terraform execution result
`{{ .Link }}` | The link of the build page on CI

On GitHub, tfnotify can also put a warning message if the plan result contains resource deletion (optional).

#### Template Examples

<details>
<summary>For GitHub</summary>

```yaml
---
ci: circleci
notifier:
  github:
    token: $GITHUB_TOKEN
    repository:
      owner: "mercari"
      name: "tfnotify"
terraform:
  fmt:
    template: |
      {{ .Title }}

      {{ .Message }}

      {{ .Result }}

      {{ .Body }}
  plan:
    template: |
      {{ .Title }} <sup>[CI link]( {{ .Link }} )</sup>
      {{ .Message }}
      {{if .Result}}
      <pre><code>{{ .Result }}
      </pre></code>
      {{end}}
      <details><summary>Details (Click me)</summary>

      <pre><code>{{ .Body }}
      </pre></code></details>
  apply:
    template: |
      {{ .Title }}
      {{ .Message }}
      {{if .Result}}
      <pre><code>{{ .Result }}
      </pre></code>
      {{end}}
      <details><summary>Details (Click me)</summary>

      <pre><code>{{ .Body }}
      </pre></code></details>
```

If you would like to let tfnotify warn the resource deletion, add `when_destroy` configuration as below.

```yaml
---
# ...
terraform:
  # ...
  plan:
    template: |
      {{ .Title }} <sup>[CI link]( {{ .Link }} )</sup>
      {{ .Message }}
      {{if .Result}}
      <pre><code>{{ .Result }}
      </pre></code>
      {{end}}
      <details><summary>Details (Click me)</summary>

      <pre><code>{{ .Body }}
      </pre></code></details>
    when_destroy:
      template: |
        ## :warning: WARNING: Resource Deletion will happen :warning:

        This plan contains **resource deletion**. Please check the plan result very carefully!
  # ...
```

You can also let tfnotify add a label to PRs depending on the `terraform plan` output result. Currently, this feature is for Github labels only.

```yaml
---
# ...
terraform:
  # ...
  plan:
    template: |
      {{ .Title }} <sup>[CI link]( {{ .Link }} )</sup>
      {{ .Message }}
      {{if .Result}}
      <pre><code>{{ .Result }}
      </pre></code>
      {{end}}
      <details><summary>Details (Click me)</summary>

      <pre><code>{{ .Body }}
      </pre></code></details>
    when_add_or_update_only:
      label: "add-or-update"
    when_destroy:
      label: "destroy"
    when_no_changes:
      label: "no-changes"
    when_plan_error:
      label: "error"
  # ...
```

Sometimes you may want not to HTML-escape Terraform command outputs.
For example, when you use code block to print command output, it's better to use raw characters instead of character references (e.g. `-/+` -> `-/&#43;`, `"` -> `&#34;`).

You can disable HTML escape by adding `use_raw_output: true` configuration.
With this configuration, Terraform doesn't HTML-escape any Terraform output.

~~~yaml
---
# ...
terraform:
  use_raw_output: true
  # ...
  plan:
    template: |
      {{ .Title }} <sup>[CI link]( {{ .Link }} )</sup>
      {{ .Message }}
      {{if .Result}}
      ```
      {{ .Result }}
      ```
      {{end}}
      <details><summary>Details (Click me)</summary>

      ```
      {{ .Body }}
      ```
  # ...
~~~

</details>

<details>
<summary>For GitHub Enterprise</summary>

```yaml
---
ci: circleci
notifier:
  github:
    token: $GITHUB_TOKEN
    base_url: $GITHUB_BASE_URL # Example: https://github.example.com/api/v3
    repository:
      owner: "mercari"
      name: "tfnotify"
terraform:
  fmt:
    template: |
      {{ .Title }}

      {{ .Message }}

      {{ .Result }}

      {{ .Body }}
  plan:
    template: |
      {{ .Title }} <sup>[CI link]( {{ .Link }} )</sup>
      {{ .Message }}
      {{if .Result}}
      <pre><code>{{ .Result }}
      </pre></code>
      {{end}}
      <details><summary>Details (Click me)</summary>

      <pre><code>{{ .Body }}
      </pre></code></details>
  apply:
    template: |
      {{ .Title }}
      {{ .Message }}
      {{if .Result}}
      <pre><code>{{ .Result }}
      </pre></code>
      {{end}}
      <details><summary>Details (Click me)</summary>

      <pre><code>{{ .Body }}
      </pre></code></details>
```

</details>


### Google Cloud Build Considerations

- These environment variables are needed to be set using [substitutions](https://cloud.google.com/cloud-build/docs/configuring-builds/substitute-variable-values)
  - `COMMIT_SHA`
  - `BUILD_ID`
  - `PROJECT_ID`
  - `_PR_NUMBER`
  - `_REGION`
- Recommended trigger events
  - `terraform plan`: Pull request
  - `terraform apply`: Push to branch

## Contribution

Please read the CLA below carefully before submitting your contribution.

https://www.mercari.com/cla/

## License

Copyright 2018 Mercari, Inc.

Licensed under the MIT License.
