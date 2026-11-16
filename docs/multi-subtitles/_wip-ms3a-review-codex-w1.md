- **Major — [LessonLanguageTabs.tsx:136](/Users/strelia/projects/courses/frontend/src/modules/admin/ui/LessonLanguageTabs.tsx:136)**  
  A background tracks/inventory refresh fails while editors contain unsaved changes. Even with cached lists available, the error branch unmounts every editor. A successful retry mounts them from server data, dropping live edits. Preserve mounted panels while showing the error. **Breaks MS-D6 dirty-state preservation / M3-R11(b).**

- **Major — [LessonLanguageTabs.tsx:235](/Users/strelia/projects/courses/frontend/src/modules/admin/ui/LessonLanguageTabs.tsx:235)**  
  Dirty `en` lines target renamed track language `fr`, whose transcript already exists. Save verifies the track, then opens replacement confirmation. Another client renames or deletes the track before confirmation; clicking Replace still sends `PUT language=fr`. Recheck the track after confirmation while retaining the displayed transcript token. **Breaks MS-D6 / M3-R10: changed track → 0 PUT.** The confirmation wait exceeds the accepted one-request race window.

**Verdict: BLOCK.**

Static, read-only review; no tests or browser checks executed.