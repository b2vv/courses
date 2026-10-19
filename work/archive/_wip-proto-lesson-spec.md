# Proto lesson editor spec — extraction (read-only)

ДАЛІ: знайти в прото lesson-view (#lesson / Lesson* компонент) у src.jsx; скласти карту класів; потім витягти CSS по кожному класу зі styles.css.

## ФАКТ: lesson JSX знайдено
- Компонент: `case'lesson'` у `pageContent()`, src.jsx рядок 30. Корінь `<section className="lesson-editor">`.
- RichText toolbar: src.jsx рядок 19. palette: рядок 24. sectionBody: рядок 25. Empty/PageHeader/ModalBox/Field/IconButton/Pill: рядки 7-16.
ДАЛІ: витягти CSS по класах зі styles.css.

## ТОКЕНИ (proto :root, dark-only — мапити на наші)
- space: --space-1 .25rem · --space-2 .5rem · --space-3 .75rem · --space-4 1rem · --space-5 1.25rem · --space-6 1.5rem · --space-8 2rem
- --control-min-size 2.75rem (=44px) · --card-padding = --space-4 · --space-icon 1.25rem · --space-prose-measure 65ch · --space-prose-min 12rem · --space-callout-rule .25rem · --space-focus-ring 3px · --space-focus-offset 3px
- radius: --radius-card .75rem · --radius-control .625rem · --radius-panel .75rem · --radius-pill 999rem
- color: background #0c0d12 · surface #111317 · elevated #17181d · text #e6e8ec · muted #8b95a3 · faint #828b99 · border #656e7b · hairline #20232b · primary #5273f2 · primary-hover/info #6986ff · on-primary #fff · focus #7b93ff · success #4ed6a8 · warning #fcb646 · danger #e5695f
- fonts: --font-body 'Inter Variable' · --font-display 'Space Grotesk Variable' · --font-utility = MONOSPACE (ui-monospace…) — eyebrows/captions/section-type/table-th усі моноширинні
- --btn-primary-bg: linear-gradient(180deg,#6986ff,#4c6bf0) · --card-shadow=--shadow-card
- h1: clamp(1.75rem,1.3rem+1.2vw,2.25rem)/750, letter-spacing -.035em. h2 1.0625rem/700. h3 1rem/700.

## 1. PageHeader (src.jsx:16,30)
DOM: `<header class="admin-page-header"><div><span class="admin-eyebrow">LESSON DOCUMENT · MODULE 01 / 02</span><h1>Lesson editor</h1><p>Create a lesson worth reading.</p></div><div class="admin-page-actions">{actions}</div></header>`
actions у lesson = Save draft (`ui-button ui-button--secondary`) + Publish (`ui-button ui-button--primary` з `<Icon name="arrow"/>` = chevron-right path `m9 5 7 7-7 7`, іконка ПІСЛЯ тексту).
CSS:
- `.admin-page-header{display:flex;justify-content:space-between;align-items:center;gap:var(--space-5);flex-wrap:wrap;padding-block:var(--space-2) var(--space-4)}`
- `.admin-page-header p{font-size:.8125rem;color:var(--color-muted);max-inline-size:var(--space-prose-measure)}`
- `.admin-page-actions{display:flex;align-items:center;gap:var(--space-3);flex-wrap:wrap}`
- `.admin-eyebrow,.lesson-editor-section-type{font:.6875rem/1.5 var(--font-utility);font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:var(--color-faint)}` (eyebrow моноширинна, caps, трекінг .08em)
- `.ui-button{display:inline-flex;align-items:center;justify-content:center;gap:var(--space-2);min-inline-size:2.75rem;min-block-size:2.75rem;padding:var(--space-2) var(--space-4);font-size:.8125rem;font-weight:650;border:1px solid var(--color-background);border-radius:var(--radius-control);background:none;color:var(--color-text);transition:background 140ms,border-color 140ms,box-shadow 140ms}`
- `.ui-button--secondary{background:var(--color-elevated);border-color:var(--color-border)}` hover border-color:text
- `.ui-button--primary{font-size:1.1875rem;font-weight:750;line-height:1.3;color:var(--color-on-primary);background:var(--btn-primary-bg);border-color:var(--color-primary-hover);box-shadow:var(--btn-primary-highlight),var(--btn-primary-glow)}`
  ⚠️ УВАГА (_ORIGIN.md Рішення 1 власника 2026-10-14): у ПРОД primary НЕ роздувати до 1.1875rem/750 — тримати нашу шкалу .8125rem/650 і поглибити ЗАЛИВКУ до color-mix(... 86%, #000)=#4763d0 (5.30:1). Великий текст — лише в еталоні.

## 2. Segmented [Editor|Preview] (src.jsx:30)
Розміщення: усередині `<div class="lesson-editor-meta">` (перший child; save-status — другий), блок ПІД PageHeader.
DOM: `<div class="ui-segmented">` → 2× `<Button class="ui-button ui-segment" aria-pressed={...}><Icon name="text|eye"/>{Editor|Preview}</Button>`. Editor icon=`text` (M4 5h16M12 5v15M8 20h8), Preview icon=`eye`. active = aria-pressed=true (НЕ клас).
CSS:
- `.ui-segmented{display:flex;align-items:center;gap:var(--space-1);padding:var(--space-1);background:var(--color-surface);border:1px solid var(--color-hairline);border-radius:var(--radius-control);flex-wrap:wrap;max-inline-size:100%}`
- `.ui-segment{color:var(--color-muted)}` (решта з .ui-button)
- `.ui-segment[aria-pressed=true]{background:var(--color-elevated);color:var(--color-text);border-color:var(--color-border);box-shadow:var(--card-shadow)}`

## 3. Save-status (src.jsx:30)
DOM (другий child .lesson-editor-meta): `<span class="lesson-save-status"><span class="saved-dot|unsaved-dot"/>{Saved in this session|Changes are not saved} · {sections.length}</span>`. Лічильник = кількість секцій через " · ".
CSS:
- `.lesson-editor-meta{display:flex;align-items:center;justify-content:space-between;gap:var(--space-4);flex-wrap:wrap;border-block:1px solid var(--color-hairline);padding-block:var(--space-4)}` (розділювальні лінії зверху й знизу)
- `.lesson-save-status{display:flex;align-items:center;gap:var(--space-2);font-size:.6875rem;color:var(--color-muted)}`
- `.saved-dot,.unsaved-dot{inline-size:var(--space-1);block-size:var(--space-1);border-radius:var(--radius-pill);background:var(--color-success);flex-shrink:0}` · `.unsaved-dot{background:var(--color-warning)}`
- @≤52rem: `.lesson-save-status{inline-size:100%}`

## 4. Title-block (src.jsx:30)
DOM: `<div class="lesson-title-block"><label class="admin-eyebrow" for="lesson-title">Name</label>{preview? <h2>{title}</h2> : <input id="lesson-title" class="lesson-title-input" .../>}<span>{course.name} / Module 01</span></div>`
CSS:
- `.lesson-title-block{padding:var(--space-4) var(--space-2)}`
- `.lesson-title-input{display:block;inline-size:100%;min-block-size:2.75rem;padding:var(--space-2) 0;margin-block:var(--space-2);border:1px solid var(--color-background);border-radius:var(--radius-control);font:750 1.75rem/1.3 var(--font-display);background:none;color:var(--color-text);letter-spacing:-.035em}` (великий display-заголовок, невидима рамка = колір фону)
- `.lesson-title-input:hover{border-block-end-color:var(--color-border)}` (підкреслення на hover)
- `.lesson-title-block h2{font-size:1.75rem;margin-block:var(--space-3)}`
- `.lesson-title-block>span{font-size:.6875rem;color:var(--color-muted)}` (caption курс/модуль)
- @40rem container: `.lesson-title-input{font-size:1.5rem}`

## 5. Section card (src.jsx:30)
Контейнер списку: `<ol class="lesson-editor-sections {preview?'lesson-document':''}">`. Кожна секція: `<li class="lesson-editor-section ui-card" data-type={type}>`.
Bar (лише !preview): `<div class="lesson-editor-section-bar"><div class="section-identity"><span class="future-handle" aria-hidden><Icon name="drag"/></span><span class="ui-icon-tile"><Icon name={type}/></span><span class="lesson-editor-section-type">{ВІДЕО|ТЕКСТ...}</span><small>01</small></div><div class="lesson-editor-section-actions">{up}{down}{delete}</div></div>`
- index `<small>` = String(i+1).padStart(2,'0') → "01","02". drag icon path `M9 5h.01...` (6 крапок).
- up/down/delete — `<IconButton>` (up `m6 10 6-6 6 6M12 4v16`, down, delete trash). boundary up/down = unavailable (aria-disabled, no native disabled). delete = destructive.
CSS:
- база картки `.ui-card{background:var(--color-background);border:1px solid var(--color-hairline);border-radius:var(--radius-card);box-shadow:var(--card-shadow);padding:var(--card-padding)}` (--card-padding=1rem)
- `.lesson-editor-sections{list-style:none;margin:0;padding:0;display:grid;gap:var(--space-6)}`
- `.lesson-editor-section:focus-within{border-color:var(--color-border)}`
- `.lesson-editor-section-bar{display:flex;align-items:center;justify-content:space-between;gap:var(--space-3);flex-wrap:wrap;padding-block-end:var(--space-3);border-block-end:1px solid var(--color-hairline)}`
- `.section-identity{display:flex;align-items:center;gap:var(--space-3);min-inline-size:0}`
- `.section-identity small{font:.6875rem var(--font-utility);color:var(--color-faint)}` (padded index 01, моноширинний, faint)
- `.future-handle{color:var(--color-faint);display:flex}` + `.future-handle>.admin-icon{inline-size:var(--space-3)}` (тьмяний, менша іконка 0.75rem, лівий край identity, aria-hidden — декоративний)
- `.ui-icon-tile{display:grid;place-items:center;inline-size:2.75rem;block-size:2.75rem;border-radius:var(--radius-control);background:var(--color-surface);border:1px solid var(--color-hairline);color:var(--color-muted);flex-shrink:0}` (type-tile 44×44)
- `.lesson-editor-section-type` = моноширинна caps eyebrow (ВІДЕО/ТЕКСТ), див. .admin-eyebrow
- `.lesson-editor-section-actions{display:flex;gap:var(--space-1)}`
- `.admin-icon-button{padding:var(--space-2);color:var(--color-muted)}` hover→text, .is-destructive hover→danger
- @≤52rem: section padding→space-3, bar gap→space-2, actions margin-inline-start:auto

## 6. Icon-toolbar тексту (src.jsx:19)
DOM: `<div class="lesson-editor-toolbar" aria-label="Editor">` → для кожної команди `<TooltipTrigger delay=250><Button aria-label aria-pressed={editor.isActive} class="ui-button toolbar-button {i===2||i===4?'toolbar-group-start':''}">{icon?<Icon/>:<span class={I?'italic':''}>B|I|H2|H3</span>}</Button><Tooltip class="ui-tooltip">{title}</Tooltip></TooltipTrigger>` + link `<IconButton name="link" pressed={isActive('link')}/>` + `<span class="toolbar-caption">Text|Task</span>`.
Команди в порядку: [B span][I span.italic][H2 span][H3 span][list Icon][ordered Icon] + link IconButton. B/I/H2/H3 — ТЕКСТ у span (не іконка); list(`M8 6h12...` 3 рядки+крапки)/ordered(з цифрами)/link — Icon-шляхи. toolbar-group-start на i===2 (H2) та i===4 (list) → візуальні групи | formatting | headings | lists |.
active-стан = aria-pressed=true. tooltip = React Aria Tooltip, delay 250ms.
CSS:
- `.lesson-editor-toolbar{display:flex;gap:var(--space-1);align-items:center;flex-wrap:wrap;padding-block:var(--space-3);border-block-end:1px solid var(--color-hairline)}`
- `.toolbar-button{font:.8125rem var(--font-utility);font-weight:700;padding:var(--space-2);color:var(--color-muted)}` (моноширинні B/I/H2/H3)
- `.toolbar-button[aria-pressed=true],.lesson-editor-toolbar .ui-button[aria-pressed=true]{background:var(--color-elevated);color:var(--color-focus);border-color:var(--color-focus)}`
- `.toolbar-group-start{margin-inline-start:var(--space-2);border-inline-start-color:var(--color-border)}` (роздільник групи)
- `.toolbar-caption{margin-inline-start:auto;font:.6875rem var(--font-utility);color:var(--color-faint);padding-inline:var(--space-2)}` (справа, "Text"/"Task")
- `.italic{font-style:italic}`
- `.ui-tooltip{background:var(--color-elevated);color:var(--color-text);border:1px solid var(--color-border);padding:var(--space-2) var(--space-3);border-radius:var(--radius-control);box-shadow:var(--shadow-floating);font-size:.75rem;max-inline-size:var(--space-dialog);z-index:30}`
- editor body: `.lesson-editor-input{padding:var(--space-6)}` + `[contenteditable]{min-block-size:var(--space-prose-min)}`; nested: padding→space-3, toolbar-caption display:none

## 7. Link ModalBox (src.jsx:13,19)
НЕ inline-поле — окремий діалог (React Aria ModalOverlay/Modal/Dialog). Відкривається кнопкою link у toolbar.
DOM: `<ModalOverlay class="ui-modal-overlay" isDismissable><Modal class="ui-modal"><Dialog class="ui-dialog" aria-label="Add link"><div class="dialog-heading"><h2>Add link</h2><IconButton name="revoke" label="Закрити / Close"/></div>{children=<Field label="URL" type="url" hint="https://…"/>}<div class="dialog-actions"><Button class="ui-button ui-button--secondary">Cancel</Button><Button class="ui-button ui-button--primary">Confirm</Button></div></Dialog></Modal></ModalOverlay>`
Логіка Confirm: валідна http(s) → setLink; порожній URL → unsetLink. Поле — компонент Field (label зверху, .ui-input, aria-invalid/describedby).
CSS:
- `.ui-modal-overlay{position:fixed;inset:0;display:grid;place-items:center;padding:var(--space-6);background:var(--color-modal-backdrop);z-index:20;overflow-y:auto}`
- `.ui-modal{inline-size:min(100%,var(--space-dialog));background:var(--color-surface);border:1px solid var(--color-border);border-radius:var(--radius-panel);box-shadow:var(--shadow-floating);padding:var(--space-6);max-block-size:calc(100dvh - var(--space-8));overflow-y:auto}` (--space-dialog=36rem)
- `.ui-dialog{outline:none}` · `.ui-dialog>p{font-size:.875rem;color:var(--color-muted);line-height:1.65}`
- `.dialog-heading{display:flex;align-items:center;justify-content:space-between;gap:var(--space-3);margin-block-end:var(--space-4)}` · `.dialog-heading h2{font-size:1.25rem}`
- `.dialog-actions{display:flex;justify-content:flex-end;gap:var(--space-3);flex-wrap:wrap;margin-block-start:var(--space-6)}`
- Field: `.ui-field` label зверху; `.ui-input`, `.ui-description` (hint), `.ui-field-error`.

## 8. Empty-states video та image (src.jsx:25)
VIDEO (вбудований, не компонент Empty): `<div class="lesson-editor-video"><span class="ui-icon-tile"><Icon name="video"/></span><span class="admin-eyebrow">Video</span><h3>See hierarchy in action</h3><p>Attach a video in course settings</p><Button class="ui-button ui-button--secondary">Video settings<Icon name="arrow"/></Button></div>`
- `.lesson-editor-video{display:flex;flex-direction:column;align-items:center;justify-content:center;gap:var(--space-3);padding:var(--space-8);margin-block-start:var(--space-4);border:1px dashed var(--color-border);border-radius:var(--radius-control);text-align:center;background:var(--color-background)}` · icon-tile color:focus · h3 margin:0 · p .8125rem muted
IMAGE: якщо немає src → компонент `<Empty icon="image" title="Upload image" description="PNG / JPG / WebP · 2 MB"/>`; якщо src → `<img>`. Плюс (не preview) file input + Field alt-text.
Компонент Empty DOM: `<div class="ui-state"><span class="ui-icon-tile"><Icon/></span><h3>{title}</h3><p>{description}</p>{action}</div>`
- `.ui-state{display:flex;flex-direction:column;align-items:center;justify-content:center;gap:var(--space-3);text-align:center;padding:var(--space-8);color:var(--color-muted);min-block-size:var(--space-prose-min)}` · `.ui-state .ui-icon-tile{color:var(--color-focus)}` · h3 color:text · p .8125rem max 65ch
- `.lesson-editor-image{display:grid;gap:var(--space-4);padding-block-start:var(--space-4)}` · img max-inline-size:100%;border-radius:control
- `.lesson-editor-upload::file-selector-button{min-block-size:2.75rem;padding-inline:var(--space-3);margin-inline-end:var(--space-3);border:1px solid var(--color-border);border-radius:var(--radius-control);background:var(--color-elevated);color:var(--color-text)}`
Empty для всього уроку (0 секцій): `<Empty icon="lesson" title="Add section" description="Choose the first block below."/>`.

## 9. Add-section палітра (src.jsx:24,30)
Обгортка (top-level, !preview): `<fieldset class="lesson-editor-add"><legend><Icon name="plus"/>Add section</legend><p>Choose a block to continue your lesson</p>{palette()}</fieldset>`
Палітра: `<div class="lesson-editor-palette {is-nested}">` → для кожного типу `<Button class="ui-button ui-button--secondary palette-tile" aria-disabled={exists}><Icon name={key}/><strong>{label}</strong><small>{caption}</small><Icon name="plus"/></Button>`
Типи top-level: video, text, image, row, task (sectionNames). У вкладеній (nested) палітрі — лише text/image/task. Icon-шляхи: video `M4 5h16v14H4zM10 9l5 3-5 3z`, text, image `M4 4h16v16H4zM4 16l5-5...`, row `M4 4h6v16H4zM14 4h6v16h-6z`, task `M8 4h11v16H5V7...`. captions (uk): text="Ідеї та пояснення", video="Один блок на урок", image="Ілюстрації й схеми", row="2–3 колонки", task="Практика для студентів". Video, якщо вже є → aria-disabled + "Вже в уроці".
CSS:
- `.lesson-editor-add{border:1px dashed var(--color-border);border-radius:var(--radius-panel);padding:var(--space-5);margin:0}` · `legend{display:flex;align-items:center;gap:var(--space-2);padding-inline:var(--space-3);font-size:.875rem;font-weight:700}` · legend svg color:focus · `>p{font-size:.6875rem;color:var(--color-muted);margin-block-end:var(--space-4)}`
- `.lesson-editor-palette{display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,var(--space-tile-min)),1fr));gap:var(--space-3)}` (--space-tile-min=9rem)
- `.palette-tile{display:grid;grid-template-columns:minmax(0,1fr) auto;justify-items:start;text-align:start;padding:var(--space-4);gap:var(--space-2);background:var(--color-background);border-color:var(--color-border);box-shadow:var(--card-shadow)}` hover: bg surface, border focus
- `.palette-tile>svg:first-child{color:var(--color-muted);grid-column:1}` · `strong{grid-column:1;font-size:.8125rem}` · `small{grid-column:1;font-size:.6875rem;color:var(--color-muted);font-weight:650}` · `>svg:last-child{grid-column:2;grid-row:1;color:var(--color-faint);inline-size:var(--space-4)}` (plus-іконка у правому-верхньому кутку)
- nested: `.lesson-editor-palette.is-nested{grid-template-columns:minmax(0,1fr)}`; `.is-nested .palette-tile{grid-template-columns:auto minmax(0,1fr);align-items:center}`; is-nested ховає small та plus-іконку.

## Row/columns (контекст для секції type=row, src.jsx:25)
- `.column-settings{display:flex;gap:var(--space-2);align-items:center;padding-block:var(--space-4)}` (перемикач 2/3 колонки, aria-pressed=true→focus колір)
- `.lesson-editor-columns{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-4)}` · `[data-columns="3"]{repeat(3,...)}`
- `.lesson-editor-column{border:1px dashed var(--color-border);padding:var(--space-3);border-radius:var(--radius-control);display:grid;align-content:start;gap:var(--space-3)}`
- @container(inline-size<40rem): columns→1fr. (README: 3→2 <60rem, →1 <40rem — через container queries на canvas)
- `.nested-section` всередині колонок; preview (.lesson-document) прибирає рамки/padding/box-shadow.

## Preview-режим (.lesson-document)
`.lesson-document .lesson-editor-section{border-color:var(--color-background);box-shadow:none}` — у перегляді картки стають безрамковими; task/video прибирають margin-top; колонки без рамки.
