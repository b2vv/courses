# A-батч — PR9 (наступний): catalog — СКОУПИНГ (read-only)

## Слід (публічна вітрина)
- Немає domain/catalog — тягне `contracts/course`.
- application/catalog/catalog.ts(+test).
- ports/CourseRepository.ts (course PR лишив flat «у catalog»).
- adapters/api/course-repository.ts(+test).
- presentation/storefront: CatalogView(+test), CourseView, HomeView, RatingSummary, ValueStrip, CatalogStates, ratings.test.

## Рішення
- feature-local, нові contracts НЕ треба; модуль тягне `contracts/course` (kernel) — port/application/adapter імпортують звідти (module→contracts дозволено check-ом).
- **Лишити flat (→admin-модуль):** admin-course-repository(+test), AdminCourseRepository.
- UI storefront → catalog/ui. CourseView має крос-модульні barrel-імпорти (ReportDialog=enrollment, можливо ReviewBlock=learn, RatingSummary) — presentation→barrel ок, поправити глибини.
- Перевірити: ratings розщеплено (RatingSummary=storefront, ReviewBlock=learn); ReviewBlock лишити в learn? ratings поки в межах course/catalog (немає окремого ratings-домену). Прочитати зв'язність CourseView/CatalogView при виконанні.
- adapter-тести (course-repository.test) під node — конфіг ловить.

## ВИКОНАТИ ПІСЛЯ: зеленого гейта transcript + merge, від свіжого main.
