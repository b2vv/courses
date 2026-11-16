import { createRequire } from 'node:module';
import { writeFileSync, existsSync } from 'node:fs';
const require=createRequire('/Users/strelia/projects/courses/.claude/worktrees/dg-archive-test-c6c9f4/frontend/package.json');
const { chromium } = require('@playwright/test');
const S=process.env.S; const [,,out,...ids]=process.argv;
const b=await chromium.launch(); const p=await b.newPage({viewport:{width:1900,height:800}});
const f=(id,k)=>{for(const t of [k+'-light',k+'-dark']){const x=`${S}/shots/${id}/${t}.png`; if(existsSync(x)) return x;}};
const html='<body style="margin:6px;font:13px sans-serif;background:#ccc;display:flex;gap:14px;align-items:flex-start">'+ids.map(id=>`<div><b>${id}</b><div style="display:flex;gap:4px;align-items:flex-start"><div style="width:156px;max-height:1150px;overflow:hidden"><img src="file://${f(id,'phone')}" style="width:156px"></div><div style="width:300px;max-height:1150px;overflow:hidden"><img src="file://${f(id,'desk')}" style="width:300px"></div></div></div>`).join('')+'</body>';
writeFileSync(out+'.html',html); await p.goto('file://'+out+'.html'); await p.waitForTimeout(300); await p.screenshot({path:out,fullPage:true}); await b.close();
