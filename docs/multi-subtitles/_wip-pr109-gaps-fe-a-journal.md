NEXT: Read current rules and fe-a mutation journal; enumerate GAP rows, then close each with baseline and exact mutation proof.
FACT: Added adapter malformed-envelope/generation cases and event-driven in-flight inventory cancellation test for A-007/A-011/A-012/A-028/X-007/X-009.
FACT: Extended request/event-boundary tests for I-003/I-004/I-007/I-009–I-015/I-024/I-025/X-005, including transaction error and wrong-key cleanup.
FACT: Added application cleanup same-revision field-isolation test at supplied store seam for R-018/R-021–R-024. NEXT: Run unmutated files, then exact mutations individually.
NEXT: A-007 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: A-007 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
NEXT: A-007 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: A-007 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-A-007-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for obsolete lessonId

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for obsolete lessonId'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for obsolete lessonId",
  "status": "fail",
  "duration": 3,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: [ { …(12) } ] } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 96,
        "column": 67,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "error": {
-     "code": "invalid_response",
+   "ok": true,
+   "value": [
+     {
+       "assetId": "50000000-0000-4000-8000-000000000001",
+       "failure": null,
+       "generation": 1,
+       "language": "uk",
+       "lessonId": "20000000-0000-4000-8000-000000000001",
+       "segments": [
+         {
+           "confidence": 0.9,
+           "endMs": 1000,
+           "id": "70000000-0000-4000-8000-000000000001",
+           "speaker": "Lecturer",
+           "startMs": 0,
+           "text": "Вітаю",
          },
-   "ok": false,
+       ],
+       "stale": false,
+       "state": "ready",
+       "trackId": "60000000-0000-4000-8000-000000000001",
+       "trackLanguage": "uk",
+       "trackMultilingual": true,
+       "updatedAt": "2026-11-15T12:00:00.123456Z",
+     },
+   ],
  }
```

codeFrame (error 1):

```text
94 |   ])('rejects the whole list for $name', async ({ response }) => {
95 |     server.use(http.get(api, () => HttpResponse.json(response)));
96 |     expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual(invalid);
   |                                                                   ^
97 |   });
98 | 
```


~~~~
FACT: A-007 | src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for obsolete lessonId | baseline PASS (recorded batch baseline) | 96 |     expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual(invalid); | CLOSED | DEFECT=no
FACT: A-007 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
NEXT: A-011 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: A-011 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-A-011-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for fractional generation

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for fractional generation'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for fractional generation",
  "status": "fail",
  "duration": 4,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: [ { …(12) } ] } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 96,
        "column": 67,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "error": {
-     "code": "invalid_response",
+   "ok": true,
+   "value": [
+     {
+       "assetId": "50000000-0000-4000-8000-000000000001",
+       "failure": null,
+       "generation": 1.5,
+       "language": "uk",
+       "lessonId": "20000000-0000-4000-8000-000000000001",
+       "segments": [
+         {
+           "confidence": 0.9,
+           "endMs": 1000,
+           "id": "70000000-0000-4000-8000-000000000001",
+           "speaker": "Lecturer",
+           "startMs": 0,
+           "text": "Вітаю",
          },
-   "ok": false,
+       ],
+       "stale": false,
+       "state": "ready",
+       "trackId": "60000000-0000-4000-8000-000000000001",
+       "trackLanguage": "uk",
+       "trackMultilingual": true,
+       "updatedAt": "2026-11-15T12:00:00.123456Z",
+     },
+   ],
  }
```

codeFrame (error 1):

```text
94 |   ])('rejects the whole list for $name', async ({ response }) => {
95 |     server.use(http.get(api, () => HttpResponse.json(response)));
96 |     expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual(invalid);
   |                                                                   ^
97 |   });
98 | 
```


~~~~
FACT: A-011 | src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for fractional generation | baseline PASS (recorded batch baseline) | 96 |     expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual(invalid); | CLOSED | DEFECT=no
FACT: A-011 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
NEXT: A-012 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: A-012 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-A-012-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for zero generation

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for zero generation'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for zero generation",
  "status": "fail",
  "duration": 4,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: [ { …(12) } ] } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 96,
        "column": 67,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "error": {
-     "code": "invalid_response",
+   "ok": true,
+   "value": [
+     {
+       "assetId": "50000000-0000-4000-8000-000000000001",
+       "failure": null,
+       "generation": 0,
+       "language": "uk",
+       "lessonId": "20000000-0000-4000-8000-000000000001",
+       "segments": [
+         {
+           "confidence": 0.9,
+           "endMs": 1000,
+           "id": "70000000-0000-4000-8000-000000000001",
+           "speaker": "Lecturer",
+           "startMs": 0,
+           "text": "Вітаю",
          },
-   "ok": false,
+       ],
+       "stale": false,
+       "state": "ready",
+       "trackId": "60000000-0000-4000-8000-000000000001",
+       "trackLanguage": "uk",
+       "trackMultilingual": true,
+       "updatedAt": "2026-11-15T12:00:00.123456Z",
+     },
+   ],
  }
```

codeFrame (error 1):

```text
94 |   ])('rejects the whole list for $name', async ({ response }) => {
95 |     server.use(http.get(api, () => HttpResponse.json(response)));
96 |     expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual(invalid);
   |                                                                   ^
97 |   });
98 | 
```

### [F02] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for negative generation

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for negative generation'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for negative generation",
  "status": "fail",
  "duration": 5,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: [ { …(12) } ] } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 96,
        "column": 67,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "error": {
-     "code": "invalid_response",
+   "ok": true,
+   "value": [
+     {
+       "assetId": "50000000-0000-4000-8000-000000000001",
+       "failure": null,
+       "generation": -1,
+       "language": "uk",
+       "lessonId": "20000000-0000-4000-8000-000000000001",
+       "segments": [
+         {
+           "confidence": 0.9,
+           "endMs": 1000,
+           "id": "70000000-0000-4000-8000-000000000001",
+           "speaker": "Lecturer",
+           "startMs": 0,
+           "text": "Вітаю",
          },
-   "ok": false,
+       ],
+       "stale": false,
+       "state": "ready",
+       "trackId": "60000000-0000-4000-8000-000000000001",
+       "trackLanguage": "uk",
+       "trackMultilingual": true,
+       "updatedAt": "2026-11-15T12:00:00.123456Z",
+     },
+   ],
  }
```

codeFrame (error 1):

```text
94 |   ])('rejects the whole list for $name', async ({ response }) => {
95 |     server.use(http.get(api, () => HttpResponse.json(response)));
96 |     expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual(invalid);
   |                                                                   ^
97 |   });
98 | 
```


~~~~
FACT: A-012 | src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for zero generation; src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for negative generation | baseline PASS (recorded batch baseline) | 96 |     expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual(invalid); / 96 |     expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual(invalid); | CLOSED | DEFECT=no
FACT: A-012 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
NEXT: A-028 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: A-028 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-A-028-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: GAP inventory cancellation reports request_cancelled after request starts

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'GAP inventory cancellation reports request_cancelled after request starts'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "GAP inventory cancellation reports request_cancelled after request starts",
  "status": "fail",
  "duration": 3,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: [] } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 214,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 214
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "error": {
-     "code": "request_cancelled",
-   },
-   "ok": false,
+   "ok": true,
+   "value": [],
  }
```

codeFrame (error 1):

```text
212 |   controller.abort();
213 |   releaseResponse();
214 |   expect(await pending).toEqual({ ok: false, error: { code: 'request_cancelled' } });
    |                         ^
215 | });
216 | 
```


~~~~
FACT: A-028 | src/modules/transcript/adapters/draft-repository.test.ts :: GAP inventory cancellation reports request_cancelled after request starts | baseline PASS (recorded batch baseline) | 214 |   expect(await pending).toEqual({ ok: false, error: { code: 'request_cancelled' } }); | CLOSED | DEFECT=no
FACT: A-028 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
NEXT: I-003 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-003 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-003-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects missing key without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects missing key without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects missing key without offering corrupt recovery",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { key: undefined, capturedAt: 1, …(4) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 217,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 217
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": undefined,
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "1e724549-0dfa-48dc-9265-6273268f1afc",
}
```

codeFrame (error 1):

```text
215 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
216 |   storedValue(value);
217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
218 | });
219 | 
```


~~~~
FACT: I-003 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects missing key without offering corrupt recovery | baseline PASS (recorded batch baseline) | 217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-003 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-004 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-004 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-004-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects missing capturedAt without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects missing capturedAt without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects missing capturedAt without offering corrupt recovery",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 217,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 217
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": undefined,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "a99e5eef-4c95-42df-856d-f93d06540056",
}
```

codeFrame (error 1):

```text
215 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
216 |   storedValue(value);
217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
218 | });
219 | 
```


~~~~
FACT: I-004 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects missing capturedAt without offering corrupt recovery | baseline PASS (recorded batch baseline) | 217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-004 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-007 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-007 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-007-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric language without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric language without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric language without offering corrupt recovery",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 217,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 217
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": 42,
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "2750d1c7-747b-4f52-ab83-0886e0a82eb1",
}
```

codeFrame (error 1):

```text
215 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
216 |   storedValue(value);
217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
218 | });
219 | 
```


~~~~
FACT: I-007 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric language without offering corrupt recovery | baseline PASS (recorded batch baseline) | 217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-007 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-009 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-009 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-009-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric expectedUpdatedAt without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric expectedUpdatedAt without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric expectedUpdatedAt without offering corrupt recovery",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 217,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 217
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": 42,
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "864a505b-6b87-46cc-8745-0c32e2b7aa66",
}
```

codeFrame (error 1):

```text
215 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
216 |   storedValue(value);
217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
218 | });
219 | 
```


~~~~
FACT: I-009 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric expectedUpdatedAt without offering corrupt recovery | baseline PASS (recorded batch baseline) | 217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-009 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-010 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-010 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-010-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric draftUpdatedAt without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric draftUpdatedAt without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric draftUpdatedAt without offering corrupt recovery",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 217,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 217
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "draftUpdatedAt": 42,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "8b257be8-d639-4c39-9218-f7aebc1e6ed4",
}
```

codeFrame (error 1):

```text
215 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
216 |   storedValue(value);
217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
218 | });
219 | 
```


~~~~
FACT: I-010 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric draftUpdatedAt without offering corrupt recovery | baseline PASS (recorded batch baseline) | 217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-010 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-011 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-011 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-011-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric line id without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric line id without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric line id without offering corrupt recovery",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 217,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 217
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": 42,
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "fa5ffd11-c4f3-40ae-8717-cbf1b4038d5b",
}
```

codeFrame (error 1):

```text
215 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
216 |   storedValue(value);
217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
218 | });
219 | 
```


~~~~
FACT: I-011 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric line id without offering corrupt recovery | baseline PASS (recorded batch baseline) | 217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-011 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-012 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-012 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-012-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric startMs without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric startMs without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric startMs without offering corrupt recovery",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 217,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 217
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": 42,
      "text": "Unsaved",
    },
  ],
  "revision": "432097f3-f800-4702-b22c-54d964d7a92a",
}
```

codeFrame (error 1):

```text
215 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
216 |   storedValue(value);
217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
218 | });
219 | 
```


~~~~
FACT: I-012 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric startMs without offering corrupt recovery | baseline PASS (recorded batch baseline) | 217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-012 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-013 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-013 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-013-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric endMs without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric endMs without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric endMs without offering corrupt recovery",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 217,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 217
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": 42,
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "dccee9d0-cb23-4eb9-9413-5e9920b3f437",
}
```

codeFrame (error 1):

```text
215 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
216 |   storedValue(value);
217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
218 | });
219 | 
```


~~~~
FACT: I-013 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric endMs without offering corrupt recovery | baseline PASS (recorded batch baseline) | 217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-013 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-014 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-014 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-014-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric text without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric text without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric text without offering corrupt recovery",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 217,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 217
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": 42,
    },
  ],
  "revision": "7c66350c-3164-460a-8389-9898f9727ef1",
}
```

codeFrame (error 1):

```text
215 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
216 |   storedValue(value);
217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
218 | });
219 | 
```


~~~~
FACT: I-014 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric text without offering corrupt recovery | baseline PASS (recorded batch baseline) | 217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-014 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-015 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-015 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-015-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric documentIds without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric documentIds without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric documentIds without offering corrupt recovery",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 217,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 217
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [
        42,
      ],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "9b73ce95-e790-4aec-a50f-f986fe779342",
}
```

codeFrame (error 1):

```text
215 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
216 |   storedValue(value);
217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
218 | });
219 | 
```


~~~~
FACT: I-015 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric documentIds without offering corrupt recovery | baseline PASS (recorded batch baseline) | 217 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-015 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-024 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-024 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-024-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB preserves the ASR draft token alongside valid recovery text

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB preserves the ASR draft token alongside valid recovery text'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB preserves the ASR draft token alongside valid recovery text",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to deeply equal { Object (key, revision, ...) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 223,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 223
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,8 +1,8 @@
  {
    "capturedAt": 1,
-   "draftUpdatedAt": "D1",
+   "draftUpdatedAt": undefined,
    "expectedUpdatedAt": "T1",
    "key": "author:lesson:track:one",
    "language": "es",
    "lines": [
      {
```

codeFrame (error 1):

```text
221 |   const value = { ...snapshot, draftUpdatedAt: 'D1' };
222 |   storedValue(value);
223 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toEqual(value);
    |                                                                        ^
224 | });
225 | 
```


~~~~
FACT: I-024 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB preserves the ASR draft token alongside valid recovery text | baseline PASS (recorded batch baseline) | 223 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toEqual(value); | CLOSED | DEFECT=no
FACT: I-024 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: I-025 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-025 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-025-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: MS3b fix4 atomic IndexedDB cleanup wrong-key uses one readwrite transaction and waits for commit

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'MS3b fix4 atomic IndexedDB cleanup wrong-key uses one readwrite transaction and waits for commit'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "MS3b fix4 atomic IndexedDB cleanup wrong-key uses one readwrite transaction and waits for commit",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"author:lesson:track:one\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 111,
        "column": 29,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 111
    }
  ]
}
```

codeFrame (error 1):

```text
109 |     if (outcome === 'same' || outcome === 'abort' || outcome === 'error')
110 |       expect(remove).toHaveBeenCalledExactlyOnceWith(snapshot.key);
111 |     else expect(remove).not.toHaveBeenCalled();
    |                             ^
112 |     expect(start).toHaveBeenCalledTimes(1);
113 |     await Promise.resolve();
```


~~~~
FACT: I-025 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: MS3b fix4 atomic IndexedDB cleanup wrong-key uses one readwrite transaction and waits for commit | baseline PASS (recorded batch baseline) | 110 |       expect(remove).toHaveBeenCalledExactlyOnceWith(snapshot.key); / 111 |     else expect(remove).not.toHaveBeenCalled(); / 112 |     expect(start).toHaveBeenCalledTimes(1); | CLOSED | DEFECT=no
FACT: I-025 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: R-018 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/application/recovery.test.ts`. Production SHA-256 5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69.
FACT: R-018 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-R-018-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"key":"another-key"}

repro:

```bash
npx rstest 'src/modules/transcript/application/recovery.test.ts' --testNamePattern 'GAP accepted cleanup preserves a same-revision snapshot differing in {"key":"another-key"}'
```

details:

```json
{
  "testPath": "src/modules/transcript/application/recovery.test.ts",
  "project": "dom",
  "fullName": "GAP accepted cleanup preserves a same-revision snapshot differing in {\"key\":\"another-key\"}",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"user:lesson\",\n      \"snapshot-revision\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/application/recovery.test.ts",
        "line": 458,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/application/recovery.test.ts",
      "line": 458
    }
  ]
}
```

codeFrame (error 1):

```text
456 |   const clearIfVersion = rstest.fn(async () => true);
457 |   await clearAcceptedTranscriptSnapshot({ ...memory, read: async () => replacement, clearIfVersion }, accepted, () => false);
458 |   expect(clearIfVersion).not.toHaveBeenCalled();
    |                              ^
459 |   expect(await memory.read(replacement.key)).toEqual(replacement);
460 | });
```


~~~~
FACT: R-018 | src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"key":"another-key"} | baseline PASS (recorded batch baseline) | 458 |   expect(clearIfVersion).not.toHaveBeenCalled(); / 459 |   expect(await memory.read(replacement.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: R-018 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/application/recovery.ts` exit=0.
NEXT: R-021 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/application/recovery.test.ts`. Production SHA-256 5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69.
FACT: R-021 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-R-021-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"language":"es"}

repro:

```bash
npx rstest 'src/modules/transcript/application/recovery.test.ts' --testNamePattern 'GAP accepted cleanup preserves a same-revision snapshot differing in {"language":"es"}'
```

details:

```json
{
  "testPath": "src/modules/transcript/application/recovery.test.ts",
  "project": "dom",
  "fullName": "GAP accepted cleanup preserves a same-revision snapshot differing in {\"language\":\"es\"}",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"user:lesson\",\n      \"snapshot-revision\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/application/recovery.test.ts",
        "line": 458,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/application/recovery.test.ts",
      "line": 458
    }
  ]
}
```

codeFrame (error 1):

```text
456 |   const clearIfVersion = rstest.fn(async () => true);
457 |   await clearAcceptedTranscriptSnapshot({ ...memory, read: async () => replacement, clearIfVersion }, accepted, () => false);
458 |   expect(clearIfVersion).not.toHaveBeenCalled();
    |                              ^
459 |   expect(await memory.read(replacement.key)).toEqual(replacement);
460 | });
```


~~~~
FACT: R-021 | src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"language":"es"} | baseline PASS (recorded batch baseline) | 458 |   expect(clearIfVersion).not.toHaveBeenCalled(); / 459 |   expect(await memory.read(replacement.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: R-021 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/application/recovery.ts` exit=0.
NEXT: R-022 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/application/recovery.test.ts`. Production SHA-256 5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69.
FACT: R-022 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-R-022-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"baseLanguage":"fr"}

repro:

```bash
npx rstest 'src/modules/transcript/application/recovery.test.ts' --testNamePattern 'GAP accepted cleanup preserves a same-revision snapshot differing in {"baseLanguage":"fr"}'
```

details:

```json
{
  "testPath": "src/modules/transcript/application/recovery.test.ts",
  "project": "dom",
  "fullName": "GAP accepted cleanup preserves a same-revision snapshot differing in {\"baseLanguage\":\"fr\"}",
  "status": "fail",
  "duration": 6,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"user:lesson\",\n      \"snapshot-revision\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/application/recovery.test.ts",
        "line": 458,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/application/recovery.test.ts",
      "line": 458
    }
  ]
}
```

codeFrame (error 1):

```text
456 |   const clearIfVersion = rstest.fn(async () => true);
457 |   await clearAcceptedTranscriptSnapshot({ ...memory, read: async () => replacement, clearIfVersion }, accepted, () => false);
458 |   expect(clearIfVersion).not.toHaveBeenCalled();
    |                              ^
459 |   expect(await memory.read(replacement.key)).toEqual(replacement);
460 | });
```


~~~~
FACT: R-022 | src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"baseLanguage":"fr"} | baseline PASS (recorded batch baseline) | 458 |   expect(clearIfVersion).not.toHaveBeenCalled(); / 459 |   expect(await memory.read(replacement.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: R-022 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/application/recovery.ts` exit=0.
NEXT: R-023 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/application/recovery.test.ts`. Production SHA-256 5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69.
FACT: R-023 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-R-023-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"expectedUpdatedAt":"T2"}

repro:

```bash
npx rstest 'src/modules/transcript/application/recovery.test.ts' --testNamePattern 'GAP accepted cleanup preserves a same-revision snapshot differing in {"expectedUpdatedAt":"T2"}'
```

details:

```json
{
  "testPath": "src/modules/transcript/application/recovery.test.ts",
  "project": "dom",
  "fullName": "GAP accepted cleanup preserves a same-revision snapshot differing in {\"expectedUpdatedAt\":\"T2\"}",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"user:lesson\",\n      \"snapshot-revision\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/application/recovery.test.ts",
        "line": 458,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/application/recovery.test.ts",
      "line": 458
    }
  ]
}
```

codeFrame (error 1):

```text
456 |   const clearIfVersion = rstest.fn(async () => true);
457 |   await clearAcceptedTranscriptSnapshot({ ...memory, read: async () => replacement, clearIfVersion }, accepted, () => false);
458 |   expect(clearIfVersion).not.toHaveBeenCalled();
    |                              ^
459 |   expect(await memory.read(replacement.key)).toEqual(replacement);
460 | });
```


~~~~
FACT: R-023 | src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"expectedUpdatedAt":"T2"} | baseline PASS (recorded batch baseline) | 458 |   expect(clearIfVersion).not.toHaveBeenCalled(); / 459 |   expect(await memory.read(replacement.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: R-023 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/application/recovery.ts` exit=0.
NEXT: R-024 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/application/recovery.test.ts`. Production SHA-256 5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69.
FACT: R-024 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-R-024-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"draftUpdatedAt":"D2"}

repro:

```bash
npx rstest 'src/modules/transcript/application/recovery.test.ts' --testNamePattern 'GAP accepted cleanup preserves a same-revision snapshot differing in {"draftUpdatedAt":"D2"}'
```

details:

```json
{
  "testPath": "src/modules/transcript/application/recovery.test.ts",
  "project": "dom",
  "fullName": "GAP accepted cleanup preserves a same-revision snapshot differing in {\"draftUpdatedAt\":\"D2\"}",
  "status": "fail",
  "duration": 6,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"user:lesson\",\n      \"snapshot-revision\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/application/recovery.test.ts",
        "line": 458,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/application/recovery.test.ts",
      "line": 458
    }
  ]
}
```

codeFrame (error 1):

```text
456 |   const clearIfVersion = rstest.fn(async () => true);
457 |   await clearAcceptedTranscriptSnapshot({ ...memory, read: async () => replacement, clearIfVersion }, accepted, () => false);
458 |   expect(clearIfVersion).not.toHaveBeenCalled();
    |                              ^
459 |   expect(await memory.read(replacement.key)).toEqual(replacement);
460 | });
```


~~~~
FACT: R-024 | src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"draftUpdatedAt":"D2"} | baseline PASS (recorded batch baseline) | 458 |   expect(clearIfVersion).not.toHaveBeenCalled(); / 459 |   expect(await memory.read(replacement.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: R-024 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/application/recovery.ts` exit=0.
NEXT: X-005 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: X-005 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-005-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: MS3b fix4 atomic IndexedDB cleanup error uses one readwrite transaction and waits for commit

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'MS3b fix4 atomic IndexedDB cleanup error uses one readwrite transaction and waits for commit'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "MS3b fix4 atomic IndexedDB cleanup error uses one readwrite transaction and waits for commit",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise resolved \"false\" instead of rejecting",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 117,
        "column": 20,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 117
    }
  ]
}
```

diff:

```diff
- Expected:
Error {
  "message": "rejected promise",
}

+ Received:
false
```

codeFrame (error 1):

```text
115 |     if (outcome === 'error') {
116 |       transaction.dispatchEvent(new Event('error'));
117 |       await expect(observed).rejects.toThrow('transaction failed');
    |                    ^
118 |     } else if (outcome === 'abort') {
119 |       transaction.dispatchEvent(new Event('abort'));
```


~~~~
FACT: X-005 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: MS3b fix4 atomic IndexedDB cleanup error uses one readwrite transaction and waits for commit | baseline PASS (recorded batch baseline) | 117 |       await expect(observed).rejects.toThrow('transaction failed'); | CLOSED | DEFECT=no
FACT: X-005 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
NEXT: X-007 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: X-007 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-007-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for null envelope

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for null envelope'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for null envelope",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "TypeError",
      "message": "Cannot convert undefined or null to object",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.ts",
        "line": 36,
        "column": 17,
        "method": "hasOnlyKeys"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.ts",
      "line": 36
    },
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

codeFrame (error 1):

```text
34 | 
35 | function hasOnlyKeys(value: Record<string, unknown>, keys: readonly string[]): boolean {
36 |   return Object.keys(value).every((key) => keys.includes(key));
   |                 ^
37 | }
38 | 
```


~~~~
FACT: X-007 | src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for null envelope | baseline PASS (recorded batch baseline) | NO ATTRIBUTABLE ASSERTION | REMAINING | DEFECT=no
FACT: X-007 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
NEXT: X-009 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: X-009 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-009-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: GAP inventory cancellation reports request_cancelled after request starts

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'GAP inventory cancellation reports request_cancelled after request starts'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "GAP inventory cancellation reports request_cancelled after request starts",
  "status": "fail",
  "duration": 3,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: false, …(1) } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 214,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 214
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
    "error": {
-     "code": "request_cancelled",
+     "code": "invalid_response",
    },
    "ok": false,
  }
```

codeFrame (error 1):

```text
212 |   controller.abort();
213 |   releaseResponse();
214 |   expect(await pending).toEqual({ ok: false, error: { code: 'request_cancelled' } });
    |                         ^
215 | });
216 | 
```


~~~~
FACT: X-009 | src/modules/transcript/adapters/draft-repository.test.ts :: GAP inventory cancellation reports request_cancelled after request starts | baseline PASS (recorded batch baseline) | 214 |   expect(await pending).toEqual({ ok: false, error: { code: 'request_cancelled' } }); | CLOSED | DEFECT=no
FACT: X-009 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
FACT: Initial baseline exit 0, 118 tests / 3 files. Exact mutation batch: 23 CLOSED; X-007 reached null-envelope rejection but assertion wrapped await rather than resolves; strengthened assertion for an attributable named rejection failure. Added event-driven hook lifecycle tests for all H GAPs plus X-012/X-013/X-014. NEXT: Unmutated hook baseline.
NEXT: H-001 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-001 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-001-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP switching a loaded visit hides its offer until the new store read resolves

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP switching a loaded visit hides its offer until the new store read resolves'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP switching a loaded visit hides its offer until the new store read resolves",
  "status": "fail",
  "duration": 12,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected true to be false // Object.is equality",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 280,
        "column": 37,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 280
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- false
+ true
```

codeFrame (error 1):

```text
278 |   expect(f.result.current.offer).toMatchObject({ snapshot: a });
279 |   f.rerender({ key: 'gap-switch:b', source: other });
280 |   expect(f.result.current.isLoaded).toBe(false);
    |                                     ^
281 |   expect(f.result.current.offer).toEqual({ kind: 'none' });
282 |   await act(async () => next.resolve(null));
```


~~~~
FACT: H-001 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP switching a loaded visit hides its offer until the new store read resolves | baseline PASS (recorded batch baseline) | 278 |   expect(f.result.current.offer).toMatchObject({ snapshot: a }); / 280 |   expect(f.result.current.isLoaded).toBe(false); / 281 |   expect(f.result.current.offer).toEqual({ kind: 'none' }); | CLOSED | DEFECT=no
FACT: H-001 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-002 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-002 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-002-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete read completion cannot authorize retirement in an unread visit'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete read completion cannot authorize retirement in an unread visit",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-unread:b\",\n      \"133f827b-8d07-4455-a521-d82d6c0e09b9\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 322,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 322
    }
  ]
}
```

codeFrame (error 1):

```text
320 |   act(() => f.result.current.finalize());
321 |   await act(async () => {});
322 |   expect(clear).not.toHaveBeenCalled();
    |                     ^
323 |   await act(async () => b.resolve(null));
324 | });
```

### [F02] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalize cannot retire an own write before the initial read finishes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP finalize cannot retire an own write before the initial read finishes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP finalize cannot retire an own write before the initial read finishes",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-initial:finalize\",\n      \"be3c299f-3f70-4f1a-9490-0686ef8334e2\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 344,
        "column": 23,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 344
    }
  ]
}
```

codeFrame (error 1):

```text
342 |     });
343 |     await act(async () => {});
344 |     expect(clear).not.toHaveBeenCalled();
    |                       ^
345 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1');
346 |     await act(async () => pending.resolve(null));
```

### [F03] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP clean capture cannot retire an own write before the initial read finishes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP clean capture cannot retire an own write before the initial read finishes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP clean capture cannot retire an own write before the initial read finishes",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-initial:clean capture\",\n      \"d2750bcd-fb7b-4719-b976-85ef2c454b48\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 344,
        "column": 23,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 344
    }
  ]
}
```

codeFrame (error 1):

```text
342 |     });
343 |     await act(async () => {});
344 |     expect(clear).not.toHaveBeenCalled();
    |                       ^
345 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1');
346 |     await act(async () => pending.resolve(null));
```


~~~~
FACT: H-002 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalize cannot retire an own write before the initial read finishes; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP clean capture cannot retire an own write before the initial read finishes | baseline PASS (recorded batch baseline) | 322 |   expect(clear).not.toHaveBeenCalled(); / 344 |     expect(clear).not.toHaveBeenCalled(); / 345 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1'); / 344 |     expect(clear).not.toHaveBeenCalled(); / 345 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1'); | CLOSED | DEFECT=no
FACT: H-002 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-003 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-003 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-003-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read resolve cannot replace the current visit offer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete initial read resolve cannot replace the current visit offer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete initial read resolve cannot replace the current visit offer",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'offer', snapshot: { …(7) } } to match object { snapshot: { …(7) } }\n(1 matching property omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 303,
        "column": 36,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 303
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
    "snapshot": {
      "baseLanguage": "en",
      "capturedAt": 1,
      "expectedUpdatedAt": "T1",
-     "key": "gap-obsolete-resolve:b",
+     "key": "obsolete-a",
      "language": "uk",
      "lines": [
        {
          "documentIds": [],
          "endMs": "1000",
          "id": "line",
          "startMs": "0",
          "text": "Recovered",
        },
      ],
-     "revision": "revision:gap-obsolete-resolve:b:Recovered",
+     "revision": "revision:obsolete-a:Recovered",
    },
  }
```

codeFrame (error 1):

