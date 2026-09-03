# UI kit foundation

## Design decisions

Artifact: a system for students finding a course and, later, instructors maintaining its material. This delivery provides reusable controls and public catalog states; it does not implement instructor actions.

The repository had no existing UI inventory or tokens when this task started. `frontend/src/app/styles/tokens.css` is the single token source. The final palette follows the user-provided CourseForge reference: white/dark slate canvases, purple actions and covers, lime hero emphasis, bold system headings and utility monospace. Card cover imagery is CSS/code typography; it makes no product claim. Components use semantic CSS classes, React Aria data attributes, and tokens; feature components do not introduce literal colors.

Light/dark uses `prefers-color-scheme`, overridden by `data-theme="light"` or `data-theme="dark"` on the root element. `--font-body`, `--font-display`, `--font-utility`, semantic `--color-*`, spacing/radius tokens and `--control-min-size` are shared. Controls have a minimum 2.75rem target, visible focus, wrapping labels, no fixed text heights, reduced-motion treatment and forced-color fallbacks. Wide tables scroll in a keyboard-focusable named section; a targeted lint exception documents why that section is focusable.

## Public API

Import from `src/presentation/ui/index.ts`. Types come from React Aria where applicable. No transport calls, domain policies, server state stores or global queues exist here.

| Component | API / behavior |
|---|---|
| Button | React Aria Button props; `variant: primary / secondary / danger`; `onPress`, disabled and keyboard behavior |
| Link | React Aria Link props; ordinary `href` supports document navigation and SSR recovery |
| TextField | `label`, optional `description`, `errorMessage`, `placeholder`, plus React Aria field props including controlled `value/onChange` |
| TextArea | Same field contract, multiline and vertically resizable |
| Select | `label`, `options: {id,label}[]`, optional help/errors, React Aria selection props |
| ComboBox | Same option shape, labelled text input, client-side option filtering, React Aria selection/input props |
| Checkbox | Label as `children`; React Aria checked/disabled props |
| Switch | Label as `children`; React Aria selection/disabled props |
| Dialog | `triggerLabel`, `title`, `children`; modal focus containment, Escape/close button and trigger focus restoration |
| Menu | `label`, `items: {id,label,isDisabled?}[]`, `onAction(key)` |
| Tabs | `label`, `items: {id,label,content}[]`, plus React Aria selected-key props; caller controls URL state when relevant |
| Tooltip | `content: string`, one React Aria pressable child; hover/focus help |
| ProgressBar | `label`, React Aria `value`, min/max and indeterminate props |
| Table | `label`, `columns: {id,label}[]`, `rows: {id,cells: ReactNode[]}[]`, optional `emptyMessage`; cells follow column order |
| Pagination | One-based `page`, `totalPages`, `onPageChange`; previous/next boundary controls; caller owns URL navigation |
| FileUpload | `label`, optional `isDisabled`, React Aria FileTrigger props including `onSelect`, accepted file types and multiple selection |
| Toast | Controlled `message/onDismiss`; polite status announcement and keyboard dismiss action |
| EmptyState | `title`, optional explanatory/action `children` |
| ErrorState | `title`, `children`, optional `retryHref`; keeps recovery URL intact |
| Skeleton | Optional `label/lines`; status announcement, decorative bars hidden from assistive technology |

## Verification

The user specified the testing seams: React component behavior, keyboard navigation and focus. No additional approval was needed to use TDD at these seams. The first test failed on the absent UI export before implementation; the pagination boundary test also failed before its component was implemented. A test-runner JSX configuration issue was fixed in the shared Rstest configuration.

`node node_modules/@rstest/core/bin/rstest.js run src/presentation/ui/ui.test.tsx` passed 10 tests, 0 skipped:

- Label/description associations and keyboard activation for TextField/Button.
- Modal name, contained focus, Escape and trigger focus restoration.
- Labelled Select open/choose/close and focus restoration.
- Space activation of Checkbox and Switch.
- Last-page pagination boundaries and back navigation.
- ComboBox filtering while typing.
- Keyboard tab selection and corresponding labelled panel.
- Disabled menu items and action dispatch.
- Validation message association, complete retry URL and progress semantics.
- Polite toast status and keyboard dismissal.

`npm run typecheck` passed. Targeted Oxlint passed with zero warnings. Final formatting and combined gates are recorded in the parent foundation report.

Happy DOM has no layout geometry, so responsive behavior remained a browser gate. The final Playwright run passed keyboard behavior, axe WCAG A/AA, dark/mobile, 200% text zoom without horizontal overflow, and `uk/en` switching; screenshots are linked from `report.md`.

## Limits

Table is a semantic read-only table, without sorting/virtualization/editing. FileUpload selects files but does not upload them. Toast is controlled by its caller, without automatic dismissal or an application queue. Dialog is a basic modal with an included trigger and close action. Transcoding, actual progress, authentication and administration workflows are later phases. No unused TanStack Table/Virtual, upload store, form library or business interface was added for those future features.

Next smallest task: specify Phase 3 Identity and session behavior. No user decision is required.
