package openapigen

// TransportErrorCodes documents codes without a domain declaration. Values name
// the actual WriteError emission file, including the terminal site of a wrapper.
// ScanErrorCodes rejects missing, stale, redundant, and unused entries.
func TransportErrorCodes() map[string]string {
	return map[string]string{
		"method_not_allowed":     "platform/httpserver/server.go",
		"precondition_failed":    "internal/media/adapters/tus/server.go",
		"request_too_large":      "internal/media/adapters/tus/server.go",
		"unsupported_media_type": "internal/media/adapters/tus/server.go",
		"upload_offset_conflict": "internal/media/adapters/tus/server.go",
		"insufficient_storage":   "internal/media/adapters/tus/server.go",
		"invalid_callback":       "internal/orders/adapters/http/handler.go",
		"playback_forbidden":     "internal/playback/adapters/http/handler.go",
		"comment_forbidden":      "internal/discussion/adapters/http/handler.go",
		"comment_removed":        "internal/discussion/adapters/http/handler.go",
	}
}
