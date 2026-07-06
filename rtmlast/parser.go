package rtmlast

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenText TokenType = iota
	TokenCommand
	TokenVarOpen
	TokenVarClose
	TokenEOF
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) Lex() []Token {
	var tokens []Token
	var text strings.Builder

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if ch == '{' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '{' {
			if text.Len() > 0 {
				tokens = append(tokens, Token{Type: TokenText, Value: text.String()})
				text.Reset()
			}
			l.pos += 2
			var expr strings.Builder
			for l.pos < len(l.input) {
				if l.pos+1 < len(l.input) && l.input[l.pos] == '}' && l.input[l.pos+1] == '}' {
					l.pos += 2
					break
				}
				expr.WriteByte(l.input[l.pos])
				l.pos++
			}
			tokens = append(tokens, Token{Type: TokenVarOpen, Value: strings.TrimSpace(expr.String())})
			tokens = append(tokens, Token{Type: TokenVarClose, Value: "}}"})

		} else if ch == '@' && isCommandPrefix(l.input[l.pos+1:]) {
			if text.Len() > 0 {
				tokens = append(tokens, Token{Type: TokenText, Value: text.String()})
				text.Reset()
			}
			l.pos++
			start := l.pos
			for l.pos < len(l.input) && !isNewline(l.input[l.pos]) && l.input[l.pos] != '<' {
				l.pos++
			}
			cmdText := strings.TrimSpace(l.input[start:l.pos])
			tokens = append(tokens, Token{Type: TokenCommand, Value: cmdText})

		} else {
			text.WriteByte(ch)
			l.pos++
		}
	}

	if text.Len() > 0 {
		tokens = append(tokens, Token{Type: TokenText, Value: text.String()})
	}
	tokens = append(tokens, Token{Type: TokenEOF})
	return tokens
}

func isCommandPrefix(s string) bool {
	prefixes := []string{"if:", "else-if:", "else", "endif", "for:", "endfor", "foreach:", "endforeach", "include:", "slot:", "endslot", "on:", "store:", "signal:", "prop:", "h:", "plugin:", "expr:"}
	for _, p := range prefixes {
		if len(s) > len(p) && s[:len(p)] == p {
			return true
		}
		if s == p {
			return true
		}
	}
	return false
}

func isNewline(ch byte) bool {
	return ch == '\n' || ch == '\r'
}

func Parse(template string) ([]Node, error) {
	lex := NewLexer(template)
	tokens := lex.Lex()
	p := &parser{tokens: tokens, pos: 0}
	return p.parseNodes()
}

type parser struct {
	tokens []Token
	pos    int
}

func (p *parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) next() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	t := p.tokens[p.pos]
	p.pos++
	return t
}

func (p *parser) parseNodes() ([]Node, error) {
	var nodes []Node
	for {
		t := p.peek()
		if t.Type == TokenEOF {
			break
		}
		switch t.Type {
		case TokenText:
			p.next()
			if t.Value != "" {
				nodes = append(nodes, TextNode{Text: t.Value})
			}
		case TokenVarOpen:
			p.next()
			expr := ParseExpr(t.Value)
			nodes = append(nodes, VarNode{Expr: expr})
			if p.peek().Type == TokenVarClose {
				p.next()
			}
		case TokenCommand:
			p.next()
			cmdText := t.Value
			node, consumed, err := p.parseCommand(cmdText)
			if err != nil {
				return nodes, err
			}
			if node != nil {
				nodes = append(nodes, node)
			}
			_ = consumed
		default:
			p.next()
		}
	}
	return nodes, nil
}

