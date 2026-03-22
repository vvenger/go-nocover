package ast

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithoutLogRegexp(t *testing.T) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\.Info\(`),
		regexp.MustCompile(`\.Error\(`),
		regexp.MustCompile(`\.Debug\(`),
	}

	tests := []struct {
		name       string
		src        string
		patterns   []*regexp.Regexp
		wantRanges []ExcludeRange
	}{
		{
			// single log statement in a function with other code → only ExprStmt is excluded
			name: "single log stmt in mixed func body",
			src: `package p
func f() {
	log.Info("start")
	doWork()
}`,
			wantRanges: []ExcludeRange{{StartLine: 3, EndLine: 3}},
		},
		{
			// if block with only log calls → entire IfStmt is excluded (single entry)
			name: "if block with only log calls",
			src: `package p
func f(debug bool) {
	if debug {
		log.Info("debug mode")
	}
}`,
			wantRanges: []ExcludeRange{{StartLine: 3, EndLine: 5}},
		},
		{
			// if block with log and other code → only the log statement is excluded
			name: "if block with log and other stmt",
			src: `package p
func f(ok bool) {
	if ok {
		log.Info("ok")
		doWork()
	}
}`,
			wantRanges: []ExcludeRange{{StartLine: 4, EndLine: 4}},
		},
		{
			// func body with only log calls → entire FuncDecl is excluded (single entry)
			name: "func body with only log calls",
			src: `package p
func logStart() {
	log.Info("start")
}`,
			wantRanges: []ExcludeRange{{StartLine: 2, EndLine: 4}},
		},
		{
			// multiple log statements in a mixed function → each statement excluded separately
			name: "multiple log stmts in mixed func",
			src: `package p
func f() {
	log.Info("a")
	doWork()
	log.Error("b")
}`,
			wantRanges: []ExcludeRange{
				{StartLine: 3, EndLine: 3},
				{StartLine: 5, EndLine: 5},
			},
		},
		{
			// chained call logger.Ctx(ctx).Info → excluded
			name: "chained log call",
			src: `package p
func f(ok bool) {
	if ok {
		logger.Ctx(ctx).Info("ok")
	}
}`,
			wantRanges: []ExcludeRange{{StartLine: 3, EndLine: 5}},
		},
		{
			// pattern .logger. matches
			name: "pattern .logger. matches",
			src: `package p
func f(ok bool) {
	if ok {
		s.logger.Info("msg")
	}
}`,
			patterns:   []*regexp.Regexp{regexp.MustCompile(`\.logger\.`)},
			wantRanges: []ExcludeRange{{StartLine: 3, EndLine: 5}},
		},
		{
			// two patterns both match
			name: "2 pattern logger matches",
			src: `package p
func f(ok bool) {
	ctx = logger.WithFields(ctx,
		zap.String("order_id", id.String()),
	)
	if ok {
		s.logger.Info("msg")
	}
}`,
			patterns:   []*regexp.Regexp{regexp.MustCompile(`logger\.WithFields`), regexp.MustCompile(`\.logger\.`)},
			wantRanges: []ExcludeRange{{StartLine: 3, EndLine: 5}, {StartLine: 6, EndLine: 8}},
		},
		{
			// non-log call → not excluded
			name: "non-log call not excluded",
			src: `package p
func f(ok bool) {
	if ok {
		doWork()
	}
}`,
		},
		{
			// empty if body → not excluded
			name: "empty if body not excluded",
			src: `package p
func f(ok bool) {
	if ok {
	}
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := patterns
			if tt.patterns != nil {
				p = tt.patterns
			}
			ranges, err := findExcludeRanges([]byte(tt.src), WithoutLogRegexp(p))
			require.NoError(t, err)
			require.Len(t, ranges, len(tt.wantRanges))
			require.ElementsMatch(t, tt.wantRanges, ranges)
		})
	}
}
