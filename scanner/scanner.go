package scanner

type Scanner struct {
	source  string
	start   int
	current int
	line    int
}

func New(source string) *Scanner {
	return &Scanner{source: source}
}

func (s *Scanner) ScanToken() Token {
	s.skipWhitespace()
	s.start = s.current

	if s.isAtEnd() {
		return NewToken(TokenEof, s.source[s.start:s.current], s.line)
	}

	c := s.advance()
	if isAlpha(c) {
		return s.identifier()
	}
	if isDigit(c) {
		return s.number()
	}

	switch c {
	case '(':
		return NewToken(TokenLParen, s.source[s.start:s.current], s.line)
	case ')':
		return NewToken(TokenRParen, s.source[s.start:s.current], s.line)
	case '{':
		return NewToken(TokenLBrace, s.source[s.start:s.current], s.line)
	case '}':
		return NewToken(TokenRBrace, s.source[s.start:s.current], s.line)
	case ';':
		return NewToken(TokenSemicolon, s.source[s.start:s.current], s.line)
	case ',':
		return NewToken(TokenComma, s.source[s.start:s.current], s.line)
	case '.':
		return NewToken(TokenDot, s.source[s.start:s.current], s.line)
	case '-':
		return NewToken(TokenMinus, s.source[s.start:s.current], s.line)
	case '+':
		return NewToken(TokenPlus, s.source[s.start:s.current], s.line)
	case '/':
		return NewToken(TokenSlash, s.source[s.start:s.current], s.line)
	case '*':
		return NewToken(TokenStar, s.source[s.start:s.current], s.line)
	case '!':
		if s.match('=') {
			return NewToken(TokenBangEq, s.source[s.start:s.current], s.line)
		} else {
			return NewToken(TokenBang, s.source[s.start:s.current], s.line)
		}
	case '=':
		if s.match('=') {
			return NewToken(TokenEqualEq, s.source[s.start:s.current], s.line)
		} else {
			return NewToken(TokenEqual, s.source[s.start:s.current], s.line)
		}
	case '<':
		if s.match('=') {
			return NewToken(TokenLessEq, s.source[s.start:s.current], s.line)
		} else {
			return NewToken(TokenLess, s.source[s.start:s.current], s.line)
		}
	case '>':
		if s.match('=') {
			return NewToken(TokenGreaterEq, s.source[s.start:s.current], s.line)
		} else {
			return NewToken(TokenGreater, s.source[s.start:s.current], s.line)
		}
	case '"':
		return s.string()
	}

	return errorToken("Unexpected character: '" + string(c) + "'", s.line)
}

func (s *Scanner) skipWhitespace() {
	for {
		c := s.peek()
		switch c {
		case ' ', '\r', '\t':
			s.advance()
		case '\n':
			s.line++
			s.advance()
		case '/':
			if s.peekNext() == '/' {
				for s.peek() != '\n' && !s.isAtEnd() {
					s.advance()
				}
			} else {
				return
			}
		default:
			return
		}
	}
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) advance() byte {
	s.current++
	return s.source[s.current-1]
}

func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return 0
	}
	return s.source[s.current]
}

func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return 0
	}
	return s.source[s.current+1]
}

func (s *Scanner) match(c byte) bool {
	if s.isAtEnd() || s.source[s.current] != c {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) string() Token {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		return errorToken("Unterminated string.", s.line)
	}

	s.advance() // Consume last quote
	return NewToken(TokenString, s.source[s.start:s.current], s.line)
}

func (s *Scanner) number() Token {
	for isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance() // consume the .
		for isDigit(s.peek()) {
			s.advance()
		}
	}

	return NewToken(TokenNumber, s.source[s.start:s.current], s.line)
}

func (s *Scanner) identifier() Token {
	for c := s.peek(); isAlpha(c) || isDigit(c); c = s.peek() {
		s.advance()
	}

	return NewToken(s.identifierType(), s.source[s.start:s.current], s.line)
}

func (s *Scanner) identifierType() TokenType {
	switch s.source[s.start] {
	case 'a':
		return s.checkKeyword("and", TokenAnd)
	case 'c':
		return s.checkKeyword("class", TokenClass)
	case 'e':
		return s.checkKeyword("else", TokenElse)
	case 'f':
		if s.current-s.start > 1 {
			switch s.source[s.start+1] {
			case 'a':
				s.checkKeyword("false", TokenFalse)
			case 'o':
				s.checkKeyword("for", TokenFor)
			case 'u':
				s.checkKeyword("fun", TokenFun)
			}
		}
	case 'i':
		return s.checkKeyword("if", TokenIf)
	case 'n':
		return s.checkKeyword("nil", TokenNil)
	case 'o':
		return s.checkKeyword("or", TokenOr)
	case 'p':
		return s.checkKeyword("print", TokenPrint)
	case 'r':
		return s.checkKeyword("return", TokenReturn)
	case 's':
		return s.checkKeyword("super", TokenSuper)
	case 't':
		if s.current-s.start > 1 {
			switch s.source[s.start+1] {
			case 'h':
				return s.checkKeyword("this", TokenThis)
			case 'r':
				return s.checkKeyword("true", TokenTrue)
			}
		}
	case 'v':
		return s.checkKeyword("var", TokenVar)
	case 'w':
		return s.checkKeyword("while", TokenWhile)
	}

	return TokenIdent
}

func (s *Scanner) checkKeyword(keyword string, tok TokenType) TokenType {
	if s.source[s.start:s.current] == keyword {
		return tok
	}
	return TokenIdent
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}
