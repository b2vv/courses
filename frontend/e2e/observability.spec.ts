import { expect, test, type Page } from '@playwright/test';

// AC1: a client-side error is captured and shipped as exactly one event
// carrying route + release, transported through the port's adapter. Channel
// separation (no dedup) guarantees one failure -> one event; a swallowed 429
// would be a non-204 status and must fail the test. Two of the three AC1
// channels have a deterministic browser trigger (a thrown error reaches
// window 'error'; an unhandled rejection reaches 'unhandledrejection'); the
// onCaughtError boundary channel funnels into the same capture() and is
// covered by the adapter unit tests.
async function armIngestCapture(page: Page) {
  const bodies: string[] = [];
  const statuses: number[] = [];
  page.on('request', (request) => {
    if (request.method() === 'POST' && request.url().includes('/api/v1/client-logs')) {
      bodies.push(request.postData() ?? '');
    }
  });
  page.on('response', (response) => {
    if (response.request().method() === 'POST' && response.url().includes('/api/v1/client-logs')) {
      statuses.push(response.status());
    }
  });
  // Module scripts run before `load`, so the client entry has installed the
  // error bridge by the time navigation resolves; a short cushion covers
  // hydration ordering.
  await page.goto('/', { waitUntil: 'load' });
  await page.waitForTimeout(500);
  bodies.length = 0;
  statuses.length = 0;
  return { bodies, statuses };
}

async function expectExactlyOne204(
  page: Page,
  bodies: string[],
  statuses: number[],
  probe: string,
) {
  await expect.poll(() => statuses.length, { timeout: 5000 }).toBe(1);
  expect(statuses[0]).toBe(204);
  expect(bodies[0]).toContain(probe);
  expect(bodies[0]).toContain('"release"');
  expect(bodies[0]).toContain('"route"');
  // Bounded wait: no second POST in a quiet window (no dedup, one event).
  await page.waitForTimeout(1500);
  expect(bodies).toHaveLength(1);
  expect(statuses).toEqual([204]);
}

test('a thrown error produces exactly one /client-logs POST returning 204', async ({ page }) => {
  const { bodies, statuses } = await armIngestCapture(page);
  const probe = `e2e-observability-error-${Date.now()}`;
  await page.evaluate((message) => {
    window.dispatchEvent(new ErrorEvent('error', { error: new Error(message), message }));
  }, probe);
  await expectExactlyOne204(page, bodies, statuses, probe);
});

test('an unhandled rejection produces exactly one /client-logs POST returning 204', async ({
  page,
}) => {
  const { bodies, statuses } = await armIngestCapture(page);
  const probe = `e2e-observability-rejection-${Date.now()}`;
  await page.evaluate((message) => {
    // A rejected promise with no handler fires 'unhandledrejection'.
    void Promise.reject(new Error(message));
  }, probe);
  await expectExactlyOne204(page, bodies, statuses, probe);
});
