package ast

import (
	goast "go/ast"
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
