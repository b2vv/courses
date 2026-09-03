# ADR 006 — Timestamped transcript segments
Status: accepted design; implementation scheduled for Phase 7.

TranscriptSegment has id, lessonId, startMs, endMs, text, optional speaker/confidence and documentLinks. Validate nonnegative ordered times and confidence 0..1. DocumentLink has id, label and an allowed http(s) URL; never render arbitrary HTML from transcript text.

Manual fixtures establish the API before automatic speech recognition. A transcription port can later wrap faster-whisper or whisper.cpp. No Python runtime now. UI will map ?t= seconds to milliseconds, seek on activation, highlight active segments, pause autoscroll after manual movement and keep keyboard access. WebVTT export is a later adapter.
