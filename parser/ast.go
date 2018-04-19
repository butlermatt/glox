package parser

import "github.com/butlermatt/glpc/lexer"

type Expr interface {
	Accept(ExprVisitor) (interface{}, error)
}

type Stmt interface {
	Accept(StmtVisitor) error
}


type AssignExpr struct {
	Name *lexer.Token
	Value Expr
}

func (a *AssignExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitAssignExpr(a) }

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

type VariableExpr struct {
	Name *lexer.Token
}

func (v *VariableExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitVariableExpr(v) }


type ExprVisitor interface {
	VisitAssignExpr(expr *AssignExpr) (interface{}, error)
	VisitBinaryExpr(expr *BinaryExpr) (interface{}, error)
	VisitGroupingExpr(expr *GroupingExpr) (interface{}, error)
	VisitLiteralExpr(expr *LiteralExpr) (interface{}, error)
	VisitUnaryExpr(expr *UnaryExpr) (interface{}, error)
	VisitVariableExpr(expr *VariableExpr) (interface{}, error)
}
type BlockStmt struct {
	Statements []Stmt
}

func (b *BlockStmt) Accept(visitor StmtVisitor) error { return visitor.VisitBlockStmt(b) }

type ExpressionStmt struct {
	Expression Expr
}

func (e *ExpressionStmt) Accept(visitor StmtVisitor) error { return visitor.VisitExpressionStmt(e) }

type PrintStmt struct {
	Expression Expr
}

func (p *PrintStmt) Accept(visitor StmtVisitor) error { return visitor.VisitPrintStmt(p) }

type VarStmt struct {
	Name *lexer.Token
	Initializer Expr
}

func (v *VarStmt) Accept(visitor StmtVisitor) error { return visitor.VisitVarStmt(v) }


type StmtVisitor interface {
	VisitBlockStmt(stmt *BlockStmt) error
	VisitExpressionStmt(stmt *ExpressionStmt) error
	VisitPrintStmt(stmt *PrintStmt) error
	VisitVarStmt(stmt *VarStmt) error
}
