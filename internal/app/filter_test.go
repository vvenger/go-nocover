package app

import (
	"nocover/internal/ast"
	"nocover/internal/config"
	"nocover/internal/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []parser.Block
		ranges   map[string][]ast.ExcludeRange
		mark     config.MarkBlocks
		expected []parser.Block
	}{
		{
			name: "deleted: all blocks excluded",
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
			mark:     config.MarkBlocksDeleted,
			expected: []parser.Block{},
		},
		{
			name:     "deleted: block inside range removed",
			blocks:   []parser.Block{{StartLine: 5, EndLine: 8}},
			ranges:   map[string][]ast.ExcludeRange{"": {{StartLine: 3, EndLine: 10}}},
			mark:     config.MarkBlocksDeleted,
			expected: []parser.Block{},
		},
		{
			name:     "deleted: block outside range kept",
			blocks:   []parser.Block{{StartLine: 1, EndLine: 2}},
			ranges:   map[string][]ast.ExcludeRange{"": {{StartLine: 5, EndLine: 10}}},
			mark:     config.MarkBlocksDeleted,
			expected: []parser.Block{{StartLine: 1, EndLine: 2}},
		},
		{
			name:     "deleted: block partial overlap kept",
			blocks:   []parser.Block{{StartLine: 3, EndLine: 12}},
			ranges:   map[string][]ast.ExcludeRange{"": {{StartLine: 5, EndLine: 10}}},
			mark:     config.MarkBlocksDeleted,
			expected: []parser.Block{{StartLine: 3, EndLine: 12}},
		},
		{
			name:     "deleted: empty ranges keeps all",
			blocks:   []parser.Block{{StartLine: 1, EndLine: 2}, {StartLine: 5, EndLine: 8}},
			ranges:   nil,
			mark:     config.MarkBlocksDeleted,
			expected: []parser.Block{{StartLine: 1, EndLine: 2}, {StartLine: 5, EndLine: 8}},
		},
		{
			name:     "deleted: empty blocks returns empty",
			blocks:   nil,
			ranges:   map[string][]ast.ExcludeRange{"": {{StartLine: 1, EndLine: 10}}},
			mark:     config.MarkBlocksDeleted,
			expected: []parser.Block{},
		},
		{
			name:     "tested: excluded block marked as covered",
			blocks:   []parser.Block{{StartLine: 5, EndLine: 8, Count: 0}},
			ranges:   map[string][]ast.ExcludeRange{"": {{StartLine: 3, EndLine: 10}}},
			mark:     config.MarkBlocksTested,
			expected: []parser.Block{{StartLine: 5, EndLine: 8, Count: 1}},
		},
		{
			name: "tested: only excluded blocks get Count=1",
			blocks: []parser.Block{
				{StartLine: 2, EndLine: 4, Count: 0},
				{StartLine: 6, EndLine: 9, Count: 0},
			},
			ranges: map[string][]ast.ExcludeRange{
				"": {{StartLine: 5, EndLine: 10}},
			},
			mark: config.MarkBlocksTested,
			expected: []parser.Block{
				{StartLine: 2, EndLine: 4, Count: 0},
				{StartLine: 6, EndLine: 9, Count: 1},
			},
		},
		{
			name:     "tested: block outside range kept as-is",
			blocks:   []parser.Block{{StartLine: 1, EndLine: 2, Count: 0}},
			ranges:   map[string][]ast.ExcludeRange{"": {{StartLine: 5, EndLine: 10}}},
			mark:     config.MarkBlocksTested,
			expected: []parser.Block{{StartLine: 1, EndLine: 2, Count: 0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter(tt.blocks, tt.ranges, tt.mark)
			assert.Equal(t, tt.expected, result)
		})
	}
}
