# ADR 001 — Modular monolith
Status: accepted, 2026-09-03.

Use one Go module with inward-facing module boundaries and separate API/worker executables. Independent services would add distributed transaction and deployment costs without present scale evidence. Worker isolation allows FFmpeg resource limits without duplicating domain rules. Extract a service only with measured operational need.
