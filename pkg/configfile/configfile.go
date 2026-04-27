// package config defines the app configuration file and tooling for config i/o
package configfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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

type ConfigFile struct {
	AWSRegions          []string `yaml:"aws_regions"`
	StarredRegions      []string `yaml:"starred_regions"`
	DefaultRegion       string   `yaml:"default_region"`
	LastUsedRegion      string   `yaml:"last_used_region"`       // TODO: impl
	DefaultToLastRegion bool     `yaml:"default_to_last_region"` // TODO: impl

	DefaultProfile string `yaml:"default_profile"`

	// tables will be paged in automatically on boot. To prevent excessive
	// calls, we specify a limit on how many pages (size of 100) can be
	// retrieved. This parameter specifies the number of tables, not pages.
	MaxTables int `yaml:"max_tables"`
}

func defaultConfig() ConfigFile {
	return ConfigFile{
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

func (m *ConfigManager) LoadConfig(create bool) (ConfigFile, error) {
	f, err := os.Open(m.full)
	if err != nil {
		if os.IsNotExist(err) && create {
			return m.create()
		}
	}
	defer f.Close()

	dflt := defaultConfig()
	var cfg ConfigFile
	bytes, err := io.ReadAll(f)
	if err != nil {
		return dflt, fmt.Errorf("failed to read config file; %w", err)
	}
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return dflt, fmt.Errorf("failed to unmarshal config file; %w", err)
	}

	if cfg.MaxTables == 0 {
		cfg.MaxTables = dflt.MaxTables
	}

	return cfg, nil
}

func (m *ConfigManager) create() (ConfigFile, error) {
	cfg := defaultConfig()

	dir := filepath.Dir(m.full)
	err := os.MkdirAll(filepath.Dir(m.full), 0755)
	if err != nil {
		return cfg, fmt.Errorf("mkDirAll: %s; %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to marshal config file to YAML: %w", err)
	}

	if err := os.WriteFile(m.full, data, 0644); err != nil {
		return cfg, fmt.Errorf("failed to write config file: %w", err)
	}

	return cfg, nil
}

// used for storing default profile on first use (if used with profile)
func (m *ConfigManager) StoreConfig(c ConfigFile) error {
	panic("implement me!")
}
