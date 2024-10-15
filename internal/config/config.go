package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Spotify struct {
	ClientID        string   `yaml:"client_id"`
	ClientSecret    string   `yaml:"client_secret"`
	RedirectURIPort string   `yaml:"redirect_uri_port"`
	Scopes          []string `yaml:"scopes"`
}

type AppConfig struct {
	LogLevel string  `yaml:"log_level"`
	Spotify  Spotify `yaml:"spotify"`
}

func NewFromFile(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var appConfig AppConfig

	if err := yaml.Unmarshal(data, &appConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return &appConfig, nil
}
