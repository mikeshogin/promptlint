package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ModelTier defines a model with its routing parameters.
type ModelTier struct {
	Name          string `yaml:"name" json:"name"`
	Tier          string `yaml:"tier" json:"tier"`
	CostWeight    int    `yaml:"cost_weight" json:"cost_weight"`
	MaxComplexity int    `yaml:"max_complexity" json:"max_complexity"`
}

// Config holds promptlint routing configuration.
type Config struct {
	Models []ModelTier `yaml:"models" json:"models"`
}

// DefaultConfig returns the built-in haiku/sonnet/opus configuration.
func DefaultConfig() *Config {
	return &Config{
		Models: []ModelTier{
			{Name: "haiku", Tier: "low", CostWeight: 1, MaxComplexity: 30},
			{Name: "sonnet", Tier: "standard", CostWeight: 10, MaxComplexity: 70},
			{Name: "opus", Tier: "high", CostWeight: 30, MaxComplexity: 100},
		},
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

	if len(cfg.Models) == 0 {
		return nil, fmt.Errorf("config has no models defined")
	}

	return &cfg, nil
}

// RouteByScore returns the cheapest model that can handle the given complexity score.
// Score is 0-100, matched against MaxComplexity of each tier.
func (c *Config) RouteByScore(score int) string {
	// Sort by cost weight ascending (cheapest first)
	for _, m := range c.Models {
		if score <= m.MaxComplexity {
			return m.Name
		}
	}
	// Fallback to most expensive model
	if len(c.Models) > 0 {
		return c.Models[len(c.Models)-1].Name
	}
	return "sonnet"
}
