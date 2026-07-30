(function (global) {
    function createBar(opts) {
        const bar = document.createElement("div");
        const color = opts.color || "#ff0000";
        const blur = opts.blur || "8px";
        Object.assign(bar.style, {
            position: "fixed",
            top: 0,
            left: 0,
            width: "0%",
            height: opts.height || "4px",
            background: color,
            boxShadow: `0 0 ${blur} ${color}`,
            zIndex: 9999,
            transition: "width 0.3s ease",
        });
        document.body.appendChild(bar);
        let interval;
        return {
            start() {
                let progress = 0;
                interval = setInterval(() => {
                    progress = Math.min(progress + Math.random() * 10, 90);
                    bar.style.width = progress + "%";
                }, 200);
            },
            finish(success = true) {
                if (interval) {
                    clearInterval(interval);
                }
                if (success) {
                    bar.style.width = "100%";
                    setTimeout(() => bar.remove(), 300);
                } else {
                    bar.remove();
                }
            },
        };
    }

    function buildCandidates(url) {
        const trimmed = url.trim();
        if (!trimmed) return [];
        const queryIndex = trimmed.indexOf("?");
        const base = queryIndex === -1 ? trimmed : trimmed.slice(0, queryIndex);
        const query = queryIndex === -1 ? "" : trimmed.slice(queryIndex);
        const candidates = [];
        if (base.endsWith(".wasm") && !base.endsWith(".wasm.br")) {
            candidates.push(`${base}.br${query}`);
        }
        candidates.push(trimmed);
        return candidates;
    }

    async function fetchWithFallback(candidates) {
        let lastError;
        for (const candidate of candidates) {
            try {
                const resp = await fetch(candidate);
                if (!resp.ok) {
                    lastError = new Error(`unexpected status ${resp.status}`);
                    continue;
                }
                return resp;
            } catch (err) {
                lastError = err;
            }
        }
        throw lastError || new Error("no wasm candidates provided");
    }

    async function load(url, { go, color, height, blur, skipLoader } = {}) {
        const candidates = buildCandidates(url);
        if (candidates.length === 0) {
            throw new Error("wasm url is empty");
        }

        let bar;
        if (!skipLoader) {
            bar = createBar({ color, height, blur });
            bar.start();
        }

        let response;
        try {
            response = await fetchWithFallback(candidates);
        } catch (err) {
            if (bar) bar.finish(false);
            console.error("Failed to load Wasm bundle", candidates, err);
            throw err;
        }

        let result;
        try {
            if (WebAssembly.instantiateStreaming) {
                const fallback = response.clone();
                try {
                    result = await WebAssembly.instantiateStreaming(
                        response,
                        go.importObject,
                    );
                } catch {
                    const bytes = await fallback.arrayBuffer();
                    result = await WebAssembly.instantiate(
                        bytes,
                        go.importObject,
                    );
                }
            } else {
                const bytes = await response.arrayBuffer();
                result = await WebAssembly.instantiate(bytes, go.importObject);
            }
        } catch (err) {
            if (bar) bar.finish(false);
            console.error("Failed to instantiate Wasm bundle", err);
            throw err;
        }
        if (bar) bar.finish(true);
        go.run(result.instance);
        return result;
    }

    global.WasmLoader = { load };
})(window);
