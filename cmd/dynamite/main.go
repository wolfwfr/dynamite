package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/urfave/cli/v3"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/configfile"
	"github.com/wolfwfr/dynamite/pkg/ui"
)

const (
	aws_profile_key = "profile"
	config_key      = "cfg"
	dynamo_url_key  = "url"
	region_key      = "region"
	table_key       = "table"
	index_key       = "index"
	hash_val_key    = "hash_key_value"
	range_val_1_key = "range_key_value"
	range_val_2_key = "range_key_value_2"
	range_op_key    = "range_operator"
	range_order_key = "range_order_descending"

	corrupt_config_dir = "<config_dir_not_found>"

	env_config_dir  = "DYNAMITE_TUI_CONFIG_DIR"
	env_aws_profile = "AWS_PROFILE"
	env_aws_region  = "AWS_REGION"
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
				Sources: cli.EnvVars(env_aws_profile),
				Aliases: []string{"p"},
				Value:   "",
				Usage:   "aws-profile",
			},
			&cli.StringFlag{
				Name:    region_key,
				Sources: cli.EnvVars(env_aws_region),
				Aliases: []string{"r"},
				Value:   "",
				Usage:   "aws-region (takes precedence over default or region in config-file, when set)",
			},
			&cli.StringFlag{
				Name:    config_key,
				Sources: cli.EnvVars(env_config_dir),
				Aliases: []string{"c"},
				Value:   filepath.Join(configDir, "dynamite/"),
				Usage:   "path to directory hosting 'config.yaml' (relative or absolute)",
			},
			&cli.StringFlag{
				Name:    dynamo_url_key,
				Aliases: []string{"u"},
				Value:   "",
				Usage:   "override the dynamodb host URL, useful for connecting to a local dynamodb compatible API (e.g. 'http://localhost:8000')",
			},
			&cli.StringFlag{
				Name:    table_key,
				Aliases: []string{},
				Value:   "",
				Usage:   "select the specified table, circumvents the table-view",
			},
			&cli.StringFlag{
				Name:    index_key,
				Aliases: []string{},
				Value:   "",
				Usage:   fmt.Sprintf("specify a table-index, only takes effect when '%s' is specified", table_key),
			},
			&cli.StringFlag{
				Name:    hash_val_key,
				Aliases: []string{"hk"},
				Value:   "",
				Usage:   fmt.Sprintf("specify a hash-key-value, only takes effect when '%s' is specified", table_key),
			},
			&cli.StringFlag{
				Name:    range_val_1_key,
				Aliases: []string{"rk"},
				Value:   "",
				Usage:   fmt.Sprintf("specify a range-key-value, only takes effect when '%s' is specified", table_key),
			},
			&cli.StringFlag{
				Name:    range_val_2_key,
				Aliases: []string{"rk2"},
				Value:   "",
				Usage:   fmt.Sprintf("specify a second range-key-value, only takes effect when '%s' is specified, and the 'BETWEEN' operator is used", table_key),
			},
			&cli.StringFlag{
				Name:      range_op_key,
				Aliases:   []string{"op"},
				Value:     "equals",
				Usage:     fmt.Sprintf("specify the range-key operator, only takes effect when '%s' is specified", table_key),
				Validator: validateRangeOp(),
			},
			&cli.BoolFlag{
				Name:    range_order_key,
				Aliases: []string{"descending", "dsc"},
				Value:   false,
				Usage:   fmt.Sprintf("specify the range order, defaults to true ⇒ 'ascending', only takes effect when '%s' is specified", table_key),
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

	rk1, rk2 := resolveQueryRangeKeyValues(cmd)

	cfg := appconfig.Config{
		Profile:          resolveProfile(cmd, cfgf),
		Region:           resolveRegion(cmd, cfgf),
		URL:              urlP,
		AvailableRegions: cfgf.AWSRegions,
		StarredRegions:   cfgf.StarredRegions,
		Tables: appconfig.Tables{
			MaxTables:    cfgf.TablesMax,
			Pagesize:     cfgf.TablesPageSize,
			PrimaryWidth: cfgf.TablesPrimaryWidth,
		},
		Items: appconfig.Items{
			PrimaryWidth: cfgf.ItemsPrimaryWidth,
			PageSize:     cfgf.ItemsPageSize,
		},
		MFACredentialCB: f,
		MFACredentialC:  credsC,

		Initialisation: appconfig.Initialisation{
			Table: cmd.String(table_key),
			Index: cmd.String(index_key),
			Query: appconfig.Queryinitialisation{
				HashkeyValue:     resolveQueryHashKey(cmd),
				RangekeyValue1:   rk1,
				RangekeyValue2:   rk2,
				RangeKeyOperator: resolveQueryRangeOp(cmd),
				RangeDescending:  resolveQueryRangeOrder(cmd),
			},
		},
	}

	p = tea.NewProgram(ui.NewModel(ctx, cfg, uiopts...))
	_, err = p.Run()
	return err
}

func loadConfig(path string) (configfile.Config, *configfile.ConfigManager, error) {
	path = filepath.Join(path, "config.yaml")
	full, err1 := filepath.Abs(path)
	if err1 != nil {
		err1 = fmt.Errorf("failed to construct a valid config-path: %w", err1)
	}

	configman := configfile.NewConfigManager(full)
	cfgf, err2 := configman.LoadConfig()
	if err1 != nil {
		return cfgf, configman, err1
	}
	if err2 != nil {
		return cfgf, configman, fmt.Errorf("failed to load local config: %w", err2)
	}

	return cfgf, configman, nil
}

func resolveQueryHashKey(cmd *cli.Command) string {
	return cmd.String(hash_val_key)
}

func resolveQueryRangeKeyValues(cmd *cli.Command) (v1, v2 *string) {
	flags := cmd.FlagNames()
	if slices.Contains(flags, range_val_1_key) {
		vv1 := cmd.String(range_val_1_key)
		v1 = &vv1
	}
	if slices.Contains(flags, range_val_2_key) {
		vv1 := cmd.String(range_val_2_key)
		v1 = &vv1
	}
	return
}

// TODO: parsing & validation
func resolveQueryRangeOp(cmd *cli.Command) string {
	return strings.ToLower(cmd.String(range_op_key))
}

func validateRangeOp() func(s string) error {
	ss := []string{
		"equals",
		"greater than or equal",
		"greater than",
		"less than or equal",
		"less than",
		"between",
		"begins with",
	}
	return func(s string) error {
		if slices.Contains(ss, strings.ToLower(s)) {
			return nil
		}
		return fmt.Errorf("operator '%s', not supported. Must be one of %q.", s, ss)
	}
}

func resolveQueryRangeOrder(cmd *cli.Command) bool {
	return cmd.Bool(range_order_key)
}

func resolveProfile(cmd *cli.Command, cfg configfile.Config) *string {
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

func resolveRegion(cmd *cli.Command, cfg configfile.Config) string {
	if r := cmd.String(region_key); r != "" {
		return r
	}
	if r := cfg.DefaultRegion; r != "" {
		return r
	}
	return "us-east-1"
}
