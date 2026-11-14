import { chromium, devices } from '/Users/strelia/projects/courses/frontend/node_modules/playwright/index.mjs';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
const root = process.argv[2];
const browser = await chromium.launch();
const context = await browser.newContext({ ...devices['Desktop Chrome'] });
const page = await context.newPage();
await page.route('http://probe.test/**', async (route) => {
  const p = new URL(route.request().url()).pathname;
  const file = path.join(root, p === '/' ? 'index.html' : p);
  const type = file.endsWith('.html') ? 'text/html' : file.endsWith('.js') ? 'text/javascript' : file.endsWith('.m3u8') ? 'application/vnd.apple.mpegurl' : 'video/mp4';
  try { await route.fulfill({ body: await readFile(file), contentType: type }); } catch { await route.fulfill({ status: 404 }); }
});
await page.goto('http://probe.test/');
await page.waitForTimeout(3000);
console.log('pre', JSON.stringify(await page.evaluate(() => ({ err: window.playErr ?? null, t: document.getElementById('v').currentTime, rs: document.getElementById('v').readyState, mse: MediaSource.isTypeSupported('video/mp4; codecs="avc1.64000c,mp4a.40.2"') }))));
await page.click('video');
await page.evaluate(() => document.getElementById('v').play().catch(e => (window.playErr2 = String(e))));
await page.waitForFunction(() => document.getElementById('v').currentTime > 1, null, { timeout: 15000 });
const sample = async (label, n = 6) => { for (let i = 0; i < n; i++) { console.log(label, JSON.stringify(await page.evaluate(() => window.peak()))); await page.waitForTimeout(300); } };
await sample('before');
const t0 = Date.now();
await page.evaluate(() => { window.hls.nextAudioTrack = 1; });
for (let i = 0; i < 30; i++) { const r = await page.evaluate(() => window.peak()); console.log('after', Date.now() - t0, JSON.stringify(r)); await page.waitForTimeout(300); }
console.log('playErr', await page.evaluate(() => window.playErr ?? null));
await browser.close();
