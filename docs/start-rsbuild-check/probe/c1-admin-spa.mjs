// M1 C1 (amended 2026-11-20): SPA navigation over Router <Link> in the admin shell. Read-only.
import { chromium } from '<repo>/frontend/node_modules/playwright/index.mjs';

const USERS = 'Користувачі';
const STATE = '<repo>/frontend/playwright/.auth/admin.json';
const out = { steps: [], documents: [], pageErrors: [], errors: [] };
const browser = await chromium.launch();
try {
  const page = await (await browser.newContext({ baseURL: 'http://localhost:8080', storageState: STATE })).newPage();
  page.on('request', (r) => {
    if (r.resourceType() === 'document' && r.frame() === page.mainFrame()) out.documents.push(new URL(r.url()).pathname);
  });
  page.on('pageerror', (e) => out.pageErrors.push(String(e.message).slice(0, 200)));
  const h1 = (name) => page.getByRole('heading', { level: 1, name, exact: true }).waitFor({ state: 'visible', timeout: 15000 });
  const step = async (name, fn) => {
    try {
      await fn();
      out.steps.push({ name, ok: true, url: new URL(page.url()).pathname, marker: await page.evaluate(() => window.__m1 ?? null), h1: (await page.getByRole('heading', { level: 1 }).first().textContent())?.trim() });
    } catch (e) {
      out.steps.push({ name, ok: false, url: new URL(page.url()).pathname, error: String(e.message).split('\n')[0] });
      throw e;
    }
  };
  await step('goto /admin', async () => {
    const r = await page.goto('/admin');
    if (r?.status() !== 200) throw new Error(`status ${r?.status()}`);
    await page.locator('nav.admin-nav a[href="/admin/analytics"]').waitFor({ state: 'visible', timeout: 15000 });
    await h1('Курси');
    await page.evaluate(() => (window.__m1 = 1));
  });
  await step('click analytics', async () => {
    await page.locator('nav.admin-nav a[href="/admin/analytics"]').click();
    await page.waitForURL((u) => u.pathname === '/admin/analytics', { timeout: 15000 });
    await h1('Аналітика');
  });
  await step('click users', async () => {
    await page.locator('nav.admin-nav a[href="/admin/users"]').click();
    await page.waitForURL((u) => u.pathname === '/admin/users', { timeout: 15000 });
    await h1(USERS);
  });
  await step('goBack', async () => {
    await page.goBack();
    await page.waitForURL((u) => u.pathname === '/admin/analytics', { timeout: 15000 });
    await h1('Аналітика');
  });
} catch (e) {
  out.errors.push(String(e.message).split('\n')[0]);
} finally {
  await browser.close();
}
const ok =
  out.errors.length === 0 && out.steps.length === 4 && out.steps.every((s) => s.ok && s.marker === 1) &&
  out.documents.length === 1 && out.pageErrors.length === 0;
out.verdict = ok ? 'PASS' : 'FAIL';
console.log(JSON.stringify(out));
