package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    Config
		wantErr bool
	}{
		{
			name: "empty file",
			yaml: "",
			want: Config{},
		},
		{
			name: "exclude-err-nil true",
			yaml: `exclude-err-nil: true`,
			want: Config{ExcludeErrNil: true},
		},
		{
			name: "exclude-err-nil false",
			yaml: `exclude-err-nil: false`,
			want: Config{ExcludeErrNil: false},
		},
		{
			name: "exclude-err-regexp",
			yaml: `
exclude-err-regexp:
  - json\.Marshal\(
  - json\.Unmarshal\(
`,
			want: Config{
				ExcludeErrRegexp: []string{`json\.Marshal\(`, `json\.Unmarshal\(`},
			},
		},
		{
			name: "exclude-err-nil and exclude-err-regexp together",
			yaml: `
exclude-err-nil: true
exclude-err-regexp:
  - json\.Marshal\(
`,
			want: Config{
				ExcludeErrNil:    true,
				ExcludeErrRegexp: []string{`json\.Marshal\(`},
			},
		},
		{
			name:    "malformed yaml",
			yaml:    "exclude-err-nil: [invalid",
			wantErr: true,
		},
		{
			name: "mark-no-cover deleted",
			yaml: `mark-no-cover: deleted`,
			want: Config{MarkNoCover: MarkBlocksDeleted},
		},
		{
			name: "mark-no-cover tested",
			yaml: `mark-no-cover: tested`,
			want: Config{MarkNoCover: MarkBlocksTested},
		},
		{
			name: "mark-options deleted",
			yaml: `mark-options: deleted`,
			want: Config{MarkOptions: MarkBlocksDeleted},
		},
		{
			name: "mark-options tested",
			yaml: `mark-options: tested`,
			want: Config{MarkOptions: MarkBlocksTested},
		},
		{
			name: "mark-no-cover and mark-options together",
			yaml: "mark-no-cover: deleted\nmark-options: tested",
			want: Config{MarkNoCover: MarkBlocksDeleted, MarkOptions: MarkBlocksTested},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(strings.NewReader(tt.yaml))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
