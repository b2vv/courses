package postgres

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"courses/internal/playback/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	progressUserID   = "40000000-0000-4000-8000-000000000001"
	progressLessonID = "50000000-0000-4000-8000-000000000002"
)

type fakeProgressTransaction struct {
	rows       []pgx.Row
	execArgs   [][]any
	execErr    error
	committed  bool
	commitErr  error
	rolledBack bool
}

func (tx *fakeProgressTransaction) QueryRow(context.Context, string, ...any) pgx.Row {
	row := tx.rows[0]
	tx.rows = tx.rows[1:]
	return row
}

func (tx *fakeProgressTransaction) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	tx.execArgs = append(tx.execArgs, args)
	return pgconn.NewCommandTag("UPDATE 1"), tx.execErr
}

func (tx *fakeProgressTransaction) Commit(context.Context) error {
	tx.committed = true
	return tx.commitErr
}

func (tx *fakeProgressTransaction) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}

type durationRow struct {
	duration pgtype.Numeric
	err      error
}

func (row durationRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	*dest[0].(*pgtype.Numeric) = row.duration
	return nil
}

type storedProgressRow struct {
	position    float64
	completedAt pgtype.Timestamptz
	sequence    int64
	updatedAt   time.Time
	err         error
}

func (row storedProgressRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	*dest[0].(*float64) = row.position
	*dest[1].(*pgtype.Timestamptz) = row.completedAt
	*dest[2].(*int64) = row.sequence
	*dest[3].(*time.Time) = row.updatedAt
	return nil
}

func TestProgressWriterUsesTheDomainPolicyInsideItsTransaction(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.September, 8, 14, 0, 0, 0, time.UTC)
	tx := &fakeProgressTransaction{rows: []pgx.Row{
		durationRow{duration: pgtype.Numeric{Int: big.NewInt(100), Valid: true}},
		storedProgressRow{position: 90, sequence: 1, updatedAt: now.Add(-time.Minute)},
	}}
	writer := newProgressWriter(func(context.Context) (progressTransaction, error) { return tx, nil })
	update, err := domain.NewProgressUpdate(95, 2, "tick")
	if err != nil {
		t.Fatal(err)
	}

	if err := writer.Save(context.Background(), progressUserID, progressLessonID, update, now); err != nil {
		t.Fatal(err)
	}
	if !tx.committed || !tx.rolledBack || len(tx.execArgs) != 2 {
		t.Fatalf("committed=%t rolledBack=%t execs=%d", tx.committed, tx.rolledBack, len(tx.execArgs))
	}
	stored := tx.execArgs[1]
	completedAt, ok := stored[3].(*time.Time)
	if len(stored) != 6 || stored[2] != 95.0 || !ok || completedAt == nil || !completedAt.Equal(now) || stored[4] != int64(2) || stored[5] != now {
		t.Fatalf("stored args=%v", stored)
	}
}

func TestProgressWriterDoesNotPersistACompletedLessonRegression(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.September, 8, 14, 0, 0, 0, time.UTC)
	tx := &fakeProgressTransaction{rows: []pgx.Row{
		durationRow{},
		storedProgressRow{position: 100, completedAt: pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true}, sequence: 3, updatedAt: now.Add(-time.Minute)},
	}}
	writer := newProgressWriter(func(context.Context) (progressTransaction, error) { return tx, nil })
	update, err := domain.NewProgressUpdate(0, 4, "nav")
	if err != nil {
		t.Fatal(err)
	}

	if err := writer.Save(context.Background(), progressUserID, progressLessonID, update, now); err != nil {
		t.Fatal(err)
	}
	if tx.committed || !tx.rolledBack || len(tx.execArgs) != 1 {
		t.Fatalf("committed=%t rolledBack=%t execs=%d", tx.committed, tx.rolledBack, len(tx.execArgs))
	}
}

func TestProgressWriterClassifiesMissingLessonAndDependencyFailure(t *testing.T) {
	t.Parallel()
	update, err := domain.NewProgressUpdate(1, 1, "tick")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("begin failure", func(t *testing.T) {
		failure := errors.New("connection failed")
		writer := newProgressWriter(func(context.Context) (progressTransaction, error) { return nil, failure })
		if err := writer.Save(context.Background(), progressUserID, progressLessonID, update, time.Now()); !errors.Is(err, domain.ErrDependencyUnavailable) {
			t.Fatalf("got %v, want dependency unavailable", err)
		}
	})
	t.Run("missing lesson", func(t *testing.T) {
		tx := &fakeProgressTransaction{rows: []pgx.Row{durationRow{err: pgx.ErrNoRows}}}
		writer := newProgressWriter(func(context.Context) (progressTransaction, error) { return tx, nil })
		if err := writer.Save(context.Background(), progressUserID, progressLessonID, update, time.Now()); !errors.Is(err, domain.ErrLessonNotFound) {
			t.Fatalf("got %v, want lesson not found", err)
		}
		if !tx.rolledBack {
			t.Fatal("missing lesson transaction was not rolled back")
		}
	})
}
