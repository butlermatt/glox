package parser

import "github.com/butlermatt/glpc/lexer"

type Expr interface {
	Accept(ExprVisitor) (interface{}, error)
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

func (b *Binary) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitBinary(b) }

type Grouping struct {
	Expression Expr
}

func (g *Grouping) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitGrouping(g) }

type Literal struct {
	Value interface{}
}

func (l *Literal) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitLiteral(l) }

type Unary struct {
	Operator *lexer.Token
	Right    Expr
}

func (u *Unary) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitUnary(u) }

type ExprVisitor interface {
	VisitBinary(binary *Binary) (interface{}, error)
	VisitGrouping(grouping *Grouping) (interface{}, error)
	VisitLiteral(literal *Literal) (interface{}, error)
	VisitUnary(unary *Unary) (interface{}, error)
}
