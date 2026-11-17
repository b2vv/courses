**No new findings. The w1 Major is closed.**

- **Destination:** [recovery.ts:233](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.ts:233) selects a case-insensitive matching **track**, then the default **track**. Orphan-only input and a nonmatching track without a default both return `[]`; neither prohibited fallback remains.
- **Snapshot preserved:** omission changes only the returned offers. The key, revision and contents remain stored. The [hook:28](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts:28) rereads when the tab set changes. Direct execution confirmed the same snapshot becomes offered when a default track appears.
- **Mechanism verified:** [LessonLanguageTabs.tsx:74](/Users/strelia/projects/courses/frontend/src/modules/admin/ui/LessonLanguageTabs.tsx:74) enters tab mode only with a Ready video and waits for both inventories. [audiotrack.go:160](/Users/strelia/projects/courses/backend/internal/media/domain/audiotrack.go:160) requires exactly one default track, itself Ready, for a Ready asset.
- **Single editor unchanged:** no tabs still yields destination `''`; the legacy key remains owned, while detached track/language snapshots remain eligible.

`npm run typecheck` **passed**. Direct production-function checks **passed** for destination selection, retained storage and single-editor behavior. Rstest was not rerun because of the known sandbox listener restriction. Unchanged w1 conclusions are retained.

The live frontend diff matches frozen SHA-256 `61b2e0c186f00eb0fea78cf9f4fd039947650f744c5915fc3edfc812ee24075f`. No edits, git writes or Docker operations.

Counts: **Critical 0 · Major 0 · Minor 0**.

**VERDICT (ms3d-fix-w2): SHIP**