package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	ErrBadFormat = errors.New("bad format")
)

type ProfileMode string

const (
	ModeSet    ProfileMode = "set"
	ModeCount  ProfileMode = "count"
	ModeAtomic ProfileMode = "atomic"
)

type Block struct {
	File      string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	Stmts     int
	Count     int
}

type Profile struct {
	Mode   ProfileMode
	Blocks []Block
}

func ParseFile(path string) (*Profile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open coverage file: %w", err)
	}
	defer f.Close()

	p, err := parseFile(f)
	if err != nil {
		return nil, fmt.Errorf("parse coverage file: %w", err)
	}

	return p, nil
}

func parseFile(f io.Reader) (*Profile, error) {
	var p Profile

	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if lineNum == 1 {
			mode, err := parseMode(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}
			p.Mode = mode
			continue
		}

		b, err := parseBlock(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		p.Blocks = append(p.Blocks, b)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read coverage file: %w", err)
	}
	return &p, nil
}

func parseMode(line string) (ProfileMode, error) {
	const prefix = "mode: "

	mode, ok := strings.CutPrefix(line, prefix)
	if !ok {
		return "", fmt.Errorf("invalid mode: %w", ErrBadFormat)
	}

	mode = strings.TrimSpace(mode)
	if mode == "" {
		return "", fmt.Errorf("invalid mode value: %w", ErrBadFormat)
	}

	return ProfileMode(mode), nil
}

// file:startLine.startCol,endLine.endCol stmts count
// main.go:8.13,12.20 3 0
func parseBlock(line string) (Block, error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return Block{}, fmt.Errorf("wrong field count: %w", ErrBadFormat)
	}

	stmts, err := strconv.Atoi(fields[1])
	if err != nil {
		return Block{}, fmt.Errorf("invalid stmts %q: %w", fields[1], err)
	}
	count, err := strconv.Atoi(fields[2])
	if err != nil {
		return Block{}, fmt.Errorf("invalid count %q: %w", fields[2], err)
	}

	filepos := fields[0]
	sep := strings.LastIndex(filepos, ":")
	if sep == -1 {
		return Block{}, ErrBadFormat
	}
	file := filepos[:sep]
	pos := filepos[sep+1:]

	comma := strings.Index(pos, ",")
	if comma == -1 {
		return Block{}, ErrBadFormat
	}
	startLine, startCol, err := parseLineCol(pos[:comma])
	if err != nil {
		return Block{}, fmt.Errorf("start position: %w", err)
	}
	endLine, endCol, err := parseLineCol(pos[comma+1:])
	if err != nil {
		return Block{}, fmt.Errorf("end position: %w", err)
	}

	return Block{
		File:      file,
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		Stmts:     stmts,
		Count:     count,
	}, nil
}

func ResolvePath(file, moduleName, root string) string {
	rel := strings.TrimPrefix(file, moduleName+"/")
	return filepath.Join(root, rel)
}

func parseLineCol(s string) (line, col int, err error) {
	dot := strings.Index(s, ".")
	if dot == -1 {
		return 0, 0, ErrBadFormat
	}

	line, err = strconv.Atoi(s[:dot])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line %q: %w", s[:dot], err)
	}

	col, err = strconv.Atoi(s[dot+1:])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid col %q: %w", s[dot+1:], err)
	}

	return line, col, nil
}
