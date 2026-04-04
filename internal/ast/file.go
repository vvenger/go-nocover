package ast

import (
	goast "go/ast"
	"go/token"
	"strings"
)

type astFile struct {
	imports map[string]string
	ranges  []ExcludeRange
}

func newAstFile(f *goast.File) *astFile {
	return &astFile{
		imports: parseImports(f),
	}
}

func (f *astFile) add(r ExcludeRange) {
	if !f.excluded(r) {
		f.ranges = append(f.ranges, r)
	}
}

func (f *astFile) excluded(r ExcludeRange) bool {
	for _, v := range f.ranges {
		if r.StartLine >= v.StartLine && r.EndLine <= v.EndLine {
			return true
		}
	}
	return false
}

// Statements that are not excluded by other options.
func (f *astFile) unexcludedStmts(fset *token.FileSet, stmts []goast.Stmt) []goast.Stmt {
	result := make([]goast.Stmt, 0, len(stmts))

	for _, stmt := range stmts {
		item := ExcludeRange{
			StartLine: fset.Position(stmt.Pos()).Line,
			EndLine:   fset.Position(stmt.End()).Line,
		}
		if !f.excluded(item) {
			result = append(result, stmt)
		}
	}

	return result
}

func parseImports(f *goast.File) map[string]string {
	imports := make(map[string]string, len(f.Imports))
	for _, spec := range f.Imports {
		path := strings.Trim(spec.Path.Value, `"`)

		var name string
		if spec.Name != nil {
			name = spec.Name.Name
		} else {
			name = path
			if idx := strings.LastIndex(path, "/"); idx != -1 {
				name = path[idx+1:]
			}
		}
		imports[name] = path
	}
	return imports
}
