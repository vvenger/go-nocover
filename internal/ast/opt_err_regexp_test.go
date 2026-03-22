package ast

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithoutIfErrReturnRegexp(t *testing.T) {
	marshal := []*regexp.Regexp{regexp.MustCompile(`json\.Marshal\(`)}

	tests := []struct {
		name       string
		src        string
		patterns   []*regexp.Regexp
		wantRanges []ExcludeRange
	}{
		{
			name: "init-stmt match",
			src: `package p
func f() error {
	if err := json.Marshal(v); err != nil {
		return err
	}
	return nil
}`,
			patterns:   marshal,
			wantRanges: []ExcludeRange{{StartLine: 3, EndLine: 5}},
		},
		{
			name: "preceding-stmt simple assign",
			src: `package p
func f() error {
	err = json.Marshal(v)
	if err != nil {
		return err
	}
	return nil
}`,
			patterns:   marshal,
			wantRanges: []ExcludeRange{{StartLine: 4, EndLine: 6}},
		},
		{
			name: "preceding-stmt multi-assign",
			src: `package p
func f() ([]byte, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return body, nil
}`,
			patterns:   marshal,
			wantRanges: []ExcludeRange{{StartLine: 4, EndLine: 6}},
		},
		{
			name: "no match — different function",
			src: `package p
func f() error {
	err = otherFn()
	if err != nil {
		return err
	}
	return nil
}`,
			patterns: marshal,
		},
		{
			name: "no match — body has extra statement",
			src: `package p
func f() error {
	err = json.Marshal(v)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}`,
			patterns: marshal,
		},
		{
			name: "no match — has else branch",
			src: `package p
func f() error {
	err = json.Marshal(v)
	if err != nil {
		return err
	} else {
		return nil
	}
}`,
			patterns: marshal,
		},
		{
			name: "no match — variable not err",
			src: `package p
func f() error {
	body, e := json.Marshal(v)
	if e != nil {
		return e
	}
	return body
}`,
			patterns: marshal,
		},
		{
			name: "multiple patterns, two matches",
			src: `package p
func f() error {
	err = json.Marshal(v)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	return nil
}`,
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`json\.Marshal\(`),
				regexp.MustCompile(`json\.Unmarshal\(`),
			},
			wantRanges: []ExcludeRange{
				{StartLine: 4, EndLine: 6},
				{StartLine: 8, EndLine: 10},
			},
		},
		{
			name: "multiple patterns, one match",
			src: `package p
func f() error {
	err = json.Marshal(v)
	if err != nil {
		return err
	}
	err = db.Query()
	if err != nil {
		return err
	}
	return nil
}`,
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`json\.Marshal\(`),
				regexp.MustCompile(`json\.Unmarshal\(`),
			},
			wantRanges: []ExcludeRange{{StartLine: 4, EndLine: 6}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranges, err := findExcludeRanges([]byte(tt.src), WithoutIfErrReturnRegexp(tt.patterns))
			require.NoError(t, err)
			require.Len(t, ranges, len(tt.wantRanges))
			assert.ElementsMatch(t, tt.wantRanges, ranges)
		})
	}
}
