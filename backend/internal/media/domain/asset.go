package domain

import (
	"encoding/hex"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MinimumDeclaredBytes      int64 = 1
	MaximumDeclaredBytes      int64 = 2 * 1024 * 1024 * 1024
	MaximumOriginalNameLength       = 255
)

type State string

const (
	Created    State = "created"
	Uploading  State = "uploading"
	Queued     State = "queued"
	Processing State = "processing"
	Ready      State = "ready"
	Failed     State = "failed"
)

// States is the ordered set used by transition validation and the HTTP contract.
var States = []State{Created, Uploading, Queued, Processing, Ready, Failed}

type ClaimToken string

type Generation int32

type FailureReason string

const (
	UnsupportedSource  FailureReason = "unsupported_source"
	ProbeFailed        FailureReason = "probe_failed"
	RenderFailed       FailureReason = "render_failed"
	PublishFailed      FailureReason = "publish_failed"
	StorageUnavailable FailureReason = "storage_unavailable"
)

func (reason FailureReason) Permanent() bool {
	return reason == UnsupportedSource
}

type VideoAsset struct {
	ID            string
	LessonID      string
	Generation    Generation
	State         State
	ClaimToken    ClaimToken
	DeclaredBytes int64
	ReceivedBytes int64
	OriginalName  string
	Attempt       int
	FailureReason FailureReason
	Renditions    []Rendition
	ExpiresAt     time.Time
	EvictedAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewAsset(asset VideoAsset) (VideoAsset, error) {
	if err := ValidateTicket(asset.DeclaredBytes, asset.OriginalName); err != nil {
		return VideoAsset{}, err
	}
	if !validID(asset.ID) || !validID(asset.LessonID) || asset.Generation < 1 || asset.State != Created ||
		asset.ClaimToken != "" || asset.ReceivedBytes != 0 || asset.Attempt != 0 || asset.FailureReason != "" ||
		len(asset.Renditions) != 0 || asset.ExpiresAt.IsZero() || asset.CreatedAt.IsZero() ||
		!asset.ExpiresAt.After(asset.CreatedAt) || !asset.UpdatedAt.Equal(asset.CreatedAt) || asset.EvictedAt != nil {
		return VideoAsset{}, ErrInvalidRequest
	}
	return asset, nil
}

func ValidateTicket(declaredBytes int64, originalName string) error {
	if declaredBytes > MaximumDeclaredBytes {
		return ErrUploadTooLarge
	}
	if declaredBytes < MinimumDeclaredBytes || !validOriginalName(originalName) {
		return ErrInvalidRequest
	}
	return nil
}

type Transition string

const (
	T1 Transition = "T1"
	T2 Transition = "T2"
	T3 Transition = "T3"
	T4 Transition = "T4"
	T5 Transition = "T5"
	T6 Transition = "T6"
	T7 Transition = "T7"
	T8 Transition = "T8"
	T9 Transition = "T9"
)

type TransitionWriter string

const (
	API           TransitionWriter = "api"
	UploadService TransitionWriter = "upload_service"
	Worker        TransitionWriter = "worker"
)

func (transition Transition) Writers() []TransitionWriter {
	switch transition {
	case T1, T9:
		return []TransitionWriter{API}
	case T2, T3:
		return []TransitionWriter{UploadService}
	case T4, T5, T6, T7:
		return []TransitionWriter{Worker}
	case T8:
		return []TransitionWriter{Worker, UploadService}
	default:
		return nil
	}
}

func ValidateTransition(transition Transition, from, to State) error {
	if !slices.Contains(States, to) || (from != "" && !slices.Contains(States, from)) {
		return ErrInvalidStateTransition
	}
	legal := false
	switch transition {
	case T1:
		legal = from == "" && to == Created
	case T2:
		legal = from == Created && to == Uploading
	case T3:
		legal = from == Uploading && to == Queued
	case T4:
		legal = from == Queued && to == Processing
	case T5:
		legal = from == Processing && to == Ready
	case T6, T7:
		legal = from == Processing && to == Queued
	case T8:
		legal = (from == Created || from == Uploading || from == Processing) && to == Failed
	case T9:
		legal = (from == Failed || from == Queued || from == Processing) && to == Queued
	}
	if !legal {
		return ErrInvalidStateTransition
	}
	return nil
}

type StorageKind string

const (
	StorageSource    StorageKind = "source"
	StorageHLS       StorageKind = "hls"
	StorageThumbnail StorageKind = "thumbnails"
	StorageTemporary StorageKind = "temporary"
)

type StorageKey struct {
	Kind StorageKind
	ID   string
	Leaf string
}

func validID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16 && value != "00000000-0000-0000-0000-000000000000"
}

func validOriginalName(value string) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) && utf8.RuneCountInString(value) >= 1 && utf8.RuneCountInString(value) <= MaximumOriginalNameLength
}
