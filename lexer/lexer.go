package lexer

import "strconv"

type Lexer struct {
	input   string
	tokens  []*Token
	start   int // Start of current token
	current int // Current position
	line    int // Current line
	index   int // token index in tokens.
}

var keywords = map[string]TokenType{
	"and":    And,
	"class":  Class,
	"else":   Else,
	"false":  False,
	"fun":    Fun,
	"for":    For,
	"if":     If,
	"null":   Null,
	"or":     Or,
	"print":  Print,
	"return": Return,
	"super":  Super,
	"this":   This,
	"true":   True,
	"var":    Var,
	"while":  While,
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

	l.tokens = append(l.tokens, NewToken(EOF, "", nil, l.line))
}

// NextToken steps through the input to generate the next token
func (l *Lexer) NextToken() *Token {
	if l.index >= len(l.tokens) {
		return nil
	}

	tok := l.tokens[l.index]
	l.index += 1
	return tok
}

func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.input)
}

func (l *Lexer) readChar() byte {
	ch := l.input[l.current]
	l.current += 1
	return ch
}

func (l *Lexer) match(expected byte) bool {
	if l.isAtEnd() || l.input[l.current] != expected {
		return false
	}

	l.current += 1
	return true
}

func (l *Lexer) peek() byte {
	if l.current >= len(l.input) {
		return 0
	}
	return l.input[l.current]
}

func (l *Lexer) peekNext() byte {
	if l.current+1 >= len(l.input) {
		return 0
	}
	return l.input[l.current+1]
}

func (l *Lexer) addToken(ty TokenType, literal interface{}) {
	l.tokens = append(l.tokens, NewToken(ty, l.input[l.start:l.current], literal, l.line))
}

func (l *Lexer) scanToken() {
	c := l.readChar()
	switch c {
	case ' ', '\t', '\r': // Blah whitespace!
		break
	case '\n': // Whitespace, but we want the new line.
		l.line += 1
	case ';':
		l.addToken(Semicolon, nil)
	case ':':
		l.addToken(Colon, nil)
	case '(':
		l.addToken(LParen, nil)
	case ')':
		l.addToken(RParen, nil)
	case '[':
		l.addToken(LBracket, nil)
	case ']':
		l.addToken(RBracket, nil)
	case '{':
		l.addToken(LBrace, nil)
	case '}':
		l.addToken(RBrace, nil)
	case ',':
		l.addToken(Comma, nil)
	case '.':
		l.addToken(Dot, nil)
	case '+':
		l.addToken(Plus, nil)
	case '-':
		l.addToken(Minus, nil)
	case '*':
		l.addToken(Star, nil)
	case '/':
		if l.match('/') {
			for l.peek() != '\n' && l.peek() != 0 {
				l.readChar()
			}
		} else {
			l.addToken(Slash, nil)
		}
	case '=':
		if l.match('=') {
			l.addToken(EqualEq, nil)
		} else {
			l.addToken(Equal, nil)
		}
	case '>':
		if l.match('=') {
			l.addToken(GreaterEq, nil)
		} else {
			l.addToken(Greater, nil)
		}
	case '<':
		if l.match('=') {
			l.addToken(LessEq, nil)
		} else {
			l.addToken(Less, nil)
		}
	case '!':
		if l.match('=') {
			l.addToken(BangEq, nil)
		} else {
			l.addToken(Bang, nil)
		}
	case '"':
		l.string()
	case '`':
		l.rawString()
	default:
		if isDigit(c) {
			l.number()
		} else if isAlpha(c) {
			l.identifier()
		} else {
			l.addToken(Illegal, nil)
		}
	}
}

func (l *Lexer) string() {
	for l.peek() != '"' && l.peek() != '\n' && !l.isAtEnd() {
		l.readChar()
	}

	if l.isAtEnd() || l.peek() == '\n' {
		l.addToken(UTString, l.input[l.start+1:l.current])
		return
	}

	l.readChar()
	l.addToken(String, l.input[l.start+1:l.current-1])
}

func (l *Lexer) rawString() {
	line := l.line

	for l.peek() != '`' && !l.isAtEnd() {
		if l.peek() == '\n' {
			l.line += 1
		}

		l.readChar()
	}

	if l.isAtEnd() {
		// Error points to line at start of string not end of string.
		l.tokens = append(l.tokens, NewToken(UTString, l.input[l.start:l.current], l.input[l.start:l.current], line))
		return
	}

	l.readChar()
	l.tokens = append(l.tokens, NewToken(String, l.input[l.start:l.current], l.input[l.start+1:l.current-1], line))
}

func (l *Lexer) number() {
	for isDigit(l.peek()) {
		l.readChar()
	}

	if l.peek() == '.' && isDigit(l.peekNext()) {
		l.readChar()

		for isDigit(l.peek()) {
			l.readChar()
		}
	}

	value, err := strconv.ParseFloat(l.input[l.start:l.current], 64)
	if err != nil {
		l.addToken(Illegal, nil)
		return
	}
	l.addToken(Number, value)
}

func (l *Lexer) identifier() {
	for isAlphaNumeric(l.peek()) {
		l.readChar()
	}

	text := l.input[l.start:l.current]
	if tokenType, ok := keywords[text]; ok {
		l.addToken(tokenType, nil)
	} else {
		l.addToken(Ident, nil)
	}
}
