// Package rtmlast defines the abstract syntax tree for RTML templates.
// It replaces the regex-based parser in v1/core/rtml.go.
package rtmlast

// Node is the root of the AST.
type Node interface {
	node()
}

func (TextNode) node()      {}
func (VarNode) node()       {}
func (ExprNode) node()      {}
func (ElementNode) node()   {}
func (CommandNode) node()   {}
func (IfNode) node()        {}
func (ForNode) node()       {}
func (SlotNode) node()      {}
func (IncludeNode) node()   {}

// TextNode is literal text.
type TextNode struct {
	Text string
}

// ExprNode is a standalone reactive expression @expr:expression.
type ExprNode struct {
	Expr Expr
}

// VarNode is a reactive interpolation {expr}.
type VarNode struct {
	Expr Expr
}

// ElementNode is a rendered HTML element. It holds the tag name,
// attributes (plain + bound), children, and optional [ref] / [key].
type ElementNode struct {
	Tag        string
	Attrs      []Attr
	BoundAttrs []BoundAttr
	Children   []Node
	Ref        string   // from [ref] constructor
	Key        Expr     // from [key {expr}] constructor
	IsVoid     bool     // self-closing
}

// Attr is a static HTML attribute.
type Attr struct {
	Name  string
	Value string
}

// BoundAttr is an attribute whose value is computed from a reactive Expr.
type BoundAttr struct {
	Name string
	Expr Expr
	Bool bool // if true, present only when truthy
}

// CommandNode is a top-level command or constructor that does not produce markup.
type CommandNode struct {
	Kind  string
	Value string
}

// IfNode is a conditional block.
type IfNode struct {
	Cond      Expr
	Then      []Node
	ElseIf    []ElseIfBranch
	Else      []Node
}

// ElseIfBranch is an @else-if branch.
type ElseIfBranch struct {
	Cond Expr
	Body []Node
}

// ForNode is a list loop.
type ForNode struct {
	Alias   string      // e.g. "item"
	KeyAlias string     // e.g. "key" in @for:key,val in obj
	Expr    Expr        // collection expression
	Body    []Node
}

// SlotNode is a named/placeholder slot.
type SlotNode struct {
	Name     string
	Fallback []Node
}

// IncludeNode is a component inclusion with inline props.
type IncludeNode struct {
	Name  string
	Props map[string]Expr
}

// Expr represents a reactive expression.
// It is later evaluated against component scope.
type Expr interface {
	expr()
}

func (IdentExpr) expr()      {}
func (LiteralExpr) expr()    {}
func (BinaryExpr) expr()     {}
func (UnaryExpr) expr()      {}
func (CallExpr) expr()       {}
func (FieldExpr) expr()      {}
func (TernaryExpr) expr()    {}

// IdentExpr is a variable reference by name.
type IdentExpr struct {
	Name string
}

// LiteralExpr is a string, number, or bool literal.
type LiteralExpr struct {
	Value any // string, float64, bool
}

// BinaryExpr is lhs op rhs.
type BinaryExpr struct {
	Op  BinOp
	LHS Expr
	RHS Expr
}

// BinOp is a binary operator.
type BinOp int

const (
	OpUnknown BinOp = iota
	OpEq
	OpNeq
	OpLt
	OpLte
	OpGt
	OpGte
	OpAnd
	OpOr
	OpAdd
	OpSub
	OpMul
	OpDiv
)

// UnaryExpr is !expr or -expr.
type UnaryExpr struct {
	Op   UnaryOp
	Expr Expr
}

// UnaryOp is a unary operator.
type UnaryOp int

const (
	UnaryUnknown UnaryOp = iota
	UnaryNot
	UnaryNeg
)

// CallExpr is function(args).
type CallExpr struct {
	Fn   string
	Args []Expr
}

// FieldExpr is obj.Field (dotted access).
type FieldExpr struct {
	Obj   Expr
	Field string
}

// TernaryExpr is cond ? then : else (rare, but supported).
type TernaryExpr struct {
	Cond Expr
	Then Expr
	Else Expr
}
