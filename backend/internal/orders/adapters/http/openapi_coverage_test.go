package http

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// Traverse registration bodies, including guarded calls, without executing middleware.
func registeredRoutes(t *testing.T, source any) map[string]bool {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "handler.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	wired := map[string]bool{}
	found := map[string]bool{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(fn.Name.Name, "Register") {
			continue
		}
		if fn.Name.Name != "Register" && fn.Name.Name != "RegisterAdmin" && fn.Name.Name != "RegisterTeach" {
			t.Fatalf("unknown registration function %s", fn.Name.Name)
		}
		if found[fn.Name.Name] || fn.Body == nil {
			t.Fatal("ambiguous registration")
		}
		found[fn.Name.Name] = true
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			expression, ok := node.(*ast.ExprStmt)
			if !ok {
				return true
			}
			call, ok := expression.X.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 {
				t.Fatal("unsupported registration statement")
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Handle" {
				t.Fatal("expected mux.Handle")
			}
			receiver, ok := selector.X.(*ast.Ident)
			if !ok || receiver.Name != "mux" {
				t.Fatal("expected mux receiver")
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				t.Fatal("route must be a literal")
			}
			pattern, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			if wired[pattern] {
				t.Fatalf("duplicate wired route %s", pattern)
			}
			wired[pattern] = true
			return false
		})
	}
	if len(found) != 3 {
		t.Fatalf("expected three registration functions; got %v", found)
	}
	return wired
}

func TestOpenAPIContributionCoversRegistrations(t *testing.T) {
	wired := registeredRoutes(t, nil)
	if len(wired) != 22 {
		t.Fatalf("expected 22 routes, got %d", len(wired))
	}
	declared := map[string]bool{}
	for _, route := range OpenAPIContribution().Routes {
		pattern := route.Method + " " + route.Path
		if declared[pattern] {
			t.Fatalf("duplicate contribution route %s", pattern)
		}
		declared[pattern] = true
	}
	if !reflect.DeepEqual(wired, declared) {
		t.Fatalf("wired=%v contribution=%v", wired, declared)
	}
}

func TestRegistrationCensusIncludesGuardedBranches(t *testing.T) {
	source := `package http
 func Register(){if enabled { mux.Handle("POST /callback", h) } else { for enabled { mux.Handle("GET /nested", h) } }}
 func RegisterAdmin(){mux.Handle("GET /admin",h)}
 func RegisterTeach(){mux.Handle("GET /teach",h)}`
	got := registeredRoutes(t, source)
	want := map[string]bool{"POST /callback": true, "GET /nested": true, "GET /admin": true, "GET /teach": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("guarded census: %v", got)
	}
	delete(got, "POST /callback")
	if reflect.DeepEqual(got, want) {
		t.Fatal("missing guarded route escaped comparison")
	}
}
