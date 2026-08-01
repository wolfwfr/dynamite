package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/urfave/cli/v3"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/configfile"
	"github.com/wolfwfr/dynamite/pkg/logging"
	"github.com/wolfwfr/dynamite/pkg/ui"
)

// CLI flags
const (
	version_key = "version"

	aws_profile_key = "profile"

	config_key = "cfg"

	dynamo_url_key = "url"
	region_key     = "region"

	table_key       = "table"
	index_key       = "index"
	hash_val_key    = "hash_key_value"
	range_val_1_key = "range_key_value"
	range_val_2_key = "range_key_value_2"
	range_op_key    = "range_operator"
	range_order_key = "range_order_descending"

	log_debug_key  = "debug"
	log_key        = "log"
	log_loc_key    = "log_path"
	log_level_key  = "log_level"
	log_text_key   = "log_text"
	log_append_key = "log_append"
)

// categories
const (
	cat_query   = "query/scan"
	cat_logging = "logging"
)

// environment variables
const (
	env_config_dir  = "DYNAMITE_TUI_CONFIG_DIR"
	env_aws_profile = "AWS_PROFILE"
	env_aws_region  = "AWS_REGION"
)

// other
const (
	corrupt_config_dir = "<config_dir_not_found>"
	dynamite_subdir    = "dynamite-tui"
	dynamite_logfile   = "dynamite.log"
)

var (
	configDir string
	logDir    string
	version   = ">v0.0.0"
	commit    = ">b99208e0e5532df6c62e6a2011195084d3eb7a0d"
	buildDate = ">2026-07-31"
)

func init() {
	// Local user configuration.
	var err error
	configDir, _ = os.UserConfigDir()
	if err != nil {
		configDir = corrupt_config_dir
	}

	logDir = configDir
	if logDir != "" {
		logDir = filepath.Join(logDir, dynamite_subdir)
	}
}

