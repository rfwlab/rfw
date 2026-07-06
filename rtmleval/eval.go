// Package rtmleval evaluates RTML expressions with full operator support.
// Supports both symbol operators (==, &&, ||, !) and word operators
// (is, is not, and, or, not, then, else) with word operators preferred.
package rtmleval

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Eval evaluates an RTML expression string against a variable lookup.
// lookup(name) returns the value for a variable.
func Eval(expr string, lookup func(string) (any, bool)) (any, error) {
	p := &parser{input: expr, lookup: lookup}
	p.readChar()
	p.next()
	return p.parseTernary()
}

// Bool evaluates and coerces the result to bool.
func Bool(expr string, lookup func(string) (any, bool)) (bool, error) {
	v, err := Eval(expr, lookup)
	if err != nil {
		return false, err
	}
	return toBool(v), nil
}

// String returns the evaluated result as string.
func String(expr string, lookup func(string) (any, bool)) (string, error) {
	v, err := Eval(expr, lookup)
	if err != nil {
		return "", err
	}
	return toString(v), nil
}

// toBool coerces a Go value to bool.
func toBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val != "" && val != "false" && val != "0"
	case nil:
		return false
	default:
		return v != nil
	}
}

// toFloat coerces a Go value to float64.
func toFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// toString coerces a Go value to string.
func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// ── Lexer ─────────────────────────────────────────────────────────────────────

type tokenType int

const (
	tEOF tokenType = iota
	tIdent
	tString
	tNumber
	tTrue
	tFalse
	tDot
	tColon
	tLParen
	tRParen
	tPlus
	tMinus
	tStar
	tSlash
	tEq
	tNeq
	tLt
	tLte
	tGt
	tGte
	tAnd
	tOr
	tNot
	tIs
	tThen
	tElse
	tQuestion
)

type token struct {
	typ tokenType
	val string
}

type parser struct {
	input  string
	pos    int
	ch     byte
	cur    token
	lookup func(string) (any, bool)
}

func (p *parser) readChar() {
	if p.pos >= len(p.input) {
		p.ch = 0
	} else {
		p.ch = p.input[p.pos]
	}
	p.pos++
}

func (p *parser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *parser) skipWhitespace() {
	for p.ch != 0 && (p.ch == ' ' || p.ch == '\t' || p.ch == '\n' || p.ch == '\r') {
		p.readChar()
	}
}

func (p *parser) next() {
	p.skipWhitespace()
	if p.ch == 0 {
		p.cur = token{tEOF, ""}
		return
	}
	switch p.ch {
	case '(':
		p.cur = token{tLParen, "("}
		p.readChar()
		return
	case ')':
		p.cur = token{tRParen, ")"}
		p.readChar()
		return
	case '+':
		p.cur = token{tPlus, "+"}
		p.readChar()
		return
	case '-':
		p.cur = token{tMinus, "-"}
		p.readChar()
		return
	case '*':
		p.cur = token{tStar, "*"}
		p.readChar()
		return
	case '/':
		p.cur = token{tSlash, "/"}
		p.readChar()
		return
	case '.':
		p.cur = token{tDot, "."}
		p.readChar()
		return
	case ':':
		p.cur = token{tColon, ":"}
		p.readChar()
		return
	case '"':
		p.readString('"')
		return
	case '\'':
		p.readString('\'')
		return
	case '=':
		if p.peek() == '=' {
			p.readChar()
			p.readChar()
			p.cur = token{tEq, "=="}
			return
		}
		// Single '=' is not an operator in our grammar; skip it.
		p.readChar()
		p.next()
		return
	case '!':
		if p.peek() == '=' {
			p.readChar()
			p.readChar()
			p.cur = token{tNeq, "!="}
			return
		}
		p.cur = token{tNot, "!"}
		p.readChar()
		return
	case '<':
		if p.peek() == '=' {
			p.readChar()
			p.readChar()
			p.cur = token{tLte, "<="}
			return
		}
		p.cur = token{tLt, "<"}
		p.readChar()
		return
	case '>':
		if p.peek() == '=' {
			p.readChar()
			p.readChar()
			p.cur = token{tGte, ">="}
			return
		}
		p.cur = token{tGt, ">"}
		p.readChar()
		return
	case '&':
		if p.peek() == '&' {
			p.readChar()
			p.readChar()
			p.cur = token{tAnd, "&&"}
			return
		}
		// Single '&' is not an operator; skip it.
		p.readChar()
		p.next()
		return
	case '|':
		if p.peek() == '|' {
			p.readChar()
			p.readChar()
			p.cur = token{tOr, "||"}
			return
		}
		// Single '|' is not an operator; skip it.
		p.readChar()
		p.next()
		return
	case '?':
		p.cur = token{tQuestion, "?"}
		p.readChar()
		return
	}

	if p.ch == '"' {
		p.readString('"')
		return
	}
	if p.ch == '\'' {
		p.readString('\'')
		return
	}
	if isDigit(p.ch) || (p.ch == '.' && isDigit(p.peek())) {
		p.readNumber()
		return
	}
	if isIdentStart(p.ch) {
		p.readIdent()
		return
	}

	// Unknown character: skip and continue.
	p.readChar()
	p.next()
}

