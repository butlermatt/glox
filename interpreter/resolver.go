package interpreter

import (
	"github.com/butlermatt/glox/lexer"
	"github.com/butlermatt/glox/parser"
)

type FunctionType int
type ClassType int

const (
	NoneFT FunctionType = iota
	FuncFT
	InitializerFT
	MethodFT
)

const (
	NoneCT ClassType = iota
	ClassCT
	SubsclassCT
)

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{interpreter: interpreter, curFunc: NoneFT}
}

type Resolver struct {
	interpreter *Interpreter
	stack       []map[string]bool
	curFunc     FunctionType
	curClass    ClassType
	inLoop      bool
}

func (r *Resolver) beginScope() {
	r.stack = append(r.stack, make(map[string]bool))
}

func (r *Resolver) endScope() {
	if len(r.stack) == 0 {
		return
	}
	r.stack = r.stack[:len(r.stack)-1]
}

func (r *Resolver) peekScope() map[string]bool {
	if len(r.stack) == 0 {
		return nil
	}

	return r.stack[len(r.stack)-1]
}

func (r *Resolver) VisitBlockStmt(stmt *parser.BlockStmt) error {
	r.beginScope()
	err := r.Resolve(stmt.Statements)
	r.endScope()
	return err
}

func (r *Resolver) VisitClassStmt(stmt *parser.ClassStmt) error {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	enclosingClass := r.curClass
	r.curClass = ClassCT

	if stmt.Superclass != nil {
		r.curClass = SubsclassCT
		err := r.resolveExpr(stmt.Superclass)
		if err != nil {
			return err
		}
		r.beginScope()
		sc := r.peekScope()
		sc["super"] = true
	}

	r.beginScope()
	scope := r.peekScope()
	scope["this"] = true

	var err error
	for _, method := range stmt.Methods {
		declaration := MethodFT
		if method.Name.Lexeme == "init" {
			declaration = InitializerFT
		}
		err = r.resolveFunction(method, declaration)
		if err != nil {
			break
		}
	}

	r.endScope()

	if stmt.Superclass != nil {
		r.endScope()
	}

	r.curClass = enclosingClass
	return err
}

func (r *Resolver) VisitExpressionStmt(stmt *parser.ExpressionStmt) error {
	return r.resolveExpr(stmt.Expression)
}

func (r *Resolver) VisitIfStmt(stmt *parser.IfStmt) error {
	err := r.resolveExpr(stmt.Condition)
	if err != nil {
		return err
	}
	err = r.resolveStmt(stmt.Then)
	if err != nil {
		return err
	}
	if stmt.Else != nil {
		return r.resolveStmt(stmt.Else)
	}
	return nil
}

func (r *Resolver) VisitFunctionStmt(stmt *parser.FunctionStmt) error {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	return r.resolveFunction(stmt, FuncFT)
}

func (r *Resolver) VisitPrintStmt(stmt *parser.PrintStmt) error {
	return r.resolveExpr(stmt.Expression)
}

func (r *Resolver) VisitReturnStmt(stmt *parser.ReturnStmt) error {
	if r.curFunc == NoneFT {
		return newError(stmt.Keyword, "Cannot return from top-level code.")
	}

	if stmt.Value != nil {
		if r.curFunc == InitializerFT {
			return newError(stmt.Keyword, "Cannot return a value from an initializer.")
		}
		return r.resolveExpr(stmt.Value)
	}
	return nil
}

func (r *Resolver) VisitForStmt(stmt *parser.ForStmt) error {
	r.beginScope()
	err := r.resolveStmt(stmt.Initializer)
	if err != nil {
		return err
	}
	err = r.resolveExpr(stmt.Condition)
	if err != nil {
		return err
	}

	oldLoop := r.inLoop
	r.inLoop = true
	err = r.resolveStmt(stmt.Body)
	if err != nil {
		r.inLoop = oldLoop
		return err
	}
	r.inLoop = oldLoop
	err = r.resolveExpr(stmt.Increment)
	r.endScope()
	return err
}

func (r *Resolver) VisitVarStmt(stmt *parser.VarStmt) error {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		err := r.resolveExpr(stmt.Initializer)
		if err != nil {
			return err
		}
	}
	r.define(stmt.Name)
	return nil
}

func (r *Resolver) VisitBreakStmt(stmt *parser.BreakStmt) error {
	if !r.inLoop {
		return newError(stmt.Keyword, "Cannot break when not in loop.")
	}
	return nil
}

func (r *Resolver) VisitContinueStmt(stmt *parser.ContinueStmt) error {
	if !r.inLoop {
		return newError(stmt.Keyword, "Cannot continue when not in loop.")
	}
	return nil
}