```text
301 |     });
302 |     expect(read).toHaveBeenCalledTimes(2);
303 |     expect(f.result.current.offer).toMatchObject({ snapshot: b });
    |                                    ^
304 |     expect(f.result.current.isLoaded).toBe(true);
305 |   });
```


~~~~
FACT: H-003 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read resolve cannot replace the current visit offer | baseline PASS (recorded batch baseline) | 302 |     expect(read).toHaveBeenCalledTimes(2); / 303 |     expect(f.result.current.offer).toMatchObject({ snapshot: b }); / 304 |     expect(f.result.current.isLoaded).toBe(true); | CLOSED | DEFECT=no
FACT: H-003 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-004 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-004 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-004-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read resolve cannot replace the current visit offer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete initial read resolve cannot replace the current visit offer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete initial read resolve cannot replace the current visit offer",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'none' } to match object { snapshot: { …(7) } }\n(1 matching property omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 303,
        "column": 36,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 303
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "snapshot": {
-     "baseLanguage": "en",
-     "capturedAt": 1,
-     "expectedUpdatedAt": "T1",
-     "key": "gap-obsolete-resolve:b",
-     "language": "uk",
-     "lines": [
-       {
-         "documentIds": [],
-         "endMs": "1000",
-         "id": "line",
-         "startMs": "0",
-         "text": "Recovered",
-       },
-     ],
-     "revision": "revision:gap-obsolete-resolve:b:Recovered",
-   },
+   "kind": "none",
  }
```

codeFrame (error 1):

```text
301 |     });
302 |     expect(read).toHaveBeenCalledTimes(2);
303 |     expect(f.result.current.offer).toMatchObject({ snapshot: b });
    |                                    ^
304 |     expect(f.result.current.isLoaded).toBe(true);
305 |   });
```

### [F02] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read reject cannot replace the current visit offer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete initial read reject cannot replace the current visit offer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete initial read reject cannot replace the current visit offer",
  "status": "fail",
  "duration": 2,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'none' } to match object { snapshot: { …(7) } }\n(1 matching property omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 303,
        "column": 36,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 303
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "snapshot": {
-     "baseLanguage": "en",
-     "capturedAt": 1,
-     "expectedUpdatedAt": "T1",
-     "key": "gap-obsolete-reject:b",
-     "language": "uk",
-     "lines": [
-       {
-         "documentIds": [],
-         "endMs": "1000",
-         "id": "line",
-         "startMs": "0",
-         "text": "Recovered",
-       },
-     ],
-     "revision": "revision:gap-obsolete-reject:b:Recovered",
-   },
+   "kind": "none",
  }
```

codeFrame (error 1):

```text
301 |     });
302 |     expect(read).toHaveBeenCalledTimes(2);
303 |     expect(f.result.current.offer).toMatchObject({ snapshot: b });
    |                                    ^
304 |     expect(f.result.current.isLoaded).toBe(true);
305 |   });
```

### [F03] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete read completion cannot authorize retirement in an unread visit'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete read completion cannot authorize retirement in an unread visit",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-unread:b\",\n      \"d84e5e78-f686-494a-93a8-a6ad68df5f76\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 322,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 322
    }
  ]
}
```

codeFrame (error 1):

```text
320 |   act(() => f.result.current.finalize());
321 |   await act(async () => {});
322 |   expect(clear).not.toHaveBeenCalled();
    |                     ^
323 |   await act(async () => b.resolve(null));
324 | });
```


~~~~
FACT: H-004 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read resolve cannot replace the current visit offer; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read reject cannot replace the current visit offer; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit | baseline PASS (recorded batch baseline) | 302 |     expect(read).toHaveBeenCalledTimes(2); / 303 |     expect(f.result.current.offer).toMatchObject({ snapshot: b }); / 304 |     expect(f.result.current.isLoaded).toBe(true); / 302 |     expect(read).toHaveBeenCalledTimes(2); / 303 |     expect(f.result.current.offer).toMatchObject({ snapshot: b }); / 304 |     expect(f.result.current.isLoaded).toBe(true); / 322 |   expect(clear).not.toHaveBeenCalled(); | CLOSED | DEFECT=no
FACT: H-004 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-006 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-006 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-006-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalization waits for its write commit before conditional retirement

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP finalization waits for its write commit before conditional retirement'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP finalization waits for its write commit before conditional retirement",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-commit\",\n      \"74a75432-d893-43f8-ac16-8c99c699cecd\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 363,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 363
    }
  ]
}
```

codeFrame (error 1):

```text
361 |   await act(async () => {});
362 |   expect(write).toHaveBeenCalledTimes(1);
363 |   expect(clear).not.toHaveBeenCalled();
    |                     ^
364 |   await act(async () => commit.resolve());
365 |   expect(clear).toHaveBeenCalledExactlyOnceWith('gap-commit', write.mock.calls[0]?.[0].revision);
```


~~~~
FACT: H-006 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalization waits for its write commit before conditional retirement | baseline PASS (recorded batch baseline) | 362 |   expect(write).toHaveBeenCalledTimes(1); / 363 |   expect(clear).not.toHaveBeenCalled(); / 365 |   expect(clear).toHaveBeenCalledExactlyOnceWith('gap-commit', write.mock.calls[0]?.[0].revision); | CLOSED | DEFECT=no
FACT: H-006 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-008 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-008 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-008-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: MS3b fix6 ownership Save / newer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'MS3b fix6 ownership Save / newer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "MS3b fix6 ownership Save / newer",
  "status": "fail",
  "duration": 64,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'offer', snapshot: { …(7) } } to deeply equal { kind: 'none' }",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 212,
        "column": 40,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 212
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "kind": "none",
+   "kind": "offer",
+   "snapshot": {
+     "baseLanguage": "en",
+     "capturedAt": 1794928790327,
+     "expectedUpdatedAt": "T1",
+     "key": "fix6:Save:newer",
+     "language": "uk",
+     "lines": [
+       {
+         "documentIds": [],
+         "endMs": "1000",
+         "id": "line",
+         "startMs": "0",
+         "text": "C",
+       },
+     ],
+     "revision": "fe05a975-7454-424d-b718-68a8eb4f3d83",
+   },
  }
```

codeFrame (error 1):

```text
210 |       expect(await store.read(key)).toEqual(version === 'same' ? null : c);
211 |       if (version === 'newer' && path === 'Save')
212 |         expect(f.result.current.offer).toEqual({ kind: 'none' });
    |                                        ^
213 |       if (version === 'newer' && path !== 'unmount' && path !== 'Save')
214 |         expect(f.result.current.offer).toMatchObject({ kind: 'offer', snapshot: c });
```


~~~~
FACT: H-008 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: MS3b fix6 ownership Save / newer | baseline PASS (recorded batch baseline) | 210 |       expect(await store.read(key)).toEqual(version === 'same' ? null : c); / 212 |         expect(f.result.current.offer).toEqual({ kind: 'none' }); / 214 |         expect(f.result.current.offer).toMatchObject({ kind: 'offer', snapshot: c }); | CLOSED | DEFECT=no
FACT: H-008 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-009 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-009 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-009-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP mismatch refresh from an obsolete visit cannot replace the current offer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP mismatch refresh from an obsolete visit cannot replace the current offer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP mismatch refresh from an obsolete visit cannot replace the current offer",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'offer', snapshot: { …(7) } } to match object { snapshot: { …(7) } }\n(1 matching property omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 403,
        "column": 34,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 403
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
    "snapshot": {
      "baseLanguage": "en",
      "capturedAt": 1,
      "expectedUpdatedAt": "T1",
-     "key": "gap-refresh:b",
+     "key": "gap-refresh:a",
      "language": "uk",
      "lines": [
        {
          "documentIds": [],
          "endMs": "1000",
          "id": "line",
          "startMs": "0",
-         "text": "Recovered",
+         "text": "foreign",
        },
      ],
-     "revision": "revision:gap-refresh:b:Recovered",
+     "revision": "foreign",
    },
  }
```

codeFrame (error 1):

```text
401 |   expect(f.result.current.offer).toMatchObject({ snapshot: b });
402 |   await act(async () => delayed.resolve({ ...a, revision: 'foreign', lines: [{ ...line, text: 'foreign' }] }));
403 |   expect(f.result.current.offer).toMatchObject({ snapshot: b });
    |                                  ^
404 | });
405 | 
```


~~~~
FACT: H-009 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP mismatch refresh from an obsolete visit cannot replace the current offer | baseline PASS (recorded batch baseline) | 401 |   expect(f.result.current.offer).toMatchObject({ snapshot: b }); / 403 |   expect(f.result.current.offer).toMatchObject({ snapshot: b }); | CLOSED | DEFECT=no
FACT: H-009 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-010 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-010 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-010-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP repeated Discard retires the refreshed revision rather than its predecessor

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP repeated Discard retires the refreshed revision rather than its predecessor'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP repeated Discard retires the refreshed revision rather than its predecessor",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to be called once with arguments: [ 'gap-discard', 'C' ], but got []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 421,
        "column": 17,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 421
    }
  ]
}
```

codeFrame (error 1):

```text
419 |   act(() => f.result.current.dismiss());
420 |   await act(async () => {});
421 |   expect(clear).toHaveBeenCalledExactlyOnceWith(c.key, 'C');
    |                 ^
422 |   expect(await store.read(c.key)).toBeNull();
423 | });
```


~~~~
FACT: H-010 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP repeated Discard retires the refreshed revision rather than its predecessor | baseline PASS (recorded batch baseline) | 421 |   expect(clear).toHaveBeenCalledExactlyOnceWith(c.key, 'C'); / 422 |   expect(await store.read(c.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: H-010 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-013 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-013 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-013-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended captureNow does not arm capture when resumed or unmounted

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP suspended captureNow does not arm capture when resumed or unmounted'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP suspended captureNow does not arm capture when resumed or unmounted",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"write\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st write call:\n\n    Array [\n      Object {\n        \"baseLanguage\": \"en\",\n        \"capturedAt\": 1794928796831,\n        \"expectedUpdatedAt\": \"T1\",\n        \"key\": \"user:lesson:track:a\",\n        \"language\": \"uk\",\n        \"lines\": Array [\n          Object {\n            \"documentIds\": Array [],\n            \"endMs\": \"1000\",\n            \"id\": \"line\",\n            \"startMs\": \"0\",\n            \"text\": \"suspended edit\",\n          },\n        ],\n        \"revision\": \"698599ba-62a9-4afd-91be-d7a672663159\",\n      },\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 445,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 445
    }
  ]
}
```

codeFrame (error 1):

```text
443 |     act(() => f.result.current.setCapturing(false));
444 |     act(() => f.result.current[action]([{ ...line, text: 'suspended edit' }]));
445 |     expect(f.write).not.toHaveBeenCalled();
    |                         ^
446 |     act(() => f.result.current.setCapturing(true));
447 |     act(() => rstest.advanceTimersByTime(idleCaptureMs));
```


~~~~
FACT: H-013 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended captureNow does not arm capture when resumed or unmounted | baseline PASS (recorded batch baseline) | 445 |     expect(f.write).not.toHaveBeenCalled(); | CLOSED | DEFECT=no
FACT: H-013 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-015 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-015 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-015-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP clean capture cannot retire an own write before the initial read finishes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP clean capture cannot retire an own write before the initial read finishes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP clean capture cannot retire an own write before the initial read finishes",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-initial:clean capture\",\n      \"367b7870-6c71-460c-ad09-32c082d350a6\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 344,
        "column": 23,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 344
    }
  ]
}
```

codeFrame (error 1):

```text
342 |     });
343 |     await act(async () => {});
344 |     expect(clear).not.toHaveBeenCalled();
    |                       ^
345 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1');
346 |     await act(async () => pending.resolve(null));
```


~~~~
FACT: H-015 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP clean capture cannot retire an own write before the initial read finishes | baseline PASS (recorded batch baseline) | 344 |     expect(clear).not.toHaveBeenCalled(); / 345 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1'); | CLOSED | DEFECT=no
FACT: H-015 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-025 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-025 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-025-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete read completion cannot authorize retirement in an unread visit'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete read completion cannot authorize retirement in an unread visit",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-unread:b\",\n      \"703b059d-6a3b-470c-a5f7-86be6ebb3081\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 322,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 322
    }
  ]
}
```

codeFrame (error 1):

```text
320 |   act(() => f.result.current.finalize());
321 |   await act(async () => {});
322 |   expect(clear).not.toHaveBeenCalled();
    |                     ^
323 |   await act(async () => b.resolve(null));
324 | });
```

### [F02] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalize cannot retire an own write before the initial read finishes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP finalize cannot retire an own write before the initial read finishes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP finalize cannot retire an own write before the initial read finishes",
  "status": "fail",
  "duration": 3,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-initial:finalize\",\n      \"091162f1-9cd8-4483-aa88-32dab7abf99e\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 344,
        "column": 23,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 344
    }
  ]
}
```

codeFrame (error 1):

```text
342 |     });
343 |     await act(async () => {});
344 |     expect(clear).not.toHaveBeenCalled();
    |                       ^
345 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1');
346 |     await act(async () => pending.resolve(null));
```


~~~~
FACT: H-025 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalize cannot retire an own write before the initial read finishes | baseline PASS (recorded batch baseline) | 322 |   expect(clear).not.toHaveBeenCalled(); / 344 |     expect(clear).not.toHaveBeenCalled(); / 345 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1'); | CLOSED | DEFECT=no
FACT: H-025 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-026 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-026 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-026-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalization without an own write retains the accepted recovery source

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP finalization without an own write retains the accepted recovery source'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP finalization without an own write retains the accepted recovery source",
  "status": "fail",
  "duration": 20,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-accepted\",\n      \"revision:gap-accepted:Recovered\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 379,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 379
    }
  ]
}
```

codeFrame (error 1):

```text
377 |   act(() => f.result.current.finalize());
378 |   await act(async () => {});
379 |   expect(clear).not.toHaveBeenCalled();
    |                     ^
380 |   expect(await store.read(a.key)).toEqual(a);
381 | });
```


~~~~
FACT: H-026 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalization without an own write retains the accepted recovery source | baseline PASS (recorded batch baseline) | 379 |   expect(clear).not.toHaveBeenCalled(); / 380 |   expect(await store.read(a.key)).toEqual(a); | CLOSED | DEFECT=no
FACT: H-026 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-028 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-028 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-028-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended captureNow does not arm capture when resumed or unmounted

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP suspended captureNow does not arm capture when resumed or unmounted'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP suspended captureNow does not arm capture when resumed or unmounted",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"write\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st write call:\n\n    Array [\n      Object {\n        \"baseLanguage\": \"en\",\n        \"capturedAt\": 1794928805459,\n        \"expectedUpdatedAt\": \"T1\",\n        \"key\": \"user:lesson:track:a\",\n        \"language\": \"uk\",\n        \"lines\": Array [\n          Object {\n            \"documentIds\": Array [],\n            \"endMs\": \"1000\",\n            \"id\": \"line\",\n            \"startMs\": \"0\",\n            \"text\": \"suspended edit\",\n          },\n        ],\n        \"revision\": \"3b8d498a-0e1f-491c-845e-17240ef44015\",\n      },\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 445,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 445
    }
  ]
}
```

codeFrame (error 1):

```text
443 |     act(() => f.result.current.setCapturing(false));
444 |     act(() => f.result.current[action]([{ ...line, text: 'suspended edit' }]));
445 |     expect(f.write).not.toHaveBeenCalled();
    |                         ^
446 |     act(() => f.result.current.setCapturing(true));
447 |     act(() => rstest.advanceTimersByTime(idleCaptureMs));
```

### [F02] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended noteChange does not arm capture when resumed or unmounted

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP suspended noteChange does not arm capture when resumed or unmounted'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP suspended noteChange does not arm capture when resumed or unmounted",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"write\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st write call:\n\n    Array [\n      Object {\n        \"baseLanguage\": \"en\",\n        \"capturedAt\": 1794928895471,\n        \"expectedUpdatedAt\": \"T1\",\n        \"key\": \"user:lesson:track:a\",\n        \"language\": \"uk\",\n        \"lines\": Array [\n          Object {\n            \"documentIds\": Array [],\n            \"endMs\": \"1000\",\n            \"id\": \"line\",\n            \"startMs\": \"0\",\n            \"text\": \"suspended edit\",\n          },\n        ],\n        \"revision\": \"a0a19991-6419-42ac-ad27-3c298dc9298e\",\n      },\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 449,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 449
    }
  ]
}
```

codeFrame (error 1):

```text
447 |     act(() => rstest.advanceTimersByTime(idleCaptureMs));
448 |     f.unmount();
449 |     expect(f.write).not.toHaveBeenCalled();
    |                         ^
450 |   });
451 | }
```


~~~~
FACT: H-028 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended captureNow does not arm capture when resumed or unmounted; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended noteChange does not arm capture when resumed or unmounted | baseline PASS (recorded batch baseline) | 445 |     expect(f.write).not.toHaveBeenCalled(); / 449 |     expect(f.write).not.toHaveBeenCalled(); | CLOSED | DEFECT=no
FACT: H-028 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: H-029 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-029 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-029-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP switching a loaded visit hides its offer until the new store read resolves

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP switching a loaded visit hides its offer until the new store read resolves'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP switching a loaded visit hides its offer until the new store read resolves",
  "status": "fail",
  "duration": 16,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'offer', snapshot: { …(7) } } to deeply equal { kind: 'none' }",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 281,
        "column": 34,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 281
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "kind": "none",
+   "kind": "offer",
+   "snapshot": {
+     "baseLanguage": "en",
+     "capturedAt": 1,
+     "expectedUpdatedAt": "T1",
+     "key": "gap-switch:a",
+     "language": "uk",
+     "lines": [
+       {
+         "documentIds": [],
+         "endMs": "1000",
+         "id": "line",
+         "startMs": "0",
+         "text": "Recovered",
+       },
+     ],
+     "revision": "revision:gap-switch:a:Recovered",
+   },
  }
```

codeFrame (error 1):

```text
279 |   f.rerender({ key: 'gap-switch:b', source: other });
280 |   expect(f.result.current.isLoaded).toBe(false);
281 |   expect(f.result.current.offer).toEqual({ kind: 'none' });
    |                                  ^
282 |   await act(async () => next.resolve(null));
283 |   expect(f.result.current.isLoaded).toBe(true);
```


~~~~
FACT: H-029 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP switching a loaded visit hides its offer until the new store read resolves | baseline PASS (recorded batch baseline) | 280 |   expect(f.result.current.isLoaded).toBe(false); / 281 |   expect(f.result.current.offer).toEqual({ kind: 'none' }); / 283 |   expect(f.result.current.isLoaded).toBe(true); | CLOSED | DEFECT=no
FACT: H-029 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: X-007 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: X-007 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-007-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for null envelope

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for null envelope'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for null envelope",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise rejected \"TypeError: Cannot convert undefined or nu…\" instead of resolving",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 96,
        "column": 56,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

codeFrame (error 1):

```text
94 |   ])('rejects the whole list for $name', async ({ response }) => {
95 |     server.use(http.get(api, () => HttpResponse.json(response)));
96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid);
   |                                                        ^
97 |   });
98 | 
```


~~~~
FACT: X-007 | src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for null envelope | baseline PASS (recorded batch baseline) | 96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid); | CLOSED | DEFECT=no
FACT: X-007 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
NEXT: X-012 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: Added detached inventory lifecycle/error tests (U rows, X-020) and field integration tests (L-001/F-002/F-009) at the named public seams. Inventory rejection test observes unhandledRejection in a named assertion across an event-loop barrier; no sleep-based scheduling. NEXT: Finish hook mutation batch before new unmutated baseline.
FACT: X-012 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-012-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read reject cannot replace the current visit offer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete initial read reject cannot replace the current visit offer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete initial read reject cannot replace the current visit offer",
  "status": "fail",
  "duration": 15,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'none' } to match object { snapshot: { …(7) } }\n(1 matching property omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 303,
        "column": 36,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 303
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "snapshot": {
-     "baseLanguage": "en",
-     "capturedAt": 1,
-     "expectedUpdatedAt": "T1",
-     "key": "gap-obsolete-reject:b",
-     "language": "uk",
-     "lines": [
-       {
-         "documentIds": [],
-         "endMs": "1000",
-         "id": "line",
-         "startMs": "0",
-         "text": "Recovered",
-       },
-     ],
-     "revision": "revision:gap-obsolete-reject:b:Recovered",
-   },
+   "kind": "none",
  }
```

codeFrame (error 1):

```text
301 |     });
302 |     expect(read).toHaveBeenCalledTimes(2);
303 |     expect(f.result.current.offer).toMatchObject({ snapshot: b });
    |                                    ^
304 |     expect(f.result.current.isLoaded).toBe(true);
305 |   });
```


~~~~
FACT: X-012 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read reject cannot replace the current visit offer | baseline PASS (recorded batch baseline) | 302 |     expect(read).toHaveBeenCalledTimes(2); / 303 |     expect(f.result.current.offer).toMatchObject({ snapshot: b }); / 304 |     expect(f.result.current.isLoaded).toBe(true); | CLOSED | DEFECT=no
FACT: X-012 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: X-013 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: X-013 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-013-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended noteChange does not arm capture when resumed or unmounted

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP suspended noteChange does not arm capture when resumed or unmounted'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP suspended noteChange does not arm capture when resumed or unmounted",
  "status": "fail",
  "duration": 19,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"write\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st write call:\n\n    Array [\n      Object {\n        \"baseLanguage\": \"en\",\n        \"capturedAt\": 1794928903003,\n        \"expectedUpdatedAt\": \"T1\",\n        \"key\": \"user:lesson:track:a\",\n        \"language\": \"uk\",\n        \"lines\": Array [\n          Object {\n            \"documentIds\": Array [],\n            \"endMs\": \"1000\",\n            \"id\": \"line\",\n            \"startMs\": \"0\",\n            \"text\": \"suspended edit\",\n          },\n        ],\n        \"revision\": \"cad4080c-efe9-40fb-98ab-5bff560b20d4\",\n      },\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 449,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 449
    }
  ]
}
```

codeFrame (error 1):

```text
447 |     act(() => rstest.advanceTimersByTime(idleCaptureMs));
448 |     f.unmount();
449 |     expect(f.write).not.toHaveBeenCalled();
    |                         ^
450 |   });
451 | }
```


~~~~
FACT: X-013 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended noteChange does not arm capture when resumed or unmounted | baseline PASS (recorded batch baseline) | 449 |     expect(f.write).not.toHaveBeenCalled(); | CLOSED | DEFECT=no
FACT: X-013 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
NEXT: X-014 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: X-014 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-014-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP a vanished mismatch closes the stale own recovery prompt

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP a vanished mismatch closes the stale own recovery prompt'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP a vanished mismatch closes the stale own recovery prompt",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'offer', snapshot: { …(7) } } to deeply equal { kind: 'none' }",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 435,
        "column": 34,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 435
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "kind": "none",
+   "kind": "offer",
+   "snapshot": {
+     "baseLanguage": "en",
+     "capturedAt": 1,
+     "expectedUpdatedAt": "T1",
+     "key": "gap-vanished",
+     "language": "uk",
+     "lines": [
+       {
+         "documentIds": [],
+         "endMs": "1000",
+         "id": "line",
+         "startMs": "0",
+         "text": "Recovered",
+       },
+     ],
+     "revision": "revision:gap-vanished:Recovered",
+   },
  }
```

codeFrame (error 1):

```text
433 |   act(() => f.result.current.captureNow([line]));
434 |   await act(async () => {});
435 |   expect(f.result.current.offer).toEqual({ kind: 'none' });
    |                                  ^
436 | });
437 | 
```


~~~~
FACT: X-014 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP a vanished mismatch closes the stale own recovery prompt | baseline PASS (recorded batch baseline) | 435 |   expect(f.result.current.offer).toEqual({ kind: 'none' }); | CLOSED | DEFECT=no
FACT: X-014 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: Unmutated hook batch `npx rstest run` (files named in raw log), exit 0; counts {
    "testFiles": 2,
    "failedFiles": 0,
    "tests": 78,
    "failedTests": 0,
    "passedTests": 78,
    "skippedTests": 0,
    "todoTests": 0
  }; raw `fe-a-hook-baseline.txt`.
FACT: Unmutated inventory batch `npx rstest run` (files named in raw log), exit 0; counts {
    "testFiles": 3,
    "failedFiles": 0,
    "tests": 30,
    "failedTests": 0,
    "passedTests": 30,
    "skippedTests": 0,
    "todoTests": 0
  }; raw `fe-a-inventory-baseline.txt`.
NEXT: U-001 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: U-001 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-U-001-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP disabled detached inventory performs no storage IO

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP disabled detached inventory performs no storage IO'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP disabled detached inventory performs no storage IO",
  "status": "fail",
  "duration": 24,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"keys\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st keys call:\n\n    Array []\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 18,
        "column": 20,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 18
    }
  ]
}
```

codeFrame (error 1):

```text
16 |   const f = renderHook(() => useDetachedTranscriptRecovery('lesson', { authorId: 'author', store }, tabs, false));
17 |   await act(async () => {});
18 |   expect(keys).not.toHaveBeenCalled();
   |                    ^
19 |   expect(read).not.toHaveBeenCalled();
20 |   expect(f.result.current).toEqual([]);
```


~~~~
FACT: U-001 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP disabled detached inventory performs no storage IO | baseline PASS (recorded batch baseline) | 18 |   expect(keys).not.toHaveBeenCalled(); / 19 |   expect(read).not.toHaveBeenCalled(); / 20 |   expect(f.result.current).toEqual([]); | CLOSED | DEFECT=no
FACT: U-001 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
NEXT: U-003 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: U-003 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-U-003-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory ignores a read result belonging to a different listed key

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP detached inventory ignores a read result belonging to a different listed key'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP detached inventory ignores a read result belonging to a different listed key",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { instanceId: 'track:en', …(2) } ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 30,
        "column": 28,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 30
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   {
+     "instanceId": "track:en",
+     "legacy": false,
+     "snapshot": {
+       "capturedAt": 1,
+       "expectedUpdatedAt": "T1",
+       "key": "author:lesson:track:gone",
+       "language": "en",
+       "lines": [
+         {
+           "documentIds": [],
+           "endMs": "1000",
+           "id": "line",
+           "startMs": "0",
+           "text": "A",
+         },
+       ],
+       "revision": "A",
+     },
+   },
+ ]
```

codeFrame (error 1):

```text
28 |   await act(async () => {});
29 |   expect(read).toHaveBeenCalledExactlyOnceWith('author:lesson:track:listed');
30 |   expect(f.result.current).toEqual([]);
   |                            ^
31 | });
32 | 
```


~~~~
FACT: U-003 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory ignores a read result belonging to a different listed key | baseline PASS (recorded batch baseline) | 29 |   expect(read).toHaveBeenCalledExactlyOnceWith('author:lesson:track:listed'); / 30 |   expect(f.result.current).toEqual([]); | CLOSED | DEFECT=no
FACT: U-003 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
NEXT: U-004 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: U-004 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-U-004-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP obsolete detached inventory cannot erase the completed current inventory

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete detached inventory cannot erase the completed current inventory'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete detached inventory cannot erase the completed current inventory",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [] to deeply equal [ { …(6) } ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 44,
        "column": 59,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 44
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- [
-   {
-     "capturedAt": 1,
-     "expectedUpdatedAt": "T1",
-     "key": "author:lesson:track:gone",
-     "language": "en",
-     "lines": [
-       {
-         "documentIds": [],
-         "endMs": "1000",
-         "id": "line",
-         "startMs": "0",
-         "text": "A",
-       },
-     ],
-     "revision": "A",
-   },
- ]
+ []
```

codeFrame (error 1):

```text
42 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);
43 |   await act(async () => old.resolve([]));
44 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);
   |                                                           ^
45 | });
46 | 
```


~~~~
FACT: U-004 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP obsolete detached inventory cannot erase the completed current inventory | baseline PASS (recorded batch baseline) | 42 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]); / 44 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]); | CLOSED | DEFECT=no
FACT: U-004 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
NEXT: U-005 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: U-005 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-U-005-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after store changes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP detached inventory hides stale same-lesson offers after store changes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP detached inventory hides stale same-lesson offers after store changes",
  "status": "fail",
  "duration": 18,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { instanceId: 'track:en', …(2) } ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 60,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 60
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   {
+     "instanceId": "track:en",
+     "legacy": false,
+     "snapshot": {
+       "capturedAt": 1,
+       "expectedUpdatedAt": "T1",
+       "key": "author:lesson:track:gone",
+       "language": "en",
+       "lines": [
+         {
+           "documentIds": [],
+           "endMs": "1000",
+           "id": "line",
+           "startMs": "0",
+           "text": "A",
+         },
+       ],
+       "revision": "A",
+     },
+   },
+ ]
```

codeFrame (error 1):

```text
58 |     f.rerender({ source: change === 'store' ? other : store,
59 |       currentTabs: change === 'tabs' ? [{ ...tabs[0], language: 'es' }] : [...tabs], enabled: change !== 'disabled' });
60 |     expect(f.result.current).toEqual([]);
   |                              ^
61 |     await act(async () => pending.resolve([]));
62 |   });
```

### [F02] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after tabs changes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP detached inventory hides stale same-lesson offers after tabs changes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP detached inventory hides stale same-lesson offers after tabs changes",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { instanceId: 'track:en', …(2) } ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 60,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 60
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   {
+     "instanceId": "track:en",
+     "legacy": false,
+     "snapshot": {
+       "capturedAt": 1,
+       "expectedUpdatedAt": "T1",
+       "key": "author:lesson:track:gone",
+       "language": "en",
+       "lines": [
+         {
+           "documentIds": [],
+           "endMs": "1000",
+           "id": "line",
+           "startMs": "0",
+           "text": "A",
+         },
+       ],
+       "revision": "A",
+     },
+   },
+ ]
```

codeFrame (error 1):

```text
58 |     f.rerender({ source: change === 'store' ? other : store,
59 |       currentTabs: change === 'tabs' ? [{ ...tabs[0], language: 'es' }] : [...tabs], enabled: change !== 'disabled' });
60 |     expect(f.result.current).toEqual([]);
   |                              ^
