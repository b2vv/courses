# Plan — multi-audio S-backfill (embedded-доріжка для старих Ready-ассетів)

Base `origin/main@2292869` (S5-L2b злито, PR #103). Гілка `multi-audio-sbackfill`. Writer = Courses 1,
review_mode = independent_subagent (Codex `exec` напряму). Вибір зрізу — власник, 2026-11-14 («S-backfill»).
Джерело вимоги: `plan.md` §5 (S-backfill), §10 (rollback), D3/D4/D5.

## Проблема (виміряно на коді, 2026-11-14)
Таблиця `audio_tracks` з'явилась у 00025 без backfill. Embedded-рядок матеріалізує лише `Publish` при
транскоді **після S3** (`postgres/publish.go:86`). Отже, кожен Ready-ассет, транскодований до S3, має
**нуль** рядків доріжок. Наслідки на `main` зараз (S5 злито — UI і upload живі):
1. **Втрата оригінального звуку в плеєрі.** Перша вивантажена доріжка на такому ассеті → `PromoteAudioTrack`
   робить її default і перегенеровує master **лише з Ready-доріжок у БД** (`application/transcode.go:259`
   `audioRenditions`, D7) → `EXT-X-MEDIA` на `audio/index.m3u8` (вбудований звук) зникає з master.
2. Редактор показує порожній список доріжок: автор не може перейменувати чи зробити default оригінал.
3. D4 «Ready ⇒ ≥1 Ready-доріжка» хибний для старих даних (S6-ASR візьме вхід із default-доріжки — її нема).

**Чому кожен такий ассет має звук (механізм, не припущення):** до S3 `AcceptSource`
(`df9a313^:backend/internal/media/domain/source.go:50-56`) відхиляв джерело без AAC/MP3-аудіо, тож
Ready-ассет до S3 завжди мав аудіопотік, а транскодер завжди пакував його як `audio/index.m3u8`
(`group_aac`). Після S3 Ready-ассет без доріжок виникнути не може: `Publish` з аудіо вставляє embedded,
video-only іде в `awaiting_audio`, delete останньої Ready-доріжки заборонено.

## Рішення (2026-11-14, author Courses 1)

### D-B1 — Data-міграція `00028_audio_track_embedded_backfill.sql`
Для кожного `video_assets` зі `state='ready' AND evicted_at IS NULL`, у якого **немає жодної Ready**
доріжки (`NOT EXISTS … state='ready'`), вставити embedded-доріжку рівно як `embeddedTrack()` (`transcode.go:478`):
`kind='embedded', language='und', label='Original', is_default=true, state='ready', source_leaf='original'`,
`created_at=updated_at=now()`, байтові колонки — дефолти 00027.
- **Умова «нема Ready-доріжки», не «нуль рядків» і не «нема embedded» (critique #1, 2026-11-14).**
  Механізм: після S3 Ready-ассет **завжди** має ≥1 Ready-доріжку — `Publish` вставляє embedded разом із
  Ready (`publish.go:86`), video-only стає Ready лише з першою Ready uploaded (`PromoteAudioTrack`), delete
  останньої Ready заборонено (`DeleteAudioTrack`). Отже «Ready без Ready-доріжки» ⇒ ассет до S3 ⇒ мав звук.
  Це покриває старий ассет, на який після S5 уже почали вантажити доріжку (`MintAudioTrack` не перевіряє
  embedded): незавершена/failed uploaded-доріжка лишається, embedded додається поруч і стає default;
  коли uploaded згодом стане Ready, `PromoteAudioTrack` зберігає наявну default (`audiotrack.go:403-405`).
  «Нема embedded» відкинуто: embedded можна навмисно видалити, коли є інша Ready (тест
  `audiotrack_integration_test.go:570-589`) — такий ассет коректний, дописувати embedded не можна.
- **Конфлікт мови `und` (critique #2 F4) — без ручного ремонту.** Якщо на такому ассеті не-failed
  uploaded-доріжка вже має `und` (індекс `audio_tracks_language_per_asset` включає й queued/processing,
  а `CancelAudioUpload` скасовує лише created/uploading — `audiotrack.go:724-732`), backfill бере мову
  `und-x-original`: приватний BCP 47-тег (валідний для HLS `LANGUAGE`), проходить серверне правило мови
  (`domain/audiotrack.go:267`, ASCII-літери/цифри/дефіс ≤35), а з UI-переліку (`uk,en,pl,de,fr,es,und`) його
  обрати неможливо — фронтенд покаже його як є (легасі-гілка `languageName`). Автор може перейменувати.
  Закриває «впасти й чекати оператора» (для queued/processing скасування нема, rename відкочується без
  Ready-доріжок — `master.go:32-33`). Лише якщо зайняті **обидві** мови (`und-x-original` можна поставити
  тільки прямим API-викликом) — `RAISE EXCEPTION` з id ассета; магнітуда: потрібні два навмисні API-
  виклики з цими тегами на старому ассеті між S5 і міграцією, у test mode.
- **Серіалізація з живими writers (critique #2):** `make migrate` не зупиняє сервіси (`Makefile:13-17`),
  тож міграція починається з `SET LOCAL lock_timeout = '5s'` **і одразу за ним** (critique #2 F6 — інакше
  очікування самого LOCK не обмежене) `LOCK TABLE video_assets, audio_tracks IN SHARE ROW EXCLUSIVE MODE`.
  Міграція лишається транзакційною (без `-- +goose NO TRANSACTION`): goose виконує її й запис версії в
  одній транзакції.
  Режим конфліктує з `ROW EXCLUSIVE` (будь-який INSERT/UPDATE/DELETE mint/promote/delete/publish/evict),
  тому перевірка `NOT EXISTS`, INSERT і D-B3 бачать один знімок, а записувачі чекають коміту міграції.
  Promote/delete перегенеровують master у своїй транзакції після DB-запису (`rewrite`-callback), тож
  regen, що йде після міграції, читає вже backfilled embedded. Якщо writer тримає таблицю довше за 5 с,
  міграція падає з lock timeout і повторюється — тихого часткового стану нема.
Альтернатива «backfill у Go при старті worker» відкинута: міграція виконується один раз до сервісів,
транзакційна з перевіркою (D-B3) і доводиться тим самим goose-провайдером, що й 00015.

### D-B2 — Детермінований id як маркер для down (замість «за ознакою Kind:embedded»)
`plan.md` §10 вимагав, щоб down прибирав **лише** згенеровані рядки. Ознака `kind='embedded'` цього не
забезпечує: пост-S3 embedded-рядки (від `Publish`) мають той самий kind, і down видалив би їх теж.
Механізм: id backfill-рядка = `md5('multi-audio-backfill:' || asset_id)::uuid`. Down видаляє рівно
`WHERE kind='embedded' AND id = md5('multi-audio-backfill:' || video_asset_id::text)::uuid`.
Up ідемпотентний (`NOT EXISTS` + `ON CONFLICT (id) DO NOTHING`). Колізія з `freshUUID` (v4 random) —
ймовірність ~2⁻¹²², і в разі колізії `ON CONFLICT` пропустить вставку, а D-B3 зупинить міграцію
(гучно, не тихо). Перевірка: integration-тест down прибирає backfill-рядок і лишає пост-S3 embedded.

### D-B3 — Активація Ready-ґейта = перевірка інваріанта в тій самій міграції
`plan.md` §5 вимагає злити backfill «разом з активацією Ready-ґейта, не раніше». Для **нових** даних
ґейт D4 уже активний з S3 (Publish/AwaitingAudio/promote/delete-guard, транзакційно). Для **наявних**
даних ґейт активується цим зрізом: після INSERT міграція в `DO`-блоці перевіряє, що **кожен**
`ready AND evicted_at IS NULL` ассет має рівно одну default Ready-доріжку і ≥1 Ready-доріжку, і
`RAISE EXCEPTION` інакше → міграція (і `make up`) падає з названою причиною, транзакція відкочується.
Альтернатива «DB constraint trigger на двох таблицях» відкинута: інваріант уже тримають транзакції
застосунку (D4), а deferred-тригери на `video_assets`+`audio_tracks` зачепили б кожну транзакцію
publish/promote/delete/evict — магнітуда: 4 write-шляхи (`publish.go`, `PromoteAudioTrack`,
`DeleteAudioTrack`, eviction) плюс їхні integration-тести, без нового захисту понад наявний.

### D-B3a — Діагностика відмови (critique #3 F7)
Названа помилка (`RAISE EXCEPTION` з id ассета) видна лише integration-тестам через goose-провайдер:
продакшн-runner навмисно замінює помилку goose загальним повідомленням (`cmd/migrate/main.go:67-69`,
правило «без секретів у логах») — runner не міняємо. Оператор при відмові `make migrate` запускає
read-only запити, що повторюють перевірки міграції (у report, § S-backfill «Діагностика»): (1) Ready
ассети без Ready-доріжки, у яких зайняті обидві мови `und`/`und-x-original`; (2) Ready ассети, у яких
після backfill не було б рівно однієї default Ready-доріжки; (3) `pg_locks` на `video_assets`/`audio_tracks`
для lock timeout.

### D-B4 — master.m3u8 на диску не чіпаємо
Backfill — лише БД. Master старого ассета (ffmpeg-ів, `NAME="audio_0"`) уже вказує на `audio/index.m3u8`
і валідний; перший regen (upload/rename/default/delete доріжки) перепише його з БД, включно з
embedded (`audioURI` → `audio/index.m3u8`, `master.go:109`). Розбіжність назви («audio_0» на диску vs
«Original» у БД до першого regen) — та сама косметика, що S3 Minor 4, задокументована.
Альтернатива «regen усіх master у міграції» відкинута: міграція не має доступу до сховища/ffmpeg-шару,
а файл і так валідний.

### D-B5 — ASR до S6 без змін
ASR сьогодні читає `source/<id>/original` (`plan.md` D5 fallback). Backfill-рядок має
`source_leaf='original'`, тож S6 («вхід із default-доріжки») для старих ассетів дасть той самий шлях.
Цей зріз ASR не чіпає.

### D-B6 — Нерозв'язний з БД стан: старий ассет, чия uploaded-доріжка вже стала Ready
Такий ассет у БД не відрізнити від video-only ассета, промоутнутого першою uploaded-доріжкою: маркера
«джерело мало звук» у схемі нема (`00005_media.sql` — лише стани/байти/renditions), а master на диску
вже перегенеровано без `audio/index.m3u8`. Міграція його **не чіпає** (має Ready-доріжку), D-B3 проходить.
Закриває «вгадувати за диском у міграції» (міграція не має доступу до сховища) і «Go-backfill за наявністю
каталогу `audio/`» (deleted-embedded ассет теж зберігає `audio/` — хибнопозитив). Магнітуда: стан
досяжний лише з S5 (upload API — L1b `0b04f5a`, UI — L2b `2292869`) до цієї міграції, лише в test mode
(деплою нема, рішення власника 03.10; pre-launch правило «ріж сміливо» 08.10). Ремонт вручну — повторно
вивантажити оригінальний звук як доріжку. Перевірочний запит для пошуку кандидатів — у report.
Це **звуження охоплення, не ремонт** (critique #2): report не може заявляти «відновлено оригінальний звук
усіх старих ассетів» — лише тих, що не мали Ready-доріжки на момент міграції.

## Тести
- **Integration, populated data (media postgres, goose-провайдер, прецедент 00015):** `DownTo(27)`,
  вставити ассети: (A) ready без доріжок; (A2) ready з лише uploaded `created` (незавершений upload після
  S5); (A3) ready з лише uploaded `failed`; (B) ready evicted без доріжок; (C) ready з embedded від Publish;
  (D) ready video-only з uploaded Ready default; (E) processing; (F) failed; (G) awaiting_audio з uploaded
  processing. `UpTo(28)` → A, A2, A3 отримують embedded default з очікуваними полями й детермінованим id,
  uploaded-рядки A2/A3 без змін; B–G без змін. **Reversibility:** `DownTo(27)` прибирає лише backfill-рядки
  A/A2/A3, embedded C і uploaded D/A2/A3 на місці; повторний `UpTo(28)` відтворює ті самі рядки (той самий
  id). Потім `Up` до голови. (Ідемпотентність самого INSERT — через `NOT EXISTS`: C/D пропущено.)
- **Integration, A2 після міграції:** promote uploaded-доріжки A2 → default лишається embedded, rewrite
  отримує 2 rendition-и (регресія наслідку 1).
- **Integration, конфлікт `und`:** (H) ready без Ready-доріжки + uploaded `queued` з мовою `und` → embedded
  отримує `und-x-original`, default; uploaded без змін. (H2) зайняті і `und`, і `und-x-original` →
  `UpTo(28)` падає з названою помилкою; версія лишилась 27, рядків не додано.
- **Integration, негатив D-B3 (critique #2 F5):** ready-ассет з Ready-доріжкою **без** default (штучно) +
  окремий backfill-кандидат → `UpTo(28)` падає з помилкою інваріанта; версія 27, embedded кандидата не
  додано (вся транзакція відкочена).
- **Integration, серіалізація:** окрема транзакція тримає незакомічений INSERT в `audio_tracks` → `UpTo(28)`
  падає з lock timeout (не проходить поверх); після rollback тієї транзакції `UpTo(28)` успішний.
- Повний ordered `make check` (персистна зміна → повний integration + e2e; §8 «наявний upload→Ready
  лишається зеленим»).

## Ризики
- Міграція на великій таблиці: один INSERT…SELECT + один перевірочний SELECT; індекс
  `audio_tracks_by_asset` покриває `NOT EXISTS`. Фаза не задеплоєна (test mode) — продакшн-обсягу нема.
- Seed-тест `seed.sql` на понижених схемах: seed не вставляє `video_assets` (`sql/seed.sql`) — не зачеплено.
- Спільне дерево з піром: явні шляхи в `git add`.

## Pre-code checkpoint
| задача | READY? | що бракує |
|---|---|---|
| S-backfill | **READY** (plan-critique #3 ACCEPTED, 2026-11-14) | — |

## Журнал
- 2026-11-14: plan-critique #1 (Codex, 4 лінзи в одному контексті) — BLOCKED: #1 «нуль рядків» пропускає
  старий ассет з uploaded-доріжками → умова «нема Ready-доріжки» + D-B6 + конфлікт `und`; #2 конкурентні
  writers → `LOCK TABLE … SHARE ROW EXCLUSIVE` + тест серіалізації; #3 (non-blocking) down→up названо
  reversibility. Raw: `_wip-sbackfill-plan-critique-1.md`.
- 2026-11-14: plan-critique #2 (Codex) — BLOCKED: B1 narrowed (D-B6 — звуження, не ремонт), B2 і N1
  resolved; новий F4 (blocking) recovery `und`-конфлікту для queued/processing → `und-x-original` без
  ручного ремонту; F5 негатив D-B3 → тест; F6 порядок `lock_timeout` перед LOCK. Raw:
  `_wip-sbackfill-plan-critique-2.md`.
- 2026-11-14: plan-critique #3 (Codex) — **ACCEPTED** на plan sha256 `b6b22958…a96c`, snapshot `2292869`;
  F4–F6 resolved, B1 narrowed (D-B6), B2/N1 resolved; F7 (non-blocking) діагностика → D-B3a. Raw:
  `_wip-sbackfill-plan-critique-3.md`. Implementation review і ordered gate — окремі gates.
