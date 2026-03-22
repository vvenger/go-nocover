package ast

import (
	"bytes"
	"fmt"
	goast "go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
)

const nocoverComment = "//nocover:block"

type ExcludeRange struct {
	StartLine int
	EndLine   int
}

type ExcludeFunc func(ast *astFile, fset *token.FileSet, f *goast.File) error

func FileExcludeRanges(filePath string, opts ...ExcludeFunc) ([]ExcludeRange, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("can't read %s: %w", filePath, err)
	}

	return findExcludeRanges(file, opts...)
}

func FindExcludeRanges(file []byte, opts ...ExcludeFunc) ([]ExcludeRange, error) {
	return findExcludeRanges(file, opts...)
}

func findExcludeRanges(file []byte, opts ...ExcludeFunc) ([]ExcludeRange, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", file, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	aFile := newAstFile(f)

	goast.Inspect(f, func(n goast.Node) bool {
		if n == nil {
			return false
		}

		if !isTrackedNode(n) {
			return true
		}

		if !hasNocoverComment(fset, f.Comments, n) {
			return true
		}

		aFile.add(ExcludeRange{
			StartLine: fset.Position(n.Pos()).Line,
			EndLine:   fset.Position(n.End()).Line,
		})

		return false
	})

	// Apply each ExcludeFunc.
	for _, opt := range opts {
		err := opt(aFile, fset, f)
		if err != nil {
			return nil, fmt.Errorf("exclude option: %w", err)
		}
	}

	return aFile.ranges, nil
}

func isTrackedNode(n goast.Node) bool {
	switch n.(type) {
	case *goast.FuncDecl, *goast.BlockStmt, *goast.IfStmt, *goast.ForStmt,
		*goast.RangeStmt, *goast.SwitchStmt, *goast.TypeSwitchStmt,
		*goast.SelectStmt, *goast.CaseClause:
		return true
	}
	return false
}

func hasNocoverComment(fset *token.FileSet, comments []*goast.CommentGroup, n goast.Node) bool {
	nodeLine := fset.Position(n.Pos()).Line
	lbraceLine := lbraceLineOf(fset, n)

	for _, cg := range comments {
		for _, c := range cg.List {
			if !strings.Contains(c.Text, nocoverComment) {
				continue
			}
			commentLine := fset.Position(c.Pos()).Line
			if lbraceLine != 0 && commentLine == lbraceLine {
				return true
			}
			if commentLine == nodeLine-1 {
				return true
			}
		}
	}
	return false
}

func lbraceLineOf(fset *token.FileSet, n goast.Node) int {
	var pos token.Pos
	switch node := n.(type) {
	case *goast.FuncDecl:
		if node.Body != nil {
			pos = node.Body.Lbrace
		}
	case *goast.BlockStmt:
		pos = node.Lbrace
	case *goast.IfStmt:
		pos = node.Body.Lbrace
	case *goast.ForStmt:
		pos = node.Body.Lbrace
	case *goast.RangeStmt:
		pos = node.Body.Lbrace
	case *goast.SwitchStmt:
		pos = node.Body.Lbrace
	case *goast.TypeSwitchStmt:
		pos = node.Body.Lbrace
	case *goast.SelectStmt:
		pos = node.Body.Lbrace
	case *goast.CaseClause:
		return 0
	}
	if pos == token.NoPos {
		return 0
	}
	return fset.Position(pos).Line
}

func isErrNotNil(expr goast.Expr) bool {
	bin, ok := expr.(*goast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}

	ident, ok := bin.X.(*goast.Ident)
	if !ok || ident.Name != "err" {
		return false
	}

	nilIdent, ok := bin.Y.(*goast.Ident)
	return ok && nilIdent.Name == "nil"
}

func bodyOnlyReturns(stmts []goast.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}

	ret, ok := stmts[len(stmts)-1].(*goast.ReturnStmt)
	if !ok {
		return false
	}

	return len(stmts) == 1 || (len(stmts) == 2 && len(ret.Results) == 0)
}

// Filtering operators already covered by another option.
func effectiveStmts(fset *token.FileSet, stmts []goast.Stmt, aFile *astFile) []goast.Stmt {
	result := make([]goast.Stmt, 0, len(stmts))

	for _, stmt := range stmts {
		item := ExcludeRange{
			StartLine: fset.Position(stmt.Pos()).Line,
			EndLine:   fset.Position(stmt.End()).Line,
		}
		if !aFile.excluded(item) {
			result = append(result, stmt)
		}
	}

	return result
}

func matchesAnyPattern(fset *token.FileSet, node goast.Node, patterns []*regexp.Regexp) bool {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return false
	}

	text := buf.String()
	for _, p := range patterns {
		if p.MatchString(text) {
			return true
		}
	}

	return false
}
