package architecture_test

import (
	"testing"
	"time"

	coursedomain "courses/internal/course/domain"
	playbackdomain "courses/internal/playback/domain"
	transcriptdomain "courses/internal/transcript/domain"
)

// The table below checks the mapping from each module's own actor/course shapes to the single
// access.CanView contract. `playback.Actor` has `Administrator bool` where `course.Actor` has
// `Role ActorRole`, and the owner is `OwnerID` on one side and `InstructorID` on the other; this
// keeps those mappings aligned without re-testing three local implementations of the rule.

func TestTheAccessPredicateAgreesAcrossModules(t *testing.T) {
	t.Parallel()

	const owner = "40000000-0000-4000-8000-000000000002"
	const stranger = "40000000-0000-4000-8000-000000000009"

	for _, test := range []struct {
		name          string
		actor         string
		administrator bool
		courseOwner   string
		status        string
		want          bool
	}{
		{name: "published, stranger", actor: stranger, courseOwner: owner, status: "published", want: true},
		{name: "published, owner", actor: owner, courseOwner: owner, status: "published", want: true},
		{name: "published, admin", actor: stranger, administrator: true, courseOwner: owner, status: "published", want: true},
		{name: "draft, stranger", actor: stranger, courseOwner: owner, status: "draft", want: false},
		{name: "draft, owner", actor: owner, courseOwner: owner, status: "draft", want: true},
		{name: "draft, admin", actor: stranger, administrator: true, courseOwner: owner, status: "draft", want: true},
		{name: "archived, stranger", actor: stranger, courseOwner: owner, status: "archived", want: false},
		{name: "archived, owner", actor: owner, courseOwner: owner, status: "archived", want: true},
		{name: "archived, admin", actor: stranger, administrator: true, courseOwner: owner, status: "archived", want: true},
		// An empty actor is what an unauthenticated caller reduces to before the predicate is
		// reached; neither copy may let it through on an unpublished course.
		{name: "draft, anonymous", actor: "", courseOwner: owner, status: "draft", want: false},
		{name: "published, anonymous", actor: "", courseOwner: owner, status: "published", want: true},
		// Earlier, safety lived only in the schema because strengthening one copy meant it
		// diverged from the others; now the rule is one place and protects this case (ADR 013).
		{name: "draft, empty owner and empty actor", actor: "", courseOwner: "", status: "draft", want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			role := coursedomain.ActorStudent
			if test.administrator {
				role = coursedomain.ActorAdmin
			}
			viaCourse := coursedomain.LearningCourse{
				Course:       coursedomain.Course{Status: coursedomain.Status(test.status)},
				InstructorID: test.courseOwner,
			}.CanBeViewedBy(coursedomain.Actor{UserID: test.actor, Role: role})

			viaPlayback := playbackdomain.VideoCourse{
				OwnerID: test.courseOwner,
				Status:  playbackdomain.CourseStatus(test.status),
			}.CanBeViewedBy(playbackdomain.Actor{UserID: test.actor, Administrator: test.administrator})

			viaTranscript := transcriptdomain.Course{
				OwnerID: test.courseOwner,
				Status:  transcriptdomain.Status(test.status),
			}.CanBeViewedBy(transcriptdomain.Actor{UserID: test.actor, Administrator: test.administrator})

			if viaCourse != test.want || viaPlayback != test.want || viaTranscript != test.want {
				t.Fatalf("course=%v playback=%v transcript=%v want=%v: the copies of the access rule disagree, or one of them changed alone",
					viaCourse, viaPlayback, viaTranscript, test.want)
			}
		})
	}
}

func TestThePositionBoundAgreesAcrossModules(t *testing.T) {
	t.Parallel()

	duration := func(seconds float64) *float64 { return &seconds }

	for _, test := range []struct {
		name     string
		position float64
		duration *float64
		want     float64
	}{
		{name: "inside the asset", position: 40, duration: duration(100), want: 40},
		{name: "exactly at the end", position: 100, duration: duration(100), want: 100},
		{name: "past a shorter replacement", position: 2050, duration: duration(480), want: 480},
		{name: "no duration known", position: 2050, duration: nil, want: 2050},
		{name: "zero-length asset", position: 12, duration: duration(0), want: 0},
		{name: "position zero", position: 0, duration: duration(100), want: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// The read boundary: course bounds what it hands back, and only for a ready asset.
			viaCourse := coursedomain.NewLearningLesson(coursedomain.LearningLesson{
				VideoState: coursedomain.VideoReady,
				Position:   test.position,
			}, test.duration).Position

			// The write boundary: playback bounds what it stores. A seek is event-triggered, so
			// the update is not discarded for moving backwards, which is what lets the bound show.
			update, err := playbackdomain.NewProgressUpdate(test.position, 2, "seek")
			if err != nil {
				t.Fatal(err)
			}
			stored, _ := playbackdomain.ApplyProgress(
				playbackdomain.Progress{Present: true, PositionSeconds: 0, ClientSequence: 1},
				update, test.duration, time.Now().UTC(),
			)

			if viaCourse != test.want || stored.PositionSeconds != test.want {
				t.Fatalf("course read=%v playback write=%v want=%v: the two copies of the position bound disagree. B07 was this defect with the read side missing the bound entirely",
					viaCourse, stored.PositionSeconds, test.want)
			}
		})
	}
}
