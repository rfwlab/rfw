//go:build js && wasm

package rtmlast

import (
	"fmt"
	"strings"

	"github.com/rfwlab/rfw/v2/state"
)

type HTMLComponent interface {
	GetID() string
	GetProps() map[string]any
	AddUnsubscribe(fn func())
}

type RenderContext struct {
	Component HTMLComponent
	Props     map[string]any
	StoreMgr  *state.StoreManager
}

func RenderNodes(nodes []Node, ctx *RenderContext) string {
	var sb strings.Builder
	for _, n := range nodes {
		sb.WriteString(renderNode(n, ctx))
	}
	return sb.String()
}

func renderNode(n Node, ctx *RenderContext) string {
	switch v := n.(type) {
	case TextNode:
		return v.Text
	case ExprNode:
		return renderExprNode(v, ctx)
	case VarNode:
		return renderVar(v, ctx)
	case IfNode:
		return renderIf(v, ctx)
	case ForNode:
		return renderFor(v, ctx)
	case IncludeNode:
		return renderInclude(v, ctx)
	case SlotNode:
		return renderSlot(v, ctx)
	case CommandNode:
		return renderCommand(v, ctx)
	default:
		return ""
	}
}

func renderVar(v VarNode, ctx *RenderContext) string {
	val := evalExpr(v.Expr, ctx)
	return fmt.Sprintf(`<span data-var>%v</span>`, val)
}

func renderExprNode(v ExprNode, ctx *RenderContext) string {
	val := evalExpr(v.Expr, ctx)
	return fmt.Sprintf(`<span data-expr>%v</span>`, val)
}

func renderIf(v IfNode, ctx *RenderContext) string {
	condVal := evalBool(v.Cond, ctx)
	if condVal {
		return RenderNodes(v.Then, ctx)
	}
	for _, branch := range v.ElseIf {
		if evalBool(branch.Cond, ctx) {
			return RenderNodes(branch.Body, ctx)
		}
	}
	if len(v.Else) > 0 {
		return RenderNodes(v.Else, ctx)
	}
	return ""
}

func renderFor(v ForNode, ctx *RenderContext) string {
	collection := evalExpr(v.Expr, ctx)
	var items []any
	switch c := collection.(type) {
	case []any:
		items = c
	case map[string]any:
		for k, val := range c {
			items = append(items, map[string]any{"key": k, "value": val})
		}
	default:
		return ""
	}
	var sb strings.Builder
	for i, item := range items {
		childCtx := *ctx
		if childCtx.Props == nil {
			childCtx.Props = map[string]any{}
		}
		childCtx.Props[v.Alias] = item
		if v.KeyAlias != "" {
			switch k := item.(type) {
			case map[string]any:
				childCtx.Props[v.KeyAlias] = k["key"]
			default:
				childCtx.Props[v.KeyAlias] = i
			}
		}
		sb.WriteString(RenderNodes(v.Body, &childCtx))
	}
	return sb.String()
}

func renderInclude(v IncludeNode, ctx *RenderContext) string {
	return fmt.Sprintf(`@include:%s`, v.Name)
}

func renderSlot(v SlotNode, ctx *RenderContext) string {
	if content, ok := ctx.Props["slot:"+v.Name]; ok {
		if s, ok := content.(string); ok {
			return s
		}
	}
	return RenderNodes(v.Fallback, ctx)
}

func renderCommand(v CommandNode, ctx *RenderContext) string {
	switch v.Kind {
	case "store":
		return renderStoreCmd(v.Value, ctx)
	case "signal":
		return renderSignalCmd(v.Value, ctx)
	case "prop":
		return renderPropCmd(v.Value, ctx)
	case "on":
		return renderEventCmd(v.Value)
	case "h":
		return fmt.Sprintf(`<span data-host-var="%s" data-host-expected=""></span>`, v.Value)
	default:
		return fmt.Sprintf(`@%s:%s`, v.Kind, v.Value)
	}
}

func renderStoreCmd(val string, ctx *RenderContext) string {
	parts := strings.Split(val, ".")
	isW := strings.HasSuffix(val, ":w")
	if isW && len(parts) >= 3 {
		cleanVal := strings.TrimSuffix(val, ":w")
		cparts := strings.Split(cleanVal, ".")
		if len(cparts) == 3 {
			placeholder := fmt.Sprintf("@store:%s:w", cleanVal)
			return placeholder
		}
	}
	if len(parts) == 3 && !isW {
		store := ctx.StoreMgr.GetStore(parts[0], parts[1])
		if store != nil {
			v := store.Get(parts[2])
			return fmt.Sprintf(`<span data-store="%s">%v</span>`, val, v)
		}
	}
	return fmt.Sprintf(`@store:%s`, val)
}

