package architecture_test

import (
	"testing"
	"time"

	coursedomain "courses/internal/course/domain"
	playbackdomain "courses/internal/playback/domain"
	transcriptdomain "courses/internal/transcript/domain"
)

// Two product rules exist twice each, once per module, and both duplications are deliberate: a
// module's domain may not import another module's domain (AGENTS.md, enforced by
// TestDomainAndApplicationDependenciesPointInward), so a shared function has nowhere to live.
// That makes the copies structural rather than sloppy - and it also means nothing holds them
// together. The position bound carries a comment in each copy saying the duplication is
// intended; a comment does not fail. The access predicate carries not even that.
//
// A textual pin is impossible here, unlike the numbers in TestContractCeilingsEqualTheDomainConstants:
// the copies do not share a shape. `playback.Actor` has `Administrator bool` where
// `course.Actor` has `Role ActorRole`, and the owner is `OwnerID` on one side and
// `InstructorID` on the other. So this pins BEHAVIOUR: one truth table per rule, every copy
// asked the same question, every copy required to answer the same way.
//
// The third module arrived: the transcript read needs the access predicate, so it carries its
// own copy and a column here. Three copies now, and the table is what keeps them honest.

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
		// Both copies answer TRUE here, and the pin records that rather than wishing otherwise:
		// `actor.UserID == course.OwnerID` is `"" == ""`, so an anonymous caller reads as the
		// owner of an ownerless course. Neither predicate defends against it, and neither has
		// to - `courses.instructor_id` is NOT NULL (migrations/00003_authoring.sql:19), so an
		// ownerless course cannot exist, and authentication runs before either call site. The
		// safety lives in the schema, not in the rule. This row is here so that hardening one
		// copy to reject an empty actor fails the pin and forces the other to match.
		{name: "draft, empty owner and empty actor: agreed, and unreachable by schema", actor: "", courseOwner: "", status: "draft", want: true},
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
