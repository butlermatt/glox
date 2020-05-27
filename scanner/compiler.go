package scanner

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

import bc "github.com/butlermatt/glox/bytecode"

type Parser struct {
	current   Token
	previous  Token
	hadError  bool
	panicMode bool
}

type Compiler struct {
	scan           *Scanner
	parser         Parser
	compilingChunk *bc.Chunk
}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(source string, chunk *bc.Chunk) bool {
	c.scan = New(source)
	c.compilingChunk = chunk
	c.parser = Parser{}

	c.Advance()
	c.Expression()
	c.Consume(TokenEof, "Expected end of expression")

	c.endCompiler()
	return !c.parser.hadError
}

func (c *Compiler) Advance() {
	c.parser.previous = c.parser.current

	for {
		c.parser.current = c.scan.ScanToken()
		if c.parser.current.Type != TokenError {
			break
		}

		c.errorAtCurrent(c.parser.current.Lexeme)
	}
}

func (c *Compiler) Consume(tt TokenType, msg string) {
	if c.parser.current.Type == tt {
		c.Advance()
		return
	}

	c.errorAtCurrent(msg)
}

func (c *Compiler) expression() {

}

func (c *Compiler) number() {
	value, err := strconv.ParseFloat(c.parser.previous.Lexeme, 64)
	if err != nil {
		c.error("failed to parse number")
		return
	}

	c.emitConstant(bc.Value(value))
}

func (c *Compiler) CurrentChunk() *bc.Chunk {
	return c.compilingChunk
}

func (c *Compiler) MakeConstant(value bc.Value) byte {
	cInd := c.CurrentChunk().AddConstant(value)
	if cInd > math.MaxUint8 {
		c.error("Too many constants in one chunk.")
		return 0
	}

	return byte(cInd)
}

func (c *Compiler) emitBytes(op bc.OpCode, bytes ...byte) {
	c.CurrentChunk().WriteOp(c.parser.previous.Line, op, bytes...)
}

func (c *Compiler) emitReturn() {
	c.emitBytes(bc.OpReturn)
}

func (c *Compiler) emitConstant(value bc.Value) {
	c.emitBytes(bc.OpConstant, c.MakeConstant(value))
}

func (c *Compiler) endCompiler() {
	c.emitReturn()
}

func (c *Compiler) errorAtCurrent(msg string) {
	c.errorAt(c.parser.current, msg)
}

func (c *Compiler) error(msg string) {
	c.errorAt(c.parser.previous, msg)
}

func (c *Compiler) errorAt(t Token, msg string) {
	if c.parser.panicMode {
		return
	}

	c.parser.panicMode = true

	_, _ = fmt.Fprintf(os.Stderr, "[line %d] Error", t.Line)

	if t.Type == TokenEof {
		_, _ = fmt.Fprintf(os.Stderr, " at end")
	} else if t.Type == TokenError {
		// Nothing
	} else {
		_, _ = fmt.Fprintf(os.Stderr, " at '%s'", t.Lexeme)
	}

	_, _ = fmt.Fprintf(os.Stderr, ": %s\n", msg)
	c.parser.hadError = true
}
