NEXT: Read shared rules, repository routes and fe-b mutation journal; close each GAP with a baseline-passing behavior test and exact mutation failure.
FACT: Clean HEAD is 64f21ef (gaps fe-a), with only eight test files changed since requested 5a56965; production bytes match requested baseline. Preserve prior tests. Source journal final matrix identifies 60 GAP rows. No delegation.
DECISION: Small test-only repair at owner-agreed seams; existing exact mutation diffs are the discriminators. No new production behavior or architectural choice; plan critique/defense N/A for this scoped verification. File inspection fallback for impact, no graph refresh. Tests/targeted mutation commands and end-of-part frontend gates explicitly authorized.
NEXT: Add draft query overlay/expiry, lesson isolation, repeated request and error behavior tests, then run their exact mutations.
NEXT: Verify exact B017 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B017 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B017.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B017 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: regeneration retains the complete prior snapshot but clears its server timestamp

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'regeneration retains the complete prior snapshot but clears its server timestamp'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "regeneration retains the complete prior snapshot but clears its server timestamp",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 239,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 239
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,21 +1,14 @@
  [
    {
-     "assetId": "asset",
-     "failure": "recognition_failed",
-     "generation": 7,
-     "language": "fr",
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
      "lessonId": "lesson",
-     "segments": [
-       {
-         "endMs": 300,
-         "id": "line",
-         "startMs": 10,
-         "text": "Bonjour",
-       },
-     ],
-     "stale": true,
+     "segments": [],
+     "stale": false,
      "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
      "updatedAt": null,
```

codeFrame (error 1):

```text
237 |   await h.read();
238 |   await h.regenerate();
239 |   expect(await h.read()).toEqual([{ ...previous, state: 'pending', updatedAt: null }]);
    |                          ^
240 |   h.client.clear();
241 | });
```


```
NEXT: Verify exact B019 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B019 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B019.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B019 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: regeneration retains the complete prior snapshot but clears its server timestamp

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'regeneration retains the complete prior snapshot but clears its server timestamp'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "regeneration retains the complete prior snapshot but clears its server timestamp",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 239,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 239
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -16,8 +16,8 @@
      "stale": true,
      "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": null,
+     "updatedAt": "2026-11-16T12:00:00Z",
    },
  ]
```

codeFrame (error 1):

```text
237 |   await h.read();
238 |   await h.regenerate();
239 |   expect(await h.read()).toEqual([{ ...previous, state: 'pending', updatedAt: null }]);
    |                          ^
240 |   h.client.clear();
241 | });
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 0,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to match object [ { state: 'pending', …(1) } ]\n(10 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 251,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 251
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    {
      "state": "pending",
-     "updatedAt": null,
+     "updatedAt": "2026-11-16T12:00:00Z",
    },
  ]
```

codeFrame (error 1):

```text
249 |   await h.regenerate();
250 |   await h.read();
251 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toMatchObject([{ state: 'pending', updatedAt: null }]);
    |                                                                                ^
252 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
253 |   expect(h.interval()).toBe(false);
```


```
NEXT: Verify exact B021 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B021 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B021.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B021 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: markers never leak rows or polling into a sibling lesson

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'markers never leak rows or polling into a sibling lesson'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "markers never leak rows or polling into a sibling lesson",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'sibling', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 263,
        "column": 46,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 263
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "sibling",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
261 |   await h.regenerate();
262 |   const sibling = h.queries.lesson('sibling', h.client);
263 |   expect(await h.client.fetchQuery(sibling)).toEqual([]);
    |                                              ^
264 |   expect(h.interval('sibling')).toBe(false);
265 |   h.client.clear();
```


```
NEXT: Verify exact B025 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B025 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B025.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B025 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 254,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 254
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
252 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
253 |   expect(h.interval()).toBe(false);
254 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(exists ? [previous] : []);
    |                                                                                ^
255 |   h.client.clear();
256 | });
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 254,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 254
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -5,12 +5,12 @@
      "generation": 1,
      "language": "en",
      "lessonId": "lesson",
      "segments": [],
      "stale": false,
-     "state": "ready",
+     "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": "2026-11-16T12:00:00Z",
+     "updatedAt": null,
    },
  ]
```

codeFrame (error 1):

```text
252 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
253 |   expect(h.interval()).toBe(false);
254 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(exists ? [previous] : []);
    |                                                                                ^
255 |   h.client.clear();
256 | });
```

### [F03] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 291,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 291
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

`
```
NEXT: Verify exact B026 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B026 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B026.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B026 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 254,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 254
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
252 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
253 |   expect(h.interval()).toBe(false);
254 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(exists ? [previous] : []);
    |                                                                                ^
255 |   h.client.clear();
256 | });
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 291,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 291
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
289 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
290 |   expect(h.interval()).toBe(false);
291 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual([]);
    |                                                                                ^
292 |   h.client.clear();
293 | });
```


```
NEXT: Verify exact B027 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B027 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B027.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B027 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 254,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 254
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -5,12 +5,12 @@
      "generation": 1,
      "language": "en",
      "lessonId": "lesson",
      "segments": [],
      "stale": false,
-     "state": "ready",
+     "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": "2026-11-16T12:00:00Z",
+     "updatedAt": null,
    },
  ]
```

codeFrame (error 1):

```text
252 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
253 |   expect(h.interval()).toBe(false);
254 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(exists ? [previous] : []);
    |                                                                                ^
255 |   h.client.clear();
256 | });
```


```
NEXT: Verify exact B029 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B029 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B029.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B029 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: running with the same timestamp acknowledges the claim and terminal data stops polling

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'running with the same timestamp acknowledges the claim and terminal data stops polling'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "running with the same timestamp acknowledges the claim and terminal data stops polling",
  "status": "fail",
  "duration": 12,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to match object [ { state: 'running' } ]\n(11 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 273,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 273
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    {
-     "state": "running",
+     "state": "pending",
    },
  ]
```

codeFrame (error 1):

```text
271 |   await h.regenerate();
272 |   h.repository.listDrafts = async () => ({ ok: true, value: [draft(trackT, { state: 'running' })] });
273 |   expect(await h.read()).toMatchObject([{ state: 'running' }]);
    |                          ^
274 |   h.repository.listDrafts = async () => ({ ok: true, value: [draft(trackT)] });
275 |   expect(await h.read()).toEqual([draft(trackT)]);
```


```
NEXT: Verify exact B034 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B034 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B034.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B034 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 5000 to be false // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 290,
        "column": 24,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 290
    }
  ]
}
```

diff:

```diff
- Expected:
false

+ Received:
5000
```

codeFrame (error 1):

```text
288 |   await h.regenerate();
289 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
290 |   expect(h.interval()).toBe(false);
    |                        ^
291 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual([]);
292 |   h.client.clear();
```


```
NEXT: Verify exact B036 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B036 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B036.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B036 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: inventory errors preserve their named code

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'inventory errors preserve their named code'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "inventory errors preserve their named code",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise resolved \"[]\" instead of rejecting",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 300,
        "column": 36,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 300
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
[]
```

codeFrame (error 1):

```text
298 |   h.repository.listDrafts = async () => ({ ok: false, error: { code: 'dependency_unavailable' } });
299 |   const opts = kind === 'inventory' ? h.queries.inventory(lessonId) : h.queries.lesson(lessonId, h.client);
300 |   await expect(h.client.fetchQuery(opts)).rejects.toMatchObject({ name: 'DraftQueryError', code: 'dependency_unavailable' });
    |                                    ^
301 |   h.client.clear();
302 | });
```


```
NEXT: Verify exact B037 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B037 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B037.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B037 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: lesson errors preserve their named code

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'lesson errors preserve their named code'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "lesson errors preserve their named code",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise resolved \"[]\" instead of rejecting",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 300,
        "column": 36,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 300
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
[]
```

codeFrame (error 1):

```text
298 |   h.repository.listDrafts = async () => ({ ok: false, error: { code: 'dependency_unavailable' } });
299 |   const opts = kind === 'inventory' ? h.queries.inventory(lessonId) : h.queries.lesson(lessonId, h.client);
300 |   await expect(h.client.fetchQuery(opts)).rejects.toMatchObject({ name: 'DraftQueryError', code: 'dependency_unavailable' });
    |                                    ^
301 |   h.client.clear();
302 | });
```


```
NEXT: Verify exact B039 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B039 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B039.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B039 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 254,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 254
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
252 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
253 |   expect(h.interval()).toBe(false);
254 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(exists ? [previous] : []);
    |                                                                                ^
255 |   h.client.clear();
256 | });
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 2,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 254,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 254
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -5,12 +5,12 @@
      "generation": 1,
      "language": "en",
      "lessonId": "lesson",
      "segments": [],
      "stale": false,
-     "state": "ready",
+     "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": "2026-11-16T12:00:00Z",
+     "updatedAt": null,
    },
  ]
```

codeFrame (error 1):

```text
252 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
253 |   expect(h.interval()).toBe(false);
254 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(exists ? [previous] : []);
    |                                                                                ^
255 |   h.client.clear();
256 | });
```

### [F03] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 291,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 291
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):


```
NEXT: Verify exact B040 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B040 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B040.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B040 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: markers never leak rows or polling into a sibling lesson

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'markers never leak rows or polling into a sibling lesson'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "markers never leak rows or polling into a sibling lesson",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 5000 to be false // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 264,
        "column": 33,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 264
    }
  ]
}
```

diff:

```diff
- Expected:
false

+ Received:
5000
```

codeFrame (error 1):

```text
262 |   const sibling = h.queries.lesson('sibling', h.client);
263 |   expect(await h.client.fetchQuery(sibling)).toEqual([]);
264 |   expect(h.interval('sibling')).toBe(false);
    |                                 ^
265 |   h.client.clear();
266 | });
```


```
NEXT: Verify exact B047 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B047 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B047.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B047 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: failed regeneration never creates a pending row or starts polling

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'failed regeneration never creates a pending row or starts polling'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "failed regeneration never creates a pending row or starts polling",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 309,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 309
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
307 |   h.repository.regenerate = async () => ({ ok: false, error: { code: 'dependency_unavailable' } });
308 |   await expect(h.regenerate()).rejects.toMatchObject({ code: 'dependency_unavailable' });
309 |   expect(await h.read()).toEqual([]);
    |                          ^
310 |   expect(h.interval()).toBe(false);
311 |   h.client.clear();
```


```
NEXT: Verify exact B160 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B160 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B160.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B160 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 254,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 254
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
252 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
253 |   expect(h.interval()).toBe(false);
254 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(exists ? [previous] : []);
    |                                                                                ^
255 |   h.client.clear();
256 | });
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 291,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 291
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
289 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
290 |   expect(h.interval()).toBe(false);
291 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual([]);
    |                                                                                ^
292 |   h.client.clear();
293 | });
```


```
NEXT: Verify exact B161 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B161 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B161.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B161 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: null-track placeholder does not rewrite the cache while a requested marker waits

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'null-track placeholder does not rewrite the cache while a requested marker waits'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "null-track placeholder does not rewrite the cache while a requested marker waits",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 2000 to be 1000 // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 324,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 324
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- 1000
+ 2000
```

codeFrame (error 1):

```text
322 |   rstest.setSystemTime(2000);
323 |   expect(h.interval()).toBe(draftPollMs);
324 |   expect(h.client.getQueryState(key)?.dataUpdatedAt).toBe(before);
    |                                                      ^
325 |   expect(h.client.getQueryData(key)).toHaveLength(1);
326 |   h.client.clear();
```


```
NEXT: Verify exact B004 mutation against src/app/composition/admin-audio-queries.test.ts.
FACT: B004 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-audio-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B004.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-audio-queries.ts` exit 0.
FACT: B004 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-audio-queries.test.ts :: terminal audio transition refetches only the affected mounted draft

repro:

```bash
npx rstest 'src/app/composition/admin-audio-queries.test.ts' --testNamePattern 'terminal audio transition refetches only the affected mounted draft'
```

details:

```json
{
  "testPath": "src/app/composition/admin-audio-queries.test.ts",
  "project": "dom",
  "fullName": "terminal audio transition refetches only the affected mounted draft",
  "status": "fail",
  "duration": 20,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 1 to be 2 // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-audio-queries.test.ts",
        "line": 223,
        "column": 27,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-audio-queries.test.ts",
      "line": 223
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- 2
+ 1
```

codeFrame (error 1):

```text
221 |     state = 'ready';
222 |     await client.fetchQuery(queries.list(courseId, videoId));
223 |     expect(affectedReads).toBe(before + 1);
    |                           ^
224 |     expect(siblingReads).toBe(siblingBefore);
225 |   } finally { stop(); stopSibling(); client.clear(); }
```


```
NEXT: Verify exact B005 mutation against src/app/composition/admin-audio-queries.test.ts.
FACT: B005 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-audio-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B005.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-audio-queries.ts` exit 0.
FACT: B005 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-audio-queries.test.ts :: list envelope with queued audio schedules another port read

repro:

```bash
npx rstest 'src/app/composition/admin-audio-queries.test.ts' --testNamePattern 'list envelope with queued audio schedules another port read'
```

details:

```json
{
  "testPath": "src/app/composition/admin-audio-queries.test.ts",
  "project": "dom",
  "fullName": "list envelope with queued audio schedules another port read",
  "status": "fail",
  "duration": 25,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 1 to be 2 // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-audio-queries.test.ts",
        "line": 245,
        "column": 19,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-audio-queries.test.ts",
      "line": 245
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- 2
+ 1
```

codeFrame (error 1):

```text
243 |     const before = reads;
244 |     await rstest.advanceTimersByTimeAsync(videoPollMs);
245 |     expect(reads).toBe(before + 1);
    |                   ^
246 |   } finally { stop(); client.clear(); rstest.useRealTimers(); }
247 | });
```

### [F02] src/app/composition/admin-audio-queries.test.ts :: list envelope with processing audio schedules another port read

repro:

```bash
npx rstest 'src/app/composition/admin-audio-queries.test.ts' --testNamePattern 'list envelope with processing audio schedules another port read'
```

details:

```json
{
  "testPath": "src/app/composition/admin-audio-queries.test.ts",
  "project": "dom",
  "fullName": "list envelope with processing audio schedules another port read",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 1 to be 2 // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-audio-queries.test.ts",
        "line": 245,
        "column": 19,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-audio-queries.test.ts",
      "line": 245
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- 2
+ 1
```

codeFrame (error 1):

```text
243 |     const before = reads;
244 |     await rstest.advanceTimersByTimeAsync(videoPollMs);
245 |     expect(reads).toBe(before + 1);
    |                   ^
246 |   } finally { stop(); client.clear(); rstest.useRealTimers(); }
247 | });
```


```
NEXT: Verify exact B053 mutation against src/modules/media/adapters/audio-track-repository.test.ts.
FACT: B053 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/adapters/audio-track-repository.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B053.log. Production restore `git diff --quiet -- frontend/src/modules/media/adapters/audio-track-schema.ts` exit 0.
FACT: B053 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/adapters/audio-track-repository.test.ts :: audio track API repository > rejects a list with invalid list multilingualDefault false

repro:

```bash
npx rstest 'src/modules/media/adapters/audio-track-repository.test.ts' --testNamePattern 'audio track API repository > rejects a list with invalid list multilingualDefault false'
```

details:

```json
{
  "testPath": "src/modules/media/adapters/audio-track-repository.test.ts",
  "project": "adapter-node",
  "fullName": "audio track API repository > rejects a list with invalid list multilingualDefault false",
  "status": "fail",
  "duration": 4,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: { …(2) } } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/media/adapters/audio-track-repository.test.ts",
        "line": 121,
        "column": 65,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/adapters/audio-track-repository.test.ts",
      "line": 121
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
+   "value": {
+     "items": [
+       {
+         "isDefault": false,
+         "kind": "uploaded",
+         "label": "English",
+         "language": "en",
+         "multilingual": false,
+         "state": "ready",
+         "trackId": "60000000-0000-4000-8000-000000000001",
+         "updatedAt": "2026-11-13T12:00:00.123456Z",
+         "videoId": "50000000-0000-4000-8000-000000000001",
+       },
+     ],
+     "multilingualDefault": "false",
    },
-   "ok": false,
  }
