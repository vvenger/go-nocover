package app

import (
	"bufio"
	"errors"
	"fmt"
	"nocover/internal/ast"
	"nocover/internal/config"
	"nocover/internal/parser"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	ErrNoModule = errors.New("wrong go.mod")
)

func Run(coveragePath, outputPath, root string, cfg *config.Config) error {
	opts, err := cfgOptions(cfg)
	if err != nil {
		return fmt.Errorf("parse options:%w", err)
	}

	profile, err := process(coveragePath, root, opts)
	if err != nil {
		return fmt.Errorf("process coverage: %w", err)
	}

	if err := writeFile(outputPath, profile); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func process(coveragePath string, root string, opts []ast.ExcludeFunc) (*parser.Profile, error) {
	moduleName, err := moduleName(root)
	if err != nil {
		return nil, fmt.Errorf("read module name: %w", err)
	}

	profile, err := parser.ParseFile(coveragePath)
	if err != nil {
		return nil, fmt.Errorf("parse coverage: %w", err)
	}

	var (
		seen  = make(map[string]struct{})
		files []string
	)
	for _, b := range profile.Blocks {
		if _, ok := seen[b.File]; !ok && b.Count == 0 {
			seen[b.File] = struct{}{}
			files = append(files, b.File)
		}
	}

	excludeRanges := make(map[string][]ast.ExcludeRange)
	for _, file := range files {
		absPath := parser.ResolvePath(file, moduleName, root)

		ranges, err := ast.FileExcludeRanges(absPath, opts...)
		if err != nil {
			return nil, fmt.Errorf("find exclude ranges for %s: %w", file, err)
		}

		excludeRanges[file] = ranges
	}

	profile.Blocks = filter(profile.Blocks, excludeRanges)

	return profile, nil
}

func moduleName(root string) (string, error) {
	f, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("open go.mod: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if v, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(v), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}

	return "", ErrNoModule
}

func cfgOptions(cfg *config.Config) ([]ast.ExcludeFunc, error) {
	var opts []ast.ExcludeFunc

	if len(cfg.ExcludeLogRegexp) > 0 {
		patterns, err := compilePatterns(cfg.ExcludeLogRegexp)
		if err != nil {
			return nil, fmt.Errorf("invalid exclude-log-regexp: %w", err)
		}
		opts = append(opts, ast.WithoutLogRegexp(patterns))
	}

	if cfg.ExcludeErrNil {
		opts = append(opts, ast.WithoutIfErrReturn())
	} else {
		if len(cfg.ExcludeErrRegexp) > 0 {
			patterns, err := compilePatterns(cfg.ExcludeErrRegexp)
			if err != nil {
				return nil, fmt.Errorf("invalid exclude-err-regexp: %w", err)
			}
			opts = append(opts, ast.WithoutIfErrReturnRegexp(patterns))
		}
		// TODO: support exclude-err-method
	}

	return opts, nil
}

func compilePatterns(raw []string) ([]*regexp.Regexp, error) {
	patterns := make([]*regexp.Regexp, 0, len(raw))
	for _, s := range raw {
		p, err := regexp.Compile(s)
		if err != nil {
			return nil, fmt.Errorf("compile pattern %q: %w", s, err)
		}
		patterns = append(patterns, p)
	}
	return patterns, nil
}
