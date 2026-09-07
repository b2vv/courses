
## Лінза Edge cases — r4, запуск
ДАЛІ: читаю docs/playback/plan.md (ред. 4), потім spec.md §2.5/§5, spike-edge.md/.conf, _wip-phase6.md; далі історія критик r1–r3, щоб не повторювати закрите.

ФАКТ: план ред.4 прочитано цілком (522 рядки). Ключове для лінзи:
- D-2 §3: `proxy_cache_key "$pb|$vid"`, `proxy_cache_valid 200 204 30s`, гард на форму id (32 hex).
- D-3: 204 вимагає (кука дійсна) + (сесія в Redis, userId==актор, videoId збіг, owner==pb) + видимість курсу.
- D-4: play завжди забирає слот; heartbeat не-власника -> 409; logout DEL owner+session.
- D-5: clientSeq засівається з сервера; межа 2^31; обрізання позиції — на читанні.
- D-8: heartbeat тіло {playbackId, positionSeconds, clientSeq, reason} — lessonId у тілі НЕМАЄ.
- §8: R11/R12/R13 кажуть «недоступний Go/Redis/PG -> 403», а D-3/D-2/B36 кажуть 503. Розбіжність усередині ред.4.
- §4 і §9 повторюють «403 (A10)» і «playbackId сам по собі доступу не дає — auth_request однаково вимагає куку».
ДАЛІ: spec.md §2.5/§5 + spike-edge.md (що саме заміряно), тоді точкові перевірки коду (deleteMissingAuthoringLessons, logout, StrictMode).
