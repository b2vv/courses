# A-батч — PR7 (наступний): purchase — СКОУПИНГ (read-only, гейт playback ще йшов)

## Слід
- domain/purchase: grant.ts, purchase.ts.
- adapters/api: course-grant-repository(+test), purchase-repository(+test), price-schema.ts.
- ports: CourseGrantRepository.ts, PurchaseRepository.ts.
- UI: presentation/learn/OrderStatusView(+test), presentation/learn/LockedLessonPanel, presentation/admin/CourseGrantPanel(+test).

## Рішення
- **feature-local, contracts НЕ треба:** flat app/ports споживачі домену — лише самі ports (переїдуть у модуль); жодного flat application. grant живе в purchase.
- **price-schema ЛИШАЄТЬСЯ flat** у adapters/api: `modules/playback/adapters/playback-schema` тягне `../../../adapters/api/price-schema`. Це спільна zod-схема (як transport/generated/api-origin/course-schema). Якщо перенести в purchase — playback-модуль не зможе тягти (module→module deep заборонено). Обидва тягнуть flat. НЕ переносити.
- CourseGrantRepository: enrollment (PR4) лишив його flat САМЕ для purchase — тепер у modules/purchase/ports.
- UI-власність (при виконанні): OrderStatusView→purchase/ui; CourseGrantPanel (admin grant-панель) — радше purchase/ui (керування грантами) або лишити admin; LockedLessonPanel (замкнений урок) — purchase vs learn. Прочитати зв'язність перед рішенням.
- Нагадування: adapter-тести (course-grant-repository.test, purchase-repository.test) під node — конфіг уже ловить modules/*/adapters.

## ВИКОНАТИ ПІСЛЯ: зеленого гейта playback + merge, від свіжого main.
