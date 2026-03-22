package app

import (
	"nocover/internal/parser"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sampleProfile = &parser.Profile{
	Mode: parser.ModeSet,
	Blocks: []parser.Block{
		{File: "github.com/user/repo/pkg/foo.go", StartLine: 10, StartCol: 20, EndLine: 15, EndCol: 3, Stmts: 2, Count: 1},
		{File: "github.com/user/repo/pkg/foo.go", StartLine: 20, StartCol: 5, EndLine: 25, EndCol: 10, Stmts: 3, Count: 0},
	},
}

func output(t *testing.T, p *parser.Profile) []string {
	t.Helper()
	var buf strings.Builder
	require.NoError(t, write(&buf, p))
	return strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
}

func TestWrite_ModeFirstLine(t *testing.T) {
	lines := output(t, sampleProfile)
	assert.Equal(t, "mode: set", lines[0])
}

func TestWrite_LineCount(t *testing.T) {
	lines := output(t, sampleProfile)
	assert.Len(t, lines, 1+len(sampleProfile.Blocks))
}

func TestWrite_BlockFormat(t *testing.T) {
	lines := output(t, sampleProfile)
	assert.Equal(t, "github.com/user/repo/pkg/foo.go:10.20,15.3 2 1", lines[1])
	assert.Equal(t, "github.com/user/repo/pkg/foo.go:20.5,25.10 3 0", lines[2])
}
