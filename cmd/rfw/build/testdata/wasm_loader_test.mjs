import assert from "node:assert/strict";
import fs from "node:fs/promises";
import vm from "node:vm";
import zlib from "node:zlib";

const loaderPath = process.argv[2];
if (!loaderPath) throw new Error("wasm loader path is required");

function element() {
    return {
        style: {},
        append() {},
        appendChild() {},
        addEventListener() {},
        remove() {},
        set id(value) { this._id = value; },
        get id() { return this._id; },
        textContent: "",
        type: "",
    };
}

globalThis.window = globalThis;
globalThis.document = {
    body: element(),
    createElement: element,
    getElementById() { return null; },
};
globalThis.location = { pathname: "/", reload() {} };
const storage = new Map();
globalThis.sessionStorage = {
    getItem(key) { return storage.get(key) ?? null; },
    setItem(key, value) { storage.set(key, value); },
    removeItem(key) { storage.delete(key); },
};

let instantiated;
globalThis.WebAssembly = {
    async instantiateStreaming(response) {
        instantiated = Buffer.from(await response.arrayBuffer());
        return { instance: {} };
    },
    async instantiate(bytes) {
        instantiated = Buffer.from(bytes);
        return { instance: {} };
    },
};

vm.runInThisContext(await fs.readFile(loaderPath, "utf8"), {
    filename: loaderPath,
});
assert.equal(typeof WasmLoader.load, "function");

const wasm = Buffer.from("repeatable wasm loader fixture");

async function load(config, fetcher) {
    window.RFW_WASM_ENCODINGS = config.encodings;
    window.RFW_WASM_NEGOTIATED = config.negotiated;
    window.RFW_BUILD_MODE = config.production ? "production" : "development";
    window.RFW_WASM_DELIVERY = config.delivery ?? "network";
    const requests = [];
    globalThis.fetch = async (url) => {
        requests.push(url);
        return fetcher(url);
    };
    instantiated = undefined;
    await WasmLoader.load("/app.wasm?v=test", {
        go: { importObject: {}, run: () => new Promise(() => {}) },
        skipLoader: true,
        reloadOnExit: false,
        recoveryWindow: 1,
    });
    assert.deepEqual(instantiated, wasm);
    return requests;
}

assert.deepEqual(
    await load(
        { encodings: ["br", "gzip"], negotiated: true, production: true },
        (url) => {
            assert.equal(url, "/app.wasm?v=test");
            return new Response(wasm, {
                status: 200,
                headers: {
                    "content-encoding": "br",
                    "content-type": "application/wasm",
                },
            });
        },
    ),
    ["/app.wasm?v=test"],
);

const gzip = zlib.gzipSync(wasm);
assert.deepEqual(
    await load(
        { encodings: ["br", "gzip"], negotiated: false, production: true },
        (url) => {
            if (url.endsWith(".br?v=test")) return new Response(null, { status: 404 });
            assert.ok(url.endsWith(".gz?v=test"));
            return new Response(gzip, {
                status: 200,
                headers: { "content-type": "application/octet-stream" },
            });
        },
    ),
    ["/app.wasm.br?v=test", "/app.wasm.gz?v=test"],
);

window.RFW_WASM_ENCODINGS = ["br", "gzip"];
window.RFW_WASM_NEGOTIATED = false;
window.RFW_BUILD_MODE = "production";
const failedRequests = [];
globalThis.fetch = async (url) => {
    failedRequests.push(url);
    return new Response(null, { status: 404 });
};
await assert.rejects(
    WasmLoader.load("/app.wasm?v=test", {
        go: { importObject: {}, run() {} },
        skipLoader: true,
    }),
    /Tried br, gzip/,
);
assert.deepEqual(failedRequests, [
    "/app.wasm.br?v=test",
    "/app.wasm.gz?v=test",
]);

assert.deepEqual(
    await load(
        { encodings: [], negotiated: false, production: false },
        (url) => {
            assert.equal(url, "/app.wasm?v=test");
            return new Response(wasm, {
                status: 200,
                headers: { "content-type": "application/wasm" },
            });
        },
    ),
    ["/app.wasm?v=test"],
);

// An embedded production build loads the raw bundle and nothing else. The
// encodings and the negotiation flag are deliberately set to what a network
// build would emit: delivery decides, not the leftovers.
assert.deepEqual(
    await load(
        {
            encodings: ["br", "gzip"],
            negotiated: true,
            production: true,
            delivery: "embedded",
        },
        (url) => {
            assert.equal(url, "/app.wasm?v=test");
            return new Response(wasm, {
                status: 200,
                headers: { "content-type": "application/wasm" },
            });
        },
    ),
    ["/app.wasm?v=test"],
);

// A packaged bundle that is missing fails visibly, and the error names the
// packaged path rather than sending the reader to look for server headers.
window.RFW_WASM_ENCODINGS = [];
window.RFW_WASM_NEGOTIATED = false;
window.RFW_BUILD_MODE = "production";
window.RFW_WASM_DELIVERY = "embedded";
const embeddedRequests = [];
globalThis.fetch = async (url) => {
    embeddedRequests.push(url);
    return new Response(null, { status: 404 });
};
await assert.rejects(
    WasmLoader.load("/app.wasm?v=test", {
        go: { importObject: {}, run() {} },
        skipLoader: true,
    }),
    /Tried embedded.+build\.delivery set to embedded/s,
);
assert.deepEqual(embeddedRequests, ["/app.wasm?v=test"]);

// A corrupt packaged bundle fails at instantiation instead of at fetch, and
// that path has to be as visible as a failed download.
globalThis.fetch = async () =>
    new Response(Buffer.from("corrupt"), {
        status: 200,
        headers: { "content-type": "application/wasm" },
    });
const instantiateStreaming = WebAssembly.instantiateStreaming;
const instantiate = WebAssembly.instantiate;
WebAssembly.instantiateStreaming = async () => {
    throw new Error("magic word not detected");
};
WebAssembly.instantiate = async () => {
    throw new Error("magic word not detected");
};
await assert.rejects(
    WasmLoader.load("/app.wasm?v=test", {
        go: { importObject: {}, run() {} },
        skipLoader: true,
    }),
    /Failed to instantiate.+build\.delivery set to embedded/s,
);
WebAssembly.instantiateStreaming = instantiateStreaming;
WebAssembly.instantiate = instantiate;

// A configuration written by an older build defines no delivery mode. That is
// the network contract, unchanged.
delete window.RFW_WASM_DELIVERY;
window.RFW_WASM_ENCODINGS = ["br", "gzip"];
window.RFW_WASM_NEGOTIATED = true;
window.RFW_BUILD_MODE = "production";
const legacyRequests = [];
globalThis.fetch = async (url) => {
    legacyRequests.push(url);
    return new Response(wasm, {
        status: 200,
        headers: {
            "content-encoding": "br",
            "content-type": "application/wasm",
        },
    });
};
instantiated = undefined;
await WasmLoader.load("/app.wasm?v=test", {
    go: { importObject: {}, run: () => new Promise(() => {}) },
    skipLoader: true,
    reloadOnExit: false,
    recoveryWindow: 1,
});
assert.deepEqual(instantiated, wasm);
assert.deepEqual(legacyRequests, ["/app.wasm?v=test"]);
