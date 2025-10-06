const markup = `
<style>
:root{
  /* Base palette */
  --bg:#0f1115; --bg-2:#0b0d12; --panel:#121419; --elev:#171923; --text:#e6e9f2; --muted:#8b93a7;
  --border:#4a3737; --border-2:#3a2426; --chip-bg:#0000007d; --tile-bg:#12080b; --tile-border:#2b191b; --tile-hover:#1a0f12;
  --rose-50:#ffe6e7; --rose-100:#ffe1e2; --rose-200:#ffd6d8; --rose-300:#ffccd0; --rose-400:#ffb3b5;
  --accent:#ff4d4f; --accent-2:#ff6b6b; --good:#22c55e; --warn:#f59e0b; --bad:#ef4444;
  --shadow:0 10px 30px rgba(0,0,0,.4); --round:14px;
}
*{box-sizing:border-box}
.rfw-button{font:inherit}
.hidden{display:none !important}

/* Toggle FAB */
.rfw-fab{ position:fixed; right:16px; bottom:16px; z-index:2147483000; height:48px; width:48px; border-radius:999px; border:1px solid var(--border); background:linear-gradient(180deg, var(--panel), var(--bg)); color:var(--text); display:inline-flex; align-items:center; justify-content:center; box-shadow:var(--shadow); cursor:pointer; transition:transform .12s ease, background .2s ease, border-color .2s ease }
.rfw-fab:hover{transform:translateY(-1px); border-color:var(--border-2)}
.rfw-fab svg{width:22px;height:22px}

/* Overlay shell (bottom dock only) */
.rfw-overlay{ position:fixed; inset:auto 0 0 0; height:48vh; z-index:2147483640; background:linear-gradient(180deg, rgba(20,11,13,.8), rgba(13,7,9,.95)); backdrop-filter: blur(10px); border-top:1px solid var(--border); box-shadow:var(--shadow); display:flex; flex-direction:column }

.rfw-header{ display:flex; align-items:center; gap:10px; padding:0 12px; border-bottom:1px solid var(--border) }
.rfw-badge{ padding:3px 8px; border-radius:999px; color:var(--rose-300); font-size:12px }
.rfw-title{font-weight:600; letter-spacing:.2px}
.rfw-spacer{flex:1}
.rfw-ctl{display:flex; align-items:center; gap:6px}
.rfw-iconbtn{ display:inline-flex; align-items:center; justify-content:center; height:30px; width:30px; color:var(--rose-200); cursor:pointer; transition:background .15s ease, border-color .15s ease }
.rfw-iconbtn:hover{color:var(--border-2)}
.rfw-iconbtn svg{display:flex; align-items:center; justify-content:center; width:16px; height:16px}

/* KPI row */
.rfw-kpi{ display:flex; gap:10px; padding:5px 12px; border-bottom:1px dashed var(--border-2) }
.rfw-chip{ display:inline-flex; align-items:center; gap:8px; padding:3px 10px; border:1px solid var(--border); border-radius:999px; background:var(--chip-bg); color:var(--rose-200); font-variant-numeric:tabular-nums }
.rfw-dot{width:7px; height:7px; border-radius:999px}
.ok{background:var(--good)} .warn{background:var(--warn)} .bad{background:var(--bad)} .info{background:var(--accent)}

/* Tabs */
.rfw-tabs{display:flex; align-items:center; gap:6px; padding:0 8px 0 8px}
.rfw-tab{ padding: 8px 12px; border-bottom: 2px solid transparent; color: var(--rose-300); cursor: pointer; user-select: none; }
.rfw-tab[aria-selected="true"]{background: var(--chip-bg); color: var(--rose-50); border-color: var(--rose-300);}
.rfw-panels{ display:flex; gap:0; flex:1; min-height:0; border-top:1px solid var(--border-2); margin:0 0 8px; border-radius:0 0 var(--round) var(--round); overflow:hidden }

/* Components panel */
.rfw-split{display:flex; flex:1; min-height:0}
.rfw-tree{ width:42%; min-width:240px; max-width:55%; border-right:1px solid var(--border); background:var(--chip-bg); display:flex; flex-direction:column }
.rfw-tree .rfw-search{ display:flex; align-items:center; gap:4px; border-bottom:1px solid var(--border) }
.rfw-tree .rfw-search .rfw-input{flex:1}
.rfw-input{ width:100%; padding:8px 10px; border:0; background:transparent; color:var(--text); outline:none; transition:border-color .12s ease, background .12s ease }
.tree-scroll{overflow:auto; padding:8px}
.node{display:flex; align-items:center; gap:8px; padding:6px 8px; border-radius:8px; cursor:pointer; color:var(--rose-100)}
.node:hover{background:var(--tile-hover)}
.node.active{background:var(--tile-hover)}
.node .kind{font-size:11px; color:var(--accent-2); padding:2px 6px; border:1px solid var(--border-2); border-radius:999px; background:var(--tile-bg)}
.node .name{font-weight:600}
.node .meta{margin-left:auto; display:flex; align-items:center; gap:6px; font-variant-numeric:tabular-nums; color:var(--rose-300)}
.node .meta span{display:inline-flex; align-items:center; gap:4px; padding:2px 8px; border-radius:999px; border:1px solid var(--tile-border); background:var(--tile-bg)}
.route-action{padding:4px 10px; border-radius:8px; border:1px solid var(--tile-border); background:var(--tile-bg); color:var(--rose-50); font-size:12px; cursor:pointer; transition:background .12s ease, border-color .12s ease}
.route-action:hover{background:var(--tile-hover); border-color:var(--border-2)}
.rfw-route-popup{position:absolute; right:18px; bottom:70px; width:280px; background:var(--panel); border:1px solid var(--border); border-radius:12px; box-shadow:var(--shadow); display:flex; flex-direction:column; padding:12px; gap:12px; z-index:2147483650}
.rfw-route-popup.hidden{display:none !important}
.rfw-route-popup h3{margin:0; font-size:16px; color:var(--rose-50)}
.rfw-route-popup p{margin:0; font-size:12px; color:var(--rose-200)}
.route-fields{display:flex; flex-direction:column; gap:8px}
.route-field{display:flex; flex-direction:column; gap:4px; font-size:12px; color:var(--rose-200)}
.route-field input{padding:6px 8px; border-radius:8px; border:1px solid var(--tile-border); background:var(--bg-2); color:var(--rose-50); font:inherit; outline:none}
.route-actions{display:flex; gap:8px; justify-content:flex-end}
.route-actions button{padding:6px 12px; border-radius:8px; border:1px solid var(--tile-border); background:var(--tile-bg); color:var(--rose-50); cursor:pointer; font:inherit}
.route-actions button.primary{background:var(--accent); border-color:var(--accent); color:#fff}
.route-actions button.primary:hover{background:var(--accent-2); border-color:var(--accent-2)}
.route-actions button:hover{background:var(--tile-hover)}
.route-error{color:var(--bad); font-size:12px; min-height:16px}

.rfw-detail{overflow-y:auto; flex:1;background:var(--chip-bg);display:flex;flex-direction:column}
.rfw-detail .rfw-subheader{display:flex;align-items:center;gap:10px;padding:8px 12px;border-bottom:1px solid var(--border)}
.kv{display:grid; grid-template-columns:160px 1fr; gap:8px 12px; padding:12px}
.kv div{padding:6px 8px; background:var(--tile-bg); border:1px solid var(--tile-border); border-radius:10px; overflow:auto}
.kv b{color:var(--rose-200)}
.store-bindings{display:flex; flex-direction:column; gap:6px}
.store-binding{display:flex; align-items:center; gap:8px; flex-wrap:wrap}
.store-badge{padding:2px 8px; border-radius:999px; border:1px solid var(--tile-border); background:var(--tile-bg); color:var(--rose-200)}
.store-keys{display:flex; flex-wrap:wrap; gap:6px}
.store-key{padding:2px 6px; border-radius:6px; border:1px solid var(--tile-border); background:var(--bg-2); font-size:12px; color:var(--rose-100)}
.store-key.muted{opacity:.65}

.mono{font-variant-numeric:tabular-nums}
.timeline{display:flex; flex-direction:column; gap:6px; font-variant-numeric:tabular-nums}
.timeline-item{display:flex; align-items:center; gap:8px}
.timeline-item .mono{background:var(--tile-bg); border:1px solid var(--tile-border); padding:2px 6px; border-radius:6px}
.timeline-item .duration{opacity:.75}

/* Network panel */
.net-list{flex:1; overflow:auto; padding:8px}
.net{display:grid; grid-template-columns:1fr 80px 80px; gap:8px; padding:8px; border-bottom:1px dashed var(--tile-border); color:var(--rose-100)}
.net .url{word-break:break-all}
.net .mono{font-variant-numeric:tabular-nums}

/* Logs panel */
.rfw-logs{display:flex;flex-direction:column;height:100%;flex:1}
.log-toolbar{display:flex; gap:8px; border-bottom:1px solid var(--border);}
.log-list{flex:1; overflow:auto; padding:8px}
.log{ display:grid; grid-template-columns:100px 80px 1fr; gap:8px; padding:8px; border-bottom:1px dashed var(--tile-border); color:var(--rose-100) }
.log .lvl{font-weight:600}
.log[data-lvl="warn"] .lvl{color:var(--warn)}
.log[data-lvl="error"] .lvl{color:var(--bad)}
.log[data-lvl="info"] .lvl{color:var(--accent)}
.log .msg{white-space:pre-wrap; word-break:break-word}

/* Resize handle (bottom only) */
.rfw-resize-h{position:absolute; left:0; right:0; top:-4px; height:8px; cursor:ns-resize}

.kbd{border:1px solid var(--border-2); background:var(--tile-bg); padding:2px 6px; border-radius:6px; font-size:12px; color:var(--rose-200)}
@media (max-width:960px){ .rfw-tree{width:48%} }
@media (max-width:720px){ .rfw-split{flex-direction:column} .rfw-tree{width:100%; max-width:none; border-right:none; border-bottom:1px solid var(--border)} }
</style>

<button id="rfwDevtoolsToggle" class="rfw-button rfw-fab" data-rfw-ignore aria-label="Open DevTools" title="DevTools (Ctrl+Shift+D)">
  <span>rfw</span>
</button>

<section id="rfwDevtools" class="rfw-overlay hidden" data-rfw-ignore role="dialog" aria-modal="false" aria-label="rfw DevTools">
  <div class="rfw-resize-h" data-resize="h"></div>

  <header class="rfw-header">
    <span class="rfw-badge">RFW DevTools</span>
    <span class="rfw-spacer"></span>
    <div class="rfw-ctl">
      <button class="rfw-button rfw-iconbtn" id="minBtn" title="Minimize">
        <svg viewBox="0 0 24 24" fill="none"><path d="M6 12h12" stroke="var(--rose-400)" stroke-width="2" stroke-linecap="round"/></svg>
      </button>
      <button class="rfw-iconbtn" id="closeBtn" title="Close">
        <svg viewBox="0 0 24 24" fill="none"><path d="M7 7l10 10M17 7 7 17" stroke="var(--accent-2)" stroke-width="1.8" stroke-linecap="round"/></svg>
      </button>
    </div>
  </header>

  <div class="rfw-kpi">
    <span class="rfw-chip"><span class="rfw-dot ok"></span><b>FPS</b> <span id="kpiFps" class="mono">0</span></span>
    <span class="rfw-chip"><span class="rfw-dot warn"></span><b>Mem</b> <span id="kpiMem" class="mono">n/a</span><span id="kpiMemSrc" class="mono" style="opacity:.7;font-size:12px;margin-left:6px"></span></span>
    <span class="rfw-chip"><span class="rfw-dot ok"></span><b>Nodes</b> <span id="kpiNodes" class="mono">0</span></span>
    <span class="rfw-chip"><span class="rfw-dot ok"></span><b>Render</b> <span id="kpiRender" class="mono">n/a</span></span>
    <span class="rfw-spacer"></span>
    <span class="rfw-chip"><b>Hotkey</b> <span class="kbd">Ctrl</span>+<span class="kbd">Shift</span>+<span class="kbd">D</span></span>
  </div>

    <nav class="rfw-tabs" role="tablist" aria-label="Tabs">
    <button class="rfw-button rfw-tab" role="tab" aria-selected="true" aria-controls="tab-components" id="tabbtn-components">Components</button>
    <button class="rfw-button rfw-tab" role="tab" aria-selected="false" aria-controls="tab-routes" id="tabbtn-routes">Routes</button>
    <button class="rfw-button rfw-tab" role="tab" aria-selected="false" aria-controls="tab-store" id="tabbtn-store">Store</button>
      <button class="rfw-button rfw-tab" role="tab" aria-selected="false" aria-controls="tab-signals" id="tabbtn-signals">Signals</button>
      <button class="rfw-button rfw-tab" role="tab" aria-selected="false" aria-controls="tab-plugins" id="tabbtn-plugins">Plugins</button>
      <button class="rfw-button rfw-tab" role="tab" aria-selected="false" aria-controls="tab-network" id="tabbtn-network">Network</button>
      <button class="rfw-button rfw-tab" role="tab" aria-selected="false" aria-controls="tab-logs" id="tabbtn-logs">Logs</button>
    <button class="rfw-button rfw-tab" role="tab" aria-selected="false" aria-controls="tab-vars" id="tabbtn-vars">Vars</button>
    <button class="rfw-button rfw-tab" role="tab" aria-selected="false" aria-controls="tab-pprof" id="tabbtn-pprof">Pprof</button>
  </nav>

  <div class="rfw-panels">

    <!-- Components -->
    <section id="tab-components" role="tabpanel" aria-labelledby="tabbtn-components" style="display:flex;flex:1">
      <div class="rfw-split">
        <aside class="rfw-tree">
          <div class="rfw-search">
            <input id="treeFilter" class="rfw-input" type="search" placeholder="Filter components… (e.g. Header, Button)" />
            <button class="rfw-button rfw-iconbtn" id="refreshTree" title="Refresh components">
              <svg viewBox="0 0 24 24" fill="none"><path d="M4 4v6h6M20 20v-6h-6M5 19a9 9 0 0 1 14-7M19 5a9 9 0 0 0-14 7" stroke="var(--rose-400)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button>
          </div>
          <div class="tree-scroll" id="treeContainer"></div>
        </aside>

        <article class="rfw-detail">
          <div class="rfw-subheader">
            <span style="font-weight:600" id="detailTitle">Select a component</span>
            <span class="rfw-badge" id="detailKind">-</span>
            <span class="rfw-spacer"></span>
          </div>
          <div class="kv" id="detailKV"></div>
        </article>
      </div>
    </section>

    <!-- Routes -->
    <section id="tab-routes" role="tabpanel" aria-labelledby="tabbtn-routes" class="hidden" style="display:flex;flex:1">
      <div class="rfw-split">
        <aside class="rfw-tree">
          <div class="rfw-search">
            <input id="routeFilter" class="rfw-input" type="search" placeholder="Filter routes…" />
            <button class="rfw-button rfw-iconbtn" id="refreshRoutes" title="Refresh routes">
              <svg viewBox="0 0 24 24" fill="none"><path d="M4 4v6h6M20 20v-6h-6M5 19a9 9 0 0 1 14-7M19 5a9 9 0 0 0-14 7" stroke="var(--rose-400)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button>
          </div>
          <div class="tree-scroll" id="routeTree"></div>
        </aside>
        <article class="rfw-detail">
          <div class="rfw-subheader">
            <span style="font-weight:600" id="routeTitle">Select a route</span>
            <span class="rfw-spacer"></span>
            <button class="rfw-button route-action" id="routeDetailGo" title="Navigate to route">Open</button>
          </div>
          <div class="kv" id="routeDetail"></div>
        </article>
      </div>
    </section>

    <!-- Store -->
    <section id="tab-store" role="tabpanel" aria-labelledby="tabbtn-store" class="hidden" style="display:flex;flex:1">
      <div class="rfw-split">
        <aside class="rfw-tree">
          <div class="rfw-search">
            <input id="storeFilter" class="rfw-input" type="search" placeholder="Filter stores…" />
            <button class="rfw-button rfw-iconbtn" id="refreshStore" title="Refresh stores">
              <svg viewBox="0 0 24 24" fill="none"><path d="M4 4v6h6M20 20v-6h-6M5 19a9 9 0 0 1 14-7M19 5a9 9 0 0 0-14 7" stroke="var(--rose-400)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button>
          </div>
          <div class="tree-scroll" id="storeTree"></div>
        </aside>
        <article class="rfw-detail">
          <div class="rfw-subheader">
            <span style="font-weight:600" id="storeTitle">Select a key</span>
            <span class="rfw-spacer"></span>
          </div>
          <pre id="storeContent" style="flex:1;margin:0;padding:12px;overflow:auto"></pre>
        </article>
      </div>
    </section>

      <!-- Signals -->
      <section id="tab-signals" role="tabpanel" aria-labelledby="tabbtn-signals" class="hidden" style="display:flex;flex:1">
      <div class="rfw-split">
        <aside class="rfw-tree">
          <div class="rfw-search">
            <input id="signalFilter" class="rfw-input" type="search" placeholder="Filter signals…" />
            <button class="rfw-button rfw-iconbtn" id="refreshSignals" title="Refresh signals">
              <svg viewBox="0 0 24 24" fill="none"><path d="M4 4v6h6M20 20v-6h-6M5 19a9 9 0 0 1 14-7M19 5a9 9 0 0 0-14 7" stroke="var(--rose-400)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button>
          </div>
          <div class="tree-scroll" id="signalList"></div>
        </aside>
        <article class="rfw-detail">
          <div class="rfw-subheader">
            <span style="font-weight:600" id="signalTitle">Select a signal</span>
            <span class="rfw-spacer"></span>
          </div>
          <pre id="signalContent" style="flex:1;margin:0;padding:12px;overflow:auto"></pre>
        </article>
      </div>
      </section>
      <!-- Plugins -->
      <section id="tab-plugins" role="tabpanel" aria-labelledby="tabbtn-plugins" class="hidden" style="display:flex;flex:1">
      <div class="rfw-split">
        <aside class="rfw-tree">
          <div class="rfw-search">
            <input id="pluginFilter" class="rfw-input" type="search" placeholder="Filter plugins…" />
            <button class="rfw-button rfw-iconbtn" id="refreshPlugins" title="Refresh plugins">
              <svg viewBox="0 0 24 24" fill="none"><path d="M4 4v6h6M20 20v-6h-6M5 19a9 9 0 0 1 14-7M19 5a9 9 0 0 0-14 7" stroke="var(--rose-400)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button>
          </div>
          <div class="tree-scroll" id="pluginTree"></div>
        </aside>
        <article class="rfw-detail">
          <div class="rfw-subheader">
            <span style="font-weight:600" id="pluginTitle">Select a plugin</span>
            <span class="rfw-spacer"></span>
          </div>
          <pre id="pluginContent" style="flex:1;margin:0;padding:12px;overflow:auto"></pre>
        </article>
      </div>
      </section>

      <!-- Network -->
      <section id="tab-network" role="tabpanel" aria-labelledby="tabbtn-network" class="hidden" style="width:100%">
        <div class="rfw-logs">
          <div class="log-toolbar">
            <span class="rfw-spacer"></span>
            <button class="rfw-button rfw-iconbtn" id="clearNet" title="Clear requests">
              <svg viewBox="0 0 24 24" fill="none"><path d="M4 7h16M9 7v12m6-12v12M6 7l1-2h10l1 2" stroke="var(--rose-400)"/></svg>
            </button>
          </div>
          <div id="netList" class="net-list" aria-live="polite"></div>
        </div>
      </section>

      <!-- Logs -->
      <section id="tab-logs" role="tabpanel" aria-labelledby="tabbtn-logs" class="hidden" style="width:100%">
      <div class="rfw-logs">
        <div class="log-toolbar">
          <input id="logFilter" class="rfw-input" placeholder="Filter logs… (text, level, tag)" />
          <span class="rfw-spacer"></span>
          <button class="rfw-button rfw-iconbtn" id="clearLogs" title="Clear logs">
            <svg viewBox="0 0 24 24" fill="none"><path d="M4 7h16M9 7v12m6-12v12M6 7l1-2h10l1 2" stroke="var(--rose-400)"/></svg>
          </button>
        </div>
        <div id="logList" class="log-list" aria-live="polite"></div>
      </div>
    </section>

    <!-- Vars -->
    <section id="tab-vars" role="tabpanel" aria-labelledby="tabbtn-vars" class="hidden" style="display:flex;flex:1">
      <div class="rfw-split">
        <aside class="rfw-tree">
          <div class="rfw-search">
            <input id="varsFilter" class="rfw-input" type="search" placeholder="Filter vars…" />
          </div>
          <div class="tree-scroll" id="varsTree"></div>
        </aside>
        <article class="rfw-detail">
          <div class="rfw-subheader">
            <span style="font-weight:600" id="varsTitle">Select a variable</span>
            <span class="rfw-spacer"></span>
          </div>
          <pre id="varsContent" style="flex:1;margin:0;padding:12px;overflow:auto"></pre>
        </article>
      </div>
    </section>

    <!-- Pprof -->
    <section id="tab-pprof" role="tabpanel" aria-labelledby="tabbtn-pprof" class="hidden" style="display:flex;flex:1">
      <div class="rfw-split">
        <aside class="rfw-tree">
          <div class="rfw-search">
            <input id="pprofFilter" class="rfw-input" type="search" placeholder="Filter profiles…" />
          </div>
          <div class="tree-scroll" id="pprofLinks"></div>
        </aside>
        <article class="rfw-detail">
          <div class="rfw-subheader">
            <span style="font-weight:600" id="pprofTitle">Select a profile</span>
            <span class="rfw-spacer"></span>
          </div>
          <pre id="pprofContent" style="flex:1;margin:0;padding:12px;overflow:auto"></pre>
        </article>
      </div>
    </section>

  </div>
  <div id="routePopup" class="rfw-route-popup hidden" data-rfw-ignore role="dialog" aria-modal="false" aria-labelledby="routePopupTitle">
    <h3 id="routePopupTitle">Configure route</h3>
    <p id="routePopupInfo">Provide values for dynamic parameters.</p>
    <div class="route-fields" id="routePopupFields"></div>
    <div class="route-error" id="routePopupError"></div>
    <div class="route-actions">
      <button class="rfw-button" id="routePopupCancel" type="button">Cancel</button>
      <button class="rfw-button primary" id="routePopupConfirm" type="button">Navigate</button>
    </div>
  </div>
</section>
`;

