package app

import (
	"nocover/internal/ast"
	"nocover/internal/parser"
)

func filter(blocks []parser.Block, ranges map[string][]ast.ExcludeRange) []parser.Block {
	result := make([]parser.Block, 0, len(blocks))
	for _, b := range blocks {
		v, ok := ranges[b.File]
		if ok && excluded(b, v) {
			continue
		}
		result = append(result, b)
	}
	return result
}

func excluded(b parser.Block, ranges []ast.ExcludeRange) bool {
	for _, r := range ranges {
		if b.StartLine >= r.StartLine && b.EndLine <= r.EndLine {
			return true
		}
	}
	return false
}
