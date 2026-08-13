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
	root := dom.ComponentRoot(c.ID)
	if root.IsNull() || root.IsUndefined() {
		return true
	}
	anchor := root.Query(fmt.Sprintf(`template[data-for-anchor="%s"]`, loopID))
	if anchor.IsNull() || anchor.IsUndefined() {
		return loopInHiddenConditionalBranch(c, loopID)
	}
	if !incrementalForBody(loopContent) {
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

	markup := ""
	if rows != "" {
		markup = c.renderRowFragment(rows)
	}
	if !reconcileForRows(anchor, loopID, markup) {
		return false
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

func loopInHiddenConditionalBranch(c *HTMLComponent, loopID string) bool {
	marker := fmt.Sprintf(`data-for-anchor="%s"`, loopID)
	for _, content := range c.conditionContents {
		belongs := false
		for _, branch := range content.Branches {
			if strings.Contains(branch.Content, marker) {
				belongs = true
				break
			}
		}
		if !belongs {
			continue
		}
		for _, branch := range content.Branches {
			if branch.Condition == "" {
				return !strings.Contains(branch.Content, marker)
			}
			matched, _ := evaluateCondition(branch.Condition, c)
			if matched {
				return !strings.Contains(branch.Content, marker)
			}
		}
		return true
	}
	return false
}

func reconcileForRows(anchor dom.Element, loopID, markup string) bool {
	selector := fmt.Sprintf(`[data-for="%s"]`, loopID)
	oldRows := dom.Element{Value: anchor.Get("parentNode").Call("querySelectorAll", selector)}
	template := dom.CreateElement("template")
	template.SetHTML(markup)
	newRows := dom.Element{Value: template.Get("content").Call("querySelectorAll", selector)}

	oldByKey := make(map[string]dom.Element, oldRows.Length())
	for i := 0; i < oldRows.Length(); i++ {
		row := oldRows.Index(i)
		key := row.Attr("data-key")
		if key == "" {
			return false
		}
		if _, duplicate := oldByKey[key]; duplicate {
			return false
		}
		oldByKey[key] = row
	}

	keys := make([]string, newRows.Length())
	seen := make(map[string]struct{}, newRows.Length())
	for i := 0; i < newRows.Length(); i++ {
		key := newRows.Index(i).Attr("data-key")
		if key == "" {
			return false
		}
		if _, duplicate := seen[key]; duplicate {
			return false
		}
		seen[key] = struct{}{}
		keys[i] = key
	}

	parent := dom.Element{Value: anchor.Get("parentNode")}
	cursor := anchor.Get("nextSibling")
	for i, key := range keys {
		newRow := newRows.Index(i)
		row, exists := oldByKey[key]
		if exists {
			dom.PatchElement(row, newRow)
			delete(oldByKey, key)
		} else {
			row = dom.Element{Value: newRow.Call("cloneNode", true)}
		}
		if !row.Equal(cursor) {
			parent.Call("insertBefore", row.Value, cursor)
		} else {
			cursor = cursor.Get("nextSibling")
		}
	}
	for _, row := range oldByKey {
		row.Call("remove")
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
