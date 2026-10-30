# LB09a — orphaned-image compensating delete

**Status:** pre-code → implementing. **Branch:** `lb09a-orphan-delete` off `main` @ `5a671d5`.
**Review:** `review_mode=existing_session`, reviewer `courses-96 [557618]` (mandatory per task).

## Pre-code checkpoint (2026-10-30)
- **Задача / готовність:** AGENDA next move #3 (owner sequence 1→3→2). READY — cite
  `docs/lesson-builder/report.md:698` §Lim3 (LB09a). Один зворотний зріз (git revert, без міграції/даних).
- **Вікно (звірено `internal/image/application/upload.go:43-49`):** `storage.Put` успішно пише байти
  під свіжо-згенерований унікальний `image.ID`, далі `repository.Insert`. Insert падає → байти
  осиротіли (коментар :46 це називає). ID мінтиться `newID()` у цьому ж виклику → на нього НІКОЛИ
  не існувало живого посилання (жодного рядка в БД).
- **Рішення (owner 2026-10-30): компенсаційне видалення, НЕ фоновий scanning sweeper.** При збої
  Insert видалити саме ті байти, що цей виклик щойно записав. Закриває назване вікно (Insert повернув
  помилку) дешево, локально, без ризику (ID унікальний цьому виклику). Фоновий sweeper foreclosed:
  важкий (scan+scheduler+критерій orphan) і вносить ризик видалити живі байти при гонці з in-flight
  upload, якого зараз нема — over-engineering для pre-launch продукту з KB-orphan-ами.
- **Залишок (named, строго менший за сьогоднішній):** краш процесу МІЖ Put і Insert, або сам Delete
  падає (подвійний збій) — там байти все ще лишаються. Це справжня територія фонового sweeper'а;
  винесено як named-limitation у report, не закривається цим зрізом.
- **Форма фіксу:**
  1. `ports.Storage` += `Delete(ctx, id, extension string) error`.
  2. filesystem adapter: `Delete` = guard (validUUID+validExtension, як у Put) → `s.root.Remove(name)`.
  3. `upload.go`: при збої Insert `return domain.Image{}, errors.Join(err, u.storage.Delete(ctx, image.ID, extension))`
     — дзеркалить наявну модульну ідіому компенсації (`storage.go:50`). `errors.Join(err, nil)`
     зберігає error-code (handler мапить Insert-збій через `default:`→503; join не ламає `errors.Is`).
- **Межі (для рев'ю):** sweeper видаляє РІВНО байти цього виклику (`image.ID.ext`), лише коли ЙОГО
  Insert упав. Не scan, не глобальний критерій → 0 ризику зачепити живий образ. Відсутність живого
  посилання доведена тим, що ID свіжий і Insert не створив рядка. Ідемпотентність: один виклик Delete
  на один failed Upload; повторний прогін того ж Upload неможливий (новий ID).
- **Докази (production path):**
  - unit `upload_test.go`: кейс "insert failure" → assert Delete викликано 1× з (want.ID, wantExt);
    успішний/storage-failure кейси → Delete 0×. RED до фіксу (Delete не викликається), GREEN після.
  - unit `upload_test.go`: кейс "insert and delete failure" → `errors.Is(err, insertErr)` **і**
    `errors.Is(err, deleteErr)` обидва true (join surфейсить обидва, не маскує).
  - filesystem `storage_test.go`: `TestDelete` видаляє наявний файл (директорія порожня по тому);
    `TestDeleteInvalidKey` — guard відхиляє `errInvalidImageKey`, файл не чіпається.
- **Impact / гейти:** backend Go (image adapter+app+тести). Чіпає **persisted on-disk state** → гейт
  **повним ordered `make check`** (не таргетним): format-check·lint·typecheck·test·**integration
  (image filesystem, авторитет)**·build·**e2e** — не таргетний, бо sweep зачіпає диск. e2e: upload
  happy-path незмінний (Delete лише на Insert-failure path).
- **Review:** writer Courses 1 (ця сесія), reviewer `courses-96 [557618]` read-only через git show /
  gh pr diff. Обов'язкове незалежне рев'ю перед merge.

## Журнал
- 2026-10-30: pre-code checkpoint; механізм вікна звірено в коді; owner обрав компенсаційне видалення.
- 2026-10-30: реалізовано. `ports.Storage`+=`Delete`; filesystem adapter `Delete` (guard+Remove, ctx-cancel не чіпає файл); `upload.go` join-компенсація при збої Insert (+errors import). Тести: upload_test 2 нові кейси (insert-failure→Delete 1× з (id,ext); insert+delete-failure→join surфейсить обидва), storage_test 4 нові (Delete happy, missing-file errors, invalid-key не чіпає сусідній файл, ctx-cancel). gofmt/vet clean, image unit зелені.
- 2026-10-30: **RED→GREEN непорожність доведено** — stash upload.go до main: `insert_failure_deletes_bytes` (Delete calls=0 want 1) + `insert_and_delete_failure_surface_both` (join не wrap-ить delete err) падають; з фіксом зелено.
- 2026-10-30: запущено повний ordered `make check` (gate-lb09a-1.log, GATE_EXIT) — не таргетний, бо чіпає persisted on-disk стан. Чекаю.
