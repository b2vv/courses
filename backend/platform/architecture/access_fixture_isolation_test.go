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

// The access truth table in internal/contracts/access/accesstest is a test
// fixture: it exists so module tests agree on one table of role x status
// outcomes. Production code must never reach for it — a fixture on the
// production path would let a test double stand in for the real access rule.
// This walks every non-test Go file in the module and refuses the import,
// narrowing by the predicate "does this production file import the fixture?"
// rather than asserting the fixture appears at any named place.
func TestAccessFixtureIsNotImportedByProductionCode(t *testing.T) {
	t.Parallel()

	const fixtureImport = "courses/internal/contracts/access/accesstest"
	root := filepath.Join("..", "..")

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
			name, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if name == fixtureImport {
				t.Errorf("%s imports the access test fixture %s; production code must not depend on it", path, fixtureImport)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
