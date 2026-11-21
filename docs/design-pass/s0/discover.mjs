import { createRequire } from 'node:module';
import { readFileSync, writeFileSync } from 'node:fs';
const FE=new URL('../../../frontend', import.meta.url).pathname; // repo-relative, resolved from this file
const require=createRequire(FE+'/package.json');
const { request } = require('@playwright/test');
const src=readFileSync(FE+'/e2e/auth.ts','utf8');
const roles={};
for (const r of ['student','instructor','admin']) {
  const block=src.match(new RegExp(r+':\\s*\\{([^}]*)\\}'))[1];
  roles[r]={email:block.match(/email:\s*'([^']+)'/)[1], password:(block.match(/password:\s*'([^']+)'/)||[])[1]};
}
const pw=src.match(/password\s*[:=]\s*'([^']+)'/); // fallback shared password
const BASE='http://localhost:8080';
const out={};
for (const [r,{email,password}] of Object.entries(roles)) {
  const ctx=await request.newContext({baseURL:BASE, extraHTTPHeaders:{Origin:BASE}});
  const res=await ctx.post('/api/v1/auth/login',{data:{email,password:password??pw?.[1]}});
  out[r]={login:res.status()};
  await ctx.storageState({path:`${process.argv[2]}/auth/${r}.json`});
  const get=async p=>{const x=await ctx.get(p); return x.ok()? x.json(): {status:x.status()}};
  if(r==='admin'){ out.adminCourses=(await get('/api/v1/admin/courses')); }
  if(r==='student'){ out.memberships=await get('/api/v1/learn/memberships'); }
  if(r==='instructor'){ out.instructorCourses=await get('/api/v1/admin/courses'); }
  await ctx.dispose();
}
writeFileSync(process.argv[2]+'/discover.json', JSON.stringify(out,null,1));
console.log(JSON.stringify(Object.fromEntries(Object.entries(out).map(([k,v])=>[k, v.login ?? (v.items? v.items.length : v)]))));
