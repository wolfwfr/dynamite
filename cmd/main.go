package main

import (
	"context"
	"fmt"
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
	region_key      = "region"

	corrupt_config_dir = "<config_dir_not_found>"
)

var configDir string

func init() {
	// Local user configuration.
	var err error
	configDir, _ = os.UserConfigDir()
	if err != nil {
		configDir = corrupt_config_dir
	}
}

func main() {
	cmd := &cli.Command{
		Name:        "Dynamite",
		Description: "TUI for Amazon DynamoDB queries",
		Usage:       "Amazon DynamoDB query engine",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    aws_profile_key,
				Aliases: []string{"p"},
				Value:   "",
				Usage:   "aws-profile",
			},
			&cli.StringFlag{
				Name:    region_key,
				Aliases: []string{"r"},
				Value:   "",
				Usage:   "aws-region (takes precedence over default or region in config-file, when set)",
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
				Usage:   "override the dynamodb host URL, useful for connecting to a local dynamodb compatible API (e.g. 'http://localhost:8000')",
			},
		},
		Action: runApplication,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func runApplication(ctx context.Context, cmd *cli.Command) error {
	var uiopts []ui.Option

	cfgf, _, err := loadConfig(cmd.String(config_key))
	if err != nil {
		uiopts = append(uiopts, ui.WithInitialErrorNotification(err))
	}

	urlS := cmd.String(dynamo_url_key)
	var urlP *string
	if urlS != "" {
		urlP = &urlS
	}

	// set up credentials channel for the aws-client mfa-token-provider
	credsC := make(chan appconfig.CredentialsResponse, 1)
	var p *tea.Program
	f := func() (string, error) {
		p.Send(appconfig.CredentialsRequest{})
		resp := <-credsC
		return resp.Token, resp.Error
	}

	cfg := appconfig.Config{
		Profile:          resolveProfile(cmd, cfgf),
		Region:           resolveRegion(cmd, cfgf),
		URL:              urlP,
		AvailableRegions: cfgf.AWSRegions,
		StarredRegions:   cfgf.StarredRegions,
		MaxTables:        cfgf.MaxTables,

		MFACredentialCB: f,
		MFACredentialC:  credsC,
	}

	p = tea.NewProgram(ui.NewModel(ctx, cfg, uiopts...))
	_, err = p.Run()
	return err
}

func loadConfig(path string) (configfile.ConfigFile, *configfile.ConfigManager, error) {
	full, err1 := filepath.Abs(path)
	if err1 != nil {
		err1 = fmt.Errorf("failed to construct a valid config-path: %w", err1)
	}

	configman := configfile.NewConfigManager(full)
	cfgf, err2 := configman.LoadConfig(true)
	if err1 != nil {
		return cfgf, configman, err1
	}
	if err2 != nil {
		return cfgf, configman, fmt.Errorf("failed to load local config: %w", err2)
	}

	return cfgf, configman, nil
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

func resolveRegion(cmd *cli.Command, cfg configfile.ConfigFile) string {
	if r := cmd.String(region_key); r != "" {
		return r
	}
	if r := os.Getenv("AWS_REGION"); r != "" {
		return r
	}
	if r := cfg.DefaultRegion; r != "" {
		return r
	}
	return "us-east-1"
}
