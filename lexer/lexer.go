package lexer

import "github.com/butlermatt/glpc/token"

type Lexer struct {
	input   string
	tokens  []token.Token
	start   int
	current int
	line    int
}

func New(input string) *Lexer {
	return &Lexer{input: input, line: 1}
}

func (l *Lexer) ScanTokens() {

}
