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

    function recoveryKey(url) {
        return `rfw:runtime-reload:${location.pathname}:${url}`;
    }

    function readRecovery(key) {
        try {
            return Number(sessionStorage.getItem(key) || 0);
        } catch (error) {
            console.warn("Unable to read WebAssembly recovery state", error);
            return 0;
        }
    }

    function writeRecovery(key, value) {
        try {
            sessionStorage.setItem(key, String(value));
            return true;
        } catch (error) {
            console.warn("Unable to persist WebAssembly recovery state", error);
            return false;
        }
    }

    function clearRecoveryState(key) {
        try {
            sessionStorage.removeItem(key);
        } catch (error) {
            console.warn("Unable to clear WebAssembly recovery state", error);
        }
    }

    function showRuntimeFailure(error) {
        const existing = document.getElementById("rfw-runtime-failure");
        if (existing) return;

        const panel = document.createElement("div");
        panel.id = "rfw-runtime-failure";
        Object.assign(panel.style, {
            position: "fixed",
            inset: 0,
            zIndex: 1000000,
            display: "grid",
            placeItems: "center",
            padding: "24px",
            background: "rgba(17, 24, 39, 0.9)",
            fontFamily: "system-ui, sans-serif",
        });
        const message = document.createElement("div");
        Object.assign(message.style, {
            maxWidth: "640px",
            padding: "24px",
            borderRadius: "12px",
            background: "white",
            color: "#111827",
            boxShadow: "0 20px 60px rgba(0, 0, 0, 0.35)",
        });
        const title = document.createElement("strong");
        title.textContent = "The application runtime stopped";
        const detail = document.createElement("pre");
        detail.textContent = String(error || "WebAssembly runtime exited");
        Object.assign(detail.style, {
            whiteSpace: "pre-wrap",
            overflowWrap: "anywhere",
            color: "#991b1b",
        });
        const reload = document.createElement("button");
        reload.type = "button";
        reload.textContent = "Reload application";
        reload.addEventListener("click", () => location.reload());
        message.append(title, detail, reload);
        panel.appendChild(message);
        document.body.appendChild(panel);
    }

    function handleRuntimeExit(url, error, reloadOnExit, recoveryWindow) {
        console.error("WebAssembly runtime stopped", error);
        const key = recoveryKey(url);
        const lastReload = readRecovery(key);
        if (
            reloadOnExit &&
            Date.now() - lastReload > recoveryWindow &&
            writeRecovery(key, Date.now())
        ) {
            location.reload();
            return;
        }
        showRuntimeFailure(error);
    }

    async function load(
        url,
        {
            go,
            color,
            height,
            blur,
            skipLoader,
            reloadOnExit = true,
            recoveryWindow = 30000,
        } = {},
    ) {
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
        const key = recoveryKey(url);
        const recoveryTimer = setTimeout(
            () => clearRecoveryState(key),
            recoveryWindow,
        );
        Promise.resolve(go.run(result.instance)).then(
            () => {
                clearTimeout(recoveryTimer);
                handleRuntimeExit(
                    url,
                    new Error("WebAssembly runtime exited"),
                    reloadOnExit,
                    recoveryWindow,
                );
            },
            (err) => {
                clearTimeout(recoveryTimer);
                handleRuntimeExit(url, err, reloadOnExit, recoveryWindow);
            },
        );
        return result;
    }

    global.WasmLoader = { load };
})(window);
