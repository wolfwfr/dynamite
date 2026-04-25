package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/urfave/cli/v3"

	"github.com/wolfwfr/dynamite/pkg/ui"
)

const (
	aws_profile_key = "aws_profile"
	config_key      = "config"
	dynamo_url_key  = "dynamo_url"
)

var configDir string

func init() {
	// Local user configuration.
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		configDir = os.Getenv("XDG_CONFIG_HOME")
	} else {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
}

func main() {
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
				Usage:   "path to config file (relative or absolute)",
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
	// profile := cmd.String(aws_profile_key)
	// configPath := cmd.String(config_key)
	// TODO: init dependencies
	p := tea.NewProgram(ui.NewModel())
	_, err := p.Run()
	return err
}
