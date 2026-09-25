package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestEveryCanBeViewedByDelegatesToAccessCanView(t *testing.T) {
	t.Parallel()

	const accessImport = "courses/internal/contracts/access"
	const methodName = "CanBeViewedBy"
	const expectedMethods = 3
	root := filepath.Join("..", "..", "internal")
	methods := 0

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		aliases := accessImportAliases(file, accessImport)
		ast.Inspect(file, func(node ast.Node) bool {
			decl, ok := node.(*ast.FuncDecl)
			if !ok || decl.Name.Name != methodName || decl.Recv == nil {
				return true
			}
			methods++
			if decl.Body == nil || len(decl.Body.List) != 1 {
				t.Errorf("%s: %s must have exactly one statement", path, methodName)
				return true
			}
			result, ok := decl.Body.List[0].(*ast.ReturnStmt)
			if !ok || len(result.Results) != 1 || !isAccessCanViewCall(result.Results[0], aliases) {
				t.Errorf("%s: %s must return access.CanView(...)", path, methodName)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if methods != expectedMethods {
		t.Fatalf("found %d %s methods, want exactly %d", methods, methodName, expectedMethods)
	}
}

func accessImportAliases(file *ast.File, importPath string) map[string]struct{} {
	aliases := make(map[string]struct{})
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path != importPath {
			continue
		}
		if spec.Name != nil {
			if spec.Name.Name != "." && spec.Name.Name != "_" {
				aliases[spec.Name.Name] = struct{}{}
			}
			continue
		}
		aliases["access"] = struct{}{}
	}
	return aliases
}

func isAccessCanViewCall(expression ast.Expr, aliases map[string]struct{}) bool {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "CanView" {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}
	_, ok = aliases[packageName.Name]
	return ok
}
