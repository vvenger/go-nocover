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
		name        string
		blocks      []parser.Block
		ranges      map[string][]ast.ExcludeRange
		markNoCover config.MarkBlocks
		markOptions config.MarkBlocks
		expected    []parser.Block
	}{
		{
			name: "nocover deleted",
			blocks: []parser.Block{
				{StartLine: 5, EndLine: 8, Count: 0},
			},
			ranges: map[string][]ast.ExcludeRange{
				"": {{StartLine: 3, EndLine: 10, Source: ast.ExcludeSourceNoCover}},
			},
			markNoCover: config.MarkBlocksDeleted,
			markOptions: config.MarkBlocksTested,
			expected:    []parser.Block{},
		},
		{
			name: "nocover tested",
			blocks: []parser.Block{
				{StartLine: 5, EndLine: 8, Count: 0},
			},
			ranges: map[string][]ast.ExcludeRange{
				"": {{StartLine: 3, EndLine: 10, Source: ast.ExcludeSourceNoCover}},
			},
			markNoCover: config.MarkBlocksTested,
			markOptions: config.MarkBlocksDeleted,
			expected:    []parser.Block{{StartLine: 5, EndLine: 8, Count: 1}},
		},
		{
			name: "option deleted",
			blocks: []parser.Block{
				{StartLine: 5, EndLine: 8, Count: 0},
			},
			ranges: map[string][]ast.ExcludeRange{
				"": {{StartLine: 3, EndLine: 10, Source: ast.ExcludeSourceOption}},
			},
			markNoCover: config.MarkBlocksTested,
			markOptions: config.MarkBlocksDeleted,
			expected:    []parser.Block{},
		},
		{
			name: "option tested",
			blocks: []parser.Block{
				{StartLine: 5, EndLine: 8, Count: 0},
			},
			ranges: map[string][]ast.ExcludeRange{
				"": {{StartLine: 3, EndLine: 10, Source: ast.ExcludeSourceOption}},
			},
			markNoCover: config.MarkBlocksDeleted,
			markOptions: config.MarkBlocksTested,
			expected:    []parser.Block{{StartLine: 5, EndLine: 8, Count: 1}},
		},
		{
			name: "nocover deleted and option tested",
			blocks: []parser.Block{
				{StartLine: 2, EndLine: 4, Count: 0},
				{StartLine: 6, EndLine: 9, Count: 0},
			},
			ranges: map[string][]ast.ExcludeRange{
				"": {
					{StartLine: 1, EndLine: 5, Source: ast.ExcludeSourceNoCover},
					{StartLine: 5, EndLine: 10, Source: ast.ExcludeSourceOption},
				},
			},
			markNoCover: config.MarkBlocksDeleted,
			markOptions: config.MarkBlocksTested,
			expected:    []parser.Block{{StartLine: 6, EndLine: 9, Count: 1}},
		},
		{
			name:        "block outside all ranges",
			blocks:      []parser.Block{{StartLine: 1, EndLine: 2, Count: 0}},
			ranges:      map[string][]ast.ExcludeRange{"": {{StartLine: 5, EndLine: 10}}},
			markNoCover: config.MarkBlocksDeleted,
			markOptions: config.MarkBlocksDeleted,
			expected:    []parser.Block{{StartLine: 1, EndLine: 2, Count: 0}},
		},
		{
			name:        "empty ranges",
			blocks:      []parser.Block{{StartLine: 1, EndLine: 2}, {StartLine: 5, EndLine: 8}},
			ranges:      nil,
			markNoCover: config.MarkBlocksDeleted,
			markOptions: config.MarkBlocksDeleted,
			expected:    []parser.Block{{StartLine: 1, EndLine: 2}, {StartLine: 5, EndLine: 8}},
		},
		{
			name:        "empty blocks",
			blocks:      nil,
			ranges:      map[string][]ast.ExcludeRange{"": {{StartLine: 1, EndLine: 10}}},
			markNoCover: config.MarkBlocksDeleted,
			markOptions: config.MarkBlocksDeleted,
			expected:    []parser.Block{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter(tt.blocks, tt.ranges, tt.markNoCover, tt.markOptions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_findExcluded(t *testing.T) {
	tests := []struct {
		name   string
		b      parser.Block
		ranges []ast.ExcludeRange
		want   ast.ExcludeRange
		wantOK bool
	}{
		{
			name:   "not excluded",
			b:      parser.Block{StartLine: 1, EndLine: 2, Count: 0},
			ranges: []ast.ExcludeRange{{StartLine: 5, EndLine: 10}},
			wantOK: false,
		},
		{
			name:   "excluded",
			b:      parser.Block{StartLine: 3, EndLine: 8},
			ranges: []ast.ExcludeRange{{StartLine: 2, EndLine: 10}},
			want:   ast.ExcludeRange{StartLine: 2, EndLine: 10},
			wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := findExcluded(tt.b, tt.ranges)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}
