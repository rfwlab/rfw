//go:build js && wasm

package dom

import (
	"fmt"
	"strings"

	js "github.com/rfwlab/rfw/v2/js"
)

const renderedAttrsProperty = "__rfwRenderedAttrs"

// patchPlan separates reconciliation decisions from DOM mutation. Planning is
// read-only: invalid identities or ownership conflicts are reported before the
// first operation is committed.
type patchPlan struct {
	ownerID string
	ops     []func()
}

type childPlacement struct {
	existing js.Value
	source   js.Value
}

type liveFormState struct {
	value         string
	checked       bool
	selectedIndex int
	selectionAt   int
	selectionEnd  int
	selectionDir  string
	hasValue      bool
	hasChecked    bool
	hasSelected   bool
	hasCaret      bool
}

type activeFormState struct {
	element js.Value
	state   *liveFormState
}

func captureActiveFormState(root js.Value) activeFormState {
	active := js.Global().Get("document").Get("activeElement")
	if !active.Truthy() || !root.Call("contains", active).Bool() {
		return activeFormState{}
	}
	return activeFormState{element: active, state: captureLiveFormState(active, active)}
}

func (snapshot activeFormState) restore() {
	if snapshot.state == nil || !snapshot.element.Truthy() || !snapshot.element.Get("isConnected").Bool() {
		return
	}
	active := js.Global().Get("document").Get("activeElement")
	if !active.Truthy() || !active.Equal(snapshot.element) {
		return
	}
	restoreLiveFormState(snapshot.element, *snapshot.state)
}

func patchInnerHTML(element js.Value, html string) {
	template := CreateElement("template")
	template.Set("innerHTML", html)
	newContent := template.Get("content")

	ownerID := attribute(element, "data-component-id")
	plan := &patchPlan{ownerID: ownerID}
	firstChild := newContent.Get("firstChild")
	if firstChild.Truthy() && firstChild.Get("nodeName").String() == "ROOT" &&
		ownerID != "" && attribute(firstChild, "data-component-id") == ownerID {
		if err := plan.planNode(element, firstChild); err != nil {
			panic(err)
		}
	} else if err := plan.planChildren(element, newContent); err != nil {
		panic(err)
	}
	plan.commit()
}

func (plan *patchPlan) commit() {
	for _, operation := range plan.ops {
		operation()
	}
}

func (plan *patchPlan) planNode(existing, replacement js.Value) error {
	if existing.Get("nodeType").Int() != replacement.Get("nodeType").Int() ||
		existing.Get("nodeName").String() != replacement.Get("nodeName").String() {
		return fmt.Errorf("rfw DOM patch: incompatible nodes %s and %s", existing.Get("nodeName").String(), replacement.Get("nodeName").String())
	}

	nodeType := replacement.Get("nodeType").Int()
	if nodeType == 3 || nodeType == 8 { // text or comment
		oldValue := existing.Get("nodeValue").String()
		newValue := replacement.Get("nodeValue").String()
		if oldValue != newValue {
			plan.ops = append(plan.ops, func() { existing.Set("nodeValue", newValue) })
		}
		return nil
	}
	if nodeType != 1 {
		return plan.planChildren(existing, replacement)
	}

	// A parent owns the presence of a child component, never its internals. The
	// child schedules and patches its own root. Router outlets follow the same
	// rule for their contents while allowing the shell to own outlet attributes.
	componentID := attribute(existing, "data-component-id")
	if componentID != "" && componentID != plan.ownerID {
		if componentID != attribute(replacement, "data-component-id") {
			return fmt.Errorf("rfw DOM patch: component ownership changed from %q", componentID)
		}
		return nil
	}

	formState := captureLiveFormState(existing, replacement)
	plan.planAttributes(existing, replacement)
	if isRouterOutlet(existing) && isRouterOutlet(replacement) {
		return nil
	}
	if attribute(existing, "data-condition") != "" &&
		attribute(existing, "data-condition-branch") != attribute(replacement, "data-condition-branch") {
		return plan.planReplaceChildren(existing, replacement)
	}
	if err := plan.planChildren(existing, replacement); err != nil {
		return err
	}
	if formState != nil {
		plan.ops = append(plan.ops, func() { restoreLiveFormState(existing, *formState) })
	}
	return nil
}

