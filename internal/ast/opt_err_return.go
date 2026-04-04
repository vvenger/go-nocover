package ast

import (
	goast "go/ast"
	"go/token"
)

// Finds `if err != nil { return }` blocks.
// Where `return` is a single statement.
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

			if !isErrNotNil(ifStmt.Cond) || ifStmt.Else != nil || !bodyOnlyReturns(aFile.unexcludedStmts(fset, ifStmt.Body.List)) {
				return true
			}

			aFile.add(item)

			return false
		})

		return nil
	}
}
