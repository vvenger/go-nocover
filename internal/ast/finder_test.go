package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ranges(t *testing.T, src string) []ExcludeRange {
	t.Helper()
	rs, err := findExcludeRanges([]byte(src))
	require.NoError(t, err)
	return rs
}

func TestFindExcludeRanges_FuncCommentAbove(t *testing.T) {
	src := `package main

//nocover:block
func foo() {
	_ = 1
}
`
	rs := ranges(t, src)
	require.Len(t, rs, 1)
	assert.Equal(t, 4, rs[0].StartLine)
	assert.Equal(t, 6, rs[0].EndLine)
}

func TestFindExcludeRanges_FuncCommentSameLine(t *testing.T) {
	src := `package main

func foo() { //nocover:block
	_ = 1
}
`
	rs := ranges(t, src)
	require.Len(t, rs, 1)
	assert.Equal(t, 3, rs[0].StartLine)
	assert.Equal(t, 5, rs[0].EndLine)
}

func TestFindExcludeRanges_IfBlock(t *testing.T) {
	src := `package main

func foo(x int) {
	if x > 0 { //nocover:block
		_ = x
	}
}
`
	rs := ranges(t, src)
	require.Len(t, rs, 1)
	assert.Equal(t, 4, rs[0].StartLine)
	assert.Equal(t, 6, rs[0].EndLine)
}

func TestFindExcludeRanges_NoComment(t *testing.T) {
	src := `package main

func foo() {
	_ = 1
}
`
	rs := ranges(t, src)
	assert.Empty(t, rs)
}

func TestFindExcludeRanges_CommentAboveRecognized(t *testing.T) {
	src := `package main

func foo(x int) {
	//nocover:block
	if x > 0 {
		_ = x
	}
}
`
	rs := ranges(t, src)
	require.Len(t, rs, 1)
	assert.Equal(t, 5, rs[0].StartLine)
	assert.Equal(t, 7, rs[0].EndLine)
}
