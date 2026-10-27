package openapigen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"courses/platform/httpserver"
)

func errorFixture(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"internal/sample/domain/other.go": `package domain
import "errors"
type Error string
func(e Error) Error() string{return string(e)}
const ErrTyped Error = "typed"
var ErrSentinel = errors.New("sentinel")
`,
		"internal/sample/adapters/http/handler.go": `package http
import (
 "courses/platform/httpserver"
 "courses/internal/sample/domain"
 "errors"
)
` + body,
	}
	for name, source := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
func TestErrorCodeForms(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       []string
	}{
		{"literal", `func f(){httpserver.WriteError(w,r,400,"typed","")}`, []string{"typed"}},
		{"cast", `func f(){httpserver.WriteError(w,r,400,string(domain.ErrTyped),"")}`, []string{"typed"}},
		{"direct typed", `func f(){httpserver.WriteError(w,r,400,domain.ErrTyped,"")}`, []string{"typed"}},
		{"finite Error variable", `func f(){err:=errors.New("sentinel");code:=err.Error();httpserver.WriteError(w,r,400,code,"")}`, []string{"sentinel"}},
		{"sentinel variable", `func f(){err:=domain.ErrSentinel;code:=err.Error();httpserver.WriteError(w,r,400,code,"")}`, []string{"sentinel"}},
		{"error forwarding", `func f(){send(domain.ErrSentinel)};func send(err error){httpserver.WriteError(w,r,400,err.Error(),"")}`, []string{"sentinel"}},
		{"sentinel method", `func f(){httpserver.WriteError(w,r,400,domain.ErrSentinel.Error(),"")}`, []string{"sentinel"}},
		{"wrapper", `func f(){send("typed");send("sentinel")};func send(code string){httpserver.WriteError(w,r,400,code,"")}`, []string{"sentinel", "typed"}},
		{"switch", `func f(){var code string;switch{case condition:code="typed";default:code="sentinel"};httpserver.WriteError(w,r,400,code,"")}`, []string{"sentinel", "typed"}},
		{"guard", `func f(override string){switch{case override=="typed":httpserver.WriteError(w,r,400,override,"")}}`, []string{"typed"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := ScanErrorCodes(errorFixture(t, tc.body), nil)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(c.Codes, tc.want) {
				t.Fatalf("codes=%v want=%v", c.Codes, tc.want)
			}
			if len(c.Domain) != 2 || len(c.Sites) == 0 {
				t.Fatalf("vacuous census: %+v", c)
			}
		})
	}
}
func TestErrorCodeFailClosed(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"unknown function", `func f(){httpserver.WriteError(w,r,400,mystery(),"")}`},
		{"typo", `func f(){httpserver.WriteError(w,r,400,"typo","")}`},
		{"missing symbol", `func f(){httpserver.WriteError(w,r,400,string(domain.ErrMissing),"")}`},
		{"unbounded Error", `func f(err error){code:=err.Error();httpserver.WriteError(w,r,400,code,"")}`},
		{"unknown wrapper caller", `func f(){send(mystery())};func send(code string){httpserver.WriteError(w,r,400,code,"")}`},
		{"escaped writer", `func f(){emit:=httpserver.WriteError;emit(w,r,400,"typed","")}`},
		{"escaped wrapper", `func f(){callback:=send;callback("typed")};func send(code string){httpserver.WriteError(w,r,400,code,"")}`},
		{"nonexhaustive switch", `func f(){var code string;switch{case condition:code="typed"};httpserver.WriteError(w,r,400,code,"")}`},
		{"overwrite", `func f(){code:="typed";code="sentinel";httpserver.WriteError(w,r,400,code,"")}`},
		{"conditional assignment", `func f(){var code string;if condition{code="typed"};httpserver.WriteError(w,r,400,code,"")}`},
		{"fallthrough", `func f(){var code string;switch{case condition:code="typed";fallthrough;default:code="sentinel"};httpserver.WriteError(w,r,400,code,"")}`},
		{"guard mutation", `func f(override string){switch{case override=="typed":override=mystery();httpserver.WriteError(w,r,400,override,"")}}`},
		{"address escape", `func f(){code:="typed";mutate(&code);httpserver.WriteError(w,r,400,code,"")}`},
		{"raw error argument", `func f(){httpserver.WriteError(w,r,400,domain.ErrSentinel,"")}`},
		{"cycle", `func f(code string){f(code);httpserver.WriteError(w,r,400,code,"")}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if c, err := ScanErrorCodes(errorFixture(t, tc.body), nil); err == nil {
				t.Fatalf("accepted unresolved census: %+v", c)
			}
		})
	}
}
func TestTransportProvenance(t *testing.T) {
	root := errorFixture(t, `func f(){httpserver.WriteError(w,r,400,"transport","")}`)
	file := "internal/sample/adapters/http/handler.go"
	if _, err := ScanErrorCodes(root, map[string]string{"transport": file}); err != nil {
		t.Fatal(err)
	}
	for _, allow := range []map[string]string{{"transport": "wrong.go"}, {"transport": file, "unused": file}} {
		if _, err := ScanErrorCodes(root, allow); err == nil {
			t.Fatal("unused provenance accepted")
		}
	}
}
func TestTreeErrorCensus(t *testing.T) {
	census, err := ScanErrorCodes(filepath.Join("..", "..", ".."), TransportErrorCodes())
	if err != nil {
		t.Fatal(err)
	}
	if len(census.Sites) < 100 || len(census.Domain) < 100 {
		t.Fatalf("vacuous census: sites=%d domain=%d", len(census.Sites), len(census.Domain))
	}
	// Independent syntactic count: even an unresolved target must not disappear.
	count := 0
	err = filepath.WalkDir(filepath.Join("..", "..", ".."), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || d.Name() == "vendor" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				switch x := c.Fun.(type) {
				case *ast.SelectorExpr:
					if x.Sel.Name == "WriteError" {
						count++
					}
				case *ast.Ident:
					if x.Name == "WriteError" {
						count++
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != len(census.Sites) {
		t.Fatalf("coverage: found %d calls, resolved %d", count, len(census.Sites))
	}
	t.Logf("WriteError sites=%d domain codes=%d emitted=%d\n%s", len(census.Sites), len(census.Domain), len(census.Codes), strings.Join(census.Codes, ", "))
}

func TestErrorSourceDiscovery(t *testing.T) {
	root := errorFixture(t, `func f(){httpserver.WriteError(w,r,400,"typed","")}`)
	path := filepath.Join(root, "internal/sample/domain/other.go")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, extra string
		valid       bool
	}{
		{"unused domain code", `const ErrInternal Error = "internal_only"`, true},
		{"unknown initializer", `var ErrUnknown = factory()`, false},
		{"computed constant", `const ErrComputed Error = "typed" + "suffix"`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(path, append(append([]byte{}, original...), []byte("\n"+tc.extra)...), 0644); err != nil {
				t.Fatal(err)
			}
			census, err := ScanErrorCodes(root, nil)
			if !tc.valid {
				if err == nil {
					t.Fatal("unknown declaration accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(census.Domain["internal_only"]) != 1 || !reflect.DeepEqual(census.Codes, []string{"typed"}) {
				t.Fatalf("source universe leaked into emitted set: %+v", census)
			}
		})
	}
}

func TestErrorSchemaRequiresSource(t *testing.T) {
	c := ModuleContribution{Routes: []RouteBinding{{Method: "GET", Path: "/errors", Responses: map[int]Response{400: {Body: httpserver.ErrorResponse{}}}}}}
	if _, err := Generate([]ModuleContribution{c}, overlay(t, "{}")); err == nil || !strings.Contains(err.Error(), "requires resolved Go error codes") {
		t.Fatalf("missing source census accepted: %v", err)
	}
}
func TestOverlayCannotSupplyErrorEnum(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{"ErrorResponse": map[string]any{"properties": map[string]any{"error": map[string]any{"properties": map[string]any{"code": map[string]any{"type": "string", "enum": []any{"typed"}}}}}}}}}
	patch := overlay(t, `components:
  schemas:
    ErrorResponse:
      properties:
        error:
          properties:
            code:
              enum: [hidden]
`)
	if _, err := applyOverlay(doc, patch); err == nil || !strings.Contains(err.Error(), "error-code enum must come from Go") {
		t.Fatalf("overlay error enum accepted: %v", err)
	}
}

func TestChangedErrorMethodFailsClosed(t *testing.T) {
	root := errorFixture(t, `func f(){httpserver.WriteError(w,r,400,domain.ErrTyped.Error(),"")}`)
	path := filepath.Join(root, "internal/sample/domain/other.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), "return string(e)", `return "prefix:"+string(e)`, 1))
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ScanErrorCodes(root, nil); err == nil || !strings.Contains(err.Error(), "noncanonical Error method") {
		t.Fatalf("changed wire value accepted: %v", err)
	}
}
