# ADR 004 — Local private storage first
Status: accepted design; adapter implementation scheduled for Phase 5.

Local volumes minimize initial operations. Application code exchanges opaque storage keys through MediaStorage rather than filesystem paths. Worker gets writable access; Nginx read-only; API and SSR get no media mount. Atomic publication is a port-level operation with adapter-specific semantics. S3 replacement uses immutable objects plus manifest commit, never assumed directory rename. Disk quotas, backup, cleanup and traversal/symlink defenses gate uploads.
