package cli

import (
	"github.com/mercari/tfnotify/v1/pkg/config"
	"github.com/urfave/cli/v3"
)

func newConfig(cmd *cli.Command) (config.Config, error) {
	cfg := config.Config{}
	confPath, err := cfg.Find(cmd.String("config"))
	if err != nil {
		return cfg, err
	}
	if confPath != "" {
		if err := cfg.LoadFile(confPath); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}
