package ast

import (
	goast "go/ast"
	"go/token"
	"regexp"
)

func WithoutLogRegexp(patterns []*regexp.Regexp) ExcludeFunc {
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

			switch node := n.(type) {
			case *goast.IfStmt:
				if node.Body != nil && bodyOnlyLogCalls(fset, node.Body, patterns) {
					aFile.add(item)
					return false
				}
			case *goast.FuncDecl:
				if node.Body != nil && bodyOnlyLogCalls(fset, node.Body, patterns) {
					aFile.add(item)
					return false
				}
			case *goast.ExprStmt:
				if matchesAnyPattern(fset, node.X, patterns) {
					aFile.add(item)
				}
				return false
			case *goast.AssignStmt:
				for _, rhs := range node.Rhs {
					if matchesAnyPattern(fset, rhs, patterns) {
						aFile.add(item)
						break
					}
				}
				return false
			}

			return true
		})

		return nil
	}
}

func bodyOnlyLogCalls(fset *token.FileSet, body *goast.BlockStmt, patterns []*regexp.Regexp) bool {
	if len(body.List) == 0 {
		return false
	}

	for _, stmt := range body.List {
		switch s := stmt.(type) {
		case *goast.ExprStmt:
			if !matchesAnyPattern(fset, s.X, patterns) {
				return false
			}
		case *goast.AssignStmt:
			matched := false
			for _, rhs := range s.Rhs {
				if matchesAnyPattern(fset, rhs, patterns) {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		default:
			return false
		}
	}
	return true
}
