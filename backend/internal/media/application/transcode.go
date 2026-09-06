package application

import "courses/internal/media/domain"

// TranscodeOutcome is the job's business result. An error returned beside it
// is reserved for dependency faults, which are always retryable.
type TranscodeOutcome struct {
	Failed    bool
	Reason    domain.FailureReason
	Permanent bool
}