func (p *parser) readString(quote byte) {
	var sb strings.Builder
	p.readChar()
	for p.ch != quote && p.ch != 0 {
		if p.ch == '\\' {
			p.readChar()
			switch p.ch {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case '\'':
				sb.WriteByte('\'')
			default:
				sb.WriteByte(p.ch)
			}
		} else {
			sb.WriteByte(p.ch)
		}
		p.readChar()
	}
	if p.ch == quote {
		p.readChar()
	}
	p.cur = token{tString, sb.String()}
}

func (p *parser) readNumber() {
	start := p.pos - 1
	for isDigit(p.ch) || p.ch == '.' {
		p.readChar()
	}
	p.cur = token{tNumber, p.input[start : p.pos-1]}
}

func (p *parser) readIdent() {
	start := p.pos - 1
	for isIdentPart(p.ch) || p.ch == '-' {
		p.readChar()
	}
	val := p.input[start : p.pos-1]
	switch val {
	case "true":
		p.cur = token{tTrue, val}
	case "false":
		p.cur = token{tFalse, val}
	case "and":
		p.cur = token{tAnd, val}
	case "or":
		p.cur = token{tOr, val}
	case "not":
		p.cur = token{tNot, val}
	case "is":
		p.cur = token{tIs, val}
	case "then":
		p.cur = token{tThen, val}
	case "else":
		p.cur = token{tElse, val}
	default:
		p.cur = token{tIdent, val}
	}
}