```

codeFrame (error 1):

```text
119 |   ])('rejects a list with $name', async ({ response }) => {
120 |     server.use(http.get(api, () => HttpResponse.json(response)));
121 |     expect(await createApiAudioTrackRepository().list(videoId)).toEqual({
    |                                                                 ^
122 |       ok: false,
123 |       error: { code: 'invalid_response' },
```


```
NEXT: Verify exact B054 mutation against src/modules/media/adapters/audio-track-repository.test.ts.
FACT: B054 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/adapters/audio-track-repository.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B054.log. Production restore `git diff --quiet -- frontend/src/modules/media/adapters/audio-track-schema.ts` exit 0.
FACT: B054 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/adapters/audio-track-repository.test.ts :: audio track API repository > rejects a list with unknown envelope key

repro:

```bash
npx rstest 'src/modules/media/adapters/audio-track-repository.test.ts' --testNamePattern 'audio track API repository > rejects a list with unknown envelope key'
```

details:

```json
{
  "testPath": "src/modules/media/adapters/audio-track-repository.test.ts",
  "project": "adapter-node",
  "fullName": "audio track API repository > rejects a list with unknown envelope key",
  "status": "fail",
  "duration": 5,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { ok: true, value: { …(3) } } to deeply equal { ok: false, …(1) }",
      "topFrame": {
        "file": "src/modules/media/adapters/audio-track-repository.test.ts",
        "line": 121,
        "column": 65,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/adapters/audio-track-repository.test.ts",
      "line": 121
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
+   "value": {
+     "extra": true,
+     "items": [
+       {
+         "isDefault": false,
+         "kind": "uploaded",
+         "label": "English",
+         "language": "en",
+         "multilingual": false,
+         "state": "ready",
+         "trackId": "60000000-0000-4000-8000-000000000001",
+         "updatedAt": "2026-11-13T12:00:00.123456Z",
+         "videoId": "50000000-0000-4000-8000-000000000001",
        },
-   "ok": false,
+     ],
+     "multilingualDefault": false,
+   },
  }
```

codeFrame (error 1):

```text
119 |   ])('rejects a list with $name', async ({ response }) => {
120 |     server.use(http.get(api, () => HttpResponse.json(response)));
121 |     expect(await createApiAudioTrackRepository().list(videoId)).toEqual({
    |                                                                 ^
122 |       ok: false,
123 |       error: { code: 'invalid_response' },
```


```
NEXT: Verify exact B189 mutation against src/modules/media/adapters/audio-track-repository.test.ts.
FACT: B189 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/adapters/audio-track-repository.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B189.log. Production restore `git diff --quiet -- frontend/src/modules/media/application/audio-tracks.ts` exit 0.
FACT: B189 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/adapters/audio-track-repository.test.ts :: rename forwards the same cancellation signal as the sixth repository argument

repro:

```bash
npx rstest 'src/modules/media/adapters/audio-track-repository.test.ts' --testNamePattern 'rename forwards the same cancellation signal as the sixth repository argument'
```

details:

```json
{
  "testPath": "src/modules/media/adapters/audio-track-repository.test.ts",
  "project": "adapter-node",
  "fullName": "rename forwards the same cancellation signal as the sixth repository argument",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"rename\" to be called with arguments: [ …(6) ], but got [ [ …(6) ] ]",
      "topFrame": {
        "file": "src/modules/media/adapters/audio-track-repository.test.ts",
        "line": 357,
        "column": 18,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/adapters/audio-track-repository.test.ts",
      "line": 357
    }
  ]
}
```

codeFrame (error 1):

```text
355 |   const controller = new AbortController();
356 |   await createAudioTracks(repository).rename(videoId, trackId, 'en', 'English', true, controller.signal);
357 |   expect(rename).toHaveBeenCalledWith(videoId, trackId, 'en', 'English', true, controller.signal);
    |                  ^
358 | });
359 | 
```


```
NEXT: Verify exact B060 mutation against src/modules/media/ui/LessonAudioTracksField.test.tsx.
FACT: B060 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/ui/LessonAudioTracksField.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B060.log. Production restore `git diff --quiet -- frontend/src/modules/media/ui/LessonAudioTracksField.tsx` exit 0.
FACT: B060 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/ui/LessonAudioTracksField.test.tsx :: upload switch is described, remains focused and read-only during upload, then resets to server default

repro:

```bash
npx rstest 'src/modules/media/ui/LessonAudioTracksField.test.tsx' --testNamePattern 'upload switch is described, remains focused and read-only during upload, then resets to server default'
```

details:

```json
{
  "testPath": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
  "project": "dom",
  "fullName": "upload switch is described, remains focused and read-only during upload, then resets to server default",
  "status": "fail",
  "duration": 96,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <input …(6)></input> to have property \"checked\" with value true",
      "topFrame": {
        "file": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
        "line": 481,
        "column": 17,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
      "line": 481
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
479 |   finish();
480 |   const reset = await screen.findByRole('switch', { name: 'Multiple languages' });
481 |   expect(reset).toHaveProperty('checked', true);
    |                 ^
482 |   await user.click(screen.getByRole('button', { name: /Choose a language Track language/ }));
483 |   await user.click(screen.getByRole('option', { name: 'Polish' }));
```


```
NEXT: Verify exact B061 mutation against src/modules/media/ui/LessonAudioTracksField.test.tsx.
FACT: B061 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/ui/LessonAudioTracksField.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B061.log. Production restore `git diff --quiet -- frontend/src/modules/media/ui/LessonAudioTracksField.tsx` exit 0.
FACT: B061 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/ui/LessonAudioTracksField.test.tsx :: pending row action makes the visible upload switch read-only without dropping focus

repro:

```bash
npx rstest 'src/modules/media/ui/LessonAudioTracksField.test.tsx' --testNamePattern 'pending row action makes the visible upload switch read-only without dropping focus'
```

details:

```json
{
  "testPath": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
  "project": "dom",
  "fullName": "pending row action makes the visible upload switch read-only without dropping focus",
  "status": "fail",
  "duration": 27,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to be 'true' // Object.is equality",
      "topFrame": {
        "file": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
        "line": 501,
        "column": 48,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
      "line": 501
    }
  ]
}
```

diff:

```diff
- Expected:
"true"

+ Received:
null
```

codeFrame (error 1):

```text
499 |   await user.click(screen.getByRole('switch', { name: /Multiple languages in track/ }));
500 |   upload.focus();
501 |   expect(upload.getAttribute('aria-readonly')).toBe('true');
    |                                                ^
502 |   expect(upload.getAttribute('aria-disabled')).toBe('true');
503 |   await user.keyboard(' ');
```


```
NEXT: Verify exact B063 mutation against src/modules/media/ui/LessonAudioTracksField.test.tsx.
FACT: B063 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/ui/LessonAudioTracksField.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B063.log. Production restore `git diff --quiet -- frontend/src/modules/media/ui/LessonAudioTracksField.tsx` exit 0.
FACT: B063 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/ui/LessonAudioTracksField.test.tsx :: pending row action makes the visible upload switch read-only without dropping focus

repro:

```bash
npx rstest 'src/modules/media/ui/LessonAudioTracksField.test.tsx' --testNamePattern 'pending row action makes the visible upload switch read-only without dropping focus'
```

details:

```json
{
  "testPath": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
  "project": "dom",
  "fullName": "pending row action makes the visible upload switch read-only without dropping focus",
  "status": "fail",
  "duration": 25,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to be 'true' // Object.is equality",
      "topFrame": {
        "file": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
        "line": 502,
        "column": 48,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
      "line": 502
    }
  ]
}
```

diff:

```diff
- Expected:
"true"

+ Received:
null
```

codeFrame (error 1):

```text
500 |   upload.focus();
501 |   expect(upload.getAttribute('aria-readonly')).toBe('true');
502 |   expect(upload.getAttribute('aria-disabled')).toBe('true');
    |                                                ^
503 |   await user.keyboard(' ');
504 |   await user.click(upload);
```


```
NEXT: Verify exact B064 mutation against src/modules/media/ui/LessonAudioTracksField.test.tsx.
FACT: B064 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/ui/LessonAudioTracksField.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B064.log. Production restore `git diff --quiet -- frontend/src/modules/media/ui/LessonAudioTracksField.tsx` exit 0.
FACT: B064 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/ui/LessonAudioTracksField.test.tsx :: upload switch is described, remains focused and read-only during upload, then resets to server default

repro:

```bash
npx rstest 'src/modules/media/ui/LessonAudioTracksField.test.tsx' --testNamePattern 'upload switch is described, remains focused and read-only during upload, then resets to server default'
```

details:

```json
{
  "testPath": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
  "project": "dom",
  "fullName": "upload switch is described, remains focused and read-only during upload, then resets to server default",
  "status": "fail",
  "duration": 16,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to be truthy",
      "topFrame": {
        "file": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
        "line": 466,
        "column": 16,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
      "line": 466
    }
  ]
}
```

diff:

```diff
- Expected:
true

+ Received:
null
```

codeFrame (error 1):

```text
464 |   expect(new Set(descriptions).size).toBe(3);
465 |   for (const id of descriptions) {
466 |     expect(id).toBeTruthy();
    |                ^
467 |     expect(document.getElementById(id ?? '')?.textContent).toBe('Recognition detects the language of each passage.');
468 |   }
```


```
NEXT: Verify exact B066 mutation against src/modules/media/ui/LessonAudioTracksField.test.tsx.
FACT: B066 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/ui/LessonAudioTracksField.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B066.log. Production restore `git diff --quiet -- frontend/src/modules/media/ui/LessonAudioTracksField.tsx` exit 0.
FACT: B066 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/ui/LessonAudioTracksField.test.tsx :: row switch toggles with Space, sends complete metadata and keeps focus while pending/error

repro:

```bash
npx rstest 'src/modules/media/ui/LessonAudioTracksField.test.tsx' --testNamePattern 'row switch toggles with Space, sends complete metadata and keeps focus while pending/error'
```

details:

```json
{
  "testPath": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
  "project": "dom",
  "fullName": "row switch toggles with Space, sends complete metadata and keeps focus while pending/error",
  "status": "fail",
  "duration": 21,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to be 'true' // Object.is equality",
      "topFrame": {
        "file": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
        "line": 441,
        "column": 48,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
      "line": 441
    }
  ]
}
```

diff:

```diff
- Expected:
"true"

+ Received:
null
```

codeFrame (error 1):

```text
439 |   expect(calls).toEqual([[row, row.language, row.label, false]]);
440 |   expect(toggle.getAttribute('aria-disabled')).toBe('true');
441 |   expect(toggle.getAttribute('aria-readonly')).toBe('true');
    |                                                ^
442 |   expect(document.activeElement).toBe(toggle);
443 |   await user.keyboard(' ');
```


```
NEXT: Verify exact B070 mutation against src/modules/media/ui/LessonAudioTracksField.test.tsx.
FACT: B070 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/ui/LessonAudioTracksField.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B070.log. Production restore `git diff --quiet -- frontend/src/modules/media/ui/LessonAudioTracksField.tsx` exit 0.
FACT: B070 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/ui/LessonAudioTracksField.test.tsx :: upload switch is described, remains focused and read-only during upload, then resets to server default

repro:

```bash
npx rstest 'src/modules/media/ui/LessonAudioTracksField.test.tsx' --testNamePattern 'upload switch is described, remains focused and read-only during upload, then resets to server default'
```

details:

```json
{
  "testPath": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
  "project": "dom",
  "fullName": "upload switch is described, remains focused and read-only during upload, then resets to server default",
  "status": "fail",
  "duration": 18,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 2 to be 3 // Object.is equality",
      "topFrame": {
        "file": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
        "line": 464,
        "column": 38,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
      "line": 464
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- 3
+ 2
```

codeFrame (error 1):

```text
462 |   const switches = screen.getAllByRole('switch');
463 |   const descriptions = switches.map((toggle) => toggle.getAttribute('aria-describedby'));
464 |   expect(new Set(descriptions).size).toBe(3);
    |                                      ^
465 |   for (const id of descriptions) {
466 |     expect(id).toBeTruthy();
```


```
NEXT: Verify exact B079 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B079 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B079.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B079 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: orphan languages sort after tracks and uppercase track language never duplicates an orphan

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'orphan languages sort after tracks and uppercase track language never duplicates an orphan'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "orphan languages sort after tracks and uppercase track language never duplicates an orphan",
  "status": "fail",
  "duration": 25,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'track:one', 'lang:uk', …(2) ] to deeply equal [ 'track:one', 'lang:de', …(2) ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1165,
        "column": 6,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1165
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    "track:one",
-   "lang:de",
-   "lang:fr",
    "lang:uk",
+   "lang:fr",
+   "lang:de",
  ]
```

codeFrame (error 1):

```text
1163 | test('orphan languages sort after tracks and uppercase track language never duplicates an orphan', () => {
1164 |   expect(lessonLanguageTabs([track('one', 'EN')], [item('uk'), item('en'), item('de'), item('fr')]).map((tab) => tab.id))
1165 |     .toEqual(['track:one', 'lang:de', 'lang:fr', 'lang:uk']);
     |      ^
1166 | });
1167 | 
```


```
NEXT: Verify exact B082 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B082 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B082.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B082 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: tabs select newest Ready generation of this lesson, ignoring foreign and processing videos

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'tabs select newest Ready generation of this lesson, ignoring foreign and processing videos'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "tabs select newest Ready generation of this lesson, ignoring foreign and processing videos",
  "status": "fail",
  "duration": 41,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'foreign' ] to deeply equal [ 'newest-ready' ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1180,
        "column": 60,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1180
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
-   "newest-ready",
+   "foreign",
  ]
```

codeFrame (error 1):

```text
1178 |   const asked = rstest.spyOn(f.video.audio, 'useTracks');
1179 |   await mount(f);
1180 |   expect([...new Set(asked.mock.calls.map(([id]) => id))]).toEqual(['newest-ready']);
     |                                                            ^
1181 | });
1182 | 
```


```
NEXT: Verify exact B083 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B083 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B083.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B083 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: tabs select newest Ready generation of this lesson, ignoring foreign and processing videos

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'tabs select newest Ready generation of this lesson, ignoring foreign and processing videos'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "tabs select newest Ready generation of this lesson, ignoring foreign and processing videos",
  "status": "fail",
  "duration": 36,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'processing' ] to deeply equal [ 'newest-ready' ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1180,
        "column": 60,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1180
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
-   "newest-ready",
+   "processing",
  ]
```

codeFrame (error 1):

```text
1178 |   const asked = rstest.spyOn(f.video.audio, 'useTracks');
1179 |   await mount(f);
1180 |   expect([...new Set(asked.mock.calls.map(([id]) => id))]).toEqual(['newest-ready']);
     |                                                            ^
1181 | });
1182 | 
```

### [F02] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: only processing videos and missing optional audio both retain the single editor

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'only processing videos and missing optional audio both retain the single editor'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "only processing videos and missing optional audio both retain the single editor",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <div class=\"ui-tab-list\" …(7)>…(1)</div> to be null",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1187,
        "column": 41,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1187
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
<div
  aria-label="Transcript languages"
  aria-orientation="horizontal"
  class="ui-tab-list"
  data-collection="react-aria-_r_2n8_"
  data-orientation="horizontal"
  data-rac=""
  id="react-aria-_r_2n9_"
  role="tablist"
>
  <div
    aria-controls="react-aria-_r_2n9_-tabpanel-track:one"
    aria-selected="true"
    class="ui-tab"
    data-collection="react-aria-_r_2n8_"
    data-key="track:one"
    data-rac=""
    data-react-aria-pressable="true"
    data-selected="true"
    id="react-aria-_r_2n9_-tab-track:one"
    role="tab"
    tabindex="0"
  >
    English · Default
  </div>
</div>
```

codeFrame (error 1):

```text
1185 |   f.video.rows = f.video.rows.map((row) => ({ ...row, state: 'processing' }));
1186 |   const first = await mount(f);
1187 |   expect(screen.queryByRole('tablist')).toBeNull();
     |                                         ^
1188 |   expect(screen.getByRole('heading', { name: 'Lesson transcript' })).toBeTruthy();
1189 |   first.unmount();
```


```
NEXT: Verify exact B084 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B084 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B084.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B084 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: tabs select newest Ready generation of this lesson, ignoring foreign and processing videos

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'tabs select newest Ready generation of this lesson, ignoring foreign and processing videos'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "tabs select newest Ready generation of this lesson, ignoring foreign and processing videos",
  "status": "fail",
  "duration": 37,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'older' ] to deeply equal [ 'newest-ready' ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1180,
        "column": 60,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1180
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
-   "newest-ready",
+   "older",
  ]
```

codeFrame (error 1):

```text
1178 |   const asked = rstest.spyOn(f.video.audio, 'useTracks');
1179 |   await mount(f);
1180 |   expect([...new Set(asked.mock.calls.map(([id]) => id))]).toEqual(['newest-ready']);
     |                                                            ^
1181 | });
1182 | 
```


```
NEXT: Verify exact B086 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B086 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B086.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B086 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: single editor known state comes from complete inventory: loading, empty, and other language

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'single editor known state comes from complete inventory: loading, empty, and other language'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "single editor known state comes from complete inventory: loading, empty, and other language",
  "status": "fail",
  "duration": 29,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ false, false ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1202,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1202
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
+   false,
+   false,
+ ]
```

codeFrame (error 1):

```text
1200 |   f.languages.useInventory = () => ({ data: undefined });
1201 |   const view = await mount(f, false);
1202 |   expect(f.state.known).toEqual([]);
     |                         ^
1203 |   f.languages.useInventory = () => ({ data: [] });
1204 |   view.refresh();
```


```
NEXT: Verify exact B087 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B087 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B087.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B087 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: single editor known state comes from complete inventory: loading, empty, and other language

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'single editor known state comes from complete inventory: loading, empty, and other language'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "single editor known state comes from complete inventory: loading, empty, and other language",
  "status": "fail",
  "duration": 30,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected false to be true // Object.is equality",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1208,
        "column": 32,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1208
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
1206 |   f.languages.useInventory = () => ({ data: [item('fr')] });
1207 |   view.refresh();
1208 |   expect(f.state.known.at(-1)).toBe(true);
     |                                ^
1209 |   expect(f.state.known.filter((known) => !known)).toHaveLength(1);
1210 | });
```


```
NEXT: Verify exact B089 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B089 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B089.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B089 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: recovery waits for both lists before reading and choosing ownership

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'recovery waits for both lists before reading and choosing ownership'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "recovery waits for both lists before reading and choosing ownership",
  "status": "fail",
  "duration": 25,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected \"keys\" to not be called at all, but actually been called 1 times\n\nReceived:\n\n  1st keys call:\n\n    Array []\n\n\nNumber of calls: 1\n",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1219,
        "column": 20,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1219
    }
  ]
}
```

codeFrame (error 1):

```text
1217 |   const read = rstest.spyOn(f.historyStore, 'read');
1218 |   const view = await mount(f);
1219 |   expect(keys).not.toHaveBeenCalled();
     |                    ^
1220 |   expect(read).not.toHaveBeenCalled();
1221 |   expect(screen.queryByRole('button', { name: 'Restore changes' })).toBeNull();
```


```
NEXT: Verify exact B094 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B094 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B094.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B094 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: only default track includes Default in its accessible tab name

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'only default track includes Default in its accessible tab name'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "only default track includes Default in its accessible tab name",
  "status": "fail",
  "duration": 37,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'English', 'Ukrainian' ] to deeply equal [ 'English · Default', 'Ukrainian' ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1232,
        "column": 68,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1232
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
-   "English · Default",
+   "English",
    "Ukrainian",
  ]
```

codeFrame (error 1):

```text
1230 |   f.state.tracks.push(track('two', 'uk', { isDefault: false }));
1231 |   await mount(f);
1232 |   expect(screen.getAllByRole('tab').map((tab) => tab.textContent)).toEqual(['English · Default', 'Ukrainian']);
     |                                                                    ^
1233 | });
1234 | 
```


