package domain

import (
	"math"
	"slices"
	"time"
)

const MaximumClientSequence int64 = 1<<31 - 1

type ProgressReason string

const (
	ProgressTick  ProgressReason = "tick"
	ProgressPause ProgressReason = "pause"
	ProgressSeek  ProgressReason = "seek"
	ProgressHide  ProgressReason = "hide"
	ProgressNav   ProgressReason = "nav"
	ProgressEnded ProgressReason = "ended"
)

// ProgressReasons is the ordered set used by validation and the HTTP contract.
var ProgressReasons = []ProgressReason{ProgressTick, ProgressPause, ProgressSeek, ProgressHide, ProgressNav, ProgressEnded}

type ProgressUpdate struct {
	positionSeconds float64
	clientSequence  int64
	reason          ProgressReason
}

func NewProgressUpdate(positionSeconds float64, clientSequence int64, reason string) (ProgressUpdate, error) {
	if positionSeconds < 0 || math.IsNaN(positionSeconds) || math.IsInf(positionSeconds, 0) ||
		clientSequence < 0 || clientSequence > MaximumClientSequence {
		return ProgressUpdate{}, ErrInvalidProgress
	}
	progressReason := ProgressReason(reason)
	if !slices.Contains(ProgressReasons, progressReason) {
		return ProgressUpdate{}, ErrInvalidProgress
	}
	return ProgressUpdate{positionSeconds: positionSeconds, clientSequence: clientSequence, reason: progressReason}, nil
}

func (update ProgressUpdate) PositionSeconds() float64 { return update.positionSeconds }

func (update ProgressUpdate) ClientSequence() int64 { return update.clientSequence }

func (update ProgressUpdate) Reason() ProgressReason { return update.reason }

func (update ProgressUpdate) eventTriggered() bool { return update.reason != ProgressTick }

type Progress struct {
	Present         bool
	PositionSeconds float64
	CompletedAt     *time.Time
	ClientSequence  int64
	UpdatedAt       time.Time
}

func ApplyProgress(saved Progress, update ProgressUpdate, durationSeconds *float64, now time.Time) (Progress, bool) {
	if saved.Present && update.clientSequence <= saved.ClientSequence {
		return saved, false
	}

	positionSeconds := update.positionSeconds
	// NewLearningLesson deliberately repeats this bound at the course read boundary; both use
	// only the duration of a ready asset.
	if durationSeconds != nil && positionSeconds > *durationSeconds {
		positionSeconds = *durationSeconds
	}
	if saved.Present && positionSeconds <= saved.PositionSeconds &&
		(!update.eventTriggered() || saved.CompletedAt != nil) {
		return saved, false
	}
	completedAt := saved.CompletedAt
	if completedAt == nil && durationSeconds != nil &&
		positionSeconds >= 0.95**durationSeconds && positionSeconds > saved.PositionSeconds {
		completed := now
		completedAt = &completed
	}
	return Progress{
		Present:         true,
		PositionSeconds: positionSeconds,
		CompletedAt:     completedAt,
		ClientSequence:  update.clientSequence,
		UpdatedAt:       now,
	}, true
}
