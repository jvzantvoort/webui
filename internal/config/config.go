package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port  int    `yaml:"port"`
	Index string `yaml:"index"`
}

// BrowserConfig holds browser autostart settings.
type BrowserConfig struct {
	Application string `yaml:"application"`
	Autostart   bool   `yaml:"autostart"`
}

// ContentItem defines a content source (markdown folder, static files).
type ContentItem struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Content     string   `yaml:"content"`
	ServeImages bool     `yaml:"serve_images"`
	Index       string   `yaml:"index"`  // file served at the section root; defaults to README.md
	Ignore      []string `yaml:"ignore"` // glob patterns excluded from the nav listing
	Menu        string   `yaml:"menu"`
}

// DataItem defines a data source (CSV with CRUD form).
type DataItem struct {
	Name           string `yaml:"name"`
	Path           string `yaml:"path"`
	Form           string `yaml:"form"`
	Schema         string `yaml:"schema"`
	RecordTemplate string `yaml:"record_template"`
	Menu           string `yaml:"menu"`
}

// Config is the top-level configuration.
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Browser BrowserConfig `yaml:"browser"`
	Content []ContentItem `yaml:"content"`
	Data    []DataItem    `yaml:"data"`
}

// Load reads and parses a YAML config file.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 3110
	}

	return &cfg, nil
}