```
NEXT: Verify exact B107 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B107 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B107.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B107 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: fresh Save rejects different id in same language and retains lines

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'fresh Save rejects different id in same language and retains lines'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "fresh Save rejects different id in same language and retains lines",
  "status": "fail",
  "duration": 1077,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an element with the text: The track changed or was removed. Reload the tracks and try again.. This could be because the text is broken up by multiple elements. In this case, you can provide a function for your text matcher to make your matcher more flexible.\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <form\n      data-testid=\"course-form\"\n    >\n      <template />\n      <div\n        class=\"ui-tabs\"\n        data-orientation=\"horizontal\"\n        data-rac=\"\"\n      >\n        <div\n          aria-label=\"Transcript languages\"\n          aria-orientation=\"horizontal\"\n          class=\"ui-tab-list\"\n          data-collection=\"react-aria-_r_2r0_\"\n          data-orientation=\"horizontal\"\n          data-rac=\"\"\n          id=\"react-aria-_r_2r1_\"\n          role=\"tablist\"\n        >\n          <div\n            aria-controls=\"react-aria-_r_2r1_-tabpanel-track:one\"\n            aria-selected=\"true\"\n            class=\"ui-tab\"\n            data-collection=\"react-aria-_r_2r0_\"\n            data-key=\"track:one\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            data-selected=\"true\"\n            id=\"react-aria-_r_2r1_-tab-track:one\"\n            role=\"tab\"\n            tabindex=\"0\"\n          >\n            English · Default\n          </div>\n        </div>\n        <div\n          aria-labelledby=\"react-aria-_r_2r1_-tab-track:one\"\n          class=\"lesson-language-panel ui-tab-panel\"\n          data-rac=\"\"\n          id=\"react-aria-_r_2r1_-tabpanel-track:one\"\n          role=\"tabpanel\"\n        >\n          <section\n            aria-busy=\"false\"\n            aria-labelledby=\"lesson-draft-lesson-track:one\"\n            class=\"admin-draft-field\"\n          >\n            <h3\n              id=\"lesson-draft-lesson-track:one\"\n            >\n              Speech recognition\n            </h3>\n            <div\n              class=\"admin-draft-regenerate\"\n            >\n              <button\n                class=\"ui-button ui-button--secondary\"\n                data-rac=\"\"\n                data-react-aria-pressable=\"true\"\n                id=\"react-aria-_r_2r4_\"\n                tabindex=\"0\"\n                type=\"button\"\n              >\n                Generate transcript\n              </button>\n            </div>\n          </section>\n          <section\n            aria-labelledby=\"transcript-editor-lesson-track:one\"\n            class=\"transcript-editor\"\n          >\n            <h4\n              id=\"transcript-editor-lesson-track:one\"\n            >\n              Lesson transcript\n            </h4>\n            <ol\n              class=\"transcript-editor-lines\"\n            >\n              <li>\n                <div\n                  class=\"ui-field\"\n                  data-rac=\"\"\n                >\n                  <label\n                    class=\"react-aria-Label\"\n                    for=\"react-aria-_r_2r6_\"\n                    id=\"react-aria-_r_2r7_\"\n                  >\n                    Line 1 start, ms\n                  </label>\n                  <input\n                    aria-labelledby=\"react-aria-_r_2r7_\"\n                    class=\"ui-input\"\n                    data-rac=\"\"\n                    id=\"react-aria-_r_2r6_\"\n                    tabindex=\"0\"\n                    title=\"\"\n                    type=\"text\"\n                    value=\"0\"\n                  />\n                </div>\n                <div\n                  class=\"ui-field\"\n                  data-rac=\"\"\n                >\n                  <label\n                    class=\"react-aria-Label\"\n                    for=\"react-aria-_r_2rb_\"\n                    id=\"react-aria-_r_2rc_\"\n                  >\n                    Line 1 end, ms\n                  </label>\n                  <input\n                    aria-labelledby=\"react-aria-_r_2rc_\"\n                    class=\"ui-input\"\n                    data-rac=\"\"\n                    id=\"react-aria-_r_2rb_\"\n                    tabindex=\"0\"\n                    title=\"\"\n                    type=\"text\"\n                    value=\"1000\"\n                  />\n                </div>\n                <div\n                  class
```
NEXT: Verify exact B108 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B108 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B108.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B108 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: fresh Save rejects processing and retains lines

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'fresh Save rejects processing and retains lines'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "fresh Save rejects processing and retains lines",
  "status": "fail",
  "duration": 1083,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an element with the text: The track changed or was removed. Reload the tracks and try again.. This could be because the text is broken up by multiple elements. In this case, you can provide a function for your text matcher to make your matcher more flexible.\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <form\n      data-testid=\"course-form\"\n    >\n      <template />\n      <div\n        class=\"ui-tabs\"\n        data-orientation=\"horizontal\"\n        data-rac=\"\"\n      >\n        <div\n          aria-label=\"Transcript languages\"\n          aria-orientation=\"horizontal\"\n          class=\"ui-tab-list\"\n          data-collection=\"react-aria-_r_2rs_\"\n          data-orientation=\"horizontal\"\n          data-rac=\"\"\n          id=\"react-aria-_r_2rt_\"\n          role=\"tablist\"\n        >\n          <div\n            aria-controls=\"react-aria-_r_2rt_-tabpanel-track:one\"\n            aria-selected=\"true\"\n            class=\"ui-tab\"\n            data-collection=\"react-aria-_r_2rs_\"\n            data-key=\"track:one\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            data-selected=\"true\"\n            id=\"react-aria-_r_2rt_-tab-track:one\"\n            role=\"tab\"\n            tabindex=\"0\"\n          >\n            English · Default\n          </div>\n        </div>\n        <div\n          aria-labelledby=\"react-aria-_r_2rt_-tab-track:one\"\n          class=\"lesson-language-panel ui-tab-panel\"\n          data-rac=\"\"\n          id=\"react-aria-_r_2rt_-tabpanel-track:one\"\n          role=\"tabpanel\"\n        >\n          <section\n            aria-busy=\"false\"\n            aria-labelledby=\"lesson-draft-lesson-track:one\"\n            class=\"admin-draft-field\"\n          >\n            <h3\n              id=\"lesson-draft-lesson-track:one\"\n            >\n              Speech recognition\n            </h3>\n            <div\n              class=\"admin-draft-regenerate\"\n            >\n              <button\n                class=\"ui-button ui-button--secondary\"\n                data-rac=\"\"\n                data-react-aria-pressable=\"true\"\n                id=\"react-aria-_r_2s0_\"\n                tabindex=\"0\"\n                type=\"button\"\n              >\n                Generate transcript\n              </button>\n            </div>\n          </section>\n          <section\n            aria-labelledby=\"transcript-editor-lesson-track:one\"\n            class=\"transcript-editor\"\n          >\n            <h4\n              id=\"transcript-editor-lesson-track:one\"\n            >\n              Lesson transcript\n            </h4>\n            <ol\n              class=\"transcript-editor-lines\"\n            >\n              <li>\n                <div\n                  class=\"ui-field\"\n                  data-rac=\"\"\n                >\n                  <label\n                    class=\"react-aria-Label\"\n                    for=\"react-aria-_r_2s2_\"\n                    id=\"react-aria-_r_2s3_\"\n                  >\n                    Line 1 start, ms\n                  </label>\n                  <input\n                    aria-labelledby=\"react-aria-_r_2s3_\"\n                    class=\"ui-input\"\n                    data-rac=\"\"\n                    id=\"react-aria-_r_2s2_\"\n                    tabindex=\"0\"\n                    title=\"\"\n                    type=\"text\"\n                    value=\"0\"\n                  />\n                </div>\n                <div\n                  class=\"ui-field\"\n                  data-rac=\"\"\n                >\n                  <label\n                    class=\"react-aria-Label\"\n                    for=\"react-aria-_r_2s7_\"\n                    id=\"react-aria-_r_2s8_\"\n                  >\n                    Line 1 end, ms\n                  </label>\n                  <input\n                    aria-labelledby=\"react-aria-_r_2s8_\"\n                    class=\"ui-input\"\n                    data-rac=\"\"\n                    id=\"react-aria-_r_2s7_\"\n                    tabindex=\"0\"\n                    title=\"\"\n                    type=\"text\"\n                    value=\"1000\"\n                  />\n                </div>\n                <div\n                  class=\"ui-field\"\n                  data-rac=\"\"\n         
```
NEXT: Verify exact B121 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B121 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B121.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B121 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: thrown rename retains own recovery and lines without PUT

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'thrown rename retains own recovery and lines without PUT'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "thrown rename retains own recovery and lines without PUT",
  "status": "fail",
  "duration": 1086,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an element with the text: Could not rename the track — try again.. This could be because the text is broken up by multiple elements. In this case, you can provide a function for your text matcher to make your matcher more flexible.\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <form\n      data-testid=\"course-form\"\n    >\n      <template />\n      <div\n        class=\"ui-tabs\"\n        data-orientation=\"horizontal\"\n        data-rac=\"\"\n      >\n        <div\n          aria-label=\"Transcript languages\"\n          aria-orientation=\"horizontal\"\n          class=\"ui-tab-list\"\n          data-collection=\"react-aria-_r_2ug_\"\n          data-orientation=\"horizontal\"\n          data-rac=\"\"\n          id=\"react-aria-_r_2uh_\"\n          role=\"tablist\"\n        >\n          <div\n            aria-controls=\"react-aria-_r_2uh_-tabpanel-track:one\"\n            aria-selected=\"true\"\n            class=\"ui-tab\"\n            data-collection=\"react-aria-_r_2ug_\"\n            data-key=\"track:one\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            data-selected=\"true\"\n            id=\"react-aria-_r_2uh_-tab-track:one\"\n            role=\"tab\"\n            tabindex=\"0\"\n          >\n            Unknown language · Default · Multiple languages\n          </div>\n        </div>\n        <div\n          aria-labelledby=\"react-aria-_r_2uh_-tab-track:one\"\n          class=\"lesson-language-panel ui-tab-panel\"\n          data-rac=\"\"\n          id=\"react-aria-_r_2uh_-tabpanel-track:one\"\n          role=\"tabpanel\"\n        >\n          <section\n            aria-busy=\"false\"\n            aria-labelledby=\"lesson-draft-lesson-track:one\"\n            class=\"admin-draft-field\"\n          >\n            <h3\n              id=\"lesson-draft-lesson-track:one\"\n            >\n              Speech recognition\n            </h3>\n            <output\n              class=\"admin-draft-state\"\n              data-draft-state=\"ready\"\n            >\n              Draft ready\n            </output>\n            <button\n              class=\"ui-button ui-button--secondary\"\n              data-rac=\"\"\n              data-react-aria-pressable=\"true\"\n              id=\"react-aria-_r_2uk_\"\n              tabindex=\"0\"\n              type=\"button\"\n            >\n              Load draft into editor\n            </button>\n            <div\n              class=\"admin-draft-regenerate\"\n            >\n              <button\n                class=\"ui-button ui-button--secondary\"\n                data-rac=\"\"\n                data-react-aria-pressable=\"true\"\n                id=\"react-aria-_r_2um_\"\n                tabindex=\"0\"\n                type=\"button\"\n              >\n                Regenerate transcript\n              </button>\n            </div>\n          </section>\n          <section\n            aria-labelledby=\"transcript-editor-lesson-track:one\"\n            class=\"transcript-editor\"\n          >\n            <h4\n              id=\"transcript-editor-lesson-track:one\"\n            >\n              Lesson transcript\n            </h4>\n            <p>\n              No lines yet.\n            </p>\n            <ol\n              class=\"transcript-editor-lines\"\n            />\n            <button\n              aria-disabled=\"false\"\n              class=\"ui-button ui-button--secondary\"\n              data-add-line=\"true\"\n              data-rac=\"\"\n              data-react-aria-pressable=\"true\"\n              id=\"react-aria-_r_2uo_\"\n              tabindex=\"0\"\n              type=\"button\"\n            >\n              Add a line\n            </button>\n            <button\n              aria-disabled=\"false\"\n              class=\"ui-button ui-button--primary\"\n              data-rac=\"\"\n              data-react-aria-pressable=\"true\"\n              id=\"react-aria-_r_2uq_\"\n              tabindex=\"0\"\n              type=\"button\"\n            >\n              Save the transcript\n            </button>\n            <output>\n              Transcript saved.\n            </output>\n          </section>\n        </div>\n      </div>\n    </form>\n  </div>\n</body>\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <form\n      data-testid=\"course-form\"\n   
```
NEXT: Verify exact B122 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B122 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B122.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B122 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: und restored token is never forwarded to newly chosen language

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'und restored token is never forwarded to newly chosen language'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "und restored token is never forwarded to newly chosen language",
  "status": "fail",
  "duration": 73,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'lesson', 'en', 'SOURCE' ] to deeply equal [ 'lesson', 'en', undefined ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1272,
        "column": 40,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1272
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    "lesson",
    "en",
-   undefined,
+   "SOURCE",
  ]
```

codeFrame (error 1):

```text
1270 |   await screen.findByText('Transcript saved.');
1271 |   expect(f.state.operations).toEqual(['PATCH', 'PUT']);
1272 |   expect(f.state.puts[0]?.slice(0, 3)).toEqual(['lesson', 'en', undefined]);
     |                                        ^
1273 | });
1274 | 
```


```
NEXT: Verify exact B129 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B129 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B129.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B129 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: post-confirmation track read failure retains lines without PUT

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'post-confirmation track read failure retains lines without PUT'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "post-confirmation track read failure retains lines without PUT",
  "status": "fail",
  "duration": 1115,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an element with the text: Could not verify the track — try again.. This could be because the text is broken up by multiple elements. In this case, you can provide a function for your text matcher to make your matcher more flexible.\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <form\n      data-testid=\"course-form\"\n    >\n      <template />\n      <div\n        class=\"ui-tabs\"\n        data-orientation=\"horizontal\"\n        data-rac=\"\"\n      >\n        <div\n          aria-label=\"Transcript languages\"\n          aria-orientation=\"horizontal\"\n          class=\"ui-tab-list\"\n          data-collection=\"react-aria-_r_31m_\"\n          data-orientation=\"horizontal\"\n          data-rac=\"\"\n          id=\"react-aria-_r_31n_\"\n          role=\"tablist\"\n        >\n          <div\n            aria-controls=\"react-aria-_r_31n_-tabpanel-track:one\"\n            aria-selected=\"true\"\n            class=\"ui-tab\"\n            data-collection=\"react-aria-_r_31m_\"\n            data-key=\"track:one\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            data-selected=\"true\"\n            id=\"react-aria-_r_31n_-tab-track:one\"\n            role=\"tab\"\n            tabindex=\"0\"\n          >\n            French · Default\n          </div>\n          <div\n            aria-selected=\"false\"\n            class=\"ui-tab\"\n            data-collection=\"react-aria-_r_31m_\"\n            data-key=\"lang:en\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            id=\"react-aria-_r_31n_-tab-lang:en\"\n            role=\"tab\"\n            tabindex=\"-1\"\n          >\n            English\n          </div>\n        </div>\n        <div\n          aria-labelledby=\"react-aria-_r_31n_-tab-track:one\"\n          class=\"lesson-language-panel ui-tab-panel\"\n          data-rac=\"\"\n          id=\"react-aria-_r_31n_-tabpanel-track:one\"\n          role=\"tabpanel\"\n        >\n          <section\n            aria-busy=\"false\"\n            aria-labelledby=\"lesson-draft-lesson-track:one\"\n            class=\"admin-draft-field\"\n          >\n            <h3\n              id=\"lesson-draft-lesson-track:one\"\n            >\n              Speech recognition\n            </h3>\n            <div\n              class=\"admin-draft-regenerate\"\n            >\n              <button\n                class=\"ui-button ui-button--secondary\"\n                data-rac=\"\"\n                data-react-aria-pressable=\"true\"\n                id=\"react-aria-_r_31q_\"\n                tabindex=\"0\"\n                type=\"button\"\n              >\n                Generate transcript\n              </button>\n            </div>\n          </section>\n          <section\n            aria-labelledby=\"transcript-editor-lesson-track:one\"\n            class=\"transcript-editor\"\n          >\n            <h4\n              id=\"transcript-editor-lesson-track:one\"\n            >\n              Lesson transcript\n            </h4>\n            <p>\n              No lines yet.\n            </p>\n            <ol\n              class=\"transcript-editor-lines\"\n            />\n            <button\n              aria-disabled=\"false\"\n              class=\"ui-button ui-button--secondary\"\n              data-add-line=\"true\"\n              data-rac=\"\"\n              data-react-aria-pressable=\"true\"\n              id=\"react-aria-_r_32d_\"\n              tabindex=\"0\"\n              type=\"button\"\n            >\n              Add a line\n            </button>\n            <button\n              aria-disabled=\"false\"\n              class=\"ui-button ui-button--primary\"\n              data-rac=\"\"\n              data-react-aria-pressable=\"true\"\n              id=\"react-aria-_r_32f_\"\n              tabindex=\"0\"\n              type=\"button\"\n            >\n              Save the transcript\n            </button>\n            <output>\n              Transcript saved.\n            </output>\n          </section>\n        </div>\n        <div\n          class=\"lesson-language-panel ui-tab-panel\"\n          data-inert=\"true\"\n          data-rac=\"\"\n          inert=\"\"\n        >\n          <section\n            aria-labelledby=\"transcript-editor-lesson-lang:en\"\n            class=\"transcr
```
NEXT: Verify exact B130 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B130 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 0; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B130.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B130 mutation failure excerpt (verbatim):
```text


No test failures reported.

```
FACT: Unmutated baseline `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B130.log.
NEXT: Verify exact B130 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B130 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B130.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B130 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: orphan has transcript editor but no ASR controls even when draft service exists

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'orphan has transcript editor but no ASR controls even when draft service exists'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "orphan has transcript editor but no ASR controls even when draft service exists",
  "status": "fail",
  "duration": 52,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <h3 …(1)></h3> to be null",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1294,
        "column": 73,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1294
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
<h3
  id="lesson-draft-lesson-lang:fr"
>
  Speech recognition
</h3>
```

codeFrame (error 1):

```text
1292 |   expect(screen.getByRole('heading', { name: 'Lesson transcript' })).toBeTruthy();
1293 |   expect(screen.queryByRole('button', { name: 'Regenerate' })).toBeNull();
1294 |   expect(screen.queryByRole('heading', { name: 'Speech recognition' })).toBeNull();
     |                                                                         ^
1295 |   expect(screen.queryByRole('button', { name: 'Generate transcript' })).toBeNull();
1296 | });
```


```
NEXT: Verify exact B133 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B133 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B133.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B133 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: loaded ASR lines save with an explicit empty document list

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'loaded ASR lines save with an explicit empty document list'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "loaded ASR lines save with an explicit empty document list",
  "status": "fail",
  "duration": 71,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { …(5) } ] to deeply equal [ { id: Any<String>, …(4) } ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1303,
        "column": 32,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1303
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    {
-     "documentIds": [],
+     "documentIds": [
+       "foreign",
+     ],
      "endMs": 1000,
-     "id": Any<String>,
+     "id": "ef37abf9-3d74-4c86-8754-a1a32c37ccd8",
      "startMs": 0,
      "text": "Original",
    },
  ]
```

codeFrame (error 1):

```text
1301 |   fireEvent.click(screen.getByRole('button', { name: 'Rename track and save' }));
1302 |   await screen.findByText('Transcript saved.');
1303 |   expect(f.state.puts[0]?.[3]).toEqual([{ id: expect.any(String), startMs: 0, endMs: 1000, text: 'Original', documentIds: [] }]);
     |                                ^
1304 | });
1305 | 
```


```
NEXT: Verify exact B136 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B136 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 0; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B136.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B136 mutation failure excerpt (verbatim):
```text


No test failures reported.

```
FACT: Unmutated baseline `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx` exit 1; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B136.log.
FACT: Unmutated baseline `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B136.log.
NEXT: Verify exact B136 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B136 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B136.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B136 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: manual recovery captures the current draft token after a draft-only refresh

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'manual recovery captures the current draft token after a draft-only refresh'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "manual recovery captures the current draft token after a draft-only refresh",
  "status": "fail",
  "duration": 44,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { Object (key, capturedAt, ...) } to match object { draftUpdatedAt: 'D1' }\n(7 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1310,
        "column": 64,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1310
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "draftUpdatedAt": "D1",
+   "baseLanguage": "en",
+   "capturedAt": 1794929860860,
+   "expectedUpdatedAt": "T1",
+   "key": "author:lesson:track:one",
+   "language": "en",
+   "lines": [
+     {
+       "documentIds": [],
+       "endMs": "1000",
+       "id": "source-id",
+       "startMs": "0",
+       "text": "My edit",
+     },
+   ],
+   "revision": "936c37f3-6c6b-4a24-a2b8-602545ab973a",
  }
```

codeFrame (error 1):

```text
1308 |   const view = await mount(f); edit();
1309 |   fireEvent.focusOut(screen.getByRole('textbox', { name: 'Line 1 text' }));
1310 |   expect(await f.historyStore.read('author:lesson:track:one')).toMatchObject({ draftUpdatedAt: 'D1' });
     |                                                                ^
1311 |   f.state.draft = { ...f.state.draft, updatedAt: 'D2' }; view.refresh();
1312 |   fireEvent.change(screen.getByRole('textbox', { name: 'Line 1 text' }), { target: { value: 'Second edit' } });
```


