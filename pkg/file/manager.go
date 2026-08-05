package file

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configName = "config.yaml"
	stateName  = ".state.yaml"
)

type Manager struct {
	path       string
	configpath string
	statepath  string
}

func NewManager(dirpath string) *Manager {
	return &Manager{
		path:       dirpath,
		configpath: filepath.Join(dirpath, configName),
		statepath:  filepath.Join(dirpath, stateName),
	}
}

// LoadConfig will always return a valid config, either the default config, or
// the one it could find, regardless of whether errors occurred.
func (m *Manager) LoadConfig() (Config, error) {
	dflt := defaultConfig()

	cfg, err := m.readConfigFromDisk()
	if err != nil || cfg == nil {
		return dflt, err
	}

	return mergeWithDefault(*cfg), nil
}

func (m *Manager) LoadState() (StateFile, error) {
	state, err := m.readStateFromDisk()
	if err != nil {
		return StateFile{}, err
	}
	if state == nil {
		return StateFile{}, nil
	}

	return *state, nil
}

func (m *Manager) readConfigFromDisk() (*configFile, error) {
	f, err := os.Open(m.configpath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}
	defer f.Close()

	var cfg configFile
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file; %w", err)
	}

	// TODO: move to toml config
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file; %w", err)
	}

	return &cfg, nil
}

func (m *Manager) readStateFromDisk() (*StateFile, error) {
	f, err := os.Open(m.statepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}
	defer f.Close()

	var state StateFile
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file; %w", err)
	}

	err = yaml.Unmarshal(bytes, &state)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal state file; %w", err)
	}

	return &state, nil
}

func (m *Manager) writeStateToDisk(f StateFile) error {
	header := "# This is a generated file. DO NOT EDIT.\n"
	bts, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("failed to marshal state file; %w", err)
	}
	buf := bytes.NewBuffer([]byte(header))
	_, err = buf.Write(bts)
	if err != nil {
		return fmt.Errorf("failed to assemble state file; %w", err)
	}
	err = os.WriteFile(m.statepath, buf.Bytes(), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write state file; %w", err)
	}

	return nil
}

func (m *Manager) UpdateRegion(region string) error {
	state, err := m.readStateFromDisk()
	if err != nil {
		return err
	}
	if state == nil {
		state = &StateFile{}
	}
	state.LastUsedRegion = region
	if err := m.writeStateToDisk(*state); err != nil {
		return err
	}

	return nil
}
