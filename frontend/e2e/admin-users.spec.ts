import { randomUUID } from 'node:crypto';
import { expect, test, type Page } from '@playwright/test';
import { authStatePath, trackPageRequests } from './auth';

const resetPassword = `t08-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;

const users = {
  block: {
    id: '40000000-0000-4000-8000-0000000000d1',
    email: 'block-me@courseforge.local',
    name: 'Тест блокування',
  },
  delete: {
    id: '40000000-0000-4000-8000-0000000000d2',
    email: 'delete-me@courseforge.local',
    name: 'Тест видалення',
  },
  reset: {
    id: '40000000-0000-4000-8000-0000000000d3',
    email: 'reset-me@courseforge.local',
    name: 'Тест скидання пароля',
  },
} as const;

let finishRequestTracking = () => Promise.resolve();

test.beforeEach(({ page }) => {
  finishRequestTracking = trackPageRequests(page);
});

test.afterEach(async () => {
  await finishRequestTracking();
});

function userRow(page: Page, email: string) {
  return page.getByRole('row').filter({ hasText: email });
}

test.describe('with a stored admin session', () => {
  test.use({ storageState: authStatePath('admin') });

  test.beforeAll(async ({ browser, baseURL }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const context = await browser.newContext({
      baseURL,
      storageState: authStatePath('admin'),
    });
    try {
      for (const user of Object.values(users)) {
        const message = `${user.email} is not an active disposable fixture; run make seed-identities`;
        const response = await context.request.get(
          `/api/v1/admin/users?q=${encodeURIComponent(user.email)}`,
        );
        expect(response.status(), message).toBe(200);
        expect(await response.json(), message).toMatchObject({
          items: [{ id: user.id, email: user.email, role: 'student', status: 'active' }],
        });
      }
    } finally {
      await context.close();
    }
  });

  test('P08 blocks, restores, resets, and deletes disposable users', async ({ page }) => {
    await page.goto('/admin/users');
    await expect(page.getByRole('heading', { name: 'Користувачі' })).toBeVisible();

    const blockRow = userRow(page, users.block.email);
    await expect(blockRow).toContainText('Активний');
    await blockRow.getByRole('button', { name: `Заблокувати ${users.block.name}` }).click();
    await expect(blockRow).toContainText('Заблокований');
    await blockRow.getByRole('button', { name: `Відновити ${users.block.name}` }).click();
    await expect(blockRow).toContainText('Активний');

    const resetRow = userRow(page, users.reset.email);
    await resetRow.getByRole('button', { name: `Скинути пароль ${users.reset.name}` }).click();
    const resetDialog = page.getByRole('dialog', {
      name: `Новий пароль для ${users.reset.name}`,
    });
    await resetDialog.getByLabel('Новий пароль').fill(resetPassword);
    await resetDialog.getByRole('button', { name: 'Зберегти новий пароль' }).click();
    await expect(resetDialog.getByRole('status')).toHaveText('Пароль оновлено.');
    await resetDialog.getByRole('button', { name: 'Закрити' }).click();

    const deleteRow = userRow(page, users.delete.email);
    // delete-me is consumed for this seed; the reviewer reseeds before the next e2e run.
    await deleteRow.getByRole('button', { name: `Видалити ${users.delete.name}` }).click();
    await expect(page.getByRole('status')).toHaveText('Обліковий запис видалено.');
    await expect(deleteRow).toHaveCount(0);

    await page.getByRole('radio', { name: 'Видалений' }).click();
    const deletedRow = userRow(page, users.delete.email);
    await expect(deletedRow).toContainText('Видалений');
    await expect(deletedRow.getByText('Немає доступних дій')).toBeVisible();
  });
});
