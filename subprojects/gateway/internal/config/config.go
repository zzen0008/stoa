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
	RateLimit  RateLimit  `yaml:"ratelimit"`
	Strategies []Strategy `yaml:"strategies"`
	Providers  []Provider `yaml:"providers"`
}

type Auth struct {
	Enabled  bool          `yaml:"enabled"`
	Issuer   string        `yaml:"issuer"`
	Audience string        `yaml:"audience"`
	CacheTTL time.Duration `yaml:"cache_ttl"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Logging struct {
	Level string `yaml:"level"`
}

type RateLimit struct {
	Enabled      bool                `yaml:"enabled"`
	Backend      string              `yaml:"backend"`
	RedisAddress string              `yaml:"redis_address"`
	Default      RateLimitConfig     `yaml:"default"`
	Groups       map[string]RateLimitConfig `yaml:"groups"`
}

type RateLimitConfig struct {
	Name     string        `yaml:"name"`
	Requests int64         `yaml:"requests"`
	Window   time.Duration `yaml:"window"`
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
	Models     []Model       `yaml:"models"`
}

type Model struct {
	Name          string   `yaml:"name"`
	AllowedGroups []string `yaml:"allowed_groups"`
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