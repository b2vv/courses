# LB00 — де-ризик-спайк (жива нота спайка; викидний код)

ДАЛІ: Codex виконує спайк, дописує докази нижче; я вношу go/no-go.

## Мета

Зняти ДВА архітектурні ризики, що єдині можуть змусити переписати план (GATE 2: найгостріше —
Tiptap-поза-бандлом). Решта Q9-пунктів (image-serve під auth, клавіатура 200%, CSP/XSS, тіло-не-в-
outline) простежені дослідниками конкретно й ловляться приймальними сценаріями у своїх тасках —
це correctness, не плано-змінні невідомі; тут НЕ спайкуються.

## Ризик A — ізоляція бандла (НАЙГОСТРІШИЙ)

**Твердження плану:** Tiptap у `React.lazy`-піддереві admin-редактора авто-спліситься в окремий
async-chunk і НЕ потрапляє ні в спільний main `index.*.js`, ні в learn-chunk — ЯКЩО його не вплести
в `createAdminDependencies`/router-контекст (живий контрприклад: hls.js протік у 1.64MB main, бо
композиційний корінь `composition/learn.ts:7,29` статично його імпортує).

**Метод:** реальний Tiptap-деп (лишається для LB12) + викидний admin-route з `React.lazy`-редактором,
НЕ вплетений у composition root; prod-білд; grep chunk'ів.

**Критерій go:** `tiptap`/`prosemirror` ВІДСУТНІ в main `index.*.js` і в learn-route-chunk; присутні
лише в окремому async-chunk. no-go: протікають у main/learn.

## Ризик B — round-trip власної схеми ↔ Tiptap + версіонування

**Твердження плану:** власна валідована rich-text-модель (не сирий ProseMirror JSON) переживає
serialize→store(jsonb)→load→render→edit без втрат, і schemaVersion дає міграцію.

**Метод:** мінімальний TS-тип власної моделі + Tiptap↔модель-мапер; тест на зразковому уроці
езотерики (video + text + row[image|text] + task): serialize→JSON.parse(JSON.stringify)→validate→
render-to-HTML (чиста ф-ція) → edit → re-serialize; + shim міграції v1→v2.

**Критерій go:** документ валідний після round-trip, render-HTML семантичний, міграція v1→v2 проходить.

---

## Докази (заповнює Codex)

### A — grep chunk'ів + розміри

2026-10-12, Codex. `npm install @tiptap/core @tiptap/pm @tiptap/starter-kit @tiptap/react`
— exit 0, додано 49 пакетів. Усі чотири прямі пакети **3.31.4, MIT** (перевірено їхні
локальні package.json). npm повідомив 13 audit vulnerabilities (1 moderate, 10 high,
2 critical); належність новим/наявним залежностям не досліджувалась, auto-fix не запускався.

`cd frontend && npm run build` — **exit 0**. Build містить попередження, що
`routes/route-guards.test.tsx` не експортує Route; збірка успішна.
Викидний `admin.spike-editor.tsx` → React.lazy → `SpikeEditor.tsx`; жодного нового
імпорту в composition/router context. Build автоматично оновив `src/routeTree.gen.ts`.

Ідентифікація chunk'ів за фактичними dynamic imports у minified main:

- `/learn/$`: `Promise.all([r.e(157),r.e(938)]).then(r.bind(r,7393))`.
  Основний lesson-route — **938.2af9adcb2d.js, 29359 bytes**, спільний async dependency
  **157.b1368bdfd8.js, 9931 bytes**. Error component — 180.fc7b73d4e5.js.
- `/learn`: `r.e(243).then(r.bind(r,8166))` — **243.25b397d655.js, 2056 bytes**;
  error component — 946.a9e3556386.js.
- `/admin/spike-editor`: `r.e(263).then(r.bind(r,7650))`;
  263.78f34dd7d2.js містить `lazy(()=>Promise.all([_.e(807),_.e(308)]).then(_.bind(_,9159)))`.
  **807.e6f067a740.js, 398152 bytes** — редакторські бібліотеки;
  **308.3550ffd678.js, 441 bytes** — компонент редактора.
