package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OutputDir    string   `yaml:"outputDir"`
	Values       []string `yaml:"values"`
	Set          []string `yaml:"set"`
	Include      []string `yaml:"include"`
	Exclude      []string `yaml:"exclude"`
	Strip        string   `yaml:"strip"`
	CopyOnly     []string `yaml:"copyOnly"`
	TemplateOnly []string `yaml:"templateOnly"`
	Readonly     bool     `yaml:"readonly"`
	Clean        bool     `yaml:"clean"`
	Force        bool     `yaml:"force"`
	DryRun       bool     `yaml:"dryRun"`
	Verbose      bool     `yaml:"verbose"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid YAML in config file: %w", err)
	}

	return &cfg, nil
}
