package tokenizer

import "strings"

// TokenKind classifies the type of UA token.
type TokenKind uint8

const (
	KindProduct TokenKind = iota // Mozilla/5.0, Chrome/120
	KindComment                  // (Windows NT 10.0; x64)
	KindFBBlock                  // [FBAN/...] — non-standard Facebook block
)

// Token is a single parsed element from a User-Agent string.
type Token struct {
	Kind    TokenKind
	Name    string   // "Mozilla", "Chrome"
	Version string   // "5.0", "120.0.0.0"
	Attrs   []string // for Comment: split by ";"
	Raw     string
}

const initialTokenCap = 8

// Tokenize parses a User-Agent string into a slice of tokens.
// It uses a finite state machine with three states: scanning products,
// reading comments (parenthesized), and reading FB blocks (bracketed).
func Tokenize(ua string) []Token {
	if len(ua) == 0 {
		return nil
	}

	tokens := make([]Token, 0, initialTokenCap)
	i := 0
	n := len(ua)

	for i < n {
		for i < n && ua[i] == ' ' {
			i++
		}

		if i >= n {
			break
		}

		switch ua[i] {
		case '(':
			tok, end := readComment(ua, i, n)

			tokens = append(tokens, tok)
			i = end
		case '[':
			tok, end := readFBBlock(ua, i, n)

			tokens = append(tokens, tok)
			i = end
		default:
			tok, end := readProduct(ua, i, n)

			tokens = append(tokens, tok)
			i = end
		}
	}

	return tokens
}

// readProduct reads a product token like "Mozilla/5.0" or "Chrome/120.0.0.0".
// Stops at whitespace, '(' or '['.
func readProduct(ua string, start, n int) (Token, int) {
	i := start

	for i < n && ua[i] != ' ' && ua[i] != '(' && ua[i] != '[' {
		i++
	}

	raw := ua[start:i]
	name := raw
	version := ""

	if n, v, ok := strings.Cut(raw, "/"); ok {
		name = n
		version = v
	}

	return Token{
		Kind:    KindProduct,
		Name:    name,
		Version: version,
		Raw:     raw,
	}, i
}

// readComment reads a parenthesized comment like "(Windows NT 10.0; Win64; x64)".
// Handles nested parentheses by tracking depth.
func readComment(ua string, start, n int) (Token, int) {
	depth := 0
	i := start

	for i < n {
		switch ua[i] {
		case '(':
			depth++
		case ')':
			depth--

			if depth == 0 {
				i++

				raw := ua[start:i]
				inner := raw[1 : len(raw)-1]
				attrs := splitAttrs(inner)

				return Token{
					Kind:  KindComment,
					Attrs: attrs,
					Raw:   raw,
				}, i
			}
		}

		i++
	}

	raw := ua[start:i]
	inner := raw[1:]
	attrs := splitAttrs(inner)

	return Token{
		Kind:  KindComment,
		Attrs: attrs,
		Raw:   raw,
	}, i
}

// readFBBlock reads a Facebook-style bracketed block like "[FBAN/FBIOS;FBAV/450.0.0;...]".
func readFBBlock(ua string, start, n int) (Token, int) {
	i := start + 1

	for i < n && ua[i] != ']' {
		i++
	}

	if i < n {
		i++
	}

	raw := ua[start:i]

	inner := raw[1:]
	if len(inner) > 0 && inner[len(inner)-1] == ']' {
		inner = inner[:len(inner)-1]
	}

	attrs := splitAttrs(inner)

	return Token{
		Kind:  KindFBBlock,
		Attrs: attrs,
		Raw:   raw,
	}, i
}

// splitAttrs splits inner text by ";", trimming whitespace.
func splitAttrs(s string) []string {
	if len(s) == 0 {
		return nil
	}

	parts := strings.Split(s, ";")

	attrs := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) > 0 {
			attrs = append(attrs, p)
		}
	}

	return attrs
}