```
NEXT: Verify exact B137 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B137 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B137.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B137 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: handled draft offer never reappears on language editor rerender

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'handled draft offer never reappears on language editor rerender'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "handled draft offer never reappears on language editor rerender",
  "status": "fail",
  "duration": 46,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <button …(7)></button> to be null",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1323,
        "column": 68,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1323
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
  data-rac=""
  data-react-aria-pressable="true"
  id="react-aria-_r_38i_"
  tabindex="0"
  type="button"
>
  Load the draft
</button>
```

codeFrame (error 1):

```text
1321 |   fireEvent.click(screen.getByRole('button', { name: 'Skip' }));
1322 |   view.refresh();
1323 |   expect(screen.queryByRole('button', { name: 'Load the draft' })).toBeNull();
     |                                                                    ^
1324 | });
1325 | 
```


```
NEXT: Verify exact B165 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B165 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B165.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B165 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: orphan languages sort after tracks and uppercase track language never duplicates an orphan

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'orphan languages sort after tracks and uppercase track language never duplicates an orphan'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "orphan languages sort after tracks and uppercase track language never duplicates an orphan",
  "status": "fail",
  "duration": 23,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'track:one', 'lang:de', …(3) ] to deeply equal [ 'track:one', 'lang:de', …(2) ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1165,
        "column": 6,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1165
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    "track:one",
    "lang:de",
+   "lang:en",
    "lang:fr",
    "lang:uk",
  ]
```

codeFrame (error 1):

```text
1163 | test('orphan languages sort after tracks and uppercase track language never duplicates an orphan', () => {
1164 |   expect(lessonLanguageTabs([track('one', 'EN')], [item('uk'), item('en'), item('de'), item('fr')]).map((tab) => tab.id))
1165 |     .toEqual(['track:one', 'lang:de', 'lang:fr', 'lang:uk']);
     |      ^
1166 | });
1167 | 
```


```
NEXT: Verify exact B168 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B168 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B168.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B168 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: only processing videos and missing optional audio both retain the single editor

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'only processing videos and missing optional audio both retain the single editor'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "only processing videos and missing optional audio both retain the single editor",
  "status": "fail",
  "duration": 101,
  "retryCount": 0,
  "errors": [
    {
      "type": "TypeError",
      "message": "Cannot read properties of undefined (reading 'useTracks')",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.tsx",
        "line": 141,
        "column": 24,
        "method": "ReadyLanguageTabs"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.tsx",
      "line": 141
    }
  ]
}
```

codeFrame (error 1):

```text
139 | }) {
140 |   const { t } = useTranslation();
141 |   const tracks = audio.useTracks(videoId);
    |                        ^
142 |   const inventory = languages.useInventory(props.lessonId);
143 |   const { lessonId, onTranscriptKnown } = props;
```


```
NEXT: Verify exact B169 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B169 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B169.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B169 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: single editor known state comes from complete inventory: loading, empty, and other language

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'single editor known state comes from complete inventory: loading, empty, and other language'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "single editor known state comes from complete inventory: loading, empty, and other language",
  "status": "fail",
  "duration": 37,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [] to deeply equal [ false ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1205,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1205
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- [
-   false,
- ]
+ []
```

codeFrame (error 1):

```text
1203 |   f.languages.useInventory = () => ({ data: [] });
1204 |   view.refresh();
1205 |   expect(f.state.known).toEqual([false]);
     |                         ^
1206 |   f.languages.useInventory = () => ({ data: [item('fr')] });
1207 |   view.refresh();
```


```
NEXT: Verify exact B170 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B170 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B170.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B170 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: detached copy of orphan language is offered only to a track, preserving mounted orphan ownership

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'detached copy of orphan language is offered only to a track, preserving mounted orphan ownership'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "detached copy of orphan language is offered only to a track, preserving mounted orphan ownership",
  "status": "fail",
  "duration": 1074,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find role=\"heading\" and name \"Unsaved lines from another tab (language: uk)\"\n\nIgnored nodes: comments, script, style\n<body>\n  <div>\n    <form\n      data-testid=\"course-form\"\n    >\n      <template />\n      <div\n        class=\"ui-tabs\"\n        data-orientation=\"horizontal\"\n        data-rac=\"\"\n      >\n        <div\n          aria-label=\"Transcript languages\"\n          aria-orientation=\"horizontal\"\n          class=\"ui-tab-list\"\n          data-collection=\"react-aria-_r_3ak_\"\n          data-orientation=\"horizontal\"\n          data-rac=\"\"\n          id=\"react-aria-_r_3al_\"\n          role=\"tablist\"\n        >\n          <div\n            aria-controls=\"react-aria-_r_3al_-tabpanel-track:one\"\n            aria-selected=\"true\"\n            class=\"ui-tab\"\n            data-collection=\"react-aria-_r_3ak_\"\n            data-key=\"track:one\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            data-selected=\"true\"\n            id=\"react-aria-_r_3al_-tab-track:one\"\n            role=\"tab\"\n            tabindex=\"0\"\n          >\n            English · Default\n          </div>\n          <div\n            aria-selected=\"false\"\n            class=\"ui-tab\"\n            data-collection=\"react-aria-_r_3ak_\"\n            data-key=\"lang:uk\"\n            data-rac=\"\"\n            data-react-aria-pressable=\"true\"\n            id=\"react-aria-_r_3al_-tab-lang:uk\"\n            role=\"tab\"\n            tabindex=\"-1\"\n          >\n            Ukrainian\n          </div>\n        </div>\n        <div\n          aria-labelledby=\"react-aria-_r_3al_-tab-track:one\"\n          class=\"lesson-language-panel ui-tab-panel\"\n          data-rac=\"\"\n          id=\"react-aria-_r_3al_-tabpanel-track:one\"\n          role=\"tabpanel\"\n        >\n          <section\n            aria-busy=\"false\"\n            aria-labelledby=\"lesson-draft-lesson-track:one\"\n            class=\"admin-draft-field\"\n          >\n            <h3\n              id=\"lesson-draft-lesson-track:one\"\n            >\n              Speech recognition\n            </h3>\n            <div\n              class=\"admin-draft-regenerate\"\n            >\n              <button\n                class=\"ui-button ui-button--secondary\"\n                data-rac=\"\"\n                data-react-aria-pressable=\"true\"\n                id=\"react-aria-_r_3ao_\"\n                tabindex=\"0\"\n                type=\"button\"\n              >\n                Generate transcript\n              </button>\n            </div>\n          </section>\n          <section\n            aria-labelledby=\"transcript-editor-lesson-track:one\"\n            class=\"transcript-editor\"\n          >\n            <h4\n              id=\"transcript-editor-lesson-track:one\"\n            >\n              Lesson transcript\n            </h4>\n            <ol\n              class=\"transcript-editor-lines\"\n            >\n              <li>\n                <div\n                  class=\"ui-field\"\n                  data-rac=\"\"\n                >\n                  <label\n                    class=\"react-aria-Label\"\n                    for=\"react-aria-_r_3aq_\"\n                    id=\"react-aria-_r_3ar_\"\n                  >\n                    Line 1 start, ms\n                  </label>\n                  <input\n                    aria-labelledby=\"react-aria-_r_3ar_\"\n                    class=\"ui-input\"\n                    data-rac=\"\"\n                    id=\"react-aria-_r_3aq_\"\n                    tabindex=\"0\"\n                    title=\"\"\n                    type=\"text\"\n                    value=\"0\"\n                  />\n                </div>\n                <div\n                  class=\"ui-field\"\n                  data-rac=\"\"\n                >\n                  <label\n                    class=\"react-aria-Label\"\n                    for=\"react-aria-_r_3av_\"\n                    id=\"react-aria-_r_3b0_\"\n                  >\n                    Line 1 end, ms\n                  </label>\n                  <input\n                    aria-labelledby=\"react-aria-_r_3b0_\"\n                    class=\"ui-input\"\n 
```
NEXT: Verify exact B179 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B179 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B179.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B179 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: orphan replacement confirmation never performs a track read

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'orphan replacement confirmation never performs a track read'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "orphan replacement confirmation never performs a track read",
  "status": "fail",
  "duration": 66,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'inventory', 'video' ] to deeply equal [ 'inventory' ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1334,
        "column": 25,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1334
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    "inventory",
+   "video",
  ]
```

codeFrame (error 1):

```text
1332 |   fireEvent.click(screen.getByRole('button', { name: 'Replace' }));
1333 |   await screen.findByText('Transcript saved.');
1334 |   expect(f.state.reads).toEqual(['inventory']);
     |                         ^
1335 |   expect(f.state.puts[0]?.slice(0, 3)).toEqual(['lesson', 'fr', 'TARGET']);
1336 | });
```


```
NEXT: Verify exact B187 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B187 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B187.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B187 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: M3-R4 options use fresh track/inventory reads, excluding und and taken languages case-insensitively; no invalid preselection

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'M3-R4 options use fresh track/inventory reads, excluding und and taken languages case-insensitively; no invalid preselection'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "M3-R4 options use fresh track/inventory reads, excluding und and taken languages case-insensitively; no invalid preselection",
  "status": "fail",
  "duration": 99,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <p role=\"alert\"></p> to be null",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 520,
        "column": 47,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 520
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
<p
  role="alert"
>
  Choose a language before saving.
</p>
```

codeFrame (error 1):

```text
518 |   ]);
519 |   await user.click(screen.getByRole('option', { name: 'Spanish' }));
520 |   expect(within(dialog).queryByRole('alert')).toBeNull();
    |                                               ^
521 |   await user.click(confirm);
522 |   await waitFor(() => expect(f.state.puts).toHaveLength(1));
```


```
FACT: Unmutated baseline `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B100.log.
NEXT: Verify exact B100 mutation against src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx.
FACT: B100 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B100.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B100 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx :: unmount cancels pending und Save promise without a write

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx' --testNamePattern 'unmount cancels pending und Save promise without a write'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
  "project": "dom",
  "fullName": "unmount cancels pending und Save promise without a write",
  "status": "fail",
  "duration": 93,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 'pending' to be 'language_cancelled' // Object.is equality",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
        "line": 41,
        "column": 20,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
      "line": 41
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- language_cancelled
+ pending
```

codeFrame (error 1):

```text
39 |     void probe.completion?.then((value) => { result = value; });
40 |     await act(async () => { view.unmount(); });
41 |     expect(result).toBe(mode === 'und' ? 'language_cancelled' : 'replacement_cancelled');
   |                    ^
42 |     expect(writes).toEqual([]);
43 |   });
```


```
NEXT: Verify exact B103 mutation against src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx.
FACT: B103 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B103.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B103 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx :: unmount cancels pending replacement Save promise without a write

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx' --testNamePattern 'unmount cancels pending replacement Save promise without a write'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
  "project": "dom",
  "fullName": "unmount cancels pending replacement Save promise without a write",
  "status": "fail",
  "duration": 40,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 'pending' to be 'replacement_cancelled' // Object.is equality",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
        "line": 41,
        "column": 20,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
      "line": 41
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- replacement_cancelled
+ pending
```

codeFrame (error 1):

```text
39 |     void probe.completion?.then((value) => { result = value; });
40 |     await act(async () => { view.unmount(); });
41 |     expect(result).toBe(mode === 'und' ? 'language_cancelled' : 'replacement_cancelled');
   |                    ^
42 |     expect(writes).toEqual([]);
43 |   });
```


```
FACT: Unmutated baseline `cd frontend && npx rstest run src/modules/admin/ui/CourseEditorView.test.tsx` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B144.log.
NEXT: Verify exact B144 mutation against src/modules/admin/ui/CourseEditorView.test.tsx.
FACT: B144 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/CourseEditorView.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B144.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/CourseEditorView.tsx` exit 0.
FACT: B144 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/CourseEditorView.test.tsx :: new unsaved lesson never invokes transcript, inventory or draft service hooks

repro:

```bash
npx rstest 'src/modules/admin/ui/CourseEditorView.test.tsx' --testNamePattern 'new unsaved lesson never invokes transcript, inventory or draft service hooks'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/CourseEditorView.test.tsx",
  "project": "dom",
  "fullName": "new unsaved lesson never invokes transcript, inventory or draft service hooks",
  "status": "fail",
  "duration": 113,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ …(12) ] to deeply equal []",
      "topFrame": {
        "file": "src/modules/admin/ui/CourseEditorView.test.tsx",
        "line": 999,
        "column": 17,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/CourseEditorView.test.tsx",
      "line": 999
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
+   "inventory:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "draft:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "draft:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "transcript:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "inventory:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "draft:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "draft:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "transcript:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "inventory:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "draft:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "draft:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+   "transcript:abcca47f-d10b-4fba-9953-e06eb355ab7f",
+ ]
```

codeFrame (error 1):

```text
 997 |   await user.click(screen.getByRole('button', { name: 'Add lesson to module 1' }));
 998 |   expect(screen.getByRole('textbox', { name: 'Lesson 1 title in module 1' })).toBeTruthy();
 999 |   expect(reads).toEqual([]);
     |                 ^
1000 | });
1001 | 
```

console:

```text
[stderr] error: AuthoringQueryError: dependency_unavailable
    at unwrap (/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend/src/app/composition/admin-queries.ts:32:25)
    at queryFn (/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend/src/app/composition/admin-queries.ts:97:40)
    at processTicksAndRejections (node:internal/process/task_queues:103:5) {
  [stack]: ... [truncated]
[stderr] warn: Warning: Error in route match: /admin/courses/$courseId/admin/courses/10000000-0000-4000-8000-000000000001
[stderr] error: AuthoringQueryError: dependency_unavailable
    at unwrap (/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend/src/app/composition/admin-queries.ts:32:25)
    at queryFn (/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend/src/app/composition/admin-queries.ts:97:40)
    at processTicksAndRejections (node:internal/process/task_queues:103:5) {
  [stack]: ... [truncated]
[stderr] warn: Warning: Error in route match: /admin/courses/$courseId/admin/courses/10000000-0000-4000-8000-000000000001
[stderr] error: AuthoringQueryError: dependency_unavailable
    at unwrap (/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend/src/app/composition/admin-queries.ts:32:25)
    at queryFn (/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend/src/app/composition/admin-queries.ts:97:40)
    at processTicksAndRejections (node:internal/process/task_queues:103:5) {
  [stack]: ... [truncated]
[stderr] warn: Warning: Error in route match: /admin/courses/$courseId/admin/courses/10000000-0000-4000-8000-000000000001
```


