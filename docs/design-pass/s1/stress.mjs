// design-pass S1 stress: replaces the text of data-bearing elements with one 190-character unbroken
// string (a valid course title is up to 200) and reports anything that leaves the viewport or the
// main column. Usage: S=<dir with auth/> [ONLY=id,id] [CSS=<file>] node stress.mjs > out.json
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
const FE=fileURLToPath(new URL('../../../frontend', import.meta.url));
const require=createRequire(FE+'/package.json');
const { chromium } = require('@playwright/test');
const S=process.env.S, BASE='http://localhost:8080';
const only=process.env.ONLY?.split(',');
const css=process.env.CSS?readFileSync(process.env.CSS,'utf8'):null;
const surfaces=JSON.parse(readFileSync(new URL('./surfaces.json', import.meta.url),'utf8')).filter(s=>!only||only.includes(s[0]));
const V={phone:[390,844,false],p200:[390,844,true],desk:[1440,900,false]};
const stress=()=>{
  const long='Ш'.repeat(190);
  const targets=[...document.querySelectorAll('main h1, main h2, main h3, main p, main a, main td, main li > strong, main dd')]
    .filter(e=>e.childElementCount===0 && e.textContent.trim().length>2 && e.getClientRects().length);
  for (const e of targets) e.textContent=long;
  const W=document.documentElement.clientWidth, main=document.querySelector('main')?.getBoundingClientRect();
  const scr=e=>{for(let p=e.parentElement;p&&p!==document.body;p=p.parentElement){const o=getComputedStyle(p).overflowX; if(o!=='visible') return true;} return false;};
  const nm=e=>e.tagName.toLowerCase()+'.'+[...e.classList].slice(0,2).join('.');
  const out=[...document.querySelectorAll('main *')].filter(e=>{const r=e.getBoundingClientRect(); return r.width>0 && !scr(e) && (r.right>W+1 || (main && r.right>main.right+1));}).slice(0,8).map(e=>nm(e)+' r'+Math.round(e.getBoundingClientRect().right));
  return {targets:targets.length, docOverflow:document.documentElement.scrollWidth-W, mainRight:main&&Math.round(main.right), spill:out};
};
const browser=await chromium.launch(); const res={};
for (const [id,role,path] of surfaces){ res[id]={};
  for (const [k,[w,h,t200]] of Object.entries(V)){
    const ctx=await browser.newContext({viewport:{width:w,height:h},reducedMotion:'reduce',...(role!=='guest'?{storageState:`${S}/auth/${role}.json`}:{})});
    await ctx.addCookies([{name:'courseforge_locale',value:'uk',url:BASE}]); const page=await ctx.newPage();
    try{ await page.goto(BASE+path,{waitUntil:'load',timeout:30000}); await page.waitForTimeout(1200);
      if(css) await page.addStyleTag({content:css});
      if(t200) await page.addStyleTag({content:'html{font-size:200% !important}'});
      await page.waitForTimeout(200); res[id][k]=await page.evaluate(stress);
    }catch(e){ res[id][k]={error:e.message.slice(0,150)}; }
    await ctx.close(); }}
await browser.close(); console.log(JSON.stringify(res,null,1));