func (plan *patchPlan) planReplaceChildren(parent, replacementParent js.Value) error {
	replacements := significantChildren(replacementParent)
	if _, err := indexIdentities(replacements); err != nil {
		return err
	}
	oldChildren := significantChildren(parent)
	plan.ops = append(plan.ops, func() {
		for _, child := range oldChildren {
			child.Call("remove")
		}
		for _, replacement := range replacements {
			clone := replacement.Call("cloneNode", true)
			parent.Call("appendChild", clone)
			recordRenderedTreeFromSource(clone, replacement)
		}
	})
	return nil
}

func (plan *patchPlan) planAttributes(existing, replacement js.Value) {
	previous := renderedAttributes(existing)
	next := attributeSnapshot(replacement)
	names := make(map[string]struct{}, len(previous)+len(next))
	for name := range previous {
		names[name] = struct{}{}
	}
	for name := range next {
		names[name] = struct{}{}
	}
	for name := range names {
		previousValue, previouslyRendered := previous[name]
		nextValue, renderedNext := next[name]
		if !frameworkOwnedAttribute(name) && previouslyRendered == renderedNext && previousValue == nextValue {
			continue
		}
		if !renderedNext {
			if existing.Call("hasAttribute", name).Bool() {
				attrName := name
				plan.ops = append(plan.ops, func() { existing.Call("removeAttribute", attrName) })
			}
			continue
		}
		if !existing.Call("hasAttribute", name).Bool() || existing.Call("getAttribute", name).String() != nextValue {
			attrName, attrValue := name, nextValue
			plan.ops = append(plan.ops, func() { existing.Call("setAttribute", attrName, attrValue) })
		}
	}
	plan.ops = append(plan.ops, func() { setRenderedAttributes(existing, next) })
}

func (plan *patchPlan) planChildren(parent, replacementParent js.Value) error {
	oldChildren := significantChildren(parent)
	newChildren := significantChildren(replacementParent)

	oldByIdentity, err := indexIdentities(oldChildren)
	if err != nil {
		return err
	}
	if _, err := indexIdentities(newChildren); err != nil {
		return err
	}

	consumed := make([]bool, len(oldChildren))
	placements := make([]childPlacement, 0, len(newChildren))
	nextUnkeyed := 0
	for _, replacement := range newChildren {
		identity := nodeIdentity(replacement)
		if identity != "" {
			if index, ok := oldByIdentity[identity]; ok {
				existing := oldChildren[index]
				if !samePatchType(existing, replacement) {
					return fmt.Errorf("rfw DOM patch: identity %q changed node type", identity)
				}
				consumed[index] = true
				if err := plan.planNode(existing, replacement); err != nil {
					return err
				}
				placements = append(placements, childPlacement{existing: existing})
			} else {
				placements = append(placements, childPlacement{source: replacement})
			}
			continue
		}

		for nextUnkeyed < len(oldChildren) && (consumed[nextUnkeyed] || nodeIdentity(oldChildren[nextUnkeyed]) != "") {
			nextUnkeyed++
		}
		if nextUnkeyed < len(oldChildren) && samePatchType(oldChildren[nextUnkeyed], replacement) {
			existing := oldChildren[nextUnkeyed]
			consumed[nextUnkeyed] = true
			nextUnkeyed++
			if err := plan.planNode(existing, replacement); err != nil {
				return err
			}
			placements = append(placements, childPlacement{existing: existing})
			continue
		}

		placements = append(placements, childPlacement{source: replacement})
	}

	plan.ops = append(plan.ops, func() {
		cursor := firstSignificantChild(parent)
		for index := range placements {
			placement := &placements[index]
			node := placement.existing
			if !node.Truthy() {
				node = placement.source.Call("cloneNode", true)
				recordRenderedTreeFromSource(node, placement.source)
			}
			if !cursor.Truthy() {
				parent.Call("appendChild", node)
			} else if node.Equal(cursor) {
				cursor = nextSignificantSibling(cursor)
			} else {
				parent.Call("insertBefore", node, cursor)
			}
		}
		for index, child := range oldChildren {
			if consumed[index] {
				continue
			}
			currentParent := child.Get("parentNode")
			if currentParent.Truthy() && currentParent.Equal(parent) {
				child.Call("remove")
			}
		}
	})
	return nil
}