```
FACT: Draft baseline `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts` exit 0, 21 passed / 0 failed / 0 skipped. Audio/media baseline `cd frontend && npx rstest run src/app/composition/admin-audio-queries.test.ts src/modules/media/adapters/audio-track-repository.test.ts src/modules/media/ui/LessonAudioTracksField.test.tsx` exit 0, 65 passed / 0 failed / 0 skipped.
FACT: Preserved baseline raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-draft-baseline.log.
FACT: Preserved baseline raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-media-baseline.log.
DECISION: The suggested pending-upload switch fixture cannot reach its target: upload progress unmounts the new-track form (LessonAudioTracksField.tsx:251). Deferred row rename keeps that form mounted with isBusy=true and directly discriminates B061/B063. A speculative live reclassification of an already-read recovery offer was discarded: recovery is read per key/store visit (use-transcript-recovery.ts:99–133); B136 instead verifies current draft token on author capture, before and after a token-only refresh. No production defect claimed from an invented live-offer refresh expectation.
FACT: Unmutated baseline `cd frontend && npx rstest run src/routes/route-guards.test.tsx` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B149.log.
NEXT: Verify exact B149 mutation against src/routes/route-guards.test.tsx.
FACT: B149 exact source patch from mutation journal; command `cd frontend && npx rstest run src/routes/route-guards.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B149.log. Production restore `git diff --quiet -- frontend/src/routes/admin.courses.$courseId.tsx` exit 0.
FACT: B149 mutation failure excerpt (verbatim):
```text


### [F01] src/routes/route-guards.test.tsx :: guarded editor forwards multilingual server default and language-specific transcript reads

repro:

```bash
npx rstest 'src/routes/route-guards.test.tsx' --testNamePattern 'guarded editor forwards multilingual server default and language-specific transcript reads'
```

details:

```json
{
  "testPath": "src/routes/route-guards.test.tsx",
  "project": "dom",
  "fullName": "guarded editor forwards multilingual server default and language-specific transcript reads",
  "status": "fail",
  "duration": 225,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <input …(6)></input> to have property \"checked\" with value true",
      "topFrame": {
        "file": "src/routes/route-guards.test.tsx",
        "line": 1404,
        "column": 70,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/routes/route-guards.test.tsx",
      "line": 1404
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
1402 | test('guarded editor forwards multilingual server default and language-specific transcript reads', async () => {
1403 |   const f = await multilingualEditor();
1404 |   expect(screen.getByRole('switch', { name: 'Multiple languages' })).toHaveProperty('checked', true);
     |                                                                      ^
1405 |   expect(f.state.requested).toContain('en');
1406 |   expect(f.state.requested).toContain('en-gb');
```

console:

```text
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
```


```
NEXT: Verify exact B150 mutation against src/routes/route-guards.test.tsx.
FACT: B150 exact source patch from mutation journal; command `cd frontend && npx rstest run src/routes/route-guards.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B150.log. Production restore `git diff --quiet -- frontend/src/routes/admin.courses.$courseId.tsx` exit 0.
FACT: B150 mutation failure excerpt (verbatim):
```text


### [F01] src/routes/route-guards.test.tsx :: guarded editor forwards multilingual server default and language-specific transcript reads

repro:

```bash
npx rstest 'src/routes/route-guards.test.tsx' --testNamePattern 'guarded editor forwards multilingual server default and language-specific transcript reads'
```

details:

```json
{
  "testPath": "src/routes/route-guards.test.tsx",
  "project": "dom",
  "fullName": "guarded editor forwards multilingual server default and language-specific transcript reads",
  "status": "fail",
  "duration": 1174,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Found multiple elements with the display value: English lines.\n\nHere are the matching elements:\n\nIgnored nodes: comments, script, style\n<textarea\n  aria-labelledby=\"react-aria-_r_u7_\"\n  class=\"ui-input ui-textarea\"\n  data-rac=\"\"\n  id=\"react-aria-_r_u6_\"\n  tabindex=\"0\"\n  title=\"\"\n>\n  English lines\n</textarea>\n\nIgnored nodes: comments, script, style\n<textarea\n  aria-labelledby=\"react-aria-_r_uv_\"\n  class=\"ui-input ui-textarea\"\n  data-rac=\"\"\n  id=\"react-aria-_r_uu_\"\n  tabindex=\"0\"\n  title=\"\"\n>\n  English lines\n</textarea>\n\n(If this is intentional, then use the `*AllBy*` variant of the query (like `queryAllByText`, `getAllByText`, or `findAllByText`)).\n\nIgnored nodes: comments, script, style\n<body\n  data-surface=\"admin\"\n>\n  <div>\n    <div\n      class=\"app-shell\"\n      data-surface=\"admin\"\n    >\n      <a\n        class=\"skip-link\"\n        href=\"#main\"\n      >\n        Skip to main content\n      </a>\n      <div\n        class=\"admin-shell\"\n      >\n        <aside\n          class=\"admin-sidebar\"\n        >\n          <div\n            class=\"admin-workspace\"\n          >\n            <span\n              class=\"admin-workspace-name\"\n            >\n              CourseForge\n            </span>\n            <span\n              class=\"admin-workspace-area\"\n            >\n              Admin\n            </span>\n          </div>\n          <nav\n            aria-label=\"Admin\"\n            class=\"admin-nav\"\n          >\n            <div\n              class=\"admin-nav-section\"\n            >\n              <a\n                class=\"admin-nav-link\"\n                href=\"/admin\"\n              >\n                <svg\n                  aria-hidden=\"true\"\n                  class=\"admin-nav-icon\"\n                  fill=\"none\"\n                  height=\"16\"\n                  stroke=\"currentColor\"\n                  stroke-linecap=\"round\"\n                  stroke-linejoin=\"round\"\n                  stroke-width=\"1.5\"\n                  viewBox=\"0 0 16 16\"\n                  width=\"16\"\n                >\n                  <rect\n                    height=\"5\"\n                    rx=\"1\"\n                    width=\"5\"\n                    x=\"2\"\n                    y=\"2\"\n                  />\n                  <rect\n                    height=\"5\"\n                    rx=\"1\"\n                    width=\"5\"\n                    x=\"9\"\n                    y=\"2\"\n                  />\n                  <rect\n                    height=\"5\"\n                    rx=\"1\"\n                    width=\"5\"\n                    x=\"2\"\n                    y=\"9\"\n                  />\n                  <rect\n                    height=\"5\"\n                    rx=\"1\"\n                    width=\"5\"\n                    x=\"9\"\n                    y=\"9\"\n                  />\n                </svg>\n                <span>\n                  Courses\n                </span>\n              </a>\n              <a\n                class=\"admin-nav-link\"\n                href=\"/admin/analytics\"\n              >\n                <svg\n                  aria-hidden=\"true\"\n                  class=\"admin-nav-icon\"\n                  fill=\"none\"\n                  height=\"16\"\n                  stroke=\"currentColor\"\n                  stroke-linecap=\"round\"\n                  stroke-linejoin=\"round\"\n                  stroke-width=\"1.5\"\n                  viewBox=\"0 0 16 16\"\n                  width=\"16\"\n                >\n                  <path\n                    d=\"M2.5 13.5h11\"\n                  />\n                  <rect\n                    height=\"4\"\n                    rx=\"0.5\"\n                    width=\"2.2\"\n                    x=\"3.5\"\n                    y=\"8.5\"\n                  />\n                  <rect\n                    height=\"7\"\n                    rx=\"0.5\"\n                    width=\"2.2\"\n                    x=\"7\"\n                    y=\"5.5\"\n                  />\n                  <rect\n                    height=\"3\"\n                    rx=\"0.5\"\n                    width=\"2.2\"\n                    x=\"10.5\"\n                    y=\"9.5\"\n             
```
NEXT: Verify exact B151 mutation against src/routes/route-guards.test.tsx.
FACT: B151 exact source patch from mutation journal; command `cd frontend && npx rstest run src/routes/route-guards.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B151.log. Production restore `git diff --quiet -- frontend/src/routes/admin.courses.$courseId.tsx` exit 0.
FACT: B151 mutation failure excerpt (verbatim):
```text


### [F01] src/routes/route-guards.test.tsx :: guarded rename retains previous lines while new language response waits, dirty=false

repro:

```bash
npx rstest 'src/routes/route-guards.test.tsx' --testNamePattern 'guarded rename retains previous lines while new language response waits, dirty=false'
```

details:

```json
{
  "testPath": "src/routes/route-guards.test.tsx",
  "project": "dom",
  "fullName": "guarded rename retains previous lines while new language response waits, dirty=false",
  "status": "fail",
  "duration": 232,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an accessible element with the role \"textbox\" and name \"Line 1 text\"\n\nHere are the accessible roles:\n\n  link:\n\n  Name \"Skip to main content\":\n  <a\n    class=\"skip-link\"\n    href=\"#main\"\n  />\n\n  Name \"Courses\":\n  <a\n    class=\"admin-nav-link\"\n    href=\"/admin\"\n  />\n\n  Name \"Analytics\":\n  <a\n    class=\"admin-nav-link\"\n    href=\"/admin/analytics\"\n  />\n\n  Name \"Learners\":\n  <a\n    class=\"admin-nav-link\"\n    href=\"/admin/learners\"\n  />\n\n  Name \"Users\":\n  <a\n    class=\"admin-nav-link\"\n    href=\"/admin/users\"\n  />\n\n  Name \"Participant reports\":\n  <a\n    class=\"admin-nav-link\"\n    href=\"/admin/participant-reports\"\n  />\n\n  Name \"Refunds\":\n  <a\n    class=\"admin-nav-link\"\n    href=\"/admin/refunds\"\n  />\n\n  Name \"To client\":\n  <a\n    class=\"admin-to-client\"\n    href=\"/\"\n  />\n\n  Name \"← All courses\":\n  <a\n    class=\"ui-link \"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    href=\"/admin\"\n    tabindex=\"0\"\n  />\n\n  --------------------------------------------------\n  complementary:\n\n  Name \"\":\n  <aside\n    class=\"admin-sidebar\"\n  />\n\n  --------------------------------------------------\n  navigation:\n\n  Name \"Admin\":\n  <nav\n    aria-label=\"Admin\"\n    class=\"admin-nav\"\n  />\n\n  --------------------------------------------------\n  banner:\n\n  Name \"\":\n  <header\n    class=\"admin-topbar\"\n  />\n\n  Name \"\":\n  <header\n    class=\"admin-page-header\"\n  />\n\n  --------------------------------------------------\n  button:\n\n  Name \"Main navigation\":\n  <button\n    aria-expanded=\"false\"\n    aria-label=\"Main navigation\"\n    class=\"ui-button ui-button--secondary admin-nav-toggle\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_vc_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"UA\":\n  <button\n    aria-pressed=\"false\"\n    lang=\"uk\"\n    type=\"button\"\n  />\n\n  Name \"EN\":\n  <button\n    aria-pressed=\"true\"\n    lang=\"en\"\n    type=\"button\"\n  />\n\n  Name \"Sign out\":\n  <button\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_vg_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Publish\":\n  <button\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_10a_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Delete draft\":\n  <button\n    aria-expanded=\"false\"\n    class=\"ui-button ui-button--danger\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_10d_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Add a material\":\n  <button\n    class=\"ui-button ui-button--primary\"\n    data-disabled=\"true\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    disabled=\"\"\n    id=\"react-aria-_r_10r_\"\n    type=\"button\"\n  />\n\n  Name \"Replace video\":\n  <button\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_13h_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Edit track Original (French)\":\n  <button\n    aria-expanded=\"false\"\n    aria-label=\"Edit track Original (French)\"\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_13o_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Delete track Original (French)\":\n  <button\n    aria-expanded=\"false\"\n    aria-label=\"Delete track Original (French)\"\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_13s_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Make British (en-gb) the default track\":\n  <button\n    aria-label=\"Make British (en-gb) the default track\"\n    class=\"ui-button ui-button--secondary\"\n    data-rac=\"\"\n    data-react-aria-pressable=\"true\"\n    id=\"react-aria-_r_142_\"\n    tabindex=\"0\"\n    type=\"button\"\n  />\n\n  Name \"Edit track British (en-gb)\":\n  <button\n    aria-expanded=\"false\"\n    aria-label=\"Edit track British (en-gb)\"\n    class=\"ui-but
```
NEXT: Verify exact B154 mutation against src/routes/route-guards.test.tsx.
FACT: B154 exact source patch from mutation journal; command `cd frontend && npx rstest run src/routes/route-guards.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B154.log. Production restore `git diff --quiet -- frontend/src/routes/admin.courses.$courseId.tsx` exit 0.
FACT: B154 mutation failure excerpt (verbatim):
```text


### [F01] src/routes/route-guards.test.tsx :: guarded fresh inventory error prevents PUT and retains edited lines

repro:

```bash
npx rstest 'src/routes/route-guards.test.tsx' --testNamePattern 'guarded fresh inventory error prevents PUT and retains edited lines'
```

details:

```json
{
  "testPath": "src/routes/route-guards.test.tsx",
  "project": "dom",
  "fullName": "guarded fresh inventory error prevents PUT and retains edited lines",
  "status": "fail",
  "duration": 1182,
  "retryCount": 0,
  "errors": [
    {
      "type": "TestingLibraryElementError",
      "message": "Unable to find an element with the text: Could not verify the track — try again.. This could be because the text is broken up by multiple elements. In this case, you can provide a function for your text matcher to make your matcher more flexible.\n\nIgnored nodes: comments, script, style\n<body\n  data-surface=\"admin\"\n>\n  <div>\n    <div\n      class=\"app-shell\"\n      data-surface=\"admin\"\n    >\n      <a\n        class=\"skip-link\"\n        href=\"#main\"\n      >\n        Skip to main content\n      </a>\n      <div\n        class=\"admin-shell\"\n      >\n        <aside\n          class=\"admin-sidebar\"\n        >\n          <div\n            class=\"admin-workspace\"\n          >\n            <span\n              class=\"admin-workspace-name\"\n            >\n              CourseForge\n            </span>\n            <span\n              class=\"admin-workspace-area\"\n            >\n              Admin\n            </span>\n          </div>\n          <nav\n            aria-label=\"Admin\"\n            class=\"admin-nav\"\n          >\n            <div\n              class=\"admin-nav-section\"\n            >\n              <a\n                class=\"admin-nav-link\"\n                href=\"/admin\"\n              >\n                <svg\n                  aria-hidden=\"true\"\n                  class=\"admin-nav-icon\"\n                  fill=\"none\"\n                  height=\"16\"\n                  stroke=\"currentColor\"\n                  stroke-linecap=\"round\"\n                  stroke-linejoin=\"round\"\n                  stroke-width=\"1.5\"\n                  viewBox=\"0 0 16 16\"\n                  width=\"16\"\n                >\n                  <rect\n                    height=\"5\"\n                    rx=\"1\"\n                    width=\"5\"\n                    x=\"2\"\n                    y=\"2\"\n                  />\n                  <rect\n                    height=\"5\"\n                    rx=\"1\"\n                    width=\"5\"\n                    x=\"9\"\n                    y=\"2\"\n                  />\n                  <rect\n                    height=\"5\"\n                    rx=\"1\"\n                    width=\"5\"\n                    x=\"2\"\n                    y=\"9\"\n                  />\n                  <rect\n                    height=\"5\"\n                    rx=\"1\"\n                    width=\"5\"\n                    x=\"9\"\n                    y=\"9\"\n                  />\n                </svg>\n                <span>\n                  Courses\n                </span>\n              </a>\n              <a\n                class=\"admin-nav-link\"\n                href=\"/admin/analytics\"\n              >\n                <svg\n                  aria-hidden=\"true\"\n                  class=\"admin-nav-icon\"\n                  fill=\"none\"\n                  height=\"16\"\n                  stroke=\"currentColor\"\n                  stroke-linecap=\"round\"\n                  stroke-linejoin=\"round\"\n                  stroke-width=\"1.5\"\n                  viewBox=\"0 0 16 16\"\n                  width=\"16\"\n                >\n                  <path\n                    d=\"M2.5 13.5h11\"\n                  />\n                  <rect\n                    height=\"4\"\n                    rx=\"0.5\"\n                    width=\"2.2\"\n                    x=\"3.5\"\n                    y=\"8.5\"\n                  />\n                  <rect\n                    height=\"7\"\n                    rx=\"0.5\"\n                    width=\"2.2\"\n                    x=\"7\"\n                    y=\"5.5\"\n                  />\n                  <rect\n                    height=\"3\"\n                    rx=\"0.5\"\n                    width=\"2.2\"\n                    x=\"10.5\"\n                    y=\"9.5\"\n                  />\n                </svg>\n                <span>\n                  Analytics\n                </span>\n              </a>\n            </div>\n            <div\n              class=\"admin-nav-section\"\n            >\n              <span\n                class=\"admin-nav-section-label\"\n              >\n                People\n              </span>\n              <a\n                class=\"admin-nav-link\"\n                href=\"/admin/learners\"\n              >\n                <svg\n                  aria-hid
```
NEXT: Verify exact B157 mutation against src/routes/route-guards.test.tsx.
FACT: B157 exact source patch from mutation journal; command `cd frontend && npx rstest run src/routes/route-guards.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B157.log. Production restore `git diff --quiet -- frontend/src/routes/admin.courses.$courseId.tsx` exit 0.
FACT: B157 mutation failure excerpt (verbatim):
```text


### [F01] src/routes/route-guards.test.tsx :: guarded Save refreshes inventory and every lesson language, isolates sibling and uses latest token

repro:

```bash
npx rstest 'src/routes/route-guards.test.tsx' --testNamePattern 'guarded Save refreshes inventory and every lesson language, isolates sibling and uses latest token'
```

details:

```json
{
  "testPath": "src/routes/route-guards.test.tsx",
  "project": "dom",
  "fullName": "guarded Save refreshes inventory and every lesson language, isolates sibling and uses latest token",
  "status": "fail",
  "duration": 185,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected false to be true // Object.is equality",
      "topFrame": {
        "file": "src/routes/route-guards.test.tsx",
        "line": 1444,
        "column": 85,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/routes/route-guards.test.tsx",
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
1442 |   await screen.findByText('Transcript saved.');
1443 |   expect(f.state.inventoryReads).toBeGreaterThan(before);
1444 |   expect(client.getQueryState(lessonTranscriptQueryKey(f.lessonId))?.isInvalidated).toBe(true);
     |                                                                                     ^
1445 |   expect(client.getQueryData(lessonTranscriptQueryKey(f.lessonId, 'en'))).toMatchObject({ updatedAt: 'SAVED2' });
1446 |   expect(client.getQueryData(lessonTranscriptQueryKey(f.lessonId, 'en-gb'))).toMatchObject({ updatedAt: 'GB1' });
```

console:

```text
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
```


```
NEXT: Verify exact B158 mutation against src/routes/route-guards.test.tsx.
FACT: B158 exact source patch from mutation journal; command `cd frontend && npx rstest run src/routes/route-guards.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B158.log. Production restore `git diff --quiet -- frontend/src/routes/admin.courses.$courseId.tsx` exit 0.
FACT: B158 mutation failure excerpt (verbatim):
```text


### [F01] src/routes/route-guards.test.tsx :: guarded Save refreshes inventory and every lesson language, isolates sibling and uses latest token

repro:

```bash
npx rstest 'src/routes/route-guards.test.tsx' --testNamePattern 'guarded Save refreshes inventory and every lesson language, isolates sibling and uses latest token'
```

details:

```json
{
  "testPath": "src/routes/route-guards.test.tsx",
  "project": "dom",
  "fullName": "guarded Save refreshes inventory and every lesson language, isolates sibling and uses latest token",
  "status": "fail",
  "duration": 157,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected { …(6) } to match object { updatedAt: 'SAVED2' }\n(5 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/routes/route-guards.test.tsx",
        "line": 1445,
        "column": 75,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/routes/route-guards.test.tsx",
      "line": 1445
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  {
-   "updatedAt": "SAVED2",
+   "updatedAt": "EN1",
  }
```

codeFrame (error 1):

```text
1443 |   expect(f.state.inventoryReads).toBeGreaterThan(before);
1444 |   expect(client.getQueryState(lessonTranscriptQueryKey(f.lessonId))?.isInvalidated).toBe(true);
1445 |   expect(client.getQueryData(lessonTranscriptQueryKey(f.lessonId, 'en'))).toMatchObject({ updatedAt: 'SAVED2' });
     |                                                                           ^
1446 |   expect(client.getQueryData(lessonTranscriptQueryKey(f.lessonId, 'en-gb'))).toMatchObject({ updatedAt: 'GB1' });
1447 |   expect(f.state.requested.filter((language) => language === 'en-gb').length).toBeGreaterThan(1);
```

console:

```text
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
```


```
NEXT: Verify exact B193 mutation against src/routes/route-guards.test.tsx.
FACT: B193 exact source patch from mutation journal; command `cd frontend && npx rstest run src/routes/route-guards.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B193.log. Production restore `git diff --quiet -- frontend/src/routes/admin.courses.$courseId.tsx` exit 0.
FACT: B193 mutation failure excerpt (verbatim):
```text


### [F01] src/routes/route-guards.test.tsx :: guarded single-editor fallback loads the first default draft while no Ready video exists

repro:

```bash
npx rstest 'src/routes/route-guards.test.tsx' --testNamePattern 'guarded single-editor fallback loads the first default draft while no Ready video exists'
```

details:

```json
{
  "testPath": "src/routes/route-guards.test.tsx",
  "project": "dom",
  "fullName": "guarded single-editor fallback loads the first default draft while no Ready video exists",
  "status": "fail",
  "duration": 112,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <textarea tabindex=\"0\" …(5)></textarea> to have property \"value\" with value 'First default draft'",
      "topFrame": {
        "file": "src/routes/route-guards.test.tsx",
        "line": 1477,
        "column": 64,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/routes/route-guards.test.tsx",
      "line": 1477
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- First default draft
+ Last sibling draft
```

codeFrame (error 1):

```text
1475 |   fireEvent.click(await screen.findByRole('button', { name: 'Load draft into editor' }));
1476 |   fireEvent.click(screen.getByRole('button', { name: 'Load the draft' }));
1477 |   expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', 'First default draft');
     |                                                                ^
1478 | });
1479 | 
```

console:

```text
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
```


```
NEXT: Verify exact B195 mutation against src/routes/route-guards.test.tsx.
FACT: B195 exact source patch from mutation journal; command `cd frontend && npx rstest run src/routes/route-guards.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B195.log. Production restore `git diff --quiet -- frontend/src/routes/admin.courses.$courseId.tsx` exit 0.
FACT: B195 mutation failure excerpt (verbatim):
```text


### [F01] src/routes/route-guards.test.tsx :: failed guarded PUT does not invalidate settled inventory or any lesson transcript

repro:

```bash
npx rstest 'src/routes/route-guards.test.tsx' --testNamePattern 'failed guarded PUT does not invalidate settled inventory or any lesson transcript'
```

details:

```json
{
  "testPath": "src/routes/route-guards.test.tsx",
  "project": "dom",
  "fullName": "failed guarded PUT does not invalidate settled inventory or any lesson transcript",
  "status": "fail",
  "duration": 160,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 4 to be 3 // Object.is equality",
      "topFrame": {
        "file": "src/routes/route-guards.test.tsx",
        "line": 1466,
        "column": 34,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/routes/route-guards.test.tsx",
      "line": 1466
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- 3
+ 4
```

codeFrame (error 1):

```text
1464 |   await screen.findByText('The service is unavailable. Try again.');
1465 |   stop();
1466 |   expect(f.state.inventoryReads).toBe(before + 1); // fresh Save verification only
     |                                  ^
1467 |   expect(f.state.requested).toHaveLength(reads);
1468 |   expect(invalidations).toEqual([]);
```

console:

```text
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
[stderr] warn: Generated path "/learn" for route "/learn" matched route "/learn/$" instead. This can happen when multiple route templates resolve to the same URL. Use the route template that matches the intended route, or adjust params.stringify if it changed the target path.
```


```
FACT: End-of-part `cd frontend && npm run format` exit 0; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-final-format.log.
FACT: End-of-part `cd frontend && npm run format:check` exit 0; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-final-format-check.log.
FACT: End-of-part `cd frontend && npm run lint` exit 0; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-final-lint.log.
FACT: End-of-part `cd frontend && npm run typecheck` exit 2; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-final-typecheck.log.
FACT: Unmutated baseline `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B017.log.
NEXT: Verify exact B017 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B017 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B017.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B017 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: regeneration retains the complete prior snapshot but clears its server timestamp

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'regeneration retains the complete prior snapshot but clears its server timestamp'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "regeneration retains the complete prior snapshot but clears its server timestamp",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 248,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 248
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,21 +1,14 @@
  [
    {
-     "assetId": "asset",
-     "failure": "recognition_failed",
-     "generation": 7,
-     "language": "fr",
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
      "lessonId": "lesson",
-     "segments": [
-       {
-         "endMs": 300,
-         "id": "line",
-         "startMs": 10,
-         "text": "Bonjour",
-       },
-     ],
-     "stale": true,
+     "segments": [],
+     "stale": false,
      "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
      "updatedAt": null,
```

codeFrame (error 1):

```text
246 |   await h.read();
247 |   await h.regenerate();
248 |   expect(await h.read()).toEqual([{ ...previous, state: 'pending', updatedAt: null }]);
    |                          ^
249 |   h.client.clear();
250 | });
```


```
NEXT: Verify exact B019 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B019 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B019.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B019 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: regeneration retains the complete prior snapshot but clears its server timestamp

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'regeneration retains the complete prior snapshot but clears its server timestamp'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "regeneration retains the complete prior snapshot but clears its server timestamp",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 248,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 248
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -16,8 +16,8 @@
      "stale": true,
      "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": null,
+     "updatedAt": "2026-11-16T12:00:00Z",
    },
  ]
```

codeFrame (error 1):

```text
246 |   await h.read();
247 |   await h.regenerate();
248 |   expect(await h.read()).toEqual([{ ...previous, state: 'pending', updatedAt: null }]);
    |                          ^
249 |   h.client.clear();
250 | });
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to match object [ { state: 'pending', …(1) } ]\n(10 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 262,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 262
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    {
      "state": "pending",
-     "updatedAt": null,
+     "updatedAt": "2026-11-16T12:00:00Z",
    },
  ]
```

codeFrame (error 1):

```text
260 |     await h.regenerate();
261 |     await h.read();
262 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toMatchObject([
    |                                                                                  ^
263 |       { state: 'pending', updatedAt: null },
264 |     ]);
```


```
NEXT: Verify exact B021 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B021 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B021.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B021 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: markers never leak rows or polling into a sibling lesson

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'markers never leak rows or polling into a sibling lesson'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "markers never leak rows or polling into a sibling lesson",
  "status": "fail",
  "duration": 13,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'sibling', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 279,
        "column": 46,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 279
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "sibling",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
277 |   await h.regenerate();
278 |   const sibling = h.queries.lesson('sibling', h.client);
279 |   expect(await h.client.fetchQuery(sibling)).toEqual([]);
    |                                              ^
280 |   expect(h.interval('sibling')).toBe(false);
281 |   h.client.clear();
```


```
NEXT: Verify exact B025 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B025 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B025.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B025 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 267,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 267
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
265 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
266 |     expect(h.interval()).toBe(false);
267 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
268 |       exists ? [previous] : [],
269 |     );
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 3,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 267,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 267
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -5,12 +5,12 @@
      "generation": 1,
      "language": "en",
      "lessonId": "lesson",
      "segments": [],
      "stale": false,
-     "state": "ready",
+     "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": "2026-11-16T12:00:00Z",
+     "updatedAt": null,
    },
  ]
```

codeFrame (error 1):

```text
265 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
266 |     expect(h.interval()).toBe(false);
267 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
268 |       exists ? [previous] : [],
269 |     );
```

### [F03] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 310,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 310
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```tex
```
NEXT: Verify exact B026 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B026 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B026.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B026 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 13,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 267,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 267
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
265 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
266 |     expect(h.interval()).toBe(false);
267 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
268 |       exists ? [previous] : [],
269 |     );
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 310,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 310
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
308 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
309 |   expect(h.interval()).toBe(false);
310 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual([]);
    |                                                                                ^
311 |   h.client.clear();
312 | });
```


```
NEXT: Verify exact B027 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B027 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B027.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B027 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 267,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 267
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -5,12 +5,12 @@
      "generation": 1,
      "language": "en",
      "lessonId": "lesson",
      "segments": [],
      "stale": false,
-     "state": "ready",
+     "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": "2026-11-16T12:00:00Z",
+     "updatedAt": null,
    },
  ]
```

codeFrame (error 1):

```text
265 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
266 |     expect(h.interval()).toBe(false);
267 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
268 |       exists ? [previous] : [],
269 |     );
```


```
NEXT: Verify exact B029 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B029 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B029.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B029 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: running with the same timestamp acknowledges the claim and terminal data stops polling

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'running with the same timestamp acknowledges the claim and terminal data stops polling'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "running with the same timestamp acknowledges the claim and terminal data stops polling",
  "status": "fail",
  "duration": 13,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to match object [ { state: 'running' } ]\n(11 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 292,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 292
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    {
-     "state": "running",
+     "state": "pending",
    },
  ]
```

codeFrame (error 1):

```text
290 |     value: [draft(trackT, { state: 'running' })],
291 |   });
292 |   expect(await h.read()).toMatchObject([{ state: 'running' }]);
    |                          ^
