# Foundation delivery evidence

Дата: 2026-09-03. Scope: Phase 1 foundation + Phase 2 public Course vertical slice.

## 1. Що зроблено

Результат відповідає задачам F01–F06 і C01–C05 у [плані](plan.md): modular-monolith Go runtime, PostgreSQL/Redis, окремий worker executable, закритий media data plane, Start/Rsbuild SSR, React Aria kit, OpenAPI→Orval, published Course domain/use cases/repository/API і публічні SSR сторінки.

Дельти до плану: за наданим mockup застосовано візуальну мову CourseForge без вигаданих цін/рейтингів; за уточненням користувача додано request-scoped i18next `uk`/`en`; після review вимкнено raw Nginx error sink, а safe structured access log збережено. Commits: `5608a03`, `31e8c39`, `4d73072`.

## 2. Протестовані сценарії

| ID | Фактичний результат | Вердикт |
|---|---|---|
| H01 | `make up`: 6 runtime services healthy, migrate exited 0 | збіглося |
| H02 | no-JS `/courses` містить seeded Course у початковому HTML | збіглося |
| H03 | перехід із каталогу показує title/description Course | збіглося |
| H04 | q/sort/limit/offset зберігаються в URL; HTTP order стабільний | збіглося |
| H05 | keyboard flow, skip link і axe WCAG A/AA — без порушень | збіглося |
| H06 | API/worker health/readiness/status повертають документовані дані | збіглося |
| H07 | hydration без console errors і повторного browser fetch | збіглося |
| B01 | empty page має EmptyState; 0 не маскує dependency error | збіглося |
| B02 | repository/HTTP перевірено для limit 1, 50 і 51 records | збіглося |
| B03 | offset 10000/q 100 code points прийнято; NUL відхилено | збіглося |
| B04 | beyond-end зберігає total і пропонує first page | збіглося |
| B05 | title/published ties мають стабільний id tie-breaker | збіглося |
| B06 | draft/archived відсутні в list | збіглося |
| B07 | draft/archived/unknown detail дають однаковий 404 flow | збіглося |
| B08 | mobile dark + 200% text zoom не мають horizontal overflow | збіглося |
| B09 | два seed: `INSERT 0 5`, потім `INSERT 0 0` | збіглося |
| E01 | malformed/duplicate/bounded query і slug дають typed 400 | збіглося |
| E02 | SSR unknown course повертає status 404 + recoverable page | збіглося |
| E03 | зупинений PostgreSQL дає safe 503, після start API відновлюється | збіглося |
| E04 | зупинений API дає ErrorState, а не порожній каталог | збіглося |
| E05 | frontend deadline і backend DB timeout завершують read | збіглося |
| E06 | caller/router abort доходить до fetch/pgx; pool відновлюється | збіглося |
| E07 | `/media/*` через Nginx повертає 403 | збіглося |
| E08 | `/admin/*` і `/learn/*` fail closed без приватних даних | збіглося |
| E09 | burst через Nginx містить 429 | збіглося |
| E10 | restart зберігає три published demo courses | збіглося |

Додаткові gates: i18next перемикає `uk/en`, `<html lang>` і SSR cookie; failure→retry зберігає весь URL; raw Nginx logs не містять query/Referer marker. Неперевірених acceptance-рядків: **0**.

Фінальний порядок: `make format-check` PASS; `make lint` PASS; `make typecheck` PASS; `make test` PASS (53 Rstest + Go race); `make integration` PASS (PostgreSQL 17 Testcontainers); `make check-generated` PASS; `make build` PASS (client + SSR + API/worker/migrate images); `make e2e` PASS (14 Playwright); smoke PASS.

## 3. Неочевидне

- npm 11.6 створив неповний platform lock; npm 11.19 package-lock-only відновив optional dependencies, після чого Linux `npm ci` пройшов ([package-lock.json](../../frontend/package-lock.json)).
- TanStack default search parser перетворював literal `q=2026/true` і exponent numbers; raw URL parser тепер зберігає strings/repeats, а validator їх перевіряє ([url-search.ts](../../frontend/src/app/router/url-search.ts)).
- `ensureQueryData` у loader не підключав lifetime navigation; QueryObserver consumer тепер скасовує останній obsolete read і не зберігає старий signal у cache ([catalog-queries.ts](../../frontend/src/app/composition/catalog-queries.ts)).
- Nginx raw error log записував query/Referer; safe access format не впливав на нього. HTTP error sink вимкнено, verified marker count після fix = 0 ([nginx.conf](../../infrastructure/nginx/nginx.conf)).

## 4. Скріни

- [Desktop catalog](screenshots/catalog-desktop.png)
- [Mobile dark catalog](screenshots/catalog-mobile-dark.png)
- [Course detail](screenshots/course-desktop.png)
- [Dependency error](screenshots/catalog-error.png)

## 5. Проходка по user flow

Playwright відкрив `/courses` без JavaScript, прочитав PostgreSQL course у initial HTML, перевірив відсутність draft/archived і відкрив detail. У JS-сесії змінив search/sort, відновив beyond-end сторінку, перемкнув EN та перезавантажив SSR з тією самою locale. Потім зупинив API й PostgreSQL, перевірив доступні error/503, підняв залежності, натиснув retry зі збереженим URL і підтвердив дані після повного restart.

## 6. Ризики й обмеження

Identity, admin CRUD, upload/transcoding, playback authorization/player, transcript і orders залишаються Phase 3–8. Worker запускає Asynq runtime, але чесно повідомляє 0 handlers; `/media` заборонено. Sentry, frontend OTel exporter і advanced UI/player/editor dependencies додаються лише з реальним caller у наступних phases. Авторський Course content у demo seed український і не перекладається i18next.

Це локальний foundation: TLS, production secrets, backups/restore drills, S3/MinIO adapter, quotas та deployment topology ще не реалізовано. HLS не є DRM і protection не заявляється.

## Post-deploy

Не застосовно: public deployment не виконувався.
