// Benchmark harness: rfw vs Svelte vs Solid vs htmx on TodoMVC + live update.
//
// Usage:  cd bench && npm install && node measure.mjs
//
// Prereqs (see bench/README notes in docs/articles/benchmarks.md):
//   - bench/todomvc/rfw:    GOOS=js GOARCH=wasm go build (app.wasm, app.opt.wasm)
//   - bench/todomvc/svelte: npm install && npm run build (dist/)
//   - bench/todomvc/solid:  npm install && npm run build (dist/)
//   - bench/todomvc/htmx:   go build -o htmx-server server.go, htmx.min.js copied locally
//
// Measures per framework (3 runs, medians):
//   load_ms        navigation start -> app interactive (#new-todo + #todo-list in DOM)
//   add100_ms      adding 100 todos through the UI (real input + click events)
//   live_*         30 counter updates at 10 Hz (3 s expected): wall time + rendered updates
//   heap_bytes     JSHeapUsedSize (CDP Performance.getMetrics) after load
// Plus bundle sizes (raw + gzip -9) computed from the build artifacts.

import { chromium } from 'playwright-core';
import http from 'node:http';
import { spawn } from 'node:child_process';
import { readFileSync, writeFileSync, statSync } from 'node:fs';
import { gzipSync } from 'node:zlib';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const CHROME = process.env.CHROME_PATH || '/usr/bin/google-chrome';
const RUNS = 3;

const MIME = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.wasm': 'application/wasm',
  '.json': 'application/json',
};

function serveStatic(root, port, aliases = {}) {
  const server = http.createServer((req, res) => {
    let url = req.url.split('?')[0];
    if (url === '/') url = '/index.html';
    if (aliases[url]) url = aliases[url];
    const file = path.join(root, url);
    try {
      const data = readFileSync(file);
      res.writeHead(200, { 'Content-Type': MIME[path.extname(file)] || 'application/octet-stream' });
      res.end(data);
    } catch {
      res.writeHead(404);
      res.end('not found');
    }
  });
  return new Promise((resolve) => server.listen(port, '127.0.0.1', () => resolve(server)));
}

function waitForHttp(url, timeoutMs = 10000) {
  const start = Date.now();
  return new Promise((resolve, reject) => {
    const attempt = () => {
      http.get(url, (res) => { res.resume(); resolve(); }).on('error', () => {
        if (Date.now() - start > timeoutMs) reject(new Error(`timeout waiting for ${url}`));
        else setTimeout(attempt, 100);
      });
    };
    attempt();
  });
}

function gz(buf) {
  return gzipSync(buf, { level: 9 }).length;
}

function fileSizes(files) {
  let raw = 0, gzip = 0;
  const detail = {};
  for (const [label, file] of Object.entries(files)) {
    const buf = readFileSync(file);
    detail[label] = { raw: buf.length, gzip: gz(buf) };
    raw += buf.length;
    gzip += detail[label].gzip;
  }
  return { raw, gzip, detail };
}

function median(arr) {
  const s = [...arr].sort((a, b) => a - b);
  return s[Math.floor(s.length / 2)];
}

const READY_SNIPPET = `
  (() => {
    const check = () => {
      if (document.querySelector('#new-todo') && document.querySelector('#todo-list')) {
        window.__appReady = performance.now();
      } else {
        requestAnimationFrame(check);
      }
    };
    check();
  })();
`;

