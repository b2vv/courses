claude-opus-4-8

ДАЛІ: читаю docs/accounts/plan.md §4d (CSP-директиви, nginx-скоуп), §6 (формулювання), spec.md AC14, tasks.md A11; потім звіряю NB1 і NB2 з кодом (server.mjs, nginx conf).

ФАКТ NB1 — ЗАКРИТО (з застереженням):
- plan §4d рядок 135-136: директива містить `style-src 'self' 'unsafe-inline'` (і script-src без unsafe-inline, nonce). Підтверджено.
- §6 рядок 171-172: формулювання звужено до «script-src без unsafe-inline», style-src 'unsafe-inline' як назване обмеження. Підтверджено.
- Премиса (inline-стилі існують) підтверджена кодом: 4 inline style= у прод-src — data.tsx:17 (ProgressBar width), data.tsx:42 (Meter width), CourseView.tsx:230 і ReviewBlock.tsx:137 (whiteSpace:'pre-line'). style-src 'unsafe-inline' покриває style-атрибути (CSP2/3 fallback). OK.
- Застереження (non-blocking): план називає лише ProgressBar/Meter + React Aria оверлеї; дві whiteSpace:'pre-line' не перелічені, але висновок (лишити 'unsafe-inline') їх покриває — дефекту нема.

ФАКТ NB2 — ЗАКРИТО:
- plan §4d рядки 140-142: nginx-CSP лише на /api, /media та error-локаціях, НЕ на location / (SSR). Підтверджено текстом.
- nginx.conf: server-level add_header лише X-Content-Type-Options/Referrer-Policy/X-Frame-Options (49-51); location / (153-166) не має власного add_header, CSP там не додається → на SSR лишиться тільки nonce-CSP із server.mjs. Один заголовок. Коректно уникає AND-подвоєння.
- server.mjs сьогодні CSP не ставить (лише Cache-Control/Vary, рядки 23-24) — тобто сьогодні CSP на SSR взагалі нема; T11 його додасть у server.mjs. Узгоджено.
- Застереження (non-blocking, nginx add_header inheritance gotcha): якщо CSP add_header класти всередину location ^~ /api/ (яка сьогодні НЕ має власного add_header), ця локація перестане успадковувати три server-level security-заголовки — їх треба повторити. /media/p/ їх уже повторює (66-68). Деталь імплементації T11, не блокує план.

ФАКТ NB3 — НОВИЙ BLOCKING (регрес в уже-закриту Phase 6 playback):
- Директива §4d (135-136) = `default-src 'self'; script-src 'self' 'nonce'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'`. НЕМАЄ `media-src` і `worker-src` → fallback на default-src 'self' (без blob:).
- Код playback (Phase 6, e2e 74/74):
  - adapters/playback/hls-media.ts:17 `new Hls({ enableWorker: true })` → hls.js на Chromium (гілка Hls.isSupported(), :83) прикріплює MediaSource через `URL.createObjectURL(mediaSource)` → video.src = `blob:` → керується media-src. Під `default-src 'self'` без `media-src blob:` MSE-attach блокується → playback у Chromium (основний e2e-браузер) ламається.
  - enableWorker:true → hls.js створює Web Worker з blob URL → `worker-src`/`child-src` (fallback default-src 'self') блокує spawn → CSP-порушення в консолі (hls.js падає на main-thread fallback, але порушення лишається).
  - LessonPlayer.tsx:424 `URL.createObjectURL(new Blob([captions],{type:'text/vtt'}))` → `<track src=blob:>` → media-src → блокується.
- Наслідок: директива §4d у enforcing-режимі порушує Acc25 («e2e-консоль без CSP-порушень») і ламає вже-закриту Phase 6. Бракує щонайменше `media-src blob:` і `worker-src blob:` (та/або `child-src blob:`). Fallback report-only (145-146) обумовлений лише «дорогим nonce», blob:-проблему не покриває; у report-only замір порушень не буде нульовим.
- Сьогодні CSP на SSR відсутній (підтверджено) → це саме НОВИЙ регрес, внесений правкою CSP, а не наявний стан.

ВЕРДИКТ: ЧЕРВОНИЙ. NB1 закрито, NB2 закрито, але відкрито новий blocking NB3 (директива §4d неповна для реальних ресурсів фронту: бракує media-src blob: / worker-src blob: для hls.js MSE+worker та blob: VTT-субтитрів — ламає Phase 6 playback і Acc25). Гейт лишається червоним, поки NB3 відкрито.