func (p *parser) parseCommand(cmdText string) (Node, bool, error) {
	colonIdx := strings.Index(cmdText, ":")
	if colonIdx < 0 {
		switch cmdText {
		case "else":
			return nil, false, nil
		case "endif":
			return nil, false, nil
		case "endfor":
			return nil, false, nil
		case "endslot":
			return nil, false, nil
		case "endforeach":
			return nil, false, nil
		}
		return CommandNode{Kind: cmdText, Value: ""}, true, nil
	}

	kind := cmdText[:colonIdx]
	rest := cmdText[colonIdx+1:]

	switch kind {
	case "if":
		node, err := p.parseIf(rest)
		return node, true, err
	case "for":
		node, err := p.parseFor(rest)
		return node, true, err
	case "foreach":
		node, err := p.parseForeach(rest)
		return node, true, err
	case "include":
		return IncludeNode{Name: rest, Props: nil}, true, nil
	case "slot":
		return p.parseSlot(rest)
	case "on":
		return CommandNode{Kind: "on", Value: rest}, true, nil
	case "store", "signal", "prop":
		return CommandNode{Kind: kind, Value: rest}, true, nil
	case "h":
		return CommandNode{Kind: "h", Value: rest}, true, nil
	case "plugin":
		return CommandNode{Kind: "plugin", Value: rest}, true, nil
	case "expr":
		return ExprNode{Expr: ParseExpr(rest)}, true, nil
	default:
		return CommandNode{Kind: kind, Value: rest}, true, nil
	}
}

func (p *parser) parseIf(condStr string) (Node, error) {
	thenNodes, _ := p.parseUntilCommands("else-if", "else", "endif")
	node := IfNode{Cond: ParseExpr(condStr), Then: thenNodes}

	for p.peek().Type == TokenCommand && strings.HasPrefix(p.peek().Value, "else-if:") {
		cmdText := p.next().Value
		elseCond := strings.TrimPrefix(cmdText, "else-if:")
		body, _ := p.parseUntilCommands("else-if", "else", "endif")
		node.ElseIf = append(node.ElseIf, ElseIfBranch{Cond: ParseExpr(elseCond), Body: body})
	}

	if p.peek().Type == TokenCommand && p.peek().Value == "else" {
		p.next()
		elseBody, _ := p.parseUntilCommands("endif")
		node.Else = elseBody
	}

	if p.peek().Type == TokenCommand && p.peek().Value == "endif" {
		p.next()
	}
	return node, nil
}

func (p *parser) parseFor(detail string) (Node, error) {
	body, _ := p.parseUntilCommands("endfor")
	if p.peek().Type == TokenCommand && p.peek().Value == "endfor" {
		p.next()
	}

	parts := strings.SplitN(detail, " in ", 2)
	alias := strings.TrimSpace(parts[0])
	exprStr := ""
	if len(parts) > 1 {
		exprStr = strings.TrimSpace(parts[1])
	}
	keyAlias := ""
	if commaIdx := strings.Index(alias, ","); commaIdx != -1 {
		keyAlias = strings.TrimSpace(alias[commaIdx+1:])
		alias = strings.TrimSpace(alias[:commaIdx])
	}
	return ForNode{Alias: alias, KeyAlias: keyAlias, Expr: ParseExpr(exprStr), Body: body}, nil
}

func (p *parser) parseForeach(detail string) (Node, error) {
	parts := strings.SplitN(detail, " as ", 2)
	exprStr := strings.TrimSpace(parts[0])
	alias := ""
	if len(parts) > 1 {
		alias = strings.TrimSpace(parts[1])
	}
	body, _ := p.parseUntilCommands("endforeach")
	if p.peek().Type == TokenCommand && p.peek().Value == "endforeach" {
		p.next()
	}
	return ForNode{Alias: alias, Expr: ParseExpr(exprStr), Body: body}, nil
}

func (p *parser) parseSlot(name string) (Node, bool, error) {
	body, _ := p.parseUntilCommands("endslot")
	if p.peek().Type == TokenCommand && p.peek().Value == "endslot" {
		p.next()
	}
	return SlotNode{Name: name, Fallback: body}, true, nil
}

