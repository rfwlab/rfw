# rtml

```go
import "github.com/rfwlab/rfw/v2/rtml"
```

RTML AST, parsing, rendering, and node types for the template language.

## Parsing & Rendering

| Function | Description |
| --- | --- |
| `Parse(template string) ([]Node, error)` | Parse an RTML template into nodes. |
| `RenderNodes(nodes []Node, ctx *RenderContext) string` | Render nodes to an HTML string. |

## Node Types

| Type | Description |
| --- | --- |
| `TextNode` | Raw text content. |
| `VarNode` | Variable interpolation `{name}`. |
| `ExprNode` | Expression output `{expr}`. |
| `ElementNode` | HTML element with tag, attrs, children. |
| `CommandNode` | Generic `@command` directive. |
| `IfNode` | Conditional block `@if`. |
| `ForNode` | Loop block `@for`. |
| `SlotNode` | Named slot insertion point. |
| `IncludeNode` | Template inclusion `@include`. |

## Expression Types

| Type | Description |
| --- | --- |
| `IdentExpr` | Identifier reference. |
| `LiteralExpr` | Literal value (string, number, bool). |
| `BinaryExpr` | Binary operation (e.g. `+`, `&&`). |
| `UnaryExpr` | Unary operation (e.g. `!`, `-`). |
| `CallExpr` | Function call. |
| `FieldExpr` | Field access (`a.b`). |
| `TernaryExpr` | Ternary conditional (`a ? b : c`). |