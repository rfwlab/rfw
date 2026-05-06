//go:build js && wasm

package core

import (
	"fmt"
	"runtime/debug"
	"strings"

	js "github.com/rfwlab/rfw/v2/js"
)

var globalOverlay = &errorOverlay{}

// ShowErrorOverlay displays a styled error recovery UI in the browser when a
// panic occurs. It categorizes the error, shows the Go stack trace, and
// provides actionable hints based on the panic message.
func ShowErrorOverlay(err any, context string) {
	globalOverlay.show(err, context)
}

type errorOverlay struct {
	shown     bool
	container js.Value
	errCount  int
}

func (eo *errorOverlay) show(err any, context string) {
	errStr := fmt.Sprintf("%v", err)
	goStack := string(debug.Stack())

	if !eo.shown {
		eo.shown = true
		eo.createContainer(errStr, goStack, context)
		return
	}

	eo.errCount++
	doc := js.Document()
	list := doc.Call("getElementById", "rfw-error-list")
	if !list.Truthy() {
		return
	}
	item := doc.Call("createElement", "div")
	item.Get("style").Set("borderTop", "1px solid #e5e7eb")
	item.Set("innerHTML", eo.buildErrorItem(eo.errCount, errStr, goStack, context))
	list.Call("appendChild", item)
}

func (eo *errorOverlay) createContainer(errStr, goStack, context string) {
	doc := js.Document()
	body := doc.Get("body")

	overlay := doc.Call("createElement", "div")
	overlay.Set("id", "rfw-error-overlay")
	style := overlay.Get("style")
	style.Set("position", "fixed")
	style.Set("top", "0")
	style.Set("left", "0")
	style.Set("width", "100%")
	style.Set("height", "100%")
	style.Set("backgroundColor", "rgba(0,0,0,0.85)")
	style.Set("zIndex", "999999")
	style.Set("display", "flex")
	style.Set("alignItems", "center")
	style.Set("justifyContent", "center")
	style.Set("padding", "20px")
	style.Set("boxSizing", "border-box")
	style.Set("fontFamily", "system-ui,-apple-system,sans-serif")

	card := doc.Call("createElement", "div")
	cs := card.Get("style")
	cs.Set("background", "#ffffff")
	cs.Set("borderRadius", "12px")
	cs.Set("boxShadow", "0 20px 60px rgba(0,0,0,0.3)")
	cs.Set("maxWidth", "900px")
	cs.Set("width", "100%")
	cs.Set("maxHeight", "90vh")
	cs.Set("overflow", "auto")
	cs.Set("display", "flex")
	cs.Set("flexDirection", "column")

	card.Set("innerHTML", eo.buildMainHTML(errStr, goStack, context))
	overlay.Call("appendChild", card)
	body.Call("appendChild", overlay)

	eo.bindActions(overlay)
	eo.container = overlay
}

func (eo *errorOverlay) bindActions(overlay js.Value) {
	doc := js.Document()

	reload := doc.Call("getElementById", "rfw-error-reload")
	if reload.Truthy() {
		reload.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			js.Location().Call("reload")
			return nil
		}))
	}

	copyBtn := doc.Call("getElementById", "rfw-error-copy")
	if copyBtn.Truthy() {
		copyBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			pre := doc.Call("getElementById", "rfw-error-full")
			if pre.Truthy() {
				text := pre.Get("textContent").String()
				navigator := js.Global().Get("navigator")
				if clipboard := navigator.Get("clipboard"); clipboard.Truthy() {
					clipboard.Call("writeText", text)
				}
			}
			return nil
		}))
	}
}

func (eo *errorOverlay) buildMainHTML(errStr, goStack, context string) string {
	cat := eo.categorize(errStr)
	hint := eo.hintHTML(errStr, context)

	versionStr := Version
	if versionStr == "" {
		versionStr = "dev"
	}

	return fmt.Sprintf(`
<div style="padding:24px 24px 0;">
    <div style="display:flex;align-items:flex-start;gap:12px;margin-bottom:16px;">
        <div style="flex:1;">
            <div style="font-size:11px;text-transform:uppercase;letter-spacing:0.08em;color:#6b7280;font-weight:700;">%s</div>
            <h2 style="margin:4px 0 0;font-size:20px;color:#111827;font-weight:800;">Something went wrong</h2>
        </div>
    </div>
    <div style="background:#fef2f2;border-left:4px solid #ef4444;border-radius:6px;padding:16px;margin-bottom:16px;">
        <div style="font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:13px;color:#991b1b;word-break:break-word;line-height:1.5;">%s</div>
    </div>
</div>
%s
<div style="padding:0 24px;">
    <div style="display:flex;gap:8px;flex-wrap:wrap;margin-bottom:16px;">
        <button id="rfw-error-reload" style="background:#111827;color:#fff;border:none;border-radius:6px;padding:10px 18px;cursor:pointer;font-size:14px;font-weight:500;">Reload Page</button>
        <button id="rfw-error-copy" style="background:#f3f4f6;color:#374151;border:none;border-radius:6px;padding:10px 18px;cursor:pointer;font-size:14px;font-weight:500;">Copy Error</button>
    </div>
</div>
<div style="padding:0 24px 16px;">
    <div style="font-size:13px;font-weight:700;color:#374151;margin-bottom:8px;">Stack Trace</div>
    <pre style="background:#f9fafb;border:1px solid #e5e7eb;border-radius:6px;padding:12px;overflow-x:auto;font-size:11px;color:#4b5563;margin:0;line-height:1.5;">%s</pre>
</div>
<pre id="rfw-error-full" style="display:none;">%s</pre>
<div id="rfw-error-list" style="display:none;"></div>
<div style="padding:8px 24px 16px;text-align:center;">
    <div style="font-size:11px;color:#9ca3af;">
        rfw recovery mode &middot; %s
    </div>
</div>
    `, cat, htmlEscape(errStr), hint,
		htmlEscape(goStack),
		htmlEscape(fmt.Sprintf("Error: %s\nContext: %s\n\n%s", errStr, context, goStack)),
		versionStr)
}

