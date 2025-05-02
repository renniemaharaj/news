package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Keywords         []string `json:"keywords"`
	NumSitesPerQuery int      `json:"num_sites_per_query"`
}

// Idiomatic load function for config
func Load(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	return &cfg, err
}
