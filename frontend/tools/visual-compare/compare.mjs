// Visual compare harness.
//
// Philosophy (borrowed from cassiopeia-admin-ui's .ds-sync/storybook/compare.mjs):
// capture the TWO TRUE images — the design reference and our own render — and
// pair them into ONE sheet per surface/width. It does NOT compute pixel-diff or
// similarity heuristics: those mislead whenever layout, copy or framing
// legitimately differ. The judge is a human/agent reading the real screenshots.
//
// Usage:
//   node tools/visual-compare/compare.mjs
//   REF_URL=https://… LOCAL_URL=http://localhost:8080 PP_WIDTHS=1280,390 \
//     node tools/visual-compare/compare.mjs
//
// Output (gitignored): tools/visual-compare/out/<surface>-<width>.png sheets,
// with the raw single shots under out/raw/. Requires the local stack to be up
// (LOCAL_URL reachable) and playwright installed in frontend/.
import { chromium } from 'playwright';
import { mkdirSync, writeFileSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const HERE = dirname(fileURLToPath(import.meta.url));
const REF = process.env.REF_URL ?? 'https://courseforge-academy.strelchenko-dev.chatgpt.site';
const LOCAL = process.env.LOCAL_URL ?? 'http://localhost:8080';
const WIDTHS = (process.env.PP_WIDTHS ?? '1280,390').split(',').map((n) => Number(n.trim()));
const SCHEME = process.env.PP_SCHEME ?? 'light';
const OUT = join(HERE, 'out');
const RAW = join(OUT, 'raw');
mkdirSync(RAW, { recursive: true });

// Each surface: how to reach the region on the reference (a one-pager) and on
// ours. `ready` is a painted selector we wait for (never networkidle alone);
// `scrollTo` brings a lower region (the cards) into the viewport.
const SURFACES = [
  {
    name: 'hero',
    ref: { path: '/', ready: 'h1', scrollTo: null },
    ours: { path: '/', ready: 'h1', scrollTo: null },
  },
  {
    name: 'cards',
    ref: { path: '/', ready: 'article', scrollTo: 'article' },
    ours: { path: '/courses', ready: '.course-card', scrollTo: '.course-card' },
  },
];

const KILL_ANIM =
  '*,*::before,*::after{animation:none!important;transition:none!important;caret-color:transparent!important;scroll-behavior:auto!important}';

async function shot(browser, url, width, { ready, scrollTo }, file) {
  const ctx = await browser.newContext({
    viewport: { width, height: 900 },
    colorScheme: SCHEME,
    deviceScaleFactor: 2,
    reducedMotion: 'reduce',
  });
  const page = await ctx.newPage();
  await page.goto(url, { waitUntil: 'networkidle', timeout: 45000 }).catch(() => {});
  await page.addStyleTag({ content: KILL_ANIM }).catch(() => {});
  await page.evaluate(() => document.fonts && document.fonts.ready).catch(() => {});
  await page.waitForSelector(ready, { timeout: 15000 }).catch(() => {});
  if (scrollTo) {
    await page
      .evaluate((sel) => {
        const el = document.querySelector(sel);
        if (el) el.scrollIntoView({ block: 'start', behavior: 'instant' });
        window.scrollBy(0, -24);
      }, scrollTo)
      .catch(() => {});
    await page.waitForTimeout(250);
  }
  await page.waitForTimeout(300);
  await page.screenshot({ path: file });
  await ctx.close();
}

// Compose a labelled side-by-side sheet (reference | ours) from two PNGs.
async function sheet(browser, refPng, oursPng, width, title, file) {
  const b64 = (p) => 'data:image/png;base64,' + readFileSync(p).toString('base64');
  const html = `<!doctype html><meta charset=utf8><style>
    :root{color-scheme:light}
    body{margin:0;background:#0f1420;font:14px/1.4 ui-sans-serif,system-ui,sans-serif;color:#e7ebf3}
    h1{font-size:15px;margin:0;padding:14px 18px;font-weight:600;letter-spacing:.01em}
    .row{display:flex;gap:2px;align-items:flex-start;background:#2a3550}
    .col{flex:1;min-width:0;background:#0f1420}
    .cap{padding:8px 14px;font-size:12px;font-weight:600;text-transform:uppercase;letter-spacing:.08em}
    .cap.a{color:#b9ff4a}.cap.b{color:#8fb4ff}
    img{display:block;width:100%;height:auto}
  </style>
  <h1>${title}</h1>
  <div class="row">
    <div class="col"><div class="cap a">Reference mockup</div><img src="${b64(refPng)}"></div>
    <div class="col"><div class="cap b">Our render</div><img src="${b64(oursPng)}"></div>
  </div>`;
  const ctx = await browser.newContext({
    viewport: { width: width * 2 + 60, height: 900 },
    deviceScaleFactor: 1,
  });
  const page = await ctx.newPage();
  await page.setContent(html, { waitUntil: 'networkidle' });
  await page.waitForTimeout(150);
  await page.screenshot({ path: file, fullPage: true });
  await ctx.close();
}

const browser = await chromium.launch();
const sheets = [];
for (const w of WIDTHS) {
  for (const s of SURFACES) {
    const refPng = join(RAW, `${s.name}-${w}-ref.png`);
    const oursPng = join(RAW, `${s.name}-${w}-ours.png`);
    await shot(browser, REF + s.ref.path, w, s.ref, refPng);
    await shot(browser, LOCAL + s.ours.path, w, s.ours, oursPng);
    const sheetFile = join(OUT, `${s.name}-${w}.png`);
    await sheet(browser, refPng, oursPng, w, `${s.name.toUpperCase()} · ${w}px`, sheetFile);
    sheets.push(sheetFile);
    console.log('sheet:', sheetFile);
  }
}
await browser.close();
writeFileSync(join(OUT, 'index.json'), JSON.stringify({ ref: REF, local: LOCAL, sheets }, null, 2));
console.log('done', sheets.length, 'sheets ->', OUT);