- Main **index.72fbd846c4.js, 1645190 bytes**.

Фактичний stdout (case-insensitive grep filenames; `.LICENSE.txt` не включено):

```text
$ grep -il tiptap dist/client/assets/js/index.*.js dist/client/assets/js/async/*.js
dist/client/assets/js/async/807.e6f067a740.js
exit=0
$ grep -il prosemirror dist/client/assets/js/index.*.js dist/client/assets/js/async/*.js
dist/client/assets/js/async/807.e6f067a740.js
exit=0

$ wc -c dist/client/assets/js/index.*.js dist/client/assets/js/async/*.js
 1645190 dist/client/assets/js/index.72fbd846c4.js
    9931 dist/client/assets/js/async/157.b1368bdfd8.js
    1169 dist/client/assets/js/async/162.9972ff0610.js
    3553 dist/client/assets/js/async/178.7b33dec4cf.js
     170 dist/client/assets/js/async/180.fc7b73d4e5.js
    1455 dist/client/assets/js/async/181.0f48812d86.js
     291 dist/client/assets/js/async/215.eee6cd5b44.js
    2056 dist/client/assets/js/async/243.25b397d655.js
     356 dist/client/assets/js/async/263.78f34dd7d2.js
    5101 dist/client/assets/js/async/307.74347855c7.js
     441 dist/client/assets/js/async/308.3550ffd678.js
    5520 dist/client/assets/js/async/325.209aa4626e.js
    1424 dist/client/assets/js/async/355.8747bb8943.js
     170 dist/client/assets/js/async/369.b43a378e5b.js
     170 dist/client/assets/js/async/384.e95816a337.js
     402 dist/client/assets/js/async/417.8095b6dfe2.js
     170 dist/client/assets/js/async/463.3edf397035.js
    5046 dist/client/assets/js/async/474.142735dc56.js
    7527 dist/client/assets/js/async/505.73f1602fdf.js
    3005 dist/client/assets/js/async/506.b90d18be36.js
     170 dist/client/assets/js/async/511.4ff7e23392.js
     167 dist/client/assets/js/async/515.95b6d59930.js
     170 dist/client/assets/js/async/535.3e4521d035.js
    5106 dist/client/assets/js/async/537.0aba8b30f3.js
    1343 dist/client/assets/js/async/538.d2fff2863f.js
     484 dist/client/assets/js/async/559.d521c13b00.js
     170 dist/client/assets/js/async/561.47db0311e7.js
     169 dist/client/assets/js/async/568.bf2948b949.js
   12389 dist/client/assets/js/async/573.176654b61a.js
    6840 dist/client/assets/js/async/577.41f96838ca.js
     173 dist/client/assets/js/async/592.8cc91805ff.js
     168 dist/client/assets/js/async/65.0f567599f9.js
     290 dist/client/assets/js/async/69.f9b6e845d5.js
    1441 dist/client/assets/js/async/697.152dabb558.js
    1245 dist/client/assets/js/async/726.c30a25b15e.js
   16248 dist/client/assets/js/async/764.befe7f7275.js
  398152 dist/client/assets/js/async/807.e6f067a740.js
    1657 dist/client/assets/js/async/815.908610ffa6.js
     496 dist/client/assets/js/async/851.afe85b3116.js
    3570 dist/client/assets/js/async/923.b6d52e82e4.js
   29359 dist/client/assets/js/async/938.2af9adcb2d.js
     432 dist/client/assets/js/async/946.a9e3556386.js
    5518 dist/client/assets/js/async/95.fc73177dab.js
    1904 dist/client/assets/js/async/963.90a186f6d5.js
 2180808 total
```

**A: критерій GO виконано** на цьому prod-build: обидва маркери лише у 807 async;
у main, усіх learn chunks та решті client JS збігів немає. Це статична bundle-перевірка,
не браузерний network trace / CSP-тест.


### B — round-trip + міграція

