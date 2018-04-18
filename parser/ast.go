package parser

import "github.com/butlermatt/glpc/lexer"

type Expr interface {
	Accept(ExprVisitor) (interface{}, error)
}

type Stmt interface {
	Accept(StmtVisitor) (interface{}, error)
}


type BinaryExpr struct {
	Left Expr
	Operator *lexer.Token
	Right Expr
}

func (b *BinaryExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitBinaryExpr(b) }
type GroupingExpr struct {
	Expression Expr
}

func (g *GroupingExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitGroupingExpr(g) }
type LiteralExpr struct {
	Value interface{}
}

func (l *LiteralExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitLiteralExpr(l) }
type UnaryExpr struct {
	Operator *lexer.Token
	Right Expr
}

func (u *UnaryExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitUnaryExpr(u) }

type ExprVisitor interface {
	VisitBinaryExpr(expr *BinaryExpr) (interface{}, error)
	VisitGroupingExpr(expr *GroupingExpr) (interface{}, error)
	VisitLiteralExpr(expr *LiteralExpr) (interface{}, error)
	VisitUnaryExpr(expr *UnaryExpr) (interface{}, error)
}
type ExpressionStmt struct {
	Expression Expr
}

func (e *ExpressionStmt) Accept(visitor StmtVisitor) (interface{}, error) { return visitor.VisitExpressionStmt(e) }
type PrintStmt struct {
	Expression Expr
}

func (p *PrintStmt) Accept(visitor StmtVisitor) (interface{}, error) { return visitor.VisitPrintStmt(p) }

type StmtVisitor interface {
	VisitExpressionStmt(stmt *ExpressionStmt) (interface{}, error)
	VisitPrintStmt(stmt *PrintStmt) (interface{}, error)
}
