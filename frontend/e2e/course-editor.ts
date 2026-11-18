import { expect, type Locator, type Page } from '@playwright/test';

/** The caller selects Lessons; reopening an expanded row must leave it expanded. */
export async function openLesson(
  editor: Page | Locator,
  module: number,
  lesson: number,
): Promise<Locator> {
  const row = editor
    .getByRole('group', { name: new RegExp(`^(Модуль|Module) ${module}$`) })
    .getByRole('group', { name: new RegExp(`^(Урок|Lesson) ${lesson}$`) });
  const toggle = row.locator('.admin-lesson-summary button[aria-expanded][aria-controls]');
  if ((await toggle.getAttribute('aria-expanded')) === 'false') await toggle.click();
  await expect(toggle).toHaveAttribute('aria-expanded', 'true');
  const bodyId = await toggle.getAttribute('aria-controls');
  if (bodyId === null) throw new Error('Lesson toggle has no controlled body');
  await expect(row.locator(`[id="${bodyId}"]`)).toBeVisible();
  return row;
}
