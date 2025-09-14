//go:build devtools && js && wasm

package devtools

import (
	"fmt"
	"strings"

	"github.com/rfwlab/rfw/v1/composition"
	"github.com/rfwlab/rfw/v1/dom"
	js "github.com/rfwlab/rfw/v1/js"
)

const errorBoxCSS = `/* === Error Box === */
.rfw-errorbox{
    position:fixed;left:24px;right:24px;bottom:24px;max-width:860px;margin:0 auto;
    background:linear-gradient(180deg,rgba(20,11,13,.8),rgba(13,7,9,.95));color:var(--text);
    border:1px solid var(--border);border-radius:var(--round);box-shadow:var(--shadow);
    z-index:2147483650;overflow:hidden;backdrop-filter:blur(10px);
}
.rfw-eb-top{display:flex;align-items:center;justify-content:space-between;padding:10px 12px;border-bottom:1px solid var(--border)}
.rfw-eb-nav{display:flex;align-items:center;gap:8px}
.rfw-eb-iconbtn{display:inline-flex;align-items:center;justify-content:center;height:30px;padding:0 10px;border:1px solid var(--border);border-radius:10px;background:var(--chip-bg);color:var(--rose-200);cursor:pointer}
.rfw-eb-iconbtn:disabled{opacity:.45;cursor:default}
.rfw-eb-count{color:var(--rose-300)}
.rfw-eb-close{border:0;background:transparent;color:var(--rose-200);font-size:18px;cursor:pointer}
.rfw-eb-title{margin:14px 16px 6px;font-weight:700;color:var(--text)}
.rfw-eb-msg{margin:0 16px 12px;color:var(--accent)}
.rfw-eb-frame{margin:0 16px 16px;border:1px solid var(--tile-border);border-radius:12px;overflow:hidden;background:var(--chip-bg)}
.rfw-eb-framehead{display:flex;align-items:center;justify-content:space-between;padding:8px 10px;background:var(--tile-bg);border-bottom:1px solid var(--tile-border)}
.rfw-eb-code{max-height:50vh;overflow:auto}
.rfw-eb-row{display:flex;gap:12px;padding:0 12px}
.rfw-eb-gutter{width:ch;text-align:right;opacity:.55;user-select:none;color:var(--rose-300)}
.rfw-eb-pre{margin:0;white-space:pre;color:var(--rose-50)}
.rfw-eb-hl{background:rgba(255,77,79,.10)}
.rfw-eb-actions{display:flex;align-items:center;gap:8px;padding:10px 12px;border-top:1px solid var(--border)}
#ebCopy,#ebReload{padding:8px 12px;border:1px solid var(--border);border-radius:12px;background:var(--chip-bg);color:var(--rose-200);cursor:pointer}
@media (max-width:640px){.rfw-errorbox{left:12px;right:12px;bottom:12px}}
`

var (
	errEvtFn js.Func
	rejEvtFn js.Func
	renderFn func()
)

func init() {
	setupErrorListeners()

	doc := js.Document()
	if doc.Get("readyState").String() == "loading" {
		var readyFn js.Func
		readyFn = js.FuncOf(func(this js.Value, args []js.Value) any {
			readyFn.Release()
			setupErrorBox()
			return nil
		})
		doc.Call("addEventListener", "DOMContentLoaded", readyFn)
	} else {
		setupErrorBox()
	}
}