async function benchOne(browser, url) {
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.addInitScript(READY_SNIPPET);

  const cdp = await context.newCDPSession(page);
  await cdp.send('Performance.enable');

  // (a) cold load to interactive
  await page.goto(url, { waitUntil: 'commit' });
  await page.waitForFunction(() => window.__appReady !== undefined, null, { timeout: 60000 });
  const loadMs = await page.evaluate(() => window.__appReady);

  // (d) heap after load (let things settle briefly)
  await page.waitForTimeout(500);
  const metrics = await cdp.send('Performance.getMetrics');
  const heap = metrics.metrics.find((m) => m.name === 'JSHeapUsedSize')?.value ?? 0;
  // JSHeapUsedSize does not account for WebAssembly linear memory; for the
  // rfw (Go/wasm) app read it off the Go runtime instance directly.
  const wasmMem = await page.evaluate(() => {
    try { return globalThis.go?._inst?.exports?.mem?.buffer?.byteLength ?? 0; } catch { return 0; }
  });

  // (b) add 100 todos through the UI
  const add100Ms = await page.evaluate(async () => {
    const list = () => document.querySelectorAll('#todo-list li').length;
    const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set;
    const start = performance.now();
    const deadline = start + 120000;
    for (let i = 0; i < 100; i++) {
      // elements may be re-rendered between iterations, re-query each time
      const input = document.querySelector('#new-todo');
      const btn = document.querySelector('#add-todo');
      const before = list();
      setter.call(input, 'todo item ' + i);
      input.dispatchEvent(new Event('input', { bubbles: true }));
      btn.click();
      while (list() !== before + 1) {
        if (performance.now() > deadline) throw new Error('add100 timeout at item ' + i);
        await new Promise((r) => setTimeout(r, 0));
      }
    }
    return performance.now() - start;
  });

  // (c) live update: 30 ticks at 10 Hz, expected 3000 ms
  const live = await page.evaluate(async () => {
    const counterText = () => document.querySelector('#live-counter')?.textContent.trim();
    let updates = 0;
    let last = counterText();
    const mo = new MutationObserver(() => {
      const v = counterText();
      if (v !== last) { last = v; updates++; }
    });
    mo.observe(document.body, { childList: true, characterData: true, subtree: true });
    const start = performance.now();
    document.querySelector('#start-live').click();
    const deadline = start + 15000;
    while (counterText() !== '30' && performance.now() < deadline) {
      await new Promise((r) => setTimeout(r, 5));
    }
    const elapsed = performance.now() - start;
    mo.disconnect();
    return { elapsed, updates, final: counterText() };
  });

  await context.close();
  return { loadMs, heap, wasmMem, add100Ms, live };
}

async function benchFramework(browser, name, url) {
  const runs = [];
  for (let i = 0; i < RUNS; i++) {
    process.stderr.write(`  ${name} run ${i + 1}/${RUNS}...\n`);
    runs.push(await benchOne(browser, url));
  }
  return {
    runs,
    median: {
      load_ms: median(runs.map((r) => r.loadMs)),
      add100_ms: median(runs.map((r) => r.add100Ms)),
      live_elapsed_ms: median(runs.map((r) => r.live.elapsed)),
      live_updates: median(runs.map((r) => r.live.updates)),
      heap_bytes: median(runs.map((r) => r.heap)),
      wasm_memory_bytes: median(runs.map((r) => r.wasmMem)),
    },
  };
}

function findDistAsset(dir, ext) {
  const assets = path.join(dir, 'dist', 'assets');
  const { readdirSync } = require('node:fs');
  return readdirSync(assets).filter((f) => f.endsWith(ext)).map((f) => path.join(assets, f));
}

