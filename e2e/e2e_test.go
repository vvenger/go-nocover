package e2e

import (
	"io"
	"nocover/internal/app"
	"nocover/internal/config"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func configErrNil() io.Reader {
	return strings.NewReader(`exclude-errnil: true`)
}

func TestExcludeErrNil(t *testing.T) {
	e2eDir, err := os.Getwd()
	require.NoError(t, err)

	cfg, err := config.Load(configErrNil())
	require.NoError(t, err)

	root := filepath.Join(e2eDir, "testdata")
	coveragePath := filepath.Join(root, "coverprofile.out")
	outputPath := filepath.Join(root, "exclude-errnil.out")

	err = app.Run(coveragePath, outputPath, root, &cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(data)
	require.NotContains(t, output, "cmd/main.go:71.13,", "main() block must be excluded via //nocover:block")
}