func indexIdentities(children []js.Value) (map[string]int, error) {
	indexed := make(map[string]int)
	for index, child := range children {
		identity := nodeIdentity(child)
		if identity == "" {
			continue
		}
		if _, exists := indexed[identity]; exists {
			return nil, fmt.Errorf("rfw DOM patch: duplicate sibling identity %q", identity)
		}
		indexed[identity] = index
	}
	return indexed, nil
}

func nodeIdentity(node js.Value) string {
	if node.Get("nodeType").Int() != 1 {
		return ""
	}
	if key := attribute(node, "data-key"); key != "" {
		return "key:" + attribute(node, "data-for") + ":" + key
	}
	for _, candidate := range []struct {
		attribute string
		prefix    string
	}{
		{"data-component-id", "component:"},
		{"data-condition", "condition:"},
		{"data-for-anchor", "for-anchor:"},
		{"data-portal-id", "portal:"},
	} {
		if value := attribute(node, candidate.attribute); value != "" {
			return candidate.prefix + value
		}
	}
	if node.Call("hasAttribute", "data-router-outlet").Bool() {
		return "router-outlet"
	}
	if node.Call("hasAttribute", "data-portal-anchor").Bool() {
		return "portal-anchor"
	}
	if node.Call("hasAttribute", "data-keepalive-host").Bool() {
		return "keepalive-host"
	}
	return ""
}

func samePatchType(existing, replacement js.Value) bool {
	return existing.Get("nodeType").Int() == replacement.Get("nodeType").Int() &&
		existing.Get("nodeName").String() == replacement.Get("nodeName").String()
}

func significantChildren(parent js.Value) []js.Value {
	children := parent.Get("childNodes")
	out := make([]js.Value, 0, children.Length())
	for i := 0; i < children.Length(); i++ {
		child := children.Index(i)
		if child.Get("nodeType").Int() == 3 && strings.TrimSpace(child.Get("nodeValue").String()) == "" {
			continue
		}
		out = append(out, child)
	}
	return out
}

func firstSignificantChild(parent js.Value) js.Value {
	return nextSignificantNode(parent.Get("firstChild"))
}

func nextSignificantSibling(node js.Value) js.Value {
	return nextSignificantNode(node.Get("nextSibling"))
}

func nextSignificantNode(node js.Value) js.Value {
	for node.Truthy() && node.Get("nodeType").Int() == 3 && strings.TrimSpace(node.Get("nodeValue").String()) == "" {
		node = node.Get("nextSibling")
	}
	return node
}

func isRouterOutlet(node js.Value) bool {
	return node.Get("nodeType").Int() == 1 && node.Call("hasAttribute", "data-router-outlet").Bool()
}

func attribute(node js.Value, name string) string {
	if node.Get("nodeType").Int() != 1 || !node.Call("hasAttribute", name).Bool() {
		return ""
	}
	return node.Call("getAttribute", name).String()
}

func attributeSnapshot(node js.Value) map[string]string {
	attributes := make(map[string]string)
	if node.Get("nodeType").Int() != 1 {
		return attributes
	}
	names := node.Call("getAttributeNames")
	for i := 0; i < names.Length(); i++ {
		name := names.Index(i).String()
		attributes[name] = node.Call("getAttribute", name).String()
	}
	return attributes
}

func renderedAttributes(node js.Value) map[string]string {
	rendered := node.Get(renderedAttrsProperty)
	if rendered.Type() != js.TypeObject {
		return attributeSnapshot(node)
	}
	attributes := make(map[string]string)
	keys := js.Object().Call("keys", rendered)
	for i := 0; i < keys.Length(); i++ {
		name := keys.Index(i).String()
		attributes[name] = rendered.Get(name).String()
	}
	return attributes
}

func setRenderedAttributes(node js.Value, attributes map[string]string) {
	rendered := js.NewDict()
	for name, value := range attributes {
		rendered.Set(name, value)
	}
	node.Set(renderedAttrsProperty, rendered.Value)
}