if (!document.getElementById("rfwDevtools")) {
  document.body.insertAdjacentHTML("beforeend", markup);
}

const $ = (s, r = document) => r.querySelector(s);
const $$ = (s, r = document) => Array.from(r.querySelectorAll(s));
const overlay = $("#rfwDevtools");
const fab = $("#rfwDevtoolsToggle");
const minBtn = $("#minBtn");
const closeBtn = $("#closeBtn");
const hHandle = $('[data-resize="h"]');
const routeTree = $("#routeTree");
const routeTitle = $("#routeTitle");
const routeDetail = $("#routeDetail");
const routeDetailGo = $("#routeDetailGo");
const routeFilter = $("#routeFilter");
const routePopup = $("#routePopup");
const routePopupTitle = $("#routePopupTitle");
const routePopupInfo = $("#routePopupInfo");
const routePopupFields = $("#routePopupFields");
const routePopupError = $("#routePopupError");
const routePopupConfirm = $("#routePopupConfirm");
const routePopupCancel = $("#routePopupCancel");
let routeSnapshot = [];
let selectedRoutePath = "";
let selectedRoute = null;
let popupRoute = null;
routeDetailGo?.setAttribute("disabled", "true");

const tabs = [
  { btn: $("#tabbtn-components"), panel: $("#tab-components") },
  {
    btn: $("#tabbtn-routes"),
    panel: $("#tab-routes"),
    onShow: refreshRoutes,
  },
  { btn: $("#tabbtn-store"), panel: $("#tab-store"), onShow: refreshStore },
  {
    btn: $("#tabbtn-signals"),
    panel: $("#tab-signals"),
    onShow: refreshSignals,
  },
  {
    btn: $("#tabbtn-plugins"),
    panel: $("#tab-plugins"),
    onShow: refreshPlugins,
  },
  { btn: $("#tabbtn-network"), panel: $("#tab-network") },
  { btn: $("#tabbtn-logs"), panel: $("#tab-logs") },
  { btn: $("#tabbtn-vars"), panel: $("#tab-vars"), onShow: loadVars },
  { btn: $("#tabbtn-pprof"), panel: $("#tab-pprof"), onShow: loadPprof },
];

