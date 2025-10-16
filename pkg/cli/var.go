package cli

import (
	"errors"
	"os"
	"strings"

	"github.com/mercari/tfnotify/v1/pkg/config"
	"github.com/mercari/tfnotify/v1/pkg/mask"
	"github.com/urfave/cli/v3"
)

func parseVars(vars []string, envs []string, varsM map[string]string) error {
	parseVarEnvs(envs, varsM)
	return parseVarOpts(vars, varsM)
}

func parseVarOpts(vars []string, varsM map[string]string) error {
	for _, v := range vars {
		a := strings.Index(v, ":")
		if a == -1 {
			return errors.New("the value of var option is invalid. the format should be '<name>:<value>': " + v)
		}
		varsM[v[:a]] = v[a+1:]
	}
	return nil
}

func parseVarEnvs(envs []string, m map[string]string) {
	for _, kv := range envs {
		k, v, _ := strings.Cut(kv, "=")
		if a := strings.TrimPrefix(k, "TFCMT_VAR_"); k != a {
			m[a] = v
		}
	}
}

func parseOpts(cmd *cli.Command, cfg *config.Config, envs []string) error { //nolint:cyclop
	if owner := cmd.String("owner"); owner != "" {
		cfg.CI.Owner = owner
	}

	if repo := cmd.String("repo"); repo != "" {
		cfg.CI.Repo = repo
	}

	if sha := cmd.String("sha"); sha != "" {
		cfg.CI.SHA = sha
	}

	if pr := cmd.Int("pr"); pr != 0 {
		cfg.CI.PRNumber = pr
	}

	if cmd.IsSet("patch") {
		cfg.PlanPatch = cmd.Bool("patch")
	}

	if buildURL := cmd.String("build-url"); buildURL != "" {
		cfg.CI.Link = buildURL
	}

	if output := cmd.String("output"); output != "" {
		cfg.Output = output
	}

	if cmd.IsSet("skip-no-changes") {
		cfg.Terraform.Plan.WhenNoChanges.DisableComment = cmd.Bool("skip-no-changes")
	}

	if cmd.IsSet("ignore-warning") {
		cfg.Terraform.Plan.IgnoreWarning = cmd.Bool("ignore-warning")
	}

	vars := cmd.StringSlice("var")
	vm := make(map[string]string, len(vars))
	if err := parseVars(vars, envs, vm); err != nil {
		return err
	}
	cfg.Vars = vm

	masks, err := mask.ParseMasksFromEnv()
	if err != nil {
		return err
	}
	cfg.Masks = masks

	if cmd.IsSet("disable-label") {
		cfg.Terraform.Plan.DisableLabel = cmd.Bool("disable-label")
	}

	if cfg.GHEBaseURL == "" {
		cfg.GHEBaseURL = os.Getenv("GITHUB_API_URL")
	}
	if cfg.GHEGraphQLEndpoint == "" {
		cfg.GHEGraphQLEndpoint = os.Getenv("GITHUB_GRAPHQL_URL")
	}

	return nil
}