293 |   h.repository.listDrafts = async () => ({ ok: true, value: [draft(trackT)] });
294 |   expect(await h.read()).toEqual([draft(trackT)]);
```


```
NEXT: Verify exact B034 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B034 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B034.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B034 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 5000 to be false // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 309,
        "column": 24,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 309
    }
  ]
}
```

diff:

```diff
- Expected:
false

+ Received:
5000
```

codeFrame (error 1):

```text
307 |   await h.regenerate();
308 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
309 |   expect(h.interval()).toBe(false);
    |                        ^
310 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual([]);
311 |   h.client.clear();
```


```
NEXT: Verify exact B036 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B036 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B036.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B036 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: inventory errors preserve their named code

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'inventory errors preserve their named code'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "inventory errors preserve their named code",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise resolved \"[]\" instead of rejecting",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 324,
        "column": 16,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 324
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
[]
```

codeFrame (error 1):

```text
322 |     ? h.client.fetchQuery(h.queries.inventory(lessonId))
323 |     : h.client.fetchQuery(h.queries.lesson(lessonId, h.client));
324 |   await expect(read).rejects.toMatchObject({
    |                ^
325 |     name: 'DraftQueryError',
326 |     code: 'dependency_unavailable',
```


```
NEXT: Verify exact B037 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B037 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B037.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B037 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: lesson errors preserve their named code

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'lesson errors preserve their named code'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "lesson errors preserve their named code",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise resolved \"[]\" instead of rejecting",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 324,
        "column": 16,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 324
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
[]
```

codeFrame (error 1):

```text
322 |     ? h.client.fetchQuery(h.queries.inventory(lessonId))
323 |     : h.client.fetchQuery(h.queries.lesson(lessonId, h.client));
324 |   await expect(read).rejects.toMatchObject({
    |                ^
325 |     name: 'DraftQueryError',
326 |     code: 'dependency_unavailable',
```


```
NEXT: Verify exact B039 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B039 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B039.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B039 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 267,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 267
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
265 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
266 |     expect(h.interval()).toBe(false);
267 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
268 |       exists ? [previous] : [],
269 |     );
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 2,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 267,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 267
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -5,12 +5,12 @@
      "generation": 1,
      "language": "en",
      "lessonId": "lesson",
      "segments": [],
      "stale": false,
-     "state": "ready",
+     "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": "2026-11-16T12:00:00Z",
+     "updatedAt": null,
    },
  ]
```

codeFrame (error 1):

```text
265 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
266 |     expect(h.interval()).toBe(false);
267 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
268 |       exists ? [previous] : [],
269 |     );
```

### [F03] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 310,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 310
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```tex
```
NEXT: Verify exact B040 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B040 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B040.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B040 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: markers never leak rows or polling into a sibling lesson

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'markers never leak rows or polling into a sibling lesson'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "markers never leak rows or polling into a sibling lesson",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 5000 to be false // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 280,
        "column": 33,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 280
    }
  ]
}
```

diff:

```diff
- Expected:
false

+ Received:
5000
```

codeFrame (error 1):

```text
278 |   const sibling = h.queries.lesson('sibling', h.client);
279 |   expect(await h.client.fetchQuery(sibling)).toEqual([]);
280 |   expect(h.interval('sibling')).toBe(false);
    |                                 ^
281 |   h.client.clear();
282 | });
```


```
NEXT: Verify exact B047 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B047 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B047.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B047 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: failed regeneration never creates a pending row or starts polling

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'failed regeneration never creates a pending row or starts polling'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "failed regeneration never creates a pending row or starts polling",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 336,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 336
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
334 |   h.repository.regenerate = async () => ({ ok: false, error: { code: 'dependency_unavailable' } });
335 |   await expect(h.regenerate()).rejects.toMatchObject({ code: 'dependency_unavailable' });
336 |   expect(await h.read()).toEqual([]);
    |                          ^
337 |   expect(h.interval()).toBe(false);
338 |   h.client.clear();
```


```
NEXT: Verify exact B160 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B160 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B160.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B160 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 267,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 267
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
265 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
266 |     expect(h.interval()).toBe(false);
267 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
268 |       exists ? [previous] : [],
269 |     );
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 310,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 310
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
308 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
309 |   expect(h.interval()).toBe(false);
310 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual([]);
    |                                                                                ^
311 |   h.client.clear();
312 | });
```


```
NEXT: Verify exact B161 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B161 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B161.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B161 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: null-track placeholder does not rewrite the cache while a requested marker waits

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'null-track placeholder does not rewrite the cache while a requested marker waits'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "null-track placeholder does not rewrite the cache while a requested marker waits",
  "status": "fail",
  "duration": 15,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 2000 to be 1000 // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 351,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 351
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- 1000
+ 2000
```

codeFrame (error 1):

```text
349 |   rstest.setSystemTime(2000);
350 |   expect(h.interval()).toBe(draftPollMs);
351 |   expect(h.client.getQueryState(key)?.dataUpdatedAt).toBe(before);
    |                                                      ^
352 |   expect(h.client.getQueryData(key)).toHaveLength(1);
353 |   h.client.clear();
```


```
FACT: Unmutated baseline `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B100.log.
NEXT: Verify exact B100 mutation against src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx.
FACT: B100 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B100.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B100 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx :: unmount cancels pending und Save promise without a write

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx' --testNamePattern 'unmount cancels pending und Save promise without a write'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
  "project": "dom",
  "fullName": "unmount cancels pending und Save promise without a write",
  "status": "fail",
  "duration": 125,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 'pending' to be 'language_cancelled' // Object.is equality",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
        "line": 117,
        "column": 20,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
      "line": 117
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- language_cancelled
+ pending
```

codeFrame (error 1):

```text
115 |       view.unmount();
116 |     });
117 |     expect(result).toBe(mode === 'und' ? 'language_cancelled' : 'replacement_cancelled');
    |                    ^
118 |     expect(writes).toEqual([]);
119 |   });
```


```
NEXT: Verify exact B103 mutation against src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx.
FACT: B103 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B103.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B103 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx :: unmount cancels pending replacement Save promise without a write

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx' --testNamePattern 'unmount cancels pending replacement Save promise without a write'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
  "project": "dom",
  "fullName": "unmount cancels pending replacement Save promise without a write",
  "status": "fail",
  "duration": 56,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 'pending' to be 'replacement_cancelled' // Object.is equality",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
        "line": 117,
        "column": 20,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx",
      "line": 117
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- replacement_cancelled
+ pending
```

codeFrame (error 1):

```text
115 |       view.unmount();
116 |     });
117 |     expect(result).toBe(mode === 'und' ? 'language_cancelled' : 'replacement_cancelled');
    |                    ^
118 |     expect(writes).toEqual([]);
119 |   });
```


```
FACT: Unmutated baseline `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B082.log.
NEXT: Verify exact B082 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B082 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B082.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B082 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: tabs select newest Ready generation of this lesson, ignoring foreign and processing videos

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'tabs select newest Ready generation of this lesson, ignoring foreign and processing videos'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "tabs select newest Ready generation of this lesson, ignoring foreign and processing videos",
  "status": "fail",
  "duration": 93,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'foreign' ] to deeply equal [ 'newest-ready' ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1197,
        "column": 60,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1197
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
-   "newest-ready",
+   "foreign",
  ]
```

codeFrame (error 1):

```text
1195 |   const asked = rstest.spyOn(f.video.audio, 'useTracks');
1196 |   await mount(f);
1197 |   expect([...new Set(asked.mock.calls.map(([id]) => id))]).toEqual(['newest-ready']);
     |                                                            ^
1198 | });
1199 | 
```


```
NEXT: Verify exact B083 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B083 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B083.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B083 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: tabs select newest Ready generation of this lesson, ignoring foreign and processing videos

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'tabs select newest Ready generation of this lesson, ignoring foreign and processing videos'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "tabs select newest Ready generation of this lesson, ignoring foreign and processing videos",
  "status": "fail",
  "duration": 40,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'processing' ] to deeply equal [ 'newest-ready' ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1197,
        "column": 60,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1197
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
-   "newest-ready",
+   "processing",
  ]
```

codeFrame (error 1):

```text
1195 |   const asked = rstest.spyOn(f.video.audio, 'useTracks');
1196 |   await mount(f);
1197 |   expect([...new Set(asked.mock.calls.map(([id]) => id))]).toEqual(['newest-ready']);
     |                                                            ^
1198 | });
1199 | 
```

### [F02] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: only processing videos and missing optional audio both retain the single editor

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'only processing videos and missing optional audio both retain the single editor'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "only processing videos and missing optional audio both retain the single editor",
  "status": "fail",
  "duration": 15,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <div class=\"ui-tab-list\" …(7)>…(1)</div> to be null",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1204,
        "column": 41,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1204
    }
  ]
}
```

diff:

```diff
- Expected:
null

+ Received:
<div
  aria-label="Transcript languages"
  aria-orientation="horizontal"
  class="ui-tab-list"
  data-collection="react-aria-_r_2n8_"
  data-orientation="horizontal"
  data-rac=""
  id="react-aria-_r_2n9_"
  role="tablist"
>
  <div
    aria-controls="react-aria-_r_2n9_-tabpanel-track:one"
    aria-selected="true"
    class="ui-tab"
    data-collection="react-aria-_r_2n8_"
    data-key="track:one"
    data-rac=""
    data-react-aria-pressable="true"
    data-selected="true"
    id="react-aria-_r_2n9_-tab-track:one"
    role="tab"
    tabindex="0"
  >
    English · Default
  </div>
</div>
```

codeFrame (error 1):

```text
1202 |   f.setVideo({ ...f.video, rows: f.video.rows.map((row) => ({ ...row, state: 'processing' })) });
1203 |   const first = await mount(f);
1204 |   expect(screen.queryByRole('tablist')).toBeNull();
     |                                         ^
1205 |   expect(screen.getByRole('heading', { name: 'Lesson transcript' })).toBeTruthy();
1206 |   first.unmount();
```


```
NEXT: Verify exact B084 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B084 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B084.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B084 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: tabs select newest Ready generation of this lesson, ignoring foreign and processing videos

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'tabs select newest Ready generation of this lesson, ignoring foreign and processing videos'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "tabs select newest Ready generation of this lesson, ignoring foreign and processing videos",
  "status": "fail",
  "duration": 45,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ 'older' ] to deeply equal [ 'newest-ready' ]",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1197,
        "column": 60,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1197
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
-   "newest-ready",
+   "older",
  ]
```

codeFrame (error 1):

```text
1195 |   const asked = rstest.spyOn(f.video.audio, 'useTracks');
1196 |   await mount(f);
1197 |   expect([...new Set(asked.mock.calls.map(([id]) => id))]).toEqual(['newest-ready']);
     |                                                            ^
1198 | });
1199 | 
```


```
NEXT: Verify exact B168 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B168 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B168.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B168 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: only processing videos and missing optional audio both retain the single editor

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'only processing videos and missing optional audio both retain the single editor'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "only processing videos and missing optional audio both retain the single editor",
  "status": "fail",
  "duration": 84,
  "retryCount": 0,
  "errors": [
    {
      "type": "TypeError",
      "message": "Cannot read properties of undefined (reading 'useTracks')",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.tsx",
        "line": 141,
        "column": 24,
        "method": "ReadyLanguageTabs"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.tsx",
      "line": 141
    }
  ]
}
```

codeFrame (error 1):

```text
139 | }) {
140 |   const { t } = useTranslation();
141 |   const tracks = audio.useTracks(videoId);
    |                        ^
142 |   const inventory = languages.useInventory(props.lessonId);
143 |   const { lessonId, onTranscriptKnown } = props;
```


```
FACT: Unmutated baseline `cd frontend && npx rstest run src/modules/media/ui/LessonAudioTracksField.test.tsx` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B060.log.
NEXT: Verify exact B060 mutation against src/modules/media/ui/LessonAudioTracksField.test.tsx.
FACT: B060 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/ui/LessonAudioTracksField.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B060.log. Production restore `git diff --quiet -- frontend/src/modules/media/ui/LessonAudioTracksField.tsx` exit 0.
FACT: B060 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/ui/LessonAudioTracksField.test.tsx :: upload and row switches have distinct hints, and successful upload resets the server default

repro:

```bash
npx rstest 'src/modules/media/ui/LessonAudioTracksField.test.tsx' --testNamePattern 'upload and row switches have distinct hints, and successful upload resets the server default'
```

details:

```json
{
  "testPath": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
  "project": "dom",
  "fullName": "upload and row switches have distinct hints, and successful upload resets the server default",
  "status": "fail",
  "duration": 109,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected <input …(6)></input> to have property \"checked\" with value true",
      "topFrame": {
        "file": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
        "line": 493,
        "column": 17,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
      "line": 493
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
491 |   finish();
492 |   const reset = await screen.findByRole('switch', { name: 'Multiple languages' });
493 |   expect(reset).toHaveProperty('checked', true);
    |                 ^
494 |   await user.click(screen.getByRole('button', { name: /Choose a language Track language/ }));
495 |   await user.click(screen.getByRole('option', { name: 'Polish' }));
```


```
NEXT: Verify exact B064 mutation against src/modules/media/ui/LessonAudioTracksField.test.tsx.
FACT: B064 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/ui/LessonAudioTracksField.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B064.log. Production restore `git diff --quiet -- frontend/src/modules/media/ui/LessonAudioTracksField.tsx` exit 0.
FACT: B064 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/ui/LessonAudioTracksField.test.tsx :: upload and row switches have distinct hints, and successful upload resets the server default

repro:

```bash
npx rstest 'src/modules/media/ui/LessonAudioTracksField.test.tsx' --testNamePattern 'upload and row switches have distinct hints, and successful upload resets the server default'
```

details:

```json
{
  "testPath": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
  "project": "dom",
  "fullName": "upload and row switches have distinct hints, and successful upload resets the server default",
  "status": "fail",
  "duration": 19,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected null to be truthy",
      "topFrame": {
        "file": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
        "line": 476,
        "column": 16,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
      "line": 476
    }
  ]
}
```

diff:

```diff
- Expected:
true