function setTab(name) {
  tabs.forEach((t) => {
    const active = t.btn && t.btn.id === `tabbtn-${name}`;
    t.btn?.setAttribute("aria-selected", active ? "true" : "false");
    t.panel?.classList.toggle("hidden", !active);
    if (t.panel?.id === "tab-components")
      t.panel.style.display = active ? "flex" : "none";
    if (active && typeof t.onShow === "function") t.onShow();
  });
}

tabs.forEach((t) =>
  t.btn?.addEventListener("click", () =>
    setTab(t.btn.id.replace("tabbtn-", "")),
  ),
);
setTab("components");

function openDevtools() {
  overlay?.classList.remove("hidden");
  if (fab) {
    fab.style.transform = "scale(0.98)";
    setTimeout(() => (fab.style.transform = ""), 120);
  }
  refreshTree();
  refreshRoutes();
  refreshStore();
  refreshSignals();
  refreshPlugins();
}
function closeDevtools() {
  overlay?.classList.add("hidden");
  closeRoutePopup();
}
function toggleDevtools() {
  overlay?.classList.contains("hidden") ? openDevtools() : closeDevtools();
}

fab?.addEventListener("click", toggleDevtools);
closeBtn?.addEventListener("click", closeDevtools);
minBtn?.addEventListener("click", () => {
  if (!overlay) return;
  overlay.style.height = overlay.style.height === "38vh" ? "48vh" : "38vh";
});

