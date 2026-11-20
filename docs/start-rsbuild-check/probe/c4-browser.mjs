// M1 C4 browser half: the route error component is visible on /m1-throw. Diagnostic only.
import { chromium } from '<repo>/frontend/node_modules/playwright/index.mjs';
const out = { pageErrors: [], console: [] };
const browser = await chromium.launch();
try {
  const page = await (await browser.newContext({ baseURL: 'http://localhost:8080' })).newPage();
  page.on('pageerror', (e) => out.pageErrors.push(String(e.message).slice(0, 160)));
  page.on('console', (m) => m.type() === 'error' && out.console.push(m.text().slice(0, 160)));
  const r = await page.goto('/m1-throw');
  out.status = r?.status();
  const alert = page.getByRole('alert').filter({ has: page.getByRole('heading', { level: 1 }) });
  await alert.waitFor({ state: 'visible', timeout: 15000 });
  out.alertH1 = (await alert.getByRole('heading', { level: 1 }).textContent())?.trim();
  out.retryButton = await alert.getByRole('button').count();
  out.verdict = 'VISIBLE';
} catch (e) {
  out.verdict = 'NOT_VISIBLE';
  out.error = String(e.message).split('\n')[0];
} finally {
  await browser.close();
}
console.log(JSON.stringify(out));
