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
			name: "exclude-errnil true",
			yaml: `exclude-errnil: true`,
			want: Config{ExcludeErrNil: true},
		},
		{
			name: "exclude-errnil false",
			yaml: `exclude-errnil: false`,
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
			name: "exclude-errnil and exclude-err-regexp together",
			yaml: `
exclude-errnil: true
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
			yaml:    "exclude-errnil: [invalid",
			wantErr: true,
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
