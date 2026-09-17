package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Backlog B-10 is about a counter that does not exist until the thing it counts happens: a
// dashboard and an alert are blind exactly when they are needed. Its cure - assert the series by
// name rather than by a zero value - is in place, and TestMetricsArePreinitialised proves the
// registration itself where the precondition is free.
//
// Neither of those closes the way back into the same blindness. The five bounded failure reasons
// are string literals in three places: the domain owns them, NewMetrics pre-touches each one, and
// the observability test asserts each one appears. Add a sixth reason to the domain and nothing
// fails: the counter for it is never pre-touched, so it springs into existence on the first
// failure of that kind - the exact behaviour B-10 exists to forbid, reached by a route B-10 does
// not watch.
//
// So this pins the three sets equal. It reads the files rather than importing them, because the
// question is what the source says, and because platform must not depend on a module.
// infrastructure/tests/smoke.sh holds a fourth copy and is deliberately out of scope: the Go test
// container mounts only backend/, the same limit that split R30's nginx half onto the host. Smoke
// remains a live spot check of the wired worker; the guarantee lives here and in the unit test.
func TestBoundedFailureReasonsAreRegisteredEverywhere(t *testing.T) {
	t.Parallel()

	domainReasons := constantValues(t,
		filepath.Join("..", "..", "internal", "media", "domain", "asset.go"), "FailureReason")
	registered := sliceLiteralAfter(t,
		filepath.Join("..", "..", "platform", "observability", "metrics.go"), "TranscodingFailures.WithLabelValues")
	asserted := labelValues(t,
		filepath.Join("..", "..", "platform", "observability", "metrics_test.go"),
		`courses_transcoding_failures_total{reason="`)

	if len(domainReasons) == 0 {
		t.Fatal("no FailureReason constants found; the domain moved and this test now proves nothing")
	}
	for _, pair := range []struct {
		name string
		got  []string
	}{
		{name: "NewMetrics pre-touches", got: registered},
		{name: "TestMetricsArePreinitialised asserts", got: asserted},
	} {
		if strings.Join(pair.got, ",") != strings.Join(domainReasons, ",") {
			t.Errorf("%s %v, the domain declares %v: a reason the domain knows and the metrics do not is a counter that appears on first failure",
				pair.name, pair.got, domainReasons)
		}
	}
}

// constantValues collects the string values of constants declared with the named type.
func constantValues(t *testing.T, path, typeName string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var values []string
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.CONST {
			continue
		}
		for _, spec := range general.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			identifier, ok := value.Type.(*ast.Ident)
			if !ok || identifier.Name != typeName {
				continue
			}
			for _, expression := range value.Values {
				if literal, ok := stringLiteralOf(expression); ok {
					values = append(values, literal)
				}
			}
		}
	}
	sort.Strings(values)
	return values
}

// sliceLiteralAfter finds the `for _, x := range []string{...}` whose body mentions the marker and
// returns the literal's elements. The loop and its body are matched together so that reordering the
// loops in NewMetrics cannot silently point this test at the wrong label set.
func sliceLiteralAfter(t *testing.T, path, marker string) []string {
	t.Helper()
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	source := readFile(t, path)
	var values []string
	ast.Inspect(file, func(node ast.Node) bool {
		loop, ok := node.(*ast.RangeStmt)
		if !ok {
			return true
		}
		composite, ok := loop.X.(*ast.CompositeLit)
		if !ok {
			return true
		}
		start := fileSet.Position(loop.Body.Pos()).Offset
		end := fileSet.Position(loop.Body.End()).Offset
		if !strings.Contains(source[start:end], marker) {
			return true
		}
		for _, element := range composite.Elts {
			if literal, ok := stringLiteralOf(element); ok {
				values = append(values, literal)
			}
		}
		return false
	})
	sort.Strings(values)
	return values
}

// labelValues collects the label value that follows each occurrence of the prefix.
func labelValues(t *testing.T, path, prefix string) []string {
	t.Helper()
	source := readFile(t, path)
	var values []string
	for _, fragment := range strings.Split(source, prefix)[1:] {
		end := strings.Index(fragment, `"`)
		if end < 0 {
			t.Fatalf("unterminated label value after %q", prefix)
		}
		values = append(values, fragment[:end])
	}
	sort.Strings(values)
	return values
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func stringLiteralOf(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	unquoted, err := strconv.Unquote(literal.Value)
	if err != nil {
		return "", false
	}
	return unquoted, true
}
