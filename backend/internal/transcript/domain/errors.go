package domain

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrInvalidRequest Error = "invalid_request"

	// ErrForbidden is the answer when the access rule refuses, and it is the SAME answer whether
	// the lesson has a transcript or not: an absent transcript must not confirm that a course
	// the caller may not see exists.
	ErrForbidden Error = "forbidden"
	ErrNotFound  Error = "not_found"

	ErrDependencyUnavailable Error = "dependency_unavailable"
	ErrAlreadyRunning        Error = "already_running"

	// ErrSegmentLimit and ErrTextLimit are named rather than folded into invalid_request
	// because the ceilings they guard are expressible in the contract and a client that hits
	// one can act on it. B-3 is the record of what a nameless refusal costs.
	ErrSegmentLimit Error = "transcript_segment_limit"
	ErrTextLimit    Error = "transcript_text_limit"

	// ErrDocumentLimit covers both ceilings of G6: documents on one segment and documents in
	// one course library.
	ErrDocumentLimit Error = "transcript_document_limit"

	// ErrDocumentReferenced is the refusal that keeps a segment from pointing at nothing. Named
	// rather than generic because the editor can then tell the author to unlink it first.
	ErrDocumentReferenced Error = "transcript_document_referenced"

	ErrDocumentExists Error = "transcript_document_exists"

	// ErrTranscriptModified is the optimistic-lock refusal (D-9). The same shape as
	// course.ErrCourseModified, because the expectation it sets for the author is the same.
	ErrTranscriptModified Error = "transcript_modified"

	// ErrTranscriptExists is what a second writer sees when no updatedAt token exists yet,
	// which is the case for the very first write into a lesson (D-9).
	ErrTranscriptExists Error = "transcript_exists"

	// ErrDraftSuperseded is what a finishing recognition job hears when the draft it claimed is
	// no longer the draft on the row: the video was replaced while it ran, and a newer
	// generation owns the lesson now (C09).
	//
	// It is NAMED rather than folded into invalid_request because the caller must act on it,
	// and act in the direction that is not obvious: this is a SUCCESS for the task. Retrying
	// would recognise an asset nobody will read, three times, at tens of minutes a go. The same
	// reasoning makes `media` answer a vanished lesson with processDone rather than an error.
	ErrDraftSuperseded Error = "asr_draft_superseded"
)