+ Received:
null
```

codeFrame (error 1):

```text
474 |   expect(new Set(descriptions).size).toBe(3);
475 |   for (const id of descriptions) {
476 |     expect(id).toBeTruthy();
    |                ^
477 |     expect(document.getElementById(id ?? '')?.textContent).toBe(
478 |       'Recognition detects the language of each passage.',
```


```
NEXT: Verify exact B070 mutation against src/modules/media/ui/LessonAudioTracksField.test.tsx.
FACT: B070 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/media/ui/LessonAudioTracksField.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B070.log. Production restore `git diff --quiet -- frontend/src/modules/media/ui/LessonAudioTracksField.tsx` exit 0.
FACT: B070 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/media/ui/LessonAudioTracksField.test.tsx :: upload and row switches have distinct hints, and successful upload resets the server default

repro:

```bash
npx rstest 'src/modules/media/ui/LessonAudioTracksField.test.tsx' --testNamePattern 'upload and row switches have distinct hints, and successful upload resets the server default'
```

details:

```json
{
  "testPath": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
  "project": "dom",
  "fullName": "upload and row switches have distinct hints, and successful upload resets the server default",
  "status": "fail",
  "duration": 21,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 2 to be 3 // Object.is equality",
      "topFrame": {
        "file": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
        "line": 474,
        "column": 38,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/media/ui/LessonAudioTracksField.test.tsx",
      "line": 474
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- 3
+ 2
```

codeFrame (error 1):

```text
472 |   const switches = screen.getAllByRole('switch');
473 |   const descriptions = switches.map((toggle) => toggle.getAttribute('aria-describedby'));
474 |   expect(new Set(descriptions).size).toBe(3);
    |                                      ^
475 |   for (const id of descriptions) {
476 |     expect(id).toBeTruthy();
```


```
FACT: Unmutated baseline `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B017.log.
NEXT: Verify exact B017 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B017 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B017.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B017 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: regeneration retains the complete prior snapshot but clears its server timestamp

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'regeneration retains the complete prior snapshot but clears its server timestamp'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "regeneration retains the complete prior snapshot but clears its server timestamp",
  "status": "fail",
  "duration": 12,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 249,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 249
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -1,21 +1,14 @@
  [
    {
-     "assetId": "asset",
-     "failure": "recognition_failed",
-     "generation": 7,
-     "language": "fr",
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
      "lessonId": "lesson",
-     "segments": [
-       {
-         "endMs": 300,
-         "id": "line",
-         "startMs": 10,
-         "text": "Bonjour",
-       },
-     ],
-     "stale": true,
+     "segments": [],
+     "stale": false,
      "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
      "updatedAt": null,
```

codeFrame (error 1):

```text
247 |   await h.read();
248 |   await h.regenerate();
249 |   expect(await h.read()).toEqual([{ ...previous, state: 'pending', updatedAt: null }]);
    |                          ^
250 |   h.client.clear();
251 | });
```


```
NEXT: Verify exact B019 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B019 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B019.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B019 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: regeneration retains the complete prior snapshot but clears its server timestamp

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'regeneration retains the complete prior snapshot but clears its server timestamp'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "regeneration retains the complete prior snapshot but clears its server timestamp",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 249,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 249
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -16,8 +16,8 @@
      "stale": true,
      "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": null,
+     "updatedAt": "2026-11-16T12:00:00Z",
    },
  ]
```

codeFrame (error 1):

```text
247 |   await h.read();
248 |   await h.regenerate();
249 |   expect(await h.read()).toEqual([{ ...previous, state: 'pending', updatedAt: null }]);
    |                          ^
250 |   h.client.clear();
251 | });
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 2,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to match object [ { state: 'pending', …(1) } ]\n(10 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 263,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 263
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    {
      "state": "pending",
-     "updatedAt": null,
+     "updatedAt": "2026-11-16T12:00:00Z",
    },
  ]
```

codeFrame (error 1):

```text
261 |     await h.regenerate();
262 |     await h.read();
263 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toMatchObject([
    |                                                                                  ^
264 |       { state: 'pending', updatedAt: null },
265 |     ]);
```


```
NEXT: Verify exact B021 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B021 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B021.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B021 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: markers never leak rows or polling into a sibling lesson

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'markers never leak rows or polling into a sibling lesson'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "markers never leak rows or polling into a sibling lesson",
  "status": "fail",
  "duration": 8,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'sibling', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 280,
        "column": 46,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 280
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "sibling",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
278 |   await h.regenerate();
279 |   const sibling = h.queries.lesson('sibling', h.client);
280 |   expect(await h.client.fetchQuery(sibling)).toEqual([]);
    |                                              ^
281 |   expect(h.interval('sibling')).toBe(false);
282 |   h.client.clear();
```


```
NEXT: Verify exact B025 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B025 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B025.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B025 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 13,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 268,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 268
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
266 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
267 |     expect(h.interval()).toBe(false);
268 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
269 |       exists ? [previous] : [],
270 |     );
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 268,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 268
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -5,12 +5,12 @@
      "generation": 1,
      "language": "en",
      "lessonId": "lesson",
      "segments": [],
      "stale": false,
-     "state": "ready",
+     "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": "2026-11-16T12:00:00Z",
+     "updatedAt": null,
    },
  ]
```

codeFrame (error 1):

```text
266 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
267 |     expect(h.interval()).toBe(false);
268 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
269 |       exists ? [previous] : [],
270 |     );
```

### [F03] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 311,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 311
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```tex
```
NEXT: Verify exact B026 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B026 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B026.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B026 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 10,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 268,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 268
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
266 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
267 |     expect(h.interval()).toBe(false);
268 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
269 |       exists ? [previous] : [],
270 |     );
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 311,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 311
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
309 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
310 |   expect(h.interval()).toBe(false);
311 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual([]);
    |                                                                                ^
312 |   h.client.clear();
313 | });
```


```
NEXT: Verify exact B027 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B027 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B027.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B027 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 268,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 268
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -5,12 +5,12 @@
      "generation": 1,
      "language": "en",
      "lessonId": "lesson",
      "segments": [],
      "stale": false,
-     "state": "ready",
+     "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": "2026-11-16T12:00:00Z",
+     "updatedAt": null,
    },
  ]
```

codeFrame (error 1):

```text
266 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
267 |     expect(h.interval()).toBe(false);
268 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
269 |       exists ? [previous] : [],
270 |     );
```


```
NEXT: Verify exact B029 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B029 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B029.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B029 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: running with the same timestamp acknowledges the claim and terminal data stops polling

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'running with the same timestamp acknowledges the claim and terminal data stops polling'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "running with the same timestamp acknowledges the claim and terminal data stops polling",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to match object [ { state: 'running' } ]\n(11 matching properties omitted from actual)",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 293,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 293
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

  [
    {
-     "state": "running",
+     "state": "pending",
    },
  ]
```

codeFrame (error 1):

```text
291 |     value: [draft(trackT, { state: 'running' })],
292 |   });
293 |   expect(await h.read()).toMatchObject([{ state: 'running' }]);
    |                          ^
294 |   h.repository.listDrafts = async () => ({ ok: true, value: [draft(trackT)] });
295 |   expect(await h.read()).toEqual([draft(trackT)]);
```


```
NEXT: Verify exact B034 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B034 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B034.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B034 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 7,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 5000 to be false // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 310,
        "column": 24,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 310
    }
  ]
}
```

diff:

```diff
- Expected:
false

+ Received:
5000
```

codeFrame (error 1):

```text
308 |   await h.regenerate();
309 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
310 |   expect(h.interval()).toBe(false);
    |                        ^
311 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual([]);
312 |   h.client.clear();
```


```
NEXT: Verify exact B036 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B036 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B036.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B036 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: inventory errors preserve their named code

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'inventory errors preserve their named code'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "inventory errors preserve their named code",
  "status": "fail",
  "duration": 12,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise resolved \"[]\" instead of rejecting",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 325,
        "column": 16,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 325
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
[]
```

codeFrame (error 1):

```text
323 |     ? h.client.fetchQuery(h.queries.inventory(lessonId))
324 |     : h.client.fetchQuery(h.queries.lesson(lessonId, h.client));
325 |   await expect(read).rejects.toMatchObject({
    |                ^
326 |     name: 'DraftQueryError',
327 |     code: 'dependency_unavailable',
```


```
NEXT: Verify exact B037 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B037 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B037.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B037 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: lesson errors preserve their named code

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'lesson errors preserve their named code'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "lesson errors preserve their named code",
  "status": "fail",
  "duration": 11,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise resolved \"[]\" instead of rejecting",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 325,
        "column": 16,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 325
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
[]
```

codeFrame (error 1):

```text
323 |     ? h.client.fetchQuery(h.queries.inventory(lessonId))
324 |     : h.client.fetchQuery(h.queries.lesson(lessonId, h.client));
325 |   await expect(read).rejects.toMatchObject({
    |                ^
326 |     name: 'DraftQueryError',
327 |     code: 'dependency_unavailable',
```


```
NEXT: Verify exact B039 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B039 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B039.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B039 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 19,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 268,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 268
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
266 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
267 |     expect(h.interval()).toBe(false);
268 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
269 |       exists ? [previous] : [],
270 |     );
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=true

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=true'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=true",
  "status": "fail",
  "duration": 1,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal [ { lessonId: 'lesson', …(11) } ]",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 268,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 268
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

@@ -5,12 +5,12 @@
      "generation": 1,
      "language": "en",
      "lessonId": "lesson",
      "segments": [],
      "stale": false,
-     "state": "ready",
+     "state": "pending",
      "trackId": "T",
      "trackLanguage": "en",
      "trackMultilingual": false,
-     "updatedAt": "2026-11-16T12:00:00Z",
+     "updatedAt": null,
    },
  ]
```

codeFrame (error 1):

```text
266 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
267 |     expect(h.interval()).toBe(false);
268 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
269 |       exists ? [previous] : [],
270 |     );
```

### [F03] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 2,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 311,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 311
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```tex
```
NEXT: Verify exact B040 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B040 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B040.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B040 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: markers never leak rows or polling into a sibling lesson

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'markers never leak rows or polling into a sibling lesson'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "markers never leak rows or polling into a sibling lesson",
  "status": "fail",
  "duration": 16,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 5000 to be false // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 281,
        "column": 33,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 281
    }
  ]
}
```

diff:

```diff
- Expected:
false

+ Received:
5000
```

codeFrame (error 1):

```text
279 |   const sibling = h.queries.lesson('sibling', h.client);
280 |   expect(await h.client.fetchQuery(sibling)).toEqual([]);
281 |   expect(h.interval('sibling')).toBe(false);
    |                                 ^
282 |   h.client.clear();
283 | });
```


```
NEXT: Verify exact B047 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B047 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B047.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B047 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: failed regeneration never creates a pending row or starts polling

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'failed regeneration never creates a pending row or starts polling'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "failed regeneration never creates a pending row or starts polling",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 337,
        "column": 26,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 337
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
335 |   h.repository.regenerate = async () => ({ ok: false, error: { code: 'dependency_unavailable' } });
336 |   await expect(h.regenerate()).rejects.toMatchObject({ code: 'dependency_unavailable' });
337 |   expect(await h.read()).toEqual([]);
    |                          ^
338 |   expect(h.interval()).toBe(false);
339 |   h.client.clear();
```


```
NEXT: Verify exact B160 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B160 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B160.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B160 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: interval expiry restores cache without a fetch; prior snapshot=false

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'interval expiry restores cache without a fetch; prior snapshot=false'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "interval expiry restores cache without a fetch; prior snapshot=false",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 268,
        "column": 82,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 268
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
266 |     rstest.setSystemTime(1000 + draftTaskTimeoutMs);
267 |     expect(h.interval()).toBe(false);
268 |     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
    |                                                                                  ^
269 |       exists ? [previous] : [],
270 |     );
```

### [F02] src/app/composition/admin-draft-queries.test.ts :: repeated regeneration keeps the original fifteen-minute queue deadline

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'repeated regeneration keeps the original fifteen-minute queue deadline'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "repeated regeneration keeps the original fifteen-minute queue deadline",
  "status": "fail",
  "duration": 2,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected [ { lessonId: 'lesson', …(11) } ] to deeply equal []",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 311,
        "column": 80,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 311
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
+     "assetId": null,
+     "failure": null,
+     "generation": null,
+     "language": null,
+     "lessonId": "lesson",
+     "segments": [],
+     "stale": false,
+     "state": "pending",
+     "trackId": "T",
+     "trackLanguage": "und",
+     "trackMultilingual": false,
+     "updatedAt": null,
+   },
+ ]
```

codeFrame (error 1):

```text
309 |   rstest.setSystemTime(1000 + draftTaskTimeoutMs);
310 |   expect(h.interval()).toBe(false);
311 |   expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual([]);
    |                                                                                ^
312 |   h.client.clear();
313 | });
```


```
NEXT: Verify exact B161 mutation against src/app/composition/admin-draft-queries.test.ts.
FACT: B161 exact source patch from mutation journal; command `cd frontend && npx rstest run src/app/composition/admin-draft-queries.test.ts`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B161.log. Production restore `git diff --quiet -- frontend/src/app/composition/admin-draft-queries.ts` exit 0.
FACT: B161 mutation failure excerpt (verbatim):
```text


### [F01] src/app/composition/admin-draft-queries.test.ts :: null-track placeholder does not rewrite the cache while a requested marker waits

repro:

```bash
npx rstest 'src/app/composition/admin-draft-queries.test.ts' --testNamePattern 'null-track placeholder does not rewrite the cache while a requested marker waits'
```

details:

```json
{
  "testPath": "src/app/composition/admin-draft-queries.test.ts",
  "project": "dom",
  "fullName": "null-track placeholder does not rewrite the cache while a requested marker waits",
  "status": "fail",
  "duration": 9,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "expected 2000 to be 1000 // Object.is equality",
      "topFrame": {
        "file": "src/app/composition/admin-draft-queries.test.ts",
        "line": 352,
        "column": 54,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/app/composition/admin-draft-queries.test.ts",
      "line": 352
    }
  ]
}
```

diff:

```diff
- Expected
+ Received

- 1000
+ 2000
```

codeFrame (error 1):

```text
350 |   rstest.setSystemTime(2000);
351 |   expect(h.interval()).toBe(draftPollMs);
352 |   expect(h.client.getQueryState(key)?.dataUpdatedAt).toBe(before);
    |                                                      ^
353 |   expect(h.client.getQueryData(key)).toHaveLength(1);
354 |   h.client.clear();
```


```
FACT: All 60 original GAPs produced behavior-attributable mutation failures. Later typecheck exposed only test generic-inference errors (QueryCache key/data tagging, inventory/draft promise union and hoisted probe narrowing). Corrected tests without weakening assertions; isolated typecheck retry exit 0. Rechecking every affected query/lifecycle discriminator.
DECISION: B100/B103 probe the pending Save promise at the existing LessonTranscriptField.editorSave callback. Real LanguageEditor Save logic, real dialogs and React unmount effects execute; only the child editor is a probe because after unmount there is no DOM in which to observe cancellation. B168 additionally wraps mount in a resolves assertion, making the missing-audio regression an explicit assertion failure rather than only a production useTracks exception.
NEXT: Finish affected mutation reruns, run ordered frontend formatting/lint/typecheck/full unit gates, append final 60-row table, diff/stat and SHA-256 inventory.
FACT: Unmutated baseline `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx` exit 0; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-baseline-B168.log.
NEXT: Verify exact B168 mutation against src/modules/admin/ui/LessonLanguageTabs.test.tsx.
FACT: B168 exact source patch from mutation journal; command `cd frontend && npx rstest run src/modules/admin/ui/LessonLanguageTabs.test.tsx`; exit 1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/gap-B168.log. Production restore `git diff --quiet -- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` exit 0.
FACT: B168 mutation failure excerpt (verbatim):
```text


### [F01] src/modules/admin/ui/LessonLanguageTabs.test.tsx :: only processing videos and missing optional audio both retain the single editor

repro:

```bash
npx rstest 'src/modules/admin/ui/LessonLanguageTabs.test.tsx' --testNamePattern 'only processing videos and missing optional audio both retain the single editor'
```

details:

```json
{
  "testPath": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
  "project": "dom",
  "fullName": "only processing videos and missing optional audio both retain the single editor",
  "status": "fail",
  "duration": 88,
  "retryCount": 0,
  "errors": [
    {
      "type": "AssertionError",
      "message": "promise rejected \"TypeError: Cannot read properties of unde…\" instead of resolving",
      "topFrame": {
        "file": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
        "line": 1209,
        "column": 22,
        "method": "<unknown>"
      },
      "stackFrames": []
    }
  ],
  "candidateFiles": [
    {
      "path": "src/modules/admin/ui/LessonLanguageTabs.test.tsx",
      "line": 1209
    }
  ]
}
```

codeFrame (error 1):

```text
1207 |   const { audio: _audio, ...withoutAudio } = f.video;
1208 |   f.setVideo(withoutAudio);
1209 |   await expect(mount(f)).resolves.toBeDefined();
     |                      ^
1210 |   expect(screen.queryByRole('tablist')).toBeNull();
1211 |   expect(screen.getByRole('heading', { name: 'Lesson transcript' })).toBeTruthy();
```


