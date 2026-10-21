// A sink for browser-captured failures. It is best-effort and fire-and-forget:
// `capture` must never throw and never surface a transport failure to the
// caller (a lost log is lost silently, never re-captured — otherwise a failing
// ingest would loop). Presentation never calls this directly; the global bridge
// in the client entry does, through this port.
export interface ClientLogReporter {
  capture(error: unknown): void;
}
