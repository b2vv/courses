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

func TestOrdersModuleOwnership(t *testing.T) {
	t.Parallel()

	internalRoot := filepath.Join("..", "..", "internal")
	ordersRoot := filepath.Join(internalRoot, "orders")

	// The OpenAPI collector is the one leaf composition root that registers every module's
	// contribution (openapispec/contributions.go imports each module's http OpenAPIContribution,
	// orders' included — see §16.22). It is not a business module consuming payment facts, so it
	// is exempt from the "no module imports orders" rule, which guards exactly that business coupling.
	// The exemption is scoped to the single contribution package the collector actually needs, not
	// to the whole openapispec tree (§16.22 follow-up F2): the collector reaching into any deeper
	// orders internals (domain or application) still fails.
	collectorRoot := filepath.Join(internalRoot, "openapispec")
	const ordersContribution = "courses/internal/orders/adapters/http"

	t.Run("other modules do not import orders", func(t *testing.T) {
		walkProductionImports(t, internalRoot, func(path, imported string) {
			if pathWithin(path, ordersRoot) {
				return
			}
			if pathWithin(path, collectorRoot) && imported == ordersContribution {
				return
			}
			if imported == "courses/internal/orders" || strings.HasPrefix(imported, "courses/internal/orders/") {
				t.Errorf("%s imports orders module; modules consume payment facts through their own ports and adapters", path)
			}
		})
	})

	t.Run("orders does not import another module", func(t *testing.T) {
		walkProductionImports(t, ordersRoot, func(path, imported string) {
			// The OpenAPI generator under internal/platform/openapigen (which orders uses to describe
			// its own contract, like every other module) is not another business module; it is the
			// internal-tree sibling of courses/platform/* (httpserver) which was always allowed. The
			// exemption names that one package rather than the whole internal/platform prefix (§16.22
			// follow-up F2), so pulling any other platform package in would surface for review. The
			// rule still forbids reading any OTHER business module's packages.
			if !strings.HasPrefix(imported, "courses/internal/") ||
				imported == "courses/internal/orders" || strings.HasPrefix(imported, "courses/internal/orders/") ||
				strings.HasPrefix(imported, "courses/internal/contracts/") ||
				imported == "courses/internal/platform/openapigen" ||
				strings.HasPrefix(imported, "courses/internal/platform/openapigen/") {
				return
			}
			t.Errorf("%s imports another module %s; orders owns its external reads", path, imported)
		})
	})
}

func walkProductionImports(t *testing.T, root string, inspect func(path, imported string)) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			inspect(path, imported)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func pathWithin(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
