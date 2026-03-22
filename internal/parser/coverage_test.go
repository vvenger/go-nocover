package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleProfile = `mode: set
github.com/user/repo/pkg/foo.go:10.20,15.3 2 1
github.com/user/repo/pkg/foo.go:20.5,25.10 3 0
`

func TestParseFile_Valid(t *testing.T) {
	p, err := parseFile(strings.NewReader(sampleProfile))
	require.NoError(t, err)
	require.NotNil(t, p)
}

func TestParseFile_Mode(t *testing.T) {
	p, err := parseFile(strings.NewReader(sampleProfile))
	require.NoError(t, err)
	assert.Equal(t, ModeSet, p.Mode)
}

func TestParseFile_Blocks(t *testing.T) {
	p, err := parseFile(strings.NewReader(sampleProfile))
	require.NoError(t, err)
	require.Len(t, p.Blocks, 2)

	tests := []struct {
		idx                 int
		file                string
		startLine, startCol int
		endLine, endCol     int
		stmts, count        int
	}{
		{0, "github.com/user/repo/pkg/foo.go", 10, 20, 15, 3, 2, 1},
		{1, "github.com/user/repo/pkg/foo.go", 20, 5, 25, 10, 3, 0},
	}
	for _, tt := range tests {
		b := p.Blocks[tt.idx]
		assert.Equal(t, tt.file, b.File)
		assert.Equal(t, tt.startLine, b.StartLine)
		assert.Equal(t, tt.startCol, b.StartCol)
		assert.Equal(t, tt.endLine, b.EndLine)
		assert.Equal(t, tt.endCol, b.EndCol)
		assert.Equal(t, tt.stmts, b.Stmts)
		assert.Equal(t, tt.count, b.Count)
	}
}

func TestParseFile_InvalidLine(t *testing.T) {
	_, err := parseFile(strings.NewReader("mode: set\nbadline\n"))
	require.ErrorIs(t, err, ErrBadFormat)
}

func TestParseFile_InvalidMode(t *testing.T) {
	_, err := parseFile(strings.NewReader("badmode\n"))
	require.ErrorIs(t, err, ErrBadFormat)
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name       string
		file       string
		moduleName string
		root       string
		expected   string
	}{
		{
			name:       "strips module prefix",
			file:       "github.com/user/repo/pkg/foo.go",
			moduleName: "github.com/user/repo",
			root:       "/project",
			expected:   "/project/pkg/foo.go",
		},
		{
			name:       "no module prefix match",
			file:       "other/module/pkg/foo.go",
			moduleName: "github.com/user/repo",
			root:       "/project",
			expected:   "/project/other/module/pkg/foo.go",
		},
		{
			name:       "root is dot",
			file:       "github.com/user/repo/main.go",
			moduleName: "github.com/user/repo",
			root:       ".",
			expected:   "main.go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Block{File: tt.file}
			got := ResolvePath(b.File, tt.moduleName, tt.root)
			assert.Equal(t, tt.expected, got)
		})
	}
}
