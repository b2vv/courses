`stage: plan` · Critic: Codex, поточний незалежний контекст · **S-backfill, ROUND 2**

План: [plan-sbackfill.md](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md), редакція 2026-11-14, SHA-256 `6aaef1e272b7224e41f754e3ab3a88d864f96cf062c6ab9ed9cdd0662282032e`.

Snapshot: `/Users/strelia/projects/courses`, HEAD `229286971f2d9b1539cbd4ed60dafdfc6b2140d6`. План і raw round-1 review — untracked; перевірені tracked media sources, migrations, migrate runner, Makefile, Compose, parent plan і report не мають змін проти HEAD. Dirty package metadata: SHA-256 `a74e0f035af81f5f21cb63d0b06e669cbc0bac1c6f60b2fd804af3b24d07dbd4`.

Prompt: `Plan critic — version 1`, `/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c/prompts/plan-critic.md`, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Package version `0.8.0-alpha.1+native-handoff.1`; metadata позначає його unreleased, current commit — unverified.

**Round-1 dispositions**

1. **B1 — NARROWED.** Умова [план:26–38](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:26) виправляє випадки created/failed upload. [PromoteAudioTrack:382–418](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:382) справді зберігає наявну default. Publish атомарно додає embedded або залишає AwaitingAudio; delete останньої Ready заборонено. Already-Ready uploaded випадок тепер явно виключено D-B6 з перевірним механізмом неоднозначності; це **звуження охоплення**, а не ремонт цих ассетів. Його не можна звітувати як відновлення всіх старих оригіналів.

2. **B2 — RESOLVED для чинних production write paths.** [План:47–53](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:47) встановлює table locks, які конфліктують із ROW EXCLUSIVE від INSERT/UPDATE/DELETE. Promote/delete виконують DB-запис перед rewrite, усередині транзакції. Goose v3.28.0 за замовчуванням парсить SQL як transactional і виконує migration та version update в одній транзакції (`provider_run.go:213–219`). Потрібно зберегти відсутність `NO TRANSACTION`.

3. **N1 — RESOLVED.** [План:106–108](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:106) правильно називає down→up reversibility.

**Нові findings**

4. **[Logical gaps / edge cases] BLOCKING — recovery після `und`-конфлікту не охоплює queued/processing. Disposition: OPEN.**

   [План:39–43](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:39) перевіряє всі non-failed uploads, але наказує оператору «скасувати той upload». Індекс [00026:6–7](/Users/strelia/projects/courses/backend/migrations/00026_audio_track_language_unique.sql:6) включає queued/processing, тоді як [CancelAudioUpload:724–732](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:724) приймає лише created/uploading.

   Обхід через rename також не працює: [metadata update:287–290](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:287) запускає rewrite; за відсутності Ready-доріжок [master writer:32–33](/Users/strelia/projects/courses/backend/internal/media/adapters/ffmpeg/master.go:32) відмовляє, і транзакція відкочується.

   **Потрібна правка:** визначити recovery окремо для created/uploading і queued/processing, включно з наслідком переходу в D-B6 при успішному завершенні; додати populated-data перевірку queued-конфлікту та обраного recovery. Fresh implementer зараз мусить вигадати цей протокол.

5. **[Test strategy] NON-BLOCKING — немає негативного доказу D-B3. Disposition: OPEN.**

   [План:100–114](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:100) перевіряє успішний інваріант, language conflict і lock timeout, але не власне відмову фінального invariant check. Додати Ready-ассет із Ready-доріжкою без default та інший backfill-кандидат; очікувати відмову, version 27 і відсутність частково доданого embedded.

6. **[Logical gaps] NON-BLOCKING — порядок timeout сформульовано суперечливо. Disposition: OPEN.**

   [План:48–53](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:48) називає LOCK першим оператором. `SET LOCAL lock_timeout='5s'` має виконатися **перед LOCK**, інакше очікування початкового lock ним не обмежене.

AI-slop та unnecessary-complexity: clean — SQL migration, goose, deterministic rollback marker і відмова від нового runtime backfill відповідають задачі.

Перечитано raw round-1 findings, parent D3–D7/§5/§8/§10, migrations 00025–00027, publish/mint/promote/delete/metadata/master та goose sources. Збережено попередні фактичні докази, але verdict переприв’язано до нового SHA. Чотири лінзи застосовано в обраному контексті; **0 launched / 0 returned**, без додаткових reviewer. Формальний незалежний four-agent gate не заявляється. Файлів не змінено; tests, network probes і graph refresh не виконувалися.

**Відкриті: 1 blocking, 2 non-blocking.**

VERDICT: BLOCKED