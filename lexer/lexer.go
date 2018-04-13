package lexer

import "github.com/butlermatt/glpc/token"

type Lexer struct {
	input   string
	tokens  []token.Token
	start   int
	current int
	line    int
}

func isAlpha(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlphaNumeric(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

// New returns a new Lexer populated with the specified input program.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
	return l
}

func (l *Lexer) ScanTokens() {
	for !l.isAtEnd() {
		l.start = l.current
		l.scanToken()
	}
}

func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.input)
}

func (l *Lexer) readChar() byte {
	ch := l.input[l.current]
	l.current += 1
	return ch
}

func (l *Lexer) peekChar() byte {
	if l.current >= len(l.input) {
		return 0
	}
	return l.input[l.current]
}

func (l *Lexer) addToken(ty token.TokenType, literal interface{}) {
	l.tokens = append(l.tokens, token.New(ty, l.input[l.start:l.current], literal, l.line))
}

func (l *Lexer) scanToken() {
	c := l.readChar()
	switch c {
	case ' ', '\t', '\r': // Blah whitespace!
		break
	case '\n': // Whitespace, but we want the new line.
		l.line += 1
	case ';':
		l.addToken(token.Semicolon, nil)
	case ':':
		l.addToken(token.Colon, nil)
	case '(':
		l.addToken(token.LParen, nil)
	case ')':
		l.addToken(token.RParen, nil)
	case '[':
		l.addToken(token.LBracket, nil)
	case ']':
		l.addToken(token.RBracket, nil)
	case '{':
		l.addToken(token.LBrace, nil)
	case '}':
		l.addToken(token.RBrace, nil)
	case '+':
		l.addToken(token.Plus, nil)
	case '-':
		l.addToken(token.Minus, nil)
	case '*':
		l.addToken(token.Star, nil)
	case '/':
		if l.peekChar() == '/' {
			for l.peekChar() != '\n' && l.peekChar() != 0 {
				l.readChar()
			}
		} else {
			l.addToken(token.Slash, nil)
		}
	case '=':
		if l.peekChar() == '=' {
			l.addToken(token.EqualEq, nil)
		} else {
			l.addToken(token.Equal, nil)
		}
	case '>':
		if l.peekChar() == '=' {
			l.addToken(token.GreaterEq, nil)
		} else {
			l.addToken(token.Greater, nil)
		}
	case '<':
		if l.peekChar() == '=' {
			l.addToken(token.LessEq, nil)
		} else {
			l.addToken(token.Less, nil)
		}
	case '!':
		if l.peekChar() == '=' {
			l.addToken(token.BangEq, nil)
		} else {
			l.addToken(token.Bang, nil)
		}
	}
}

// NextToken steps through the input to generate the next token
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	// TODO (iterate through tokens that are in the slice

	return tok
}