61 |     await act(async () => pending.resolve([]));
62 |   });
```

### [F03] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after disabled changes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP detached inventory hides stale same-lesson offers after disabled changes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP detached inventory hides stale same-lesson offers after disabled changes",
  "status": "fail",
  "duration": 3,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { instanceId: 'track:en', …(2) } ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 60,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 60
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   {
+     "instanceId": "track:en",
+     "legacy": false,
+     "snapshot": {
+       "capturedAt": 1,
+       "expectedUpdatedAt": "T1",
+       "key": "author:lesson:track:gone",
+       "language": "en",
+       "lines": [
+         {
+           "documentIds": [],
+           "endMs": "1000",
+           "id": "line",
+           "startMs": "0",
+           "text": "A",
+         },
+       ],
+       "revision": "A",
+     },
+   },
+ ]
```

codeFrame (error 1):

```text
58 |     f.rerender({ source: change === 'store' ? other : store,
59 |       currentTabs: change === 'tabs' ? [{ ...tabs[0], language: 'es' }] : [...tabs], enabled: change !== 'disabled' });
60 |     expect(f.result.current).toEqual([]);
   |                              ^
61 |     await act(async () => pending.resolve([]));
62 |   });
```


~~~~
FACT: U-005 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after store changes; src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after tabs changes; src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after disabled changes | baseline PASS (recorded batch baseline) | 60 |     expect(f.result.current).toEqual([]); / 60 |     expect(f.result.current).toEqual([]); / 60 |     expect(f.result.current).toEqual([]); | CLOSED | DEFECT=no
FACT: U-005 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
NEXT: U-009 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: U-009 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-U-009-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP one refused detached read does not suppress a different valid recovery

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP one refused detached read does not suppress a different valid recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP one refused detached read does not suppress a different valid recovery",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [] to deeply equal [ { …(6) } ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 72,
        "column": 59,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 72
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- [
-   {
-     "capturedAt": 1,
-     "expectedUpdatedAt": "T1",
-     "key": "author:lesson:track:gone",
-     "language": "en",
-     "lines": [
-       {
-         "documentIds": [],
-         "endMs": "1000",
-         "id": "line",
-         "startMs": "0",
-         "text": "A",
-       },
-     ],
-     "revision": "A",
-   },
- ]
+ []
```

codeFrame (error 1):

```text
70 |   await act(async () => {});
71 |   expect(read).toHaveBeenCalledTimes(2);
72 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);
   |                                                           ^
73 | });
74 | 
```


~~~~
FACT: U-009 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP one refused detached read does not suppress a different valid recovery | baseline PASS (recorded batch baseline) | 71 |   expect(read).toHaveBeenCalledTimes(2); / 72 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]); | CLOSED | DEFECT=no
FACT: U-009 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
NEXT: L-001 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/LessonDraftField.test.tsx`. Production SHA-256 c5ab86ff38e890b866e6e779d977065143be2a46758d23b8cbb07c0aebdc631c.
FACT: L-001 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-L-001-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/LessonDraftField.test.tsx :: GAP a missing shown track reports not found after Regenerate

repro:

```bash
npx rstest 'src/modules/transcript/ui/LessonDraftField.test.tsx' --testNamePattern 'GAP a missing shown track reports not found after Regenerate'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/LessonDraftField.test.tsx",
  "project": "dom",
  "fullName": "GAP a missing shown track reports not found after Regenerate",
  "status": "fail",
  "duration": 31,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 'Recognition service is temporarily un…' to be 'Save the course and upload a video be…' // Object.is equality",
      "topFrame": {
        "file": "src/modules/transcript/ui/LessonDraftField.test.tsx",
        "line": 273,
        "column": 58,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/LessonDraftField.test.tsx",
      "line": 273
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- Save the course and upload a video before generating.
+ Recognition service is temporarily unavailable. Try again later.
```

codeFrame (error 1):

```text
271 |   renderField(draft('ready'), draftActions({ regenerate: async () => 'track_not_found' }));
272 |   await userEvent.setup().click(screen.getByRole('button', { name: 'Regenerate transcript' }));
273 |   expect((await screen.findByRole('alert')).textContent).toBe('Save the course and upload a video before generating.');
    |                                                          ^
274 | });
275 | 
```


~~~~
FACT: L-001 | src/modules/transcript/ui/LessonDraftField.test.tsx :: GAP a missing shown track reports not found after Regenerate | baseline PASS (recorded batch baseline) | 273 |   expect((await screen.findByRole('alert')).textContent).toBe('Save the course and upload a video before generating.'); | CLOSED | DEFECT=no
FACT: L-001 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/LessonDraftField.tsx` exit=0.
NEXT: F-002 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/LessonTranscriptField.test.tsx`. Production SHA-256 c2c97d5c0e1af77d2814534ab1bae82b0d8f7fc640cc5574003efff0a07f606f.
FACT: F-002 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-F-002-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/LessonTranscriptField.test.tsx :: GAP first transcript load then renamed-language load retains the dirty editor instance

repro:

```bash
npx rstest 'src/modules/transcript/ui/LessonTranscriptField.test.tsx' --testNamePattern 'GAP first transcript load then renamed-language load retains the dirty editor instance'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
  "project": "dom",
  "fullName": "GAP first transcript load then renamed-language load retains the dirty editor instance",
  "status": "fail",
  "duration": 130,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find a label with the text of: Текст репліки 1\n\nIgnored nodes: comments, script, style\n<body>\n  <div />\n</body>",
      "topFrame": {
        "file": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
        "line": 26,
        "column": 10,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
      "line": 26
    }
  ]
}
```

codeFrame (error 1):

```text
24 |   loading = true;
25 |   f.rerender(<LessonTranscriptField {...props} fixedLanguage="es" />);
26 |   expect(screen.getByLabelText('Текст репліки 1')).toBe(text);
   |          ^
27 |   expect(screen.getByDisplayValue('Unsaved')).toBe(text);
28 | });
```


~~~~
FACT: F-002 | src/modules/transcript/ui/LessonTranscriptField.test.tsx :: GAP first transcript load then renamed-language load retains the dirty editor instance | baseline PASS (recorded batch baseline) | 26 |   expect(screen.getByLabelText('Текст репліки 1')).toBe(text); / 27 |   expect(screen.getByDisplayValue('Unsaved')).toBe(text); | CLOSED | DEFECT=no
FACT: F-002 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/LessonTranscriptField.tsx` exit=0.
NEXT: F-009 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/LessonTranscriptField.test.tsx`. Production SHA-256 c2c97d5c0e1af77d2814534ab1bae82b0d8f7fc640cc5574003efff0a07f606f.
FACT: F-009 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-F-009-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/LessonTranscriptField.test.tsx :: GAP fixed tab language keeps the returned transcript token attached to its real base language

repro:

```bash
npx rstest 'src/modules/transcript/ui/LessonTranscriptField.test.tsx' --testNamePattern 'GAP fixed tab language keeps the returned transcript token attached to its real base language'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
  "project": "dom",
  "fullName": "GAP fixed tab language keeps the returned transcript token attached to its real base language",
  "status": "fail",
  "duration": 46,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to match object { language: 'es', …(2) }\n(4 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
        "line": 42,
        "column": 55,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
      "line": 42
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "baseLanguage": "en",
+   "baseLanguage": "es",
    "expectedUpdatedAt": "T1",
    "language": "es",
  }
```

codeFrame (error 1):

```text
40 |   fireEvent.focusOut(text);
41 |   await act(async () => {});
42 |   expect(await store.read('author:lesson:track:one')).toMatchObject({ language: 'es', baseLanguage: 'en', expectedUpdatedAt: 'T1' });
   |                                                       ^
43 |   fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
44 |   await act(async () => {});
```


~~~~
FACT: F-009 | src/modules/transcript/ui/LessonTranscriptField.test.tsx :: GAP fixed tab language keeps the returned transcript token attached to its real base language | baseline PASS (recorded batch baseline) | 42 |   expect(await store.read('author:lesson:track:one')).toMatchObject({ language: 'es', baseLanguage: 'en', expectedUpdatedAt: 'T1' }); | CLOSED | DEFECT=no
FACT: F-009 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/LessonTranscriptField.tsx` exit=0.
NEXT: X-020 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: X-020 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-020-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "",
  "status": "fail",
  "duration": 51,
  "retryCount": 0,
  "errors": [
    {
      "type": "unhandledRejection",
      "message": "inventory refusal",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.ts",
        "line": 53,
        "column": 15,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.ts",
      "line": 53
    }
  ]
}
```

codeFrame (error 1):

```text
51 |           });
52 |       } catch {
53 |         throw new Error('inventory refusal');
   |               ^
54 |       }
55 |     })();
```

console:

```text
[stderr] error: unhandledRejection: inventory refusal
    at /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts:53:15
```

### [F02] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP inventory refusal remains best effort without an escaped rejection

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP inventory refusal remains best effort without an escaped rejection'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP inventory refusal remains best effort without an escaped rejection",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ …(1) ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 89,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 89
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   unhandledRejection {
+     "message": "inventory refusal",
+     "name": "unhandledRejection",
+   },
+ ]
```

codeFrame (error 1):

```text
87 |     expect(keys).toHaveBeenCalledTimes(1);
88 |     expect(f.result.current).toEqual([]);
89 |     expect(escaped).toEqual([]);
   |                     ^
90 |   } finally {
91 |     process.off('unhandledRejection', onRejection);
```

console:

```text
[stderr] error: unhandledRejection: inventory refusal
    at /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts:53:15
```


~~~~
FACT: X-020 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx; src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP inventory refusal remains best effort without an escaped rejection | baseline PASS (recorded batch baseline) | 87 |     expect(keys).toHaveBeenCalledTimes(1); / 88 |     expect(f.result.current).toEqual([]); / 89 |     expect(escaped).toEqual([]); | CLOSED | DEFECT=no
FACT: X-020 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
FACT: Inventory/field mutation batch all 9 CLOSED, including X-020 with named expect(escaped).toEqual([]) failure (global file failure alone not counted). Total CLOSED=50. Added editor state-transition, save-lock, ancestry, ASR/recovery and deferred-ownership behavior tests for remaining E rows plus X-017. NEXT: Unmutated editor baseline; any correct behavior failure is DEFECT and STOP + QUESTION.
FACT: First editor baseline exit 1: two new test fixture selectors used non-existent label «Відкинути зміни»; actual visible Discard label is «Не відновлювати». These are test authoring errors, not product defects; corrected selectors and required ASR button lookup rather than optional lookup. All other new editor behavior assertions passed. NEXT: Re-run editor baseline.
FACT: Unmutated editor baseline exit 0, counts {
    "testFiles": 1,
    "failedFiles": 0,
    "tests": 86,
    "failedTests": 0,
    "passedTests": 86,
    "skippedTests": 0,
    "todoTests": 0
  }; raw fe-a-editor-baseline-fixed.txt. All remaining tests now reach their intended behaviors. NEXT: Exact E-row and X-017 mutations individually, restoring bytes after each.
NEXT: E-001 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-001 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-001-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a source acquired by a mounted tab is removed from captured cleanup ancestry

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a source acquired by a mounted tab is removed from captured cleanup ancestry'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a source acquired by a mounted tab is removed from captured cleanup ancestry",
  "status": "fail",
  "duration": 25,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { key: 'acquired-source', …(1) } ] to be undefined",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1596,
        "column": 79,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1596
    }
  ]
}
```

diff:

```diff
- Expected:
undefined

+ Received:
[
  {
    "key": "acquired-source",
    "revision": "revision:acquired-source",
  },
]
```

codeFrame (error 1):

```text
1594 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1595 |   await act(async () => {});
1596 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined();
     |                                                                               ^
1597 |   view.rerender(<TranscriptEditor {...props} detachedSnapshots={offers} mountedHistoryKeys={[]} />);
1598 |   clickSave();
```


~~~~
FACT: E-001 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a source acquired by a mounted tab is removed from captured cleanup ancestry | baseline PASS (recorded batch baseline) | 1596 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined(); | CLOSED | DEFECT=no
FACT: E-001 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-002 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-002 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-002-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP source cleanup rechecks newly mounted ownership after its deferred read

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP source cleanup rechecks newly mounted ownership after its deferred read'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP source cleanup rechecks newly mounted ownership after its deferred read",
  "status": "fail",
  "duration": 28,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ [ 'newly-mounted-source', …(1) ] ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1622,
        "column": 66,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1622
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   [
+     "newly-mounted-source",
+     "revision:newly-mounted-source",
+   ],
+ ]
```

codeFrame (error 1):

```text
1620 |   view.rerender(<TranscriptEditor {...props} history={{ key: props.history.key, store }} detachedSnapshots={offers} mountedHistoryKeys={[source.key]} />);
1621 |   await act(async () => pending.resolve(source));
1622 |   expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]);
     |                                                                  ^
1623 |   expect(await memory.read(source.key)).toEqual(source);
1624 | });
```


~~~~
FACT: E-002 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP source cleanup rechecks newly mounted ownership after its deferred read | baseline PASS (recorded batch baseline) | 1622 |   expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]); / 1623 |   expect(await memory.read(source.key)).toEqual(source); | CLOSED | DEFECT=no
FACT: E-002 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-003 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-003 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-003-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a completed Save lets the next canonical server token become the next edit base

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a completed Save lets the next canonical server token become the next edit base'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a completed Save lets the next canonical server token become the next edit base",
  "status": "fail",
  "duration": 37,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected last \"rstest.fn()\" call to have been called with [ 'uk', 'T2', Any<Array>, 'uk' ], but got [ 'uk', 'T1', [ { …(5) } ], 'uk' ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1420,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1420
    }
  ]
}
```

codeFrame (error 1):

```text
1418 |   clickSave();
1419 |   await act(async () => {});
1420 |   expect(props.actions.save).toHaveBeenLastCalledWith('uk', 'T2', expect.any(Array), 'uk');
     |                              ^
1421 | });
1422 | 
```


~~~~
FACT: E-003 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a completed Save lets the next canonical server token become the next edit base | baseline PASS (recorded batch baseline) | 1420 |   expect(props.actions.save).toHaveBeenLastCalledWith('uk', 'T2', expect.any(Array), 'uk'); | CLOSED | DEFECT=no
FACT: E-003 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-005 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-005 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-005-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only base language

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a clean editor adopts a server change to only base language'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a clean editor adopts a server change to only base language",
  "status": "fail",
  "duration": 24,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to be called with arguments: [ Array(4) ], but got [ [ 'uk', 'T1', [ { …(5) } ], 'uk' ] ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1403,
        "column": 32,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1403
    }
  ]
}
```

codeFrame (error 1):

```text
1401 |     clickSave();
1402 |     await act(async () => {});
1403 |     expect(props.actions.save).toHaveBeenCalledWith('uk', change === 'token' ? 'T2' : 'T1',
     |                                ^
1404 |       expect.arrayContaining([expect.objectContaining({ text: 'Edited' })]), change === 'base language' ? 'en' : 'uk');
1405 |   });
```


~~~~
FACT: E-005 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only base language | baseline PASS (recorded batch baseline) | 1403 |     expect(props.actions.save).toHaveBeenCalledWith('uk', change === 'token' ? 'T2' : 'T1', | CLOSED | DEFECT=no
FACT: E-005 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-006 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-006 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-006-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only token

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a clean editor adopts a server change to only token'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a clean editor adopts a server change to only token",
  "status": "fail",
  "duration": 25,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to be called with arguments: [ Array(4) ], but got [ [ 'uk', 'T1', [ { …(5) } ], 'uk' ] ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1403,
        "column": 32,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1403
    }
  ]
}
```

codeFrame (error 1):

```text
1401 |     clickSave();
1402 |     await act(async () => {});
1403 |     expect(props.actions.save).toHaveBeenCalledWith('uk', change === 'token' ? 'T2' : 'T1',
     |                                ^
1404 |       expect.arrayContaining([expect.objectContaining({ text: 'Edited' })]), change === 'base language' ? 'en' : 'uk');
1405 |   });
```


~~~~
FACT: E-006 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only token | baseline PASS (recorded batch baseline) | 1403 |     expect(props.actions.save).toHaveBeenCalledWith('uk', change === 'token' ? 'T2' : 'T1', | CLOSED | DEFECT=no
FACT: E-006 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-007 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-007 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-007-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only segments

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a clean editor adopts a server change to only segments'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a clean editor adopts a server change to only segments",
  "status": "fail",
  "duration": 24,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <textarea tabindex=\"0\" …(5)></textarea> to have property \"value\" with value 'New server'",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1399,
        "column": 81,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1399
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- New server
+ Перша
```

codeFrame (error 1):

```text
1397 |       expectedUpdatedAt={change === 'token' ? 'T2' : 'T1'}
1398 |       segments={change === 'segments' ? [{ ...segments[0]!, text: 'New server' }] : segments} />);
1399 |     if (change === 'segments') expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'New server');
     |                                                                                 ^
1400 |     fireEvent.change(screen.getByLabelText('Текст репліки 1'), { target: { value: 'Edited' } });
1401 |     clickSave();
```


~~~~
FACT: E-007 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only segments | baseline PASS (recorded batch baseline) | 1399 |     if (change === 'segments') expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'New server'); | CLOSED | DEFECT=no
FACT: E-007 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-008 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-008 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-008-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix 5 a clean track rename followed by editing is not a dirty track rename

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'MS3b fix 5 a clean track rename followed by editing is not a dirty track rename'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "MS3b fix 5 a clean track rename followed by editing is not a dirty track rename",
  "status": "fail",
  "duration": 22,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <p></p> to be null",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 628,
        "column": 84,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 628
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
<p>
  Мова доріжки тепер en; рядки буде збережено як en.
</p>
```

codeFrame (error 1):

```text
626 |     target: { value: 'Edited after rename' },
627 |   });
628 |   expect(screen.queryByText('Мова доріжки тепер en; рядки буде збережено як en.')).toBeNull();
    |                                                                                    ^
629 | });
630 | 
```


~~~~
FACT: E-008 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix 5 a clean track rename followed by editing is not a dirty track rename | baseline PASS (recorded batch baseline) | 628 |   expect(screen.queryByText('Мова доріжки тепер en; рядки буде збережено як en.')).toBeNull(); | CLOSED | DEFECT=no
FACT: E-008 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-013 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-013 mutation exit=0; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-013-mutation.txt`.
~~~~text
## Failures

No test failures reported.

~~~~
FACT: E-013 | frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx | baseline PASS (recorded batch baseline) | NO ATTRIBUTABLE ASSERTION | REMAINING | DEFECT=no
FACT: E-013 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-017 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-017 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-017-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a pending Save locks fields, documents, line controls and repeat Save

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a pending Save locks fields, documents, line controls and repeat Save'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a pending Save locks fields, documents, line controls and repeat Save",
  "status": "fail",
  "duration": 27,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <input tabindex=\"0\" …(7)></input> to have property \"readOnly\" with value true",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1444,
        "column": 19,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1444
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- true
+ false
```

codeFrame (error 1):

```text
1442 |   for (const label of ['Початок репліки 1, мс', 'Кінець репліки 1, мс', 'Текст репліки 1']) {
1443 |     const field = screen.getByLabelText(label);
1444 |     expect(field).toHaveProperty('readOnly', true);
     |                   ^
1445 |     fireEvent.change(field, { target: { value: '99' } });
1446 |   }
```


~~~~
FACT: E-017 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a pending Save locks fields, documents, line controls and repeat Save | baseline PASS (recorded batch baseline) | 1444 |     expect(field).toHaveProperty('readOnly', true); | CLOSED | DEFECT=no
FACT: E-017 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-030 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
DECISION: Scope is test-only GAP repair on disposable copy HEAD 5a569653baed6532a39f8a12f0ef01c4da6354b1; owner 2026-11-17 brief authorizes exact production mutations only for comparison, never production fixes. No agents launched (0), returned (0); author verification only, no independent review claim. Existing owner working file is sole task record; no repo docs/product changes authorized.
FACT: Input contract /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-fe-a.md SHA-256=b3b4484fe1bc0eb01df38a65907aa711cb5f69778a1c5fd7e3635c1b988b8324. Exact patch per ID extracted unchanged from source journal CONFIRM entry into fe-a-gaps.json; runner applies one patch with git apply -p0 --unidiff-zero, saves raw run and restores original bytes in finally.
FACT: Input contract /Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md SHA-256=f27495ff02f70e3a2023c1a29295df62df786a48369933d7d8ed8fd68cebeb71. Exact patch per ID extracted unchanged from source journal CONFIRM entry into fe-a-gaps.json; runner applies one patch with git apply -p0 --unidiff-zero, saves raw run and restores original bytes in finally.
FACT: E-030 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-030-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry distinguishes two revisions of the same key within its eight-source limit

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP ancestry distinguishes two revisions of the same key within its eight-source limit'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP ancestry distinguishes two revisions of the same key within its eight-source limit",
  "status": "fail",
  "duration": 28,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ …(8) ] to deeply equal [ …(8) ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1495,
        "column": 79,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1495
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -2,14 +2,10 @@
    {
      "key": "source",
      "revision": "new",
    },
    {
-     "key": "source",
-     "revision": "old",
-   },
-   {
      "key": "a",
      "revision": "a1",
    },
    {
      "key": "b",
@@ -28,7 +24,11 @@
      "revision": "e1",
    },
    {
      "key": "f",
      "revision": "f1",
+   },
+   {
+     "key": "overflow",
+     "revision": "o1",
    },
  ]
```

codeFrame (error 1):

```text
1493 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1494 |   await act(async () => {});
1495 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toEqual([
     |                                                                               ^
1496 |     { key: 'source', revision: 'new' }, { key: 'source', revision: 'old' },
1497 |     { key: 'a', revision: 'a1' }, { key: 'b', revision: 'b1' }, { key: 'c', revision: 'c1' },
```


~~~~
FACT: E-030 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry distinguishes two revisions of the same key within its eight-source limit | baseline PASS (recorded batch baseline) | 1495 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toEqual([ | CLOSED | DEFECT=no
FACT: E-030 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-031 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-031 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-031-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry distinguishes two revisions of the same key within its eight-source limit

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP ancestry distinguishes two revisions of the same key within its eight-source limit'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP ancestry distinguishes two revisions of the same key within its eight-source limit",
  "status": "fail",
  "duration": 27,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ …(8) ] to deeply equal [ …(8) ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1495,
        "column": 79,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1495
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,13 +1,13 @@
  [
    {
      "key": "source",
-     "revision": "new",
+     "revision": "old",
    },
    {
      "key": "source",
-     "revision": "old",
+     "revision": "new",
    },
    {
      "key": "a",
      "revision": "a1",
    },
```

codeFrame (error 1):

```text
1493 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1494 |   await act(async () => {});
1495 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toEqual([
     |                                                                               ^
1496 |     { key: 'source', revision: 'new' }, { key: 'source', revision: 'old' },
1497 |     { key: 'a', revision: 'a1' }, { key: 'b', revision: 'b1' }, { key: 'c', revision: 'c1' },
```

### [F02] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry dedup retains the full restored snapshot for content-aware cleanup

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP ancestry dedup retains the full restored snapshot for content-aware cleanup'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP ancestry dedup retains the full restored snapshot for content-aware cleanup",
  "status": "fail",
  "duration": 16,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to deeply equal { key: 'duplicate-source', …(7) }",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1514,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1514
    }
  ]
}
```

diff:

```diff
- Expected:
{
  "acceptedSource": [
    {
      "key": "duplicate-source",
      "revision": "revision:duplicate-source",
    },
  ],
  "baseLanguage": "uk",
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "duplicate-source",
  "language": "uk",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "70000000-0000-4000-8000-000000000001",
      "startMs": "0",
      "text": "Changed with same revision",
    },
  ],
  "revision": "revision:duplicate-source",
}

+ Received:
null
```

codeFrame (error 1):

```text
1512 |   clickSave();
1513 |   await act(async () => {});
1514 |   expect(await props.history.store.read(source.key)).toEqual(replacement);
     |                                                      ^
1515 | });
1516 | 
```


~~~~
FACT: E-031 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry distinguishes two revisions of the same key within its eight-source limit; src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry dedup retains the full restored snapshot for content-aware cleanup | baseline PASS (recorded batch baseline) | 1495 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toEqual([ / 1514 |   expect(await props.history.store.read(source.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: E-031 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-036 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-036 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-036-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP restoring a descendant handles all concurrently offered ancestors

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP restoring a descendant handles all concurrently offered ancestors'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP restoring a descendant handles all concurrently offered ancestors",
  "status": "fail",
  "duration": 23,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <button …(9)></button> to be null",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1524,
        "column": 69,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1524
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
<button
  aria-disabled="false"
  class="ui-button ui-button--primary"
  data-focus-visible="true"
  data-focused="true"
  data-rac=""
  data-react-aria-pressable="true"
  id="react-aria-_r_2lp_"
  tabindex="0"
  type="button"
>
  Відновити зміни
</button>
```

codeFrame (error 1):

```text
1522 |   await act(async () => {});
1523 |   clickRestore();
1524 |   expect(screen.queryByRole('button', { name: 'Відновити зміни' })).toBeNull();
     |                                                                     ^
1525 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1526 |   await act(async () => {});
```


~~~~
FACT: E-036 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP restoring a descendant handles all concurrently offered ancestors | baseline PASS (recorded batch baseline) | 1524 |   expect(screen.queryByRole('button', { name: 'Відновити зміни' })).toBeNull(); | CLOSED | DEFECT=no
FACT: E-036 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-048 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-048 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-048-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP loading ASR after a detached restore releases and reoffers the old source

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP loading ASR after a detached restore releases and reoffers the old source'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP loading ASR after a detached restore releases and reoffers the old source",
  "status": "fail",
  "duration": 84,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an accessible element with the role \"button\" and name \"Відновити зміни\"\n\nHere are the accessible roles:\n\n  region:\n\n  Name \"Транскрипт уроку\":\n  <section\n    aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n    class=\"transcript-editor\"\n  />\n\n  --------------------------------------------------\n  heading:\n\n  Name \"Транскрипт уроку\":\n  <h4\n    id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n  />\n\n  Name \"Чернетка розпізнавання готова\":\n  <h5\n    id=\"transcript-draft-offer-30000000-0000-4000-8000-000000000001\"\n  />\n\n  --------------------------------------------------\n  paragraph:\n\n  Name \"\":\n  <p />\n\n  --------------------------------------------------\n  button:\n\n  Name \"Завантажити чернетку\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--primary\"\n    data-focus-visible=\"true\"\n    data-focused=\"true\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2mq_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Не завантажувати\":\n  <button\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2ms_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Видалити репліку 1\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2mg_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Додати репліку\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--secondary\"\n    data-add-line=\"true\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2mi_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Зберегти транскрипт\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--primary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2mk_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  --------------------------------------------------\n  list:\n\n  Name \"\":\n  <ol\n    class=\"transcript-editor-lines\"\n  />\n\n  --------------------------------------------------\n  listitem:\n\n  Name \"\":\n  <li />\n\n  --------------------------------------------------\n  textbox:\n\n  Name \"Початок репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2m2_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2m1_\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"0\"\n  />\n\n  Name \"Кінець репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2m7_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2m6_\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"1000\"\n  />\n\n  Name \"Текст репліки 1\":\n  <textarea\n    aria-labelledby=\"react-aria-_r_2mc_\"\n    class=\"ui-input ui-textarea\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2mb_\"\n    tabindex=\"0\"\n    title=\"\"\n  />\n\n  --------------------------------------------------\n  group:\n\n  Name \"Матеріали репліки 1\":\n  <fieldset />\n\n  --------------------------------------------------\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <section\n      aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      class=\"transcript-editor\"\n    >\n      <h4\n        id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      >\n        Транскрипт уроку\n      </h4>\n      <div\n        aria-labelledby=\"transcript-draft-offer-30000000-0000-4000-8000-000000000001\"\n        class=\"transcript-editor-draft-offer\"\n      >\n        <h5\n          id=\"transcript-draft-offer-30000000-0000-4000-8000-000000000001\"\n        >\n          Чернетка розпізнавання готова\n        </h5>\n        <p>\n          Розпізнавання завершилось. Завантажити текст у редактор?\n        </p>\n        <button\n          aria-disabled=\"false\"\n          class=\"ui-button ui-button--primary\"\n          data-focus-visible=\"true\"\n          data-focused=\"true\"\n          data-rac=\"\"\n          data-react-aria-pressable=\"true\"\n          id=\"react-aria-_r_2mq_\"\n          tabindex=\"0\"\n          type=\"button\"\n        >\n          Завантажити чернетку\n        </button>\n        <button\n          class=\"ui-button ui-button--secondary\"\n          data-rac=\"\"\n          data-react-aria-pressable=\"true\"\n          id=\"react-aria-_r_2ms_\"\n          tabindex=\"0\"\n          type=\"button\"\n        >\n          Не завантажувати\n        </button>\n      </div>\n      <ol\n        class=\"transcript-editor-lines\"\n      >\n        <li>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2m1_\"\n              id=\"react-aria-_r_2m2_\"\n            >\n              Початок репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2m2_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2m1_\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"0\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2m6_\"\n              id=\"react-aria-_r_2m7_\"\n            >\n              Кінець репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2m7_\"\n      
