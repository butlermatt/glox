package scanner

import "fmt"

type Compiler struct {
	scan *Scanner
}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(source string) {
	c.scan = New(source)

	line := -1
	for {
		token := c.scan.ScanToken()
		if token.Line != line {
			fmt.Printf("%4d ", token.Line)
			line = token.Line
		} else {
			fmt.Printf("   | ")
		}

		fmt.Printf("%2d '%s'\n", token.Type, token.Lexeme)
		if token.Type == TokenEof {
			break
		}
	}
}
