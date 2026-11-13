Перевірено лише `git diff ea54313..b4ee658`.

1. **Minor 1 — resolved.** Audio-resume перетворює `invalid_request` на `upload_resume_unverified`: [audio-track-repository.ts:174](/Users/strelia/projects/courses/frontend/src/modules/media/adapters/audio-track-repository.ts:174). Тест із правильним fingerprint і невідповідним `Upload-Length` перевіряє потрібний код: [тест:215](/Users/strelia/projects/courses/frontend/src/modules/media/adapters/audio-track-repository.test.ts:215).

2. **Minor 2 — resolved.** Audio-specific mapping класифікує 409 як `upload_in_progress`: [audio-track-repository.ts:35](/Users/strelia/projects/courses/frontend/src/modules/media/adapters/audio-track-repository.ts:35). Новий тест проходить через HEAD → PATCH 409 і перевіряє результат: [тест:229](/Users/strelia/projects/courses/frontend/src/modules/media/adapters/audio-track-repository.test.ts:229).

**Video tus незмінний за кодом:** [videoTusResource:45](/Users/strelia/projects/courses/frontend/src/modules/media/adapters/tus-client.ts:45) не визначає `statusCode`; нова optional-гілка тому використовує попередній `tusStatusCode`: [виклик:241](/Users/strelia/projects/courses/frontend/src/modules/media/adapters/tus-client.ts:241). Сама функція [tusStatusCode:180](/Users/strelia/projects/courses/frontend/src/modules/media/adapters/tus-client.ts:180), video URL, fingerprint, ліміти та transfer-policy дельтою не змінені.

Нових дефектів чи небезпечної типізації не виявлено. Обидва змінені тести за кодом впали б при поверненні відповідної попередньої поведінки; хибнопозитивності не знайдено.

Тести не запускав, файлів не змінював, мережу не використовував. Виконання — **NOT_TESTED**.

**SHIP — snapshot `b4ee658`.**