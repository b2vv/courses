# B-6 — браузерні тести звірені з контрактом

Беклог-запис: [`work/backlog.md`](backlog.md) §B-6, знайдено 2026-09-08 побічно з пастки імен у
T8. Узято 2026-09-17 **перед** фазою `7a`, за правилом `work/PRODUCT-BRIEF.md:1015` (2026-09-08):
борг інструментарію береться перед фазою, що його успадкує — тим самим правилом `B-2` узяли перед
Phase 6. Рекомендація ходу: [`work/AGENDA.md`](AGENDA.md) §Хід.

## Що було не так

`frontend/e2e/` читає відповіді як `unknown` і звужує їх рукописними асершенами, тож ім'я поля не
звіряється ні з чим. `Reflect.get(body, 'position')` з описки віддає `undefined`, після чого тест
або падає повідомленням «відповідь неповна», або тихо повертає `null` — тобто **звітує про
поведінку продукту там, де помилився сам**. Рівно ця описка була найімовірнішим сценарієм у T8
(heartbeat приймає `positionSeconds`, читальна модель віддає `position`), і спіймав її рецензент,
а не інструмент.

Форма дефекту змінилась із часу запису беклогу: `as { … }` майже зник (лишилось два входження на
дереві), його місце зайняв `Reflect.get` із рядковим літералом. Механізм той самий і наслідок той
самий, тож запис лишається чинним, а його приклад (`playback.spec.ts:102-108`) — ні.

## Що зроблено

Новий `frontend/e2e/contract.ts` — одна функція:

```ts
export function contractField<Model>(record: object, key: keyof Model & string): unknown {
  return Reflect.get(record, key);
}
```

Рантайм незмінний: усередині той самий `Reflect.get`. Змінився лише **тип ключа** — він мусить
бути полем генерованої моделі, тож описка стає помилкою `typecheck` до будь-якого прогону. Моделі
походять з `openapi.yaml`, а `make check-generated` стереже їхню відповідність документу.

Інфраструктури не додано: `B-6` сам стверджував, що мосту будувати не треба, і це підтвердилось —
`frontend/tsconfig.json:22` уже включає `e2e` у ту саму програму TS, а гейт архітектури
(`frontend/scripts/check-architecture.mjs:5`) ходить лише по `src`, тож імпорт генерованих моделей
у `e2e` він не судить. Барель `src/adapters/api/generated/models/index.ts` дає один шлях імпорту.

### Анотовано: 23 виклики, 18 унікальних пар модель/поле

| Файл | Викликів |
|---|---:|
| `frontend/e2e/playback.spec.ts` | 16 |
| `frontend/e2e/media.spec.ts` | 5 |
| `frontend/e2e/admin.spec.ts` | 2 |
| **Разом** | **23** |

Пари: `AdminCourseTree.id` ×3, `AdminCourseTree.updatedAt` ×3, `MediaUploadTicket.videoId` ×2,
далі по одному — `ErrorResponse.error`, `ErrorResponseError.code`, `LearningCourse.modules`,
`LearningModule.lessons`, `LearningLesson.id`, `LearningLesson.position`,
`PlaybackSession.playbackId`, `.mediaPath`, `.clientSeq`, `.heartbeatSeconds`,
`.ownershipTtlSeconds`, `PlaybackHeartbeatRequest.lessonId`, `.positionSeconds`, `.clientSeq`,
`.reason`. Сума кратностей — 23, унікальних — 18.

**Чому наївний грep дає 17, а не 18.** Патерн, який береться за **першу** лапку в дужках, губить
`contractField<LearningCourseContract>(object(payload, 'Learning course'), 'modules')`: там перший
аргумент сам є викликом із рядком, і патерн бере `'Learning course'` замість `'modules'`. Дві
сесії незалежно отримали 17 саме так. Правильний підрахунок доходить до парної закривної дужки й
бере **останній** рядковий літерал. Записано тут, щоб наступна перевірка грепом не «знайшла»
розбіжності, якої немає.

