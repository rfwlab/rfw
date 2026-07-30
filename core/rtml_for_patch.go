//go:build js && wasm

package core

import (
	"fmt"
	"strings"

	"github.com/rfwlab/rfw/v2/dom"
)

// patchForLoop replaces the rows of one loop in place. A store-driven list used
// to re-render its whole component (every dependency, every binding, the routed
// page below an app shell included) to repaint a handful of rows; here only the
// nodes carrying the loop id are touched.
//
// It reports false when the loop cannot be patched on its own, and the caller
// falls back to the full render: a body that pulls in other components or opens
// its own conditional needs the whole pipeline, which only a render provides.
func patchForLoop(c *HTMLComponent, loopID string, aliases []string, loopContent string, collection any) bool {
	if !incrementalForBody(loopContent) {
		return false
	}
	root := dom.ComponentRoot(c.ID)
	if root.IsNull() || root.IsUndefined() {
		return false
	}
	anchor := root.Query(fmt.Sprintf(`template[data-for-anchor="%s"]`, loopID))
	if anchor.IsNull() || anchor.IsUndefined() {
		return false
	}

	rows, ok := expandForRows(c, aliases, loopContent, collection, loopID)
	if !ok {
		return false
	}
	if strings.Contains(rows, "@include:") {
		// an item resolved to a component: only a render can mount it
		return false
	}

	old := root.QueryAll(fmt.Sprintf(`[data-for="%s"]`, loopID))
	for i := old.Length() - 1; i >= 0; i-- {
		old.Index(i).Call("remove")
	}

	if rows != "" {
		markup := c.renderRowFragment(rows)
		anchor.Call("insertAdjacentHTML", "afterend", markup)
	}
	dom.ReleaseInputBindings(c.ID)
	dom.BindStoreInputsForComponent(c.ID, root.Value)
	dom.BindSignalInputs(c.ID, root.Value)
	dom.BindASTStoreInputs(c.ID, root.Value)
	dom.BindASTSignalInputs(c.ID, root.Value)

	if dom.TemplateHook != nil {
		dom.TemplateHook(c.ID, rows)
	}
	return true
}

// incrementalForBody reports whether a loop body is self-contained enough to be
// patched without a full render.
func incrementalForBody(body string) bool {
	for _, directive := range []string{"@include:", "@if:", "@for:", "@slot", "rt-is="} {
		if strings.Contains(body, directive) {
			return false
		}
	}
	return singleRootRow(body)
}

// renderRowFragment runs the substitutions that normally follow the loop
// expansion over freshly built rows, so a patched row carries the same
// bindings a rendered one would.
func (c *HTMLComponent) renderRowFragment(fragment string) string {
	fragment = replaceStorePlaceholders(fragment, c)
	fragment = replaceSignalPlaceholders(fragment, c)
	fragment = replaceExprInClassAttr(fragment, c)
	fragment = replaceExprPlaceholders(fragment, c)
	fragment = replacePropPlaceholders(fragment, c)
	fragment = replacePluginPlaceholders(fragment)
	fragment = replaceEventHandlers(fragment)
	fragment = replaceConstructors(fragment)
	return minifyInline(fragment)
}