document.addEventListener("keydown", (e) => {
  if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key.toLowerCase() === "d") {
    e.preventDefault();
    toggleDevtools();
    return;
  }
  if (e.key === "Escape" && !routePopup?.classList.contains("hidden")) {
    e.preventDefault();
    closeRoutePopup();
  }
});

(function () {
  if (!hHandle || !overlay) return;
  let dragging = false,
    startY = 0,
    startH = 0;
  hHandle.addEventListener("mousedown", (e) => {
    dragging = true;
    startY = e.clientY;
    startH = overlay.getBoundingClientRect().height;
    document.body.style.userSelect = "none";
  });
  window.addEventListener("mousemove", (e) => {
    if (!dragging) return;
    const dy = startY - e.clientY;
    overlay.style.height =
      Math.max(200, Math.min(window.innerHeight - 80, startH + dy)) + "px";
  });
  window.addEventListener("mouseup", () => {
    dragging = false;
    document.body.style.userSelect = "";
  });
})();

const treeContainer = $("#treeContainer");
const detailTitle = $("#detailTitle");
const detailKind = $("#detailKind");
const detailKV = $("#detailKV");

function renderTree(list, root = true) {
  const frag = document.createDocumentFragment();
  list.forEach((node) => {
    const el = document.createElement("div");
    el.className = "node";
    el.dataset.id = node.id;
    const metrics = [];
    if (typeof node.average === "number" && Number.isFinite(node.average)) {
      metrics.push(`${node.average.toFixed(1)} ms avg`);
    } else if (typeof node.time === "number" && Number.isFinite(node.time)) {
      metrics.push(`${node.time.toFixed(1)} ms`);
    }
    if (typeof node.updates === "number" && node.updates > 0) {
      metrics.push(`${node.updates}×`);
    }
    if (Array.isArray(node.storeBindings) && node.storeBindings.length) {
      const total = node.storeBindings.length;
      metrics.push(`${total} store${total > 1 ? "s" : ""}`);
    }
    const meta = metrics.length
      ? `<span class="meta">${metrics
          .map((txt) => `<span>${escapeHTML(txt)}</span>`)
          .join("")}</span>`
      : "";
    el.innerHTML = `
      <span class="kind">${escapeHTML(node.kind || "")}</span>
      <span class="name">${escapeHTML(node.name || "")}</span>
      ${meta}
    `;
    el.addEventListener("click", () => selectNode(node));
    frag.appendChild(el);
    if (node.children?.length) {
      const pad = document.createElement("div");
      pad.style.marginLeft = "18px";
      pad.appendChild(renderTree(node.children, false));
      frag.appendChild(pad);
    }
  });
  if (root) {
    treeContainer?.replaceChildren(frag);
    return treeContainer;
  }
  return frag;
}

