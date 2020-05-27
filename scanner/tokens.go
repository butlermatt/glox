package scanner

type TokenType byte

const (
	// Single character tokens
	TokenLParen TokenType = iota
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenComma
	TokenDot
	TokenMinus
	TokenPlus
	TokenSemicolon
	TokenSlash
	TokenStar
	// One or two character tokens
	TokenBang
	TokenBangEq
	TokenEqual
	TokenEqualEq
	TokenGreater
	TokenGreaterEq
	TokenLess
	TokenLessEq
	// Literals
	TokenIdent
	TokenString
	TokenNumber
	// Keywords
	TokenAnd
	TokenClass
	TokenElse
	TokenFalse
	TokenFor
	TokenFun
	TokenIf
	TokenNil
	TokenOr
	TokenPrint
	TokenReturn
	TokenSuper
	TokenThis
	TokenTrue
	TokenVar
	TokenWhile
	// Errors
	TokenError
	TokenEof
)

type Token struct {
	Type   TokenType
	Lexeme string
	Line   int
}

func NewToken(ty TokenType, lex string, line int) Token {
	return Token{ty, lex, line}
}

func errorToken(msg string, line int) Token {
	return Token{TokenError, msg, line}
}
