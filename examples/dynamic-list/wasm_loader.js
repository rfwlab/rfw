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
        let indeterminate = false;
        function stopAnimation() {
            if (interval) {
                clearInterval(interval);
                interval = undefined;
            }
        }
        return {
            // An encoded response reports the compressed length while the
            // reader yields decoded bytes, so any ratio between them is a
            // fiction. Sweep instead of inventing a percentage.
            startIndeterminate() {
                stopAnimation();
                indeterminate = true;
                bar.style.width = "35%";
                bar.style.transition = "transform 1.1s ease-in-out";
                let forward = true;
                interval = setInterval(() => {
                    bar.style.transform = forward
                        ? "translateX(190%)"
                        : "translateX(0)";
                    forward = !forward;
                }, 1100);
            },
            // Only used when the response is not encoded, where the received
            // bytes and Content-Length measure the same thing.
            setProgress(received, total) {
                if (indeterminate || !total) return;
                stopAnimation();
                const percent = Math.min((received / total) * 100, 100);
                bar.style.width = percent + "%";
            },
            finish(success = true) {
                stopAnimation();
                bar.style.transition = "width 0.3s ease";
                bar.style.transform = "translateX(0)";
                if (success) {
                    bar.style.width = "100%";
                    setTimeout(() => bar.remove(), 300);
                } else {
                    bar.remove();
                }
            },
        };
    }

    // The suffix each content coding is stored under. This is the contract
    // between `rfw build` and the loader; a static host serves these names
    // directly because it cannot negotiate.
    const ARTIFACT_EXT = { br: ".br", gzip: ".gz" };

    // DecompressionStream implements gzip and deflate, never brotli. A brotli
    // artifact is therefore only usable when the server labels it and the
    // browser's own HTTP layer decodes it.
    const CLIENT_DECODABLE = new Set(["gzip"]);

    function buildConfig() {
        return {
            encodings: Array.isArray(global.RFW_WASM_ENCODINGS)
                ? global.RFW_WASM_ENCODINGS
                : [],
            negotiated: global.RFW_WASM_NEGOTIATED === true,
            production: global.RFW_BUILD_MODE === "production",
        };
    }

    // buildPlan orders the ways this bundle can be fetched, best first. A
    // negotiated request is preferred because the browser picks the coding it
    // actually supports; named artifacts follow for static hosting.
    function buildPlan(url, config) {
        const trimmed = url.trim();
        if (!trimmed) return [];
        const queryIndex = trimmed.indexOf("?");
        const base = queryIndex === -1 ? trimmed : trimmed.slice(0, queryIndex);
        const query = queryIndex === -1 ? "" : trimmed.slice(queryIndex);
        const plan = [];
        if (config.negotiated) {
            plan.push({ url: trimmed, kind: "negotiated" });
        }
        for (const encoding of config.encodings) {
            const ext = ARTIFACT_EXT[encoding];
            if (ext && !base.endsWith(ext)) {
                plan.push({ url: base + ext + query, kind: encoding });
            }
        }
        // The raw bundle is a development convenience. In production it is the
        // multi-megabyte download that a broken compressed path would
        // otherwise hide, so it is not an option.
        if (!config.production) {
            plan.push({ url: trimmed, kind: "identity" });
        }
        return plan;
    }

    // fetchArtifact resolves one step of the plan into a Response carrying
    // decoded wasm, or throws so the caller can try the next step.
    async function fetchArtifact(step, production) {
        const response = await fetch(step.url);
        if (!response.ok) {
            throw new Error(`${step.url} responded ${response.status}`);
        }
        const encoding = (
            response.headers.get("content-encoding") || ""
        ).toLowerCase();
        const encoded = encoding !== "" && encoding !== "identity";

        if (step.kind === "negotiated") {
            if (production && !encoded) {
                throw new Error(
                    `${step.url} was served uncompressed; the host did not negotiate an encoding`,
                );
            }
            return { response, encoded };
        }
        if (step.kind === "identity") {
            return { response, encoded: false };
        }
        if (encoding === step.kind) {
            // The server labelled it, so the browser already decoded it.
            return { response, encoded: true };
        }
        if (encoded) {
            throw new Error(
                `${step.url} was served as ${encoding}, expected ${step.kind}`,
            );
        }
        if (!CLIENT_DECODABLE.has(step.kind) || !response.body) {
            throw new Error(
                `${step.url} was served without Content-Encoding and ${step.kind} cannot be decoded in the browser`,
            );
        }
        // A static host handed over the artifact bytes unlabelled. Decoding
        // here is what makes plain static hosting work without a proxy.
        const stream = response.body.pipeThrough(
            new DecompressionStream(step.kind),
        );
        return {
            // instantiateStreaming needs a Response with the wasm media type,
            // not a bare stream.
            response: new Response(stream, {
                headers: { "Content-Type": "application/wasm" },
            }),
            encoded: true,
        };
    }

    // trackProgress re-wraps a response so the bar can follow it. It is only
    // used when the body is not encoded, where received bytes and
    // Content-Length measure the same thing.
    function trackProgress(response, total, bar) {
        const reader = response.body.getReader();
        let received = 0;
        const stream = new ReadableStream({
            async pull(controller) {
                const { done, value } = await reader.read();
                if (done) {
                    controller.close();
                    return;
                }
                received += value.byteLength;
                bar.setProgress(received, total);
                controller.enqueue(value);
            },
            cancel(reason) {
                return reader.cancel(reason);
            },
        });
        return new Response(stream, {
            headers: { "Content-Type": "application/wasm" },
        });
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
        const config = buildConfig();
        const plan = buildPlan(url, config);
        if (plan.length === 0) {
            throw new Error("wasm url is empty");
        }

        let bar;
        if (!skipLoader) {
            bar = createBar({ color, height, blur });
        }

        let artifact;
        const failures = [];
        for (const step of plan) {
            try {
                artifact = await fetchArtifact(step, config.production);
                break;
            } catch (err) {
                failures.push(`${step.kind}: ${err.message}`);
            }
        }
        if (!artifact) {
            if (bar) bar.finish(false);
            const detail = failures.join("; ");
            const error = new Error(
                `Failed to load the WebAssembly bundle. Tried ${plan
                    .map((step) => step.kind)
                    .join(", ")}. ${detail}. ` +
                    "Check that `rfw build` produced app.wasm.br and app.wasm.gz and that the server sends Content-Encoding for them.",
            );
            console.error(error);
            showRuntimeFailure(error);
            throw error;
        }

        let response = artifact.response;
        if (bar) {
            if (artifact.encoded) {
                bar.startIndeterminate();
            } else {
                const total = Number(
                    response.headers.get("content-length") || 0,
                );
                if (total > 0 && response.body) {
                    response = trackProgress(response, total, bar);
                } else {
                    bar.startIndeterminate();
                }
            }
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
