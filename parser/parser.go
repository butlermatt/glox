package parser

import (
	"github.com/butlermatt/glpc/lexer"
)

type ParseError struct {
	Line  int
	Where string
	Msg   string
}

type Parser struct {
	l       *lexer.Lexer
	curTok  *lexer.Token
	prevTok *lexer.Token
	errors  []ParseError
}

func New(lexer *lexer.Lexer) *Parser {
	lexer.ScanTokens()
	p := &Parser{l: lexer, errors: []ParseError{}}
	p.nextToken()

	return p
}

func (p *Parser) Errors() []ParseError {
	return p.errors
}

func (p *Parser) Parse() []Stmt {
	var stmts []Stmt
	for p.curTok.Type != lexer.EOF {
		stmts = append(stmts, p.declaration())
	}

	return stmts
}

func (p *Parser) addError(token *lexer.Token, message string) {
	if token.Type == lexer.EOF {
		p.errors = append(p.errors, ParseError{Line: token.Line, Where: "at end", Msg: message})
	} else {
		p.errors = append(p.errors, ParseError{Line: token.Line, Where: token.Lexeme, Msg: message})
	}
}

func (p *Parser) nextToken() {
	if p.curTok == nil || p.curTok.Type != lexer.EOF {
		p.prevTok = p.curTok
		p.curTok = p.l.NextToken()
	}
}

func (p *Parser) check(t lexer.TokenType) bool {
	if p.curTok.Type == lexer.EOF {
		return false
	}
	return p.curTok.Type == t
}

// match will check if the next token matches the provided types. If it does, then it will advance the token.
func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, tt := range types {
		if p.check(tt) {
			p.nextToken()
			return true
		}
	}

	return false
}

func (p *Parser) consume(tt lexer.TokenType, message string) bool {
	if p.check(tt) {
		p.nextToken()
		return true
	}

	p.addError(p.curTok, message)
	return false
}

func (p *Parser) expression() Expr {
	return p.assignment()
}

func (p *Parser) assignment() Expr {
	expr := p.equality()

	if p.match(lexer.Equal) {
		equals := p.prevTok
		value := p.assignment()

		if e, ok := expr.(*VariableExpr); ok {
			return &AssignExpr{Name: e.Name, Value: value}
		}

		p.addError(equals, "Invalid assignment target.")
		return nil
	}

	return expr
}

func (p *Parser) equality() Expr {
	exp := p.comparison()

	for p.match(lexer.BangEq, lexer.EqualEq) {
		if exp == nil {
			return exp
		}
		oper := p.prevTok
		right := p.comparison()
		if right == nil {
			return nil
		}
		exp = &BinaryExpr{Left: exp, Operator: oper, Right: right}
	}

	return exp
}

func (p *Parser) comparison() Expr {
	exp := p.addition()

	for p.match(lexer.Greater, lexer.GreaterEq, lexer.Less, lexer.LessEq) {
		if exp == nil {
			return exp
		}
		oper := p.prevTok
		right := p.addition()
		if right == nil {
			return nil
		}
		exp = &BinaryExpr{Left: exp, Operator: oper, Right: right}
	}

	return exp
}

func (p *Parser) addition() Expr {
	exp := p.multiplication()

	for p.match(lexer.Plus, lexer.Minus) {
		if exp == nil {
			return exp
		}
		oper := p.prevTok
		right := p.multiplication()
		if right == nil {
			return nil
		}
		exp = &BinaryExpr{Left: exp, Operator: oper, Right: right}
	}

	return exp
}

func (p *Parser) multiplication() Expr {
	exp := p.unary()

	for p.match(lexer.Slash, lexer.Star) {
		if exp == nil {
			return exp
		}
		oper := p.prevTok
		right := p.unary()
		if right == nil {
			return nil
		}
		exp = &BinaryExpr{Left: exp, Operator: oper, Right: right}
	}

	return exp
}

func (p *Parser) unary() Expr {
	if p.match(lexer.Bang, lexer.Minus) {
		oper := p.prevTok
		right := p.unary()
		if right == nil {
			return nil
		}
		return &UnaryExpr{Operator: oper, Right: right}
	}

	return p.primary()
}

func (p *Parser) primary() Expr {
	switch {
	case p.match(lexer.False):
		return &LiteralExpr{Value: false}
	case p.match(lexer.True):
		return &LiteralExpr{Value: true}
	case p.match(lexer.Null):
		return &LiteralExpr{Value: nil}
	case p.match(lexer.Number, lexer.String):
		return &LiteralExpr{Value: p.prevTok.Literal}
	case p.match(lexer.Ident):
		return &VariableExpr{Name: p.prevTok}
	case p.match(lexer.LParen):
		exp := p.expression()
		if exp == nil {
			return nil
		}
		if p.consume(lexer.RParen, "Expect ')' after expression.") {
			return &GroupingExpr{Expression: exp}
		}
	}

	p.addError(p.curTok, "Expect expression.")
	return nil
}

func (p *Parser) synchronize() {
	p.nextToken()

	for p.curTok.Type != lexer.EOF {
		if p.prevTok.Type == lexer.Semicolon {
			return
		}

		switch p.curTok.Type {
		case lexer.Class:
		case lexer.Fun:
		case lexer.Var:
		case lexer.For:
		case lexer.If:
		case lexer.While:
		case lexer.Print:
		case lexer.Return:
			return
		}

		p.nextToken()
	}
}

func (p *Parser) statement() Stmt {
	if p.match(lexer.Print) {
		return p.printStatement()
	}

	return p.expressionStatement()
}

func (p *Parser) printStatement() Stmt {
	value := p.expression()
	p.consume(lexer.Semicolon, "Expect ';' after value.")
	return &PrintStmt{Expression: value}
}

func (p *Parser) expressionStatement() Stmt {
	expr := p.expression()
	p.consume(lexer.Semicolon, "Expect ';' after value.")
	return &ExpressionStmt{Expression: expr}
}

func (p *Parser) declaration() Stmt {
	var stmt Stmt
	if p.match(lexer.Var) {
		stmt = p.varDeclaration()
	} else {
		stmt = p.statement()
	}

	if len(p.errors) > 0 {
		p.synchronize()
		return nil
	}
	return stmt
}

func (p *Parser) varDeclaration() Stmt {
	if !p.consume(lexer.Ident, "Expect variable name.") {
		return nil
	}
	name := p.prevTok

	var initializer Expr
	if p.match(lexer.Equal) {
		initializer = p.expression()
	}

	p.consume(lexer.Semicolon, "Expect ';' after variable declaration.")
	return &VarStmt{Name: name, Initializer: initializer}
}
