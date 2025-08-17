package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     Server     `yaml:"server"`
	Logging    Logging    `yaml:"logging"`
	Auth       Auth       `yaml:"auth"`
	Strategies []Strategy `yaml:"strategies"`
	Providers  []Provider `yaml:"providers"`
}

type Auth struct {
	Enabled  bool   `yaml:"enabled"`
	Issuer   string `yaml:"issuer"`
	Audience string `yaml:"audience"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Logging struct {
	Level string `yaml:"level"`
}

type Strategy struct {
	Name      string   `yaml:"name"`
	Providers []string `yaml:"providers"`
}

type Provider struct {
	Name       string        `yaml:"name"`
	Enabled    bool          `yaml:"enabled"`
	TargetURL  string        `yaml:"target_url"`
	APIKey     string        `yaml:"api_key"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
}

// Load reads the configuration file from the given path, parses it, and returns a Config struct.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expand environment variables
	data = []byte(os.ExpandEnv(string(data)))

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
