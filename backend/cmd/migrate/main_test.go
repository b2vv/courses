package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"
)

// MS1 M-R9: a migration's own RAISE EXCEPTION reaches the log through goose's
// wrapping; any other database error stays behind the generic text.
func TestAuthoredRefusalSurfacesOnlyRaiseException(t *testing.T) {
	wrap := func(cause error) error {
		partial := &goose.PartialError{
			Failed: &goose.MigrationResult{Source: &goose.Source{Type: goose.TypeSQL, Path: "00030_audio_track_language_shape.sql", Version: 30}},
			Err:    fmt.Errorf("failed to run SQL migration: %w", cause),
		}
		return executionError{cause: partial}
	}

	raised := wrap(&pgconn.PgError{Code: "P0001", Message: "audio track language shape: 1 track(s) carry a language no transcript can use; first track a4000000-0000-4000-8000-000000000001"})
	found, ok := authoredRefusal(raised)
	if !ok || found.sqlState != "P0001" || found.version != 30 || !strings.Contains(found.message, "a4000000-0000-4000-8000-000000000001") {
		t.Fatalf("authoredRefusal(raised) = %+v, %v", found, ok)
	}
	if raised.Error() != "database migration execution failed; inspect versioned SQL against the target schema" {
		t.Fatalf("the logged reason changed: %q", raised.Error())
	}

	for _, other := range []error{
		wrap(&pgconn.PgError{Code: "23514", Message: `new row for relation "audio_tracks" violates check constraint, Failing row contains (secret)`}),
		wrap(errors.New("connection reset")),
		errors.New("plain"),
	} {
		if found, ok := authoredRefusal(other); ok {
			t.Fatalf("authoredRefusal(%v) surfaced %+v", other, found)
		}
	}
}
