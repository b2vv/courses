package architecture_test

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"
)

func TestSafeRequestLogsDoNotIncludeURI(t *testing.T) {
	path := filepath.Join("..", "httpserver", "server.go")
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isHTTPrequestLogCall(call) {
			return true
		}
		calls++
		assertSafeGoRequestLog(t, call)
		return true
	})
	if calls != 1 {
		t.Fatalf("expected exactly one http request log call, found %d", calls)
	}
}

func isHTTPrequestLogCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Info" || len(call.Args) == 0 {
		return false
	}
	message, ok := stringLiteral(call.Args[0])
	return ok && message == "http request"
}

func assertSafeGoRequestLog(t *testing.T, call *ast.CallExpr) {
	t.Helper()
	expected := map[string]string{
		"method":     "method",
		"route":      "route",
		"status":     "response.status",
		"durationMs": "time.Since(start).Milliseconds()",
		"requestId":  "id",
		"traceId":    "span.SpanContext().TraceID().String()",
	}
	seen := make(map[string]bool)
	if (len(call.Args)-1)%2 != 0 {
		t.Fatalf("Go http request logger must contain key/value pairs, got %d trailing arguments", len(call.Args)-1)
	}
	for index := 1; index+1 < len(call.Args); index += 2 {
		key, literal := stringLiteral(call.Args[index])
		if !literal {
			t.Errorf("Go http request logger keys must be string literals")
			continue
		}
		expectedValue, approved := expected[key]
		if !approved {
			t.Errorf("Go http request logger contains unapproved field %q; review it for URI data before adding it to the safe allowlist", key)
			continue
		}
		actualValue := formatExpression(t, call.Args[index+1])
		if actualValue != expectedValue {
			t.Errorf("Go http request logger field %q must use %s, got %s", key, expectedValue, actualValue)
		}
		seen[key] = true
	}
	for key := range expected {
		if !seen[key] {
			t.Errorf("Go http request logger is missing approved field %q", key)
		}
	}
}

func stringLiteral(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func formatExpression(t *testing.T, expression ast.Expr) string {
	t.Helper()
	var buffer bytes.Buffer
	if err := format.Node(&buffer, token.NewFileSet(), expression); err != nil {
		t.Fatal(err)
	}
	return buffer.String()
}
