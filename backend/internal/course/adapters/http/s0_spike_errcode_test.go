//go:build s0spike

// S0.0 spike, part 2 (B1 + N3 from plan §13). Run with part 1:
//
//	go test -tags s0spike -run TestS0Spike -v ./internal/course/adapters/http/
//
// B1: the error-code enum cannot be built from WriteError call-site string
// literals alone — discussion wraps it in a local writeError and passes a
// VARIABLE (`code := err.Error()`) whose values are errors.New("…") consts in
// discussion/domain/errors.go. This part proves the Go source for those codes
// is statically parseable (go/parser), so Layer A can derive the enum. N3: the
// non-JSON orders callback route (w.Write(Acknowledge())) is identified so it
// can be put on a declared exclusion list.
package http

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// repoFile resolves a path relative to the backend module root, independent of
// the test's working directory (go test runs in the package dir).
func repoFile(rel string) string {
	_, self, _, _ := runtime.Caller(0)
	// self = .../backend/internal/course/adapters/http/s0_spike_errcode_test.go
	backend := filepath.Clean(filepath.Join(filepath.Dir(self), "..", "..", "..", ".."))
	return filepath.Join(backend, rel)
}

// errorsNewLiterals parses a Go file and returns every string literal passed to
// errors.New(...). This is the mechanism B1(a) requires for domain/errors.go.
func errorsNewLiterals(t *testing.T, relPath string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, repoFile(relPath), nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", relPath, err)
	}
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "New" {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "errors" || len(call.Args) != 1 {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if s, err := strconv.Unquote(lit.Value); err == nil {
			out = append(out, s)
		}
		return true
	})
	sort.Strings(out)
	return out
}

func TestS0SpikeDiscussionErrorCodeSource(t *testing.T) {
	codes := errorsNewLiterals(t, "internal/discussion/domain/errors.go")
	if len(codes) < 20 {
		t.Fatalf("expected the full discussion error-code set from domain/errors.go, got %d: %v", len(codes), codes)
	}
	// Spot-check the exact codes B1 names as unreachable from call-site scanning
	// (they flow through `code := err.Error()`): they MUST be present here.
	for _, want := range []string{"review_already_exists", "report_already_open", "course_forbidden"} {
		found := false
		for _, c := range codes {
			if c == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("B1 code %q not found by domain/errors.go parse — Layer A would miss it", want)
		}
	}
	t.Logf("B1 PROVEN: go/parser extracted %d error codes from discussion/domain/errors.go: %s",
		len(codes), strings.Join(codes, ", "))
	t.Logf("B1 FINDING: discussion handler passes `code := err.Error()` (variable) to local writeError; "+
		"call-site literal scan alone yields 0 of these %d codes. The enum MUST union domain/errors.go "+
		"consts (this parse) + direct WriteError literals + per-module writeError wrappers. "+
		"Assert-coverage test: every domain/errors.go code ∈ emitted enum.", len(codes))
}

// TestS0SpikeNonJSONCallback confirms the WayForPay callback writes raw bytes
// (not JSON via WriteJSON), so §4-A respType and Layer-B ValidateResponse must
// skip it via a DECLARED exclusion list (N3).
func TestS0SpikeNonJSONCallback(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, repoFile("internal/orders/adapters/http/handler.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse orders handler: %v", err)
	}
	var rawWrite, writeJSON int
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch sel.Sel.Name {
		case "Write": // w.Write(...)
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "w" {
				rawWrite++
			}
		case "WriteJSON":
			writeJSON++
		}
		return true
	})
	if rawWrite == 0 {
		t.Fatalf("expected ≥1 raw w.Write (the Acknowledge callback) in orders handler; got 0")
	}
	t.Logf("N3 PROVEN: orders handler has %d raw w.Write (non-JSON callback) and %d WriteJSON. "+
		"The callback route (POST /api/v1/payments/wayforpay/callback → w.Write(Acknowledge())) must be "+
		"on a DECLARED non-JSON exclusion list so Layer-B coverage-gate does not fail on it.", rawWrite, writeJSON)
}
