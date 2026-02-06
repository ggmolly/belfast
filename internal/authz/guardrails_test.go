package authz

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestPermissionKeysAreEnforced(t *testing.T) {
	permConsts, err := parsePermissionConsts("keys.go")
	if err != nil {
		t.Fatalf("parse keys.go: %v", err)
	}
	required := KnownPermissions()
	used, err := findUsedPermissionKeys(filepath.Join("..", "api"), permConsts)
	if err != nil {
		t.Fatalf("scan internal/api: %v", err)
	}

	for key := range required {
		if !used[key] {
			t.Fatalf("permission key %q is not enforced by middleware", key)
		}
	}

	_ = permConsts
}

func parsePermissionConsts(filename string) (map[string]string, error) {
	set := token.NewFileSet()
	f, err := parser.ParseFile(set, filename, nil, 0)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if name == nil || name.Name == "" {
					continue
				}
				if len(name.Name) < 4 || name.Name[:4] != "Perm" {
					continue
				}
				if i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				value, err := strconv.Unquote(lit.Value)
				if err != nil {
					continue
				}
				result[name.Name] = value
			}
		}
	}
	return result, nil
}

func findUsedPermissionKeys(apiDir string, permConsts map[string]string) (map[string]bool, error) {
	used := map[string]bool{}
	set := token.NewFileSet()
	err := filepath.WalkDir(apiDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "middleware" {
				return true
			}
			switch sel.Sel.Name {
			case "RequirePermission", "RequirePermissionAny", "RequirePermissionSelf", "RequirePermissionForMethod":
				if len(call.Args) == 0 {
					return true
				}
				key := extractAuthzKey(call.Args[0], permConsts)
				if key != "" {
					used[key] = true
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return used, nil
}

func extractAuthzKey(expr ast.Expr, permConsts map[string]string) string {
	s, ok := expr.(*ast.SelectorExpr)
	if !ok {
		if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			value, err := strconv.Unquote(lit.Value)
			if err == nil {
				return value
			}
		}
		return ""
	}
	pkg, ok := s.X.(*ast.Ident)
	if !ok || pkg.Name != "authz" {
		return ""
	}
	return permConsts[s.Sel.Name]
}
