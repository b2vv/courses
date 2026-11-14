Reviewed `origin/main...HEAD` at `4f29e50`. Two verified findings:

- **Major — Failed audio hides the replacement.** [LessonVideoField.tsx:68](/Users/strelia/projects/courses/frontend/src/modules/media/ui/LessonVideoField.tsx:68). When the only audio track fails, the backend changes `awaiting_audio → failed`. With an older Ready generation, `shownVideo` then selects that older video, hiding the replacement’s audio controls and retry action. D-L2-10 needs an explicit recovery path and regression coverage.

- **Major — Missing extension validation before mint.** [admin.courses.$courseId.tsx:263](/Users/strelia/projects/courses/frontend/src/routes/admin.courses.$courseId.tsx:263). Upload proceeds directly to mint; neither the application validator nor tus checks the extension. Overriding the picker filter permits `.wav` or arbitrary files to consume upload traffic before backend rejection, contrary to D-L2-5.

Default/delete guards match the backend. New e2e tests use real uploads, worker processing, scoped counts, and master-playlist assertions. Full gates and visual accessibility checks were not run.

The working file could not be created or appended: filesystem permissions prohibit writes. No repository files changed.

VERDICT: BLOCK