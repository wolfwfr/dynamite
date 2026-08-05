// package config defines the app configuration file and tooling for config i/o
package file

import (
	"slices"
	"strconv"

	"github.com/wolfwfr/dynamite/pkg/theme"
	"github.com/wolfwfr/dynamite/pkg/util"
)

var builtInRegions = []string{
	"us-east-1",
	"us-east-2",
	"us-west-1",
	"us-west-2",
	"us-gov-east-1",
	"us-gov-west-1",
	"sa-east-1",
	"mx-central-1",
	"me-south-1",
	"me-central-1",
	"il-central-1",
	"eusc-de-east-1",
	"eu-west-1",
	"eu-west-2",
	"eu-west-3",
	"eu-south-1",
	"eu-south-2",
	"eu-north-1",
	"eu-central-1",
	"eu-central-2",
	"cn-northwest-1",
	"cn-north-1",
	"ca-west-1",
	"ca-central-1",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-southeast-3",
	"ap-southeast-4",
	"ap-southeast-5",
	"ap-southeast-6",
	"ap-southeast-7",
	"ap-south-1",
	"ap-south-2",
	"ap-northeast-1",
	"ap-northeast-2",
	"ap-northeast-3",
	"ap-east-1",
	"ap-east-2",
	"af-south-1",
}

type configFile struct {
	AWSRegions          []string `yaml:"aws_regions"`
	StarredRegions      []string `yaml:"starred_regions"`
	DefaultRegion       string   `yaml:"default_region"`
	DefaultToLastRegion string   `yaml:"default_to_last_region"`

	DefaultProfile string `yaml:"default_profile"`

	// tables will be paged in automatically on boot. To prevent excessive
	// calls, we specify a limit on how many pages (size of 100) can be
	// retrieved. This parameter specifies the number of tables, not pages.
	Tables         TableViewSettings    `yaml:"tables"`
	Items          ItemViewSettings     `yaml:"items"`
	ThemeOverrides theme.ThemeOverrides `yaml:"colors"`
}

type TableViewSettings struct {
	PrimaryWidth    int      `yaml:"primary_width_percent"`
	MaxTables       int      `yaml:"max_tables"`
	PageSize        int      `yaml:"page_size"`
	HighLightRegexp []string `yaml:"highlight_regexp"`
}

type ItemViewSettings struct {
	PrimaryWidth  int    `yaml:"primary_width_percent"`
	PageSize      int    `yaml:"page_size"`
	DefaultFormat string `yaml:"format"`
}

type Config struct {
	AWSRegions            []string
	StarredRegions        []string
	DefaultRegion         string
	DefaultToLastRegion   bool
	DefaultProfile        string
	TablesPrimaryWidth    int
	TablesPageSize        int
	TablesMax             int
	TablesHighlightRegexp []string
	ItemsPrimaryWidth     int
	ItemsPageSize         int
	ItemsDefaultFormat    string
	ThemeOverrides        theme.ThemeOverrides
}

func defaultConfig() Config {
	return Config{
		AWSRegions:          builtInRegions,
		StarredRegions:      []string{},
		DefaultRegion:       "us-east-1",
		DefaultToLastRegion: false,
		DefaultProfile:      "",
		TablesPrimaryWidth:  50,
		TablesPageSize:      0, // initialised by view, based on window size
		TablesMax:           1000,
		ItemsPrimaryWidth:   50,
		ItemsPageSize:       0, // initialised by view, based on window size
	}
}

func mergeWithDefault(cfg configFile) Config {
	res := defaultConfig()
	res.AWSRegions = unique(append(res.AWSRegions, cfg.AWSRegions...))
	res.StarredRegions = unique(append(res.StarredRegions, cfg.StarredRegions...))
	res.DefaultRegion = notEmptyS(cfg.DefaultRegion, res.DefaultRegion)
	if defreg, err := strconv.ParseBool(cfg.DefaultToLastRegion); err == nil {
		res.DefaultToLastRegion = defreg
	}
	res.DefaultProfile = notEmptyS(cfg.DefaultProfile, res.DefaultProfile)
	if cfg.Tables.MaxTables > 0 {
		res.TablesMax = cfg.Tables.MaxTables
	}
	res.TablesPrimaryWidth = util.Ternary(cfg.Tables.PrimaryWidth, res.TablesPrimaryWidth, cfg.Tables.PrimaryWidth > 0)
	res.ItemsPrimaryWidth = util.Ternary(cfg.Items.PrimaryWidth, res.ItemsPrimaryWidth, cfg.Items.PrimaryWidth > 0)

	res.ItemsPageSize = util.Ternary(cfg.Items.PageSize, res.ItemsPageSize, cfg.Items.PageSize > 0)
	res.TablesPageSize = util.Ternary(cfg.Tables.PageSize, res.TablesPageSize, cfg.Tables.PageSize > 0)

	res.TablesHighlightRegexp = cfg.Tables.HighLightRegexp
	res.ThemeOverrides = cfg.ThemeOverrides

	res.ItemsDefaultFormat = cfg.Items.DefaultFormat

	return res
}

func unique[E comparable, S ~[]E](s S) S {
	seen := map[E]struct{}{}
	res := make(S, 0, len(s))
	for _, e := range s {
		if _, ok := seen[e]; ok {
			continue
		}
		seen[e] = struct{}{}
		res = append(res, e)
	}
	return slices.Clip(res)
}

func notEmptyS(strings ...string) string {
	var res string
	for _, s := range strings {
		if s != "" {
			return s
		}
	}
	return res
}