~~~~
FACT: E-048 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP loading ASR after a detached restore releases and reoffers the old source | baseline PASS (recorded batch baseline) | 1540 |   expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'ASR replacement'); / 1541 |   expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy(); | CLOSED | DEFECT=no
FACT: E-048 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-049 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-049 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-049-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP legacy recovery keeps the loaded ASR generation and current editor language

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP legacy recovery keeps the loaded ASR generation and current editor language'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP legacy recovery keeps the loaded ASR generation and current editor language",
  "status": "fail",
  "duration": 27,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { key: 'author:gap-editor', …(7) } to match object { language: 'uk', …(2) }\n(5 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1556,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1556
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,7 +1,7 @@
  {
-   "draftUpdatedAt": "D1",
+   "draftUpdatedAt": "D2",
    "language": "uk",
    "lines": [
      {
        "documentIds": [],
        "endMs": "1000",
```

codeFrame (error 1):

```text
1554 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1555 |   await act(async () => {});
1556 |   expect(await props.history.store.read(source.key)).toMatchObject({ language: 'uk', draftUpdatedAt: 'D1', lines: [storedLine] });
     |                                                      ^
1557 |   view.unmount();
1558 |   const revisit = render(<TranscriptEditor {...props} draftUpdatedAt="D2" />);
```


~~~~
FACT: E-049 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP legacy recovery keeps the loaded ASR generation and current editor language | baseline PASS (recorded batch baseline) | 1556 |   expect(await props.history.store.read(source.key)).toMatchObject({ language: 'uk', draftUpdatedAt: 'D1', lines: [storedLine] }); | CLOSED | DEFECT=no
FACT: E-049 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-052 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-052 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-052-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields",
  "status": "fail",
  "duration": 22,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "the given combination of arguments (undefined and string) is invalid for this assertion. You can use an array, a map, an object, a set, a string, or a weakset instead of a string",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 828,
        "column": 96,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 828
    }
  ]
}
```

codeFrame (error 1):

```text
826 |   expect(text.getAttribute('aria-describedby')?.split(' ')).toContain(hint.id);
827 |   for (const name of ['Line 1 start, ms', 'Line 1 end, ms']) {
828 |     expect(screen.getByRole('textbox', { name }).getAttribute('aria-describedby')?.split(' ')).toContain(hint.id);
    |                                                                                                ^
829 |   }
830 |   text.focus();
```


~~~~
FACT: E-052 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields | baseline PASS (recorded batch baseline) | 826 |   expect(text.getAttribute('aria-describedby')?.split(' ')).toContain(hint.id); / 828 |     expect(screen.getByRole('textbox', { name }).getAttribute('aria-describedby')?.split(' ')).toContain(hint.id); | CLOSED | DEFECT=no
FACT: E-052 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-059 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-059 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-059-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a rejected Save reports unavailable, retains recovery and resumes later captures

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a rejected Save reports unavailable, retains recovery and resumes later captures'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a rejected Save reports unavailable, retains recovery and resumes later captures",
  "status": "fail",
  "duration": 103,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an element with the text: Сервіс недоступний. Спробуйте ще раз.. This could be because the text is broken up by multiple elements. In this case, you can provide a function for your text matcher to make your matcher more flexible.\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <section\n      aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      class=\"transcript-editor\"\n    >\n      <h4\n        id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      >\n        Транскрипт уроку\n      </h4>\n      <ol\n        class=\"transcript-editor-lines\"\n      >\n        <li>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2ok_\"\n              id=\"react-aria-_r_2ol_\"\n            >\n              Початок репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2ol_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2ok_\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"0\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2op_\"\n              id=\"react-aria-_r_2oq_\"\n            >\n              Кінець репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2oq_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2op_\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"1000\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2ou_\"\n              id=\"react-aria-_r_2ov_\"\n            >\n              Текст репліки 1\n            </label>\n            <textarea\n              aria-labelledby=\"react-aria-_r_2ov_\"\n              class=\"ui-input ui-textarea\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2ou_\"\n              tabindex=\"0\"\n              title=\"\"\n            >\n              Before failure\n            </textarea>\n          </div>\n          <fieldset>\n            <legend>\n              Матеріали репліки 1\n            </legend>\n          </fieldset>\n          <button\n            aria-disabled=\"false\"\n            class=\"ui-button ui-button--secondary\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            id=\"react-aria-_r_2p3_\"\n            tabindex=\"0\"\n            type=\"button\"\n          >\n            Видалити репліку 1\n          </button>\n        </li>\n      </ol>\n      <button\n        aria-disabled=\"false\"\n        class=\"ui-button ui-button--secondary\"\n        data-add-line=\"true\"\n        data-rac=\"\"\n        data-react-aria-pressable=\"true\"\n        id=\"react-aria-_r_2p5_\"\n        tabindex=\"0\"\n        type=\"button\"\n      >\n        Додати репліку\n      </button>\n      <button\n        aria-disabled=\"false\"\n        class=\"ui-button ui-button--primary\"\n        data-rac=\"\"\n        data-react-aria-pressable=\"true\"\n        id=\"react-aria-_r_2p7_\"\n        tabindex=\"0\"\n        type=\"button\"\n      >\n        Зберегти транскрипт\n      </button>\n      <output>\n        Транскрипт збережено.\n      </output>\n    </section>\n  </div>\n</body>",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1576,
        "column": 10,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1576
    }
  ]
}
```

codeFrame (error 1):

```text
1574 |   clickSave();
1575 |   await act(async () => {});
1576 |   expect(screen.getByText('Сервіс недоступний. Спробуйте ще раз.')).toBeTruthy();
     |          ^
1577 |   expect(screen.queryByText('Транскрипт збережено.')).toBeNull();
1578 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ lines: [{ text: 'Before failure' }] });
```


~~~~
FACT: E-059 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a rejected Save reports unavailable, retains recovery and resumes later captures | baseline PASS (recorded batch baseline) | 1576 |   expect(screen.getByText('Сервіс недоступний. Спробуйте ще раз.')).toBeTruthy(); / 1577 |   expect(screen.queryByText('Транскрипт збережено.')).toBeNull(); / 1578 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ lines: [{ text: 'Before failure' }] }); | CLOSED | DEFECT=no
FACT: E-059 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-060 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-060 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-060-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a rejected Save reports unavailable, retains recovery and resumes later captures

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a rejected Save reports unavailable, retains recovery and resumes later captures'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a rejected Save reports unavailable, retains recovery and resumes later captures",
  "status": "fail",
  "duration": 36,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { key: 'author:gap-editor', …(6) } to match object { Object (lines) }\n(10 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1582,
        "column": 61,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1582
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
    "lines": [
      {
-       "text": "After failure",
+       "text": "Before failure",
      },
    ],
  }
```

codeFrame (error 1):

```text
1580 |   fireEvent.focusOut(text);
1581 |   await act(async () => {});
1582 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ lines: [{ text: 'After failure' }] });
     |                                                             ^
1583 | });
1584 | 
```

### [F02] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP successful detached Save removes ancestry from subsequent author captures

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP successful detached Save removes ancestry from subsequent author captures'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP successful detached Save removes ancestry from subsequent author captures",
  "status": "fail",
  "duration": 14,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to match object { Object (lines) }",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1639,
        "column": 61,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1639
    }
  ]
}
```

diff:

```diff
- Expected:
{
  "lines": [
    {
      "text": "New edit after successful Save",
    },
  ],
}

+ Received:
null
```

codeFrame (error 1):

```text
1637 |   fireEvent.focusOut(text);
1638 |   await act(async () => {});
1639 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ lines: [{ text: 'New edit after successful Save' }] });
     |                                                             ^
1640 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined();
1641 | });
```


~~~~
FACT: E-060 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a rejected Save reports unavailable, retains recovery and resumes later captures; src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP successful detached Save removes ancestry from subsequent author captures | baseline PASS (recorded batch baseline) | 1582 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ lines: [{ text: 'After failure' }] }); / 1639 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ lines: [{ text: 'New edit after successful Save' }] }); / 1640 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined(); | CLOSED | DEFECT=no
FACT: E-060 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-062 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-062 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-062-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP source cleanup rechecks newly mounted ownership after its deferred read

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP source cleanup rechecks newly mounted ownership after its deferred read'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP source cleanup rechecks newly mounted ownership after its deferred read",
  "status": "fail",
  "duration": 28,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ [ 'newly-mounted-source', …(1) ] ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1622,
        "column": 66,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1622
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   [
+     "newly-mounted-source",
+     "revision:newly-mounted-source",
+   ],
+ ]
```

codeFrame (error 1):

```text
1620 |   view.rerender(<TranscriptEditor {...props} history={{ key: props.history.key, store }} detachedSnapshots={offers} mountedHistoryKeys={[source.key]} />);
1621 |   await act(async () => pending.resolve(source));
1622 |   expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]);
     |                                                                  ^
1623 |   expect(await memory.read(source.key)).toEqual(source);
1624 | });
```


~~~~
FACT: E-062 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP source cleanup rechecks newly mounted ownership after its deferred read | baseline PASS (recorded batch baseline) | 1622 |   expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]); / 1623 |   expect(await memory.read(source.key)).toEqual(source); | CLOSED | DEFECT=no
FACT: E-062 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-063 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-063 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-063-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP successful detached Save removes ancestry from subsequent author captures

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP successful detached Save removes ancestry from subsequent author captures'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP successful detached Save removes ancestry from subsequent author captures",
  "status": "fail",
  "duration": 33,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { key: 'saved-source', …(1) } ] to be undefined",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1640,
        "column": 79,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1640
    }
  ]
}
```

diff:

```diff
- Expected:
undefined

+ Received:
[
  {
    "key": "saved-source",
    "revision": "revision:saved-source",
  },
]
```

codeFrame (error 1):

```text
1638 |   await act(async () => {});
1639 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ lines: [{ text: 'New edit after successful Save' }] });
1640 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined();
     |                                                                               ^
1641 | });
1642 | 
```


~~~~
FACT: E-063 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP successful detached Save removes ancestry from subsequent author captures | baseline PASS (recorded batch baseline) | 1639 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ lines: [{ text: 'New edit after successful Save' }] }); / 1640 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined(); | CLOSED | DEFECT=no
FACT: E-063 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-066 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-066 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-066-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields",
  "status": "fail",
  "duration": 21,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "the given combination of arguments (undefined and string) is invalid for this assertion. You can use an array, a map, an object, a set, a string, or a weakset instead of a string",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 828,
        "column": 96,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 828
    }
  ]
}
```

codeFrame (error 1):

```text
826 |   expect(text.getAttribute('aria-describedby')?.split(' ')).toContain(hint.id);
827 |   for (const name of ['Line 1 start, ms', 'Line 1 end, ms']) {
828 |     expect(screen.getByRole('textbox', { name }).getAttribute('aria-describedby')?.split(' ')).toContain(hint.id);
    |                                                                                                ^
829 |   }
830 |   text.focus();
```


~~~~
FACT: E-066 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields | baseline PASS (recorded batch baseline) | 826 |   expect(text.getAttribute('aria-describedby')?.split(' ')).toContain(hint.id); / 828 |     expect(screen.getByRole('textbox', { name }).getAttribute('aria-describedby')?.split(' ')).toContain(hint.id); | CLOSED | DEFECT=no
FACT: E-066 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-072 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-072 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-072-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP detached Restore during pending Save cannot change text or ownership

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP detached Restore during pending Save cannot change text or ownership'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP detached Restore during pending Save cannot change text or ownership",
  "status": "fail",
  "duration": 30,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <textarea tabindex=\"0\" …(7)></textarea> to have property \"value\" with value 'Saving own'",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1474,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1474
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- Saving own
+ Detached
```

codeFrame (error 1):

```text
1472 |     fireEvent.click(screen.getByRole('button', { name: action === 'Restore' ? 'Відновити зміни' : 'Не відновлювати' }));
1473 |     await act(async () => {});
1474 |     expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'Saving own');
     |                                                      ^
1475 |     expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy();
1476 |     expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]);
```


~~~~
FACT: E-072 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP detached Restore during pending Save cannot change text or ownership | baseline PASS (recorded batch baseline) | 1474 |     expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'Saving own'); / 1475 |     expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy(); / 1476 |     expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]); | CLOSED | DEFECT=no
FACT: E-072 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-073 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-073 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-073-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP detached Discard during pending Save cannot change text or ownership

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP detached Discard during pending Save cannot change text or ownership'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP detached Discard during pending Save cannot change text or ownership",
  "status": "fail",
  "duration": 122,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an accessible element with the role \"button\" and name \"Відновити зміни\"\n\nHere are the accessible roles:\n\n  region:\n\n  Name \"Транскрипт уроку\":\n  <section\n    aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n    class=\"transcript-editor\"\n  />\n\n  --------------------------------------------------\n  heading:\n\n  Name \"Транскрипт уроку\":\n  <h4\n    id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n  />\n\n  --------------------------------------------------\n  list:\n\n  Name \"\":\n  <ol\n    class=\"transcript-editor-lines\"\n  />\n\n  --------------------------------------------------\n  listitem:\n\n  Name \"\":\n  <li />\n\n  --------------------------------------------------\n  textbox:\n\n  Name \"Початок репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2iq_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2ip_\"\n    readonly=\"\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"0\"\n  />\n\n  Name \"Кінець репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2iv_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2iu_\"\n    readonly=\"\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"1000\"\n  />\n\n  Name \"Текст репліки 1\":\n  <textarea\n    aria-labelledby=\"react-aria-_r_2j4_\"\n    class=\"ui-input ui-textarea\"\n    data-focus-visible=\"true\"\n    data-focused=\"true\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2j3_\"\n    readonly=\"\"\n    tabindex=\"0\"\n    title=\"\"\n  />\n\n  --------------------------------------------------\n  group:\n\n  Name \"Матеріали репліки 1\":\n  <fieldset />\n\n  --------------------------------------------------\n  button:\n\n  Name \"Видалити репліку 1\":\n  <button\n    aria-disabled=\"true\"\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2j8_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Додати репліку\":\n  <button\n    aria-disabled=\"true\"\n    class=\"ui-button ui-button--secondary\"\n    data-add-line=\"true\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2ja_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Зберегти транскрипт\":\n  <button\n    aria-disabled=\"true\"\n    class=\"ui-button ui-button--primary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2jc_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  --------------------------------------------------\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <section\n      aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      class=\"transcript-editor\"\n    >\n      <h4\n        id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      >\n        Транскрипт уроку\n      </h4>\n      <ol\n        class=\"transcript-editor-lines\"\n      >\n        <li>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n            data-readonly=\"true\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2ip_\"\n              id=\"react-aria-_r_2iq_\"\n            >\n              Початок репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2iq_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2ip_\"\n              readonly=\"\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"0\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n            data-readonly=\"true\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2iu_\"\n              id=\"react-aria-_r_2iv_\"\n            >\n              Кінець репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2iv_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2iu_\"\n              readonly=\"\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"1000\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n            data-readonly=\"true\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2j3_\"\n              id=\"react-aria-_r_2j4_\"\n            >\n              Текст репліки 1\n            </label>\n            <textarea\n              aria-labelledby=\"react-aria-_r_2j4_\"\n              class=\"ui-input ui-textarea\"\n              data-focus-visible=\"true\"\n              data-focused=\"true\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2j3_\"\n              readonly=\"\"\n              tabindex=\"0\"\n              title=\"\"\n            >\n              Saving own\n            </textarea>\n          </div>\n          <fieldset>\n            <legend>\n              Матеріали репліки 1\n            </legend>\n          </fieldset>\n          <button\n            aria-disabled=\"true\"\n            class=\"ui-button ui-button--secondary\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            id=\"react-aria-_r_2j8_\"\n            tabindex=\"0\"\n            type=\"button\"\n          >\n            Видалити репліку 1\n          </button>\n        </li>\n      </ol>\n      <button\n        aria-disabled=\"true\"\n        class=\"ui-button ui-button--secondary\"\n        data-add-line=\"tru
~~~~
FACT: E-073 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP detached Discard during pending Save cannot change text or ownership | baseline PASS (recorded batch baseline) | 1474 |     expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'Saving own'); / 1475 |     expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy(); / 1476 |     expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]); / 1477 |     expect(await props.history.store.read(source.key)).toEqual(source); | CLOSED | DEFECT=no
FACT: E-073 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: E-074 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-074 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-074-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP Discard of a refreshed own offer releases an earlier detached restoration

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP Discard of a refreshed own offer releases an earlier detached restoration'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP Discard of a refreshed own offer releases an earlier detached restoration",
  "status": "fail",
  "duration": 71,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an accessible element with the role \"button\" and name \"Відновити зміни\"\n\nHere are the accessible roles:\n\n  region:\n\n  Name \"Транскрипт уроку\":\n  <section\n    aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n    class=\"transcript-editor\"\n  />\n\n  --------------------------------------------------\n  heading:\n\n  Name \"Транскрипт уроку\":\n  <h4\n    id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n  />\n\n  --------------------------------------------------\n  list:\n\n  Name \"\":\n  <ol\n    class=\"transcript-editor-lines\"\n  />\n\n  --------------------------------------------------\n  listitem:\n\n  Name \"\":\n  <li />\n\n  --------------------------------------------------\n  textbox:\n\n  Name \"Початок репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2rl_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2rk_\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"0\"\n  />\n\n  Name \"Кінець репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2rq_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2rp_\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"1000\"\n  />\n\n  Name \"Текст репліки 1\":\n  <textarea\n    aria-labelledby=\"react-aria-_r_2rv_\"\n    class=\"ui-input ui-textarea\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2ru_\"\n    tabindex=\"0\"\n    title=\"\"\n  />\n\n  --------------------------------------------------\n  group:\n\n  Name \"Матеріали репліки 1\":\n  <fieldset />\n\n  --------------------------------------------------\n  button:\n\n  Name \"Видалити репліку 1\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2s3_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Додати репліку\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--secondary\"\n    data-add-line=\"true\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2s5_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Зберегти транскрипт\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--primary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2s7_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  --------------------------------------------------\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <section\n      aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      class=\"transcript-editor\"\n    >\n      <h4\n        id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      >\n        Транскрипт уроку\n      </h4>\n      <ol\n        class=\"transcript-editor-lines\"\n      >\n        <li>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2rk_\"\n              id=\"react-aria-_r_2rl_\"\n            >\n              Початок репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2rl_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2rk_\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"0\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2rp_\"\n              id=\"react-aria-_r_2rq_\"\n            >\n              Кінець репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2rq_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2rp_\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"1000\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2ru_\"\n              id=\"react-aria-_r_2rv_\"\n            >\n              Текст репліки 1\n            </label>\n            <textarea\n              aria-labelledby=\"react-aria-_r_2rv_\"\n              class=\"ui-input ui-textarea\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2ru_\"\n              tabindex=\"0\"\n              title=\"\"\n            >\n              Перша\n            </textarea>\n          </div>\n          <fieldset>\n            <legend>\n              Матеріали репліки 1\n            </legend>\n          </fieldset>\n          <button\n            aria-disabled=\"false\"\n            class=\"ui-button ui-button--secondary\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            id=\"react-aria-_r_2s3_\"\n            tabindex=\"0\"\n            type=\"button\"\n          >\n            Видалити репліку 1\n          </button>\n        </li>\n      </ol>\n      <button\n        aria-disabled=\"false\"\n        class=\"ui-button ui-button--secondary\"\n        data-add-line=\"true\"\n        data-rac=\"\"\n        data-react-aria-pressable=\"true\"\n        id=\"react-aria-_r_2s5_\"\n        tabindex=\"0\"\n        type=\"button\"\n      >\n        Додати репліку\n      </button>\n      <button\n        aria-disabled=\"false\"\n        class=\"ui-button ui-button--primary\"\n        data-rac=\"\"\n        data-react-aria-pressable=\"true\"\n        i
~~~~
FACT: E-074 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP Discard of a refreshed own offer releases an earlier detached restoration | baseline PASS (recorded batch baseline) | 1661 |   expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy(); | CLOSED | DEFECT=no
FACT: E-074 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
NEXT: X-017 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: X-017 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-017-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP legacy recovery keeps the loaded ASR generation and current editor language

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP legacy recovery keeps the loaded ASR generation and current editor language'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP legacy recovery keeps the loaded ASR generation and current editor language",
  "status": "fail",
  "duration": 26,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { key: 'author:gap-editor', …(6) } to match object { language: 'uk', …(2) }\n(5 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1556,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1556
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,8 +1,7 @@
  {
    "draftUpdatedAt": "D1",
-   "language": "uk",
    "lines": [
      {
        "documentIds": [],
        "endMs": "1000",
        "id": "70000000-0000-4000-8000-000000000001",
```

codeFrame (error 1):

```text
1554 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1555 |   await act(async () => {});
1556 |   expect(await props.history.store.read(source.key)).toMatchObject({ language: 'uk', draftUpdatedAt: 'D1', lines: [storedLine] });
     |                                                      ^
1557 |   view.unmount();
1558 |   const revisit = render(<TranscriptEditor {...props} draftUpdatedAt="D2" />);
```


~~~~
FACT: X-017 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP legacy recovery keeps the loaded ASR generation and current editor language | baseline PASS (recorded batch baseline) | 1556 |   expect(await props.history.store.read(source.key)).toMatchObject({ language: 'uk', draftUpdatedAt: 'D1', lines: [storedLine] }); | CLOSED | DEFECT=no
FACT: X-017 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
DECISION: E-013 first mutation survived because changed segments are not a useMemo dependency while dirty base is retained; the test did not reach the mutated baseline expression. Added supplied draftUpdatedAt=D2 during the server refresh to trigger the real memoization path without changing retained T1. This is a test-only reachability repair, not a production defect. First raw survivor retained as gap-E-013-mutation-first-pass.txt. Other 73 GAPs have named mutation failures. NEXT: Rebaseline final editor test, rerun E-013.
NEXT: E-013 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-013 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-013-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP dirty edits compare against their retained base after a server prop refresh

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP dirty edits compare against their retained base after a server prop refresh'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP dirty edits compare against their retained base after a server prop refresh",
  "status": "fail",
  "duration": 34,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to match object { expectedUpdatedAt: 'T1', …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1434,
        "column": 61,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1434
    }
  ]
}
```

diff:

```diff
- Expected:
{
  "expectedUpdatedAt": "T1",
  "lines": [
    {
      "text": "Server moved",
    },
  ],
}

+ Received:
null
```

codeFrame (error 1):

```text
1432 |   fireEvent.focusOut(text);
1433 |   await act(async () => {});
1434 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ expectedUpdatedAt: 'T1', lines: [{ text: 'Server moved' }] });
     |                                                             ^
1435 | });
1436 | 
```


~~~~
FACT: E-013 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP dirty edits compare against their retained base after a server prop refresh | baseline PASS (recorded batch baseline) | 1434 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ expectedUpdatedAt: 'T1', lines: [{ text: 'Server moved' }] }); | CLOSED | DEFECT=no
FACT: E-013 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: Final editor baseline (with E-013 reachability repair) exit 0, 86/86 tests; raw fe-a-editor-baseline-final.txt. Latest per-ID outcome count {"CLOSED": 74, "REMAINING": 0}. NEXT: Run owner-requested frontend end checks in order, then audit diff scope and SHA-256 values.
FACT: End check `cd frontend && npm run format` exit=0; raw fe-a-format.txt.
NEXT: End check `cd frontend && npm run format:check`.
FACT: End check `cd frontend && npm run format:check` exit=0; raw fe-a-format-check.txt.
NEXT: End check `cd frontend && npm run lint`.
FACT: End check `cd frontend && npm run lint` exit=1; raw fe-a-lint.txt.
FACT: End lint exit 1: five newly added non-null assertions in editor fixtures violate typescript(no-non-null-assertion). Replaced them with a single checked firstSegment fixture reference; assertions/behavior unchanged. Typecheck and whole suite had not run (stopped on lint failure). NEXT: Reformat, repeat ordered end checks; recheck affected editor mutations on final fixture bytes.
FACT: End check `cd frontend && npm run format` exit=0; raw fe-a-format.txt.
NEXT: End check `cd frontend && npm run format:check`.
FACT: End check `cd frontend && npm run format:check` exit=0; raw fe-a-format-check.txt.
NEXT: End check `cd frontend && npm run lint`.
FACT: End check `cd frontend && npm run lint` exit=0; raw fe-a-lint.txt.
NEXT: End check `cd frontend && npm run typecheck`.
FACT: End check `cd frontend && npm run typecheck` exit=2; raw fe-a-typecheck.txt.
FACT: Typecheck exit 2: repository targets ES2022, so Promise.withResolvers (runtime-supported under Node 24) lacks its TS lib declaration; also one afterEach returned RstestUtilities instead of void. Replaced with local deferred helper using the existing LessonDocumentEditor test idiom; fixed afterEach return. No compiler/production/config changes. First typecheck raw retained as fe-a-typecheck-first-failure.txt. NEXT: Reformat and rerun ordered end checks; rerun exact mutations affected by fixture scheduling construction on final bytes.
FACT: End check `cd frontend && npm run format` exit=0; raw fe-a-format.txt.
NEXT: End check `cd frontend && npm run format:check`.
FACT: End check `cd frontend && npm run format:check` exit=0; raw fe-a-format-check.txt.
NEXT: End check `cd frontend && npm run lint`.
FACT: End check `cd frontend && npm run lint` exit=0; raw fe-a-lint.txt.
NEXT: End check `cd frontend && npm run typecheck`.
FACT: End check `cd frontend && npm run typecheck` exit=0; raw fe-a-typecheck.txt.
NEXT: End check `cd frontend && npx rstest run`.
FACT: End check `cd frontend && npx rstest run` exit=0; counts={
    "testFiles": 108,
    "failedFiles": 0,
    "tests": 1257,
    "failedTests": 0,
    "passedTests": 1257,
    "skippedTests": 0,
    "todoTests": 0
  }; raw fe-a-suite.txt.
FACT: Final ordered frontend checks PASS: format 0, format:check 0, lint 0, typecheck 0, full npx rstest run 0 with 1257/1257 tests in 108 files, 0 skipped. Baseline delta +66 tests from original 1191. NEXT: One final exact 74-row mutation confirmation on these formatted, typechecked test bytes, with per-test SHA and separate -final raw logs. No test edits during this comparison.
FACT: A-007 test SHA-256=ab2dd6b76f69ca87c4b491ad8b4294502b188eda4fe9b221c797165d3aeb9023; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: A-007 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: A-007 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-A-007-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for obsolete lessonId

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for obsolete lessonId'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for obsolete lessonId",
  "status": "fail",
  "duration": 4,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: [ { …(12) } ] } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 96,
        "column": 56,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "error": {
-     "code": "invalid_response",
+   "ok": true,
+   "value": [
+     {
+       "assetId": "50000000-0000-4000-8000-000000000001",
+       "failure": null,
+       "generation": 1,
+       "language": "uk",
+       "lessonId": "20000000-0000-4000-8000-000000000001",
+       "segments": [
+         {
+           "confidence": 0.9,
+           "endMs": 1000,
+           "id": "70000000-0000-4000-8000-000000000001",
+           "speaker": "Lecturer",
+           "startMs": 0,
+           "text": "Вітаю",
          },
-   "ok": false,
+       ],
+       "stale": false,
+       "state": "ready",
+       "trackId": "60000000-0000-4000-8000-000000000001",
+       "trackLanguage": "uk",
+       "trackMultilingual": true,
+       "updatedAt": "2026-11-15T12:00:00.123456Z",
+     },
+   ],
  }
```

codeFrame (error 1):

```text
94 |   ])('rejects the whole list for $name', async ({ response }) => {
95 |     server.use(http.get(api, () => HttpResponse.json(response)));
96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid);
   |                                                        ^
97 |   });
98 | 
```


~~~~
FACT: A-007 | src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for obsolete lessonId | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid); | CLOSED | DEFECT=no
FACT: A-007 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
FACT: A-011 test SHA-256=ab2dd6b76f69ca87c4b491ad8b4294502b188eda4fe9b221c797165d3aeb9023; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: A-011 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: A-011 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-A-011-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for fractional generation

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for fractional generation'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for fractional generation",
  "status": "fail",
  "duration": 4,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: [ { …(12) } ] } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 96,
        "column": 56,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "error": {
-     "code": "invalid_response",
+   "ok": true,
+   "value": [
+     {
+       "assetId": "50000000-0000-4000-8000-000000000001",
+       "failure": null,
+       "generation": 1.5,
+       "language": "uk",
+       "lessonId": "20000000-0000-4000-8000-000000000001",
+       "segments": [
+         {
+           "confidence": 0.9,
+           "endMs": 1000,
+           "id": "70000000-0000-4000-8000-000000000001",
+           "speaker": "Lecturer",
+           "startMs": 0,
+           "text": "Вітаю",
          },
-   "ok": false,
+       ],
+       "stale": false,
+       "state": "ready",
+       "trackId": "60000000-0000-4000-8000-000000000001",
+       "trackLanguage": "uk",
+       "trackMultilingual": true,
+       "updatedAt": "2026-11-15T12:00:00.123456Z",
+     },
+   ],
  }
```

codeFrame (error 1):

```text
94 |   ])('rejects the whole list for $name', async ({ response }) => {
95 |     server.use(http.get(api, () => HttpResponse.json(response)));
96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid);
   |                                                        ^
97 |   });
98 | 
```


~~~~
FACT: A-011 | src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for fractional generation | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid); | CLOSED | DEFECT=no
FACT: A-011 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
FACT: A-012 test SHA-256=ab2dd6b76f69ca87c4b491ad8b4294502b188eda4fe9b221c797165d3aeb9023; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: A-012 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: A-012 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-A-012-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for zero generation

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for zero generation'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for zero generation",
  "status": "fail",
  "duration": 3,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: [ { …(12) } ] } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 96,
        "column": 56,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "error": {
-     "code": "invalid_response",
+   "ok": true,
+   "value": [
+     {
+       "assetId": "50000000-0000-4000-8000-000000000001",
+       "failure": null,
+       "generation": 0,
+       "language": "uk",
+       "lessonId": "20000000-0000-4000-8000-000000000001",
+       "segments": [
+         {
+           "confidence": 0.9,
+           "endMs": 1000,
+           "id": "70000000-0000-4000-8000-000000000001",
+           "speaker": "Lecturer",
+           "startMs": 0,
+           "text": "Вітаю",
          },
-   "ok": false,
+       ],
+       "stale": false,
+       "state": "ready",
+       "trackId": "60000000-0000-4000-8000-000000000001",
+       "trackLanguage": "uk",
+       "trackMultilingual": true,
+       "updatedAt": "2026-11-15T12:00:00.123456Z",
+     },
+   ],
  }
```

