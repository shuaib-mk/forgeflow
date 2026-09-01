package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIURL         string `yaml:"api_url"`
	Token          string `yaml:"token,omitempty"`
	OrganizationID string `yaml:"organization_id,omitempty"`
}

func defaultConfig() Config { return Config{APIURL: "http://localhost:8080"} }
func configPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "forgeflow", "config.yaml"), nil
}
func loadConfig() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaultConfig(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read CLI config: %w", err)
	}
	config := defaultConfig()
	if err := yaml.Unmarshal(content, &config); err != nil {
		return Config{}, fmt.Errorf("parse CLI config: %w", err)
	}
	return config, nil
}
func saveConfig(config Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	content, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		return fmt.Errorf("write CLI config: %w", err)
	}
	return nil
}
