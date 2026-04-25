package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/urfave/cli/v3"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/configfile"
	"github.com/wolfwfr/dynamite/pkg/ui"
)

const (
	aws_profile_key = "aws_profile"
	config_key      = "config"
	dynamo_url_key  = "dynamo_url"

	corrupt_config_dir = "<config_dir_not_found>"
)

var configDir string

func initDirs() error {
	// Local user configuration.
	var err error
	configDir, err = os.UserConfigDir()
	if err != nil {
		configDir = corrupt_config_dir
		return err
	}
	return nil
}

func main() {
	err := initDirs()
	if err != nil {
		// TODO: notify user
	}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    aws_profile_key,
				Aliases: []string{"p"},
				Value:   "",
				Usage:   "aws-profile",
			},
			&cli.StringFlag{
				Name:    config_key,
				Aliases: []string{"c"},
				Value:   filepath.Join(configDir, "dynamite/config.yaml"),
				Usage:   "path to config file (relative or absolute); must be yaml",
			},
			&cli.StringFlag{
				Name:    dynamo_url_key,
				Aliases: []string{"u"},
				Value:   "",
				Usage:   "override the dynamodb host URL, useful for connecting to a local dynamodb compatible API",
			},
		},
		Action: runApplication,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func runApplication(ctx context.Context, cmd *cli.Command) error {
	full, err := filepath.Abs(cmd.String(config_key))
	if err != nil {
		// TODO: handling
	}

	configman := configfile.NewConfigManager(full)
	cfgf, err := configman.LoadConfig(true)
	if err != nil {
		// TODO: handling
	}

	cfg := appconfig.Config{
		Profile: resolveProfile(cmd, cfgf),
		Region:  cfgf.DefaultRegion,
	}

	p := tea.NewProgram(ui.NewModel(ctx, cfg))
	_, err = p.Run()
	return err
}

func resolveProfile(cmd *cli.Command, cfg configfile.ConfigFile) *string {
	if pr := cmd.String(aws_profile_key); pr != "" {
		return &pr
	}
	if pr := os.Getenv("AWS_PROFILE"); pr != "" {
		return &pr
	}
	if pr := cfg.DefaultProfile; pr != "" {
		return &pr
	}
	return nil
}
