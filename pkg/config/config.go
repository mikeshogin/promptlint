package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Tier defines a model tier with routing parameters.
type Tier struct {
	Name          string `yaml:"name"`
	MaxComplexity string `yaml:"max_complexity"` // low, medium, high
	MaxTokens     int    `yaml:"max_tokens"`
	CostPer1k     float64 `yaml:"cost_per_1k"`
}

// Config holds promptlint routing configuration.
type Config struct {
	Tiers       []Tier `yaml:"tiers"`
	DefaultTier string `yaml:"default_tier"`
}

// DefaultConfig returns the built-in haiku/sonnet/opus configuration.
func DefaultConfig() *Config {
	return &Config{
		Tiers: []Tier{
			{Name: "haiku", MaxComplexity: "low", MaxTokens: 500, CostPer1k: 0.80},
			{Name: "sonnet", MaxComplexity: "medium", MaxTokens: 5000, CostPer1k: 3.00},
			{Name: "opus", MaxComplexity: "high", MaxTokens: 100000, CostPer1k: 15.00},
		},
		DefaultTier: "sonnet",
	}
}

// Load reads configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if len(cfg.Tiers) == 0 {
		return nil, fmt.Errorf("config has no tiers defined")
	}

	return &cfg, nil
}

// LoadOrDefault loads config from .promptlint.yaml in the current directory.
// If the file does not exist, it returns DefaultConfig.
func LoadOrDefault() *Config {
	cfg, err := Load(".promptlint.yaml")
	if err != nil {
		return DefaultConfig()
	}
	return cfg
}

// RouteByComplexity returns the model name for the given complexity level.
// Complexity is one of: low, medium, high.
func (c *Config) RouteByComplexity(complexity string) string {
	for _, t := range c.Tiers {
		if t.MaxComplexity == complexity {
			return t.Name
		}
	}
	// Fallback: return default tier if set
	if c.DefaultTier != "" {
		return c.DefaultTier
	}
	// Last resort fallback
	if len(c.Tiers) > 0 {
		return c.Tiers[len(c.Tiers)-1].Name
	}
	return "sonnet"
}
