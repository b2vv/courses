claude-opus-4-8

# GATE 2 — accounts CSP NB3 фінальний звіряльний прохід

ДАЛІ: перечитав plan.md §4d/§6/Acc25 і tasks A11. Далі звіряю директиву CSP у §4d проти реального коду playback (hls.js blob/worker, VTT blob) — файли adapters/playback/hls-media.ts, learn/LessonPlayer.tsx; оцінюю report-only-first як обмеження з магнітудою; вирішую blocking для старту A01.

ФАКТ (директива): plan.md:136 містить рівно `media-src 'self' blob:; worker-src 'self' blob:; connect-src 'self'` у повній директиві (разом із default/script-nonce/style-unsafe-inline/img data:/object-none/base-uri/frame-ancestors). NB3-твердження про текст директиви — правдиве.

ФАКТ (hls.js worker): frontend/src/adapters/playback/hls-media.ts:17 `new Hls({ enableWorker: true })` → blob-Web-Worker → потребує worker-src blob:. Підтверджено.

ФАКТ (hls.js MSE blob): hls-media.ts:18 `hls.attachMedia(video)` + :37 `hls.loadSource` — MSE-шлях; hls.js чіпляє MediaSource через `URL.createObjectURL` → blob: як src <video> → media-src blob:. Сегменти тягне тим самим origin (/media/p/<id>/ за AGENTS — nginx, той самий origin) → connect-src 'self'. Підтверджено.

ФАКТ (VTT blob): frontend/src/presentation/learn/LessonPlayer.tsx:424 `URL.createObjectURL(new Blob([captions], { type: 'text/vtt' }))` → blob: для <track> → media-src blob:. Підтверджено.

ФАКТ (повний e2e): plan.md:141-142 «CSP — зміна, що зачіпає весь застосунок, тож повний e2e включно з playback обов'язковий»; Acc25 (plan.md:222) «повний e2e (вкл. playback/медіа) … консоль без CSP-порушень»; §5 (161) і A11 (tasks:303) повторюють. Названо обов'язковим.

ФАКТ (метод report-only-first): plan.md:151-156 — вмикаємо Content-Security-Policy-Report-Only з цільовою директивою, ганяємо повний e2e, збираємо реальні порушення з report у консолі/endpoint → емпірично ловить будь-яку пропущену директиву (font/connect/…). Fallback: якщо enforcing ризиковий — фіксуємо report-only + число/перелік порушень як обмеження (магнітуда). Report-Only — реальна можливість браузера (SecurityPolicyViolationEvent/report-uri), Playwright читає консоль → механізм реалізовний. Fallback дає ЧИСЛО, не обіцянку → валідне «обмеження з магнітудою/механізмом», не відмахування.

РІШЕННЯ:
- NB3 ЗАКРИТО. Директива має media-src/worker-src blob: + connect-src 'self'; усі три використання (worker, MSE-blob, VTT-blob) підтверджені в коді на названих рядках; повний e2e з playback названо обов'язковим у 4 місцях.
- report-only-first адекватно знімає ризик ще-не названих директив: емпіричний замір на повному e2e перед enforcing; fallback із числом порушень — коректне обмеження.
- Директива сама по собі playback не ламає: object-src 'none'/frame-ancestors 'none'/base-uri 'self' playback не чіпають; img data: покриває постери; відсутній font-src падає на default-src 'self' (self-hosted) і однак ловиться report-only-нетом.
- Blocking для СТАРТУ імплементації: НЕМАЄ. A01 (platform/mail+Mailpit) незалежний від CSP; CSP — окремий A11 із власним емпіричним кроком; план визначено достатньо, щоб починати.

ВЕРДИКТ: ЗЕЛЕНИЙ. Жодної відкритої blocking-знахідки. NB3 закрито; план фази accounts готовий до старту імплементації; A11 (CSP) несе власний report-only-first замір як назване обмеження.
