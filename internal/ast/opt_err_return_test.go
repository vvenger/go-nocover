package ast

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithoutIfErrReturn(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		wantRanges []ExcludeRange
	}{
		{
			name: "basic if err != nil return",
			src: `package p
func f() error {
	if err != nil {
		return err
	}
	return nil
}`,
			wantRanges: []ExcludeRange{{StartLine: 3, EndLine: 5}},
		},
		{
			name: "if with init and return",
			src: `package p
func f() error {
	if err := do(); err != nil {
		return err
	}
	return nil
}`,
			wantRanges: []ExcludeRange{{StartLine: 3, EndLine: 5}},
		},
		{
			name: "body single statement, bare return",
			src: `package p
func f() {
	if err != nil {
     	http.Error(w, err.Error(), http.StatusBadRequest)
        return
	}
	method()	
}`,
			wantRanges: []ExcludeRange{{StartLine: 3, EndLine: 6}},
		},
		{
			name: "body has extra statement — not excluded",
			src: `package p
func f() error {
	if err != nil {
		log()
		return err
	}
	return nil
}`,
		},
		{
			name: "variable not err — not excluded",
			src: `package p
func f() error {
	if x != nil {
		return x
	}
	return nil
}`,
		},
		{
			name: "operator == not != — not excluded",
			src: `package p
func f() error {
	if err == nil {
		return nil
	}
	return nil
}`,
		},
		{
			name: "has else — not excluded",
			src: `package p
func f() error {
	if err != nil {
		return err
	} else {
		return nil
	}
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranges, err := findExcludeRanges([]byte(tt.src), WithoutIfErrReturn())
			require.NoError(t, err)
			require.Len(t, ranges, len(tt.wantRanges))
			assert.ElementsMatch(t, tt.wantRanges, ranges)
		})
	}
}

func TestWithoutIfErrReturn_LogPreExcluded(t *testing.T) {
	src := `package p
func f() error {
	if err != nil {
		log()
		return err
	}
	return nil
}`

	logPat := regexp.MustCompile(`^log\(\)$`)
	ranges, err := findExcludeRanges([]byte(src),
		WithoutLogRegexp([]*regexp.Regexp{logPat}),
		WithoutIfErrReturn(),
	)
	require.NoError(t, err)
	// log() at line 4 excluded first; WithoutIfErrReturn sees effective body = [return err] → excludes if block
	assert.ElementsMatch(t, []ExcludeRange{
		{StartLine: 4, EndLine: 4},
		{StartLine: 3, EndLine: 6},
	}, ranges)
}
