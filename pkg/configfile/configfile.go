// package config defines the app configuration file and tooling for config i/o
package configfile

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"

	"gopkg.in/yaml.v3"
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
	LastUsedRegion      string   `yaml:"last_used_region"`       // TODO: impl
	DefaultToLastRegion string   `yaml:"default_to_last_region"` // TODO: impl

	DefaultProfile string `yaml:"default_profile"`

	// tables will be paged in automatically on boot. To prevent excessive
	// calls, we specify a limit on how many pages (size of 100) can be
	// retrieved. This parameter specifies the number of tables, not pages.
	MaxTables int `yaml:"max_tables"`
}

type Config struct {
	AWSRegions          []string
	StarredRegions      []string
	DefaultRegion       string
	LastUsedRegion      string
	DefaultToLastRegion bool // TODO: impl
	DefaultProfile      string
	MaxTables           int
}

func defaultConfig() Config {
	return Config{
		AWSRegions:          builtInRegions,
		StarredRegions:      []string{},
		DefaultRegion:       "us-east-1",
		LastUsedRegion:      "",
		DefaultToLastRegion: false,
		DefaultProfile:      "",
		MaxTables:           1000,
	}
}

type ConfigManager struct {
	full string
}

func NewConfigManager(absPath string) *ConfigManager {
	return &ConfigManager{
		full: absPath,
	}
}

// LoadConfig will always return a valid config, either the default config, or
// the one it could find, regardless of whether errors occurred.
func (m *ConfigManager) LoadConfig() (Config, error) {
	dflt := defaultConfig()

	f, err := os.Open(m.full)
	if err != nil {
		if os.IsNotExist(err) {
			return dflt, nil
		}
	}
	defer f.Close()

	var cfg configFile
	bytes, err := io.ReadAll(f)
	if err != nil {
		return dflt, fmt.Errorf("failed to read config file; %w", err)
	}

	// TODO: move to toml config
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return dflt, fmt.Errorf("failed to unmarshal config file; %w", err)
	}

	return mergeWithDefault(cfg), nil
}

func mergeWithDefault(cfg configFile) Config {
	res := defaultConfig()
	res.AWSRegions = unique(append(res.AWSRegions, cfg.AWSRegions...))
	res.StarredRegions = unique(append(res.StarredRegions, cfg.StarredRegions...))
	res.DefaultRegion = notEmptyS(cfg.DefaultRegion, res.DefaultRegion)
	res.LastUsedRegion = notEmptyS(cfg.LastUsedRegion, res.LastUsedRegion)
	if defreg, err := strconv.ParseBool(cfg.DefaultToLastRegion); err != nil {
		res.DefaultToLastRegion = defreg
	}
	res.DefaultProfile = notEmptyS(cfg.DefaultProfile, res.DefaultProfile)
	if cfg.MaxTables > 0 {
		res.MaxTables = cfg.MaxTables
	}

	return res
}

// used for storing default profile on first use (if used with profile)
func (m *ConfigManager) StoreConfig(c Config) error {
	panic("implement me!")
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
