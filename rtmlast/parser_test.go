package rtmlast

import (
	"testing"
)

func TestParseText(t *testing.T) {
	nodes, err := Parse("hello world")
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	tn, ok := nodes[0].(TextNode)
	if !ok {
		t.Fatalf("expected TextNode, got %T", nodes[0])
	}
	if tn.Text != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", tn.Text)
	}
}

func TestParseVarInterpolation(t *testing.T) {
	input := "hello {{name}} world"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	vn, ok := nodes[1].(VarNode)
	if !ok {
		t.Fatalf("expected VarNode, got %T", nodes[1])
	}
	ident, ok := vn.Expr.(IdentExpr)
	if !ok {
		t.Fatalf("expected IdentExpr, got %T", vn.Expr)
	}
	if ident.Name != "name" {
		t.Fatalf("expected 'name', got '%s'", ident.Name)
	}
}

func TestParseIfConditional(t *testing.T) {
	input := "@if:active\nhello\n@else\nworld\n@endif"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	ifn, ok := nodes[0].(IfNode)
	if !ok {
		t.Fatalf("expected IfNode, got %T", nodes[0])
	}
	ident, ok := ifn.Cond.(IdentExpr)
	if !ok {
		t.Fatalf("expected IdentExpr condition, got %T", ifn.Cond)
	}
	if ident.Name != "active" {
		t.Fatalf("expected condition 'active', got '%s'", ident.Name)
	}
	if len(ifn.Then) == 0 {
		t.Fatal("expected Then branch")
	}
	if len(ifn.Else) == 0 {
		t.Fatal("expected Else branch")
	}
}

func TestParseForLoop(t *testing.T) {
	input := "@for:item in items\n<div>{{item}}</div>\n@endfor"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	fn, ok := nodes[0].(ForNode)
	if !ok {
		t.Fatalf("expected ForNode, got %T", nodes[0])
	}
	if fn.Alias != "item" {
		t.Fatalf("expected alias 'item', got '%s'", fn.Alias)
	}
}

func TestParseInclude(t *testing.T) {
	input := "@include:MyComponent"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	inc, ok := nodes[0].(IncludeNode)
	if !ok {
		t.Fatalf("expected IncludeNode, got %T", nodes[0])
	}
	if inc.Name != "MyComponent" {
		t.Fatalf("expected 'MyComponent', got '%s'", inc.Name)
	}
}

func TestParseBinaryExpr(t *testing.T) {
	expr := ParseExpr("count > 0")
	bin, ok := expr.(BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}
	if bin.Op != OpGt {
		t.Fatalf("expected OpGt, got %d", bin.Op)
	}
}

func TestParseBinaryExprEq(t *testing.T) {
	expr := ParseExpr("status == \"active\"")
	bin, ok := expr.(BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}
	if bin.Op != OpEq {
		t.Fatalf("expected OpEq, got %d", bin.Op)
	}
}

func TestParseUnaryNot(t *testing.T) {
	expr := ParseExpr("!visible")
	un, ok := expr.(UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", expr)
	}
	if un.Op != UnaryNot {
		t.Fatalf("expected UnaryNot, got %d", un.Op)
	}
}

func TestParseFieldExpr(t *testing.T) {
	expr := ParseExpr("user.name")
	field, ok := expr.(FieldExpr)
	if !ok {
		t.Fatalf("expected FieldExpr, got %T", expr)
	}
	if field.Field != "name" {
		t.Fatalf("expected 'name', got '%s'", field.Field)
	}
}

func TestParseCallExpr(t *testing.T) {
	expr := ParseExpr("format(date)")
	call, ok := expr.(CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", expr)
	}
	if call.Fn != "format" {
		t.Fatalf("expected 'format', got '%s'", call.Fn)
	}
}

func TestParseElseIf(t *testing.T) {
	input := "@if:a\nA\n@else-if:b\nB\n@else\nC\n@endif"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	ifn, ok := nodes[0].(IfNode)
	if !ok {
		t.Fatalf("expected IfNode, got %T", nodes[0])
	}
	if len(ifn.ElseIf) != 1 {
		t.Fatalf("expected 1 else-if, got %d", len(ifn.ElseIf))
	}
	if len(ifn.Else) == 0 {
		t.Fatal("expected else branch")
	}
}

func TestParseSlot(t *testing.T) {
	input := "@slot:header\n<h1>fallback</h1>\n@endslot"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	sn, ok := nodes[0].(SlotNode)
	if !ok {
		t.Fatalf("expected SlotNode, got %T", nodes[0])
	}
	if sn.Name != "header" {
		t.Fatalf("expected 'header', got '%s'", sn.Name)
	}
}

func TestParseStoreIdent(t *testing.T) {
	expr := ParseExpr("store:app.user.name")
	ident, ok := expr.(IdentExpr)
	if !ok {
		t.Fatalf("expected IdentExpr, got %T", expr)
	}
	if ident.Name != "store:app.user.name" {
		t.Fatalf("expected 'store:app.user.name', got '%s'", ident.Name)
	}
}

func TestParseSignalIdent(t *testing.T) {
	expr := ParseExpr("signal:count")
	ident, ok := expr.(IdentExpr)
	if !ok {
		t.Fatalf("expected IdentExpr, got %T", expr)
	}
	if ident.Name != "signal:count" {
		t.Fatalf("expected 'signal:count', got '%s'", ident.Name)
	}
}

func TestParseComplexTemplate(t *testing.T) {
	input := "<div>{{user.name}}@if:admin\nadmin panel\n@endif</div>"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) == 0 {
		t.Fatal("expected nodes")
	}
}