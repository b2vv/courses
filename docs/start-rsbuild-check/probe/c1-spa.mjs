// M1 C1: SPA navigation through nginx, anonymous. Diagnostic only.
import { chromium } from '<repo>/frontend/node_modules/playwright/index.mjs';

const TITLE = 'Go: фундамент надійного бекенду';
const out = { steps: [], documents: [], pageErrors: [], errors: [] };
const browser = await chromium.launch();
try {
  const page = await (await browser.newContext({ baseURL: 'http://localhost:8080' })).newPage();
  page.on('request', (r) => {
    if (r.resourceType() === 'document' && r.frame() === page.mainFrame()) out.documents.push(r.url());
  });
  page.on('pageerror', (e) => out.pageErrors.push(String(e.message).slice(0, 200)));
  const step = async (name, fn) => {
    try {
      await fn();
      out.steps.push({ name, ok: true, url: new URL(page.url()).pathname, marker: await page.evaluate(() => window.__m1 ?? null) });
    } catch (e) {
      out.steps.push({ name, ok: false, error: String(e.message).split('\n')[0] });
      throw e;
    }
  };
  const visible = (loc) => loc.waitFor({ state: 'visible', timeout: 10000 });
  await step('goto /courses', async () => {
    const r = await page.goto('/courses');
    if (r?.status() !== 200) throw new Error(`status ${r?.status()}`);
    await visible(page.getByRole('heading', { name: 'Каталог курсів' }));
    await visible(page.getByRole('heading', { name: TITLE }));
    await page.evaluate(() => (window.__m1 = 1));
  });
  await step('click card', async () => {
    await page.getByRole('heading', { name: TITLE }).getByRole('link').click();
    await page.waitForURL('**/courses/go-backend-foundations', { timeout: 10000 });
    await visible(page.getByRole('heading', { level: 1, name: TITLE }));
  });
  await step('goBack', async () => {
    await page.goBack();
    await page.waitForURL((u) => u.pathname === '/courses', { timeout: 10000 });
    await visible(page.getByRole('heading', { name: 'Каталог курсів' }));
  });
} catch (e) {
  out.errors.push(String(e.message).split('\n')[0]);
} finally {
  await browser.close();
}
const ok =
  out.errors.length === 0 &&
  out.steps.length === 3 &&
  out.steps.every((s) => s.ok && s.marker === 1) &&
  out.documents.length === 1 &&
  out.pageErrors.length === 0;
out.verdict = ok ? 'PASS' : 'FAIL';
console.log(JSON.stringify(out));
