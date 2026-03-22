package app

import (
	"nocover/internal/ast"
	"nocover/internal/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []parser.Block
		ranges   map[string][]ast.ExcludeRange
		expected []parser.Block
	}{
		{
			name: "all blocks excluded",
			blocks: []parser.Block{
				{StartLine: 2, EndLine: 4},
				{StartLine: 6, EndLine: 9},
			},
			ranges: map[string][]ast.ExcludeRange{
				"": {
					{StartLine: 1, EndLine: 5},
					{StartLine: 5, EndLine: 10},
				},
			},
			expected: []parser.Block{},
		},
		{
			name:     "block inside range removed",
			blocks:   []parser.Block{{StartLine: 5, EndLine: 8}},
			ranges:   map[string][]ast.ExcludeRange{"": {{StartLine: 3, EndLine: 10}}},
			expected: []parser.Block{},
		},
		{
			name:     "block outside range kept",
			blocks:   []parser.Block{{StartLine: 1, EndLine: 2}},
			ranges:   map[string][]ast.ExcludeRange{"": {{StartLine: 5, EndLine: 10}}},
			expected: []parser.Block{{StartLine: 1, EndLine: 2}},
		},
		{
			name:     "block partial overlap kept",
			blocks:   []parser.Block{{StartLine: 3, EndLine: 12}},
			ranges:   map[string][]ast.ExcludeRange{"": {{StartLine: 5, EndLine: 10}}},
			expected: []parser.Block{{StartLine: 3, EndLine: 12}},
		},
		{
			name:     "empty ranges keeps all",
			blocks:   []parser.Block{{StartLine: 1, EndLine: 2}, {StartLine: 5, EndLine: 8}},
			ranges:   nil,
			expected: []parser.Block{{StartLine: 1, EndLine: 2}, {StartLine: 5, EndLine: 8}},
		},
		{
			name:     "empty blocks returns empty",
			blocks:   nil,
			ranges:   map[string][]ast.ExcludeRange{"": {{StartLine: 1, EndLine: 10}}},
			expected: []parser.Block{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter(tt.blocks, tt.ranges)
			assert.Equal(t, tt.expected, result)
		})
	}
}