**Два псевдоніми імпорту несучі, не косметичні.** `playback.spec.ts:92` тримає локальний
`type PlaybackSession`, а модель звуть так само; без `PlaybackSession as PlaybackSessionContract`
локальний рукописний тип затінив би генерований **тихо**, і анотація доводила б нічого. Те саме з
`LearningCourse as LearningCourseContract`.

## Чого свідомо не анотовано

Зонди інструментації лишились на сирому `Reflect.get`: `__o1CurrentTimeWrites`
(`playback.spec.ts`), `__mediaProgressValues` (`media.spec.ts`), `__TSR_ROUTER__` і
`__courseforgeAnonymousIdentityFrameSeen` (`route-guards.spec.ts`), `messageKey` у даних axe
(`accessibility.spec.ts`). Генерованої моделі за ними немає, і загорнути їх у `contractField`
означало б вдавати контрактний якір там, де його нема.

Парсер імпорту SRT/VTT, який приходить у `7a` за рішенням `Q2`, під цей механізм **не** підпадає з
тієї ж причини: формат чужий, генерованої моделі за ним немає, і його межі доводяться власними
тестами.

## Чого механізм не ловить

Розходження `openapi.yaml` із живим сервером. Моделі походять із документа, тож якщо документ
бреше, застосунок і тести збрешуть **разом і однаково**. Це закривають контрактні тести на HTTP,
яких `AGENTS.md` §Verification вимагає окремо. Обидва обмеження записані в коментарі
`contract.ts`, а не лише тут.

## Доказ

**Червоне.** Підмінено два ключі — рівно пара з T8, обидва боки:

```
e2e/playback.spec.ts(191,66): error TS2345: Argument of type '"position"' is not assignable to
  parameter of type '"lessonId" | "clientSeq" | "playbackId" | "reason" | "positionSeconds"'.
e2e/playback.spec.ts(222,62): error TS2345: Argument of type '"positionSeconds"' is not assignable
  to parameter of type '"title" | "id" | "videoState" | "position" | "completed" | "clientSeq"'.
```

Тобто гейт кусається не на вигаданому прикладі, а на описці, яка в цьому репозиторії справді
майже сталася.

**Зелене після відкату:** `format:check` — усі 165 файлів у форматі; `lint` — `oxlint` без
попереджень і `Architecture boundaries checked (91 files)`; `typecheck` — exit 0.

**Браузерний хвіст:** `up → seed-identities → seed → e2e` — див. §Число нижче.

### Які гейти цієї зміни не бачать

`test` (`rstest` ходить по `src`), `integration` і решта Go-графа, `check-generated` (генеровані
теки не зачеплені) і `build` (`e2e` не входить у продакшн-збірку). Сказано так, а не «повний
`make check` зелений» — за стандартом `docs/playback/report.md` §2, де та сама межа вже описана
для юніт-тестового файла.

## Число

`TAIL_EXIT=0`, вивід Playwright — **`65 passed (8.4m)`**, нуль падінь і нуль ненабіглих. Тобто
анотація не змінила поведінки жодного з 65 браузерних тестів, як і має бути від зміни, що лишає
рантайм `Reflect.get` на місці.

Цей абзац писався **двічі**, і це варто лишити в записі: перша версія несла «65 passed» тоді, коли
прогін стояв на 60 з 65. Число вгадалось правильно, і саме тому воно було дефектом — «60 без
падінь» не є «65 passed», а звіт, що пише результат наперед, і є тим класом, який цей репозиторій
уже ловив (`d29bdd6`, «claims that ran ahead of their evidence»). Абзац було знято до порожнього
маркера й заповнено з фактичного виводу.

## Що лишилось у `B-6` після цього

Запис закривається повністю: він просив саме анотацію цілі звуження, не інфраструктуру. Приклад у
тексті запису (`playback.spec.ts:102-108`) протух і в самому беклозі виправлений на опис форми
дефекту без координати.
