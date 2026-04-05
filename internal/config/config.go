package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultConfigName = "nocover.yaml"

type MarkBlocks string

const (
	MarkBlocksNone    MarkBlocks = ""
	MarkBlocksDeleted MarkBlocks = "deleted"
	MarkBlocksTested  MarkBlocks = "tested"
)

type Config struct {
	MarkNoCover      MarkBlocks `yaml:"mark-no-cover"`
	MarkOptions      MarkBlocks `yaml:"mark-options"`
	ExcludeErrRegexp []string   `yaml:"exclude-err-regexp"`
	ExcludeLogRegexp []string   `yaml:"exclude-log-regexp"`
	ExcludeErrNil    bool       `yaml:"exclude-err-nil"`
}

func LoadFromFile(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("can't open config: %w", err)
	}
	defer f.Close()

	return Load(f)
}

func Load(r io.Reader) (Config, error) {
	var cfg Config
	if err := yaml.NewDecoder(r).Decode(&cfg); err != nil {
		if errors.Is(err, io.EOF) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("can't parse config: %w", err)
	}
	return cfg, nil
}

func (c Config) MarkNoCoverValue() MarkBlocks {
	if c.MarkNoCover == MarkBlocksTested {
		return MarkBlocksTested
	}
	return MarkBlocksDeleted
}

func (c Config) MarkOptionsValue() MarkBlocks {
	if c.MarkOptions == MarkBlocksDeleted {
		return MarkBlocksDeleted
	}
	return MarkBlocksTested
}