```
FACT: Final ordered run `cd frontend && npm run format` exit 0; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-final-v2-format.log.
FACT: Final ordered run `cd frontend && npm run format:check` exit 0; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-final-v2-format-check.log.
FACT: Final ordered run `cd frontend && npm run lint` exit 0; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-final-v2-lint.log.
FACT: Final ordered run `cd frontend && npm run typecheck` exit 0; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-final-v2-typecheck.log.
FACT: Final GAP ledger; latest gap-Bnnn.log contains the exact command output for each mutation. Prior SURVIVED attempts (B130/B136) were test-design iterations and are superseded by their later baseline-PASS / assertion-FAIL pairs, never counted as closures on their own. Formatting-only edits after these checks preserve all assertions and fixtures.

| id | test file::test name | baseline | mutated FAIL line (verbatim) | CLOSED | DEFECT |
|---|---|---|---|---|---|
| B004 | src/app/composition/admin-audio-queries.test.ts::terminal audio transition refetches only the affected mounted draft | PASS | `223 \|     expect(affectedReads).toBe(before + 1);` | CLOSED | — |
| B005 | src/app/composition/admin-audio-queries.test.ts::list envelope with queued audio schedules another port read | PASS | `245 \|     expect(reads).toBe(before + 1);` | CLOSED | — |
| B017 | src/app/composition/admin-draft-queries.test.ts::regeneration retains the complete prior snapshot but clears its server timestamp | PASS | `249 \|   expect(await h.read()).toEqual([{ ...previous, state: 'pending', updatedAt: null }]);` | CLOSED | — |
| B019 | src/app/composition/admin-draft-queries.test.ts::regeneration retains the complete prior snapshot but clears its server timestamp | PASS | `249 \|   expect(await h.read()).toEqual([{ ...previous, state: 'pending', updatedAt: null }]);` | CLOSED | — |
| B021 | src/app/composition/admin-draft-queries.test.ts::markers never leak rows or polling into a sibling lesson | PASS | `280 \|   expect(await h.client.fetchQuery(sibling)).toEqual([]);` | CLOSED | — |
| B025 | src/app/composition/admin-draft-queries.test.ts::interval expiry restores cache without a fetch; prior snapshot=false | PASS | `268 \|     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(` | CLOSED | — |
| B026 | src/app/composition/admin-draft-queries.test.ts::interval expiry restores cache without a fetch; prior snapshot=false | PASS | `268 \|     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(` | CLOSED | — |
| B027 | src/app/composition/admin-draft-queries.test.ts::interval expiry restores cache without a fetch; prior snapshot=true | PASS | `268 \|     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(` | CLOSED | — |
| B029 | src/app/composition/admin-draft-queries.test.ts::running with the same timestamp acknowledges the claim and terminal data stops polling | PASS | `293 \|   expect(await h.read()).toMatchObject([{ state: 'running' }]);` | CLOSED | — |
| B034 | src/app/composition/admin-draft-queries.test.ts::repeated regeneration keeps the original fifteen-minute queue deadline | PASS | `310 \|   expect(h.interval()).toBe(false);` | CLOSED | — |
| B036 | src/app/composition/admin-draft-queries.test.ts::inventory errors preserve their named code | PASS | `325 \|   await expect(read).rejects.toMatchObject({` | CLOSED | — |
| B037 | src/app/composition/admin-draft-queries.test.ts::lesson errors preserve their named code | PASS | `325 \|   await expect(read).rejects.toMatchObject({` | CLOSED | — |
| B039 | src/app/composition/admin-draft-queries.test.ts::interval expiry restores cache without a fetch; prior snapshot=false | PASS | `268 \|     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(` | CLOSED | — |
| B040 | src/app/composition/admin-draft-queries.test.ts::markers never leak rows or polling into a sibling lesson | PASS | `281 \|   expect(h.interval('sibling')).toBe(false);` | CLOSED | — |
| B047 | src/app/composition/admin-draft-queries.test.ts::failed regeneration never creates a pending row or starts polling | PASS | `337 \|   expect(await h.read()).toEqual([]);` | CLOSED | — |
| B053 | src/modules/media/adapters/audio-track-repository.test.ts::audio track API repository > rejects a list with invalid list multilingualDefault false | PASS | `121 \|     expect(await createApiAudioTrackRepository().list(videoId)).toEqual({` | CLOSED | — |
| B054 | src/modules/media/adapters/audio-track-repository.test.ts::audio track API repository > rejects a list with unknown envelope key | PASS | `121 \|     expect(await createApiAudioTrackRepository().list(videoId)).toEqual({` | CLOSED | — |
| B060 | src/modules/media/ui/LessonAudioTracksField.test.tsx::upload and row switches have distinct hints, and successful upload resets the server default | PASS | `493 \|   expect(reset).toHaveProperty('checked', true);` | CLOSED | — |
| B061 | src/modules/media/ui/LessonAudioTracksField.test.tsx::pending row action makes the visible upload switch read-only without dropping focus | PASS | `501 \|   expect(upload.getAttribute('aria-readonly')).toBe('true');` | CLOSED | — |
| B063 | src/modules/media/ui/LessonAudioTracksField.test.tsx::pending row action makes the visible upload switch read-only without dropping focus | PASS | `502 \|   expect(upload.getAttribute('aria-disabled')).toBe('true');` | CLOSED | — |
| B064 | src/modules/media/ui/LessonAudioTracksField.test.tsx::upload and row switches have distinct hints, and successful upload resets the server default | PASS | `476 \|     expect(id).toBeTruthy();` | CLOSED | — |
| B066 | src/modules/media/ui/LessonAudioTracksField.test.tsx::row switch toggles with Space, sends complete metadata and keeps focus while pending/error | PASS | `441 \|   expect(toggle.getAttribute('aria-readonly')).toBe('true');` | CLOSED | — |
| B070 | src/modules/media/ui/LessonAudioTracksField.test.tsx::upload and row switches have distinct hints, and successful upload resets the server default | PASS | `474 \|   expect(new Set(descriptions).size).toBe(3);` | CLOSED | — |
| B079 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::orphan languages sort after tracks and uppercase track language never duplicates an orphan | PASS | `1165 \|     .toEqual(['track:one', 'lang:de', 'lang:fr', 'lang:uk']);` | CLOSED | — |
| B082 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::tabs select newest Ready generation of this lesson, ignoring foreign and processing videos | PASS | `1197 \|   expect([...new Set(asked.mock.calls.map(([id]) => id))]).toEqual(['newest-ready']);` | CLOSED | — |
| B083 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::tabs select newest Ready generation of this lesson, ignoring foreign and processing videos | PASS | `1197 \|   expect([...new Set(asked.mock.calls.map(([id]) => id))]).toEqual(['newest-ready']);` | CLOSED | — |
| B084 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::tabs select newest Ready generation of this lesson, ignoring foreign and processing videos | PASS | `1197 \|   expect([...new Set(asked.mock.calls.map(([id]) => id))]).toEqual(['newest-ready']);` | CLOSED | — |
| B086 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::single editor known state comes from complete inventory: loading, empty, and other language | PASS | `1202 \|   expect(f.state.known).toEqual([]);` | CLOSED | — |
| B087 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::single editor known state comes from complete inventory: loading, empty, and other language | PASS | `1208 \|   expect(f.state.known.at(-1)).toBe(true);` | CLOSED | — |
| B089 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::recovery waits for both lists before reading and choosing ownership | PASS | `1219 \|   expect(keys).not.toHaveBeenCalled();` | CLOSED | — |
| B094 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::only default track includes Default in its accessible tab name | PASS | `1232 \|   expect(screen.getAllByRole('tab').map((tab) => tab.textContent)).toEqual(['English · Default', 'Ukrainian']);` | CLOSED | — |
| B100 | src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx::unmount cancels pending und Save promise without a write | PASS | `117 \|     expect(result).toBe(mode === 'und' ? 'language_cancelled' : 'replacement_cancelled');` | CLOSED | — |
| B103 | src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx::unmount cancels pending replacement Save promise without a write | PASS | `117 \|     expect(result).toBe(mode === 'und' ? 'language_cancelled' : 'replacement_cancelled');` | CLOSED | — |
| B107 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::fresh Save rejects different id in same language and retains lines | PASS | `1244 \|   await screen.findByText('The track changed or was removed. Reload the tracks and try again.');` | CLOSED | — |
| B108 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::fresh Save rejects processing and retains lines | PASS | `1244 \|   await screen.findByText('The track changed or was removed. Reload the tracks and try again.');` | CLOSED | — |
| B121 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::thrown rename retains own recovery and lines without PUT | PASS | `1256 \|   await screen.findByText('Could not rename the track — try again.');` | CLOSED | — |
| B122 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::und restored token is never forwarded to newly chosen language | PASS | `1272 \|   expect(f.state.puts[0]?.slice(0, 3)).toEqual(['lesson', 'en', undefined]);` | CLOSED | — |
| B129 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::post-confirmation track read failure retains lines without PUT | PASS | `1283 \|   await screen.findByText('Could not verify the track — try again.');` | CLOSED | — |
| B130 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::orphan has transcript editor but no ASR controls even when draft service exists | PASS | `1294 \|   expect(screen.queryByRole('heading', { name: 'Speech recognition' })).toBeNull();` | CLOSED | — |
| B133 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::loaded ASR lines save with an explicit empty document list | PASS | `1303 \|   expect(f.state.puts[0]?.[3]).toEqual([{ id: expect.any(String), startMs: 0, endMs: 1000, text: 'Original', documentIds: [] }]);` | CLOSED | — |
| B136 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::manual recovery captures the current draft token after a draft-only refresh | PASS | `1310 \|   expect(await f.historyStore.read('author:lesson:track:one')).toMatchObject({ draftUpdatedAt: 'D1' });` | CLOSED | — |
| B137 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::handled draft offer never reappears on language editor rerender | PASS | `1323 \|   expect(screen.queryByRole('button', { name: 'Load the draft' })).toBeNull();` | CLOSED | — |
| B144 | src/modules/admin/ui/CourseEditorView.test.tsx::new unsaved lesson never invokes transcript, inventory or draft service hooks | PASS | ` 999 \|   expect(reads).toEqual([]);` | CLOSED | — |
| B149 | src/routes/route-guards.test.tsx::guarded editor forwards multilingual server default and language-specific transcript reads | PASS | `1404 \|   expect(screen.getByRole('switch', { name: 'Multiple languages' })).toHaveProperty('checked', true);` | CLOSED | — |
| B150 | src/routes/route-guards.test.tsx::guarded editor forwards multilingual server default and language-specific transcript reads | PASS | `1398 \|   await screen.findByDisplayValue('English lines');` | CLOSED | — |
| B151 | src/routes/route-guards.test.tsx::guarded rename retains previous lines while new language response waits, dirty=false | PASS | `1419 \|   expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', dirty ? 'My edit' : 'English lines');` | CLOSED | — |
| B154 | src/routes/route-guards.test.tsx::guarded fresh inventory error prevents PUT and retains edited lines | PASS | `1429 \|   await screen.findByText('Could not verify the track — try again.');` | CLOSED | — |
| B157 | src/routes/route-guards.test.tsx::guarded Save refreshes inventory and every lesson language, isolates sibling and uses latest token | PASS | `1444 \|   expect(client.getQueryState(lessonTranscriptQueryKey(f.lessonId))?.isInvalidated).toBe(true);` | CLOSED | — |
| B158 | src/routes/route-guards.test.tsx::guarded Save refreshes inventory and every lesson language, isolates sibling and uses latest token | PASS | `1445 \|   expect(client.getQueryData(lessonTranscriptQueryKey(f.lessonId, 'en'))).toMatchObject({ updatedAt: 'SAVED2' });` | CLOSED | — |
| B160 | src/app/composition/admin-draft-queries.test.ts::interval expiry restores cache without a fetch; prior snapshot=false | PASS | `268 \|     expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(` | CLOSED | — |
| B161 | src/app/composition/admin-draft-queries.test.ts::null-track placeholder does not rewrite the cache while a requested marker waits | PASS | `352 \|   expect(h.client.getQueryState(key)?.dataUpdatedAt).toBe(before);` | CLOSED | — |
| B165 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::orphan languages sort after tracks and uppercase track language never duplicates an orphan | PASS | `1165 \|     .toEqual(['track:one', 'lang:de', 'lang:fr', 'lang:uk']);` | CLOSED | — |
| B168 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::only processing videos and missing optional audio both retain the single editor | PASS | `1209 \|   await expect(mount(f)).resolves.toBeDefined();` | CLOSED | — |
| B169 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::single editor known state comes from complete inventory: loading, empty, and other language | PASS | `1205 \|   expect(f.state.known).toEqual([false]);` | CLOSED | — |
| B170 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::detached copy of orphan language is offered only to a track, preserving mounted orphan ownership | PASS | `1342 \|   await screen.findByRole('heading', { name: 'Unsaved lines from another tab (language: uk)' });` | CLOSED | — |
| B179 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::orphan replacement confirmation never performs a track read | PASS | `1334 \|   expect(f.state.reads).toEqual(['inventory']);` | CLOSED | — |
| B187 | src/modules/admin/ui/LessonLanguageTabs.test.tsx::M3-R4 options use fresh track/inventory reads, excluding und and taken languages case-insensitively; no invalid preselection | PASS | `520 \|   expect(within(dialog).queryByRole('alert')).toBeNull();` | CLOSED | — |
| B189 | src/modules/media/adapters/audio-track-repository.test.ts::rename forwards the same cancellation signal as the sixth repository argument | PASS | `357 \|   expect(rename).toHaveBeenCalledWith(videoId, trackId, 'en', 'English', true, controller.signal);` | CLOSED | — |
| B193 | src/routes/route-guards.test.tsx::guarded single-editor fallback loads the first default draft while no Ready video exists | PASS | `1477 \|   expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', 'First default draft');` | CLOSED | — |
| B195 | src/routes/route-guards.test.tsx::failed guarded PUT does not invalidate settled inventory or any lesson transcript | PASS | `1466 \|   expect(f.state.inventoryReads).toBe(before + 1); // fresh Save verification only` | CLOSED | — |

FACT: Row accounting: CLOSED=60 / DEFECT=0 / remaining=0. Every row has an attributable assertion failure in its new/extended test; collection/syntax/timeout failures count as none. Source restoration verified after every mutation. Subagents launched=0 / returned=0.
FACT: Final `git diff --name-only` exit 0; every tracked delta is a *.test.ts(x). `git ls-files --others --exclude-standard` lists only the new LessonLanguageTabs.lifecycle.test.tsx. Existing fe-a commit preserved.
```text
frontend/src/app/composition/admin-audio-queries.test.ts
frontend/src/app/composition/admin-draft-queries.test.ts
frontend/src/modules/admin/ui/CourseEditorView.test.tsx
frontend/src/modules/admin/ui/LessonLanguageTabs.test.tsx
frontend/src/modules/media/adapters/audio-track-repository.test.ts
frontend/src/modules/media/ui/LessonAudioTracksField.test.tsx
frontend/src/routes/route-guards.test.tsx
```
FACT: Production integrity against requested 5a56965 baseline and current HEAD:
frontend/src/app/composition/admin-audio-queries.ts: SHA-256 a60316fab85c4098f1ebe1e9f7744883239ee80f385dbb0acacf283593745046, byte-identical to 5a56965 and HEAD; git diff --quiet exit 0
frontend/src/app/composition/admin-draft-queries.ts: SHA-256 32ee45eaba7b9b6290e10daf1cdba3a8fc3bea2f6336dfa21b047a6006b32c10, byte-identical to 5a56965 and HEAD; git diff --quiet exit 0
frontend/src/modules/admin/ui/CourseEditorView.tsx: SHA-256 4e11966249f1c9339ba2e9580fdaa666234a4f8511c9ca0aeac28ee0115634e1, byte-identical to 5a56965 and HEAD; git diff --quiet exit 0
frontend/src/modules/admin/ui/LessonLanguageTabs.tsx: SHA-256 5751be5c1083b53ea3484f30cf2e83edb06982181838098d7ad696665ee07460, byte-identical to 5a56965 and HEAD; git diff --quiet exit 0
frontend/src/modules/media/adapters/audio-track-schema.ts: SHA-256 7a3e3e2328e568b28a22c25bf2f6ff5828a37ba8357476f55e2c028278c51b04, byte-identical to 5a56965 and HEAD; git diff --quiet exit 0
frontend/src/modules/media/application/audio-tracks.ts: SHA-256 27d7089a8e06495a7a023ecd8a722ad21a4295f30ccb1dc6d7a703d168939108, byte-identical to 5a56965 and HEAD; git diff --quiet exit 0
frontend/src/modules/media/ui/LessonAudioTracksField.tsx: SHA-256 24e71a1ed136271e51a4d1cb974a5f418c2e4a7e8e99f46cddbb8bd5c1f8e54b, byte-identical to 5a56965 and HEAD; git diff --quiet exit 0
frontend/src/routes/admin.courses.$courseId.tsx: SHA-256 7cd33bd7754ed4e0d8616e27ac2281fe1fcbbfb995c2da59c3003da2ca24fdd4, byte-identical to 5a56965 and HEAD; git diff --quiet exit 0
FACT: `git diff --stat` exit 0:
```text
 .../app/composition/admin-audio-queries.test.ts    |  85 +++++-
 .../app/composition/admin-draft-queries.test.ts    | 152 +++++++++-
 .../src/modules/admin/ui/CourseEditorView.test.tsx |  48 +++
 .../modules/admin/ui/LessonLanguageTabs.test.tsx   | 253 +++++++++++++++-
 .../media/adapters/audio-track-repository.test.ts  |  27 +-
 .../media/ui/LessonAudioTracksField.test.tsx       |  80 +++++
 frontend/src/routes/route-guards.test.tsx          | 327 ++++++++++++++++++++-
 7 files changed, 963 insertions(+), 9 deletions(-)
```
FACT: New, unstaged lifecycle test is additional to the tracked diff-stat:
```text
 .../admin/ui/LessonLanguageTabs.lifecycle.test.tsx | 120 +++++++++++++++++++++
 1 file changed, 120 insertions(+)
```
FACT: Final SHA-256 of every changed or new test file:

| Test file | SHA-256 |
|---|---|
| frontend/src/app/composition/admin-audio-queries.test.ts | `8461d556bab4d4870b4a2c09e75b589d6c5713b905e27f540507c83de470e67b` |
| frontend/src/app/composition/admin-draft-queries.test.ts | `6b513204d7b54fae82945ddf50e61cd1f92240245a5f49b03d2eb36ca7a9d2d5` |
| frontend/src/modules/admin/ui/CourseEditorView.test.tsx | `2851e08e2e9999d6b2913eecfba4aa9202c8a33afd8fbf6397240306673494af` |
| frontend/src/modules/admin/ui/LessonLanguageTabs.lifecycle.test.tsx | `6f75f5b5356e268650aa23dc41f5bba60a15db375df0576d34e592ce3d687465` |
| frontend/src/modules/admin/ui/LessonLanguageTabs.test.tsx | `d5874c37ab12be73f941e1cc0dbf81f909fd8f41af327676429376b02b5da2e5` |
| frontend/src/modules/media/adapters/audio-track-repository.test.ts | `424c6fe47a9fcc293eaa98334912e601c98a0b7176185ac761ea9b7fd7a8772b` |
| frontend/src/modules/media/ui/LessonAudioTracksField.test.tsx | `f4a5153b653b0b04fe679272903fabc5be24ebfde17d307c79b66ac1af219f1b` |
| frontend/src/routes/route-guards.test.tsx | `e649bea811a3d050fa208f7c57bfa8af91aa28b7c0c752a0fc9dda1b6148e081` |
FACT: Final ordered run `cd frontend && npx rstest run` exit 0; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/fe-b-final-v2-suite.log.
FACT: Final ordered gate COMPLETE: npm run format exit 0; npm run format:check exit 0 (403 files); npm run lint exit 0 (architecture 240 files); npm run typecheck exit 0; npx rstest run exit 0 — 109 files, 1306 passed / 0 failed / 0 skipped / 0 todo (52.957s). Final git diff --check exit 0. SHA-256 inventory rechecked after the suite with no differences.
FACT: Final result CLOSED=60 / DEFECT=0 / remaining=0; 7 tracked test files (+963/-9) plus 1 new unstaged lifecycle test (+120), total 8 test files (+1083/-9). HEAD remains 64f21ef; no commit/push, production change, dependency install, Go/Compose/e2e/live-service run or original /Users/strelia/projects/courses write. No CI/e2e/deployment claim.
NEXT: fe-b GAP closure complete. Review/apply the eight test files and this evidence; no remaining GAP id to resume.