function selectNode(n) {
  if (!detailTitle || !detailKind || !detailKV) return;
  detailTitle.textContent = n.name;
  detailKind.textContent = n.kind;
  const rows = [];
  rows.push(renderRow("Path", n.path || ""));
  if (typeof n.updates === "number")
    rows.push(renderRow("Updates", String(n.updates)));
  if (typeof n.average === "number" && Number.isFinite(n.average))
    rows.push(renderRow("Average render", formatMs(n.average)));
  if (typeof n.time === "number" && Number.isFinite(n.time))
    rows.push(renderRow("Last render", formatMs(n.time)));
  if (typeof n.total === "number" && Number.isFinite(n.total))
    rows.push(renderRow("Total render", formatMs(n.total)));
  if (n.owner) rows.push(renderRow("Owner", n.owner));
  if (n.hostComponent) rows.push(renderRow("Host component", n.hostComponent));
  if (n.props && Object.keys(n.props).length)
    rows.push(renderJSONRow("Props", n.props));
  if (n.slots && Object.keys(n.slots).length)
    rows.push(renderJSONRow("Slots", n.slots));
  if (n.signals && Object.keys(n.signals).length)
    rows.push(renderJSONRow("Signals", n.signals));
  if (n.store) {
    const storeLabel =
      [n.store.module, n.store.name].filter(Boolean).join("/") || "-";
    rows.push(renderRow("Store", storeLabel));
    if (n.store.state && Object.keys(n.store.state).length) {
      rows.push(renderJSONRow("Store state", n.store.state));
    }
  }
  if (Array.isArray(n.storeBindings) && n.storeBindings.length) {
    rows.push(renderStoreBindingsRow("Store bindings", n.storeBindings));
  }
  if (Array.isArray(n.timeline) && n.timeline.length)
    rows.push(renderTimelineRow("Timeline", n.timeline));
  detailKV.innerHTML =
    rows.join("") || renderRow("Info", "No metadata available");
}

function escapeHTML(s) {
  return String(s).replace(
    /[&<>"']/g,
    (m) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        m
      ],
  );
}

function formatMs(value) {
  if (typeof value !== "number" || !Number.isFinite(value)) return "0.00 ms";
  const digits = value >= 100 ? 0 : 2;
  return `${value.toFixed(digits)} ms`;
}

function renderTimelineRow(label, events) {
  const items = events
    .map((ev) => {
      const offset = formatMs(ev.at ?? 0);
      const duration =
        typeof ev.duration === "number" && ev.duration > 0
          ? `<span class="mono duration">${formatMs(ev.duration)}</span>`
          : "";
      return `<div class="timeline-item"><span class="mono">${offset}</span><span>${escapeHTML(
        ev.kind || "",
      )}</span>${duration}</div>`;
    })
    .join("");
  return `<b>${escapeHTML(label)}</b><div class="timeline">${items}</div>`;
}

function renderStoreBindingsRow(label, bindings) {
  const items = bindings
    .map((binding) => {
      const storeLabel =
        [binding.module, binding.name].filter(Boolean).join("/") || "-";
      const keys =
        Array.isArray(binding.keys) && binding.keys.length
          ? binding.keys
              .map((key) => `<span class="store-key">${escapeHTML(key)}</span>`)
              .join("")
          : '<span class="store-key muted">(all keys)</span>';
      return `<div class="store-binding"><span class="store-badge">${escapeHTML(
        storeLabel,
      )}</span><div class="store-keys">${keys}</div></div>`;
    })
    .join("");
  return `<b>${escapeHTML(label)}</b><div class="store-bindings">${items}</div>`;
}

function formatJSON(value) {
  try {
    return JSON.stringify(value, null, 2);
  } catch (err) {
    return String(value);
  }
}

function renderRow(label, value, isMono = false) {
  const content = escapeHTML(value == null ? "" : value);
  const cls = isMono ? ' class="mono" style="white-space:pre-wrap"' : "";
  return `<b>${escapeHTML(label)}</b><div${cls}>${content}</div>`;
}

function renderJSONRow(label, value) {
  return renderRow(label, formatJSON(value), true);
}

function renderRoutes(list) {
  if (!routeTree) return;
  routeTree.innerHTML = "";
  const frag = document.createDocumentFragment();
  const walk = (nodes, depth) => {
    nodes.forEach((route) => {
      const el = document.createElement("div");
      el.className = "node";
      if (depth) el.style.paddingLeft = `${depth * 12}px`;
      const kind = document.createElement("span");
      kind.className = "kind";
      const dynamic = Array.isArray(route.params) && route.params.length > 0;
      kind.textContent = dynamic ? "dynamic" : "static";
      el.appendChild(kind);

      const name = document.createElement("span");
      name.className = "name";
      name.textContent = route.path || route.template || "/";
      el.appendChild(name);

      const meta = document.createElement("span");
      meta.className = "meta";
      if (dynamic) {
        const badge = document.createElement("span");
        const count = route.params.length;
        badge.textContent = `${count} param${count > 1 ? "s" : ""}`;
        meta.appendChild(badge);
      }
      const btn = document.createElement("button");
      btn.className = "route-action";
      btn.type = "button";
      btn.textContent = "Open";
      btn.title = `Navigate to ${route.path || route.template || "/"}`;
      btn.addEventListener("click", (evt) => {
        evt.stopPropagation();
        navigateRoute(route);
      });
      meta.appendChild(btn);
      el.appendChild(meta);

      el.dataset.path = (route.path || route.template || "/").toLowerCase();
      el.addEventListener("click", () => selectRoute(route));
      frag.appendChild(el);
      if (Array.isArray(route.children) && route.children.length) {
        walk(route.children, depth + 1);
      }
    });
  };
  walk(list, 0);
  routeTree.replaceChildren(frag);
  highlightSelectedRoute();
}

function highlightSelectedRoute() {
  if (!routeTree) return;
  const nodes = $$(".node", routeTree);
  nodes.forEach((n) => {
    if (selectedRoutePath) {
      n.classList.toggle(
        "active",
        n.dataset.path === selectedRoutePath.toLowerCase(),
      );
    } else {
      n.classList.remove("active");
    }
  });
}

function selectRoute(route) {
  selectedRoute = route;
  selectedRoutePath = route?.path || route?.template || "";
  highlightSelectedRoute();
  if (!routeTitle || !routeDetail || !routeDetailGo) return;
  if (!route) {
    routeTitle.textContent = "Select a route";
    routeDetail.innerHTML = "";
    routeDetailGo.setAttribute("disabled", "true");
    return;
  }
  routeTitle.textContent = route.path || route.template || "/";
  const rows = [];
  rows.push(renderRow("Path", route.path || "/"));
  if (route.template && route.template !== route.path)
    rows.push(renderRow("Template", route.template));
  if (Array.isArray(route.params) && route.params.length) {
    rows.push(
      renderRow(
        "Parameters",
        route.params.join(", ") || "-",
      ),
    );
  }
  rows.push(
    renderRow(
      "Children",
      Array.isArray(route.children) ? String(route.children.length) : "0",
    ),
  );
  routeDetail.innerHTML = rows.join("");
  routeDetailGo.removeAttribute("disabled");
}

