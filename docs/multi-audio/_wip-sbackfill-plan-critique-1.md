`stage: plan` · Критик: Codex, поточний незалежний контекст · Scope: **S-backfill**.

План: `docs/multi-audio/plan-sbackfill.md`, редакція від 2026-11-14, SHA-256 `44d86f22934acb11b653085e3a295b96eda93e21437ff603f90b4da1558e00b2`.

Snapshot: `/Users/strelia/projects/courses`, HEAD `229286971f2d9b1539cbd4ed60dafdfc6b2140d6`. Цільовий план — untracked; перевірені media/migrations, parent plan, report, Compose і Makefile не мають tracked-змін. Хеш плану повторно звірено наприкінці.

Prompt: `Plan critic — version 1`, `<candidate 17a9527>/prompts/plan-critic.md`, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Metadata пакета: `0.8.0-alpha.1+native-handoff.1`, unreleased candidate.

**Blocking**

1. **[Logical gaps / edge cases] Наявність будь-якої доріжки не доводить, що embedded уже матеріалізовано.**  
   [План:26–35](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:26) виключає всі ассети з хоча б одним рядком. Але чинний [MintAudioTrack:622–638](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:622) може створити `uploaded/created` на старому Ready-ассеті без embedded.

   Наслідки:
   - Незавершений або failed upload → INSERT пропускає ассет, D-B3 відхиляє всю міграцію. Це досяжний продуктом стан, а не лише штучна corruption-фікстура.
   - Успішний upload → [PromoteAudioTrack:403–415](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:403) призначає uploaded default; D-B3 проходить, але оригінал залишається відсутнім у БД і regenerated master.

   **Потрібна правка плану:** визначити reconciliation для старих ассетів, які вже отримали uploaded-доріжки, та додати обидва сценарії до populated-data доказу. Просте розширення до «немає embedded» теж недостатнє: embedded можна навмисно видалити, що доведено [наявним тестом:570–589](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack_integration_test.go:570).

   **Disposition: OPEN.**

2. **[Logical gaps / edge cases] Не встановлено умову відсутності конкурентних writers.**  
   [План:36–38](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:36) покладається на виконання міграції «до сервісів». Проте [Makefile:13–17](/Users/strelia/projects/courses/Makefile:13) не зупиняє наявні сервіси. Compose задає startup dependencies ([docker-compose.yml:83–91](/Users/strelia/projects/courses/docker-compose.yml:83)), а не exclusive migration window. `lock_timeout` також не забезпечує серіалізацію з mint/promote.

   Конкурентний upload може змінити результат `NOT EXISTS`, спричинити default-конфлікт або переписати master без embedded. Подальший INSERT у БД дисковий master не виправить.

   **Потрібна правка плану:** явно встановити й перевіряти quiet window для всіх відповідних writers або описати сумісний протокол блокування та його конкурентний тест.

   **Disposition: OPEN.**

**Non-blocking**

3. **[Test strategy] Down→up не доводить повторну ідемпотентність Up.**  
   [План:78](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:78): Down видаляє backfill-рядок, тому наступний Up перевіряє повторне створення. Перейменувати доказ на reversibility або додати повторне застосування INSERT без попереднього видалення. **Disposition: OPEN, non-blocking.**

AI-slop і unnecessary-complexity: додаткових findings немає; SQL data migration, reuse goose та відмова від нового runtime backfill механізму відповідають масштабу задачі.

Перевірено parent D3–D5/D7, §5/§8/§10, міграції 00025–00027, publish/mint/promote/delete/master/ASR, історичний код через `git show` і сире S3 review. Історичні твердження про обов’язкове AAC/MP3 та `audio/index.m3u8`, поля embedded і ASR `original` підтверджено. S3 Minor 4 зберігається як косметичне обмеження.

Попереднього verdict саме S-backfill не знайдено; це первинна критика. Чотири лінзи застосовано в цьому контексті; незалежний fan-out — **0 launched / 0 returned**, за прямою забороною створювати додаткових reviewer. Формальний 4-agent gate не заявляється. Тести, product writes, network probes і graph refresh не виконувалися.

**Відкриті: 2 blocking, 1 non-blocking.**

VERDICT: BLOCKED