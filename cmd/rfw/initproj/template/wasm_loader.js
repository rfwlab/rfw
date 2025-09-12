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
        return {
            start() {
                let progress = 0;
                this.interval = setInterval(() => {
                    progress = Math.min(progress + Math.random() * 10, 90);
                    bar.style.width = progress + "%";
                }, 200);
            },
            finish() {
                clearInterval(this.interval);
                bar.style.width = "100%";
                setTimeout(() => bar.remove(), 300);
            },
        };
    }
    async function load(url, { go, color, height, blur, skipLoader } = {}) {
        let bar;
        if (!skipLoader) {
            bar = createBar({ color, height, blur });
            bar.start();
        }
        const resp = await fetch(url);
        const bytes = await resp.arrayBuffer();
        if (bar) bar.finish();
        const result = await WebAssembly.instantiate(bytes, go.importObject);
        go.run(result.instance);
        return result;
    }
    global.WasmLoader = { load };
})(window);
