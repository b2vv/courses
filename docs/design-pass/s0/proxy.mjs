// Local verification harness only: /api and /media go to the live stack (with its own Origin,
// so the same-origin guard sees the origin it expects); everything else goes to this
// worktree's SSR server. Not part of the product.
import http from 'node:http';
const API='http://localhost:8080', SSR='http://localhost:3101';
http.createServer((req,res)=>{
  const toApi=req.url.startsWith('/api/')||req.url.startsWith('/media/');
  const target=new URL(req.url, toApi?API:SSR);
  const headers={...req.headers, host: target.host};
  if(toApi && headers.origin) headers.origin=API;
  if(toApi && headers.referer) headers.referer=API+'/';
  const up=http.request(target,{method:req.method,headers},(r)=>{res.writeHead(r.statusCode,r.headers); r.pipe(res);});
  up.on('error',e=>{res.writeHead(502); res.end(String(e));});
  req.pipe(up);
}).listen(3100,()=>console.log('proxy :3100'));
