# System context

Courses is a self-hosted course storefront and future learning platform. Visitors browse published courses. Students, instructors and administrators become authenticated actors in Phase 3. Foundation implements public discovery only; purchases, enrollment and video playback are roadmap capabilities.

The browser connects to one Nginx origin. Nginx delegates rendering to React SSR and JSON requests to Go. PostgreSQL owns durable business records; Redis backs Asynq. Local media is private. See `../foundation/plan.md` for acceptance scope and `../foundation/report.md` for measured evidence.
