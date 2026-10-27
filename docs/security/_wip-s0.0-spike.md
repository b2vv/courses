# S0.0 expressiveness spike — WIP (live notes)

ДАЛІ: згенерувати `courseDTO` через openapi3gen, звірити структуру з hand-spec `Course`,
переписати override-поверхню (enum/format/validation/additionalProperties/Rating-divergence).

Branch: s0.0-spike (TBD) · base main @ 8021e83 · owner go 2026-10-27.

## ФАКТ (grounding, measured on main @ 8021e83)
- `courseDTO` (course/adapters/http/handler.go:28) — 13 json-полів, **жодного `,omitempty`**;
  `PublishedAt *time.Time`, `Price *priceDTO`, `Rating *ratingDTO` — вказівники.
- openapi3gen v0.149.0: вказівник → `Nullable=true` (не root); `time.Time`→format date-time;
  `int64`→format int64. **НЕ заповнює `Required`взагалі**; НЕ дає enum/format:uuid/validation/
  additionalProperties. `SchemaCustomizer(name,t,tag,schema)` — per-field хук.
- DTO **неекспортовані** (`courseDTO`) → центральний реєстр §4-A з ІНШОГО пакета їх не відрефлексить.
- hand-spec `Course.rating` — **inline** object з minimum:1/maximum:5; компонент `Rating` — БЕЗ min/max.
  Той самий Go-тип `*ratingDTO` → дві різні схеми за контекстом.
- `Price` → через `NullablePrice` (allOf[Price]+nullable) wrapper-компонент.

## РІШЕННЯ
- S0.0 доведено 4/4 зелені (`go test -tags s0spike`). Fidelity досяжна; override-поверхня виміряна.
- F1: Block override-registry НЕ потрібен для published-схеми (вона пермісивна); потрібен лише
  string-enum для BlockType. marshaler-assertion лишається як guard, не механізм коректності.
- F2: DTO неекспортовані → генератор per-module (кожен адаптер-пакет віддає route→type), не центральний.
- B1 зелено: 21 код із discussion/domain/errors.go через go/parser; call-site дає 0.
- N3 зелено: 1 raw w.Write (callback) → declared exclusion.
- Наступне: рев'ю Courses 2 (existing_session) §14, тоді S0.1.
- ВЛАСНИК 27.10: повний генератор (мета — не правити сваггер руками); reuse-path мету не вирішує.
- COURSES 2 ВЕРДИКТ: S0.0 PASS, 0 blocking (§16). F1-blocking знято push-back'ом із цитатами
  (request-схема документа в спеку ВЖЕ пермісивна; строгість лише в Go-валідаторі). Прийнято
  16.1 (error-code resolve: 4 форми виклику + const-decl), 16.2 (request single-layer — Layer B
  лише responses), 16.4 (request-DTO у fidelity), 16.5 F2 Option 1 (per-module), 16.3 (Block→string-enum).
- S0.1 entry: 16.1 parser + 16.2 request-layer рішення (ValidateRequest vs named limit) + 16.4 +
  16.5 per-module + per-module coverage-gate. Чекає owner go на pacing S0.1.

## Журнал
- grounding course DTOs/Block/discussion errors/orders callback + openapi3gen API (main @ 8021e83)
- spike A (structure+census) + B (B1 errcode parse / N3 non-json) → 4/4 PASS
- §14 виписано в план-док; F1/F2 як plan edits на рев'ю
