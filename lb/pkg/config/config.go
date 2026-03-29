package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type BackendConfig struct {
	URL     string `yaml:"url"`
	Weight  int    `yaml:"weight"`
	Timeout string `yaml:"timeout"`
}

type RouteConfig struct {
	Prefix   string          `yaml:"prefix"`
	Strategy string          `yaml:"strategy"`
	Backends []BackendConfig `yaml:"backends"`
}

type Config struct {
	Port   int           `yaml:"port"`
	Admin  int           `yaml:"admin_port"`
	Routes []RouteConfig `yaml:"routes"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