codeFrame (error 1):

```text
94 |   ])('rejects the whole list for $name', async ({ response }) => {
95 |     server.use(http.get(api, () => HttpResponse.json(response)));
96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid);
   |                                                        ^
97 |   });
98 | 
```

### [F02] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for negative generation

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for negative generation'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for negative generation",
  "status": "fail",
  "duration": 2,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: [ { …(12) } ] } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 96,
        "column": 56,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "error": {
-     "code": "invalid_response",
+   "ok": true,
+   "value": [
+     {
+       "assetId": "50000000-0000-4000-8000-000000000001",
+       "failure": null,
+       "generation": -1,
+       "language": "uk",
+       "lessonId": "20000000-0000-4000-8000-000000000001",
+       "segments": [
+         {
+           "confidence": 0.9,
+           "endMs": 1000,
+           "id": "70000000-0000-4000-8000-000000000001",
+           "speaker": "Lecturer",
+           "startMs": 0,
+           "text": "Вітаю",
          },
-   "ok": false,
+       ],
+       "stale": false,
+       "state": "ready",
+       "trackId": "60000000-0000-4000-8000-000000000001",
+       "trackLanguage": "uk",
+       "trackMultilingual": true,
+       "updatedAt": "2026-11-15T12:00:00.123456Z",
+     },
+   ],
  }
```

codeFrame (error 1):

```text
94 |   ])('rejects the whole list for $name', async ({ response }) => {
95 |     server.use(http.get(api, () => HttpResponse.json(response)));
96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid);
   |                                                        ^
97 |   });
98 | 
```


~~~~
FACT: A-012 | src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for zero generation; src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for negative generation | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid); / 96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid); | CLOSED | DEFECT=no
FACT: A-012 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
FACT: A-028 test SHA-256=ab2dd6b76f69ca87c4b491ad8b4294502b188eda4fe9b221c797165d3aeb9023; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: A-028 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: A-028 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-A-028-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: GAP inventory cancellation reports request_cancelled after request starts

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'GAP inventory cancellation reports request_cancelled after request starts'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "GAP inventory cancellation reports request_cancelled after request starts",
  "status": "fail",
  "duration": 3,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: [] } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 216,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 216
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "error": {
-     "code": "request_cancelled",
-   },
-   "ok": false,
+   "ok": true,
+   "value": [],
  }
```

codeFrame (error 1):

```text
214 |   controller.abort();
215 |   releaseResponse();
216 |   expect(await pending).toEqual({ ok: false, error: { code: 'request_cancelled' } });
    |                         ^
217 | });
218 | 
```


~~~~
FACT: A-028 | src/modules/transcript/adapters/draft-repository.test.ts :: GAP inventory cancellation reports request_cancelled after request starts | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 216 |   expect(await pending).toEqual({ ok: false, error: { code: 'request_cancelled' } }); | CLOSED | DEFECT=no
FACT: A-028 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
FACT: I-003 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-003 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-003 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-003-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects missing key without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects missing key without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects missing key without offering corrupt recovery",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { key: undefined, capturedAt: 1, …(4) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 227,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 227
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": undefined,
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "1d2a30ea-75cc-4f1d-995e-ebc13c81686f",
}
```

codeFrame (error 1):

```text
225 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
226 |   storedValue(value);
227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
228 | });
229 | 
```


~~~~
FACT: I-003 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects missing key without offering corrupt recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-003 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-004 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-004 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-004 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-004-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects missing capturedAt without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects missing capturedAt without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects missing capturedAt without offering corrupt recovery",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 227,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 227
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": undefined,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "a1207828-f9fd-4d5f-b211-298f6ef63e27",
}
```

codeFrame (error 1):

```text
225 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
226 |   storedValue(value);
227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
228 | });
229 | 
```


~~~~
FACT: I-004 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects missing capturedAt without offering corrupt recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-004 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-007 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-007 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-007 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-007-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric language without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric language without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric language without offering corrupt recovery",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 227,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 227
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": 42,
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "8649f847-648d-4b57-9c1b-a0429963a69a",
}
```

codeFrame (error 1):

```text
225 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
226 |   storedValue(value);
227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
228 | });
229 | 
```


~~~~
FACT: I-007 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric language without offering corrupt recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-007 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-009 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-009 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-009 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-009-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric expectedUpdatedAt without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric expectedUpdatedAt without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric expectedUpdatedAt without offering corrupt recovery",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 227,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 227
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": 42,
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "ea48f490-372d-47b4-b010-70ed0031b42e",
}
```

codeFrame (error 1):

```text
225 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
226 |   storedValue(value);
227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
228 | });
229 | 
```


~~~~
FACT: I-009 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric expectedUpdatedAt without offering corrupt recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-009 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-010 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-010 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-010 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-010-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric draftUpdatedAt without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric draftUpdatedAt without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric draftUpdatedAt without offering corrupt recovery",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 227,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 227
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "draftUpdatedAt": 42,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "5ae7ec02-488f-40c7-9ce6-132d57ec1441",
}
```

codeFrame (error 1):

```text
225 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
226 |   storedValue(value);
227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
228 | });
229 | 
```


~~~~
FACT: I-010 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric draftUpdatedAt without offering corrupt recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-010 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-011 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-011 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-011 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-011-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric line id without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric line id without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric line id without offering corrupt recovery",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 227,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 227
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": 42,
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "817f93d8-a19b-44d5-98f5-e1faa82c63a9",
}
```

codeFrame (error 1):

```text
225 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
226 |   storedValue(value);
227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
228 | });
229 | 
```


~~~~
FACT: I-011 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric line id without offering corrupt recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-011 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-012 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-012 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-012 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-012-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric startMs without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric startMs without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric startMs without offering corrupt recovery",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 227,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 227
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": 42,
      "text": "Unsaved",
    },
  ],
  "revision": "bd876068-8377-4c1e-8899-a4be72a4bbcf",
}
```

codeFrame (error 1):

```text
225 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
226 |   storedValue(value);
227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
228 | });
229 | 
```


~~~~
FACT: I-012 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric startMs without offering corrupt recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-012 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-013 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-013 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-013 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-013-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric endMs without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric endMs without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric endMs without offering corrupt recovery",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 227,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 227
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": 42,
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "b8827edd-7617-414e-8b5b-f9d85b3e7d2e",
}
```

codeFrame (error 1):

```text
225 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
226 |   storedValue(value);
227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
228 | });
229 | 
```


~~~~
FACT: I-013 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric endMs without offering corrupt recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-013 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-014 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-014 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-014 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-014-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric text without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric text without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric text without offering corrupt recovery",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 227,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 227
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": 42,
    },
  ],
  "revision": "04463587-1d1e-4ce2-b7d7-3fc14464fadf",
}
```

codeFrame (error 1):

```text
225 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
226 |   storedValue(value);
227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
228 | });
229 | 
```


~~~~
FACT: I-014 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric text without offering corrupt recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-014 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-015 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-015 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-015 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-015-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric documentIds without offering corrupt recovery

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB rejects numeric documentIds without offering corrupt recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB rejects numeric documentIds without offering corrupt recovery",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to be null",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 227,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 227
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
{
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "author:lesson:track:one",
  "language": "es",
  "lines": [
    {
      "documentIds": [
        42,
      ],
      "endMs": "1000",
      "id": "line",
      "startMs": "0",
      "text": "Unsaved",
    },
  ],
  "revision": "0f012abc-8f5b-448c-831f-bf60772ee15e",
}
```

codeFrame (error 1):

```text
225 | ])('GAP IndexedDB rejects $name without offering corrupt recovery', async ({ value }) => {
226 |   storedValue(value);
227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
228 | });
229 | 
```


~~~~
FACT: I-015 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB rejects numeric documentIds without offering corrupt recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 227 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: I-015 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-024 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-024 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-024 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-024-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB preserves the ASR draft token alongside valid recovery text

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'GAP IndexedDB preserves the ASR draft token alongside valid recovery text'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "GAP IndexedDB preserves the ASR draft token alongside valid recovery text",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to deeply equal { Object (key, revision, ...) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 233,
        "column": 72,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 233
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,8 +1,8 @@
  {
    "capturedAt": 1,
-   "draftUpdatedAt": "D1",
+   "draftUpdatedAt": undefined,
    "expectedUpdatedAt": "T1",
    "key": "author:lesson:track:one",
    "language": "es",
    "lines": [
      {
```

codeFrame (error 1):

```text
231 |   const value = { ...snapshot, draftUpdatedAt: 'D1' };
232 |   storedValue(value);
233 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toEqual(value);
    |                                                                        ^
234 | });
235 | 
```


~~~~
FACT: I-024 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: GAP IndexedDB preserves the ASR draft token alongside valid recovery text | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 233 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toEqual(value); | CLOSED | DEFECT=no
FACT: I-024 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: I-025 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: I-025 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: I-025 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-I-025-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: MS3b fix4 atomic IndexedDB cleanup wrong-key uses one readwrite transaction and waits for commit

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'MS3b fix4 atomic IndexedDB cleanup wrong-key uses one readwrite transaction and waits for commit'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "MS3b fix4 atomic IndexedDB cleanup wrong-key uses one readwrite transaction and waits for commit",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"author:lesson:track:one\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 115,
        "column": 29,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 115
    }
  ]
}
```

codeFrame (error 1):

```text
113 |     if (outcome === 'same' || outcome === 'abort' || outcome === 'error')
114 |       expect(remove).toHaveBeenCalledExactlyOnceWith(snapshot.key);
115 |     else expect(remove).not.toHaveBeenCalled();
    |                             ^
116 |     expect(start).toHaveBeenCalledTimes(1);
117 |     await Promise.resolve();
```


~~~~
FACT: I-025 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: MS3b fix4 atomic IndexedDB cleanup wrong-key uses one readwrite transaction and waits for commit | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 114 |       expect(remove).toHaveBeenCalledExactlyOnceWith(snapshot.key); / 115 |     else expect(remove).not.toHaveBeenCalled(); / 116 |     expect(start).toHaveBeenCalledTimes(1); | CLOSED | DEFECT=no
FACT: I-025 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: R-018 test SHA-256=f06d85b1177079158747e52d9bbb734748189e888fe4cb66ad8ab80dc6a6d3b3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: R-018 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/application/recovery.test.ts`. Production SHA-256 5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69.
FACT: R-018 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-R-018-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"key":"another-key"}

repro:

```bash
npx rstest 'src/modules/transcript/application/recovery.test.ts' --testNamePattern 'GAP accepted cleanup preserves a same-revision snapshot differing in {"key":"another-key"}'
```

details:

```json
{
  "testPath": "src/modules/transcript/application/recovery.test.ts",
  "project": "dom",
  "fullName": "GAP accepted cleanup preserves a same-revision snapshot differing in {\"key\":\"another-key\"}",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"user:lesson\",\n      \"snapshot-revision\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/application/recovery.test.ts",
        "line": 462,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/application/recovery.test.ts",
      "line": 462
    }
  ]
}
```

codeFrame (error 1):

```text
460 |     () => false,
461 |   );
462 |   expect(clearIfVersion).not.toHaveBeenCalled();
    |                              ^
463 |   expect(await memory.read(replacement.key)).toEqual(replacement);
464 | });
```


~~~~
FACT: R-018 | src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"key":"another-key"} | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 462 |   expect(clearIfVersion).not.toHaveBeenCalled(); / 463 |   expect(await memory.read(replacement.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: R-018 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/application/recovery.ts` exit=0.
FACT: R-021 test SHA-256=f06d85b1177079158747e52d9bbb734748189e888fe4cb66ad8ab80dc6a6d3b3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: R-021 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/application/recovery.test.ts`. Production SHA-256 5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69.
FACT: R-021 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-R-021-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"language":"es"}

repro:

```bash
npx rstest 'src/modules/transcript/application/recovery.test.ts' --testNamePattern 'GAP accepted cleanup preserves a same-revision snapshot differing in {"language":"es"}'
```

details:

```json
{
  "testPath": "src/modules/transcript/application/recovery.test.ts",
  "project": "dom",
  "fullName": "GAP accepted cleanup preserves a same-revision snapshot differing in {\"language\":\"es\"}",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"user:lesson\",\n      \"snapshot-revision\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/application/recovery.test.ts",
        "line": 462,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/application/recovery.test.ts",
      "line": 462
    }
  ]
}
```

codeFrame (error 1):

```text
460 |     () => false,
461 |   );
462 |   expect(clearIfVersion).not.toHaveBeenCalled();
    |                              ^
463 |   expect(await memory.read(replacement.key)).toEqual(replacement);
464 | });
```


~~~~
FACT: R-021 | src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"language":"es"} | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 462 |   expect(clearIfVersion).not.toHaveBeenCalled(); / 463 |   expect(await memory.read(replacement.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: R-021 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/application/recovery.ts` exit=0.
FACT: R-022 test SHA-256=f06d85b1177079158747e52d9bbb734748189e888fe4cb66ad8ab80dc6a6d3b3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: R-022 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/application/recovery.test.ts`. Production SHA-256 5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69.
FACT: R-022 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-R-022-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"baseLanguage":"fr"}

repro:

```bash
npx rstest 'src/modules/transcript/application/recovery.test.ts' --testNamePattern 'GAP accepted cleanup preserves a same-revision snapshot differing in {"baseLanguage":"fr"}'
```

details:

```json
{
  "testPath": "src/modules/transcript/application/recovery.test.ts",
  "project": "dom",
  "fullName": "GAP accepted cleanup preserves a same-revision snapshot differing in {\"baseLanguage\":\"fr\"}",
  "status": "fail",
  "duration": 6,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"user:lesson\",\n      \"snapshot-revision\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/application/recovery.test.ts",
        "line": 462,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/application/recovery.test.ts",
      "line": 462
    }
  ]
}
```

codeFrame (error 1):

```text
460 |     () => false,
461 |   );
462 |   expect(clearIfVersion).not.toHaveBeenCalled();
    |                              ^
463 |   expect(await memory.read(replacement.key)).toEqual(replacement);
464 | });
```


~~~~
FACT: R-022 | src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"baseLanguage":"fr"} | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 462 |   expect(clearIfVersion).not.toHaveBeenCalled(); / 463 |   expect(await memory.read(replacement.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: R-022 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/application/recovery.ts` exit=0.
FACT: R-023 test SHA-256=f06d85b1177079158747e52d9bbb734748189e888fe4cb66ad8ab80dc6a6d3b3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: R-023 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/application/recovery.test.ts`. Production SHA-256 5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69.
FACT: R-023 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-R-023-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"expectedUpdatedAt":"T2"}

repro:

```bash
npx rstest 'src/modules/transcript/application/recovery.test.ts' --testNamePattern 'GAP accepted cleanup preserves a same-revision snapshot differing in {"expectedUpdatedAt":"T2"}'
```

details:

```json
{
  "testPath": "src/modules/transcript/application/recovery.test.ts",
  "project": "dom",
  "fullName": "GAP accepted cleanup preserves a same-revision snapshot differing in {\"expectedUpdatedAt\":\"T2\"}",
  "status": "fail",
  "duration": 6,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"user:lesson\",\n      \"snapshot-revision\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/application/recovery.test.ts",
        "line": 462,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/application/recovery.test.ts",
      "line": 462
    }
  ]
}
```

codeFrame (error 1):

```text
460 |     () => false,
461 |   );
462 |   expect(clearIfVersion).not.toHaveBeenCalled();
    |                              ^
463 |   expect(await memory.read(replacement.key)).toEqual(replacement);
464 | });
```


~~~~
FACT: R-023 | src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"expectedUpdatedAt":"T2"} | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 462 |   expect(clearIfVersion).not.toHaveBeenCalled(); / 463 |   expect(await memory.read(replacement.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: R-023 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/application/recovery.ts` exit=0.
FACT: R-024 test SHA-256=f06d85b1177079158747e52d9bbb734748189e888fe4cb66ad8ab80dc6a6d3b3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: R-024 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/application/recovery.test.ts`. Production SHA-256 5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69.
FACT: R-024 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-R-024-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"draftUpdatedAt":"D2"}

repro:

```bash
npx rstest 'src/modules/transcript/application/recovery.test.ts' --testNamePattern 'GAP accepted cleanup preserves a same-revision snapshot differing in {"draftUpdatedAt":"D2"}'
```

details:

```json
{
  "testPath": "src/modules/transcript/application/recovery.test.ts",
  "project": "dom",
  "fullName": "GAP accepted cleanup preserves a same-revision snapshot differing in {\"draftUpdatedAt\":\"D2\"}",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st rstest.fn() call:\n\n    Array [\n      \"user:lesson\",\n      \"snapshot-revision\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/application/recovery.test.ts",
        "line": 462,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/application/recovery.test.ts",
      "line": 462
    }
  ]
}
```

codeFrame (error 1):

```text
460 |     () => false,
461 |   );
462 |   expect(clearIfVersion).not.toHaveBeenCalled();
    |                              ^
463 |   expect(await memory.read(replacement.key)).toEqual(replacement);
464 | });
```


~~~~
FACT: R-024 | src/modules/transcript/application/recovery.test.ts :: GAP accepted cleanup preserves a same-revision snapshot differing in {"draftUpdatedAt":"D2"} | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 462 |   expect(clearIfVersion).not.toHaveBeenCalled(); / 463 |   expect(await memory.read(replacement.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: R-024 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/application/recovery.ts` exit=0.
FACT: H-001 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-001 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-001 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-001-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP switching a loaded visit hides its offer until the new store read resolves

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP switching a loaded visit hides its offer until the new store read resolves'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP switching a loaded visit hides its offer until the new store read resolves",
  "status": "fail",
  "duration": 12,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected true to be false // Object.is equality",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 289,
        "column": 37,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 289
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- false
+ true
```

codeFrame (error 1):

```text
287 |   expect(f.result.current.offer).toMatchObject({ snapshot: a });
288 |   f.rerender({ key: 'gap-switch:b', source: other });
289 |   expect(f.result.current.isLoaded).toBe(false);
    |                                     ^
290 |   expect(f.result.current.offer).toEqual({ kind: 'none' });
291 |   await act(async () => next.resolve(null));
```


~~~~
FACT: H-001 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP switching a loaded visit hides its offer until the new store read resolves | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 287 |   expect(f.result.current.offer).toMatchObject({ snapshot: a }); / 289 |   expect(f.result.current.isLoaded).toBe(false); / 290 |   expect(f.result.current.offer).toEqual({ kind: 'none' }); | CLOSED | DEFECT=no
FACT: H-001 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-002 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-002 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-002 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-002-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete read completion cannot authorize retirement in an unread visit'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete read completion cannot authorize retirement in an unread visit",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-unread:b\",\n      \"e11a40c9-5ee6-419e-90e7-52d03b1ae695\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 333,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 333
    }
  ]
}
```

codeFrame (error 1):

```text
331 |   act(() => f.result.current.finalize());
332 |   await act(async () => {});
333 |   expect(clear).not.toHaveBeenCalled();
    |                     ^
334 |   await act(async () => b.resolve(null));
335 | });
```

### [F02] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalize cannot retire an own write before the initial read finishes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP finalize cannot retire an own write before the initial read finishes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP finalize cannot retire an own write before the initial read finishes",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-initial:finalize\",\n      \"b294c23b-60bd-4571-8c3f-54f6a20205e3\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 355,
        "column": 23,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 355
    }
  ]
}
```

codeFrame (error 1):

```text
353 |     });
354 |     await act(async () => {});
355 |     expect(clear).not.toHaveBeenCalled();
    |                       ^
356 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1');
357 |     await act(async () => pending.resolve(null));
```

### [F03] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP clean capture cannot retire an own write before the initial read finishes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP clean capture cannot retire an own write before the initial read finishes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP clean capture cannot retire an own write before the initial read finishes",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-initial:clean capture\",\n      \"0207ca7d-46dc-44e6-86f5-1e3c21c59914\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 355,
        "column": 23,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 355
    }
  ]
}
```

codeFrame (error 1):

```text
353 |     });
354 |     await act(async () => {});
355 |     expect(clear).not.toHaveBeenCalled();
    |                       ^
356 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1');
357 |     await act(async () => pending.resolve(null));
```


~~~~
FACT: H-002 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalize cannot retire an own write before the initial read finishes; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP clean capture cannot retire an own write before the initial read finishes | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 333 |   expect(clear).not.toHaveBeenCalled(); / 355 |     expect(clear).not.toHaveBeenCalled(); / 356 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1'); / 355 |     expect(clear).not.toHaveBeenCalled(); / 356 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1'); | CLOSED | DEFECT=no
FACT: H-002 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-003 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-003 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-003 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-003-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read resolve cannot replace the current visit offer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete initial read resolve cannot replace the current visit offer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete initial read resolve cannot replace the current visit offer",
  "status": "fail",
  "duration": 12,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'offer', snapshot: { …(7) } } to match object { snapshot: { …(7) } }\n(1 matching property omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 313,
        "column": 36,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 313
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
    "snapshot": {
      "baseLanguage": "en",
      "capturedAt": 1,
      "expectedUpdatedAt": "T1",
-     "key": "gap-obsolete-resolve:b",
+     "key": "obsolete-a",
      "language": "uk",
      "lines": [
        {
          "documentIds": [],
          "endMs": "1000",
          "id": "line",
          "startMs": "0",
          "text": "Recovered",
        },
      ],
-     "revision": "revision:gap-obsolete-resolve:b:Recovered",
+     "revision": "revision:obsolete-a:Recovered",
    },
  }
```

codeFrame (error 1):

```text
311 |     });
312 |     expect(read).toHaveBeenCalledTimes(2);
313 |     expect(f.result.current.offer).toMatchObject({ snapshot: b });
    |                                    ^
314 |     expect(f.result.current.isLoaded).toBe(true);
315 |   });
```


~~~~
FACT: H-003 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read resolve cannot replace the current visit offer | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 312 |     expect(read).toHaveBeenCalledTimes(2); / 313 |     expect(f.result.current.offer).toMatchObject({ snapshot: b }); / 314 |     expect(f.result.current.isLoaded).toBe(true); | CLOSED | DEFECT=no
FACT: H-003 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-004 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-004 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-004 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-004-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read resolve cannot replace the current visit offer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete initial read resolve cannot replace the current visit offer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete initial read resolve cannot replace the current visit offer",
  "status": "fail",
  "duration": 14,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'none' } to match object { snapshot: { …(7) } }\n(1 matching property omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 313,
        "column": 36,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 313
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "snapshot": {
-     "baseLanguage": "en",
-     "capturedAt": 1,
-     "expectedUpdatedAt": "T1",
-     "key": "gap-obsolete-resolve:b",
-     "language": "uk",
-     "lines": [
-       {
-         "documentIds": [],
-         "endMs": "1000",
-         "id": "line",
-         "startMs": "0",
-         "text": "Recovered",
-       },
-     ],
-     "revision": "revision:gap-obsolete-resolve:b:Recovered",
-   },
+   "kind": "none",
  }
```

codeFrame (error 1):

```text
311 |     });
312 |     expect(read).toHaveBeenCalledTimes(2);
313 |     expect(f.result.current.offer).toMatchObject({ snapshot: b });
    |                                    ^
314 |     expect(f.result.current.isLoaded).toBe(true);
315 |   });
```

### [F02] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read reject cannot replace the current visit offer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete initial read reject cannot replace the current visit offer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete initial read reject cannot replace the current visit offer",
  "status": "fail",
  "duration": 2,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'none' } to match object { snapshot: { …(7) } }\n(1 matching property omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 313,
        "column": 36,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 313
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "snapshot": {
-     "baseLanguage": "en",
-     "capturedAt": 1,
-     "expectedUpdatedAt": "T1",
-     "key": "gap-obsolete-reject:b",
-     "language": "uk",
-     "lines": [
-       {
-         "documentIds": [],
-         "endMs": "1000",
-         "id": "line",
-         "startMs": "0",
-         "text": "Recovered",
-       },
-     ],
-     "revision": "revision:gap-obsolete-reject:b:Recovered",
-   },
+   "kind": "none",
  }
```

codeFrame (error 1):

```text
311 |     });
312 |     expect(read).toHaveBeenCalledTimes(2);
313 |     expect(f.result.current.offer).toMatchObject({ snapshot: b });
    |                                    ^
314 |     expect(f.result.current.isLoaded).toBe(true);
315 |   });
```

### [F03] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete read completion cannot authorize retirement in an unread visit'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete read completion cannot authorize retirement in an unread visit",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-unread:b\",\n      \"68c66267-df14-44fd-ad72-f920893af022\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 333,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 333
    }
  ]
}
```

codeFrame (error 1):

```text
331 |   act(() => f.result.current.finalize());
332 |   await act(async () => {});
333 |   expect(clear).not.toHaveBeenCalled();
    |                     ^
334 |   await act(async () => b.resolve(null));
335 | });
```


~~~~
FACT: H-004 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read resolve cannot replace the current visit offer; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read reject cannot replace the current visit offer; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 312 |     expect(read).toHaveBeenCalledTimes(2); / 313 |     expect(f.result.current.offer).toMatchObject({ snapshot: b }); / 314 |     expect(f.result.current.isLoaded).toBe(true); / 312 |     expect(read).toHaveBeenCalledTimes(2); / 313 |     expect(f.result.current.offer).toMatchObject({ snapshot: b }); / 314 |     expect(f.result.current.isLoaded).toBe(true); / 333 |   expect(clear).not.toHaveBeenCalled(); | CLOSED | DEFECT=no
FACT: H-004 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-006 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-006 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-006 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-006-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalization waits for its write commit before conditional retirement

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP finalization waits for its write commit before conditional retirement'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP finalization waits for its write commit before conditional retirement",
  "status": "fail",
  "duration": 13,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-commit\",\n      \"d6dc8fb3-9e13-4c54-bdff-36536668a93d\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 377,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 377
    }
  ]
}
```

codeFrame (error 1):

```text
375 |   await act(async () => {});
376 |   expect(write).toHaveBeenCalledTimes(1);
377 |   expect(clear).not.toHaveBeenCalled();
    |                     ^
378 |   await act(async () => commit.resolve());
379 |   expect(clear).toHaveBeenCalledExactlyOnceWith('gap-commit', write.mock.calls[0]?.[0].revision);
```


~~~~
FACT: H-006 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalization waits for its write commit before conditional retirement | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 376 |   expect(write).toHaveBeenCalledTimes(1); / 377 |   expect(clear).not.toHaveBeenCalled(); / 379 |   expect(clear).toHaveBeenCalledExactlyOnceWith('gap-commit', write.mock.calls[0]?.[0].revision); | CLOSED | DEFECT=no
FACT: H-006 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-008 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-008 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-008 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-008-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: MS3b fix6 ownership Save / newer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'MS3b fix6 ownership Save / newer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "MS3b fix6 ownership Save / newer",
  "status": "fail",
  "duration": 68,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'offer', snapshot: { …(7) } } to deeply equal { kind: 'none' }",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 212,
        "column": 40,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 212
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "kind": "none",
+   "kind": "offer",
+   "snapshot": {
+     "baseLanguage": "en",
+     "capturedAt": 1794929233260,
+     "expectedUpdatedAt": "T1",
+     "key": "fix6:Save:newer",
+     "language": "uk",
+     "lines": [
+       {
+         "documentIds": [],
+         "endMs": "1000",
+         "id": "line",
+         "startMs": "0",
+         "text": "C",
+       },
+     ],
+     "revision": "4ddff5ac-367a-49f5-85c2-7f4e3e323096",
+   },
  }
```

codeFrame (error 1):

```text
210 |       expect(await store.read(key)).toEqual(version === 'same' ? null : c);
211 |       if (version === 'newer' && path === 'Save')
212 |         expect(f.result.current.offer).toEqual({ kind: 'none' });
    |                                        ^
213 |       if (version === 'newer' && path !== 'unmount' && path !== 'Save')
214 |         expect(f.result.current.offer).toMatchObject({ kind: 'offer', snapshot: c });
```


~~~~
FACT: H-008 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: MS3b fix6 ownership Save / newer | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 210 |       expect(await store.read(key)).toEqual(version === 'same' ? null : c); / 212 |         expect(f.result.current.offer).toEqual({ kind: 'none' }); / 214 |         expect(f.result.current.offer).toMatchObject({ kind: 'offer', snapshot: c }); | CLOSED | DEFECT=no
FACT: H-008 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-009 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-009 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-009 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-009-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP mismatch refresh from an obsolete visit cannot replace the current offer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP mismatch refresh from an obsolete visit cannot replace the current offer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP mismatch refresh from an obsolete visit cannot replace the current offer",
  "status": "fail",
  "duration": 12,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'offer', snapshot: { …(7) } } to match object { snapshot: { …(7) } }\n(1 matching property omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 421,
        "column": 34,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 421
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
    "snapshot": {
      "baseLanguage": "en",
      "capturedAt": 1,
      "expectedUpdatedAt": "T1",
-     "key": "gap-refresh:b",
+     "key": "gap-refresh:a",
      "language": "uk",
      "lines": [
        {
          "documentIds": [],
          "endMs": "1000",
          "id": "line",
          "startMs": "0",
-         "text": "Recovered",
+         "text": "foreign",
        },
      ],
-     "revision": "revision:gap-refresh:b:Recovered",
+     "revision": "foreign",
    },
  }
```

codeFrame (error 1):

```text
419 |     delayed.resolve({ ...a, revision: 'foreign', lines: [{ ...line, text: 'foreign' }] }),
420 |   );
421 |   expect(f.result.current.offer).toMatchObject({ snapshot: b });
    |                                  ^
422 | });
423 | 
```


~~~~
FACT: H-009 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP mismatch refresh from an obsolete visit cannot replace the current offer | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 421 |   expect(f.result.current.offer).toMatchObject({ snapshot: b }); | CLOSED | DEFECT=no
FACT: H-009 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-010 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-010 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-010 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-010-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP repeated Discard retires the refreshed revision rather than its predecessor

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP repeated Discard retires the refreshed revision rather than its predecessor'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP repeated Discard retires the refreshed revision rather than its predecessor",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to be called once with arguments: [ 'gap-discard', 'C' ], but got []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 439,
        "column": 17,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 439
    }
  ]
}
```

