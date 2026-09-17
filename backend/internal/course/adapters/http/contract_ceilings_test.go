package http_test

import (
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"courses/internal/course/domain"
	transcriptdomain "courses/internal/transcript/domain"
)

// Backlog B-3 left the course-wide lesson ceiling out of the contract, because a sum over
// modules[*].lessons is inexpressible in JSON Schema. Every other ceiling here IS expressible,
// so the contract carries it and the schema does the refusing - which means the same number
// exists in two places by design, and the reviewer of that work pointed out that a design
// coincidence nobody pins is B-3 again, only worse: the two numbers agree on the day they are
// written and diverge silently when one of them is edited.
//
// The divergence is not hypothetical, and it runs in the direction the original defect did not.
// Raising domain.MaximumLessons to 200 would leave every per-module maxItems at 120, so a course
// the domain accepts would be refused by the contract before reaching it - a valid course made
// unreachable by a stale schema, rather than an invalid one promised by a loose schema.
//
// So this test pins equality rather than restating the numbers: each row names where the number
// lives in the contract and which domain constant owns it. A number that moves in the domain and
// not in openapi.yaml fails here by name.
func TestContractCeilingsEqualTheDomainConstants(t *testing.T) {
	t.Parallel()

	document, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "..", "..", "openapi", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	arrayCeiling := func(t *testing.T, schema, property string) uint64 {
		t.Helper()
		definition, ok := document.Components.Schemas[schema]
		if !ok {
			t.Fatalf("schema %s is missing from the contract", schema)
		}
		field, ok := definition.Value.Properties[property]
		if !ok {
			t.Fatalf("%s has no property %s", schema, property)
		}
		if field.Value.MaxItems == nil {
			t.Fatalf("%s.%s declares no maxItems, so the contract stopped enforcing a ceiling the domain believes it enforces", schema, property)
		}
		return *field.Value.MaxItems
	}

	for _, test := range []struct {
		schema   string
		property string
		want     int
		owner    string
	}{
		{schema: "AdminModule", property: "lessons", want: domain.MaximumLessons, owner: "domain.MaximumLessons"},
		{schema: "LearningModule", property: "lessons", want: domain.MaximumLessons, owner: "domain.MaximumLessons"},
		{schema: "ProgramModule", property: "lessons", want: domain.MaximumLessons, owner: "domain.MaximumLessons"},
		{schema: "AdminCourseTree", property: "modules", want: domain.MaximumModules, owner: "domain.MaximumModules"},
		{schema: "AdminCourseSaveRequest", property: "modules", want: domain.MaximumModules, owner: "domain.MaximumModules"},
		{schema: "LearningCourse", property: "modules", want: domain.MaximumModules, owner: "domain.MaximumModules"},
		{schema: "CourseDetail", property: "program", want: domain.MaximumModules, owner: "domain.MaximumModules"},
		// The transcript ceilings arrive with the phase that owns them. They are the same shape
		// as the ones above - a flat array in one schema - unlike the course-wide lesson total,
		// which is a sum across nested arrays and therefore inexpressible here.
		{schema: "LessonTranscript", property: "segments", want: transcriptdomain.MaximumSegments, owner: "transcriptdomain.MaximumSegments"},
		{schema: "TranscriptSegment", property: "documents", want: transcriptdomain.MaximumSegmentDocuments, owner: "transcriptdomain.MaximumSegmentDocuments"},
	} {
		t.Run(test.schema+"."+test.property, func(t *testing.T) {
			t.Parallel()
			if got := arrayCeiling(t, test.schema, test.property); got != uint64(test.want) {
				t.Fatalf("%s.%s maxItems=%d, %s=%d: the contract and the domain disagree about the same ceiling", test.schema, test.property, got, test.owner, test.want)
			}
		})
	}

	t.Run("AdminLesson.description", func(t *testing.T) {
		t.Parallel()
		field, ok := document.Components.Schemas["AdminLesson"].Value.Properties["description"]
		if !ok {
			t.Fatal("AdminLesson has no description property")
		}
		if field.Value.MaxLength == nil {
			t.Fatal("AdminLesson.description declares no maxLength")
		}
		if got := *field.Value.MaxLength; got != uint64(domain.MaximumLessonDescription) {
			t.Fatalf("AdminLesson.description maxLength=%d, domain.MaximumLessonDescription=%d", got, domain.MaximumLessonDescription)
		}
	})

	// TranscriptSegment.text is counted in RUNES on both sides: OpenAPI maxLength counts
	// characters, and so does the domain constant. The byte budgets derived from it (the request
	// body limit, and nginx's client_max_body_size above it) are a different unit and are pinned
	// elsewhere — OpenAPI cannot express bytes at all.
	t.Run("TranscriptSegment.text", func(t *testing.T) {
		t.Parallel()
		field, ok := document.Components.Schemas["TranscriptSegment"].Value.Properties["text"]
		if !ok {
			t.Fatal("TranscriptSegment has no text property")
		}
		if field.Value.MaxLength == nil {
			t.Fatal("TranscriptSegment.text declares no maxLength")
		}
		if got := *field.Value.MaxLength; got != uint64(transcriptdomain.MaximumTextRunes) {
			t.Fatalf("TranscriptSegment.text maxLength=%d, transcriptdomain.MaximumTextRunes=%d: the contract and the domain disagree about the same ceiling", got, transcriptdomain.MaximumTextRunes)
		}
	})

	// The one ceiling with nothing to pin, stated so that its absence reads as a decision. A
	// course-wide lesson total cannot be expressed here (B-3), so the contract carries no number
	// for it and the domain refuses a breach with lesson_limit.
	t.Run("the course-wide lesson total is owned by the domain alone", func(t *testing.T) {
		t.Parallel()
		for _, schema := range []string{"AdminCourseSaveRequest", "AdminCourseTree"} {
			definition := document.Components.Schemas[schema].Value
			if definition.Extensions["x-course-lesson-total"] != nil {
				t.Fatalf("%s grew a second declaration of the course-wide total; it belongs to domain.MaximumLessons alone", schema)
			}
		}
	})
}
