// Real browser fetch/body cancellation against an owned local HTTP server.
import {chromium, firefox, webkit} from 'playwright-core';
import fs from 'node:fs';
import http from 'node:http';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {execFileSync} from 'node:child_process';
import assert from 'node:assert/strict';
const root=path.resolve(process.env.RFW_ROOT||fileURLToPath(new URL('..',import.meta.url)));
const out=process.env.RFW_HTTP_RESULTS;
assert(out,'RFW_HTTP_RESULTS must name an owned output directory');fs.mkdirSync(out,{recursive:true});
const go=process.env.GO_BINARY||'go';
execFileSync(go,['build','-o',path.join(out,'app.wasm'),'./bench/http-context'],{cwd:root,env:{...process.env,GOOS:'js',GOARCH:'wasm'},stdio:'inherit'});
const goroot=execFileSync(go,['env','GOROOT'],{encoding:'utf8'}).trim();
fs.copyFileSync(path.join(goroot,'lib/wasm/wasm_exec.js'),path.join(out,'wasm_exec.js'));
const records=new Map(),results=[];
const server=http.createServer((req,res)=>{
 const url=new URL(req.url,'http://127.0.0.1');const id=url.searchParams.get('id');
 if(id){const record=records.get(id)||{requests:0,closed:false,committed:0};record.requests++;records.set(id,record);res.on('close',()=>record.closed=true)}
 if(url.pathname==='/pending'||url.pathname==='/commit'){
  if(url.pathname==='/commit')records.get(id).committed++;
  return;
 }
 if(url.pathname==='/body') {res.writeHead(200,{'Content-Type':'application/octet-stream'});res.flushHeaders();res.write(Buffer.from([0,255]));return;}
 if(url.pathname==='/missing'){res.writeHead(404).end('missing');return;}
 if(url.pathname==='/binary'){res.writeHead(201).end(Buffer.from([0,255,128,1]));return;}
 if(url.pathname==='/ok'){res.end('ok');return;}
 if(url.pathname==='/wasm_exec.js'||url.pathname==='/app.wasm'){
  res.setHeader('Content-Type',url.pathname.endsWith('.wasm')?'application/wasm':'text/javascript');res.end(fs.readFileSync(path.join(out,url.pathname.slice(1))));return;
 }
 res.setHeader('Content-Type','text/html');res.end(`<!doctype html><html lang="en"><meta charset="utf-8"><title>HTTP context probe</title><body><h1>HTTP context probe</h1><p>Request results are checked by the browser harness.</p><script src="/wasm_exec.js"></script><script>const go=new Go();WebAssembly.instantiateStreaming(fetch('/app.wasm'),go.importObject).then(r=>go.run(r.instance));</script></body></html>`);
});
await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));const base='http://127.0.0.1:'+server.address().port;
async function until(condition){const end=Date.now()+5000;while(!condition()){assert(Date.now()<end,'server condition timed out');await new Promise(r=>setTimeout(r,10))}}
try{
 for(const [engine,type] of Object.entries({chromium,firefox,webkit})){
  const browser=await type.launch(engine==='chromium'?{executablePath:process.env.CHROME_PATH||'/usr/bin/google-chrome',args:['--no-sandbox']}:{});
  try{
   const page=await browser.newPage();page.setDefaultTimeout(10000);
   await page.addInitScript(()=>{window.requestBodyReads=[];const read=Response.prototype.arrayBuffer;Response.prototype.arrayBuffer=function(...args){window.requestBodyReads.push(this.url);return read.apply(this,args)}});
   const result={engine,steps:[],errors:[],consoleErrors:[]};results.push(result);
   page.on('pageerror',e=>result.errors.push(e.message));page.on('console',m=>{if(m.type()==='error'&&!m.text().includes('404'))result.consoleErrors.push(m.text())});
   try{
    await page.goto(base);await page.waitForFunction(()=>window.requestProbeReady===true);
    async function start(mode,id,endpoint,timeout=0){await page.evaluate(({mode,id,endpoint,timeout})=>window.requestProbe(mode,id,endpoint+'?id='+id,timeout),{mode,id,endpoint,timeout})}
    async function completed(id){await page.waitForFunction(id=>window.requestResults.some(r=>r.id===id),id);return page.evaluate(id=>window.requestResults.filter(r=>r.id===id),id)}
    let id=engine+'-pre';await start('precancelled',id,'/pending');assert.deepEqual(await completed(id),[{id,status:0,body:''}]);assert(!records.has(id));result.steps.push('Precancelled context does not reach server');
    id=engine+'-deadline';await start('text',id,'/pending',500);await until(()=>records.has(id));assert.deepEqual(await completed(id),[{id,status:0,body:''}]);await until(()=>records.get(id).closed);assert.equal(records.get(id).requests,1);result.steps.push('Real pending fetch deadline completes once and closes transport');
    id=engine+'-body';await start('bytes',id,'/body');await page.waitForFunction(id=>window.requestBodyReads.some(url=>url.endsWith('/body?id='+id)),id);await page.evaluate(id=>window.cancelRequestProbe(id),id);assert.deepEqual(await completed(id),[{id,status:0,body:''}]);await until(()=>records.get(id).closed);result.steps.push('Manual cancellation closes a real streaming response body');
    id=engine+'-commit';await start('text',id,'/commit',500);await until(()=>records.has(id));assert.deepEqual(await completed(id),[{id,status:0,body:''}]);await until(()=>records.get(id).closed);assert.equal(records.get(id).committed,1);assert.equal(records.get(id).requests,1);result.steps.push('Abort leaves an already admitted server effect intact and does not replay');
    id=engine+'-legacy';await start('legacy',id,'/missing');assert.deepEqual(await completed(id),[{id,status:404,body:'missing'}]);
    id=engine+'-binary';await start('bytes',id,'/binary');assert.deepEqual(await completed(id),[{id,status:201,body:'00ff8001'}]);result.steps.push('Legacy HTTP status and exact binary bytes survive');
    for(let n=0;n<10;n++){id=engine+'-repeat'+n;await start('text',id,'/pending');await until(()=>records.has(id));await page.evaluate(id=>window.cancelRequestProbe(id),id);assert.equal((await completed(id)).length,1);await until(()=>records.get(id).closed)}
    id=engine+'-last';await start('text',id,'/ok');assert.deepEqual(await completed(id),[{id,status:200,body:'ok'}]);
    await page.evaluate(()=>document.querySelector('p').textContent=JSON.stringify(window.requestResults));
    await page.screenshot({path:path.join(out,engine+'.png'),fullPage:true});assert.deepEqual(result.errors,[]);assert.deepEqual(result.consoleErrors,[]);result.steps.push('Repeated aborts leave no duplicate or released-function callback errors');
    result.status='PASS';
   }catch(error){result.status='FAIL';result.failure=error.stack;process.exitCode=1}
   console.log(JSON.stringify(result));
  }finally{await browser.close()}
 }
}finally{server.closeAllConnections();await new Promise(r=>server.close(r));fs.writeFileSync(path.join(out,'results.json'),JSON.stringify({results,requests:Object.fromEntries(records)},null,2)+'\n')}
