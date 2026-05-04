package hostclient

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	hostVarAttr        = "data-host-var"
	hostExpectedAttr   = "data-host-expected"
	expectationHashAlg = "sha1"
)

type hostVarElement interface {
	Exists() bool
	Text() string
	SetText(string)
	Attr(string) string
	SetAttr(string, string)
}

type componentRoot interface {
	HostVar(string) hostVarElement
	SetHTML(string)
}

type hydrationMismatch struct {
	VarName     string
	Expected    string
	Actual      string
	ActualHash  string
	ExpectedAlg string
}

type initSnapshotPayload struct {
	HTML string
	Vars []string
}

func encodeExpectation(value string) string {
	sum := sha1.Sum([]byte(value))
	return fmt.Sprintf("%s:%s", expectationHashAlg, hex.EncodeToString(sum[:]))
}

func expectationMatches(expectedAttr, actual string) (bool, string, string) {
	actualHash := encodeExpectation(actual)
	if expectedAttr == "" {
		return true, expectationHashAlg, actualHash
	}
	if strings.HasPrefix(expectedAttr, expectationHashAlg+":") {
		return expectedAttr == actualHash, expectationHashAlg, actualHash
	}
	return expectedAttr == actual, "raw", actualHash
}

func updateHostVar(root componentRoot, name, value string) *hydrationMismatch {
	node := root.HostVar(name)
	if !node.Exists() {
		return nil
	}
	expectedAttr := node.Attr(hostExpectedAttr)
	actualText := node.Text()
	matches, alg, actualHash := expectationMatches(expectedAttr, actualText)
	if !matches {
		return &hydrationMismatch{
			VarName:     name,
			Expected:    expectedAttr,
			Actual:      actualText,
			ActualHash:  actualHash,
			ExpectedAlg: alg,
		}
	}
	node.SetText(value)
	node.SetAttr(hostExpectedAttr, encodeExpectation(value))
	return nil
}

func handleHostPayload(root componentRoot, payload map[string]any) []hydrationMismatch {
	mismatches := make([]hydrationMismatch, 0)
	for key, raw := range payload {
		if key == "initSnapshot" || strings.HasPrefix(key, "_") {
			continue
		}
		mismatch := updateHostVar(root, key, fmt.Sprintf("%v", raw))
		if mismatch != nil {
			mismatches = append(mismatches, *mismatch)
		}
	}
	return mismatches
}

func applyInitSnapshot(root componentRoot, payload *initSnapshotPayload) {
	if payload == nil {
		return
	}
	root.SetHTML(payload.HTML)
}

func decodeInitSnapshotPayload(raw any) *initSnapshotPayload {
	if raw == nil {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	html, _ := m["html"].(string)
	if html == "" {
		return nil
	}
	var vars []string
	if list, ok := m["vars"].([]any); ok {
		vars = make([]string, 0, len(list))
		for _, item := range list {
			if s, ok := item.(string); ok {
				vars = append(vars, s)
			}
		}
	} else if list, ok := m["vars"].([]string); ok {
		vars = append(vars, list...)
	}
	return &initSnapshotPayload{HTML: html, Vars: vars}
}

func buildResyncPayload(mismatches []hydrationMismatch) map[string]any {
	entries := make([]map[string]string, 0, len(mismatches))
	for _, m := range mismatches {
		entries = append(entries, map[string]string{
			"var":         m.VarName,
			"expected":    m.Expected,
			"expectedAlg": m.ExpectedAlg,
			"actual":      m.Actual,
			"actualHash":  m.ActualHash,
		})
	}
	return map[string]any{
		"resync": map[string]any{
			"reason": "host-var-mismatch",
			"vars":   entries,
		},
	}
}
