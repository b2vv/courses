import { expect, test } from '@playwright/test';
import { authFixtures, authStatePath, recordRequest, resetLoginCount } from './auth';

const roles = ['student', 'instructor', 'admin'] as const;

test.describe.configure({ mode: 'serial' });
test.beforeAll(resetLoginCount);

for (const role of roles) {
  test(`authenticate as ${role}`, async ({ request, baseURL }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const { email, password } = authFixtures[role];
    const response = await request.post('/api/v1/auth/login', {
      data: { email, password },
      headers: { Origin: new URL(baseURL).origin },
    });
    await recordRequest('login');
    expect(response.status()).toBe(200);
    await request.storageState({ path: authStatePath(role) });
  });
}