codeFrame (error 1):

```text
437 |   act(() => f.result.current.dismiss());
438 |   await act(async () => {});
439 |   expect(clear).toHaveBeenCalledExactlyOnceWith(c.key, 'C');
    |                 ^
440 |   expect(await store.read(c.key)).toBeNull();
441 | });
```


~~~~
FACT: H-010 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP repeated Discard retires the refreshed revision rather than its predecessor | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 439 |   expect(clear).toHaveBeenCalledExactlyOnceWith(c.key, 'C'); / 440 |   expect(await store.read(c.key)).toBeNull(); | CLOSED | DEFECT=no
FACT: H-010 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-013 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-013 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-013 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-013-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended captureNow does not arm capture when resumed or unmounted

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP suspended captureNow does not arm capture when resumed or unmounted'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP suspended captureNow does not arm capture when resumed or unmounted",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"write\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st write call:\n\n    Array [\n      Object {\n        \"baseLanguage\": \"en\",\n        \"capturedAt\": 1794929239657,\n        \"expectedUpdatedAt\": \"T1\",\n        \"key\": \"user:lesson:track:a\",\n        \"language\": \"uk\",\n        \"lines\": Array [\n          Object {\n            \"documentIds\": Array [],\n            \"endMs\": \"1000\",\n            \"id\": \"line\",\n            \"startMs\": \"0\",\n            \"text\": \"suspended edit\",\n          },\n        ],\n        \"revision\": \"1548a31a-db5f-4c97-b6aa-cf7265043117\",\n      },\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 463,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 463
    }
  ]
}
```

codeFrame (error 1):

```text
461 |     act(() => f.result.current.setCapturing(false));
462 |     act(() => f.result.current[action]([{ ...line, text: 'suspended edit' }]));
463 |     expect(f.write).not.toHaveBeenCalled();
    |                         ^
464 |     act(() => f.result.current.setCapturing(true));
465 |     act(() => rstest.advanceTimersByTime(idleCaptureMs));
```


~~~~
FACT: H-013 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended captureNow does not arm capture when resumed or unmounted | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 463 |     expect(f.write).not.toHaveBeenCalled(); | CLOSED | DEFECT=no
FACT: H-013 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-015 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-015 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-015 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-015-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP clean capture cannot retire an own write before the initial read finishes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP clean capture cannot retire an own write before the initial read finishes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP clean capture cannot retire an own write before the initial read finishes",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-initial:clean capture\",\n      \"456850e4-2461-457c-8851-38ba6f3fa496\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 355,
        "column": 23,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 355
    }
  ]
}
```

codeFrame (error 1):

```text
353 |     });
354 |     await act(async () => {});
355 |     expect(clear).not.toHaveBeenCalled();
    |                       ^
356 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1');
357 |     await act(async () => pending.resolve(null));
```


~~~~
FACT: H-015 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP clean capture cannot retire an own write before the initial read finishes | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 355 |     expect(clear).not.toHaveBeenCalled(); / 356 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1'); | CLOSED | DEFECT=no
FACT: H-015 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-025 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-025 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-025 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-025-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete read completion cannot authorize retirement in an unread visit'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete read completion cannot authorize retirement in an unread visit",
  "status": "fail",
  "duration": 12,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-unread:b\",\n      \"b413738b-14a8-4383-9691-cdbe6e336986\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 333,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 333
    }
  ]
}
```

codeFrame (error 1):

```text
331 |   act(() => f.result.current.finalize());
332 |   await act(async () => {});
333 |   expect(clear).not.toHaveBeenCalled();
    |                     ^
334 |   await act(async () => b.resolve(null));
335 | });
```

### [F02] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalize cannot retire an own write before the initial read finishes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP finalize cannot retire an own write before the initial read finishes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP finalize cannot retire an own write before the initial read finishes",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-initial:finalize\",\n      \"c28980e5-8580-413d-badf-7dce1868f8ae\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 355,
        "column": 23,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 355
    }
  ]
}
```

codeFrame (error 1):

```text
353 |     });
354 |     await act(async () => {});
355 |     expect(clear).not.toHaveBeenCalled();
    |                       ^
356 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1');
357 |     await act(async () => pending.resolve(null));
```


~~~~
FACT: H-025 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete read completion cannot authorize retirement in an unread visit; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalize cannot retire an own write before the initial read finishes | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 333 |   expect(clear).not.toHaveBeenCalled(); / 355 |     expect(clear).not.toHaveBeenCalled(); / 356 |     expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1'); | CLOSED | DEFECT=no
FACT: H-025 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-026 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-026 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-026 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-026-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalization without an own write retains the accepted recovery source

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP finalization without an own write retains the accepted recovery source'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP finalization without an own write retains the accepted recovery source",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"clearIfVersion\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st clearIfVersion call:\n\n    Array [\n      \"gap-accepted\",\n      \"revision:gap-accepted:Recovered\",\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 393,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 393
    }
  ]
}
```

codeFrame (error 1):

```text
391 |   act(() => f.result.current.finalize());
392 |   await act(async () => {});
393 |   expect(clear).not.toHaveBeenCalled();
    |                     ^
394 |   expect(await store.read(a.key)).toEqual(a);
395 | });
```


~~~~
FACT: H-026 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP finalization without an own write retains the accepted recovery source | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 393 |   expect(clear).not.toHaveBeenCalled(); / 394 |   expect(await store.read(a.key)).toEqual(a); | CLOSED | DEFECT=no
FACT: H-026 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-028 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-028 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-028 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-028-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended captureNow does not arm capture when resumed or unmounted

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP suspended captureNow does not arm capture when resumed or unmounted'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP suspended captureNow does not arm capture when resumed or unmounted",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"write\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st write call:\n\n    Array [\n      Object {\n        \"baseLanguage\": \"en\",\n        \"capturedAt\": 1794929248104,\n        \"expectedUpdatedAt\": \"T1\",\n        \"key\": \"user:lesson:track:a\",\n        \"language\": \"uk\",\n        \"lines\": Array [\n          Object {\n            \"documentIds\": Array [],\n            \"endMs\": \"1000\",\n            \"id\": \"line\",\n            \"startMs\": \"0\",\n            \"text\": \"suspended edit\",\n          },\n        ],\n        \"revision\": \"4db31da7-a403-449f-83ca-f70d9baacc87\",\n      },\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 463,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 463
    }
  ]
}
```

codeFrame (error 1):

```text
461 |     act(() => f.result.current.setCapturing(false));
462 |     act(() => f.result.current[action]([{ ...line, text: 'suspended edit' }]));
463 |     expect(f.write).not.toHaveBeenCalled();
    |                         ^
464 |     act(() => f.result.current.setCapturing(true));
465 |     act(() => rstest.advanceTimersByTime(idleCaptureMs));
```

### [F02] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended noteChange does not arm capture when resumed or unmounted

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP suspended noteChange does not arm capture when resumed or unmounted'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP suspended noteChange does not arm capture when resumed or unmounted",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"write\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st write call:\n\n    Array [\n      Object {\n        \"baseLanguage\": \"en\",\n        \"capturedAt\": 1794929338113,\n        \"expectedUpdatedAt\": \"T1\",\n        \"key\": \"user:lesson:track:a\",\n        \"language\": \"uk\",\n        \"lines\": Array [\n          Object {\n            \"documentIds\": Array [],\n            \"endMs\": \"1000\",\n            \"id\": \"line\",\n            \"startMs\": \"0\",\n            \"text\": \"suspended edit\",\n          },\n        ],\n        \"revision\": \"40b16da2-932e-49de-9de7-6de6b0ea486e\",\n      },\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 467,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 467
    }
  ]
}
```

codeFrame (error 1):

```text
465 |     act(() => rstest.advanceTimersByTime(idleCaptureMs));
466 |     f.unmount();
467 |     expect(f.write).not.toHaveBeenCalled();
    |                         ^
468 |   });
469 | }
```


~~~~
FACT: H-028 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended captureNow does not arm capture when resumed or unmounted; src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended noteChange does not arm capture when resumed or unmounted | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 463 |     expect(f.write).not.toHaveBeenCalled(); / 467 |     expect(f.write).not.toHaveBeenCalled(); | CLOSED | DEFECT=no
FACT: H-028 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: H-029 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: H-029 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: H-029 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-H-029-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP switching a loaded visit hides its offer until the new store read resolves

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP switching a loaded visit hides its offer until the new store read resolves'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP switching a loaded visit hides its offer until the new store read resolves",
  "status": "fail",
  "duration": 15,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'offer', snapshot: { …(7) } } to deeply equal { kind: 'none' }",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 290,
        "column": 34,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 290
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "kind": "none",
+   "kind": "offer",
+   "snapshot": {
+     "baseLanguage": "en",
+     "capturedAt": 1,
+     "expectedUpdatedAt": "T1",
+     "key": "gap-switch:a",
+     "language": "uk",
+     "lines": [
+       {
+         "documentIds": [],
+         "endMs": "1000",
+         "id": "line",
+         "startMs": "0",
+         "text": "Recovered",
+       },
+     ],
+     "revision": "revision:gap-switch:a:Recovered",
+   },
  }
```

codeFrame (error 1):

```text
288 |   f.rerender({ key: 'gap-switch:b', source: other });
289 |   expect(f.result.current.isLoaded).toBe(false);
290 |   expect(f.result.current.offer).toEqual({ kind: 'none' });
    |                                  ^
291 |   await act(async () => next.resolve(null));
292 |   expect(f.result.current.isLoaded).toBe(true);
```


~~~~
FACT: H-029 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP switching a loaded visit hides its offer until the new store read resolves | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 289 |   expect(f.result.current.isLoaded).toBe(false); / 290 |   expect(f.result.current.offer).toEqual({ kind: 'none' }); / 292 |   expect(f.result.current.isLoaded).toBe(true); | CLOSED | DEFECT=no
FACT: H-029 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: U-001 test SHA-256=75fcfb69be6bcbfa10743f8419a65588eb37fe9ca8b3d3d60fc8deab85eca467; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: U-001 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: U-001 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-U-001-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP disabled detached inventory performs no storage IO

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP disabled detached inventory performs no storage IO'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP disabled detached inventory performs no storage IO",
  "status": "fail",
  "duration": 16,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"keys\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st keys call:\n\n    Array []\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 28,
        "column": 20,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 28
    }
  ]
}
```

codeFrame (error 1):

```text
26 |   );
27 |   await act(async () => {});
28 |   expect(keys).not.toHaveBeenCalled();
   |                    ^
29 |   expect(read).not.toHaveBeenCalled();
30 |   expect(f.result.current).toEqual([]);
```


~~~~
FACT: U-001 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP disabled detached inventory performs no storage IO | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 28 |   expect(keys).not.toHaveBeenCalled(); / 29 |   expect(read).not.toHaveBeenCalled(); / 30 |   expect(f.result.current).toEqual([]); | CLOSED | DEFECT=no
FACT: U-001 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
FACT: U-003 test SHA-256=75fcfb69be6bcbfa10743f8419a65588eb37fe9ca8b3d3d60fc8deab85eca467; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: U-003 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: U-003 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-U-003-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory ignores a read result belonging to a different listed key

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP detached inventory ignores a read result belonging to a different listed key'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP detached inventory ignores a read result belonging to a different listed key",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { instanceId: 'track:en', …(2) } ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 42,
        "column": 28,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 42
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   {
+     "instanceId": "track:en",
+     "legacy": false,
+     "snapshot": {
+       "capturedAt": 1,
+       "expectedUpdatedAt": "T1",
+       "key": "author:lesson:track:gone",
+       "language": "en",
+       "lines": [
+         {
+           "documentIds": [],
+           "endMs": "1000",
+           "id": "line",
+           "startMs": "0",
+           "text": "A",
+         },
+       ],
+       "revision": "A",
+     },
+   },
+ ]
```

codeFrame (error 1):

```text
40 |   await act(async () => {});
41 |   expect(read).toHaveBeenCalledExactlyOnceWith('author:lesson:track:listed');
42 |   expect(f.result.current).toEqual([]);
   |                            ^
43 | });
44 | 
```


~~~~
FACT: U-003 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory ignores a read result belonging to a different listed key | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 41 |   expect(read).toHaveBeenCalledExactlyOnceWith('author:lesson:track:listed'); / 42 |   expect(f.result.current).toEqual([]); | CLOSED | DEFECT=no
FACT: U-003 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
FACT: U-004 test SHA-256=75fcfb69be6bcbfa10743f8419a65588eb37fe9ca8b3d3d60fc8deab85eca467; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: U-004 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: U-004 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-U-004-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP obsolete detached inventory cannot erase the completed current inventory

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete detached inventory cannot erase the completed current inventory'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete detached inventory cannot erase the completed current inventory",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [] to deeply equal [ { …(6) } ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 58,
        "column": 59,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 58
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- [
-   {
-     "capturedAt": 1,
-     "expectedUpdatedAt": "T1",
-     "key": "author:lesson:track:gone",
-     "language": "en",
-     "lines": [
-       {
-         "documentIds": [],
-         "endMs": "1000",
-         "id": "line",
-         "startMs": "0",
-         "text": "A",
-       },
-     ],
-     "revision": "A",
-   },
- ]
+ []
```

codeFrame (error 1):

```text
56 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);
57 |   await act(async () => old.resolve([]));
58 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);
   |                                                           ^
59 | });
60 | 
```


~~~~
FACT: U-004 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP obsolete detached inventory cannot erase the completed current inventory | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 56 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]); / 58 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]); | CLOSED | DEFECT=no
FACT: U-004 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
FACT: U-005 test SHA-256=75fcfb69be6bcbfa10743f8419a65588eb37fe9ca8b3d3d60fc8deab85eca467; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: U-005 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: U-005 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-U-005-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after store changes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP detached inventory hides stale same-lesson offers after store changes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP detached inventory hides stale same-lesson offers after store changes",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { instanceId: 'track:en', …(2) } ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 91,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 91
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   {
+     "instanceId": "track:en",
+     "legacy": false,
+     "snapshot": {
+       "capturedAt": 1,
+       "expectedUpdatedAt": "T1",
+       "key": "author:lesson:track:gone",
+       "language": "en",
+       "lines": [
+         {
+           "documentIds": [],
+           "endMs": "1000",
+           "id": "line",
+           "startMs": "0",
+           "text": "A",
+         },
+       ],
+       "revision": "A",
+     },
+   },
+ ]
```

codeFrame (error 1):

```text
89 |       enabled: change !== 'disabled',
90 |     });
91 |     expect(f.result.current).toEqual([]);
   |                              ^
92 |     await act(async () => pending.resolve([]));
93 |   });
```

### [F02] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after tabs changes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP detached inventory hides stale same-lesson offers after tabs changes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP detached inventory hides stale same-lesson offers after tabs changes",
  "status": "fail",
  "duration": 2,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { instanceId: 'track:en', …(2) } ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 91,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 91
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   {
+     "instanceId": "track:en",
+     "legacy": false,
+     "snapshot": {
+       "capturedAt": 1,
+       "expectedUpdatedAt": "T1",
+       "key": "author:lesson:track:gone",
+       "language": "en",
+       "lines": [
+         {
+           "documentIds": [],
+           "endMs": "1000",
+           "id": "line",
+           "startMs": "0",
+           "text": "A",
+         },
+       ],
+       "revision": "A",
+     },
+   },
+ ]
```

codeFrame (error 1):

```text
89 |       enabled: change !== 'disabled',
90 |     });
91 |     expect(f.result.current).toEqual([]);
   |                              ^
92 |     await act(async () => pending.resolve([]));
93 |   });
```

### [F03] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after disabled changes

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP detached inventory hides stale same-lesson offers after disabled changes'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP detached inventory hides stale same-lesson offers after disabled changes",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { instanceId: 'track:en', …(2) } ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 91,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 91
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   {
+     "instanceId": "track:en",
+     "legacy": false,
+     "snapshot": {
+       "capturedAt": 1,
+       "expectedUpdatedAt": "T1",
+       "key": "author:lesson:track:gone",
+       "language": "en",
+       "lines": [
+         {
+           "documentIds": [],
+           "endMs": "1000",
+           "id": "line",
+           "startMs": "0",
+           "text": "A",
+         },
+       ],
+       "revision": "A",
+     },
+   },
+ ]
```

codeFrame (error 1):

```text
89 |       enabled: change !== 'disabled',
90 |     });
91 |     expect(f.result.current).toEqual([]);
   |                              ^
92 |     await act(async () => pending.resolve([]));
93 |   });
```


~~~~
FACT: U-005 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after store changes; src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after tabs changes; src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP detached inventory hides stale same-lesson offers after disabled changes | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 91 |     expect(f.result.current).toEqual([]); / 91 |     expect(f.result.current).toEqual([]); / 91 |     expect(f.result.current).toEqual([]); | CLOSED | DEFECT=no
FACT: U-005 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
FACT: U-009 test SHA-256=75fcfb69be6bcbfa10743f8419a65588eb37fe9ca8b3d3d60fc8deab85eca467; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: U-009 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: U-009 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-U-009-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP one refused detached read does not suppress a different valid recovery

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP one refused detached read does not suppress a different valid recovery'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP one refused detached read does not suppress a different valid recovery",
  "status": "fail",
  "duration": 6,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [] to deeply equal [ { …(6) } ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 108,
        "column": 59,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 108
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- [
-   {
-     "capturedAt": 1,
-     "expectedUpdatedAt": "T1",
-     "key": "author:lesson:track:gone",
-     "language": "en",
-     "lines": [
-       {
-         "documentIds": [],
-         "endMs": "1000",
-         "id": "line",
-         "startMs": "0",
-         "text": "A",
-       },
-     ],
-     "revision": "A",
-   },
- ]
+ []
```

codeFrame (error 1):

```text
106 |   await act(async () => {});
107 |   expect(read).toHaveBeenCalledTimes(2);
108 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);
    |                                                           ^
109 | });
110 | 
```


~~~~
FACT: U-009 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP one refused detached read does not suppress a different valid recovery | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 107 |   expect(read).toHaveBeenCalledTimes(2); / 108 |   expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]); | CLOSED | DEFECT=no
FACT: U-009 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
FACT: L-001 test SHA-256=87df0aab9beaeadc8f3b056ec3959ca8c78f4f084ea5bd2906c4238695f73467; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: L-001 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/LessonDraftField.test.tsx`. Production SHA-256 c5ab86ff38e890b866e6e779d977065143be2a46758d23b8cbb07c0aebdc631c.
FACT: L-001 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-L-001-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/LessonDraftField.test.tsx :: GAP a missing shown track reports not found after Regenerate

repro:

```bash
npx rstest 'src/modules/transcript/ui/LessonDraftField.test.tsx' --testNamePattern 'GAP a missing shown track reports not found after Regenerate'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/LessonDraftField.test.tsx",
  "project": "dom",
  "fullName": "GAP a missing shown track reports not found after Regenerate",
  "status": "fail",
  "duration": 20,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 'Recognition service is temporarily un…' to be 'Save the course and upload a video be…' // Object.is equality",
      "topFrame": {
        "file": "src/modules/transcript/ui/LessonDraftField.test.tsx",
        "line": 272,
        "column": 58,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/LessonDraftField.test.tsx",
      "line": 272
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- Save the course and upload a video before generating.
+ Recognition service is temporarily unavailable. Try again later.
```

codeFrame (error 1):

```text
270 |   renderField(draft('ready'), draftActions({ regenerate: async () => 'track_not_found' }));
271 |   await userEvent.setup().click(screen.getByRole('button', { name: 'Regenerate transcript' }));
272 |   expect((await screen.findByRole('alert')).textContent).toBe(
    |                                                          ^
273 |     'Save the course and upload a video before generating.',
274 |   );
```


~~~~
FACT: L-001 | src/modules/transcript/ui/LessonDraftField.test.tsx :: GAP a missing shown track reports not found after Regenerate | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 272 |   expect((await screen.findByRole('alert')).textContent).toBe( | CLOSED | DEFECT=no
FACT: L-001 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/LessonDraftField.tsx` exit=0.
FACT: F-002 test SHA-256=a1ea6e3097b9f8a9f65ff60bac22b5a71e4d0b9375683cd8ead24d177122a7d0; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: F-002 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/LessonTranscriptField.test.tsx`. Production SHA-256 c2c97d5c0e1af77d2814534ab1bae82b0d8f7fc640cc5574003efff0a07f606f.
FACT: F-002 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-F-002-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/LessonTranscriptField.test.tsx :: GAP first transcript load then renamed-language load retains the dirty editor instance

repro:

```bash
npx rstest 'src/modules/transcript/ui/LessonTranscriptField.test.tsx' --testNamePattern 'GAP first transcript load then renamed-language load retains the dirty editor instance'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
  "project": "dom",
  "fullName": "GAP first transcript load then renamed-language load retains the dirty editor instance",
  "status": "fail",
  "duration": 88,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find a label with the text of: Текст репліки 1\n\nIgnored nodes: comments, script, style\n<body>\n  <div />\n</body>",
      "topFrame": {
        "file": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
        "line": 39,
        "column": 10,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
      "line": 39
    }
  ]
}
```

codeFrame (error 1):

```text
37 |   loading = true;
38 |   f.rerender(<LessonTranscriptField {...props} fixedLanguage="es" />);
39 |   expect(screen.getByLabelText('Текст репліки 1')).toBe(text);
   |          ^
40 |   expect(screen.getByDisplayValue('Unsaved')).toBe(text);
41 | });
```


~~~~
FACT: F-002 | src/modules/transcript/ui/LessonTranscriptField.test.tsx :: GAP first transcript load then renamed-language load retains the dirty editor instance | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 39 |   expect(screen.getByLabelText('Текст репліки 1')).toBe(text); / 40 |   expect(screen.getByDisplayValue('Unsaved')).toBe(text); | CLOSED | DEFECT=no
FACT: F-002 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/LessonTranscriptField.tsx` exit=0.
FACT: F-009 test SHA-256=a1ea6e3097b9f8a9f65ff60bac22b5a71e4d0b9375683cd8ead24d177122a7d0; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: F-009 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/LessonTranscriptField.test.tsx`. Production SHA-256 c2c97d5c0e1af77d2814534ab1bae82b0d8f7fc640cc5574003efff0a07f606f.
FACT: F-009 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-F-009-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/LessonTranscriptField.test.tsx :: GAP fixed tab language keeps the returned transcript token attached to its real base language

repro:

```bash
npx rstest 'src/modules/transcript/ui/LessonTranscriptField.test.tsx' --testNamePattern 'GAP fixed tab language keeps the returned transcript token attached to its real base language'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
  "project": "dom",
  "fullName": "GAP fixed tab language keeps the returned transcript token attached to its real base language",
  "status": "fail",
  "duration": 25,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to match object { language: 'es', …(2) }\n(4 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
        "line": 64,
        "column": 55,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/LessonTranscriptField.test.tsx",
      "line": 64
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "baseLanguage": "en",
+   "baseLanguage": "es",
    "expectedUpdatedAt": "T1",
    "language": "es",
  }
```

codeFrame (error 1):

```text
62 |   fireEvent.focusOut(text);
63 |   await act(async () => {});
64 |   expect(await store.read('author:lesson:track:one')).toMatchObject({
   |                                                       ^
65 |     language: 'es',
66 |     baseLanguage: 'en',
```


~~~~
FACT: F-009 | src/modules/transcript/ui/LessonTranscriptField.test.tsx :: GAP fixed tab language keeps the returned transcript token attached to its real base language | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 64 |   expect(await store.read('author:lesson:track:one')).toMatchObject({ | CLOSED | DEFECT=no
FACT: F-009 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/LessonTranscriptField.tsx` exit=0.
FACT: E-001 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-001 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-001 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-001-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a source acquired by a mounted tab is removed from captured cleanup ancestry

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a source acquired by a mounted tab is removed from captured cleanup ancestry'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a source acquired by a mounted tab is removed from captured cleanup ancestry",
  "status": "fail",
  "duration": 21,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { key: 'acquired-source', …(1) } ] to be undefined",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1718,
        "column": 79,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1718
    }
  ]
}
```

diff:

```diff
- Expected:
undefined

+ Received:
[
  {
    "key": "acquired-source",
    "revision": "revision:acquired-source",
  },
]
```

codeFrame (error 1):

```text
1716 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1717 |   await act(async () => {});
1718 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined();
     |                                                                               ^
1719 |   view.rerender(<TranscriptEditor {...props} detachedSnapshots={offers} mountedHistoryKeys={[]} />);
1720 |   clickSave();
```


~~~~
FACT: E-001 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a source acquired by a mounted tab is removed from captured cleanup ancestry | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1718 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined(); | CLOSED | DEFECT=no
FACT: E-001 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-002 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-002 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-002 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-002-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP source cleanup rechecks newly mounted ownership after its deferred read

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP source cleanup rechecks newly mounted ownership after its deferred read'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP source cleanup rechecks newly mounted ownership after its deferred read",
  "status": "fail",
  "duration": 22,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ [ 'newly-mounted-source', …(1) ] ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1764,
        "column": 66,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1764
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   [
+     "newly-mounted-source",
+     "revision:newly-mounted-source",
+   ],
+ ]
```

codeFrame (error 1):

```text
1762 |   );
1763 |   await act(async () => pending.resolve(source));
1764 |   expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]);
     |                                                                  ^
1765 |   expect(await memory.read(source.key)).toEqual(source);
1766 | });
```


~~~~
FACT: E-002 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP source cleanup rechecks newly mounted ownership after its deferred read | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1764 |   expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]); / 1765 |   expect(await memory.read(source.key)).toEqual(source); | CLOSED | DEFECT=no
FACT: E-002 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-003 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-003 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-003 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-003-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a completed Save lets the next canonical server token become the next edit base

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a completed Save lets the next canonical server token become the next edit base'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a completed Save lets the next canonical server token become the next edit base",
  "status": "fail",
  "duration": 24,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected last \"rstest.fn()\" call to have been called with [ 'uk', 'T2', Any<Array>, 'uk' ], but got [ 'uk', 'T1', [ { …(5) } ], 'uk' ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1456,
        "column": 30,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1456
    }
  ]
}
```

codeFrame (error 1):

```text
1454 |   clickSave();
1455 |   await act(async () => {});
1456 |   expect(props.actions.save).toHaveBeenLastCalledWith('uk', 'T2', expect.any(Array), 'uk');
     |                              ^
1457 | });
1458 | 
```


~~~~
FACT: E-003 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a completed Save lets the next canonical server token become the next edit base | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1456 |   expect(props.actions.save).toHaveBeenLastCalledWith('uk', 'T2', expect.any(Array), 'uk'); | CLOSED | DEFECT=no
FACT: E-003 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-005 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-005 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-005 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-005-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only base language

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a clean editor adopts a server change to only base language'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a clean editor adopts a server change to only base language",
  "status": "fail",
  "duration": 22,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to be called with arguments: [ Array(4) ], but got [ [ 'uk', 'T1', [ { …(5) } ], 'uk' ] ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1429,
        "column": 32,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1429
    }
  ]
}
```

codeFrame (error 1):

```text
1427 |     clickSave();
1428 |     await act(async () => {});
1429 |     expect(props.actions.save).toHaveBeenCalledWith(
     |                                ^
1430 |       'uk',
1431 |       change === 'token' ? 'T2' : 'T1',
```


~~~~
FACT: E-005 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only base language | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1429 |     expect(props.actions.save).toHaveBeenCalledWith( | CLOSED | DEFECT=no
FACT: E-005 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-006 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-006 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-006 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-006-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only token

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a clean editor adopts a server change to only token'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a clean editor adopts a server change to only token",
  "status": "fail",
  "duration": 20,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rstest.fn()\" to be called with arguments: [ Array(4) ], but got [ [ 'uk', 'T1', [ { …(5) } ], 'uk' ] ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1429,
        "column": 32,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1429
    }
  ]
}
```

codeFrame (error 1):

```text
1427 |     clickSave();
1428 |     await act(async () => {});
1429 |     expect(props.actions.save).toHaveBeenCalledWith(
     |                                ^
1430 |       'uk',
1431 |       change === 'token' ? 'T2' : 'T1',
```


~~~~
FACT: E-006 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only token | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1429 |     expect(props.actions.save).toHaveBeenCalledWith( | CLOSED | DEFECT=no
FACT: E-006 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-007 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-007 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-007 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-007-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only segments

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a clean editor adopts a server change to only segments'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a clean editor adopts a server change to only segments",
  "status": "fail",
  "duration": 21,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <textarea tabindex=\"0\" …(5)></textarea> to have property \"value\" with value 'New server'",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1425,
        "column": 56,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1425
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- New server
+ Перша
```

codeFrame (error 1):

```text
1423 |     );
1424 |     if (change === 'segments')
1425 |       expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'New server');
     |                                                        ^
1426 |     fireEvent.change(screen.getByLabelText('Текст репліки 1'), { target: { value: 'Edited' } });
1427 |     clickSave();
```


~~~~
FACT: E-007 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a clean editor adopts a server change to only segments | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1425 |       expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'New server'); | CLOSED | DEFECT=no
FACT: E-007 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-008 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-008 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-008 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-008-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix 5 a clean track rename followed by editing is not a dirty track rename

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'MS3b fix 5 a clean track rename followed by editing is not a dirty track rename'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "MS3b fix 5 a clean track rename followed by editing is not a dirty track rename",
  "status": "fail",
  "duration": 21,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <p></p> to be null",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 631,
        "column": 84,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 631
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
<p>
  Мова доріжки тепер en; рядки буде збережено як en.
</p>
```

codeFrame (error 1):

```text
629 |     target: { value: 'Edited after rename' },
630 |   });
631 |   expect(screen.queryByText('Мова доріжки тепер en; рядки буде збережено як en.')).toBeNull();
    |                                                                                    ^
632 | });
633 | 
```


