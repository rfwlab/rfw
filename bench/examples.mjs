// Run after building all examples and the native Notes server. Uses the existing
// benchmark Playwright dependency; writes only to the explicit results directory.
import {chromium, firefox, webkit} from 'playwright-core';
import assert from 'node:assert/strict';
import http from 'node:http';
import net from 'node:net';
import fs from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {spawn} from 'node:child_process';
import {once} from 'node:events';

const root = path.resolve(process.env.RFW_EXAMPLES_ROOT || fileURLToPath(new URL('../examples', import.meta.url)));
const output = process.env.RFW_EXAMPLE_RESULTS;
const api = process.env.RFW_NOTES_SERVER;
assert(output && api, 'RFW_EXAMPLE_RESULTS and RFW_NOTES_SERVER are required');
fs.mkdirSync(output, {recursive: true});
const results = [];
const servers = [];
let notes;

async function serve(name) {
    const dir = path.join(root, name, 'build/client');
    const server = http.createServer((req, res) => {
        const url = new URL(req.url, 'http://127.0.0.1');
        const name = ['/', '/about'].includes(url.pathname) ? 'index.html' : url.pathname.slice(1);
        const file = path.resolve(dir, name);
        if (!file.startsWith(dir + path.sep)) { res.writeHead(404).end(); return; }
        try {
            const data = fs.readFileSync(file);
            const mime = {'.html': 'text/html', '.js': 'text/javascript', '.wasm': 'application/wasm'};
            res.writeHead(200, {'Content-Type': mime[path.extname(file)] || 'application/octet-stream'});
            res.end(data);
        } catch { res.writeHead(404).end(); }
    });
    server.listen(0, '127.0.0.1');
    await once(server, 'listening');
    servers.push(server);
    return `http://127.0.0.1:${server.address().port}`;
}

