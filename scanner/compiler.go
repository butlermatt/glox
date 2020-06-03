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
const LOCALS_MAX = 256

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
	strings        *bc.Table

	locals     [LOCALS_MAX]Local
	localCount int
	scopeDepth int
}

type Local struct {
	name  Token
	depth int
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

type ParseFn func(canAssign bool)
type ParseRule struct {
	prefix ParseFn
	infix  ParseFn
	prec   Precedence
}

func NewCompiler() *Compiler {
	c := &Compiler{}
	c.rules = []ParseRule{
		{c.grouping, nil, PrecNone},   // LParen
		{nil, nil, PrecNone},          // RParen
		{nil, nil, PrecNone},          // LBrace
		{nil, nil, PrecNone},          // RBrace
		{nil, nil, PrecNone},          // Comma
		{nil, nil, PrecNone},          // Dot
		{c.unary, c.binary, PrecTerm}, // Minus
		{nil, c.binary, PrecTerm},     // Plus
		{nil, nil, PrecNone},          // Semicolon
		{nil, c.binary, PrecFactor},   // Slash
		{nil, c.binary, PrecFactor},   // Star

		{c.unary, nil, PrecNone},        // Bang
		{nil, c.binary, PrecEquality},   // BangEq
		{nil, nil, PrecNone},            // Equal
		{nil, c.binary, PrecEquality},   // EqualEq
		{nil, c.binary, PrecComparison}, // Greater
		{nil, c.binary, PrecComparison}, // GreaterEq
		{nil, c.binary, PrecComparison}, // Less
		{nil, c.binary, PrecComparison}, // LessEq

		{c.variable, nil, PrecNone}, // Ident
		{c.string, nil, PrecNone},   // String
		{c.number, nil, PrecNone},   // Number

		{nil, nil, PrecNone},       // And
		{nil, nil, PrecNone},       // Class
		{nil, nil, PrecNone},       // Else
		{c.literal, nil, PrecNone}, // False
		{nil, nil, PrecNone},       // For
		{nil, nil, PrecNone},       // Fun
		{nil, nil, PrecNone},       // If
		{c.literal, nil, PrecNone}, // Nil
		{nil, nil, PrecNone},       // Or
		{nil, nil, PrecNone},       // Print
		{nil, nil, PrecNone},       // Return
		{nil, nil, PrecNone},       // Super
		{nil, nil, PrecNone},       // This
		{c.literal, nil, PrecNone}, // True
		{nil, nil, PrecNone},       // Var
		{nil, nil, PrecNone},       // While

		{nil, nil, PrecNone}, // Error
		{nil, nil, PrecNone}, // Eof
	}

	return c
}

func (c *Compiler) Compile(tbl *bc.Table, source string, chunk *bc.Chunk) bool {
	c.scan = New(source)
	c.compilingChunk = chunk
	c.parser = Parser{}
	c.strings = tbl

	c.Advance()

	for !c.match(TokenEof) {
		c.declaration()
	}

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

func (c *Compiler) beginScope() {
	c.scopeDepth++
}

func (c *Compiler) endScope() {
	c.scopeDepth--

	for c.localCount > 0 && c.locals[c.localCount - 1].depth > c.scopeDepth {
		c.emitBytes(bc.OpPop)
		c.localCount--
	}
}

func (c *Compiler) endCompiler() {
	c.emitReturn()

	if DEBUG_PRINT_CODE {
		if !c.parser.hadError {
			debug.DisassembleChunk(c.CurrentChunk(), "code")
		}
	}
}

func (c *Compiler) declaration() {
	if c.match(TokenVar) {
		c.varDeclaration()
	} else {
		c.statement()
	}

	if c.parser.panicMode {
		c.synchronize()
	}
}

func (c *Compiler) varDeclaration() {
	global := c.parseVariable("Expect variable name.")

	if c.match(TokenEqual) {
		c.expression()
	} else {
		c.emitBytes(bc.OpNil)
	}

	c.Consume(TokenSemicolon, "Expect ';' after variable declaration.")
	c.defineVariable(global)
}

func (c *Compiler) declareVariable() {
	if c.scopeDepth == 0 {
		return
	}

	name := &c.parser.previous

	for i := c.localCount - 1; i >= 0; i-- {
		local := &c.locals[i]
		if local.depth != -1 && local.depth < c.scopeDepth {
			break
		}

		if name.Lexeme == local.name.Lexeme {
			c.error("Variable with this name already declared in this scope")
		}
	}

	c.addLocal(*name)
}

func (c *Compiler) defineVariable(global byte) {
	if c.scopeDepth > 0 {
		// Don't emit bytes for local variables
		c.markInitialized()
		return
	}

	c.emitBytes(bc.OpDefineGlobal, global)
}

func (c *Compiler) statement() {
	switch {
	case c.match(TokenPrint):
		c.printStatement()
	case c.match(TokenLBrace):
		c.beginScope()
		c.block()
		c.endScope()
	default:
		c.expressionStatement()
	}
}

func (c *Compiler) printStatement() {
	c.expression()
	c.Consume(TokenSemicolon, "Expect ';' after value.")
	c.emitBytes(bc.OpPrint)
}

func (c *Compiler) expressionStatement() {
	c.expression()
	c.Consume(TokenSemicolon, "Expect ';' after expression.")
	c.emitBytes(bc.OpPop)
}

func (c *Compiler) block() {
	for !c.check(TokenRBrace) && !c.check(TokenEof) {
		c.declaration()
	}

	c.Consume(TokenRBrace, "Expect '}' after block.")
}

func (c *Compiler) expression() {
	c.parsePrecedence(PrecAssign)
}

func (c *Compiler) number(_ bool) {
	value, err := strconv.ParseFloat(c.parser.previous.Lexeme, 64)
	if err != nil {
		c.error("failed to parse number")
		return
	}

	c.emitConstant(bc.NumberValue{Value: value})
}

func (c *Compiler) grouping(_ bool) {
	c.expression()
	c.Consume(TokenRParen, "Expect ')' after expression.")
}

func (c *Compiler) unary(_ bool) {
	operType := c.parser.previous.Type

	c.parsePrecedence(PrecUnary)

	switch operType {
	case TokenBang:
		c.emitBytes(bc.OpNot)
	case TokenMinus:
		c.emitBytes(bc.OpNegate)
	default:
		return // Unreachable?
	}
}

func (c *Compiler) binary(_ bool) {
	operType := c.parser.previous.Type

	rule := c.rules[operType]
	c.parsePrecedence(rule.prec + 1)

	switch operType {
	case TokenBangEq:
		c.emitBytes(bc.OpEqual, byte(bc.OpNot))
	case TokenEqualEq:
		c.emitBytes(bc.OpEqual)
	case TokenGreater:
		c.emitBytes(bc.OpGreater)
	case TokenGreaterEq:
		c.emitBytes(bc.OpLess, byte(bc.OpNot))
	case TokenLess:
		c.emitBytes(bc.OpLess)
	case TokenLessEq:
		c.emitBytes(bc.OpGreater, byte(bc.OpNot))
	case TokenPlus:
		c.emitBytes(bc.OpAdd)
	case TokenMinus:
		c.emitBytes(bc.OpSubtract)
	case TokenStar:
		c.emitBytes(bc.OpMultiply)
	case TokenSlash:
		c.emitBytes(bc.OpDivide)
	default:
		return
	}
}

func (c *Compiler) literal(_ bool) {
	switch c.parser.previous.Type {
	case TokenFalse:
		c.emitBytes(bc.OpFalse)
	case TokenNil:
		c.emitBytes(bc.OpNil)
	case TokenTrue:
		c.emitBytes(bc.OpTrue)
	default:
		return // unreachable
	}
}

func (c *Compiler) string(_ bool) {
	str := c.parser.previous.Lexeme[1 : len(c.parser.previous.Lexeme)-1]
	sobj := bc.StringAsValue(c.strings, str)
	c.emitConstant(sobj)
}

func (c *Compiler) variable(canAssign bool) {
	c.namedVariable(c.parser.previous, canAssign)
}

func (c *Compiler) namedVariable(name Token, canAssign bool) {
	var getOp, setOp bc.OpCode
	arg := c.resolveLocal(&name)
	if arg != -1 {
		getOp = bc.OpGetLocal
		setOp = bc.OpSetLocal
	} else {
		arg = c.identifierConstant(&name)
		getOp = bc.OpGetGlobal
		setOp = bc.OpSetGlobal
	}

	if canAssign && c.match(TokenEqual) {
		c.expression()
		c.emitBytes(setOp, byte(arg))
	} else {
		c.emitBytes(getOp, byte(arg))
	}
}

func (c *Compiler) parsePrecedence(prec Precedence) {
	c.Advance()
	rule := c.rules[c.parser.previous.Type]
	if rule.prefix == nil {
		c.error("Expected expression.")
		return
	}

	canAssign := prec <= PrecAssign
	rule.prefix(canAssign)

	for prec <= c.rules[c.parser.current.Type].prec {
		c.Advance()
		infix := c.rules[c.parser.previous.Type].infix
		infix(canAssign)
	}

	if canAssign && c.match(TokenEqual) {
		c.error("Invalid assignment target.")
	}
}

func (c *Compiler) parseVariable(msg string) byte {
	c.Consume(TokenIdent, msg)

	c.declareVariable()
	if c.scopeDepth > 0 {
		return 0
	}
	return byte(c.identifierConstant(&c.parser.previous))
}

func (c *Compiler) markInitialized() {
	c.locals[c.localCount - 1].depth = c.scopeDepth
}

func (c *Compiler) identifierConstant(token *Token) int {
	return int(c.MakeConstant(bc.StringAsValue(c.strings, token.Lexeme)))
}

func (c *Compiler) addLocal(name Token) {
	if c.localCount == LOCALS_MAX {
		c.error("Too many local variables in scope.")
		return
	}

	local := &c.locals[c.localCount]
	c.localCount++

	local.name = name
	local.depth = -1
}

func (c *Compiler) match(tt TokenType) bool {
	if !c.check(tt) {
		return false
	}
	c.Advance()
	return true
}

func (c *Compiler) check(tt TokenType) bool {
	return c.parser.current.Type == tt
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

func (c *Compiler) resolveLocal(name *Token) int {
	for i := c.localCount - 1; i >= 0; i-- {
		if local := &c.locals[i]; local.name.Lexeme == name.Lexeme {
			if local.depth == -1 {
				c.error("Cannot read local variable in its own initializer.")
			}
			return i
		}
	}

	return -1
}

func (c *Compiler) synchronize() {
	c.parser.panicMode = false

	for c.parser.current.Type != TokenEof {
		if c.parser.previous.Type == TokenSemicolon {
			return
		}

		switch c.parser.current.Type {
		case TokenClass, TokenFun, TokenVar, TokenFor, TokenIf, TokenWhile, TokenPrint, TokenReturn:
			return
		default:
			// do nothing
		}

		c.Advance()
	}
}