~~~~
FACT: E-008 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix 5 a clean track rename followed by editing is not a dirty track rename | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 631 |   expect(screen.queryByText('Мова доріжки тепер en; рядки буде збережено як en.')).toBeNull(); | CLOSED | DEFECT=no
FACT: E-008 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-013 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-013 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-013 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-013-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP dirty edits compare against their retained base after a server prop refresh

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP dirty edits compare against their retained base after a server prop refresh'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP dirty edits compare against their retained base after a server prop refresh",
  "status": "fail",
  "duration": 22,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to match object { expectedUpdatedAt: 'T1', …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1477,
        "column": 61,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1477
    }
  ]
}
```

diff:

```diff
- Expected:
{
  "expectedUpdatedAt": "T1",
  "lines": [
    {
      "text": "Server moved",
    },
  ],
}

+ Received:
null
```

codeFrame (error 1):

```text
1475 |   fireEvent.focusOut(text);
1476 |   await act(async () => {});
1477 |   expect(await props.history.store.read(props.history.key)).toMatchObject({
     |                                                             ^
1478 |     expectedUpdatedAt: 'T1',
1479 |     lines: [{ text: 'Server moved' }],
```


~~~~
FACT: E-013 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP dirty edits compare against their retained base after a server prop refresh | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1477 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ | CLOSED | DEFECT=no
FACT: E-013 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-017 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-017 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-017 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-017-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a pending Save locks fields, documents, line controls and repeat Save

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a pending Save locks fields, documents, line controls and repeat Save'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a pending Save locks fields, documents, line controls and repeat Save",
  "status": "fail",
  "duration": 20,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <input tabindex=\"0\" …(7)></input> to have property \"readOnly\" with value true",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1493,
        "column": 19,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1493
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- true
+ false
```

codeFrame (error 1):

```text
1491 |   for (const label of ['Початок репліки 1, мс', 'Кінець репліки 1, мс', 'Текст репліки 1']) {
1492 |     const field = screen.getByLabelText(label);
1493 |     expect(field).toHaveProperty('readOnly', true);
     |                   ^
1494 |     fireEvent.change(field, { target: { value: '99' } });
1495 |   }
```


~~~~
FACT: E-017 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a pending Save locks fields, documents, line controls and repeat Save | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1493 |     expect(field).toHaveProperty('readOnly', true); | CLOSED | DEFECT=no
FACT: E-017 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-030 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-030 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-030 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-030-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry distinguishes two revisions of the same key within its eight-source limit

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP ancestry distinguishes two revisions of the same key within its eight-source limit'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP ancestry distinguishes two revisions of the same key within its eight-source limit",
  "status": "fail",
  "duration": 20,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ …(8) ] to deeply equal [ …(8) ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1569,
        "column": 79,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1569
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -2,14 +2,10 @@
    {
      "key": "source",
      "revision": "new",
    },
    {
-     "key": "source",
-     "revision": "old",
-   },
-   {
      "key": "a",
      "revision": "a1",
    },
    {
      "key": "b",
@@ -28,7 +24,11 @@
      "revision": "e1",
    },
    {
      "key": "f",
      "revision": "f1",
+   },
+   {
+     "key": "overflow",
+     "revision": "o1",
    },
  ]
```

codeFrame (error 1):

```text
1567 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1568 |   await act(async () => {});
1569 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toEqual([
     |                                                                               ^
1570 |     { key: 'source', revision: 'new' },
1571 |     { key: 'source', revision: 'old' },
```


~~~~
FACT: E-030 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry distinguishes two revisions of the same key within its eight-source limit | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1569 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toEqual([ | CLOSED | DEFECT=no
FACT: E-030 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-031 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-031 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-031 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-031-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry distinguishes two revisions of the same key within its eight-source limit

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP ancestry distinguishes two revisions of the same key within its eight-source limit'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP ancestry distinguishes two revisions of the same key within its eight-source limit",
  "status": "fail",
  "duration": 20,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ …(8) ] to deeply equal [ …(8) ]",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1569,
        "column": 79,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1569
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,13 +1,13 @@
  [
    {
      "key": "source",
-     "revision": "new",
+     "revision": "old",
    },
    {
      "key": "source",
-     "revision": "old",
+     "revision": "new",
    },
    {
      "key": "a",
      "revision": "a1",
    },
```

codeFrame (error 1):

```text
1567 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1568 |   await act(async () => {});
1569 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toEqual([
     |                                                                               ^
1570 |     { key: 'source', revision: 'new' },
1571 |     { key: 'source', revision: 'old' },
```

### [F02] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry dedup retains the full restored snapshot for content-aware cleanup

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP ancestry dedup retains the full restored snapshot for content-aware cleanup'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP ancestry dedup retains the full restored snapshot for content-aware cleanup",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to deeply equal { key: 'duplicate-source', …(7) }",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1596,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1596
    }
  ]
}
```

diff:

```diff
- Expected:
{
  "acceptedSource": [
    {
      "key": "duplicate-source",
      "revision": "revision:duplicate-source",
    },
  ],
  "baseLanguage": "uk",
  "capturedAt": 1,
  "expectedUpdatedAt": "T1",
  "key": "duplicate-source",
  "language": "uk",
  "lines": [
    {
      "documentIds": [],
      "endMs": "1000",
      "id": "70000000-0000-4000-8000-000000000001",
      "startMs": "0",
      "text": "Changed with same revision",
    },
  ],
  "revision": "revision:duplicate-source",
}

+ Received:
null
```

codeFrame (error 1):

```text
1594 |   clickSave();
1595 |   await act(async () => {});
1596 |   expect(await props.history.store.read(source.key)).toEqual(replacement);
     |                                                      ^
1597 | });
1598 | 
```


~~~~
FACT: E-031 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry distinguishes two revisions of the same key within its eight-source limit; src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP ancestry dedup retains the full restored snapshot for content-aware cleanup | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1569 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toEqual([ / 1596 |   expect(await props.history.store.read(source.key)).toEqual(replacement); | CLOSED | DEFECT=no
FACT: E-031 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-036 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-036 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-036 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-036-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP restoring a descendant handles all concurrently offered ancestors

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP restoring a descendant handles all concurrently offered ancestors'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP restoring a descendant handles all concurrently offered ancestors",
  "status": "fail",
  "duration": 19,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <button …(9)></button> to be null",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1613,
        "column": 69,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1613
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
<button
  aria-disabled="false"
  class="ui-button ui-button--primary"
  data-focus-visible="true"
  data-focused="true"
  data-rac=""
  data-react-aria-pressable="true"
  id="react-aria-_r_2lp_"
  tabindex="0"
  type="button"
>
  Відновити зміни
</button>
```

codeFrame (error 1):

```text
1611 |   await act(async () => {});
1612 |   clickRestore();
1613 |   expect(screen.queryByRole('button', { name: 'Відновити зміни' })).toBeNull();
     |                                                                     ^
1614 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1615 |   await act(async () => {});
```


~~~~
FACT: E-036 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP restoring a descendant handles all concurrently offered ancestors | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1613 |   expect(screen.queryByRole('button', { name: 'Відновити зміни' })).toBeNull(); | CLOSED | DEFECT=no
FACT: E-036 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-048 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-048 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-048 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-048-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP loading ASR after a detached restore releases and reoffers the old source

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP loading ASR after a detached restore releases and reoffers the old source'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP loading ASR after a detached restore releases and reoffers the old source",
  "status": "fail",
  "duration": 60,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an accessible element with the role \"button\" and name \"Відновити зміни\"\n\nHere are the accessible roles:\n\n  region:\n\n  Name \"Транскрипт уроку\":\n  <section\n    aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n    class=\"transcript-editor\"\n  />\n\n  --------------------------------------------------\n  heading:\n\n  Name \"Транскрипт уроку\":\n  <h4\n    id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n  />\n\n  Name \"Чернетка розпізнавання готова\":\n  <h5\n    id=\"transcript-draft-offer-30000000-0000-4000-8000-000000000001\"\n  />\n\n  --------------------------------------------------\n  paragraph:\n\n  Name \"\":\n  <p />\n\n  --------------------------------------------------\n  button:\n\n  Name \"Завантажити чернетку\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--primary\"\n    data-focus-visible=\"true\"\n    data-focused=\"true\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2mq_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Не завантажувати\":\n  <button\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2ms_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Видалити репліку 1\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2mg_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Додати репліку\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--secondary\"\n    data-add-line=\"true\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2mi_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Зберегти транскрипт\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--primary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2mk_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  --------------------------------------------------\n  list:\n\n  Name \"\":\n  <ol\n    class=\"transcript-editor-lines\"\n  />\n\n  --------------------------------------------------\n  listitem:\n\n  Name \"\":\n  <li />\n\n  --------------------------------------------------\n  textbox:\n\n  Name \"Початок репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2m2_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2m1_\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"0\"\n  />\n\n  Name \"Кінець репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2m7_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2m6_\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"1000\"\n  />\n\n  Name \"Текст репліки 1\":\n  <textarea\n    aria-labelledby=\"react-aria-_r_2mc_\"\n    class=\"ui-input ui-textarea\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2mb_\"\n    tabindex=\"0\"\n    title=\"\"\n  />\n\n  --------------------------------------------------\n  group:\n\n  Name \"Матеріали репліки 1\":\n  <fieldset />\n\n  --------------------------------------------------\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <section\n      aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      class=\"transcript-editor\"\n    >\n      <h4\n        id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      >\n        Транскрипт уроку\n      </h4>\n      <div\n        aria-labelledby=\"transcript-draft-offer-30000000-0000-4000-8000-000000000001\"\n        class=\"transcript-editor-draft-offer\"\n      >\n        <h5\n          id=\"transcript-draft-offer-30000000-0000-4000-8000-000000000001\"\n        >\n          Чернетка розпізнавання готова\n        </h5>\n        <p>\n          Розпізнавання завершилось. Завантажити текст у редактор?\n        </p>\n        <button\n          aria-disabled=\"false\"\n          class=\"ui-button ui-button--primary\"\n          data-focus-visible=\"true\"\n          data-focused=\"true\"\n          data-rac=\"\"\n          data-react-aria-pressable=\"true\"\n          id=\"react-aria-_r_2mq_\"\n          tabindex=\"0\"\n          type=\"button\"\n        >\n          Завантажити чернетку\n        </button>\n        <button\n          class=\"ui-button ui-button--secondary\"\n          data-rac=\"\"\n          data-react-aria-pressable=\"true\"\n          id=\"react-aria-_r_2ms_\"\n          tabindex=\"0\"\n          type=\"button\"\n        >\n          Не завантажувати\n        </button>\n      </div>\n      <ol\n        class=\"transcript-editor-lines\"\n      >\n        <li>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2m1_\"\n              id=\"react-aria-_r_2m2_\"\n            >\n              Початок репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2m2_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2m1_\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"0\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2m6_\"\n              id=\"react-aria-_r_2m7_\"\n            >\n              Кінець репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2m7_\"\n      
~~~~
FACT: E-048 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP loading ASR after a detached restore releases and reoffers the old source | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1638 |   expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'ASR replacement'); / 1639 |   expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy(); | CLOSED | DEFECT=no
FACT: E-048 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-049 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-049 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-049 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-049-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP legacy recovery keeps the loaded ASR generation and current editor language

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP legacy recovery keeps the loaded ASR generation and current editor language'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP legacy recovery keeps the loaded ASR generation and current editor language",
  "status": "fail",
  "duration": 22,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { key: 'author:gap-editor', …(7) } to match object { language: 'uk', …(2) }\n(5 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1661,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1661
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,7 +1,7 @@
  {
-   "draftUpdatedAt": "D1",
+   "draftUpdatedAt": "D2",
    "language": "uk",
    "lines": [
      {
        "documentIds": [],
        "endMs": "1000",
```

codeFrame (error 1):

```text
1659 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1660 |   await act(async () => {});
1661 |   expect(await props.history.store.read(source.key)).toMatchObject({
     |                                                      ^
1662 |     language: 'uk',
1663 |     draftUpdatedAt: 'D1',
```


~~~~
FACT: E-049 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP legacy recovery keeps the loaded ASR generation and current editor language | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1661 |   expect(await props.history.store.read(source.key)).toMatchObject({ | CLOSED | DEFECT=no
FACT: E-049 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-052 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-052 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-052 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-052-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields",
  "status": "fail",
  "duration": 19,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "the given combination of arguments (undefined and string) is invalid for this assertion. You can use an array, a map, an object, a set, a string, or a weakset instead of a string",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 833,
        "column": 7,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 833
    }
  ]
}
```

codeFrame (error 1):

```text
831 |     expect(
832 |       screen.getByRole('textbox', { name }).getAttribute('aria-describedby')?.split(' '),
833 |     ).toContain(hint.id);
    |       ^
834 |   }
835 |   text.focus();
```


~~~~
FACT: E-052 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 831 |     expect( | CLOSED | DEFECT=no
FACT: E-052 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-059 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-059 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-059 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-059-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a rejected Save reports unavailable, retains recovery and resumes later captures

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a rejected Save reports unavailable, retains recovery and resumes later captures'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a rejected Save reports unavailable, retains recovery and resumes later captures",
  "status": "fail",
  "duration": 51,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an element with the text: Сервіс недоступний. Спробуйте ще раз.. This could be because the text is broken up by multiple elements. In this case, you can provide a function for your text matcher to make your matcher more flexible.\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <section\n      aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      class=\"transcript-editor\"\n    >\n      <h4\n        id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      >\n        Транскрипт уроку\n      </h4>\n      <ol\n        class=\"transcript-editor-lines\"\n      >\n        <li>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2ok_\"\n              id=\"react-aria-_r_2ol_\"\n            >\n              Початок репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2ol_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2ok_\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"0\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2op_\"\n              id=\"react-aria-_r_2oq_\"\n            >\n              Кінець репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2oq_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2op_\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"1000\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2ou_\"\n              id=\"react-aria-_r_2ov_\"\n            >\n              Текст репліки 1\n            </label>\n            <textarea\n              aria-labelledby=\"react-aria-_r_2ov_\"\n              class=\"ui-input ui-textarea\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2ou_\"\n              tabindex=\"0\"\n              title=\"\"\n            >\n              Before failure\n            </textarea>\n          </div>\n          <fieldset>\n            <legend>\n              Матеріали репліки 1\n            </legend>\n          </fieldset>\n          <button\n            aria-disabled=\"false\"\n            class=\"ui-button ui-button--secondary\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            id=\"react-aria-_r_2p3_\"\n            tabindex=\"0\"\n            type=\"button\"\n          >\n            Видалити репліку 1\n          </button>\n        </li>\n      </ol>\n      <button\n        aria-disabled=\"false\"\n        class=\"ui-button ui-button--secondary\"\n        data-add-line=\"true\"\n        data-rac=\"\"\n        data-react-aria-pressable=\"true\"\n        id=\"react-aria-_r_2p5_\"\n        tabindex=\"0\"\n        type=\"button\"\n      >\n        Додати репліку\n      </button>\n      <button\n        aria-disabled=\"false\"\n        class=\"ui-button ui-button--primary\"\n        data-rac=\"\"\n        data-react-aria-pressable=\"true\"\n        id=\"react-aria-_r_2p7_\"\n        tabindex=\"0\"\n        type=\"button\"\n      >\n        Зберегти транскрипт\n      </button>\n      <output>\n        Транскрипт збережено.\n      </output>\n    </section>\n  </div>\n</body>",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1692,
        "column": 10,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1692
    }
  ]
}
```

codeFrame (error 1):

```text
1690 |   clickSave();
1691 |   await act(async () => {});
1692 |   expect(screen.getByText('Сервіс недоступний. Спробуйте ще раз.')).toBeTruthy();
     |          ^
1693 |   expect(screen.queryByText('Транскрипт збережено.')).toBeNull();
1694 |   expect(await props.history.store.read(props.history.key)).toMatchObject({
```


~~~~
FACT: E-059 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a rejected Save reports unavailable, retains recovery and resumes later captures | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1692 |   expect(screen.getByText('Сервіс недоступний. Спробуйте ще раз.')).toBeTruthy(); / 1693 |   expect(screen.queryByText('Транскрипт збережено.')).toBeNull(); / 1694 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ | CLOSED | DEFECT=no
FACT: E-059 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-060 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-060 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-060 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-060-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a rejected Save reports unavailable, retains recovery and resumes later captures

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP a rejected Save reports unavailable, retains recovery and resumes later captures'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP a rejected Save reports unavailable, retains recovery and resumes later captures",
  "status": "fail",
  "duration": 25,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { key: 'author:gap-editor', …(6) } to match object { Object (lines) }\n(10 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1700,
        "column": 61,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1700
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
    "lines": [
      {
-       "text": "After failure",
+       "text": "Before failure",
      },
    ],
  }
```

codeFrame (error 1):

```text
1698 |   fireEvent.focusOut(text);
1699 |   await act(async () => {});
1700 |   expect(await props.history.store.read(props.history.key)).toMatchObject({
     |                                                             ^
1701 |     lines: [{ text: 'After failure' }],
1702 |   });
```

### [F02] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP successful detached Save removes ancestry from subsequent author captures

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP successful detached Save removes ancestry from subsequent author captures'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP successful detached Save removes ancestry from subsequent author captures",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to match object { Object (lines) }",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1782,
        "column": 61,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1782
    }
  ]
}
```

diff:

```diff
- Expected:
{
  "lines": [
    {
      "text": "New edit after successful Save",
    },
  ],
}

+ Received:
null
```

codeFrame (error 1):

```text
1780 |   fireEvent.focusOut(text);
1781 |   await act(async () => {});
1782 |   expect(await props.history.store.read(props.history.key)).toMatchObject({
     |                                                             ^
1783 |     lines: [{ text: 'New edit after successful Save' }],
1784 |   });
```


~~~~
FACT: E-060 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP a rejected Save reports unavailable, retains recovery and resumes later captures; src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP successful detached Save removes ancestry from subsequent author captures | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1700 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ / 1782 |   expect(await props.history.store.read(props.history.key)).toMatchObject({ | CLOSED | DEFECT=no
FACT: E-060 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-062 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-062 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-062 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-062-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP source cleanup rechecks newly mounted ownership after its deferred read

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP source cleanup rechecks newly mounted ownership after its deferred read'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP source cleanup rechecks newly mounted ownership after its deferred read",
  "status": "fail",
  "duration": 23,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ [ 'newly-mounted-source', …(1) ] ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1764,
        "column": 66,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1764
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   [
+     "newly-mounted-source",
+     "revision:newly-mounted-source",
+   ],
+ ]
```

codeFrame (error 1):

```text
1762 |   );
1763 |   await act(async () => pending.resolve(source));
1764 |   expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]);
     |                                                                  ^
1765 |   expect(await memory.read(source.key)).toEqual(source);
1766 | });
```


~~~~
FACT: E-062 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP source cleanup rechecks newly mounted ownership after its deferred read | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1764 |   expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]); / 1765 |   expect(await memory.read(source.key)).toEqual(source); | CLOSED | DEFECT=no
FACT: E-062 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-063 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-063 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-063 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-063-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP successful detached Save removes ancestry from subsequent author captures

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP successful detached Save removes ancestry from subsequent author captures'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP successful detached Save removes ancestry from subsequent author captures",
  "status": "fail",
  "duration": 21,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { key: 'saved-source', …(1) } ] to be undefined",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1785,
        "column": 79,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1785
    }
  ]
}
```

diff:

```diff
- Expected:
undefined

+ Received:
[
  {
    "key": "saved-source",
    "revision": "revision:saved-source",
  },
]
```

codeFrame (error 1):

```text
1783 |     lines: [{ text: 'New edit after successful Save' }],
1784 |   });
1785 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined();
     |                                                                               ^
1786 | });
1787 | 
```


~~~~
FACT: E-063 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP successful detached Save removes ancestry from subsequent author captures | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1785 |   expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined(); | CLOSED | DEFECT=no
FACT: E-063 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-066 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-066 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-066 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-066-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields",
  "status": "fail",
  "duration": 20,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "the given combination of arguments (undefined and string) is invalid for this assertion. You can use an array, a map, an object, a set, a string, or a weakset instead of a string",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 833,
        "column": 7,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 833
    }
  ]
}
```

codeFrame (error 1):

```text
831 |     expect(
832 |       screen.getByRole('textbox', { name }).getAttribute('aria-describedby')?.split(' '),
833 |     ).toContain(hint.id);
    |       ^
834 |   }
835 |   text.focus();
```


~~~~
FACT: E-066 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 831 |     expect( | CLOSED | DEFECT=no
FACT: E-066 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-072 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-072 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-072 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-072-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP detached Restore during pending Save cannot change text or ownership

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP detached Restore during pending Save cannot change text or ownership'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP detached Restore during pending Save cannot change text or ownership",
  "status": "fail",
  "duration": 26,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <textarea tabindex=\"0\" …(7)></textarea> to have property \"value\" with value 'Saving own'",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1538,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1538
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- Saving own
+ Detached
```

codeFrame (error 1):

```text
1536 |     );
1537 |     await act(async () => {});
1538 |     expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'Saving own');
     |                                                      ^
1539 |     expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy();
1540 |     expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]);
```


~~~~
FACT: E-072 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP detached Restore during pending Save cannot change text or ownership | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1538 |     expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'Saving own'); / 1539 |     expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy(); / 1540 |     expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]); | CLOSED | DEFECT=no
FACT: E-072 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-073 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-073 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-073 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-073-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP detached Discard during pending Save cannot change text or ownership

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP detached Discard during pending Save cannot change text or ownership'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP detached Discard during pending Save cannot change text or ownership",
  "status": "fail",
  "duration": 63,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an accessible element with the role \"button\" and name \"Відновити зміни\"\n\nHere are the accessible roles:\n\n  region:\n\n  Name \"Транскрипт уроку\":\n  <section\n    aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n    class=\"transcript-editor\"\n  />\n\n  --------------------------------------------------\n  heading:\n\n  Name \"Транскрипт уроку\":\n  <h4\n    id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n  />\n\n  --------------------------------------------------\n  list:\n\n  Name \"\":\n  <ol\n    class=\"transcript-editor-lines\"\n  />\n\n  --------------------------------------------------\n  listitem:\n\n  Name \"\":\n  <li />\n\n  --------------------------------------------------\n  textbox:\n\n  Name \"Початок репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2iq_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2ip_\"\n    readonly=\"\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"0\"\n  />\n\n  Name \"Кінець репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2iv_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2iu_\"\n    readonly=\"\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"1000\"\n  />\n\n  Name \"Текст репліки 1\":\n  <textarea\n    aria-labelledby=\"react-aria-_r_2j4_\"\n    class=\"ui-input ui-textarea\"\n    data-focus-visible=\"true\"\n    data-focused=\"true\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2j3_\"\n    readonly=\"\"\n    tabindex=\"0\"\n    title=\"\"\n  />\n\n  --------------------------------------------------\n  group:\n\n  Name \"Матеріали репліки 1\":\n  <fieldset />\n\n  --------------------------------------------------\n  button:\n\n  Name \"Видалити репліку 1\":\n  <button\n    aria-disabled=\"true\"\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2j8_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Додати репліку\":\n  <button\n    aria-disabled=\"true\"\n    class=\"ui-button ui-button--secondary\"\n    data-add-line=\"true\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2ja_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Зберегти транскрипт\":\n  <button\n    aria-disabled=\"true\"\n    class=\"ui-button ui-button--primary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2jc_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  --------------------------------------------------\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <section\n      aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      class=\"transcript-editor\"\n    >\n      <h4\n        id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      >\n        Транскрипт уроку\n      </h4>\n      <ol\n        class=\"transcript-editor-lines\"\n      >\n        <li>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n            data-readonly=\"true\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2ip_\"\n              id=\"react-aria-_r_2iq_\"\n            >\n              Початок репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2iq_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2ip_\"\n              readonly=\"\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"0\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n            data-readonly=\"true\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2iu_\"\n              id=\"react-aria-_r_2iv_\"\n            >\n              Кінець репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2iv_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2iu_\"\n              readonly=\"\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"1000\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n            data-readonly=\"true\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2j3_\"\n              id=\"react-aria-_r_2j4_\"\n            >\n              Текст репліки 1\n            </label>\n            <textarea\n              aria-labelledby=\"react-aria-_r_2j4_\"\n              class=\"ui-input ui-textarea\"\n              data-focus-visible=\"true\"\n              data-focused=\"true\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2j3_\"\n              readonly=\"\"\n              tabindex=\"0\"\n              title=\"\"\n            >\n              Saving own\n            </textarea>\n          </div>\n          <fieldset>\n            <legend>\n              Матеріали репліки 1\n            </legend>\n          </fieldset>\n          <button\n            aria-disabled=\"true\"\n            class=\"ui-button ui-button--secondary\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            id=\"react-aria-_r_2j8_\"\n            tabindex=\"0\"\n            type=\"button\"\n          >\n            Видалити репліку 1\n          </button>\n        </li>\n      </ol>\n      <button\n        aria-disabled=\"true\"\n        class=\"ui-button ui-button--secondary\"\n        data-add-line=\"true
~~~~
FACT: E-073 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP detached Discard during pending Save cannot change text or ownership | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1538 |     expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'Saving own'); / 1539 |     expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy(); / 1540 |     expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]); / 1541 |     expect(await props.history.store.read(source.key)).toEqual(source); | CLOSED | DEFECT=no
FACT: E-073 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: E-074 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: E-074 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: E-074 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-E-074-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP Discard of a refreshed own offer releases an earlier detached restoration

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP Discard of a refreshed own offer releases an earlier detached restoration'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP Discard of a refreshed own offer releases an earlier detached restoration",
  "status": "fail",
  "duration": 57,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an accessible element with the role \"button\" and name \"Відновити зміни\"\n\nHere are the accessible roles:\n\n  region:\n\n  Name \"Транскрипт уроку\":\n  <section\n    aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n    class=\"transcript-editor\"\n  />\n\n  --------------------------------------------------\n  heading:\n\n  Name \"Транскрипт уроку\":\n  <h4\n    id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n  />\n\n  --------------------------------------------------\n  list:\n\n  Name \"\":\n  <ol\n    class=\"transcript-editor-lines\"\n  />\n\n  --------------------------------------------------\n  listitem:\n\n  Name \"\":\n  <li />\n\n  --------------------------------------------------\n  textbox:\n\n  Name \"Початок репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2rl_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2rk_\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"0\"\n  />\n\n  Name \"Кінець репліки 1, мс\":\n  <input\n    aria-labelledby=\"react-aria-_r_2rq_\"\n    class=\"ui-input\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2rp_\"\n    tabindex=\"0\"\n    title=\"\"\n    type=\"text\"\n    value=\"1000\"\n  />\n\n  Name \"Текст репліки 1\":\n  <textarea\n    aria-labelledby=\"react-aria-_r_2rv_\"\n    class=\"ui-input ui-textarea\"\n    data-rac=\"\"\n    id=\"react-aria-_r_2ru_\"\n    tabindex=\"0\"\n    title=\"\"\n  />\n\n  --------------------------------------------------\n  group:\n\n  Name \"Матеріали репліки 1\":\n  <fieldset />\n\n  --------------------------------------------------\n  button:\n\n  Name \"Видалити репліку 1\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2s3_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Додати репліку\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--secondary\"\n    data-add-line=\"true\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2s5_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Зберегти транскрипт\":\n  <button\n    aria-disabled=\"false\"\n    class=\"ui-button ui-button--primary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_2s7_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  --------------------------------------------------\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <section\n      aria-labelledby=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      class=\"transcript-editor\"\n    >\n      <h4\n        id=\"transcript-editor-30000000-0000-4000-8000-000000000001\"\n      >\n        Транскрипт уроку\n      </h4>\n      <ol\n        class=\"transcript-editor-lines\"\n      >\n        <li>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2rk_\"\n              id=\"react-aria-_r_2rl_\"\n            >\n              Початок репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2rl_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2rk_\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"0\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2rp_\"\n              id=\"react-aria-_r_2rq_\"\n            >\n              Кінець репліки 1, мс\n            </label>\n            <input\n              aria-labelledby=\"react-aria-_r_2rq_\"\n              class=\"ui-input\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2rp_\"\n              tabindex=\"0\"\n              title=\"\"\n              type=\"text\"\n              value=\"1000\"\n            />\n          </div>\n          <div\n            class=\"ui-field\"\n            data-rac=\"\"\n          >\n            <label\n              class=\"react-aria-Label\"\n              for=\"react-aria-_r_2ru_\"\n              id=\"react-aria-_r_2rv_\"\n            >\n              Текст репліки 1\n            </label>\n            <textarea\n              aria-labelledby=\"react-aria-_r_2rv_\"\n              class=\"ui-input ui-textarea\"\n              data-rac=\"\"\n              id=\"react-aria-_r_2ru_\"\n              tabindex=\"0\"\n              title=\"\"\n            >\n              Перша\n            </textarea>\n          </div>\n          <fieldset>\n            <legend>\n              Матеріали репліки 1\n            </legend>\n          </fieldset>\n          <button\n            aria-disabled=\"false\"\n            class=\"ui-button ui-button--secondary\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            id=\"react-aria-_r_2s3_\"\n            tabindex=\"0\"\n            type=\"button\"\n          >\n            Видалити репліку 1\n          </button>\n        </li>\n      </ol>\n      <button\n        aria-disabled=\"false\"\n        class=\"ui-button ui-button--secondary\"\n        data-add-line=\"true\"\n        data-rac=\"\"\n        data-react-aria-pressable=\"true\"\n        id=\"react-aria-_r_2s5_\"\n        tabindex=\"0\"\n        type=\"button\"\n      >\n        Додати репліку\n      </button>\n      <button\n        aria-disabled=\"false\"\n        class=\"ui-button ui-button--primary\"\n        data-rac=\"\"\n        data-react-aria-pressable=\"true\"\n        i
~~~~
FACT: E-074 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP Discard of a refreshed own offer releases an earlier detached restoration | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1806 |   expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy(); | CLOSED | DEFECT=no
FACT: E-074 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: X-005 test SHA-256=79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: X-005 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`. Production SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
FACT: X-005 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-005-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: MS3b fix4 atomic IndexedDB cleanup error uses one readwrite transaction and waits for commit

