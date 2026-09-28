package architecture_test

import (
	"testing"
	"time"

	"courses/internal/contracts/access/accesstest"
	coursedomain "courses/internal/course/domain"
	discussiondomain "courses/internal/discussion/domain"
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

	for _, test := range accesstest.TruthTable {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			actor := stranger
			administrator := test.Role == accesstest.Administrator
			if test.Role == accesstest.Owner {
				actor = owner
			}
			role := coursedomain.ActorStudent
			if administrator {
				role = coursedomain.ActorAdmin
			}
			viaCourse := coursedomain.LearningCourse{
				Course:       coursedomain.Course{Status: coursedomain.Status(test.Status)},
				InstructorID: owner,
			}.CanBeViewedBy(coursedomain.Actor{UserID: actor, Role: role})

			viaPlayback := playbackdomain.VideoCourse{
				OwnerID: owner,
				Status:  playbackdomain.CourseStatus(test.Status),
			}.CanBeViewedBy(playbackdomain.Actor{UserID: actor, Administrator: administrator})

			viaTranscript := transcriptdomain.Course{
				OwnerID: owner,
				Status:  transcriptdomain.Status(test.Status),
			}.CanBeViewedBy(transcriptdomain.Actor{UserID: actor, Administrator: administrator})

			viaDiscussion := discussiondomain.Course{
				OwnerID: owner,
				Status:  discussiondomain.Status(test.Status),
			}.CanBeViewedBy(discussiondomain.Actor{UserID: actor, Administrator: administrator})

			if viaCourse != test.Want || viaPlayback != test.Want || viaTranscript != test.Want || viaDiscussion != test.Want {
				t.Fatalf("course=%v playback=%v transcript=%v discussion=%v want=%v: the copies of the access rule disagree, or one of them changed alone",
					viaCourse, viaPlayback, viaTranscript, viaDiscussion, test.Want)
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

			positionMs := int64(test.position * 1000)
			_, discussionErr := discussiondomain.NewComment(
				discussiondomain.Comment{
					LessonID: "lesson", AuthorID: "author", Body: "comment",
					PositionMs: &positionMs,
				}, nil, test.duration,
			)
			discussionAccepted := discussionErr == nil
			if test.duration == nil {
				if discussionAccepted {
					t.Fatal("discussion accepted a position without a known duration")
				}
			} else if discussionAccepted != (viaCourse == test.position) {
				t.Fatalf("discussion accepted=%v course read=%v: position bounds disagree", discussionAccepted, viaCourse)
			}

			if viaCourse != test.want || stored.PositionSeconds != test.want {
				t.Fatalf("course read=%v playback write=%v want=%v: the two copies of the position bound disagree. B07 was this defect with the read side missing the bound entirely",
					viaCourse, stored.PositionSeconds, test.want)
			}
		})
	}
}
