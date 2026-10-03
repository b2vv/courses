package postgres

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestMapCourseRatingSummary(t *testing.T) {
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	id := pgtype.UUID{Bytes: [16]byte{0x50, 0, 0, 0, 0, 0, 0x40, 0, 0x80, 0, 0, 0, 0, 0, 0, 1}, Valid: true}
	for _, test := range []struct {
		name  string
		count int64
		sum   int64
		want  *float64
	}{
		{name: "below threshold is null", count: 2, sum: 9},
		{name: "visible rounded summary", count: 3, sum: 14, want: float64Pointer(4.7)},
	} {
		t.Run(test.name, func(t *testing.T) {
			course, err := mapCourse(id, "rated-course", "Rated course", "Summary", "Description", "published", pgtype.Timestamptz{Time: now, Valid: true}, pgtype.Timestamptz{Time: now, Valid: true}, pgtype.Timestamptz{Time: now, Valid: true}, pgtype.Int8{}, pgtype.Text{}, test.count, test.sum, 2)
			if err != nil {
				t.Fatal(err)
			}
			if course.LessonCount != 2 {
				t.Fatalf("lessonCount=%d want=2", course.LessonCount)
			}
			if test.want == nil {
				if course.Rating != nil {
					t.Fatalf("rating=%+v want nil", course.Rating)
				}
				return
			}
			if course.Rating == nil || course.Rating.Average != *test.want || course.Rating.Count != int(test.count) {
				t.Fatalf("rating=%+v want average=%v count=%d", course.Rating, *test.want, test.count)
			}
		})
	}
}

func float64Pointer(value float64) *float64 { return &value }
