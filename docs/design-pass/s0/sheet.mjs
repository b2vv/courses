import { createRequire } from 'node:module';
import { readdirSync, existsSync, mkdirSync, writeFileSync } from 'node:fs';
const require=createRequire('<repo>/frontend/package.json');
const { chromium } = require('@playwright/test');
const S=process.env.S; mkdirSync(S+'/'+(process.env.SHEETS||'sheets'),{recursive:true});
const b=await chromium.launch(); const p=await b.newPage({viewport:{width:1860,height:800}});
const rows=[['phone-light','tabP-light','tabL-light','desk-light'],['phone-dark','tabP-dark','tabL-dark','desk-dark'],['text200-phone','reflow320','text200-desk']];
const W={phone:195,tabP:384,tabL:512,desk:720,text200:195,reflow320:160};
const MAXH=+process.env.MAXH||1150;
for (const id of readdirSync(S+'/'+(process.env.OUT||'shots'))){
  const cell=k=>{const f=`${S}/${process.env.OUT||'shots'}/${id}/${k}.png`; if(!existsSync(f)) return ''; const w=k==='text200-desk'?720:W[k.split('-')[0]]??W[k]; return `<figure><figcaption>${k}</figcaption><div style="width:${w}px;max-height:${MAXH}px;overflow:hidden;border:1px solid #888"><img src="file://${f}" style="width:${w}px;display:block"></div></figure>`;};
  const html=`<body style="margin:8px;font:14px sans-serif;background:#ddd"><h3 style="margin:0">${id}</h3>`+rows.map(r=>`<div style="display:flex;gap:8px;align-items:flex-start">${r.map(cell).join('')}</div>`).join('')+'</body>';
  writeFileSync(`${S}/${process.env.SHEETS||'sheets'}/${id}.html`,html); await p.goto(`file://${S}/${process.env.SHEETS||'sheets'}/${id}.html`); await p.waitForTimeout(300);
  await p.screenshot({path:`${S}/${process.env.SHEETS||'sheets'}/${id}.png`,fullPage:true});
}
await b.close();
