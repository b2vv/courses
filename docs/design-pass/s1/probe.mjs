// design-pass S1 probe: broken words (WIDE = word wider than its column, FITS = it would fit) and
// horizontal overflow per surface and viewport. Logins come from discover.mjs (S/auth/<role>.json).
// Usage: S=<dir> [CSS=<file>] [ONLY=id,id] [VARS=p200,r320] [LOCALE=uk] [SHOTS=<dir>] node probe.mjs > out.json
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
const FE=fileURLToPath(new URL('../../../frontend', import.meta.url)); // repo-relative, resolved from this file
const require=createRequire(FE+'/package.json');
const { chromium } = require('@playwright/test');
const S=process.env.S, BASE='http://localhost:8080';
const css=process.env.CSS?readFileSync(process.env.CSS,'utf8'):null;
const only=process.env.ONLY?.split(','); const vars=process.env.VARS?.split(',');
const surfaces=JSON.parse(readFileSync(new URL('./surfaces.json', import.meta.url),'utf8')).filter(s=>!only||only.includes(s[0]));
const V={phone:[390,844,false],p200:[390,844,true],r320:[320,720,false],tab:[768,1024,false],desk:[1440,900,false],d200:[1440,900,true]};
const measure=()=>{const W=innerWidth; const broken=[]; const w=document.createTreeWalker(document.body,NodeFilter.SHOW_TEXT); let n;
 while((n=w.nextNode())){const el=n.parentElement; if(!el||!el.getClientRects().length||el.closest('[aria-hidden="true"],.ui-visually-hidden,script,style')) continue; const re=/\S+/g; let m;
  while((m=re.exec(n.data))){ if(m[0].length<2) continue; const r=document.createRange(); r.setStart(n,m.index); r.setEnd(n,m.index+m[0].length); const tops=new Set([...r.getClientRects()].filter(x=>x.width>0).map(x=>Math.round(x.top))); if(tops.size>1){ const cs=getComputedStyle(el); const c=document.createElement('canvas').getContext('2d'); c.font=cs.font; c.letterSpacing=cs.letterSpacing; const ww=c.measureText(m[0]).width; let blk=el; while(blk&&getComputedStyle(blk).display.startsWith('inline')) blk=blk.parentElement; const bcs=getComputedStyle(blk); const cw=blk.clientWidth-parseFloat(bcs.paddingLeft)-parseFloat(bcs.paddingRight); broken.push((ww>cw+0.5?'WIDE ':'FITS ')+Math.round(ww)+'/'+Math.round(cw)+' '+m[0].slice(0,30)+' <'+el.tagName.toLowerCase()+(el.classList.length?'.'+[...el.classList].slice(0,2).join('.'):'')+'>'); }}}
 const scr=e=>{for(let p=e.parentElement;p&&p!==document.body;p=p.parentElement){const o=getComputedStyle(p).overflowX; if(o!=='visible') return true;} return false;};
 const over=[...document.querySelectorAll('body *')].filter(e=>{const r=e.getBoundingClientRect(); return r.width>0&&r.right>W+1&&!scr(e)}).slice(0,6).map(e=>e.tagName.toLowerCase()+'.'+[...e.classList].slice(0,2).join('.')+' r'+Math.round(e.getBoundingClientRect().right));
 return {docOverflow:document.documentElement.scrollWidth-W, broken, over, title:document.title};};
const browser=await chromium.launch(); const out={};
for (const [id,role,path] of surfaces){ out[id]={};
 for (const [k,[w,h,t200]] of Object.entries(V)){ if(vars&&!vars.includes(k)) continue;
  const ctx=await browser.newContext({viewport:{width:w,height:h}, reducedMotion:'reduce', ...(role!=='guest'?{storageState:`${S}/auth/${role}.json`}:{})});
  await ctx.addCookies([{name:'courseforge_locale',value:process.env.LOCALE||'uk',url:BASE}]);
  const page=await ctx.newPage();
  try{ await page.goto(BASE+path,{waitUntil:'load',timeout:30000}); await page.waitForTimeout(1200);
   if(css) await page.addStyleTag({content:css}); if(t200) await page.addStyleTag({content:'html{font-size:200% !important}'}); await page.waitForTimeout(300);
   out[id][k]=await page.evaluate(measure); if(process.env.SHOTS) await page.screenshot({path:`${process.env.SHOTS}/${id}-${k}.png`,fullPage:true});
  }catch(e){ out[id][k]={error:e.message.slice(0,150)}; }
  await ctx.close(); }}
await browser.close(); console.log(JSON.stringify(out,null,1));
