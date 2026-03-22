package app

import (
	"fmt"
	"io"
	"nocover/internal/parser"
	"os"
)

func writeFile(outputPath string, profile *parser.Profile) error {
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	if err := write(out, profile); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func write(w io.Writer, profile *parser.Profile) error {
	if _, err := fmt.Fprintf(w, "mode: %s\n", profile.Mode); err != nil {
		return fmt.Errorf("write mode: %w", err)
	}
	for _, b := range profile.Blocks {
		_, err := fmt.Fprintf(w, "%s:%d.%d,%d.%d %d %d\n",
			b.File, b.StartLine, b.StartCol, b.EndLine, b.EndCol, b.Stmts, b.Count,
		)
		if err != nil {
			return fmt.Errorf("write block: %w", err)
		}
	}
	return nil
}