func renderSignalCmd(val string, ctx *RenderContext) string {
	name := val
	isW := strings.HasSuffix(val, ":w")
	if isW {
		name = strings.TrimSuffix(val, ":w")
		return fmt.Sprintf("@signal:%s:w", name)
	}
	if prop, ok := ctx.Props[name]; ok {
		if sig, ok := prop.(interface{ Read() any }); ok {
			v := sig.Read()
			return fmt.Sprintf(`<span data-signal="%s">%v</span>`, name, v)
		}
	}
	return fmt.Sprintf(`@signal:%s`, name)
}

func renderPropCmd(val string, ctx *RenderContext) string {
	if v, ok := ctx.Props[val]; ok {
		return fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf(`@prop:%s`, val)
}

func renderEventCmd(val string) string {
	parts := strings.SplitN(val, ":", 2)
	if len(parts) == 2 {
		event := parts[0]
		handler := parts[1]
		return fmt.Sprintf(`data-on-%s="%s"`, event, handler)
	}
	return fmt.Sprintf(`data-on-%s`, val)
}

func evalExpr(e Expr, ctx *RenderContext) any {
	switch v := e.(type) {
	case IdentExpr:
		return lookupIdent(v.Name, ctx)
	case LiteralExpr:
		return v.Value
	case BinaryExpr:
		return evalBinary(v, ctx)
	case UnaryExpr:
		return evalUnary(v, ctx)
	case FieldExpr:
		obj := evalExpr(v.Obj, ctx)
		if m, ok := obj.(map[string]any); ok {
			return m[v.Field]
		}
		return nil
	case CallExpr:
		return nil
	case TernaryExpr:
		if evalBool(v.Cond, ctx) {
			return evalExpr(v.Then, ctx)
		}
		return evalExpr(v.Else, ctx)
	default:
		return nil
	}
}

func evalBool(e Expr, ctx *RenderContext) bool {
	val := evalExpr(e, ctx)
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int:
		return v != 0
	case float64:
		return v != 0
	case nil:
		return false
	default:
		return true
	}
}

func evalBinary(b BinaryExpr, ctx *RenderContext) any {
	lhs := evalExpr(b.LHS, ctx)
	rhs := evalExpr(b.RHS, ctx)
	switch b.Op {
	case OpEq:
		return fmt.Sprintf("%v", lhs) == fmt.Sprintf("%v", rhs)
	case OpNeq:
		return fmt.Sprintf("%v", lhs) != fmt.Sprintf("%v", rhs)
	case OpAnd:
		return toBool(lhs) && toBool(rhs)
	case OpOr:
		return toBool(lhs) || toBool(rhs)
	case OpLt, OpGt, OpLte, OpGte:
		return compareValues(lhs, rhs, b.Op)
	case OpAdd:
		return toFloat(lhs) + toFloat(rhs)
	case OpSub:
		return toFloat(lhs) - toFloat(rhs)
	case OpMul:
		return toFloat(lhs) * toFloat(rhs)
	case OpDiv:
		r := toFloat(rhs)
		if r == 0 {
			return 0.0
		}
		return toFloat(lhs) / r
	default:
		return nil
	}
}

func evalUnary(u UnaryExpr, ctx *RenderContext) any {
	val := evalExpr(u.Expr, ctx)
	switch u.Op {
	case UnaryNot:
		return !toBool(val)
	case UnaryNeg:
		return -toFloat(val)
	default:
		return val
	}
}

func lookupIdent(name string, ctx *RenderContext) any {
	if strings.HasPrefix(name, "store:") {
		parts := strings.Split(strings.TrimPrefix(name, "store:"), ".")
		if len(parts) == 3 && ctx.StoreMgr != nil {
			store := ctx.StoreMgr.GetStore(parts[0], parts[1])
			if store != nil {
				return store.Get(parts[2])
			}
		}
		return nil
	}
	if strings.HasPrefix(name, "signal:") {
		sigName := strings.TrimPrefix(name, "signal:")
		if prop, ok := ctx.Props[sigName]; ok {
			if sig, ok := prop.(interface{ Read() any }); ok {
				return sig.Read()
			}
			return prop
		}
		return nil
	}
	if v, ok := ctx.Props[name]; ok {
		if sig, ok := v.(interface{ Read() any }); ok {
			return sig.Read()
		}
		return v
	}
	return nil
}

func toBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != ""
	case int:
		return val != 0
	case float64:
		return val != 0
	default:
		return v != nil
	}
}

func toFloat(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case float64:
		return val
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

func compareValues(lhs, rhs any, op BinOp) bool {
	l, r := toFloat(lhs), toFloat(rhs)
	switch op {
	case OpLt:
		return l < r
	case OpGt:
		return l > r
	case OpLte:
		return l <= r
	case OpGte:
		return l >= r
	default:
		return false
	}
}