func isIdentStart(ch byte) bool { return unicode.IsLetter(rune(ch)) || ch == '_' }
func isIdentPart(ch byte) bool  { return unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' || ch == '.' }
func isDigit(ch byte) bool      { return '0' <= ch && ch <= '9' }

// ── Recursive descent ────────────────────────────────────────────────────────

// parseTernary handles: cond then X else Y
// Also handles the symbol form: cond ? X : Y (kept for backward compat)
func (p *parser) parseTernary() (any, error) {
	cond, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.cur.typ == tThen {
		p.next()
		thenVal, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.cur.typ != tElse {
			return nil, fmt.Errorf("expected 'else' in ternary expression")
		}
		p.next()
		elseVal, err := p.parseTernary()
		if err != nil {
			return nil, err
		}
		if toBool(cond) {
			return thenVal, nil
		}
		return elseVal, nil
	}
	// Backward compat: symbol ? X : Y
	if p.cur.typ == tQuestion {
		p.next()
		thenVal, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.cur.typ != tColon {
			return nil, fmt.Errorf("expected ':' in ternary expression")
		}
		p.next()
		elseVal, err := p.parseTernary()
		if err != nil {
			return nil, err
		}
		if toBool(cond) {
			return thenVal, nil
		}
		return elseVal, nil
	}
	return cond, nil
}

func (p *parser) parseOr() (any, error) {
	lhs, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.cur.typ == tOr {
		p.next()
		rhs, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		lhs = toBool(lhs) || toBool(rhs)
	}
	return lhs, nil
}

func (p *parser) parseAnd() (any, error) {
	lhs, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.cur.typ == tAnd {
		p.next()
		rhs, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		lhs = toBool(lhs) && toBool(rhs)
	}
	return lhs, nil
}

// parseEquality handles ==, !=, is, is not
func (p *parser) parseEquality() (any, error) {
	lhs, err := p.parseRelational()
	if err != nil {
		return nil, err
	}
	for {
		switch p.cur.typ {
		case tEq:
			p.next()
			rhs, err := p.parseRelational()
			if err != nil {
				return nil, err
			}
			lhs = cmpEqual(lhs, rhs)
		case tNeq:
			p.next()
			rhs, err := p.parseRelational()
			if err != nil {
				return nil, err
			}
			lhs = !cmpEqual(lhs, rhs)
		case tIs:
			p.next()
			// Check for "is not" (negated equality)
			if p.cur.typ == tNot {
				p.next()
				rhs, err := p.parseRelational()
				if err != nil {
					return nil, err
				}
				lhs = !cmpEqual(lhs, rhs)
			} else {
				rhs, err := p.parseRelational()
				if err != nil {
					return nil, err
				}
				lhs = cmpEqual(lhs, rhs)
			}
		default:
			return lhs, nil
		}
	}
}

func (p *parser) parseRelational() (any, error) {
	lhs, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	for p.cur.typ == tLt || p.cur.typ == tLte || p.cur.typ == tGt || p.cur.typ == tGte {
		op := p.cur.typ
		p.next()
		rhs, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		a, aok := toFloat(lhs)
		b, bok := toFloat(rhs)
		if !aok || !bok {
			return false, nil
		}
		switch op {
		case tLt:
			lhs = a < b
		case tLte:
			lhs = a <= b
		case tGt:
			lhs = a > b
		case tGte:
			lhs = a >= b
		}
	}
	return lhs, nil
}

func (p *parser) parseAdditive() (any, error) {
	lhs, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	for p.cur.typ == tPlus || p.cur.typ == tMinus {
		op := p.cur.typ
		p.next()
		rhs, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		a, aok := toFloat(lhs)
		b, bok := toFloat(rhs)
		if aok && bok {
			if op == tPlus {
				lhs = a + b
			} else {
				lhs = a - b
			}
		} else {
			// String concatenation.
			lhs = toString(lhs) + toString(rhs)
		}
	}
	return lhs, nil
}

func (p *parser) parseMultiplicative() (any, error) {
	lhs, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.cur.typ == tStar || p.cur.typ == tSlash {
		op := p.cur.typ
		p.next()
		rhs, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		a, aok := toFloat(lhs)
		b, bok := toFloat(rhs)
		if !aok || !bok {
			return nil, fmt.Errorf("incompatible types for * or /")
		}
		if op == tStar {
			lhs = a * b
		} else {
			lhs = a / b
		}
	}
	return lhs, nil
}

func (p *parser) parseUnary() (any, error) {
	if p.cur.typ == tNot {
		p.next()
		v, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return !toBool(v), nil
	}
	if p.cur.typ == tMinus {
		p.next()
		v, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		f, ok := toFloat(v)
		if !ok {
			return nil, fmt.Errorf("cannot negate non-numeric value")
		}
		return -f, nil
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (any, error) {
	switch p.cur.typ {
	case tString:
		v := p.cur.val
		p.next()
		return v, nil
	case tNumber:
		v, err := strconv.ParseFloat(p.cur.val, 64)
		if err != nil {
			return nil, err
		}
		p.next()
		return v, nil
	case tTrue:
		p.next()
		return true, nil
	case tFalse:
		p.next()
		return false, nil
	case tIdent:
		name := p.cur.val
		p.next()
		// Field access: ident.ident.ident
		for p.cur.typ == tDot {
			p.next()
			if p.cur.typ != tIdent {
				return nil, fmt.Errorf("expected field name after .")
			}
			name += "." + p.cur.val
			p.next()
		}
		// Variable lookup.
		if p.lookup != nil {
			if v, ok := p.lookup(name); ok {
				return v, nil
			}
		}
		return name, nil // fallback: return the name as a string
	case tLParen:
		p.next()
		expr, err := p.parseTernary()
		if err != nil {
			return nil, err
		}
		if p.cur.typ != tRParen {
			return nil, fmt.Errorf("expected )")
		}
		p.next()
		return expr, nil
	default:
		return nil, fmt.Errorf("unexpected token %q", p.cur.val)
	}
}

func cmpEqual(a, b any) bool {
	// Fast paths for same type.
	switch av := a.(type) {
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
	case int:
		if bv, ok := b.(int); ok {
			return av == bv
		}
	}
	// Fallback: numeric comparison.
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return af == bf
	}
	return toString(a) == toString(b)
}