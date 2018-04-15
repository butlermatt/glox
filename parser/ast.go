package parser

import "github.com/butlermatt/glpc/lexer"

type Expr interface {
	Accept(ExprVisitor) interface{}
}

//
//type Stmt interface {
//	Accept(StmtVisitor) interface{}
//}

type Binary struct {
	Left     Expr
	Operator *lexer.Token
	Right    Expr
}

func (b *Binary) Accept(visitor ExprVisitor) interface{} { return visitor.VisitBinary(b) }

type Grouping struct {
	Expression Expr
}

func (g *Grouping) Accept(visitor ExprVisitor) interface{} { return visitor.VisitGrouping(g) }

type Literal struct {
	Value interface{}
}

func (l *Literal) Accept(visitor ExprVisitor) interface{} { return visitor.VisitLiteral(l) }

type Unary struct {
	Operator *lexer.Token
	Right    Expr
}

func (u *Unary) Accept(visitor ExprVisitor) interface{} { return visitor.VisitUnary(u) }

type ExprVisitor interface {
	VisitBinary(binary *Binary) interface{}
	VisitGrouping(grouping *Grouping) interface{}
	VisitLiteral(literal *Literal) interface{}
	VisitUnary(unary *Unary) interface{}
}
