import { createRequire } from 'node:module';
import { mkdirSync, writeFileSync, readFileSync, appendFileSync } from 'node:fs';
const FE='/Users/strelia/projects/courses/.claude/worktrees/dg-archive-test-c6c9f4/frontend';
const require=createRequire(FE+'/package.json');
const { chromium } = require('@playwright/test');
const AXE=readFileSync(FE+'/node_modules/axe-core/axe.min.js','utf8');
const S=process.env.S, BASE=process.env.BASE||'http://localhost:8080', OUT=process.env.OUT||'shots';
const only=process.env.ONLY?.split(',');
const surfaces=JSON.parse(readFileSync(S+'/'+(process.env.SURF||'surfaces.json'),'utf8')).filter(s=>!only||only.includes(s[0]));
const VP={phone:[390,844],tabP:[768,1024],tabL:[1024,768],desk:[1440,900]};
const log=m=>appendFileSync(S+'/capture.log',m+'\n');
const browser=await chromium.launch();
const measure=()=>{
  const W=innerWidth, out={docOverflow:document.documentElement.scrollWidth-W};
  const vis=e=>{const r=e.getBoundingClientRect(); const cs=getComputedStyle(e); return r.width>0&&r.height>0&&cs.visibility!=='hidden'&&cs.display!=='none'&&!e.closest('[inert],[aria-hidden="true"]');};
  const name=e=>(e.tagName.toLowerCase()+(e.id?'#'+e.id:'')+'.'+[...e.classList].slice(0,2).join('.')+' "'+(e.getAttribute('aria-label')||e.textContent||e.getAttribute('placeholder')||'').trim().replace(/\s+/g,' ').slice(0,40)+'"');
  const inScroller=e=>{for(let p=e.parentElement;p&&p!==document.body;p=p.parentElement){const o=getComputedStyle(p).overflowX; if(o==='auto'||o==='scroll'||o==='hidden'||o==='clip') return true;} return false;};
  out.overflowEls=[...document.querySelectorAll('body *')].filter(e=>vis(e)&&e.getBoundingClientRect().right>W+1&&!inScroller(e)).slice(0,8).map(name);
  out.clipped=[...document.querySelectorAll('body *')].filter(e=>{if(!vis(e)||!e.childNodes.length) return false; const cs=getComputedStyle(e); const hidden=['hidden','clip'].includes(cs.overflowX)||['hidden','clip'].includes(cs.overflowY); if(!hidden) return false; const hasText=[...e.childNodes].some(n=>n.nodeType===3&&n.textContent.trim()); return hasText&&(e.scrollWidth>e.clientWidth+1||e.scrollHeight>e.clientHeight+2);}).filter(e=>!e.matches('.sr-only,.visually-hidden,[class*="visually"]')&&e.clientWidth>1).slice(0,8).map(name);
  const sel='a[href],button,input:not([type=hidden]),select,textarea,summary,[role=button],[role=tab],[role=link],[role=checkbox],[role=radio],[role=switch],[role=menuitem],[role=option],[tabindex="0"]';
  out.small=[...document.querySelectorAll(sel)].filter(e=>{if(!vis(e)) return false; let t=e; if(e.matches('input[type=checkbox],input[type=radio]')){const l=e.closest('label'); if(l) t=l; else return false;} if(getComputedStyle(e).position==='absolute'&&e.matches('input')) {const l=e.closest('label'); if(l) t=l;} const r=t.getBoundingClientRect(); const cs=getComputedStyle(e); const inline=cs.display==='inline'&&e.closest('p,li,dd,td'); return !inline&&(r.width<43.5||r.height<43.5);}).slice(0,12).map(e=>{const r=e.getBoundingClientRect(); return name(e)+` ${Math.round(r.width)}x${Math.round(r.height)}`});
  out.h1=[...document.querySelectorAll('h1')].map(h=>h.textContent.trim().slice(0,60));
  out.title=document.title; out.height=document.documentElement.scrollHeight;
  return out;
};
const fontSnapshot=()=>[...document.querySelectorAll('body *')].filter(e=>[...e.childNodes].some(n=>n.nodeType===3&&n.textContent.trim())).map(e=>parseFloat(getComputedStyle(e).fontSize));
const results={};
for (const [id,role,path,surface,acts=[]] of surfaces){
  const themes= surface==='client'?['light','dark']:['dark'];
  const variants=[];
  for (const th of themes) for (const v of Object.keys(VP)) variants.push({k:`${v}-${th}`,vp:VP[v],th});
  variants.push({k:'text200-phone',vp:VP.phone,th:themes[0],t200:true},{k:'text200-desk',vp:VP.desk,th:themes[0],t200:true},{k:'reflow320',vp:[320,720],th:themes[0]});
  mkdirSync(`${S}/${OUT}/${id}`,{recursive:true});
  results[id]={role,path,variants:{}};
  for (const v of variants){
    const ctx=await browser.newContext({viewport:{width:v.vp[0],height:v.vp[1]}, colorScheme:v.th, reducedMotion:'reduce', ...(role!=='guest'?{storageState:`${S}/auth/${role}.json`}:{})});
    const page=await ctx.newPage(); const errs=[];
    page.on('console',m=>{if(m.type()==='error') errs.push(m.text().slice(0,160));}); page.on('pageerror',e=>errs.push('PAGEERROR '+e.message.slice(0,160)));
    let status=null;
    try{ const r=await page.goto(BASE+path,{waitUntil:'load',timeout:30000}); status=r?.status(); await page.waitForTimeout(1500);
      for (const a of acts){ if(a.fill) await page.getByLabel(a.fill[0],{exact:true}).fill(a.fill[1].replace('{v}',v.k)); if(a.click) await page.getByRole('button',{name:a.click,exact:true}).click(); if(a.wait) await page.getByText(a.wait).first().waitFor({timeout:15000}); }
      const rec={status,url:page.url().replace(BASE,'')};
      let before;
      if(v.t200){ before=await page.evaluate(fontSnapshot); await page.addStyleTag({content:'html{font-size:200% !important}'}); await page.waitForTimeout(400); const after=await page.evaluate(fontSnapshot); rec.pxLocked=before.length===after.length? after.filter((f,i)=>f<before[i]*1.5).length+'/'+after.length : 'n/a'; }
      Object.assign(rec, await page.evaluate(measure));
      if(!v.t200 && v.k!=='reflow320'){ await page.addScriptTag({content:AXE}); rec.axe=await page.evaluate(async()=>{const r=await axe.run(document,{runOnly:['wcag2a','wcag2aa','wcag21aa','wcag22aa']}); return r.violations.map(x=>`${x.id}(${x.impact})x${x.nodes.length}: ${x.nodes.slice(0,2).map(n=>n.target.join(' ')).join(' | ')}`)}); }
      if(v.k==='desk-'+themes[0]){ rec.tab=[]; await page.mouse.click(1,1).catch(()=>{}); await page.evaluate(()=>document.activeElement?.blur?.()); for(let i=0;i<30;i++){ await page.keyboard.press('Tab'); rec.tab.push(await page.evaluate(()=>{const e=document.activeElement; if(!e||e===document.body) return 'BODY'; const cs=getComputedStyle(e); const ring=(cs.outlineStyle!=='none'&&parseFloat(cs.outlineWidth)>0)||cs.boxShadow!=='none'; const r=e.getBoundingClientRect(); return (ring?'':'NORING ')+(r.bottom<0||r.top>innerHeight?'OFFSCREEN ':'')+e.tagName.toLowerCase()+' "'+(e.getAttribute('aria-label')||e.textContent||'').trim().replace(/\s+/g,' ').slice(0,30)+'"';})); }
        await page.evaluate(()=>window.scrollTo(0,0)); }
      rec.errors=errs;
      await page.screenshot({path:`${S}/${OUT}/${id}/${v.k}.png`, fullPage:true});
      results[id].variants[v.k]=rec;
    }catch(e){ results[id].variants[v.k]={status,error:e.message.slice(0,200),errors:errs}; }
    await ctx.close();
  }
  log(`${id} done`);
  writeFileSync(`${S}/metrics${process.env.TAG||''}.json`, JSON.stringify(results,null,1));
}
await browser.close(); log('ALL DONE');
