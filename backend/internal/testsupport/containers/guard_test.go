// TEMPORARY(testcontainers-go#3868): delete with this package (backlog B-27).
package containers

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

// The retry matches testcontainers-go v0.44.0's error text. An upgrade must force the B-27 decision (is
// #3868 in? then delete this package) instead of silently disabling the match.
func TestPinnedToTheTestcontainersVersionTheRetryWasWrittenFor(t *testing.T) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Skip("no build info")
	}
	for _, dep := range info.Deps {
		if dep.Path == "github.com/testcontainers/testcontainers-go" {
			if dep.Version != "v0.44.0" {
				t.Fatalf("testcontainers-go is %s, the TEMPORARY retry was written for v0.44.0: check whether the release contains PR #3868; if so delete internal/testsupport/containers and its call sites (backlog B-27), otherwise re-verify the error text and update this pin", dep.Version)
			}
			return
		}
	}
	t.Fatal("testcontainers-go not found in build info")
}

// Start terminates the container returned with the race error. That is safe only because no call site asks
// testcontainers to reuse a shared container: with Reuse/WithReuseByName the returned object could be a
// container other tests use. This guard fails if one appears.
func TestNoCallSiteUsesContainerReuse(t *testing.T) {
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.Contains(path, filepath.Join("testsupport", "containers")) {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(body), "WithReuseByName") || strings.Contains(string(body), "Reuse:") {
			t.Errorf("%s uses container reuse; containers.Start must not wrap it (it terminates the returned container)", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