function findRouteByPath(path, nodes) {
  for (const route of nodes) {
    const value = route.path || route.template || "";
    if (value === path) return route;
    if (Array.isArray(route.children)) {
      const found = findRouteByPath(path, route.children);
      if (found) return found;
    }
  }
  return null;
}

function refreshRoutes() {
  if (!routeTree) return;
  try {
    if (typeof globalThis.RFW_DEVTOOLS_ROUTES === "function") {
      const data = JSON.parse(globalThis.RFW_DEVTOOLS_ROUTES());
      routeSnapshot = Array.isArray(data) ? data : [];
      renderRoutes(routeSnapshot);
      if (selectedRoutePath) {
        const current = findRouteByPath(selectedRoutePath, routeSnapshot);
        selectRoute(current);
      }
      return;
    }
  } catch (err) {
    console.warn("DevTools routes refresh error", err);
  }
  routeSnapshot = [];
  if (routeTree) routeTree.textContent = "";
  selectRoute(null);
}

function navigateRoute(route) {
  if (!route) return;
  const dynamic = Array.isArray(route.params) && route.params.length > 0;
  if (dynamic) {
    openRoutePopup(route);
    return;
  }
  goToPath(route.path || route.template || "/");
}

function goToPath(path) {
  if (!path) return;
  if (typeof globalThis.goNavigate === "function") {
    globalThis.goNavigate(path);
  } else {
    window.location.assign(path);
  }
}

function openRoutePopup(route) {
  if (!routePopup || !routePopupFields) return;
  popupRoute = route;
  routePopup.classList.remove("hidden");
  if (routePopupTitle) routePopupTitle.textContent = route.path || route.template || "/";
  if (routePopupInfo) {
    const template = route.template && route.template !== route.path ? route.template : "Provide values for dynamic parameters.";
    routePopupInfo.textContent = template;
  }
  routePopupFields.innerHTML = "";
  (route.params || []).forEach((name) => {
    const field = document.createElement("label");
    field.className = "route-field";
    field.innerHTML = `<span>${escapeHTML(name)}</span>`;
    const input = document.createElement("input");
    input.type = "text";
    input.dataset.key = name;
    input.placeholder = name;
    field.appendChild(input);
    routePopupFields.appendChild(field);
  });
  if (routePopupError) routePopupError.textContent = "";
  const first = routePopupFields.querySelector("input");
  if (first) first.focus();
}

function closeRoutePopup() {
  if (!routePopup) return;
  routePopup.classList.add("hidden");
  popupRoute = null;
  routePopupFields?.replaceChildren();
  if (routePopupError) routePopupError.textContent = "";
}

function confirmRoutePopup() {
  if (!popupRoute || !routePopupFields) return;
  let target = popupRoute.path || popupRoute.template || "/";
  const inputs = Array.from(routePopupFields.querySelectorAll("input[data-key]"));
  for (const input of inputs) {
    const key = input.dataset.key;
    const value = input.value.trim();
    if (!value) {
      if (routePopupError) {
        routePopupError.textContent = `Missing value for ${key}`;
      }
      input.focus();
      return;
    }
    const encoded = encodeURIComponent(value);
    const pattern = new RegExp(`:${key}(?=/|$)`, "g");
    target = target.replace(pattern, encoded);
  }
  closeRoutePopup();
  goToPath(target);
}

$("#refreshRoutes")?.addEventListener("click", refreshRoutes);
routeFilter?.addEventListener("input", (e) => {
  const q = e.target.value.trim().toLowerCase();
  const nodes = $$(".node", routeTree);
  nodes.forEach((n) => {
    const text = n.textContent.toLowerCase();
    n.style.display = text.includes(q) ? "" : "none";
  });
});
routeDetailGo?.addEventListener("click", () => {
  const current = selectedRoute || findRouteByPath(selectedRoutePath, routeSnapshot);
  navigateRoute(current);
});
routePopupCancel?.addEventListener("click", closeRoutePopup);
routePopupConfirm?.addEventListener("click", confirmRoutePopup);

$("#treeFilter")?.addEventListener("input", (e) => {
  const q = e.target.value.trim().toLowerCase();
  const nodes = $$(".node", treeContainer);
  nodes.forEach((n) => {
    const text = n.textContent.toLowerCase();
    n.style.display = text.includes(q) ? "" : "none";
  });
});
$("#refreshTree")?.addEventListener("click", refreshTree);

function countNodes(list) {
  let total = 0;
  const walk = (nodes) => {
    nodes.forEach((n) => {
      total++;
      if (n.children) walk(n.children);
    });
  };
  walk(list);
  return total;
}

function aggregateRenderAverage(list) {
  let total = 0;
  let count = 0;
  const walk = (nodes) => {
    nodes.forEach((n) => {
      if (
        typeof n.total === "number" &&
        typeof n.updates === "number" &&
        n.updates > 0
      ) {
        total += n.total;
        count += n.updates;
      }
      if (Array.isArray(n.children) && n.children.length) walk(n.children);
    });
  };
  walk(list);
  if (!count) return null;
  return total / count;
}

function refreshTree() {
  try {
    if (typeof globalThis.RFW_DEVTOOLS_TREE === "function") {
      const data = JSON.parse(globalThis.RFW_DEVTOOLS_TREE());
      renderTree(data);
      const k = $("#kpiNodes");
      if (k) k.textContent = String(countNodes(data));
      const avg = aggregateRenderAverage(data);
      const r = $("#kpiRender");
      if (r) r.textContent = avg == null ? "n/a" : formatMs(avg);
      return;
    }
  } catch {}
  if (treeContainer) treeContainer.textContent = "";
  const k = $("#kpiNodes");
  if (k) k.textContent = "0";
  const r = $("#kpiRender");
  if (r) r.textContent = "n/a";
}

let fpsSample = 0,
  memSample = 0;
let frameCount = 0,
  lastTime = performance.now();
let renderSample = null;
let memFrom = null; // 'expvar' | 'heap' | null

function fpsLoop(t) {
  frameCount++;
  if (t - lastTime >= 1000) {
    fpsSample = frameCount;
    frameCount = 0;
    lastTime = t;
    const el = $("#kpiFps");
    if (el) el.textContent = String(Math.round(fpsSample));
  }
  requestAnimationFrame(fpsLoop);
}
requestAnimationFrame(fpsLoop);

async function pollMetrics() {
  let tickMem = 0;
  let tickFrom = null;

  try {
    const res = await fetch("/debug/vars", { cache: "no-store" });
    if (res.ok) {
      const vars = await res.json();
      const ms = vars.memstats || {};
      if (ms.Alloc) {
        tickMem = ms.Alloc / 1048576;
        tickFrom = "expvar";
      }
    }
  } catch {}

  if (!tickMem && performance && performance.memory) {
    try {
      tickMem = performance.memory.usedJSHeapSize / 1048576;
      tickFrom = "heap";
    } catch {}
  }

  if (tickFrom) {
    memSample = tickMem;
    memFrom = tickFrom;
  }

  const memEl = $("#kpiMem");
  if (memEl)
    memEl.textContent = memFrom ? Math.round(memSample) + " MB" : "n/a";
  const memSrcEl = $("#kpiMemSrc");
  if (memSrcEl) memSrcEl.textContent = `(${memFrom || "n/a"})`;

  const rEl = $("#kpiRender");
  if (rEl)
    rEl.textContent =
      renderSample != null ? Number(renderSample).toFixed(1) + " ms" : "n/a";

  setTimeout(pollMetrics, 1000);
}
pollMetrics();

