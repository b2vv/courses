
## Лінза «логічні діри» — старт 2026-09-19

ДАЛІ: прочитати docs/asr/spec.md цілком, потім AGENTS.md, потім перевірити кожну цитату координат
у backend/internal/transcript/, backend/internal/media/adapters/asynq/, backend/cmd/media-worker/main.go,
docker-compose.yml — вмістом рядка (sed -n 'Np'), не діапазоном.

ФАКТ: spec.md 976 рядків. Прочитано 1–340. Структура: §1 проблема, §2 факти дерева + §2.1 поправка
(2 ГіБ — рядок конфігу), §2.2 база образу bookworm, §3 звуження по §24.20 (знято A3), §4 сценарії,
§5 таблиця відповідей A1–A8, §6 розбір рушіїв (+ §6.5 замір, §6.8 поправка з повторами).
ФАКТ: у §6.5 ПОПРАВКА: числа §6.5 «не є опорою для рішень» (розкид до +60%), виправлена таблиця §6.8.
ПИТАННЯ (кандидат на знахідку): §6.6 «Ухвалив власник ... **за таблицею §6.5**» — а §6.5 сама себе
оголосила неопорною. Те саме §5 рядок A1 «закрите за заміром §6.5». Треба звірити з §6.8.
ПИТАННЯ: §5 A5 «38 хв при 6 CPU / 8 ГіБ (§6.8)» проти §6.5, де 38 хв — це one-shot 77.6 с; повтори
дали 112–124 с → ~55–61 хв. Перевірити §6.8.
ПИТАННЯ: §6.4 замір задумано «--cpus=2 --memory=2g», а §6.5 заміряно на 6 CPU. §2.1 пояснює зміну меж,
але перевірити, чи §6.4 не лишилась як чинний припис.
ДАЛІ: читати 340–700.

ФАКТ: прочитано 340–976 (кінець). Секції: §6.6 наслідки, §6.8 поправки, §7 A3/A8 (7.1–7.10, порядок
7.5,7.6,7.7,7.9,7.10,7.8), §8 A5, §9 A7 (9.1–9.4, 9.6, 9.5), §10 RabbitMQ, §11 out of scope.

КАНДИДАТИ (до перевірки):
K1 §6.8 «mem_limit ≈4 ГіБ і це дає 38 хв, а не 109» — але в таблиці двома рядками вище саме
   `--memory=4g` дало 109 хв. 38 хв заміряно ЛИШЕ на 8g. §9.6 тягне ≈4 ГіБ далі.
K2 §6.6 таблиця: «mem_limit не менше ~2.5 ГіБ» — мертве після §6.8 (4g→109 хв), не поправлено.
K3 §6.6:348 «cpus обирається з кривої §6.7» — §6.7 У ДОКУМЕНТІ НЕМАЄ; криву §2.1 замовила, ніхто
   не зняв. Усі числа — 6 CPU.
K4 §6.4 замір задумано на ДВОХ рушіях; faster-whisper не міряний ніколи, а A1 закрите «за заміром».
   §2.2 обрала bookworm саме щоб не закрити faster-whisper до заміру.
K5 §2.2 «чим перевірити»: «якщо A1 зупиниться на whisper.cpp — рішення можна переглянути». A1
   зупинилось на whisper.cpp; наслідок не доведено.
K6 §6.5 ПОПРАВКА (:309-313) «числа §6.5 не є опорою для рішень» — §6.8 скасувала її підставу, але
   рядок лишився; §6.6/§5 при цьому спираються «за таблицею §6.5».
K7 §7.3 (аргумент про перемонтування редактора) розвалений §7.10: автор тепер редагує ЧЕРНЕТКУ,
   а ASR у неї ж пише → колізія повернулась через updatedAt чернетки.
K8 §4.3 «структурно істинне» — після §7.10 робота автора може жити в чернетці, яку перегенерація
   переписує.
K9 Ніде не сказано, ХТО ставить задачу ASR у чергу при готовності ассету (§4.1 «починається само»).
   Перевірити, чи є в media подія/аутбокс.
