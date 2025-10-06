package hostclient

import (
	"regexp"
	"testing"
)

type fakeElement struct {
	text      string
	expected  string
	exists    bool
	attrStore map[string]string
}

func (e *fakeElement) Exists() bool { return e.exists }

func (e *fakeElement) Text() string { return e.text }

func (e *fakeElement) SetText(v string) { e.text = v }

func (e *fakeElement) Attr(name string) string {
	if name == hostExpectedAttr {
		return e.expected
	}
	if e.attrStore != nil {
		return e.attrStore[name]
	}
	return ""
}

func (e *fakeElement) SetAttr(name, value string) {
	if name == hostExpectedAttr {
		e.expected = value
		return
	}
	if e.attrStore == nil {
		e.attrStore = make(map[string]string)
	}
	e.attrStore[name] = value
}

type fakeRoot struct {
	elems map[string]*fakeElement
	html  string
}

func newFakeRoot() *fakeRoot {
	return &fakeRoot{elems: make(map[string]*fakeElement)}
}

func (r *fakeRoot) HostVar(name string) hostVarElement {
	if el, ok := r.elems[name]; ok {
		return el
	}
	return &fakeElement{}
}

func (r *fakeRoot) SetHTML(html string) {
	r.html = html
	r.elems = make(map[string]*fakeElement)
	re := regexp.MustCompile(`<span[^>]*data-host-var="([^"]+)"[^>]*data-host-expected="([^"]*)"[^>]*>([^<]*)</span>`)
	matches := re.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		name := m[1]
		expected := m[2]
		text := m[3]
		r.elems[name] = &fakeElement{exists: true, expected: expected, text: text}
	}
}

func TestHandleHostPayloadMismatchTriggersResync(t *testing.T) {
	root := newFakeRoot()
	root.elems["greeting"] = &fakeElement{
		exists:   true,
		expected: encodeExpectation("server"),
		text:     "tampered",
	}

	payload := map[string]any{"greeting": "fresh"}
	mismatches := handleHostPayload(root, payload)
	if len(mismatches) != 1 {
		t.Fatalf("expected 1 mismatch, got %d", len(mismatches))
	}
	if root.elems["greeting"].text != "tampered" {
		t.Fatalf("text was updated despite mismatch")
	}
	resync := buildResyncPayload(mismatches)
	body, ok := resync["resync"].(map[string]any)
	if !ok {
		t.Fatalf("resync payload missing body")
	}
	if body["reason"] != "host-var-mismatch" {
		t.Fatalf("unexpected reason %v", body["reason"])
	}
	vars, ok := body["vars"].([]map[string]string)
	if ok {
		if vars[0]["var"] != "greeting" {
			t.Fatalf("unexpected var name %s", vars[0]["var"])
		}
		if vars[0]["expected"] == vars[0]["actualHash"] {
			t.Fatalf("expected hashes to differ on mismatch")
		}
	}
}

func TestInitSnapshotRecoveryAndUpdate(t *testing.T) {
	root := newFakeRoot()
	root.elems["count"] = &fakeElement{
		exists:   true,
		expected: encodeExpectation("1"),
		text:     "0",
	}

	if mismatches := handleHostPayload(root, map[string]any{"count": "2"}); len(mismatches) == 0 {
		t.Fatalf("expected mismatch when expectation diverges")
	}

	snapHTML := `<span data-host-var="count" data-host-expected="` + encodeExpectation("1") + `">1</span>`
	applyInitSnapshot(root, &initSnapshotPayload{HTML: snapHTML})

	if mismatches := handleHostPayload(root, map[string]any{"count": "3"}); len(mismatches) != 0 {
		t.Fatalf("expected clean hydration after snapshot")
	}

	elem := root.HostVar("count").(*fakeElement)
	if elem.text != "3" {
		t.Fatalf("expected text to update to 3, got %s", elem.text)
	}
	if elem.expected != encodeExpectation("3") {
		t.Fatalf("expected hash to reflect new value")
	}
}
