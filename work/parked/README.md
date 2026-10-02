# Parked work

## learn-lesson-panel-tabs.patch

Origin: uncommitted WIP left by a courses session that ended (recovered from
`stash@{0}` on 2026-09-30). It moves the lesson page's transcript and discussion
into tabs (`Tabs` from `presentation/ui`). Base commit: `1c90bfc`.

Not taken into payment (O05): it is a learning-page redesign (D-c), and it changes
how the transcript and discussion are reached — the e2e for both would need to
select a tab first. Take it up with D-c, re-based onto the current tree.

It also needs the i18n key `common.panels` (uk `'Матеріали уроку'`, en
`'Lesson panels'`), which was part of the same WIP but lived in the locale files.
