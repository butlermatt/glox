package token

type TokenType string

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal interface{}
	Line    int
}

func New(ty TokenType, lex string, lit interface{}, line int) Token {
	return Token{Type: ty, Lexeme: lex, Literal: lit, Line: line}
}

const (
	// Single Character tokens
	LParen    = "("
	RParen    = ")"
	LBrace    = "{"
	RBrace    = "}"
	LBracket  = "["
	RBracket  = "]"
	Colon     = ":"
	Comma     = ","
	Dot       = "."
	Minus     = "-"
	Plus      = "+"
	Semicolon = ";"
	Slash     = "/"
	Star      = "*"

	// Single or two character tokens.
	Bang      = "!"
	BangEq    = "!="
	Equal     = "="
	EqualEq   = "=="
	Greater   = ">"
	GreaterEq = ">="
	Less      = "<"
	LessEq    = "<="

	// Literals
	Ident    = "IDENT"
	String   = "STRING"
	UTString = "UNTERMINATED STRING"
	Number   = "NUMBER"

	// Keywords
	And    = "AND"
	Class  = "CLASS"
	Else   = "ELSE"
	False  = "FALSE"
	Fun    = "FUN"
	For    = "FOR"
	If     = "IF"
	Null   = "NULL"
	Or     = "OR"
	Print  = "PRINT"
	Return = "RETURN"
	Super  = "SUPER"
	This   = "THIS"
	True   = "TRUE"
	Var    = "VAR"
	While  = "WHILE"

	EOF     = "EOF"
	ILLEGAL = "ILLEGAL"
)