try {
    const bases = {};
    for (const name of ['counter', 'dynamic-list', 'dashboard']) bases[name] = await serve(name);
    const reservation = net.createServer();
    reservation.listen(0, '127.0.0.1');
    await once(reservation, 'listening');
    const port = reservation.address().port;
    await new Promise(resolve => reservation.close(resolve));
    const fd = fs.openSync(path.join(output, 'notes-server.log'), 'w');
    notes = spawn(api, ['--port', String(port), '--assets', path.join(root, 'notes/build/client')], {stdio: ['ignore', fd, fd]});
    fs.closeSync(fd);
    bases.notes = `http://127.0.0.1:${port}`;
    let ready = false;
    for (let n = 0; n < 100; n++) {
        try { ready = (await fetch(bases.notes + '/api/notes')).ok; } catch {}
        if (ready) break;
        assert(notes.exitCode === null, 'Notes server exited');
        await new Promise(resolve => setTimeout(resolve, 100));
    }
    assert(ready, 'Notes server did not start');
    const deniedOrigin = await fetch(bases.notes + '/api/notes', {headers: {Origin: 'https://untrusted.example'}});
    assert.equal(deniedOrigin.status, 403);
    for (const [engine, type] of Object.entries({chromium, firefox, webkit})) {
        const browser = await type.launch(engine === 'chromium' ? {executablePath: process.env.CHROME_PATH || '/usr/bin/google-chrome', args: ['--no-sandbox']} : {});
        try {
            for (const [size, viewport] of Object.entries({desktop: {width: 1280, height: 900}, mobile: {width: 390, height: 844}})) {
                for (const name of ['counter', 'dynamic-list', 'dashboard', 'notes']) {
                    const context = await browser.newContext({viewport});
                    const page = await context.newPage();
                    page.setDefaultTimeout(15000);
                    const result = {engine, size, example: name, status: 'RUNNING', errors: []};
                    results.push(result);
                    page.on('pageerror', e => result.errors.push(e.message));
                    try {
                        await page.goto(bases[name]);
                        if (name === 'counter') {
                            await page.getByRole('button', {name: 'Increment', exact: true}).click();
                            await page.getByText('Count: 1', {exact: true}).waitFor();
                        } else if (name === 'dynamic-list') {
                            const text = `Literal <b>${engine} ${size}</b>`;
                            await page.locator('#item-name').fill(text);
                            await page.getByRole('button', {name: 'Add', exact: true}).click();
                            const row = page.locator('#items-rows tr').filter({hasText: text});
                            await row.waitFor();
                            assert.equal(await row.locator('b').count(), 0);
                            await row.getByRole('button', {name: 'Remove', exact: true}).click();
                            assert.equal(await row.count(), 0);
                        } else if (name === 'dashboard') {
                            await page.getByText('Status: live', {exact: true}).waitFor();
                            await page.getByRole('button', {name: 'Pause / resume', exact: true}).click();
                            await page.getByText('Status: paused', {exact: true}).waitFor();
                            await page.getByRole('link', {name: 'About this demo', exact: true}).click();
                            await page.getByRole('heading', {name: 'About', exact: true}).waitFor();
                            await page.getByRole('link', {name: 'Back to the dashboard', exact: true}).click();
                            await page.getByText('Status: live', {exact: true}).waitFor();
                            await page.waitForFunction(() => document.querySelector('ul li')?.textContent.includes('cpu sample'));
                        } else {
                            await page.getByText('Notes loaded.', {exact: true}).waitFor();
                            const text = `Literal <b>${engine} ${size}</b>`;
                            await page.getByLabel('Note text', {exact: true}).fill(text);
                            await page.getByRole('button', {name: 'Save note', exact: true}).click();
                            const row = page.locator('#notes-list li').filter({hasText: text});
                            await row.waitFor();
                            const id = await row.getAttribute('data-note');
                            assert.equal(await row.locator('b').count(), 0);
                            await page.reload();
                            await row.waitFor();
                            const other = await context.newPage();
                            other.setDefaultTimeout(15000);
                            await other.goto(bases.notes);
                            const selector = `#notes-list li[data-note="${id}"]`;
                            await other.locator(selector).getByRole('button', {name: 'Edit', exact: true}).click();
                            await page.locator(selector).getByRole('button', {name: 'Edit', exact: true}).click();
                            await page.getByLabel('Note text', {exact: true}).fill(`Updated ${engine} ${size}`);
                            await page.getByRole('button', {name: 'Save note', exact: true}).click();
                            await page.locator(selector).getByText(`Updated ${engine} ${size}`, {exact: true}).waitFor();
                            const draft = `Kept draft ${engine} ${size}`;
                            await other.getByLabel('Note text', {exact: true}).fill(draft);
                            await other.getByRole('button', {name: 'Save note', exact: true}).click();
                            await other.getByText('The note changed.', {exact: false}).waitFor();
                            assert.equal(await other.getByLabel('Note text').inputValue(), draft);
                            await other.close();
                            await page.locator(selector).getByRole('button', {name: 'Delete', exact: true}).click();
                            await page.locator(selector).waitFor({state: 'detached'});
                            await page.getByLabel('Note text').fill(draft);
                            await page.route('**/api/notes', route => route.request().method() === 'POST' ? route.abort() : route.continue(), {times: 1});
                            await page.getByRole('button', {name: 'Save note', exact: true}).click();
                            await page.getByText('Connection failed.', {exact: false}).waitFor();
                            assert.equal(await page.getByLabel('Note text').inputValue(), draft);
                            let release;
                            const delayed = new Promise(resolve => { release = resolve; });
                            let observed;
                            const started = new Promise(resolve => { observed = resolve; });
                            await page.route('**/api/notes', async route => { const response = await route.fetch(); observed(); await delayed; await route.fulfill({response}); }, {times: 1});
                            await page.getByRole('button', {name: 'Refresh list', exact: true}).click();
                            await started;
                            await page.getByRole('link', {name: 'About this example', exact: true}).click();
                            await page.getByRole('heading', {name: 'About Notes', exact: true}).waitFor();
                            release();
                            await page.waitForTimeout(100);
                            assert.equal(await page.locator('#notes-message').count(), 0);
                            await page.getByRole('link', {name: 'Back to notes', exact: true}).click();
                            await page.getByText('Notes loaded.', {exact: true}).waitFor();
                            assert.equal(await page.locator(selector).count(), 0);
                            result.steps = ['real UI create/reload/edit/delete against Go API', 'HTML-looking text stays literal', 'second-tab stale edit keeps draft', 'transport failure keeps draft without retry', 'late real GET cannot touch an unmounted page'];
                        }
                        await page.screenshot({path: path.join(output, `${name}-${engine}-${size}.png`), fullPage: true});
                        assert.deepEqual(result.errors, []);
                        result.status = 'PASS';
                    } catch (error) {
                        result.status = 'FAIL'; result.error = error.message;
                        await page.screenshot({path: path.join(output, `${name}-${engine}-${size}-failure.png`), fullPage: true}).catch(() => {});
                        throw error;
                    } finally { await context.close(); }
                }
            }
        } finally { await browser.close(); }
    }
} finally {
    for (const server of servers) await new Promise(resolve => server.close(resolve));
    if (notes?.exitCode === null) { notes.kill('SIGTERM'); await once(notes, 'exit'); }
    fs.writeFileSync(path.join(output, 'results.json'), JSON.stringify(results, null, 2) + '\n');
}
console.log(JSON.stringify(results));