func (p *parser) parseUntilCommands(commands ...string) ([]Node, error) {
	var nodes []Node
	for {
		t := p.peek()
		if t.Type == TokenEOF {
			break
		}
		if t.Type == TokenCommand {
			for _, cmd := range commands {
				if t.Value == cmd || strings.HasPrefix(t.Value, cmd+":") || strings.HasPrefix(t.Value, cmd+" ") {
					return nodes, nil
				}
			}
		}
		p.next()
		switch t.Type {
		case TokenText:
			if t.Value != "" {
				nodes = append(nodes, TextNode{Text: t.Value})
			}
		case TokenVarOpen:
			nodes = append(nodes, VarNode{Expr: ParseExpr(t.Value)})
			if p.peek().Type == TokenVarClose {
				p.next()
			}
		case TokenCommand:
			node, _, err := p.parseCommand(t.Value)
			if err != nil {
				return nodes, err
			}
			if node != nil {
				nodes = append(nodes, node)
			}
		}
	}
	return nodes, nil
}

func ParseExpr(s string) Expr {
	s = strings.TrimSpace(s)
	if s == "" {
		return LiteralExpr{Value: ""}
	}
	// Ternary: cond then X else Y (preferred) or cond ? X : Y (legacy)
	if idx, ok := findTernaryThen(s); ok {
		condStr := strings.TrimSpace(s[:idx])
		rest := s[idx+5:] // len(" then") = 5
		elseIdx, ok2 := findElse(rest)
		if ok2 {
			thenStr := strings.TrimSpace(rest[:elseIdx])
			elseStr := strings.TrimSpace(rest[elseIdx+5:]) // len(" else") = 5
			return TernaryExpr{Cond: ParseExpr(condStr), Then: ParseExpr(thenStr), Else: ParseExpr(elseStr)}
		}
	}
	if idx, ok := findTernarySymbol(s); ok {
		condStr := strings.TrimSpace(s[:idx])
		rest := s[idx+1:]
		colonIdx := findTernaryColon(rest)
		if colonIdx >= 0 {
			thenStr := strings.TrimSpace(rest[:colonIdx])
			elseStr := strings.TrimSpace(rest[colonIdx+1:])
			return TernaryExpr{Cond: ParseExpr(condStr), Then: ParseExpr(thenStr), Else: ParseExpr(elseStr)}
		}
	}
	if result, ok := tryParseBinary(s); ok {
		return result
	}
	if strings.HasPrefix(s, "!") || strings.HasPrefix(s, "not ") {
		inner := s
		if strings.HasPrefix(s, "not ") {
			inner = strings.TrimSpace(s[4:])
		} else {
			inner = strings.TrimSpace(s[1:])
		}
		return UnaryExpr{Op: UnaryNot, Expr: ParseExpr(inner)}
	}
	if strings.HasPrefix(s, "-") && len(s) > 1 {
		return UnaryExpr{Op: UnaryNeg, Expr: ParseExpr(s[1:])}
	}
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return LiteralExpr{Value: s[1 : len(s)-1]}
	}
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") && len(s) >= 2 {
		return LiteralExpr{Value: s[1 : len(s)-1]}
	}
	if n, ok := tryParseNumber(s); ok {
		return LiteralExpr{Value: n}
	}
	if strings.HasPrefix(s, "store:") || strings.HasPrefix(s, "signal:") || strings.HasPrefix(s, "prop:") {
		return IdentExpr{Name: s}
	}
	if strings.Contains(s, ".") && !strings.Contains(s, "(") {
		parts := strings.SplitN(s, ".", 2)
		if isIdent(parts[0]) && isIdent(parts[1]) {
			return FieldExpr{Obj: ParseExpr(parts[0]), Field: parts[1]}
		}
	}
	if strings.Contains(s, "(") {
		parenIdx := strings.Index(s, "(")
		fnName := s[:parenIdx]
		argsStr := s[parenIdx+1 : len(s)-1]
		var args []Expr
		if argsStr != "" {
			for _, a := range strings.Split(argsStr, ",") {
				args = append(args, ParseExpr(strings.TrimSpace(a)))
			}
		}
		return CallExpr{Fn: fnName, Args: args}
	}
	return IdentExpr{Name: s}
}