func main() {
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("version=%s commit=%s buildDate=%s\n", cmd.Root().Version, commit, buildDate)
	}

	cmd := &cli.Command{
		Name:        "Dynamite",
		Version:     version,
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
				Value:   filepath.Join(configDir, dynamite_subdir),
				Usage:   "path to directory hosting 'config.yaml' (relative or absolute)",
			},
			&cli.StringFlag{
				Name:    dynamo_url_key,
				Aliases: []string{"u"},
				Value:   "",
				Usage:   "override the dynamodb host URL, useful for connecting to a local dynamodb compatible API (e.g. 'http://localhost:8000')",
			},
			&cli.StringFlag{
				Name:     table_key,
				Category: cat_query,
				Aliases:  []string{},
				Value:    "",
				Usage:    "select the specified table, circumvents the table-view",
			},
			&cli.StringFlag{
				Name:     index_key,
				Category: cat_query,
				Aliases:  []string{},
				Value:    "",
				Usage:    fmt.Sprintf("specify a table-index, only takes effect when '%s' is specified", table_key),
			},
			&cli.StringFlag{
				Name:     hash_val_key,
				Category: cat_query,
				Aliases:  []string{"hk"},
				Value:    "",
				Usage:    fmt.Sprintf("specify a hash-key-value, only takes effect when '%s' is specified", table_key),
			},
			&cli.StringFlag{
				Name:     range_val_1_key,
				Category: cat_query,
				Aliases:  []string{"rk"},
				Value:    "",
				Usage:    fmt.Sprintf("specify a range-key-value, only takes effect when '%s' is specified", table_key),
			},
			&cli.StringFlag{
				Name:     range_val_2_key,
				Category: cat_query,
				Aliases:  []string{"rk2"},
				Value:    "",
				Usage:    fmt.Sprintf("specify a second range-key-value, only takes effect when '%s' is specified, and the 'BETWEEN' operator is used", table_key),
			},
			&cli.StringFlag{
				Name:      range_op_key,
				Category:  cat_query,
				Aliases:   []string{"op"},
				Value:     "equals",
				Usage:     fmt.Sprintf("specify the range-key operator, only takes effect when '%s' is specified", table_key),
				Validator: validateRangeOp(),
			},
			&cli.BoolFlag{
				Name:     range_order_key,
				Category: cat_query,
				Aliases:  []string{"descending", "dsc"},
				Value:    false,
				Usage:    fmt.Sprintf("specify the range order, defaults to true ⇒ 'ascending', only takes effect when '%s' is specified", table_key),
			},
			&cli.StringFlag{
				Name:     log_loc_key,
				Category: cat_logging,
				Aliases:  []string{},
				Value:    logDir,
				Usage:    fmt.Sprintf("location of the logfile, when logging is enabled with '%s'", log_debug_key),
			},
			&cli.BoolFlag{
				Name:     log_debug_key,
				Category: cat_logging,
				Aliases:  []string{},
				Value:    false,
				Usage:    fmt.Sprintf("enable debug-level logging to '%s' or the location specified with '%s'", logDir, log_loc_key),
			},
			&cli.BoolFlag{
				Name:     log_key,
				Category: cat_logging,
				Aliases:  []string{},
				Value:    false,
				Usage:    fmt.Sprintf("enable logging to '%s' or the location specified with '%s'", logDir, log_loc_key),
			},
			&cli.BoolFlag{
				Name:     log_text_key,
				Category: cat_logging,
				Aliases:  []string{},
				Value:    false,
				Usage:    "log in text instead of JSON format",
			},
			&cli.BoolFlag{
				Name:     log_append_key,
				Category: cat_logging,
				Aliases:  []string{},
				Value:    false,
				Usage:    "append the existing logfile (truncates by default)",
			},
			&cli.StringFlag{
				Name:     log_level_key,
				Category: cat_logging,
				Aliases:  []string{"level"},
				Value:    "info",
				Usage:    "log-level",
				Validator: func(s string) error {
					switch strings.ToLower(s) {
					case "trace", "debug", "info", "warn", "error":
						return nil
					default:
						return fmt.Errorf("only %q supported", []string{"trace", "debug", "info", "warn", "error"})

					}
				},
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

	logger, logFile, err := initialiseLogger(cmd)
	if err != nil {
		return err
	}

	if logFile != nil {
		defer logFile.Close()
	}

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
		Logger:           logger,
		Profile:          resolveProfile(cmd, cfgf),
		Region:           resolveRegion(cmd, cfgf),
		URL:              urlP,
		AvailableRegions: cfgf.AWSRegions,
		StarredRegions:   cfgf.StarredRegions,
		Tables: appconfig.Tables{
			MaxTables:       cfgf.TablesMax,
			Pagesize:        cfgf.TablesPageSize,
			PrimaryWidth:    cfgf.TablesPrimaryWidth,
			HighlightRegexp: cfgf.TablesHighlightRegexp,
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

	logger.Info("New DYNAMITE session")

	p = tea.NewProgram(ui.NewModel(ctx, cfg, uiopts...))
	_, err = p.Run()
	return err
}

func loadConfig(dirpath string) (configfile.Config, *configfile.ConfigManager, error) {
	dirpath = filepath.Join(dirpath, "config.yaml")
	path, err1 := filepath.Abs(dirpath)
	if err1 != nil {
		err1 = fmt.Errorf("failed to construct a valid config-path: %w", err1)
	}

	configman := configfile.NewConfigManager(path)
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

func initialiseLogger(cmd *cli.Command) (*slog.Logger, *os.File, error) {
	// loglevel
	var logLevel slog.Level
	switch strings.ToLower(cmd.String(log_level_key)) {
	case "trace":
		logLevel = logging.LevelTrace
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	var isLogging = cmd.Bool(log_key)
	if cmd.Bool(log_debug_key) {
		isLogging = true
		logLevel = min(slog.LevelDebug, logLevel)
	}
	if !isLogging {
		// noop by default
		return slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil
	}

	// logpath
	if err := findOrCreateLogLocation(cmd); err != nil {
		return nil, nil, err
	}
	path := filepath.Join(logDir, dynamite_logfile)

	// file
	fileAppend := cmd.Bool(log_append_key)
	fileopts := os.O_RDWR | os.O_CREATE
	if !fileAppend {
		os.Truncate(path, 0)
	} else {
		fileopts = fileopts | os.O_APPEND
	}
	file, err := os.OpenFile(path, fileopts, 0666)
	if err != nil {
		return nil, nil, fmt.Errorf("creating/opening logfile: %w", err)
	}

	// logger
	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return logging.ReplaceLevelName(a)
		},
	}
	logger := slog.New(slog.NewJSONHandler(file, opts))
	if cmd.Bool(log_text_key) {
		logger = slog.New(slog.NewTextHandler(file, opts))
	}

	return logger, file, nil
}

func findOrCreateLogLocation(cmd *cli.Command) error {
	if customDir := cmd.String(log_loc_key); customDir != "" {
		logDir = customDir
	}
	if logDir == "" {
		return fmt.Errorf("could not resolve an appropriate location for logging, please provide a '--%s' value", log_loc_key)
	}
	absPath, err := filepath.Abs(logDir)
	if err != nil {
		return fmt.Errorf("failed to resolve the absolute location of logging directory: %w", err)
	}
	err = os.MkdirAll(absPath, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create the required directories for storing the logfile: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("failed to verify the log directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path '%s' is not a directory", absPath)
	}
	return nil
}
