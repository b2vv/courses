package http

import (
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"slices"
	"strings"
	"testing"

	"courses/internal/course/domain"
)

// Census reads typed production constants, not a second list of their values.
// Sort's two admin-only names are deliberate exclusions; new names fail closed.
func missingConsts(scope *types.Scope, typeName string, emitted []string) []string {
	var missing []string
	for _, name := range scope.Names() {
		c, ok := scope.Lookup(name).(*types.Const)
		if !ok || c.Type() != scope.Lookup(typeName).Type() {
			continue
		}
		if typeName == "Sort" && (name == "SortStudents" || name == "SortPrice") {
			continue
		}
		if !slices.Contains(emitted, constant.StringVal(c.Val())) {
			missing = append(missing, name)
		}
	}
	return missing
}

func TestDomainConstSetCensus(t *testing.T) {
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
					// Include every named scalar so const aliases and inferred conversions
					// are typechecked too. Unsupported dependencies fail rather than disappear.
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
	for _, tc := range []struct {
		name   string
		values []string
	}{
		{"Sort", enumStrings(domain.PublicCatalogSorts)},
		{"Status", enumStrings(domain.CourseStatuses)},
		{"BlockType", enumStrings(domain.BlockTypes)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if missing := missingConsts(scope, tc.name, tc.values); len(missing) != 0 {
				t.Fatalf("domain constants absent from authoritative set: %v", missing)
			}
			// A typed declaration added without updating the set must be detected.
			name := "CensusAdded" + tc.name
			scope.Insert(types.NewConst(token.NoPos, pkg, name, scope.Lookup(tc.name).Type(), constant.MakeString("census_added")))
			if missing := missingConsts(scope, tc.name, tc.values); !slices.Equal(missing, []string{name}) {
				t.Fatalf("new const escaped census: %v", missing)
			}
		})
	}
}

func TestPublicSortSingleSource(t *testing.T) {
	original := domain.PublicCatalogSorts
	t.Cleanup(func() { domain.PublicCatalogSorts = original })
	// Changing just the authoritative set changes BOTH validation and contribution.
	domain.PublicCatalogSorts = append(slices.Clone(original), domain.Sort("future_public_sort"))
	for _, value := range domain.PublicCatalogSorts {
		if _, err := domain.NewQuery(20, 0, "", string(value)); err != nil {
			t.Fatalf("NewQuery rejected authoritative sort %q: %v", value, err)
		}
	}
	for _, route := range OpenAPIContribution().Routes {
		if route.Path != "/api/v1/courses" {
			continue
		}
		for _, p := range route.Params {
			if p.Name == "sort" {
				if !slices.Equal(p.Enum, enumStrings(domain.PublicCatalogSorts)) {
					t.Fatalf("contribution does not consume PublicCatalogSorts: %v", p.Enum)
				}
				return
			}
		}
	}
	t.Fatal("public sort parameter missing")
}