func recordRenderedTree(root js.Value) {
	if root.Get("nodeType").Int() == 1 {
		setRenderedAttributes(root, attributeSnapshot(root))
	}
	children := root.Get("childNodes")
	for i := 0; i < children.Length(); i++ {
		recordRenderedTree(children.Index(i))
	}
}

func recordRenderedTreeFromSource(node, source js.Value) {
	if node.Get("nodeType").Int() == 1 && source.Get("nodeType").Int() == 1 {
		setRenderedAttributes(node, attributeSnapshot(source))
	}
	nodeChildren := node.Get("childNodes")
	sourceChildren := source.Get("childNodes")
	limit := nodeChildren.Length()
	if sourceChildren.Length() < limit {
		limit = sourceChildren.Length()
	}
	for i := 0; i < limit; i++ {
		recordRenderedTreeFromSource(nodeChildren.Index(i), sourceChildren.Index(i))
	}
}

func frameworkOwnedAttribute(name string) bool {
	return strings.HasPrefix(name, "data-component-") ||
		strings.HasPrefix(name, "data-key") ||
		strings.HasPrefix(name, "data-for") ||
		strings.HasPrefix(name, "data-condition") ||
		strings.HasPrefix(name, "data-router-") ||
		strings.HasPrefix(name, "data-bind-") ||
		strings.HasPrefix(name, "data-on-")
}

func captureLiveFormState(existing, replacement js.Value) *liveFormState {
	controlled := isControlledForm(existing) || isControlledForm(replacement)
	tag := existing.Get("nodeName").String()
	state := &liveFormState{}
	switch tag {
	case "INPUT":
		if !controlled {
			state.value = existing.Get("value").String()
			state.hasValue = true
		}
		typ := strings.ToLower(existing.Get("type").String())
		if !controlled && (typ == "checkbox" || typ == "radio") {
			state.checked = existing.Get("checked").Bool()
			state.hasChecked = true
		}
	case "TEXTAREA":
		if !controlled {
			state.value = existing.Get("value").String()
			state.hasValue = true
		}
	case "SELECT":
		if !controlled {
			state.selectedIndex = existing.Get("selectedIndex").Int()
			state.hasSelected = true
		}
	default:
		return nil
	}
	active := js.Global().Get("document").Get("activeElement")
	if active.Truthy() && active.Equal(existing) {
		start := existing.Get("selectionStart")
		end := existing.Get("selectionEnd")
		if start.Type() == js.TypeNumber && end.Type() == js.TypeNumber {
			state.selectionAt = start.Int()
			state.selectionEnd = end.Int()
			direction := existing.Get("selectionDirection")
			if direction.Type() == js.TypeString {
				state.selectionDir = direction.String()
			}
			state.hasCaret = true
		}
	}
	return state
}

func restoreLiveFormState(element js.Value, state liveFormState) {
	if state.hasValue && element.Get("value").String() != state.value {
		element.Set("value", state.value)
	}
	if state.hasChecked && element.Get("checked").Bool() != state.checked {
		element.Set("checked", state.checked)
	}
	if state.hasSelected && element.Get("selectedIndex").Int() != state.selectedIndex {
		element.Set("selectedIndex", state.selectedIndex)
	}
	if state.hasCaret {
		setter := element.Get("setSelectionRange")
		if setter.Type() == js.TypeFunction {
			if state.selectionDir != "" {
				element.Call("setSelectionRange", state.selectionAt, state.selectionEnd, state.selectionDir)
			} else {
				element.Call("setSelectionRange", state.selectionAt, state.selectionEnd)
			}
		}
	}
}

func isControlledForm(node js.Value) bool {
	if node.Get("nodeType").Int() != 1 {
		return false
	}
	if node.Call("hasAttribute", "data-bind-store").Bool() || node.Call("hasAttribute", "data-bind-signal").Bool() {
		return true
	}
	for _, name := range []string{"value", "checked"} {
		value := attribute(node, name)
		if strings.Contains(value, ":w") && (strings.Contains(value, "@store:") || strings.Contains(value, "@signal:")) {
			return true
		}
	}
	return false
}
