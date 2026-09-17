package domain

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrInvalidRequest Error = "invalid_request"

	// ErrSegmentLimit and ErrTextLimit are named rather than folded into invalid_request
	// because the ceilings they guard are expressible in the contract and a client that hits
	// one can act on it. B-3 is the record of what a nameless refusal costs.
	ErrSegmentLimit Error = "transcript_segment_limit"
	ErrTextLimit    Error = "transcript_text_limit"

	// ErrDocumentLimit covers both ceilings of G6: documents on one segment and documents in
	// one course library.
	ErrDocumentLimit Error = "transcript_document_limit"

	// ErrTranscriptModified is the optimistic-lock refusal (D-9). The same shape as
	// course.ErrCourseModified, because the expectation it sets for the author is the same.
	ErrTranscriptModified Error = "transcript_modified"

	// ErrTranscriptExists is what a second writer sees when no updatedAt token exists yet,
	// which is the case for the very first write into a lesson (D-9).
	ErrTranscriptExists Error = "transcript_exists"
)