const storeTree = $("#storeTree");
const storeTitle = $("#storeTitle");
const storeContent = $("#storeContent");

function buildStoreTree(obj, path = "") {
  return Object.entries(obj).map(([mod, stores]) => {
    const p = path ? `${path}/${mod}` : mod;
    const node = { name: mod, path: p, kind: "module" };
    node.children = Object.entries(stores).map(([st, keys]) => {
      const sp = `${p}/${st}`;
      return {
        name: st,
        path: sp,
        kind: "store",
        children: Object.entries(keys).map(([k, v]) => ({
          name: k,
          path: `${sp}/${k}`,
          kind: Array.isArray(v) ? "array" : typeof v,
          value: v,
        })),
      };
    });
    return node;
  });
}

function renderStoreTree(list, root = true) {
  const frag = document.createDocumentFragment();
  list.forEach((node) => {
    const el = document.createElement("div");
    el.className = "node";
    el.dataset.path = node.path.toLowerCase();
    el.innerHTML = `
      <span class="kind">${node.kind}</span>
      <span class="name">${node.name}</span>
      ${node.children ? "" : `<span class="time mono">${escapeHTML(String(node.value))}</span>`}
    `;
    el.addEventListener("click", () => selectStore(node));
    frag.appendChild(el);
    if (node.children) {
      const pad = document.createElement("div");
      pad.style.marginLeft = "18px";
      pad.appendChild(renderStoreTree(node.children, false));
      frag.appendChild(pad);
    }
  });
  if (root) {
    storeTree?.replaceChildren(frag);
    return storeTree;
  }
  return frag;
}

function selectStore(node) {
  if (!storeTitle || !storeContent) return;
  storeTitle.textContent = node.path;
  try {
    storeContent.textContent = JSON.stringify(node.value, null, 2);
  } catch {
    storeContent.textContent = String(node.value);
  }
}

function refreshStore() {
  if (!storeTree) return;
  storeTree.innerHTML = "";
  try {
    if (typeof globalThis.RFW_DEVTOOLS_STORES === "function") {
      const data = JSON.parse(globalThis.RFW_DEVTOOLS_STORES());
      renderStoreTree(buildStoreTree(data));
    }
  } catch {
    storeTree.textContent = "";
  }
}

$("#refreshStore")?.addEventListener("click", refreshStore);
$("#storeFilter")?.addEventListener("input", (e) => {
  const q = e.target.value.toLowerCase();
  $$(".node", storeTree).forEach((n) => {
    const text = n.textContent.toLowerCase();
    n.style.display = text.includes(q) ? "" : "none";
  });
});

const signalList = $("#signalList");
const signalTitle = $("#signalTitle");
const signalContent = $("#signalContent");

function renderSignalList(list) {
  if (!signalList) return;
  signalList.innerHTML = "";
  list.forEach((sig) => {
    const el = document.createElement("div");
    el.className = "node";
    el.dataset.id = String(sig.id);
    el.innerHTML = `
      <span class="name">#${sig.id}</span>
      <span class="time mono">${escapeHTML(String(sig.value))}</span>
    `;
    el.addEventListener("click", () => selectSignal(sig));
    signalList.appendChild(el);
  });
}

function selectSignal(sig) {
  if (!signalTitle || !signalContent) return;
  signalTitle.textContent = `Signal #${sig.id}`;
  try {
    signalContent.textContent = JSON.stringify(sig.value, null, 2);
  } catch {
    signalContent.textContent = String(sig.value);
  }
}

function refreshSignals() {
  if (!signalList) return;
  try {
    if (typeof globalThis.RFW_DEVTOOLS_SIGNALS === "function") {
      const data = JSON.parse(globalThis.RFW_DEVTOOLS_SIGNALS());
      const list = Object.entries(data).map(([id, v]) => ({ id, value: v }));
      renderSignalList(list);
    }
  } catch {
    signalList.textContent = "";
  }
}

$("#refreshSignals")?.addEventListener("click", refreshSignals);
$("#signalFilter")?.addEventListener("input", (e) => {
  const q = e.target.value.toLowerCase();
  $$(".node", signalList).forEach((n) => {
    n.style.display = n.textContent.toLowerCase().includes(q) ? "" : "none";
  });
});

const pluginTree = $("#pluginTree");
const pluginTitle = $("#pluginTitle");
const pluginContent = $("#pluginContent");

function renderPluginList(list) {
  if (!pluginTree) return;
  pluginTree.innerHTML = "";
  list.forEach((p) => {
    const el = document.createElement("div");
    el.className = "node";
    el.dataset.name = p.name.toLowerCase();
    el.innerHTML = `<span class="name">${p.name}</span>`;
    el.addEventListener("click", () => selectPlugin(p));
    pluginTree.appendChild(el);
  });
}

function selectPlugin(p) {
  if (!pluginTitle || !pluginContent) return;
  pluginTitle.textContent = p.name;
  try {
    pluginContent.textContent = JSON.stringify(p.config, null, 2);
  } catch {
    pluginContent.textContent = String(p.config);
  }
}

function refreshPlugins() {
  if (!pluginTree) return;
  try {
    if (typeof globalThis.RFW_DEVTOOLS_PLUGINS === "function") {
      const list = globalThis.RFW_DEVTOOLS_PLUGINS();
      renderPluginList(list);
    }
  } catch {
    pluginTree.textContent = "";
  }
}

$("#refreshPlugins")?.addEventListener("click", refreshPlugins);
$("#pluginFilter")?.addEventListener("input", (e) => {
  const q = e.target.value.toLowerCase();
  $$(".node", pluginTree).forEach((n) => {
    n.style.display = n.textContent.toLowerCase().includes(q) ? "" : "none";
  });
});

window.RFW_DEVTOOLS_REFRESH_STORES = refreshStore;
window.RFW_DEVTOOLS_REFRESH_SIGNALS = refreshSignals;
window.RFW_DEVTOOLS_REFRESH_PLUGINS = refreshPlugins;
window.RFW_DEVTOOLS_REFRESH_ROUTES = refreshRoutes;

const varsTree = $("#varsTree");
const varsTitle = $("#varsTitle");
const varsContent = $("#varsContent");
let varsData = [];

function buildVarsTree(obj, path = "") {
  return Object.entries(obj).map(([k, v]) => {
    const p = path ? `${path}.${k}` : k;
    const node = {
      name: k,
      path: p,
      kind: Array.isArray(v) ? "array" : typeof v,
      value: v,
    };
    if (v && typeof v === "object") node.children = buildVarsTree(v, p);
    return node;
  });
}

