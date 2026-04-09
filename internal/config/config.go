package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level portwatch configuration.
type Config struct {
	Interval time.Duration `yaml:"interval"`
	Ports    []PortConfig  `yaml:"ports"`
}

// PortConfig describes a single port to monitor and its associated actions.
type PortConfig struct {
	Host    string   `yaml:"host"`
	Port    int      `yaml:"port"`
	OnOpen  []Action `yaml:"on_open"`
	OnClose []Action `yaml:"on_close"`
}

// Action represents either a webhook call or a shell script to execute.
type Action struct {
	Type    string `yaml:"type"`    // "webhook" or "script"
	URL     string `yaml:"url"`     // used when type == "webhook"
	Command string `yaml:"command"` // used when type == "script"
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{
		Interval: 10 * time.Second, // default poll interval
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// validate performs basic sanity checks on the loaded configuration.
func (c *Config) validate() error {
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	for i, p := range c.Ports {
		if p.Port < 1 || p.Port > 65535 {
			return fmt.Errorf("ports[%d]: port number %d out of range", i, p.Port)
		}
		if p.Host == "" {
			c.Ports[i].Host = "localhost"
		}
		for j, a := range append(p.OnOpen, p.OnClose...) {
			if a.Type != "webhook" && a.Type != "script" {
				return fmt.Errorf("ports[%d] action[%d]: unknown type %q", i, j, a.Type)
			}
		}
	}
	return nil
}