func (r *Resolver) VisitArrayExpr(expr *parser.ArrayExpr) (interface{}, error) {
	for _, value := range expr.Values {
		err := r.resolveExpr(value)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (r *Resolver) VisitAssignExpr(expr *parser.AssignExpr) (interface{}, error) {
	err := r.resolveExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	r.resolveLocal(expr, expr.Name)
	return nil, nil
}

func (r *Resolver) VisitBinaryExpr(expr *parser.BinaryExpr) (interface{}, error) {
	err := r.resolveExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	err = r.resolveExpr(expr.Right)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *Resolver) VisitCallExpr(expr *parser.CallExpr) (interface{}, error) {
	err := r.resolveExpr(expr.Callee)
	if err != nil {
		return nil, err
	}
	for _, e := range expr.Args {
		err = r.resolveExpr(e)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (r *Resolver) VisitGetExpr(expr *parser.GetExpr) (interface{}, error) {
	err := r.resolveExpr(expr.Object)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *Resolver) VisitGroupingExpr(expr *parser.GroupingExpr) (interface{}, error) {
	return nil, r.resolveExpr(expr.Expression)
}

func (r *Resolver) VisitIndexExpr(expr *parser.IndexExpr) (interface{}, error) {
	err := r.resolveExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	err = r.resolveExpr(expr.Right)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *Resolver) VisitLiteralExpr(expr *parser.LiteralExpr) (interface{}, error) {
	return nil, nil
}

func (r *Resolver) VisitLogicalExpr(expr *parser.LogicalExpr) (interface{}, error) {
	err := r.resolveExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	err = r.resolveExpr(expr.Right)
	return nil, err
}

func (r *Resolver) VisitSetExpr(expr *parser.SetExpr) (interface{}, error) {
	err := r.resolveExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	err = r.resolveExpr(expr.Object)
	return nil, err
}

func (r *Resolver) VisitSuperExpr(expr *parser.SuperExpr) (interface{}, error) {
	if r.curClass == NoneCT {
		return nil, newError(expr.Keyword, "Cannot use 'super' outside of a class.")
	} else if r.curClass != SubsclassCT {
		return nil, newError(expr.Keyword, "Cannot use 'super' in a class with no superclass.")
	}
	r.resolveLocal(expr, expr.Keyword)
	return nil, nil
}

func (r *Resolver) VisitThisExpr(expr *parser.ThisExpr) (interface{}, error) {
	if r.curClass == NoneCT {
		return nil, newError(expr.Keyword, "Cannot use 'this' outside of a class.")
	}
	r.resolveLocal(expr, expr.Keyword)
	return nil, nil
}

func (r *Resolver) VisitUnaryExpr(expr *parser.UnaryExpr) (interface{}, error) {
	return nil, r.resolveExpr(expr.Right)
}

func (r *Resolver) VisitVariableExpr(expr *parser.VariableExpr) (interface{}, error) {
	scope := r.peekScope()
	if scope != nil {
		if v, ok := scope[expr.Name.Lexeme]; ok && v == false {
			return nil, newError(expr.Name, "Cannot read local variable in its own initializer")
		}
	}

	r.resolveLocal(expr, expr.Name)
	return nil, nil
}

func (r *Resolver) Resolve(stmts []parser.Stmt) error {
	for _, stmt := range stmts {
		err := r.resolveStmt(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) resolveStmt(stmt parser.Stmt) error {
	return stmt.Accept(r)
}

func (r *Resolver) resolveExpr(expr parser.Expr) error {
	_, err := expr.Accept(r)
	return err
}

func (r *Resolver) resolveLocal(expr parser.Expr, name *lexer.Token) {
	for i := len(r.stack) - 1; i >= 0; i-- {
		if _, ok := r.stack[i][name.Lexeme]; ok {
			r.interpreter.resolve(expr, len(r.stack)-1-i)
			return
		}
	}
	// Not found assume it's global
}

func (r *Resolver) resolveFunction(function *parser.FunctionStmt, fnType FunctionType) error {
	enclosingFun := r.curFunc
	r.curFunc = fnType
	r.beginScope()
	for _, param := range function.Parameters {
		r.declare(param)
		r.define(param)
	}
	err := r.Resolve(function.Body)
	r.endScope()
	r.curFunc = enclosingFun
	return err
}

func (r *Resolver) declare(name *lexer.Token) error {
	scope := r.peekScope()
	if scope == nil {
		return nil
	}
	if _, ok := scope[name.Lexeme]; ok {
		return newError(name, "Variable with this name already declared in this scope.")
	}

	scope[name.Lexeme] = false
	return nil
}

func (r *Resolver) define(name *lexer.Token) {
	scope := r.peekScope()
	if scope == nil {
		return
	}

	scope[name.Lexeme] = true
}
