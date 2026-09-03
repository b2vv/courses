# Backend architecture

One Go module hosts API and media-worker binaries. Domain packages depend only on Go standard library. Application packages use domain and ports. Ports expose domain types; adapters translate HTTP/SQL records. `cmd` is the composition root. No global clients, mutable service locators or runtime ORM.

Course is the first implemented business module. Published list/detail use cases enforce input bounds and typed failures. PostgreSQL implements the repository via sqlc generated queries. A read-only repeatable-read transaction makes page items and total consistent. Hidden courses are indistinguishable from absent courses through public endpoints.

Identity (User, Role, Session), Enrollment and its playback policy, Media (VideoAsset, Upload, TranscodingJob), Transcript and Order are deliberately not fake runtime modules. Their domain terms and ports will be introduced with consuming use cases in the planned phases. Order/PaymentGateway and TranscriptionGateway contract sketches are documented, not injected as unused interfaces.

HTTP owns parsing/status/error envelopes. Errors expose safe codes and request IDs, never DB causes. Request duration/status and database duration are instrumented. Request traces use OpenTelemetry; Prometheus scrapes internal metrics. Dependencies have bounded startup and readiness checks. SIGTERM drains HTTP and Asynq before closing pools.
