package app

import (
	"nocover/internal/ast"
	"nocover/internal/config"
	"nocover/internal/parser"
)

func filter(blocks []parser.Block, ranges map[string][]ast.ExcludeRange, markNoCover, markOptions config.MarkBlocks) []parser.Block {
	result := make([]parser.Block, 0, len(blocks))
	for _, b := range blocks {
		if b.Count != 0 {
			result = append(result, b)
			continue
		}

		v, ok := ranges[b.File]
		if !ok {
			result = append(result, b)
			continue
		}

		r, ok := findExcluded(b, v)
		if !ok {
			result = append(result, b)
			continue
		}

		mark := markNoCover
		if r.Source == ast.ExcludeSourceOption {
			mark = markOptions
		}
		if mark == config.MarkBlocksTested {
			b.Count = 1
			result = append(result, b)
		}
	}
	return result
}

func findExcluded(b parser.Block, ranges []ast.ExcludeRange) (ast.ExcludeRange, bool) {
	for _, r := range ranges {
		if b.StartLine >= r.StartLine && b.EndLine <= r.EndLine {
			return r, true
		}
	}
	return ast.ExcludeRange{}, false
}
