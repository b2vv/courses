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
		domain := strings.Contains(filepath.ToSlash(path), "/domain/")
		application := strings.Contains(filepath.ToSlash(path), "/application/")
		if !domain && !application {
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
			standard := !strings.Contains(strings.Split(name, "/")[0], ".") && !strings.HasPrefix(name, "courses/")
			allowed := standard ||
				(domain && strings.HasPrefix(name, "courses/internal/contracts/")) ||
				(application && strings.HasPrefix(name, "courses/internal/") && (strings.Contains(name, "/domain") || strings.Contains(name, "/ports")))
			if !allowed {
				t.Errorf("%s imports forbidden dependency %s", path, name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
