# ADR 003 — Own UI kit over React Aria
Status: accepted, 2026-09-03.

Wrap React Aria Components with semantic design tokens and application copy. The library supplies keyboard/focus/ARIA behavior; local components supply consistent sizing, validation messages and themes. Avoid duplicating library interaction machinery. Test critical dialog/select/form/navigation behavior and browser zoom. Native semantic markup is appropriate for non-interactive content and simple tables.
