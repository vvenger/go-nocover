package main

import (
	"errors"
	"flag"
	"fmt"
	"nocover/internal/app"
	"nocover/internal/config"
	"os"
	"path/filepath"
)

var errNoModule = errors.New("module directive not found in go.mod")

func main() {
	coveragePath := flag.String("coverage", ".", "path to input coverage profile (default: current directory)")
	outputPath := flag.String("output", "", "path to output file (default: overwrite input)")
	root := flag.String("root", ".", "project root directory (default: current directory)")
	configPath := flag.String("config", "", "path to config file (default: nocover.yaml in root)")
	flag.Parse()

	if *coveragePath == "" {
		fmt.Printf("error: -coverage flag is required")
		flag.Usage()
		os.Exit(1)
	}
	if *outputPath == "" {
		*outputPath = *coveragePath
	}

	cfg, err := loadConfig(*configPath, *root)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	if err = app.Run(*coveragePath, *outputPath, *root, &cfg); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig(path, root string) (config.Config, error) {
	if path == "" {
		path = filepath.Join(root, config.DefaultConfigName)
	}

	cfg, err := config.LoadFromFile(path)
	if err != nil {
		return config.Config{}, fmt.Errorf("can'tload config: %w", err)
	}

	return cfg, nil
}
