package ast

import (
	goast "go/ast"
	"go/token"
)

func WithoutIfErrReturn() ExcludeFunc {
	return func(aFile *astFile, fset *token.FileSet, f *goast.File) error {
		goast.Inspect(f, func(n goast.Node) bool {
			if n == nil {
				return false
			}

			item := ExcludeRange{
				StartLine: fset.Position(n.Pos()).Line,
				EndLine:   fset.Position(n.End()).Line,
			}

			if aFile.excluded(item) {
				return false
			}

			if !isTrackedNode(n) {
				return true
			}

			ifStmt, ok := n.(*goast.IfStmt)
			if !ok {
				return true
			}

			if !isErrNotNil(ifStmt.Cond) || ifStmt.Else != nil || !bodyOnlyReturns(effectiveStmts(fset, ifStmt.Body.List, aFile)) {
				return true
			}

			aFile.add(item)

			return false
		})

		return nil
	}
}