async function main() {
  const todomvc = path.join(__dirname, 'todomvc');
  const { readdirSync } = await import('node:fs');

  // ---- bundle sizes ----
  const rfwDir = path.join(todomvc, 'rfw');
  const sizes = {};

  sizes.rfw = fileSizes({
    'app.wasm (optimized -s -w)': path.join(rfwDir, 'app.opt.wasm'),
    'wasm_exec.js': path.join(rfwDir, 'wasm_exec.js'),
    'wasm_loader.js': path.join(rfwDir, 'wasm_loader.js'),
    'index.html': path.join(rfwDir, 'index.html'),
  });
  sizes.rfw.unoptimized_wasm = {
    raw: statSync(path.join(rfwDir, 'app.wasm')).size,
    gzip: gz(readFileSync(path.join(rfwDir, 'app.wasm'))),
  };

  for (const fw of ['svelte', 'solid']) {
    const assetsDir = path.join(todomvc, fw, 'dist', 'assets');
    const files = { 'index.html': path.join(todomvc, fw, 'dist', 'index.html') };
    for (const f of readdirSync(assetsDir)) {
      if (f.endsWith('.js') || f.endsWith('.css')) files[f] = path.join(assetsDir, f);
    }
    sizes[fw] = fileSizes(files);
  }

  // htmx: library + rendered initial page
  const htmxDir = path.join(todomvc, 'htmx');
  const htmxServer = spawn(path.join(htmxDir, 'htmx-server'), ['-addr', '127.0.0.1:4614'], {
    cwd: htmxDir, stdio: 'ignore',
  });
  await waitForHttp('http://127.0.0.1:4614/');
  const pageHtml = await new Promise((resolve, reject) => {
    http.get('http://127.0.0.1:4614/', (res) => {
      const chunks = [];
      res.on('data', (c) => chunks.push(c));
      res.on('end', () => resolve(Buffer.concat(chunks)));
    }).on('error', reject);
  });
  const htmxLib = readFileSync(path.join(htmxDir, 'htmx.min.js'));
  sizes.htmx = {
    raw: htmxLib.length + pageHtml.length,
    gzip: gz(htmxLib) + gz(pageHtml),
    detail: {
      'htmx.min.js': { raw: htmxLib.length, gzip: gz(htmxLib) },
      'initial page (server-rendered)': { raw: pageHtml.length, gzip: gz(pageHtml) },
    },
  };

  // ---- runtime benchmarks ----
  const servers = [];
  servers.push(await serveStatic(rfwDir, 4611, { '/app.wasm': '/app.opt.wasm' }));
  servers.push(await serveStatic(path.join(todomvc, 'svelte', 'dist'), 4612));
  servers.push(await serveStatic(path.join(todomvc, 'solid', 'dist'), 4613));

  const browser = await chromium.launch({ executablePath: CHROME, headless: true });

  const results = {};
  const targets = {
    rfw: 'http://127.0.0.1:4611/',
    svelte: 'http://127.0.0.1:4612/',
    solid: 'http://127.0.0.1:4613/',
    htmx: 'http://127.0.0.1:4614/',
  };
  for (const [name, url] of Object.entries(targets)) {
    console.error(`benchmarking ${name}`);
    try {
      results[name] = await benchFramework(browser, name, url);
    } catch (err) {
      console.error(`  ${name} FAILED: ${err.message}`);
      results[name] = { error: String(err.message) };
    }
  }

  await browser.close();
  servers.forEach((s) => s.close());
  htmxServer.kill();

  // ---- report ----
  const report = {
    date: new Date().toISOString(),
    runs: RUNS,
    live_scenario: { ticks: 30, interval_ms: 100, expected_ms: 3000 },
    sizes,
    results,
  };
  writeFileSync(path.join(__dirname, 'results.json'), JSON.stringify(report, null, 2));
  console.error(`\nwrote ${path.join(__dirname, 'results.json')}\n`);

  const kb = (b) => (b / 1024).toFixed(1);
  const rows = [['framework', 'size raw KB', 'size gzip KB', 'load ms', 'add100 ms', 'live ms (3000 exp.)', 'live updates (30 exp.)', 'heap MB', 'wasm mem MB']];
  for (const name of Object.keys(targets)) {
    const s = sizes[name];
    const m = results[name]?.median;
    rows.push([
      name,
      kb(s.raw), kb(s.gzip),
      m ? m.load_ms.toFixed(1) : 'FAIL',
      m ? m.add100_ms.toFixed(1) : 'FAIL',
      m ? m.live_elapsed_ms.toFixed(1) : 'FAIL',
      m ? String(m.live_updates) : 'FAIL',
      m ? (m.heap_bytes / 1024 / 1024).toFixed(1) : 'FAIL',
      m ? (m.wasm_memory_bytes / 1024 / 1024).toFixed(1) : 'FAIL',
    ]);
  }
  const widths = rows[0].map((_, i) => Math.max(...rows.map((r) => String(r[i]).length)));
  for (const r of rows) {
    console.log(r.map((c, i) => String(c).padEnd(widths[i])).join('  '));
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