func (eo *errorOverlay) buildErrorItem(n int, errStr, goStack, context string) string {
	return fmt.Sprintf(`
<div style="padding:16px;">
    <div style="font-size:12px;font-weight:700;color:#6b7280;margin-bottom:8px;">Error #%d</div>
    <div style="background:#fef2f2;border-radius:6px;padding:12px;margin-bottom:8px;">
        <div style="font-family:monospace;font-size:12px;color:#991b1b;word-break:break-word;">%s</div>
    </div>
    <details open>
        <summary style="cursor:pointer;font-size:12px;color:#6b7280;">Stack trace</summary>
        <pre style="background:#f9fafb;border-radius:6px;padding:8px;font-size:11px;color:#4b5563;margin-top:8px;">%s</pre>
    </details>
</div>
    `, n, htmlEscape(errStr), htmlEscape(goStack))
}

func (eo *errorOverlay) categorize(err string) string {
	errLower := strings.ToLower(err)
	switch {
	case strings.Contains(errLower, "template"):
		return "Template Error"
	case strings.Contains(errLower, "signal"):
		return "Signal Error"
	case strings.Contains(errLower, "store"):
		return "Store Error"
	case strings.Contains(errLower, "nil pointer") || strings.Contains(errLower, "invalid memory"):
		return "Null Reference"
	case strings.Contains(errLower, "index out of range"):
		return "Index Error"
	case strings.Contains(errLower, "mount") || strings.Contains(errLower, "render") || strings.Contains(errLower, "unmount"):
		return "Lifecycle Error"
	case strings.Contains(errLower, "dom") || strings.Contains(errLower, "element"):
		return "DOM Error"
	default:
		return "Runtime Error"
	}
}

func (eo *errorOverlay) hintHTML(errStr, context string) string {
	errLower := strings.ToLower(errStr)
	hints := []string{}

	if strings.Contains(errLower, "template") && strings.Contains(errLower, "not found") {
		hints = append(hints, `Call <code style="background:#f3f4f6;padding:2px 5px;border-radius:3px;font-size:12px;">composition.RegisterFS(&amp;yourEmbedFS)</code> or add a <code style="background:#f3f4f6;padding:2px 5px;border-radius:3px;font-size:12px;">Template() string</code> method to your struct.`)
	}
	if strings.Contains(errLower, "signal") && strings.Contains(errLower, "not found") {
		hints = append(hints, `Use a signal type field (<code style="background:#f3f4f6;padding:2px 5px;border-radius:3px;font-size:12px;">t.Int</code>, <code style="background:#f3f4f6;padding:2px 5px;border-radius:3px;font-size:12px;">*t.String</code>, etc.) and initialize with <code style="background:#f3f4f6;padding:2px 5px;border-radius:3px;font-size:12px;">t.NewInt(0)</code>.`)
	}
	if strings.Contains(errLower, "nil pointer") || strings.Contains(errLower, "invalid memory") {
		hints = append(hints, `Initialize all pointer fields. Use <code style="background:#f3f4f6;padding:2px 5px;border-radius:3px;font-size:12px;">*t.Inject[T]</code> for DI dependencies.`)
	}
	if strings.Contains(errLower, "store") && strings.Contains(errLower, "not found") {
		hints = append(hints, `Register store with <code style="background:#f3f4f6;padding:2px 5px;border-radius:3px;font-size:12px;">state.GlobalStoreManager.RegisterStore()</code>.`)
	}
	if strings.Contains(errLower, "index out of range") {
		hints = append(hints, `Check bounds: <code style="background:#f3f4f6;padding:2px 5px;border-radius:3px;font-size:12px;">if len(items) > i { ... }</code>.`)
	}
	if strings.Contains(errLower, "dom") || strings.Contains(errLower, "element") {
		hints = append(hints, `Ensure element exists before access. Check component mount order.`)
	}

	if len(hints) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<div style="padding:0 24px 16px;"><div style="background:#eff6ff;border-radius:6px;padding:16px;">`)
	sb.WriteString(`<div style="font-size:13px;font-weight:700;color:#1e40af;margin-bottom:8px;">How to fix this</div>`)
	sb.WriteString(`<ul style="margin:0;padding-left:20px;font-size:13px;color:#374151;line-height:1.7;">`)
	for _, h := range hints {
		sb.WriteString(fmt.Sprintf("<li>%s</li>", h))
	}
	sb.WriteString("</ul></div></div>")
	return sb.String()
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}