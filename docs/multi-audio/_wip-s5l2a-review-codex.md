1. **Minor — CONFIRMED** — `frontend/src/modules/media/adapters/tus-client.ts:282–288`. Audio-resume з невідповідним `Upload-Length` повертає `invalid_request`, хоча D-L2-2 вимагає `upload_resume_unverified`. Тест `audio-track-repository.test.ts:215–226` закріплює неправильний код. Завантаження залишається безпечним: PATCH не виконується. **Виправлення:** повертати потрібний код для audio-resume й оновити тест, зберігши попередню поведінку відео.

2. **Minor — CONFIRMED** — `frontend/src/modules/media/adapters/audio-track-repository.ts:91`, `tus-client.ts:177–189`. Audio успадковує перетворення **409 `upload_offset_conflict` → `invalid_response`**. Backend повертає цей коректний бізнес-конфлікт (`backend/internal/media/adapters/tus/server.go:112`), зокрема коли дві вкладки продовжують один upload з однакового offset. Споживач втрачає причину відмови. Це успадковане обмеження, **не регресія відео**. **Виправлення:** додати audio-specific класифікацію конфлікту та тест відповіді 409, залишивши video mapping незмінним.

**Major/Blocker не виявлено.** Перевірено:

- Video URL, fingerprint-ключ, ліміти, chunk, error mapping та callers у repository/application зберігають попередню поведінку.
- Audio JSON paths, статуси CRUD, error/status-пари, ticket, порожні 204 та language/label rules відповідають backend.
- Resume розділяє audio/video fingerprints, перевіряє файл і HEAD; cancel працює без ticket.
- Terminal signature охоплює перший `ready`/`failed`; тотожні відповіді не інвалідують повторно. Колізій ключів і refetch-циклу не знайдено.
- Learner SQL правильно мапить `awaiting_audio → processing`; integration-тест перевіряє реальний SQL-шлях і виявив би повернення старої поведінки. Іншого неврахованого wire-споживача не знайдено.
- Порушень layering, нового `fetch` у UI/routes, `any` чи небезпечних assertions не знайдено.

Тести не запускав, файлів не змінював, мережу не використовував. Оцінка тестів — за кодом; виконання **NOT_TESTED**.

**SHIP**