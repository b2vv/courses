package http

import (
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"courses/internal/media/domain"
)

func TestFailureReasonConstSetCensus(t *testing.T) {
	fset := token.NewFileSet()
	file := &ast.File{Name: ast.NewIdent("domain")}
	sources, err := parser.ParseDir(fset, "../../domain", func(info os.FileInfo) bool { return !strings.HasSuffix(info.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range sources["domain"].Files {
		for _, decl := range source.Decls {
			group, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			selected := &ast.GenDecl{Tok: group.Tok}
			for _, spec := range group.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					// Match the course census: typecheck scalar declarations and all
					// constants, including aliases, conversions and implicit types.
					if _, ok := spec.Type.(*ast.Ident); ok {
						selected.Specs = append(selected.Specs, spec)
					}
				case *ast.ValueSpec:
					if group.Tok == token.CONST {
						selected.Specs = append(selected.Specs, spec)
					}
				}
			}
			if len(selected.Specs) > 0 {
				file.Decls = append(file.Decls, selected)
			}
		}
	}
	pkg, err := new(types.Config).Check("domain", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatal(err)
	}
	scope := pkg.Scope()
	var declared []string
	for _, name := range scope.Names() {
		c, ok := scope.Lookup(name).(*types.Const)
		if ok && types.Identical(c.Type(), scope.Lookup("FailureReason").Type()) {
			declared = append(declared, constant.StringVal(c.Val()))
		}
	}
	slices.Sort(declared)
	declared = slices.Compact(declared)
	if len(declared) == 0 {
		t.Fatal("no FailureReason constants found")
	}
	values := make([]string, len(domain.FailureReasons))
	for i, reason := range domain.FailureReasons {
		values[i] = string(reason)
	}
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	// Exact equality also rejects extra values and duplicate set entries.
	if !slices.Equal(sorted, declared) {
		t.Fatalf("FailureReasons = %v; declared constant values = %v", sorted, declared)
	}
	emitted := OpenAPIContribution().Enums[reflect.TypeFor[domain.FailureReason]()]
	if !slices.Equal(emitted, values) {
		t.Fatalf("contribution enum = %v; ordered FailureReasons = %v", emitted, values)
	}
}
