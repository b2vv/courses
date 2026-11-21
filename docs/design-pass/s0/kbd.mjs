import { createRequire } from 'node:module';
const require=createRequire('<repo>/frontend/package.json');
const { chromium } = require('@playwright/test');
const B='http://localhost:3100';
const cases=[['/login','Забули пароль?','/password/forgot'],['/login','Зареєструватися','/register'],['/register','Увійти','/login',1],['/password/forgot','Увійти','/login',1],['/password/reset','Запросити нове посилання','/password/forgot']];
const b=await chromium.launch(); const out=[];
for (const [path,name,expect,skip=0] of cases){ let nth=skip;
  const p=await b.newPage({viewport:{width:1440,height:900}}); await p.goto(B+path); await p.waitForTimeout(1200);
  let reached=false;
  for(let i=0;i<20;i++){ await p.keyboard.press('Tab'); const n=await p.evaluate(()=>document.activeElement?.textContent?.trim()); if(n===name){ if(nth===0){reached=true;break;} nth--; } }
  if(reached){ await p.keyboard.press('Enter'); await p.waitForURL('**'+expect,{timeout:8000}).catch(()=>{}); }
  out.push(`${path} Tab→"${name}" ${reached?'reached':'NOT reached'} → Enter → ${new URL(p.url()).pathname} ${new URL(p.url()).pathname===expect?'PASS':'FAIL'}`);
  await p.close();
}
await b.close(); console.log(out.join('\n'));