// findTernaryThen finds " then " outside strings and parens.
func findTernaryThen(s string) (int, bool) {
	depth := 0
	inStr := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' || ch == '\'' {
			inStr = !inStr
		}
		if inStr {
			continue
		}
		if ch == '(' {
			depth++
		}
		if ch == ')' {
			depth--
		}
		if depth > 0 {
			continue
		}
		if i+5 <= len(s) && s[i:i+5] == " then" && (i+5 >= len(s) || s[i+5] == ' ') {
			return i, true
		}
	}
	return -1, false
}

// findElse finds " else " outside strings and parens.
func findElse(s string) (int, bool) {
	depth := 0
	inStr := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' || ch == '\'' {
			inStr = !inStr
		}
		if inStr {
			continue
		}
		if ch == '(' {
			depth++
		}
		if ch == ')' {
			depth--
		}
		if depth > 0 {
			continue
		}
		if i+5 <= len(s) && s[i:i+5] == " else" && (i+5 >= len(s) || s[i+5] == ' ') {
			return i, true
		}
	}
	return -1, false
}

// findTernarySymbol finds "?" outside strings and parens.
func findTernarySymbol(s string) (int, bool) {
	depth := 0
	inStr := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' || ch == '\'' {
			inStr = !inStr
		}
		if inStr {
			continue
		}
		if ch == '(' {
			depth++
		}
		if ch == ')' {
			depth--
		}
		if depth > 0 {
			continue
		}
		if ch == '?' {
			return i, true
		}
	}
	return -1, false
}

// findTernaryColon finds ":" outside strings and parens, used for legacy ? : ternary.
func findTernaryColon(s string) int {
	depth := 0
	inStr := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' || ch == '\'' {
			inStr = !inStr
		}
		if inStr {
			continue
		}
		if ch == '(' {
			depth++
		}
		if ch == ')' {
			depth--
		}
		if depth > 0 {
			continue
		}
		if ch == ':' {
			return i
		}
	}
	return -1
}

func isIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if !unicode.IsLetter(ch) && ch != '_' && !unicode.IsDigit(ch) {
			return false
		}
	}
	return true
}

func tryParseNumber(s string) (any, bool) {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err == nil {
		if f == float64(int(f)) {
			return int(f), true
		}
		return f, true
	}
	return nil, false
}

func tryParseBinary(s string) (Expr, bool) {
	type opInfo struct {
		op   BinOp
		syms []string
	}
	ops := []opInfo{
		{OpOr, []string{" or ", "||"}},
		{OpAnd, []string{" and ", "&&"}},
		{OpEq, []string{" is ", "=="}},
		{OpNeq, []string{" is not ", "!="}},
		{OpLte, []string{"<="}},
		{OpGte, []string{">="}},
		{OpLt, []string{"<"}},
		{OpGt, []string{">"}},
		{OpAdd, []string{"+"}},
		{OpSub, []string{"-"}},
		{OpMul, []string{"*"}},
		{OpDiv, []string{"/"}},
	}
	for _, o := range ops {
		for _, sym := range o.syms {
			idx := findBinaryOp(s, sym)
			if idx >= 0 {
				lhs := strings.TrimSpace(s[:idx])
				rhs := strings.TrimSpace(s[idx+len(sym):])
				if lhs != "" && rhs != "" {
					return BinaryExpr{Op: o.op, LHS: ParseExpr(lhs), RHS: ParseExpr(rhs)}, true
				}
			}
		}
	}
	return nil, false
}

func findBinaryOp(s, op string) int {
	depth := 0
	inStr := false
	inVar := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' || ch == '\'' {
			inStr = !inStr
		}
		if inStr {
			continue
		}
		if ch == '{' {
			inVar = true
		}
		if ch == '}' {
			inVar = false
		}
		if inVar {
			continue
		}
		if ch == '(' {
			depth++
		}
		if ch == ')' {
			depth--
		}
		if depth > 0 {
			continue
		}
		if i+len(op) <= len(s) && s[i:i+len(op)] == op {
			return i
		}
	}
	return -1
}