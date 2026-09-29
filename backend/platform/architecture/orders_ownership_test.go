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

	t.Run("other modules do not import orders", func(t *testing.T) {
		walkProductionImports(t, internalRoot, func(path, imported string) {
			if pathWithin(path, ordersRoot) {
				return
			}
			if imported == "courses/internal/orders" || strings.HasPrefix(imported, "courses/internal/orders/") {
				t.Errorf("%s imports orders module; modules consume payment facts through their own ports and adapters", path)
			}
		})
	})

	t.Run("orders does not import another module", func(t *testing.T) {
		walkProductionImports(t, ordersRoot, func(path, imported string) {
			if !strings.HasPrefix(imported, "courses/internal/") ||
				imported == "courses/internal/orders" || strings.HasPrefix(imported, "courses/internal/orders/") ||
				strings.HasPrefix(imported, "courses/internal/contracts/") {
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
