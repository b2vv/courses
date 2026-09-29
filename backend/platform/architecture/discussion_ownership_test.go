package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscussionDoesNotImportCourseModule(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "discussion")
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
		for _, imported := range file.Imports {
			if strings.Trim(imported.Path.Value, `"`) == "courses/internal/course" || strings.HasPrefix(strings.Trim(imported.Path.Value, `"`), "courses/internal/course/") {
				t.Errorf("%s imports course module; discussion owns its course read", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