2026-10-12, Codex. Написані, але **не верифіковані** викидні файли:
`frontend/src/presentation/admin/spike-document.ts` (власні типи, unknown-input validator,
мапер у два боки, pure HTML renderer, демонстраційна міграція v1→v2 з полем title) і
`spike-document.test.ts` (4 тести: round-trip + реальний Tiptap Editor/edit + migration,
XSS escaping/drop href, межі валідатора, мапінг/відмова unsupported editor nodes).
Це лише JSON.stringify/parse-модель збереження, **PostgreSQL jsonb не запускався**.

**BLOCKED — тестовий runner не стартував.** Єдина спроба:
`cd frontend && npm run test -- spike-document`, exit **1**:

```text
> @courses/frontend@0.1.0 test
> rstest run spike-document

error   Failed to run Rstest.
error   listen EPERM: operation not permitted ::1:3000
    at Server.setupListenHandle [as _listen2] (node:net:1918:21)
    at listenInCluster (node:net:1997:12)
    at GetAddrInfoReqWrap.callback (node:net:2206:7)
    at GetAddrInfoReqWrap.onlookupall [as oncomplete] (node:dns:134:8)
```

Це заборона bind/listen в поточному середовищі, ще до виконання тестів; результату
pass/fail для жодного assertion немає. За прямою інструкцією власника «If anything
blocks you ... STOP» виконання зупинене без retry/обходу обмежень. Питання для
відновлення: чи можна запустити цей самий targeted Rstest у середовищі, де дозволено
слухати `::1:3000`?

Фактичного renderToHtml output snippet **немає**: тест, що друкує `LB00 HTML:`,
не виконався. Очікуваний фрагмент у тесті (не доказ виконання):
`<h2><strong>Практика</strong></h2>`.

Фінішні `npm run format` і `npm run typecheck` **не запускались через STOP на блокері**;
файли залишені як є, форматування й типи не підтверджені. Build A відбувся до
створення B-файлів і не є доказом їхньої коректності. Повний gate/e2e не запускався.

### Підсумок і рекомендація

2026-10-12, Codex: ізоляція bundle A підтверджена (398152-byte Tiptap/ProseMirror chunk
поза 1645190-byte main та learn chunks), тож lazy admin subtree є придатним шляхом
для продовження. **Загальний LB00 поки BLOCKED, B не доведено** через Rstest
`listen EPERM ::1:3000`; рекомендую відкласти загальний go/no-go до запуску написаних
тестів, перевірки фактичного HTML, обов'язкових format/typecheck у дозволеному
середовищі. Tiptap dependency, throwaway route/editor/model/tests та згенерований
route tree залишені для огляду; нічого не закомічено. Остаточне рішення нижче — власника.


## Go / No-go (вношу я після доказів)

**2026-10-12 — GO.** Обидва плано-змінні ризики зняті:

- **A (GO):** prod-build + grep — `tiptap`/`prosemirror` лише в async-chunk `807` (398 152 B);
  відсутні в main `index.72fbd846c4.js` (1 645 190 B) і в learn-chunk'ах (`938` 29 359 B, `243`
  2 056 B). Lazy admin-піддерево ізолює редактор, якщо не вплітати в composition root. (Codex-run
  під пісочницею.)
- **B (GO):** цільовий rstest `spike-document` — **4/4 passed** у цій сесії (round-trip + реальний
  Tiptap edit + міграція v1→v2; XSS escape + drop `javascript:`; межі валідатора; двобічний мапінг).
  `listen EPERM ::1:3000` був суто пісочницею Codex, не дефект тесту.

**Валідований контракт схеми** (переноситься у LB01 Go-домен / LB11-12 фронт): типи `LessonDocument`/
`Section`/`RichText`/`Span` з spike-document.ts; валідатор (0..1 video, row 1..3 col depth-1
text/image, heading 2|3, href лише http(s), schemaVersion). Round-trip Tiptap↔власна модель
доведено, тож «не зберігати сирий ProseMirror JSON» (Codex-рішення) — реалізовне.

**Прибрано після go:** весь викидний код (route/editor/model/test) + Tiptap-деп + regen routeTree
відкочено — Tiptap додамо в LB12 «коли вжито» (AGENTS: deps only when used; уникаємо lockfile-дрейфу
macOS→linux). Знання лишається тут. Далі — LB01 (Go-домен документа).
