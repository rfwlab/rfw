package server

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func componentNamesForTemplate(templatePath string) []string {
	abs := templatePath
	if v, err := filepath.Abs(templatePath); err == nil {
		abs = v
	}
	dir := filepath.Dir(abs)
	if filepath.Base(dir) == "templates" {
		dir = filepath.Dir(dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	rel, err := filepath.Rel(dir, abs)
	if err != nil {
		rel = filepath.Base(abs)
	}
	rel = filepath.ToSlash(rel)
	var names []string
	seen := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		goFile := filepath.Join(dir, entry.Name())
		fileNames := embedVariablesForTemplate(goFile, rel)
		if len(fileNames) == 0 {
			continue
		}
		comps := componentsUsingTemplates(goFile, fileNames)
		for _, name := range comps {
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

func embedVariablesForTemplate(goFile, rel string) map[string]struct{} {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, goFile, nil, parser.ParseComments)
	if err != nil {
		return nil
	}
	vars := map[string]struct{}{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if !specEmbedsPath(rel, gen.Doc, vs.Doc, vs.Comment) {
				continue
			}
			for _, name := range vs.Names {
				vars[name.Name] = struct{}{}
			}
		}
	}
	return vars
}

func specEmbedsPath(rel string, groups ...*ast.CommentGroup) bool {
	for _, grp := range groups {
		if grp == nil {
			continue
		}
		for _, comment := range grp.List {
			text := strings.TrimSpace(comment.Text)
			if !strings.HasPrefix(text, "//go:embed") {
				continue
			}
			fields := strings.Fields(strings.TrimPrefix(text, "//go:embed"))
			for _, field := range fields {
				candidate := strings.Trim(field, "`\"")
				if candidate == "" {
					continue
				}
				if filepath.ToSlash(candidate) == rel {
					return true
				}
			}
		}
	}
	return false
}

func componentsUsingTemplates(goFile string, vars map[string]struct{}) []string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, goFile, nil, 0)
	if err != nil {
		return nil
	}
	var names []string
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}
		fun := call.Fun
		switch f := fun.(type) {
		case *ast.IndexExpr:
			fun = f.X
		case *ast.IndexListExpr:
			fun = f.X
		}
		sel, ok := fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "core" {
			return true
		}
		if sel.Sel.Name != "NewComponent" && sel.Sel.Name != "NewComponentWith" {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		name, err := strconv.Unquote(lit.Value)
		if err != nil || name == "" {
			return true
		}
		ident, ok := call.Args[1].(*ast.Ident)
		if !ok {
			return true
		}
		if _, ok := vars[ident.Name]; !ok {
			return true
		}
		names = append(names, name)
		return true
	})
	return names
}
