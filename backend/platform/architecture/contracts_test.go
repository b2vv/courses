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

func TestContractsImportOnlyTheStandardLibrary(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "contracts")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		source, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range source.Imports {
			name, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if !isStandardLibraryImport(name) {
				t.Errorf("%s imports non-standard package %s", path, name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func isStandardLibraryImport(name string) bool {
	return !strings.Contains(strings.Split(name, "/")[0], ".") && !strings.HasPrefix(name, "courses/")
}
