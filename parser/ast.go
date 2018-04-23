package parser

import "github.com/butlermatt/glpc/lexer"

type Expr interface {
	Accept(ExprVisitor) (interface{}, error)
}

type Stmt interface {
	Accept(StmtVisitor) error
}

type AssignExpr struct {
	Name  *lexer.Token
	Value Expr
}

func (a *AssignExpr) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.VisitAssignExpr(a)
}

type BinaryExpr struct {
	Left     Expr
	Operator *lexer.Token
	Right    Expr
}

func (b *BinaryExpr) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.VisitBinaryExpr(b)
}

type CallExpr struct {
	Callee Expr
	Paren  *lexer.Token
	Args   []Expr
}

func (c *CallExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitCallExpr(c) }

type GetExpr struct {
	Object Expr
	Name   *lexer.Token
}

func (g *GetExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitGetExpr(g) }

type GroupingExpr struct {
	Expression Expr
}

func (g *GroupingExpr) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.VisitGroupingExpr(g)
}

type LiteralExpr struct {
	Value interface{}
}

func (l *LiteralExpr) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.VisitLiteralExpr(l)
}

type LogicalExpr struct {
	Left     Expr
	Operator *lexer.Token
	Right    Expr
}

func (l *LogicalExpr) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.VisitLogicalExpr(l)
}

type SetExpr struct {
	Object Expr
	Name   *lexer.Token
	Value  Expr
}

func (s *SetExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitSetExpr(s) }

type ThisExpr struct {
	Keyword *lexer.Token
}

func (t *ThisExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitThisExpr(t) }

type UnaryExpr struct {
	Operator *lexer.Token
	Right    Expr
}

func (u *UnaryExpr) Accept(visitor ExprVisitor) (interface{}, error) { return visitor.VisitUnaryExpr(u) }

type VariableExpr struct {
	Name *lexer.Token
}

func (v *VariableExpr) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.VisitVariableExpr(v)
}

type ExprVisitor interface {
	VisitAssignExpr(expr *AssignExpr) (interface{}, error)
	VisitBinaryExpr(expr *BinaryExpr) (interface{}, error)
	VisitCallExpr(expr *CallExpr) (interface{}, error)
	VisitGetExpr(expr *GetExpr) (interface{}, error)
	VisitGroupingExpr(expr *GroupingExpr) (interface{}, error)
	VisitLiteralExpr(expr *LiteralExpr) (interface{}, error)
	VisitLogicalExpr(expr *LogicalExpr) (interface{}, error)
	VisitSetExpr(expr *SetExpr) (interface{}, error)
	VisitThisExpr(expr *ThisExpr) (interface{}, error)
	VisitUnaryExpr(expr *UnaryExpr) (interface{}, error)
	VisitVariableExpr(expr *VariableExpr) (interface{}, error)
}
type BlockStmt struct {
	Statements []Stmt
}

func (b *BlockStmt) Accept(visitor StmtVisitor) error { return visitor.VisitBlockStmt(b) }

type ClassStmt struct {
	Name    *lexer.Token
	Methods []*FunctionStmt
}

func (c *ClassStmt) Accept(visitor StmtVisitor) error { return visitor.VisitClassStmt(c) }

type ExpressionStmt struct {
	Expression Expr
}

func (e *ExpressionStmt) Accept(visitor StmtVisitor) error { return visitor.VisitExpressionStmt(e) }

type FunctionStmt struct {
	Name       *lexer.Token
	Parameters []*lexer.Token
	Body       []Stmt
}

func (f *FunctionStmt) Accept(visitor StmtVisitor) error { return visitor.VisitFunctionStmt(f) }

type IfStmt struct {
	Condition Expr
	Then      Stmt
	Else      Stmt
}

func (i *IfStmt) Accept(visitor StmtVisitor) error { return visitor.VisitIfStmt(i) }

type PrintStmt struct {
	Expression Expr
}

func (p *PrintStmt) Accept(visitor StmtVisitor) error { return visitor.VisitPrintStmt(p) }

type ReturnStmt struct {
	Keyword *lexer.Token
	Value   Expr
}

func (r *ReturnStmt) Accept(visitor StmtVisitor) error { return visitor.VisitReturnStmt(r) }

type VarStmt struct {
	Name        *lexer.Token
	Initializer Expr
}

func (v *VarStmt) Accept(visitor StmtVisitor) error { return visitor.VisitVarStmt(v) }

type ForStmt struct {
	Initializer Stmt
	Condition   Expr
	Body        Stmt
	Increment   Expr
}

func (f *ForStmt) Accept(visitor StmtVisitor) error { return visitor.VisitForStmt(f) }

type BreakStmt struct {
	Keyword *lexer.Token
}

func (b *BreakStmt) Accept(visitor StmtVisitor) error { return visitor.VisitBreakStmt(b) }

type ContinueStmt struct {
	Keyword *lexer.Token
}

func (c *ContinueStmt) Accept(visitor StmtVisitor) error { return visitor.VisitContinueStmt(c) }

type StmtVisitor interface {
	VisitBlockStmt(stmt *BlockStmt) error
	VisitClassStmt(stmt *ClassStmt) error
	VisitExpressionStmt(stmt *ExpressionStmt) error
	VisitFunctionStmt(stmt *FunctionStmt) error
	VisitIfStmt(stmt *IfStmt) error
	VisitPrintStmt(stmt *PrintStmt) error
	VisitReturnStmt(stmt *ReturnStmt) error
	VisitVarStmt(stmt *VarStmt) error
	VisitForStmt(stmt *ForStmt) error
	VisitBreakStmt(stmt *BreakStmt) error
	VisitContinueStmt(stmt *ContinueStmt) error
}
