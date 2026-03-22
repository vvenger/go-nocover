package ast

import (
	goast "go/ast"
	"go/token"
	"regexp"
)

func WithoutIfErrReturnRegexp(patterns []*regexp.Regexp) ExcludeFunc {
	return func(afile *astFile, fset *token.FileSet, f *goast.File) error {
		return findErrRegexpRanges(afile, fset, f, patterns)
	}
}

// findErrRegexpRanges finds `if err != nil { return }` blocks where the error
// originated from a call matching one of patterns — either via the if-stmt's
// init statement or the immediately preceding statement in the same block.
func findErrRegexpRanges(aFile *astFile, fset *token.FileSet, f *goast.File, patterns []*regexp.Regexp) error {
	goast.Inspect(f, func(n goast.Node) bool {
		if n == nil {
			return false
		}

		if aFile.excluded(ExcludeRange{
			StartLine: fset.Position(n.Pos()).Line,
			EndLine:   fset.Position(n.End()).Line,
		}) {
			return false
		}

		block, ok := n.(*goast.BlockStmt)
		if !ok {
			return true
		}

		for i, stmt := range block.List {
			ifStmt, ok := stmt.(*goast.IfStmt)
			if !ok {
				continue
			}

			item := ExcludeRange{
				StartLine: fset.Position(ifStmt.Pos()).Line,
				EndLine:   fset.Position(ifStmt.End()).Line,
			}

			if aFile.excluded(item) {
				continue
			}

			if !isErrNotNil(ifStmt.Cond) || ifStmt.Else != nil || !bodyOnlyReturns(effectiveStmts(fset, ifStmt.Body.List, aFile)) {
				continue
			}

			matched := false

			if ifStmt.Init != nil {
				if call := extractCallFromStmt(ifStmt.Init); call != nil {
					if matchesAnyPattern(fset, call, patterns) {
						matched = true
					}
				}
			}

			if !matched && i > 0 {
				if call := extractCallFromStmt(block.List[i-1]); call != nil {
					if matchesAnyPattern(fset, call, patterns) {
						matched = true
					}
				}
			}

			if matched {
				aFile.add(item)
			}
		}

		return true
	})

	return nil
}

func extractCallFromStmt(stmt goast.Stmt) *goast.CallExpr {
	switch s := stmt.(type) {
	case *goast.AssignStmt:
		if !assignsErr(s.Lhs) {
			return nil
		}
		if len(s.Rhs) == 1 {
			call, _ := s.Rhs[0].(*goast.CallExpr)
			return call
		}
	case *goast.ExprStmt:
		call, _ := s.X.(*goast.CallExpr)
		return call
	}
	return nil
}

func assignsErr(exprs []goast.Expr) bool {
	for _, e := range exprs {
		if ident, ok := e.(*goast.Ident); ok && ident.Name == "err" {
			return true
		}
	}
	return false
}
