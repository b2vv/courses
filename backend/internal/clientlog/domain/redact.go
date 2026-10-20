package domain

import "regexp"

// The §17 never-log set reaches the client log path in two shapes: as a
// sensitive key carrying a value (`password=…`, `"token":"…"`, a cookie or a
// session id) and as a bare credential token (an `Authorization: Bearer …`
// header value, a JWT). A browser error message or stack can embed either when
// it quotes a request, a URL or a header. The schema itself is closed —
// DisallowUnknownFields at the adapter rejects any field the struct does not
// name, so a top-level `password` field can never arrive — which is the
// field-level guarantee; this value-scrub is the second level, over the free
// text the schema does allow (message, stack, route).
//
// The key list covers passwords, session cookies, playback/payment tokens and
// signatures — the brief §17 set. The JWT match is deliberately best-effort
// (shape only); it may miss an unusual token or, rarely, over-redact a
// JWT-shaped non-secret, which is the safe direction for a log line.
var (
	// sensitiveKeyValue matches `key = value` and `"key": "value"` forms and
	// captures the key so the replacement keeps it and masks only the value.
	sensitiveKeyValue = regexp.MustCompile(`(?i)\b(password|passwd|pwd|secret|token|authorization|auth|cookie|set-cookie|session|sessionid|session_id|merchant[_-]?secret|signature|api[_-]?key|access[_-]?token|refresh[_-]?token)\b(\s*["']?\s*[:=]\s*["']?)([^\s"'&,;}]+)`)

	// bearerToken masks the credential after a Bearer scheme.
	bearerToken = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=\-]+`)

	// jwtShape masks a three-segment base64url token (header.payload.signature).
	jwtShape = regexp.MustCompile(`\beyJ[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+`)
)

const redactedPlaceholder = "[redacted]"

// scrubSecrets masks §17 credentials embedded in free text before it becomes a
// log line. It never grows the set of fields; it only replaces values in place.
func scrubSecrets(text string) string {
	if text == "" {
		return text
	}
	text = jwtShape.ReplaceAllString(text, redactedPlaceholder)
	text = bearerToken.ReplaceAllString(text, "Bearer "+redactedPlaceholder)
	text = sensitiveKeyValue.ReplaceAllString(text, "$1$2"+redactedPlaceholder)
	return text
}
