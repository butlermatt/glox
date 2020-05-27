package scanner

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

import (
	bc "github.com/butlermatt/glox/bytecode"
	"github.com/butlermatt/glox/debug"
)

const DEBUG_PRINT_CODE = true

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
	rules          []ParseRule
}

type Precedence byte

const (
	PrecNone       Precedence = iota
	PrecAssign                // =
	PrecOr                    // or
	PrecAnd                   // and
	PrecEquality              // == !=
	PrecComparison            // < > <= >=
	PrecTerm                  // + -
	PrecFactor                // * /
	PrecUnary                 // ! -
	PrecCall                  // . ()
	PrecPrimary
)

type ParseFn func()
type ParseRule struct {
	prefix ParseFn
	infix  ParseFn
	prec   Precedence
}

func NewCompiler() *Compiler {
	c := &Compiler{}
	c.rules = []ParseRule{
		{c.grouping, nil, PrecNone},		// LParen
		{nil, nil, PrecNone},				// RParen
		{nil, nil, PrecNone},				// LBrace
		{nil, nil, PrecNone},				// RBrace
		{nil, nil, PrecNone},				// Comma
		{nil, nil, PrecNone},				// Dot
		{c.unary, c.binary, PrecTerm},	// Minus
		{nil, c.binary, PrecTerm},		// Plus
		{nil, nil, PrecNone},				// Semicolon
		{nil, c.binary, PrecFactor},		// Slash
		{nil, c.binary, PrecFactor},		// Star
		{nil, nil, PrecNone},				// Bang
		{nil, nil, PrecNone},				// BangEq
		{nil, nil, PrecNone},				// Equal
		{nil, nil, PrecNone},				// EqualEq
		{nil, nil, PrecNone},				// Greater
		{nil, nil, PrecNone},				// GreaterEq
		{nil, nil, PrecNone},				// Less
		{nil, nil, PrecNone},				// LessEq
		{nil, nil, PrecNone},				// Ident
		{nil, nil, PrecNone},				// String
		{c.number, nil, PrecNone},		// Number
		{nil, nil, PrecNone},				// And
		{nil, nil, PrecNone},				// Class
		{nil, nil, PrecNone},				// Else
		{nil, nil, PrecNone},				// False
		{nil, nil, PrecNone},				// For
		{nil, nil, PrecNone},				// Fun
		{nil, nil, PrecNone},				// If
		{nil, nil, PrecNone},				// Nil
		{nil, nil, PrecNone},				// Or
		{nil, nil, PrecNone},				// Print
		{nil, nil, PrecNone},				// Return
		{nil, nil, PrecNone},				// Super
		{nil, nil, PrecNone},				// This
		{nil, nil, PrecNone},				// True
		{nil, nil, PrecNone},				// Var
		{nil, nil, PrecNone},				// While
		{nil, nil, PrecNone},				// Error
		{nil, nil, PrecNone},				// Eof
	}

	return c
}

func (c *Compiler) Compile(source string, chunk *bc.Chunk) bool {
	c.scan = New(source)
	c.compilingChunk = chunk
	c.parser = Parser{}

	c.Advance()
	c.expression()
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

	if DEBUG_PRINT_CODE {
		if !c.parser.hadError {
			debug.DisassembleChunk(c.CurrentChunk(), "code")
		}
	}
}

func (c *Compiler) expression() {
	c.parsePrecedence(PrecAssign)
}

func (c *Compiler) number() {
	value, err := strconv.ParseFloat(c.parser.previous.Lexeme, 64)
	if err != nil {
		c.error("failed to parse number")
		return
	}

	c.emitConstant(bc.NumberValue{Value:value})
}

func (c *Compiler) grouping() {
	c.expression()
	c.Consume(TokenRParen, "Expect ')' after expression.")
}

func (c *Compiler) unary() {
	operType := c.parser.previous.Type

	c.parsePrecedence(PrecUnary)

	switch operType {
	case TokenMinus:
		c.emitBytes(bc.OpNegate)
	default:
		return // Unreachable?
	}
}

func (c *Compiler) binary() {
	operType := c.parser.previous.Type

	rule := c.rules[operType]
	c.parsePrecedence(rule.prec + 1)

	switch operType {
	case TokenPlus:
		c.emitBytes(bc.OpAdd)
	case TokenMinus:
		c.emitBytes(bc.OpSubtract)
	case TokenStar:
		c.emitBytes(bc.OpMultiply)
	case TokenSlash:
		fmt.Println("Emitting opDivide")
		c.emitBytes(bc.OpDivide)
	default:
		return
	}
}

func (c *Compiler) parsePrecedence(prec Precedence) {
	c.Advance()
	rule := c.rules[c.parser.previous.Type]
	if rule.prefix == nil {
		c.error("Expected expression.")
		return
	}

	rule.prefix()

	for prec <= c.rules[c.parser.current.Type].prec {
		c.Advance()
		infix := c.rules[c.parser.previous.Type].infix
		infix()
	}
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
