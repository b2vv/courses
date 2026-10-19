# editor — layout-diff proto ↔ prod (design-pass крок 1)

**Дата:** 2026-10-15. Viewport 1440×900. proto `:8771/#lesson` ↔ prod `:8080/admin/courses/.../lessons/...`.
Палітру НЕ рахуємо (ок). Нижче — лише layout/структура/наявність елементів. (EN у prod — бо активний
EN-тогл; мовна різниця не рахується.)

| # | Елемент | proto (ціль) | prod (поточний) | gap |
|---|---|---|---|---|
| E1 | **PageHeader eyebrow** | «LESSON DOCUMENT · MODULE 01 / 02» (caps, letter-spaced) над title | відсутній | **є** |
| E2 | **PageHeader опис** | «Створіть урок, який хочеться читати.» під title | відсутній | **є** |
| E3 | **Primary-кнопка** | «Опублікувати ›» з chevron | «Publish» без chevron | дрібний |
| E4 | **Segmented [Редактор\|Перегляд]** | є, з іконками, під хедером | **відсутній повністю** | **велике** |
| E5 | **Save-status** | «• Збережено в цій сесії · 3» (dot + лічильник секцій) справа від segmented | відсутній | **велике** |
| E6 | **Title-block** | eyebrow «НАЗВА» + великий заголовок уроку + caption курс/модуль | **відсутній** (prod одразу секції) | **велике** |
| E7 | **Section card: type-tile + caption + index** | icon-tile зліва + «ВІДЕО»/«ТЕКСТ» caption + **padded «01»** | icon-tile + caption + «1» (без паддінга) | середній |
| E8 | **Drag-handle** | тьмяний drag-хендл на дальньому лівому краю картки (future-reorder affordance) | відсутній | середній |
| E9 | **Delete-control** | тихий trash-icon | **червоний ✕** (гучний) | середній |
| E10 | **Toolbar тексту** | **icon-toolbar** (B/I/H2/H3/list/ordered/link іконки) + caption «Текст» справа | **текстові кнопки** (Bold/Italic/Heading/Subheading/Bulleted list/Numbered list/Link) | **велике** (конфлікт з PR #33) |
| E11 | **Link UI** | **ModalBox** (діалог) | **inline-поле** «Link address»+Apply під тулбаром | **велике** (конфлікт з PR #33) |
| E12 | **Empty-state (video)** | центрований icon + caption + title + опис + «Налаштування відео ›» | простіший/площинний | середній |
| E13 | **Empty-state (image)** | багатший центрований стан | dropzone-рядок + «Upload image» + «Image description» | середній |
| E14 | **Add-section палітра** | Video/Text/Image/Row/Task tiles icon+title+caption | **майже збігається** | мінімальний |
| E15 | **Sidebar chrome** | brand icon-tile «coursespace», «WORKSPACE · ADMIN» pill, reference-блок, user-footer з аватаром | текстовий «CourseForge/Admin», групи PEOPLE/MODERATION, «To client» | **shell-рівень** (спільний для всіх 8 views, не лише editor) |
| E16 | **Top sub-bar** | «Референсний прототип…» + «Специфікація дизайну ›» | відсутній | shell-рівень (демо-специфіка proto — ймовірно НЕ переносити дослівно) |

## Класифікація для editor-пілоту

- **Велике (структура, обов'язково):** E4 segmented, E5 save-status, E6 title-block, E10 icon-toolbar,
  E11 link-ModalBox.
- **Середнє (деталь картки/станів):** E7 padded index, E8 drag-handle, E9 trash-icon, E12/E13 empty-states.
- **Дрібне:** E3 chevron, E14 (вже близько).
- **Shell-рівень (E15):** спільний admin-shell — виносимо в окремий зріз «admin shell fidelity» перед/разом
  із views, бо стосується всіх 8. E16 — демо-специфіка proto, переносити лише доречне.

## Конфлікт із PR #33 (editor-polish)

E10 (text→icon toolbar) та E11 (inline→ModalBox link) прямо переробляють UI, доданий у PR #33.
Поведінку PR #33 (per-image upload, no-text-loss, focus, invalid-URL-keeps-link, announce) **зберегти**;
переносимо її на нову розмітку. Поведінкові тести лишаються регресом.

## Примітка про стан prod-скріна

Prod editor на момент зйомки мав залишковий брудний стан минулої сесії (відкрита text-секція з
«notaurl» link-error). На structural chrome (E1–E6, E15) це не впливає — вони відсутні незалежно від стану.