K10 §6.4 визначає RTF як «секунд обробки на секунду аудіо», а таблиці рахують навпаки (швидкість).
K11 §9.3 `cmd/transcript-worker` vs §9.6/§9.5 `cmd/asr-worker` — дві назви одного виконуваного.
K12 §8 цілком стоїть на «5–6 годин»; після §6.5 підстава зникла, поправки в §8 нема (§7.1 «наслідок
    числа з §8», §7.3 «п'ятигодинна задача» — теж).
K13 §11 «Воркер має 2 CPU» — мертве після §2.1.
ДАЛІ: перевіряти КОЖНУ координату вмістом рядка.

## Перевірка координат вмістом (sed -n 'Np')

ЖИВІ (збіг точний):
- docker-compose.yml:91-92 → `cpus: 2.0` / `mem_limit: 2g`, у блоці `media-worker:` (:76..:100) ✓
- media/adapters/asynq/queue.go:41 → `asynqlib.Timeout(30*time.Minute),` ✓
- cmd/media-worker/main.go:70 → `Concurrency: 1, Queues: map[string]int{"media": 1}, ...` ✓
- AGENTS.md:19 → «Ports require an actual external dependency…» ✓ ; :20 → «No global mutable clients/stores.» ✓
- backend/Dockerfile:16 → `USER 10001:10001` ✓ ; frontend/Dockerfile → node:24-bookworm-slim ✓
- transcript/domain/transcript.go:14 → MaximumSegments = 5000 ✓ ; :19 → MaximumTextRunes = 1000 ✓
- transcript/adapters/postgres/repository.go:22 → `FROM video_assets AS assets` ✓
- PRODUCT-BRIEF.md:551 → «Підготуй port для майбутнього transcription adapter.» ✓ ; :1015 → §24.19 ✓
- transport.ts:88 → `credentials: 'same-origin',` ✓ ; :96 → `return { ok: false, code: control.failureCode(error) };` ✓
- use-transcript-recovery.ts:21 → `export const idleCaptureMs = 90_000;` ✓
- __root.tsx:87 → `function Layout() {` ✓ ; :147 → `<Outlet />` ✓
- tus-client.ts:77 → `function browserStorage()…` ✓
- identity.ts:69 → `export function safeReturnTo(` ✓ ; identity.test.ts:54 → `['a recursive login target', '/login'],` ✓
- playback/adapters/redis/ownership_writer.go існує; api/media-worker/upload — один і той самий redis ✓

ЗГНИЛІ:
- TranscriptEditor.tsx:54 → насправді `expectedUpdatedAt,` (проп). useState для рядків — :61.
- LessonTranscriptField.tsx:59 → насправді коментар; `key={`${lessonId}-${expectedUpdatedAt ?? 'new'}`}` — :71.
- AGENTS.md:16 → «Application: domain and ports only…» — про транзакції там нічого немає.
- §6.6:348 «крива §6.7» → секції 6.7 у документі немає (є 6.1–6.6, 6.8).

СУПЕРЕЧАТЬ КОДУ:
- §7.1/§7.2 «правки гинуть мовчки, без localStorage/автозбереження» — у TranscriptEditor.tsx:69-73
  уже змонтовано useTranscriptRecovery (baseline з expectedUpdatedAt, offer-stale), тобто f1b98de/85c9f31,
  які сама ж спека згадує в §7.7/§7.9. Обидва коміти в git є.
- §9.2 «не треба нічого нового, щоб дізнатися шлях до ассету» — запит transcript-репозиторію читає
  ЛИШЕ `assets.generation` (repository.go:20-26); шлях — це media/domain.StorageKey{Kind,ID,Leaf:"original"}
  (upload.go:230) плюс резолвер media/adapters/filesystem/storage.go:316-334. Ні id ассету, ні розкладки
  модуль transcript не має.
- ніхто не ставить задачу в чергу: єдина згадка тригера — :483 «її запускає готовність ассету»;
  у media немає ні шини подій, ні аутбоксу (transcode.go:187 Enqueue — це власний ретрай медіа).
- §9.5:920 образ asr-worker = «Go + whisper-cli + модель», без ffmpeg, хоч §2:37 і §8:806 спираються
  на «ffmpeg уже в образі» (це образ media-worker).

РІШЕННЯ: 8 blocking + 9 non-blocking, звід нижче в фінальній відповіді.
