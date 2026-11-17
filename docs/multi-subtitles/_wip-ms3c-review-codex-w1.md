**BLOCK**

Reviewed the requested code snapshot; its hashes match run 2. No files modified or tests executed.

1. **Major — R8 and R11b–d accept the outgoing payload as their expected result.**  
   [asr.spec.ts:705](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:705), [asr.spec.ts:988](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:988).  
   R8 proves that the server recognized distinct audio, but never compares each saved PUT’s text with its corresponding draft. A UI regression loading the embedded draft into both tabs would pass: both GETs would equal their respective, incorrect PUTs. R11 similarly derives expected saved text from the PUT instead of the edits entered at lines 1008, 1014 and 1029; saving older text with fresh IDs could pass.  
   **Fix:** assert each PUT against independently expected content: the corresponding ASR draft for R8 and the entered/restored text for R11. Then assert the persisted GET against those expectations.

2. **Major — R12b cannot distinguish language matching from unconditional default placement.**  
   [asr.spec.ts:1111](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:1111), [asr.spec.ts:1123](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:1123).  
   Both variants finish with one default `und` track. The “same-language” variant also starts with `und`, so an implementation that always chooses the default passes both. Both stop after Restore, leaving the Save→200 and original-transcript preservation delegated from R11e to R12b unproved. The promotion identity check at line 1134 is also skipped when automatic recognition has not produced `oldDrafts[0]`.  
   **Fix:** restore a known-language snapshot into a matching **non-default** replacement track alongside a different default; assert the offer’s location, save it, and verify the new and preserved transcripts. Capture asset identity from media inventory/upload rather than an optional ASR draft.

3. **Major — R15’s `en` setup races automatic recognition.**  
   [asr.spec.ts:1214](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:1214).  
   The worker can claim the Ready embedded track as `und` before the rename. `startRecognition` then accepts that running/ready draft, and line 1220 fails because its immutable `trackLanguage` remains `und`. That is correct production behavior, including the `und` exception.  
   **Fix:** finish the initial recognition, rename, explicitly regenerate, and wait for an `en` draft before testing the multilingual switch.

4. **Minor — one teardown transport exception skips remaining course cleanup.**  
   [asr.spec.ts:407](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:407).  
   Soft assertions handle HTTP failures, but a thrown GET/DELETE or parsing error exits the whole loop, leaving later courses behind.  
   **Fix:** catch and record failures per course while continuing cleanup.

The row audit is:

| Row | Assessment |
|---|---|
| R2 | Proves PUT language through real requests. |
| R4 | Proves success and PATCH-success/PUT-failure/reload/restore. The 503 injection is explicitly permitted. |
| R6 | Proves two native IndexedDB keys, isolated restoration, repeated offers and retirement. |
| R8 | Real recognition and persistence; per-track loaded content is insufficiently asserted—finding 1. |
| R9 | Required migrations are present and use real recognition. |
| R10 | Proves foreign rename refusal and GET-503 refusal with retained lines and zero PUTs; retry persists. |
| R11a | Proves successful reuse across languages with fresh IDs and preservation. |
| R11b/c/d | Real saves and preservation checks; intended edited/restored content needs independent assertions—finding 1. |
| R12a | Valid promotion/unmount recovery proof. |
| R12b | Real replacement/recovery, but weak destination discrimination and incomplete linked save proof—finding 2. |
| R13a/b | Proves real T2 writes, T1-token conflicts, preserved server text and retained local text. |
| R14 | Proves real multilingual upload metadata → recognized draft. |
| R15 | Exercises invalidation/regeneration for both modes; `en` setup is flaky—finding 3. |

R12a’s ~5 seconds is credible: it delays delivery of an **actual Ready response**, without fabricating state or recognition. Focus remains in the text field until release; the recovery idle timer is 90 seconds, so this exercises unmount flushing. It does not need to wait for ASR.

The defect-1 fix at [LessonAudioTracksField.tsx:462](/Users/strelia/projects/courses/frontend/src/modules/media/ui/LessonAudioTracksField.tsx:462) is correct and complete for the reported class. It stops submission before validation and covers keyboard submission through the same handler. I found no other nested form under the course form; language dialogs use explicit button actions, and pricing/grants are outside it.

There are no fixed sleeps or Redis queue pauses in this delta. The Ready gate releases in `finally`. Existing cleanup artifacts record 33 deleted courses and zero unexpected course PUTs.

Merge additionally awaits the planned final ordered gate: full repository e2e, including unchanged `transcript.spec`, remains **NOT_RUN** in the supplied journals.