function renderVarsTree(list, root = true) {
  const frag = document.createDocumentFragment();
  list.forEach((node) => {
    const el = document.createElement("div");
    el.className = "node";
    el.dataset.path = node.path.toLowerCase();
    el.innerHTML = `
      <span class="kind">${node.kind}</span>
      <span class="name">${node.name}</span>
      ${node.children ? "" : `<span class="time mono">${escapeHTML(String(node.value))}</span>`}
    `;
    el.addEventListener("click", () => selectVar(node));
    frag.appendChild(el);
    if (node.children) {
      const pad = document.createElement("div");
      pad.style.marginLeft = "18px";
      pad.appendChild(renderVarsTree(node.children, false));
      frag.appendChild(pad);
    }
  });
  if (root) {
    varsTree?.replaceChildren(frag);
    return varsTree;
  }
  return frag;
}

function selectVar(node) {
  if (!varsTitle || !varsContent) return;
  varsTitle.textContent = node.path;
  try {
    varsContent.textContent = JSON.stringify(node.value, null, 2);
  } catch {
    varsContent.textContent = String(node.value);
  }
}

async function loadVars() {
  if (!varsTree) return;
  varsTree.innerHTML = "";
  try {
    const res = await fetch("/debug/vars", { cache: "no-store" });
    if (!res.ok) {
      varsTree.textContent = "Failed to load";
      return;
    }
    const data = await res.json();
    varsData = buildVarsTree(data);
    renderVarsTree(varsData);
  } catch {
    varsTree.textContent = "Error loading vars";
  }
}

$("#varsFilter")?.addEventListener("input", (e) => {
  const q = e.target.value.toLowerCase();
  $$(".node", varsTree).forEach((n) => {
    const text = n.textContent.toLowerCase();
    n.style.display = text.includes(q) ? "" : "none";
  });
});

const pprofLinks = $("#pprofLinks");
const pprofTitle = $("#pprofTitle");
const pprofContent = $("#pprofContent");

async function loadPprof() {
  if (!pprofLinks) return;
  pprofLinks.innerHTML = "";
  pprofContent.textContent = "";
  pprofTitle.textContent = "Select a profile";
  try {
    const res = await fetch("/debug/pprof/", { cache: "no-store" });
    if (!res.ok) {
      pprofLinks.textContent = "Failed to load";
      return;
    }
    const html = await res.text();
    const doc = new DOMParser().parseFromString(html, "text/html");
    const anchors = Array.from(doc.querySelectorAll("a[href]"));
    anchors.forEach((a) => {
      const href = a.getAttribute("href");
      if (!href || href.startsWith("?")) return;
      const el = document.createElement("div");
      el.className = "node";
      el.innerHTML = `<span class="name">${a.textContent || href}</span>`;
      el.addEventListener("click", () => loadProfile(href));
      pprofLinks.appendChild(el);
    });
  } catch {
    pprofLinks.textContent = "Error loading profiles";
  }
}

async function loadProfile(name) {
  if (!pprofContent || !pprofTitle) return;
  pprofTitle.textContent = name;
  pprofContent.textContent = "Loading...";
  try {
    const hasQuery = name.includes("?");
    const url = `/debug/pprof/${name}${hasQuery ? "&" : "?"}debug=1`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) {
      pprofContent.textContent = "Failed to load";
      return;
    }
    const type = res.headers.get("Content-Type") || "";
    if (type.includes("text")) {
      const text = await res.text();
      pprofContent.textContent = text;
    } else {
      const blob = await res.blob();
      const dl = URL.createObjectURL(blob);
      pprofContent.innerHTML = `<a href="${dl}" download="${name.replace(/\?.*/, "")}">Download profile</a>`;
    }
  } catch {
    pprofContent.textContent = "Error fetching profile";
  }
}

$("#pprofFilter")?.addEventListener("input", (e) => {
  const q = e.target.value.toLowerCase();
  $$(".node", pprofLinks).forEach((n) => {
    const text = n.textContent.toLowerCase();
    n.style.display = text.includes(q) ? "" : "none";
  });
});

const logList = $("#logList");
const netList = $("#netList");
const original = {
  log: console.log.bind(console),
  warn: console.warn.bind(console),
  error: console.error.bind(console),
  info: console.info.bind(console),
};

function addLog(level, args) {
  if (!logList) return;
  const msg = args
    .map((a) => {
      try {
        return typeof a === "string" ? a : JSON.stringify(a);
      } catch {
        return String(a);
      }
    })
    .join(" ");
  if (msg.includes("mutation:")) return;
  const time = new Date().toLocaleTimeString();
  const row = document.createElement("div");
  row.className = "log";
  row.dataset.lvl = level;
  row.innerHTML = `<div class="mono">${time}</div><div class="lvl">${level.toUpperCase()}</div><div class="msg">${escapeHTML(msg)}</div>`;
  logList.appendChild(row);
  if (logList.childElementCount > 200) logList.removeChild(logList.firstChild);
  logList.scrollTop = logList.scrollHeight;
}

console.log = (...a) => {
  addLog("info", a);
  original.log(...a);
};
console.warn = (...a) => {
  addLog("warn", a);
  original.warn(...a);
};
console.error = (...a) => {
  addLog("error", a);
  original.error(...a);
};
console.info = (...a) => {
  addLog("info", a);
  original.info(...a);
};

$("#clearLogs")?.addEventListener("click", () => {
  if (logList) logList.innerHTML = "";
});
$("#logFilter")?.addEventListener("input", (e) => {
  if (!logList) return;
  const q = e.target.value.toLowerCase();
  $$(".log", logList).forEach((row) => {
    row.style.display = row.textContent.toLowerCase().includes(q) ? "" : "none";
  });
});

function addRequest(url, status, dur) {
  if (!netList) return;
  const row = document.createElement("div");
  row.className = "net";
  row.innerHTML = `<div class="url">${escapeHTML(url)}</div><div class="mono">${status}</div><div class="mono">${Number(dur).toFixed(1)} ms</div>`;
  netList.appendChild(row);
  if (netList.childElementCount > 200) netList.removeChild(netList.firstChild);
  netList.scrollTop = netList.scrollHeight;
}

$("#clearNet")?.addEventListener("click", () => {
  if (netList) netList.innerHTML = "";
});

window.RFW_DEVTOOLS = {
  open: openDevtools,
  close: closeDevtools,
  feedMetrics(m) {
    if (m.fps != null) {
      fpsSample = m.fps;
      const el = $("#kpiFps");
      if (el) el.textContent = String(Math.round(fpsSample));
    }
    if (m.mem != null) {
      memSample = m.mem;
      memFrom = "expvar";
      const el = $("#kpiMem");
      if (el) el.textContent = Math.round(memSample) + " MB";
      const src = $("#kpiMemSrc");
      if (src) src.textContent = "(expvar)";
    }
    if (m.render != null) {
      renderSample = m.render;
      const el = $("#kpiRender");
      if (el) el.textContent = Number(m.render).toFixed(1) + " ms";
    }
  },
  feedTree(t) {
    if (Array.isArray(t)) {
      renderTree(t);
      const k = $("#kpiNodes");
      if (k) k.textContent = String(countNodes(t));
    }
  },
  network(start, url, status, dur) {
    if (!start) addRequest(url, status, dur);
  },
  log(level, ...args) {
    addLog(level || "info", args);
  },
};
window.RFW_DEVTOOLS_REFRESH = refreshTree;