func setupErrorBox() {
	doc := dom.Doc()
	style := doc.CreateElement("style")
	style.SetText(errorBoxCSS)
	doc.Head().AppendChild(style)

	box := composition.Div().Classes("rfw-errorbox", "hidden")
	boxEl := box.Element()
	boxEl.SetAttr("id", "rfwErrorBox")
	boxEl.SetAttr("role", "dialog")
	boxEl.SetAttr("aria-modal", "true")
	boxEl.SetAttr("aria-label", "Unhandled Runtime Error")
	boxEl.SetAttr("data-rfw-ignore", "")

	prevBtn := composition.Button().Classes("rfw-eb-iconbtn")
	prevBtnEl := prevBtn.Element()
	prevBtnEl.SetAttr("id", "ebPrev")
	prevBtnEl.SetAttr("title", "Previous")
	prevBtnEl.SetText("←")

	nextBtn := composition.Button().Classes("rfw-eb-iconbtn")
	nextBtnEl := nextBtn.Element()
	nextBtnEl.SetAttr("id", "ebNext")
	nextBtnEl.SetAttr("title", "Next")
	nextBtnEl.SetText("→")

	countSpan := composition.Span().Classes("rfw-eb-count")
	countEl := countSpan.Element()
	countEl.SetAttr("id", "ebCount")
	countEl.SetText("1 of 1 unhandled error")

	nav := composition.Div().Classes("rfw-eb-nav")
	nav.Element().AppendChild(prevBtnEl)
	nav.Element().AppendChild(nextBtnEl)
	nav.Element().AppendChild(countEl)

	closeBtn := composition.Button().Classes("rfw-eb-close")
	closeBtnEl := closeBtn.Element()
	closeBtnEl.SetAttr("id", "ebClose")
	closeBtnEl.SetAttr("title", "Close")
	closeBtnEl.SetText("✕")

	top := composition.Div().Classes("rfw-eb-top")
	top.Element().AppendChild(nav.Element())
	top.Element().AppendChild(closeBtnEl)

	title := composition.H(2).Classes("rfw-eb-title")
	titleEl := title.Element()
	titleEl.SetText("Unhandled Runtime Error")

	msg := composition.Div().Classes("rfw-eb-msg", "mono")
	msgEl := msg.Element()
	msgEl.SetAttr("id", "ebMsg")
	msgEl.SetText("Error: …")

	pathSpan := composition.Span().Classes("mono")
	pathEl := pathSpan.Element()
	pathEl.SetAttr("id", "ebFramePath")
	pathEl.SetText("-")

	frameHead := composition.Div().Classes("rfw-eb-framehead")
	frameHead.Element().AppendChild(pathEl)

	codeDiv := composition.Div().Classes("rfw-eb-code")
	codeEl := codeDiv.Element()
	codeEl.SetAttr("id", "ebCode")
	codeEl.SetAttr("aria-live", "polite")

	frame := composition.Div().Classes("rfw-eb-frame")
	frame.Element().AppendChild(frameHead.Element())
	frame.Element().AppendChild(codeEl)

	copyBtn := composition.Button().Classes("rfw-button")
	copyBtnEl := copyBtn.Element()
	copyBtnEl.SetAttr("id", "ebCopy")
	copyBtnEl.SetText("Copy error")

	spacer := composition.Span().Classes("rfw-spacer")
	spacerEl := spacer.Element()

	reloadBtn := composition.Button().Classes("rfw-button")
	reloadBtnEl := reloadBtn.Element()
	reloadBtnEl.SetAttr("id", "ebReload")
	reloadBtnEl.SetText("Reload")

	actions := composition.Div().Classes("rfw-eb-actions")
	actions.Element().AppendChild(copyBtnEl)
	actions.Element().AppendChild(spacerEl)
	actions.Element().AppendChild(reloadBtnEl)

	boxEl.AppendChild(top.Element())
	boxEl.AppendChild(titleEl)
	boxEl.AppendChild(msgEl)
	boxEl.AppendChild(frame.Element())
	boxEl.AppendChild(actions.Element())

	doc.Body().AppendChild(boxEl)

	render := func() {
		cur, ok := currentRuntimeError()
		if !ok {
			boxEl.AddClass("hidden")
			return
		}
		boxEl.RemoveClass("hidden")
		msgEl.SetText(cur.Message)
		codeEl.SetText(cur.Stack)
		pathEl.SetText(cur.Path)
		countEl.SetText(fmt.Sprintf("%d of %d unhandled error", runtimeErrorIndex()+1, runtimeErrorCount()))
		if runtimeErrorIndex() <= 0 {
			prevBtnEl.SetAttr("disabled", "true")
		} else {
			prevBtnEl.Call("removeAttribute", "disabled")
		}
		if runtimeErrorIndex() >= runtimeErrorCount()-1 {
			nextBtnEl.SetAttr("disabled", "true")
		} else {
			nextBtnEl.Call("removeAttribute", "disabled")
		}
	}

	prevBtnEl.On("click", func(dom.Event) { prevRuntimeError(); render() })
	nextBtnEl.On("click", func(dom.Event) { nextRuntimeError(); render() })
	closeBtnEl.On("click", func(dom.Event) { resetRuntimeErrors(); render() })
	copyBtnEl.On("click", func(dom.Event) {
		cur, ok := currentRuntimeError()
		if ok {
			js.Window().Get("navigator").Get("clipboard").Call("writeText", cur.Message+"\n"+cur.Stack)
		}
	})
	reloadBtnEl.On("click", func(dom.Event) { js.Location().Call("reload") })

	renderFn = render
	render()
}

func setupErrorListeners() {
	errEvtFn = js.FuncOf(func(this js.Value, args []js.Value) any {
		e := args[0]
		msg := e.Get("message").String()
		stack := ""
		if v := e.Get("error"); v.Type() == js.TypeObject {
			stack = v.Get("stack").String()
		}
		path := parsePath(stack)
		addRuntimeError(runtimeError{Message: msg, Stack: stack, Path: path})
		if renderFn != nil {
			renderFn()
		}
		return nil
	})
	rejEvtFn = js.FuncOf(func(this js.Value, args []js.Value) any {
		e := args[0]
		reason := e.Get("reason")
		msg := reason.Get("message").String()
		stack := reason.Get("stack").String()
		path := parsePath(stack)
		addRuntimeError(runtimeError{Message: msg, Stack: stack, Path: path})
		if renderFn != nil {
			renderFn()
		}
		return nil
	})
	js.Window().Call("addEventListener", "error", errEvtFn)
	js.Window().Call("addEventListener", "unhandledrejection", rejEvtFn)
}

func parsePath(stack string) string {
	lines := strings.Split(stack, "\n")
	for _, l := range lines {
		if strings.Contains(l, ".go") {
			l = strings.TrimSpace(l)
			return l
		}
	}
	return ""
}