repro:

```bash
npx rstest 'src/modules/transcript/adapters/transcript-history-indexed-database.test.ts' --testNamePattern 'MS3b fix4 atomic IndexedDB cleanup error uses one readwrite transaction and waits for commit'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
  "project": "adapter-node",
  "fullName": "MS3b fix4 atomic IndexedDB cleanup error uses one readwrite transaction and waits for commit",
  "status": "fail",
  "duration": 6,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise resolved \"false\" instead of rejecting",
      "topFrame": {
        "file": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
        "line": 121,
        "column": 20,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/transcript-history-indexed-database.test.ts",
      "line": 121
    }
  ]
}
```

diff:

```diff
- Expected:
Error {
  "message": "rejected promise",
}

+ Received:
false
```

codeFrame (error 1):

```text
119 |     if (outcome === 'error') {
120 |       transaction.dispatchEvent(new Event('error'));
121 |       await expect(observed).rejects.toThrow('transaction failed');
    |                    ^
122 |     } else if (outcome === 'abort') {
123 |       transaction.dispatchEvent(new Event('abort'));
```


~~~~
FACT: X-005 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts :: MS3b fix4 atomic IndexedDB cleanup error uses one readwrite transaction and waits for commit | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 121 |       await expect(observed).rejects.toThrow('transaction failed'); | CLOSED | DEFECT=no
FACT: X-005 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` exit=0.
FACT: X-007 test SHA-256=ab2dd6b76f69ca87c4b491ad8b4294502b188eda4fe9b221c797165d3aeb9023; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: X-007 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: X-007 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-007-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for null envelope

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'draft API repository > rejects the whole list for null envelope'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "draft API repository > rejects the whole list for null envelope",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise rejected \"TypeError: Cannot convert undefined or nu…\" instead of resolving",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 96,
        "column": 56,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 96
    }
  ]
}
```

codeFrame (error 1):

```text
94 |   ])('rejects the whole list for $name', async ({ response }) => {
95 |     server.use(http.get(api, () => HttpResponse.json(response)));
96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid);
   |                                                        ^
97 |   });
98 | 
```


~~~~
FACT: X-007 | src/modules/transcript/adapters/draft-repository.test.ts :: draft API repository > rejects the whole list for null envelope | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 96 |     await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid); | CLOSED | DEFECT=no
FACT: X-007 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
FACT: X-009 test SHA-256=ab2dd6b76f69ca87c4b491ad8b4294502b188eda4fe9b221c797165d3aeb9023; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: X-009 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/adapters/draft-repository.test.ts`. Production SHA-256 0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120.
FACT: X-009 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-009-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/adapters/draft-repository.test.ts :: GAP inventory cancellation reports request_cancelled after request starts

repro:

```bash
npx rstest 'src/modules/transcript/adapters/draft-repository.test.ts' --testNamePattern 'GAP inventory cancellation reports request_cancelled after request starts'
```

details:

```json
{
  "testPath": "src/modules/transcript/adapters/draft-repository.test.ts",
  "project": "adapter-node",
  "fullName": "GAP inventory cancellation reports request_cancelled after request starts",
  "status": "fail",
  "duration": 3,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: false, …(1) } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/transcript/adapters/draft-repository.test.ts",
        "line": 216,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/adapters/draft-repository.test.ts",
      "line": 216
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
    "error": {
-     "code": "request_cancelled",
+     "code": "invalid_response",
    },
    "ok": false,
  }
```

codeFrame (error 1):

```text
214 |   controller.abort();
215 |   releaseResponse();
216 |   expect(await pending).toEqual({ ok: false, error: { code: 'request_cancelled' } });
    |                         ^
217 | });
218 | 
```


~~~~
FACT: X-009 | src/modules/transcript/adapters/draft-repository.test.ts :: GAP inventory cancellation reports request_cancelled after request starts | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 216 |   expect(await pending).toEqual({ ok: false, error: { code: 'request_cancelled' } }); | CLOSED | DEFECT=no
FACT: X-009 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/adapters/draft-repository.ts` exit=0.
FACT: X-012 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: X-012 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: X-012 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-012-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read reject cannot replace the current visit offer

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP obsolete initial read reject cannot replace the current visit offer'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP obsolete initial read reject cannot replace the current visit offer",
  "status": "fail",
  "duration": 14,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'none' } to match object { snapshot: { …(7) } }\n(1 matching property omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 313,
        "column": 36,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 313
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "snapshot": {
-     "baseLanguage": "en",
-     "capturedAt": 1,
-     "expectedUpdatedAt": "T1",
-     "key": "gap-obsolete-reject:b",
-     "language": "uk",
-     "lines": [
-       {
-         "documentIds": [],
-         "endMs": "1000",
-         "id": "line",
-         "startMs": "0",
-         "text": "Recovered",
-       },
-     ],
-     "revision": "revision:gap-obsolete-reject:b:Recovered",
-   },
+   "kind": "none",
  }
```

codeFrame (error 1):

```text
311 |     });
312 |     expect(read).toHaveBeenCalledTimes(2);
313 |     expect(f.result.current.offer).toMatchObject({ snapshot: b });
    |                                    ^
314 |     expect(f.result.current.isLoaded).toBe(true);
315 |   });
```


~~~~
FACT: X-012 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP obsolete initial read reject cannot replace the current visit offer | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 312 |     expect(read).toHaveBeenCalledTimes(2); / 313 |     expect(f.result.current.offer).toMatchObject({ snapshot: b }); / 314 |     expect(f.result.current.isLoaded).toBe(true); | CLOSED | DEFECT=no
FACT: X-012 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: X-013 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: X-013 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: X-013 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-013-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended noteChange does not arm capture when resumed or unmounted

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP suspended noteChange does not arm capture when resumed or unmounted'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP suspended noteChange does not arm capture when resumed or unmounted",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"write\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st write call:\n\n    Array [\n      Object {\n        \"baseLanguage\": \"en\",\n        \"capturedAt\": 1794929411194,\n        \"expectedUpdatedAt\": \"T1\",\n        \"key\": \"user:lesson:track:a\",\n        \"language\": \"uk\",\n        \"lines\": Array [\n          Object {\n            \"documentIds\": Array [],\n            \"endMs\": \"1000\",\n            \"id\": \"line\",\n            \"startMs\": \"0\",\n            \"text\": \"suspended edit\",\n          },\n        ],\n        \"revision\": \"541a2c62-846b-42e3-bd2e-b5ed3f6d4b42\",\n      },\n    ]\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 467,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 467
    }
  ]
}
```

codeFrame (error 1):

```text
465 |     act(() => rstest.advanceTimersByTime(idleCaptureMs));
466 |     f.unmount();
467 |     expect(f.write).not.toHaveBeenCalled();
    |                         ^
468 |   });
469 | }
```


~~~~
FACT: X-013 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP suspended noteChange does not arm capture when resumed or unmounted | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 467 |     expect(f.write).not.toHaveBeenCalled(); | CLOSED | DEFECT=no
FACT: X-013 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: X-014 test SHA-256=ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: X-014 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx`. Production SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
FACT: X-014 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-014-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP a vanished mismatch closes the stale own recovery prompt

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-transcript-recovery.test.tsx' --testNamePattern 'GAP a vanished mismatch closes the stale own recovery prompt'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP a vanished mismatch closes the stale own recovery prompt",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { kind: 'offer', snapshot: { …(7) } } to deeply equal { kind: 'none' }",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
        "line": 453,
        "column": 34,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-transcript-recovery.test.tsx",
      "line": 453
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "kind": "none",
+   "kind": "offer",
+   "snapshot": {
+     "baseLanguage": "en",
+     "capturedAt": 1,
+     "expectedUpdatedAt": "T1",
+     "key": "gap-vanished",
+     "language": "uk",
+     "lines": [
+       {
+         "documentIds": [],
+         "endMs": "1000",
+         "id": "line",
+         "startMs": "0",
+         "text": "Recovered",
+       },
+     ],
+     "revision": "revision:gap-vanished:Recovered",
+   },
  }
```

codeFrame (error 1):

```text
451 |   act(() => f.result.current.captureNow([line]));
452 |   await act(async () => {});
453 |   expect(f.result.current.offer).toEqual({ kind: 'none' });
    |                                  ^
454 | });
455 | 
```


~~~~
FACT: X-014 | src/modules/transcript/ui/use-transcript-recovery.test.tsx :: GAP a vanished mismatch closes the stale own recovery prompt | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 453 |   expect(f.result.current.offer).toEqual({ kind: 'none' }); | CLOSED | DEFECT=no
FACT: X-014 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-transcript-recovery.ts` exit=0.
FACT: X-017 test SHA-256=9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: X-017 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/TranscriptEditor.test.tsx`. Production SHA-256 1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2.
FACT: X-017 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-017-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP legacy recovery keeps the loaded ASR generation and current editor language

repro:

```bash
npx rstest 'src/modules/transcript/ui/TranscriptEditor.test.tsx' --testNamePattern 'GAP legacy recovery keeps the loaded ASR generation and current editor language'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
  "project": "dom",
  "fullName": "GAP legacy recovery keeps the loaded ASR generation and current editor language",
  "status": "fail",
  "duration": 21,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { key: 'author:gap-editor', …(6) } to match object { language: 'uk', …(2) }\n(5 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
        "line": 1661,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/TranscriptEditor.test.tsx",
      "line": 1661
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,8 +1,7 @@
  {
    "draftUpdatedAt": "D1",
-   "language": "uk",
    "lines": [
      {
        "documentIds": [],
        "endMs": "1000",
        "id": "70000000-0000-4000-8000-000000000001",
```

codeFrame (error 1):

```text
1659 |   fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
1660 |   await act(async () => {});
1661 |   expect(await props.history.store.read(source.key)).toMatchObject({
     |                                                      ^
1662 |     language: 'uk',
1663 |     draftUpdatedAt: 'D1',
```


~~~~
FACT: X-017 | src/modules/transcript/ui/TranscriptEditor.test.tsx :: GAP legacy recovery keeps the loaded ASR generation and current editor language | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 1661 |   expect(await props.history.store.read(source.key)).toMatchObject({ | CLOSED | DEFECT=no
FACT: X-017 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/TranscriptEditor.tsx` exit=0.
FACT: X-020 test SHA-256=75fcfb69be6bcbfa10743f8419a65588eb37fe9ca8b3d3d60fc8deab85eca467; final baseline is fe-a-suite.txt if suffix=-final.
NEXT: X-020 exact prior mutation; command `cd frontend && npx rstest run src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx`. Production SHA-256 94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a.
FACT: X-020 mutation exit=1; raw `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-X-020-final-mutation.txt`.
~~~~text
## Failures

### [F01] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "",
  "status": "fail",
  "duration": 26,
  "retryCount": 0,
  "errors": [
    {
      "type": "unhandledRejection",
      "message": "inventory refusal",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.ts",
        "line": 53,
        "column": 15,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.ts",
      "line": 53
    }
  ]
}
```

codeFrame (error 1):

```text
51 |           });
52 |       } catch {
53 |         throw new Error('inventory refusal');
   |               ^
54 |       }
55 |     })();
```

console:

```text
[stderr] error: unhandledRejection: inventory refusal
    at /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts:53:15
```

### [F02] src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP inventory refusal remains best effort without an escaped rejection

repro:

```bash
npx rstest 'src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx' --testNamePattern 'GAP inventory refusal remains best effort without an escaped rejection'
```

details:

```json
{
  "testPath": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
  "project": "dom",
  "fullName": "GAP inventory refusal remains best effort without an escaped rejection",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ …(1) ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
        "line": 129,
        "column": 21,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx",
      "line": 129
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- []
+ [
+   unhandledRejection {
+     "message": "inventory refusal",
+     "name": "unhandledRejection",
+   },
+ ]
```

codeFrame (error 1):

```text
127 |     expect(keys).toHaveBeenCalledTimes(1);
128 |     expect(f.result.current).toEqual([]);
129 |     expect(escaped).toEqual([]);
    |                     ^
130 |   } finally {
131 |     process.off('unhandledRejection', onRejection);
```

console:

```text
[stderr] error: unhandledRejection: inventory refusal
    at /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts:53:15
```


~~~~
FACT: X-020 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx; src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx :: GAP inventory refusal remains best effort without an escaped rejection | baseline PASS (final whole suite fe-a-suite.txt: 1257/1257 if suffix=-final; otherwise recorded batch baseline) | 127 |     expect(keys).toHaveBeenCalledTimes(1); / 128 |     expect(f.result.current).toEqual([]); / 129 |     expect(escaped).toEqual([]); | CLOSED | DEFECT=no
FACT: X-020 restored original production bytes; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts` exit=0.
FACT: FINAL EXACT-SNAPSHOT AUDIT: 74/74 final mutations have named test failures, not just aggregate/file errors; baseline full suite=1257/1257 (108 files), 0 skipped. CLOSED=74, DEFECT=0, remaining=0. Every mutation target git diff --quiet exit=0. Only *.test.ts(x) changed, including two new test files. No production, configuration or dependency file changed.
FACT: FINAL PER-ROW RECORD (supersedes provisional records; baseline for every row = `cd frontend && npx rstest run`, exit 0, raw fe-a-suite.txt). Assertion below quoted verbatim from the test line identified by the named Rstest failure frame; full message/raw command/exit are in the row logs above and its -final raw artifact.

| id | test file::test name | baseline | mutated FAIL line verbatim | CLOSED | DEFECT |
|---|---|---|---|---|---|
| A-007 | src/modules/transcript/adapters/draft-repository.test.ts::draft API repository > rejects the whole list for obsolete lessonId | PASS | `await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid);` | CLOSED | no |
| A-011 | src/modules/transcript/adapters/draft-repository.test.ts::draft API repository > rejects the whole list for fractional generation | PASS | `await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid);` | CLOSED | no |
| A-012 | src/modules/transcript/adapters/draft-repository.test.ts::draft API repository > rejects the whole list for zero generation | PASS | `await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid);` | CLOSED | no |
| A-028 | src/modules/transcript/adapters/draft-repository.test.ts::GAP inventory cancellation reports request_cancelled after request starts | PASS | `expect(await pending).toEqual({ ok: false, error: { code: 'request_cancelled' } });` | CLOSED | no |
| I-003 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB rejects missing key without offering corrupt recovery | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();` | CLOSED | no |
| I-004 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB rejects missing capturedAt without offering corrupt recovery | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();` | CLOSED | no |
| I-007 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB rejects numeric language without offering corrupt recovery | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();` | CLOSED | no |
| I-009 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB rejects numeric expectedUpdatedAt without offering corrupt recovery | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();` | CLOSED | no |
| I-010 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB rejects numeric draftUpdatedAt without offering corrupt recovery | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();` | CLOSED | no |
| I-011 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB rejects numeric line id without offering corrupt recovery | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();` | CLOSED | no |
| I-012 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB rejects numeric startMs without offering corrupt recovery | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();` | CLOSED | no |
| I-013 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB rejects numeric endMs without offering corrupt recovery | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();` | CLOSED | no |
| I-014 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB rejects numeric text without offering corrupt recovery | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();` | CLOSED | no |
| I-015 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB rejects numeric documentIds without offering corrupt recovery | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();` | CLOSED | no |
| I-024 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::GAP IndexedDB preserves the ASR draft token alongside valid recovery text | PASS | `expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toEqual(value);` | CLOSED | no |
| I-025 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::MS3b fix4 atomic IndexedDB cleanup wrong-key uses one readwrite transaction and waits for commit | PASS | `else expect(remove).not.toHaveBeenCalled();` | CLOSED | no |
| R-018 | src/modules/transcript/application/recovery.test.ts::GAP accepted cleanup preserves a same-revision snapshot differing in {"key":"another-key"} | PASS | `expect(clearIfVersion).not.toHaveBeenCalled();` | CLOSED | no |
| R-021 | src/modules/transcript/application/recovery.test.ts::GAP accepted cleanup preserves a same-revision snapshot differing in {"language":"es"} | PASS | `expect(clearIfVersion).not.toHaveBeenCalled();` | CLOSED | no |
| R-022 | src/modules/transcript/application/recovery.test.ts::GAP accepted cleanup preserves a same-revision snapshot differing in {"baseLanguage":"fr"} | PASS | `expect(clearIfVersion).not.toHaveBeenCalled();` | CLOSED | no |
| R-023 | src/modules/transcript/application/recovery.test.ts::GAP accepted cleanup preserves a same-revision snapshot differing in {"expectedUpdatedAt":"T2"} | PASS | `expect(clearIfVersion).not.toHaveBeenCalled();` | CLOSED | no |
| R-024 | src/modules/transcript/application/recovery.test.ts::GAP accepted cleanup preserves a same-revision snapshot differing in {"draftUpdatedAt":"D2"} | PASS | `expect(clearIfVersion).not.toHaveBeenCalled();` | CLOSED | no |
| H-001 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP switching a loaded visit hides its offer until the new store read resolves | PASS | `expect(f.result.current.isLoaded).toBe(false);` | CLOSED | no |
| H-002 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP obsolete read completion cannot authorize retirement in an unread visit | PASS | `expect(clear).not.toHaveBeenCalled();` | CLOSED | no |
| H-003 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP obsolete initial read resolve cannot replace the current visit offer | PASS | `expect(f.result.current.offer).toMatchObject({ snapshot: b });` | CLOSED | no |
| H-004 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP obsolete initial read resolve cannot replace the current visit offer | PASS | `expect(f.result.current.offer).toMatchObject({ snapshot: b });` | CLOSED | no |
| H-006 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP finalization waits for its write commit before conditional retirement | PASS | `expect(clear).not.toHaveBeenCalled();` | CLOSED | no |
| H-008 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::MS3b fix6 ownership Save / newer | PASS | `expect(f.result.current.offer).toEqual({ kind: 'none' });` | CLOSED | no |
| H-009 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP mismatch refresh from an obsolete visit cannot replace the current offer | PASS | `expect(f.result.current.offer).toMatchObject({ snapshot: b });` | CLOSED | no |
| H-010 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP repeated Discard retires the refreshed revision rather than its predecessor | PASS | `expect(clear).toHaveBeenCalledExactlyOnceWith(c.key, 'C');` | CLOSED | no |
| H-013 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP suspended captureNow does not arm capture when resumed or unmounted | PASS | `expect(f.write).not.toHaveBeenCalled();` | CLOSED | no |
| H-015 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP clean capture cannot retire an own write before the initial read finishes | PASS | `expect(clear).not.toHaveBeenCalled();` | CLOSED | no |
| H-025 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP obsolete read completion cannot authorize retirement in an unread visit | PASS | `expect(clear).not.toHaveBeenCalled();` | CLOSED | no |
| H-026 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP finalization without an own write retains the accepted recovery source | PASS | `expect(clear).not.toHaveBeenCalled();` | CLOSED | no |
| H-028 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP suspended captureNow does not arm capture when resumed or unmounted | PASS | `expect(f.write).not.toHaveBeenCalled();` | CLOSED | no |
| H-029 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP switching a loaded visit hides its offer until the new store read resolves | PASS | `expect(f.result.current.offer).toEqual({ kind: 'none' });` | CLOSED | no |
| U-001 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx::GAP disabled detached inventory performs no storage IO | PASS | `expect(keys).not.toHaveBeenCalled();` | CLOSED | no |
| U-003 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx::GAP detached inventory ignores a read result belonging to a different listed key | PASS | `expect(f.result.current).toEqual([]);` | CLOSED | no |
| U-004 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx::GAP obsolete detached inventory cannot erase the completed current inventory | PASS | `expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);` | CLOSED | no |
| U-005 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx::GAP detached inventory hides stale same-lesson offers after store changes | PASS | `expect(f.result.current).toEqual([]);` | CLOSED | no |
| U-009 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx::GAP one refused detached read does not suppress a different valid recovery | PASS | `expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);` | CLOSED | no |
| L-001 | src/modules/transcript/ui/LessonDraftField.test.tsx::GAP a missing shown track reports not found after Regenerate | PASS | `expect((await screen.findByRole('alert')).textContent).toBe(` | CLOSED | no |
| F-002 | src/modules/transcript/ui/LessonTranscriptField.test.tsx::GAP first transcript load then renamed-language load retains the dirty editor instance | PASS | `expect(screen.getByLabelText('Текст репліки 1')).toBe(text);` | CLOSED | no |
| F-009 | src/modules/transcript/ui/LessonTranscriptField.test.tsx::GAP fixed tab language keeps the returned transcript token attached to its real base language | PASS | `expect(await store.read('author:lesson:track:one')).toMatchObject({` | CLOSED | no |
| E-001 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP a source acquired by a mounted tab is removed from captured cleanup ancestry | PASS | `expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined();` | CLOSED | no |
| E-002 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP source cleanup rechecks newly mounted ownership after its deferred read | PASS | `expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]);` | CLOSED | no |
| E-003 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP a completed Save lets the next canonical server token become the next edit base | PASS | `expect(props.actions.save).toHaveBeenLastCalledWith('uk', 'T2', expect.any(Array), 'uk');` | CLOSED | no |
| E-005 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP a clean editor adopts a server change to only base language | PASS | `expect(props.actions.save).toHaveBeenCalledWith(` | CLOSED | no |
| E-006 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP a clean editor adopts a server change to only token | PASS | `expect(props.actions.save).toHaveBeenCalledWith(` | CLOSED | no |
| E-007 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP a clean editor adopts a server change to only segments | PASS | `expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'New server');` | CLOSED | no |
| E-008 | src/modules/transcript/ui/TranscriptEditor.test.tsx::MS3b fix 5 a clean track rename followed by editing is not a dirty track rename | PASS | `expect(screen.queryByText('Мова доріжки тепер en; рядки буде збережено як en.')).toBeNull();` | CLOSED | no |
| E-013 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP dirty edits compare against their retained base after a server prop refresh | PASS | `expect(await props.history.store.read(props.history.key)).toMatchObject({` | CLOSED | no |
| E-017 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP a pending Save locks fields, documents, line controls and repeat Save | PASS | `expect(field).toHaveProperty('readOnly', true);` | CLOSED | no |
| E-030 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP ancestry distinguishes two revisions of the same key within its eight-source limit | PASS | `expect((await props.history.store.read(props.history.key))?.acceptedSource).toEqual([` | CLOSED | no |
| E-031 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP ancestry distinguishes two revisions of the same key within its eight-source limit | PASS | `expect((await props.history.store.read(props.history.key))?.acceptedSource).toEqual([` | CLOSED | no |
| E-036 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP restoring a descendant handles all concurrently offered ancestors | PASS | `expect(screen.queryByRole('button', { name: 'Відновити зміни' })).toBeNull();` | CLOSED | no |
| E-048 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP loading ASR after a detached restore releases and reoffers the old source | PASS | `expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy();` | CLOSED | no |
| E-049 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP legacy recovery keeps the loaded ASR generation and current editor language | PASS | `expect(await props.history.store.read(source.key)).toMatchObject({` | CLOSED | no |
| E-052 | src/modules/transcript/ui/TranscriptEditor.test.tsx::MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields | PASS | `).toContain(hint.id);` | CLOSED | no |
| E-059 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP a rejected Save reports unavailable, retains recovery and resumes later captures | PASS | `expect(screen.getByText('Сервіс недоступний. Спробуйте ще раз.')).toBeTruthy();` | CLOSED | no |
| E-060 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP a rejected Save reports unavailable, retains recovery and resumes later captures | PASS | `expect(await props.history.store.read(props.history.key)).toMatchObject({` | CLOSED | no |
| E-062 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP source cleanup rechecks newly mounted ownership after its deferred read | PASS | `expect(clear.mock.calls.filter(([key]) => key === source.key)).toEqual([]);` | CLOSED | no |
| E-063 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP successful detached Save removes ancestry from subsequent author captures | PASS | `expect((await props.history.store.read(props.history.key))?.acceptedSource).toBeUndefined();` | CLOSED | no |
| E-066 | src/modules/transcript/ui/TranscriptEditor.test.tsx::MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields | PASS | `).toContain(hint.id);` | CLOSED | no |
| E-072 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP detached Restore during pending Save cannot change text or ownership | PASS | `expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'Saving own');` | CLOSED | no |
| E-073 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP detached Discard during pending Save cannot change text or ownership | PASS | `expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy();` | CLOSED | no |
| E-074 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP Discard of a refreshed own offer releases an earlier detached restoration | PASS | `expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy();` | CLOSED | no |
| X-005 | src/modules/transcript/adapters/transcript-history-indexed-database.test.ts::MS3b fix4 atomic IndexedDB cleanup error uses one readwrite transaction and waits for commit | PASS | `await expect(observed).rejects.toThrow('transaction failed');` | CLOSED | no |
| X-007 | src/modules/transcript/adapters/draft-repository.test.ts::draft API repository > rejects the whole list for null envelope | PASS | `await expect(createApiDraftRepository().listDrafts(lessonId)).resolves.toEqual(invalid);` | CLOSED | no |
| X-009 | src/modules/transcript/adapters/draft-repository.test.ts::GAP inventory cancellation reports request_cancelled after request starts | PASS | `expect(await pending).toEqual({ ok: false, error: { code: 'request_cancelled' } });` | CLOSED | no |
| X-012 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP obsolete initial read reject cannot replace the current visit offer | PASS | `expect(f.result.current.offer).toMatchObject({ snapshot: b });` | CLOSED | no |
| X-013 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP suspended noteChange does not arm capture when resumed or unmounted | PASS | `expect(f.write).not.toHaveBeenCalled();` | CLOSED | no |
| X-014 | src/modules/transcript/ui/use-transcript-recovery.test.tsx::GAP a vanished mismatch closes the stale own recovery prompt | PASS | `expect(f.result.current.offer).toEqual({ kind: 'none' });` | CLOSED | no |
| X-017 | src/modules/transcript/ui/TranscriptEditor.test.tsx::GAP legacy recovery keeps the loaded ASR generation and current editor language | PASS | `expect(await props.history.store.read(source.key)).toMatchObject({` | CLOSED | no |
| X-020 | src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx::GAP inventory refusal remains best effort without an escaped rejection | PASS | `expect(escaped).toEqual([]);` | CLOSED | no |

FACT: `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --name-only` exit 0:
~~~~text
frontend/src/modules/transcript/adapters/draft-repository.test.ts
frontend/src/modules/transcript/adapters/transcript-history-indexed-database.test.ts
frontend/src/modules/transcript/application/recovery.test.ts
frontend/src/modules/transcript/ui/LessonDraftField.test.tsx
frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx
frontend/src/modules/transcript/ui/use-transcript-recovery.test.tsx
~~~~
FACT: `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --stat` exit 0 (tracked files only):
~~~~text
 .../transcript/adapters/draft-repository.test.ts   |  39 +-
 .../transcript-history-indexed-database.test.ts    |  43 +-
 .../transcript/application/recovery.test.ts        |  23 +-
 .../transcript/ui/LessonDraftField.test.tsx        |   8 +
 .../transcript/ui/TranscriptEditor.test.tsx        | 456 +++++++++++++++++++++
 .../transcript/ui/use-transcript-recovery.test.tsx | 223 ++++++++++
 6 files changed, 786 insertions(+), 6 deletions(-)
~~~~
FACT: SHA-256 ab2dd6b76f69ca87c4b491ad8b4294502b188eda4fe9b221c797165d3aeb9023  frontend/src/modules/transcript/adapters/draft-repository.test.ts
FACT: SHA-256 79277398921c346576cb840441c391199ed68f9ba25ef28f8dbdee13b21a3fa3  frontend/src/modules/transcript/adapters/transcript-history-indexed-database.test.ts
FACT: SHA-256 f06d85b1177079158747e52d9bbb734748189e888fe4cb66ad8ab80dc6a6d3b3  frontend/src/modules/transcript/application/recovery.test.ts
FACT: SHA-256 87df0aab9beaeadc8f3b056ec3959ca8c78f4f084ea5bd2906c4238695f73467  frontend/src/modules/transcript/ui/LessonDraftField.test.tsx
FACT: SHA-256 9ccc0562aedc3f9e1713ee3582b4c643d05868d5af183d7d2082742f46c15b52  frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx
FACT: SHA-256 ae202f22ab39b6ea6832c8ee9f644755370cf3a29f87821b89656896b83cc9f5  frontend/src/modules/transcript/ui/use-transcript-recovery.test.tsx
FACT: SHA-256 a1ea6e3097b9f8a9f65ff60bac22b5a71e4d0b9375683cd8ead24d177122a7d0  frontend/src/modules/transcript/ui/LessonTranscriptField.test.tsx (NEW/untracked, 72 lines; excluded by git diff --stat)
FACT: SHA-256 75fcfb69be6bcbfa10743f8419a65588eb37fe9ca8b3d3d60fc8deab85eca467  frontend/src/modules/transcript/ui/use-detached-transcript-recovery.test.tsx (NEW/untracked, 147 lines; excluded by git diff --stat)
DECISION: Test-only GAP closure part fe-a complete. Exact original 74 GAP mutations now CLOSED. Existing 13 EQUIVALENT and X-019 INCONCLUSIVE were outside this repair scope and are not relabeled. This is author behavior verification, not independent PR approval or live persistence/e2e evidence. No agents launched/returned: 0/0. No commits/push/deploy. NEXT: Owner/reviewer can consume this journal and final evidence JSON; no remaining fe-a GAP row.
FACT: Final SHA-256 manifest written to fe-a-final-sha256.txt (8 files, values rechecked); machine-readable row evidence/diff stat in fe-a-final-evidence.json. No test bytes changed after final whole-suite baseline or the 74 final mutation comparisons. Two new untracked tests add 219 lines, so total including them = 8 test files / 1005 insertions / 6 deletions; tracked git diff --stat remains 6 files / 786 insertions / 6 deletions.
