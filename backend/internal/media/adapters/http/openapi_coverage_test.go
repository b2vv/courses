package http

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"testing"
)

// Fail closed on changes to the registration grammar as well as missing routes.
// No separately maintained expected route list can hide a new mux.Handle call.
func TestOpenAPIContributionCoversRegisterAdmin(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "handler.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	wired := map[string]bool{}
	found := false
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "RegisterAdmin" {
			continue
		}
		if found || fn.Body == nil {
			t.Fatal("ambiguous RegisterAdmin")
		}
		found = true
		for _, stmt := range fn.Body.List {
			expression, ok := stmt.(*ast.ExprStmt)
			if !ok {
				t.Fatal("unsupported registration statement")
			}
			call, ok := expression.X.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 {
				t.Fatal("unsupported registration call")
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
				t.Fatal("route must be a string literal")
			}
			pattern, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			if wired[pattern] {
				t.Fatalf("duplicate wired route %s", pattern)
			}
			wired[pattern] = true
		}
	}
	if !found || len(wired) == 0 {
		t.Fatal("no RegisterAdmin routes found")
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
