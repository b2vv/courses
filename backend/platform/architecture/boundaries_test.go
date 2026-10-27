package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDomainAndApplicationDependenciesPointInward(t *testing.T) {
	root := filepath.Join("..", "..", "internal")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		layer := layerOf(path)
		if layer == "" {
			return nil
		}
		module := moduleOf(path)
		source, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range source.Imports {
			name, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if !importAllowed(module, layer, name) {
				t.Errorf("%s (module %q, %s) imports forbidden dependency %s", path, module, layer, name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestImportRuleStewardsCrossModuleIsolation pins the general isolation rule (ADR 013) rather than
// relying on ad-hoc per-pair tests: a module's domain and application see only the standard library,
// the shared contracts, and — for application — their OWN module's domain and ports. Importing
// another module's packages is forbidden for every module pair, not just the pairs someone wrote a
// bespoke test for. The foreign-module case below is the one the earlier rule wrongly allowed (it
// bound the import to `courses/internal/` but not to the file's own module).
func TestImportRuleStewardsCrossModuleIsolation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		module  string
		layer   string
		path    string
		allowed bool
	}{
		{"standard library", "orders", "domain", "time", true},
		{"standard library in application", "orders", "application", "context", true},
		{"contracts from domain", "orders", "domain", "courses/internal/contracts/access", true},
		{"contracts from application", "orders", "application", "courses/internal/contracts/access", true},
		{"own module domain from application", "orders", "application", "courses/internal/orders/domain", true},
		{"own module ports from application", "orders", "application", "courses/internal/orders/ports", true},
		{"foreign module domain from application", "orders", "application", "courses/internal/discussion/domain", false},
		{"foreign module ports from application", "discussion", "application", "courses/internal/orders/ports", false},
		{"foreign module domain from domain", "orders", "domain", "courses/internal/discussion/domain", false},
		{"own module domain from domain", "orders", "domain", "courses/internal/orders/domain", false},
		{"third party from application", "orders", "application", "github.com/jackc/pgx/v5", false},
		{"adapters from application", "orders", "application", "courses/internal/orders/adapters/http", false},
		{"own module domain from ports", "orders", "ports", "courses/internal/orders/domain", true},
		{"contracts from ports", "orders", "ports", "courses/internal/contracts/access", true},
		{"foreign module domain from ports", "orders", "ports", "courses/internal/discussion/domain", false},
		{"application from ports", "orders", "ports", "courses/internal/orders/application", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := importAllowed(tc.module, tc.layer, tc.path); got != tc.allowed {
				t.Errorf("importAllowed(%q, %q, %q)=%v, want %v", tc.module, tc.layer, tc.path, got, tc.allowed)
			}
		})
	}
}

// importAllowed reports whether a file in `module`'s `layer` (domain|application) may import
// importPath. ADR 013: domain and application see the standard library and the shared contracts;
// application may additionally reach its own module's domain and ports; neither may import any other
// module's packages, nor an adapter. Binding the import to `module` — not merely to
// `courses/internal/` — is what closes the M8 cross-module guard-gap.
func importAllowed(module, layer, importPath string) bool {
	first := strings.Split(importPath, "/")[0]
	standard := !strings.Contains(first, ".") && !strings.HasPrefix(importPath, "courses/")
	if standard {
		return true
	}
	// A domain may import a contract, and so may an application (ADR 013). Contracts are the only
	// cross-cutting courses/ package either layer may see.
	if strings.HasPrefix(importPath, "courses/internal/contracts/") {
		return true
	}
	// The sole remaining allowance: an application or a ports file reaching its OWN module's domain
	// or ports (AGENTS.md: application sees domain and ports; ports may reference domain). Never
	// another module, never an adapter, never the application layer from ports.
	if layer == "application" || layer == "ports" {
		own := "courses/internal/" + module + "/"
		if strings.HasPrefix(importPath, own) && (strings.Contains(importPath, "/domain") || strings.Contains(importPath, "/ports")) {
			return true
		}
	}
	return false
}

// layerOf returns "domain", "application" or "ports" for a file on that inward layer, or "" for any
// other file (adapters, composition roots) the inward-dependency rule does not police here. This
// assumes a FLAT layer layout — no nested packages under a module's domain/application/ports — which
// holds today; the predicate forbids a same-module same-layer subpackage import, so splitting a layer
// into subpackages later would need this rule widened. The contracts package's own purity (ADR 013:
// standard library only, zero module imports) is a separate rule, not policed here.
func layerOf(path string) string {
	slash := filepath.ToSlash(path)
	switch {
	case strings.Contains(slash, "/domain/"):
		return "domain"
	case strings.Contains(slash, "/application/"):
		return "application"
	case strings.Contains(slash, "/ports/"):
		return "ports"
	default:
		return ""
	}
}

// moduleOf returns the module segment immediately after internal/ in path, e.g. "orders" for
// internal/orders/application/application.go.
func moduleOf(path string) string {
	slash := filepath.ToSlash(path)
	const marker = "/internal/"
	index := strings.Index(slash, marker)
	if index < 0 {
		return ""
	}
	rest := slash[index+len(marker):]
	if slash := strings.IndexByte(rest, '/'); slash >= 0 {
		return rest[:slash]
	}
	return rest
}
