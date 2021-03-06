package interpreter

import (
	"errors"
	"fmt"
	"github.com/butlermatt/glox/lexer"
	"github.com/butlermatt/glox/parser"
	"time"
)

var BreakError = errors.New("Unexpected 'break' outside of loop")
var ContinueError = errors.New("Unexpected 'continue' outside of loop")

type RuntimeError struct {
	Token   *lexer.Token
	Message string
}

type ReturnError struct {
	RuntimeError
	Value interface{}
}

func NewReturnError(keyword *lexer.Token, value interface{}) *ReturnError {
	return &ReturnError{RuntimeError: RuntimeError{Token: keyword, Message: ""}, Value: value}
}

func (re *RuntimeError) Error() string {
	return fmt.Sprintf("[Runtime Error line %d] %s", re.Token.Line, re.Message)
}

type Interpreter struct {
	stmts       []parser.Stmt
	globals     *Environment
	environment *Environment
	locals      map[parser.Expr]int
}

func New(statements []parser.Stmt) *Interpreter {
	env := NewEnvironment()
	env.builtin("clock", &BuiltIn{
		arity: 0,
		callFn: func(interp *Interpreter, args []interface{}) (interface{}, error) {
			return float64(time.Now().Unix()), nil
		}},
	)
	return &Interpreter{stmts: statements, globals: env, environment: env, locals: make(map[parser.Expr]int)}
}

func (i *Interpreter) Interpret() error {
	for _, stmt := range i.stmts {
		err := i.execute(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) execute(stmt parser.Stmt) error {
	return stmt.Accept(i)
}

func (i *Interpreter) resolve(expr parser.Expr, depth int) {
	i.locals[expr] = depth
}

func (i *Interpreter) VisitArrayExpr(expr *parser.ArrayExpr) (interface{}, error) {
	var values []interface{}
	for _, val := range expr.Values {
		value, err := i.evaluate(val)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, nil
}

func (i *Interpreter) VisitBinaryExpr(binary *parser.BinaryExpr) (interface{}, error) {
	left, err := i.evaluate(binary.Left)
	if err != nil {
		return nil, err
	}
	right, err := i.evaluate(binary.Right)
	if err != nil {
		return nil, err
	}

	switch binary.Operator.Type {
	case lexer.Greater:
		l, r, err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return l > r, nil
	case lexer.GreaterEq:
		l, r, err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return l >= r, nil
	case lexer.Less:
		l, r, err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return l < r, nil
	case lexer.LessEq:
		l, r, err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return l <= r, nil
	case lexer.BangEq:
		b, err := isEqual(left, right)
		return !b, err
	case lexer.EqualEq:
		return isEqual(left, right)
	case lexer.Minus:
		l, r, err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return l - r, nil
	case lexer.Slash:
		l, r, err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		if r == 0 {
			return nil, newError(binary.Operator, "Division by zero.")
		}
		return l / r, nil
	case lexer.Star:
		l, r, err := checkNumberOperands(binary.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return l * r, nil
	case lexer.Plus:
		switch l := left.(type) {
		case float64:
			if r, ok := right.(float64); !ok {
				return nil, newError(binary.Operator, "Both operands must be of the same type.")
			} else {
				return l + r, nil
			}
		case string:
			if r, ok := right.(string); !ok {
				return nil, newError(binary.Operator, "Both operands must be of the same type.")
			} else {
				return l + r, nil
			}
		default:
			return nil, newError(binary.Operator, "Both operands must be a Number or a String.")
		}
	}

	return nil, nil // Should never reach
}

func (i *Interpreter) VisitGroupingExpr(grouping *parser.GroupingExpr) (interface{}, error) {
	return i.evaluate(grouping.Expression)
}

func (i *Interpreter) VisitIndexExpr(expr *parser.IndexExpr) (interface{}, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	l, err := checkSliceOperand(expr.Operator, left)
	if err != nil {
		return nil, err
	}
	r, err := checkNumberOperand(expr.Operator, right)
	if err != nil {
		return nil, err
	}
	ind := int(r)
	if ind >= len(l) {
		return nil, newError(expr.Operator, "Index out of range.")
	}
	return l[ind], nil
}

func (i *Interpreter) VisitLiteralExpr(literal *parser.LiteralExpr) (interface{}, error) {
	return literal.Value, nil
}
func (i *Interpreter) VisitUnaryExpr(unary *parser.UnaryExpr) (interface{}, error) {
	right, err := i.evaluate(unary.Right)
	if err != nil {
		return nil, err
	}

	switch unary.Operator.Type {
	case lexer.Minus:
		val, err := checkNumberOperand(unary.Operator, right)
		if err != nil {
			return nil, err
		}
		return -val, nil
	case lexer.Bang:
		return !isTruthy(right), nil
	}

	// Should never reach here.
	return nil, nil
}

func (i *Interpreter) VisitVariableExpr(expr *parser.VariableExpr) (interface{}, error) {
	return i.lookUpVariable(expr.Name, expr)
}

func (i *Interpreter) lookUpVariable(name *lexer.Token, expr parser.Expr) (interface{}, error) {
	if distance, ok := i.locals[expr]; ok {
		return i.environment.GetAt(distance, name)
	}
	return i.globals.Get(name)
}

func (i *Interpreter) VisitAssignExpr(expr *parser.AssignExpr) (interface{}, error) {
	value, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}
	if distance, ok := i.locals[expr]; ok {
		i.environment.AssignAt(distance, expr.Name, value)
	} else {
		i.globals.Assign(expr.Name, value)
	}

	return value, nil
}

func (i *Interpreter) VisitLogicalExpr(expr *parser.LogicalExpr) (interface{}, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	if expr.Operator.Type == lexer.Or {
		if isTruthy(left) {
			return left, nil
		}
	} else {
		if !isTruthy(left) {
			return left, nil
		}
	}

	return i.evaluate(expr.Right)
}

func (i *Interpreter) VisitSetExpr(expr *parser.SetExpr) (interface{}, error) {
	if ie, ok := expr.Object.(*parser.IndexExpr); ok {
		l, err := i.evaluate(ie.Left)
		if err != nil {
			return nil, err
		}
		arr, err := checkSliceOperand(ie.Operator, l)
		if err != nil {
			return nil, err
		}
		ind, err := i.evaluate(ie.Right)
		if err != nil {
			return nil, err
		}
		in, err := checkNumberOperand(ie.Operator, ind)
		if err != nil {
			return nil, err
		}
		index := int(in)
		val, err := i.evaluate(expr.Value)
		if err != nil {
			return nil, err
		}
		if index >= len(arr) {
			return nil, newError(ie.Operator, "Index out of range.")
		}
		arr[index] = val
		return val, nil
	}

	obj, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}

	if o, ok := obj.(*LoxInstance); ok {
		val, err := i.evaluate(expr.Value)
		if err != nil {
			return nil, err
		}
		o.Set(expr.Name, val)

		return val, nil
	}

	return nil, newError(expr.Name, "Only instances have fields.")
}

func (i *Interpreter) VisitSuperExpr(expr *parser.SuperExpr) (interface{}, error) {
	var superclass *LoxClass
	var object *LoxInstance
	dist := i.locals[expr]
	sc, err := i.environment.GetAt(dist, expr.Keyword)
	if err != nil {
		return nil, err
	}

	var ok bool
	if superclass, ok = sc.(*LoxClass); !ok {
		return nil, newError(expr.Keyword, "Superclass was not a LoxClass")
	}

	obj, err := i.environment.GetAt(dist-1, &lexer.Token{Type: lexer.This, Lexeme: "this"})
	if err != nil {
		return nil, err
	}
	if object, ok = obj.(*LoxInstance); !ok {
		return nil, newError(expr.Keyword, "this was not a LoxInstance")
	}

	method := superclass.findMethod(object, expr.Method.Lexeme)
	if method == nil {
		return nil, newError(expr.Method, "Undefined property '"+expr.Method.Lexeme+"'.")
	}
	return method, nil

}

func (i *Interpreter) VisitThisExpr(expr *parser.ThisExpr) (interface{}, error) {
	return i.lookUpVariable(expr.Keyword, expr)
}

func (i *Interpreter) VisitCallExpr(expr *parser.CallExpr) (interface{}, error) {
	callee, err := i.evaluate(expr.Callee)
	if err != nil {
		return nil, err
	}

	var args []interface{}
	for _, arg := range expr.Args {
		a, err := i.evaluate(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, a)
	}

	if function, ok := callee.(Callable); !ok {
		return nil, newError(expr.Paren, "Can only call functions and classes.")
	} else {
		if len(args) != function.Arity() {
			return nil, newError(expr.Paren, fmt.Sprintf("Expected %d arguments but got %d.", function.Arity(), len(args)))
		}
		return function.Call(i, args)
	}
}

func (i *Interpreter) VisitGetExpr(expr *parser.GetExpr) (interface{}, error) {
	obj, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}
	if li, ok := obj.(*LoxInstance); ok {
		return li.Get(expr.Name)
	}
	return nil, newError(expr.Name, "Only instances have properties.")
}

func (i *Interpreter) evaluate(expr parser.Expr) (interface{}, error) {
	return expr.Accept(i)
}

func (i *Interpreter) VisitExpressionStmt(stmt *parser.ExpressionStmt) error {
	_, err := i.evaluate(stmt.Expression)
	return err
}

func (i *Interpreter) VisitPrintStmt(stmt *parser.PrintStmt) error {
	val, err := i.evaluate(stmt.Expression)
	if err != nil {
		return err
	}
	fmt.Println(stringify(val))
	return nil
}

func (i *Interpreter) VisitVarStmt(stmt *parser.VarStmt) error {
	var value interface{}
	var err error
	if stmt.Initializer != nil {
		value, err = i.evaluate(stmt.Initializer)
		if err != nil {
			return err
		}
	}

	i.environment.Define(stmt.Name, value)
	return nil
}

func (i *Interpreter) VisitBlockStmt(stmt *parser.BlockStmt) error {
	return i.executeBlock(stmt.Statements, NewEnclosedEnvironment(i.environment))
}

func (i *Interpreter) VisitClassStmt(stmt *parser.ClassStmt) error {
	i.environment.Define(stmt.Name, nil)

	var sk *LoxClass
	if stmt.Superclass != nil {
		sc, err := i.evaluate(stmt.Superclass)
		if err != nil {
			return err
		}
		var ok bool
		if sk, ok = sc.(*LoxClass); !ok {
			return newError(stmt.Superclass.Name, "Superclass must be a class")
		}

		i.environment = NewEnclosedEnvironment(i.environment)
		i.environment.m["super"] = sk
	}

	var methods = make(map[string]*Function)
	for _, method := range stmt.Methods {
		methods[method.Name.Lexeme] = NewFunction(method, i.environment, method.Name.Lexeme == "init")
	}

	klass := NewClass(stmt.Name.Lexeme, sk, methods)
	if sk != nil {
		i.environment = i.environment.enclosing
	}

	i.environment.Assign(stmt.Name, klass)
	return nil
}

func (i *Interpreter) VisitIfStmt(stmt *parser.IfStmt) error {
	cond, err := i.evaluate(stmt.Condition)
	if err != nil {
		return err
	}

	if isTruthy(cond) {
		return i.execute(stmt.Then)
	}
	if stmt.Else != nil {
		return i.execute(stmt.Else)
	}

	return nil
}

func (i *Interpreter) VisitForStmt(stmt *parser.ForStmt) error {
	prev := i.environment
	i.environment = NewEnclosedEnvironment(i.environment)

	if stmt.Initializer != nil {
		err := i.execute(stmt.Initializer)
		if err != nil {
			i.environment = prev
			return err
		}
	}

	cond, err := i.evaluate(stmt.Condition)
	for err == nil && isTruthy(cond) {
		err = i.execute(stmt.Body)
		if err == BreakError {
			err = nil
			break
		} else if err == ContinueError {
			err = nil
		} else if err != nil {
			break
		}

		if stmt.Increment != nil {
			_, err = i.evaluate(stmt.Increment)
			if err != nil {
				i.environment = prev
				break
			}
		}
		cond, err = i.evaluate(stmt.Condition)

	}

	i.environment = prev
	return err
}

func (i *Interpreter) VisitBreakStmt(stmt *parser.BreakStmt) error {
	return BreakError
}

func (i *Interpreter) VisitContinueStmt(stmt *parser.ContinueStmt) error {
	return ContinueError
}

func (i *Interpreter) VisitFunctionStmt(stmt *parser.FunctionStmt) error {
	fun := NewFunction(stmt, i.environment, false)
	i.environment.Define(stmt.Name, fun)
	return nil
}

func (i *Interpreter) VisitReturnStmt(stmt *parser.ReturnStmt) error {
	var value interface{}
	var err error
	if stmt.Value != nil {
		value, err = i.evaluate(stmt.Value)
		if err != nil {
			return err
		}
	}

	return NewReturnError(stmt.Keyword, value)
}

func (i *Interpreter) executeBlock(statements []parser.Stmt, env *Environment) error {
	prev := i.environment
	i.environment = env
	var err error
	for _, stmt := range statements {
		err = i.execute(stmt)
		if err != nil {
			break
		}
	}

	i.environment = prev
	return err
}

func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}

	if b, ok := value.(bool); ok {
		return b
	}

	return true
}

func checkNumberOperand(operator *lexer.Token, operand interface{}) (float64, error) {
	op, ok := operand.(float64)
	if !ok {
		return 0, newError(operator, "Operand must be a number.")
	}
	return op, nil
}

func checkNumberOperands(operator *lexer.Token, left, right interface{}) (float64, float64, error) {
	l, ok := left.(float64)
	if !ok {
		return 0, 0, newError(operator, "Operands must be numbers.")
	}
	r, ok := right.(float64)
	if !ok {
		return 0, 0, newError(operator, "Operands must be numbers.")
	}

	return l, r, nil
}

func checkSliceOperand(operator *lexer.Token, operand interface{}) ([]interface{}, error) {
	sl, ok := operand.([]interface{})
	if !ok {
		return nil, newError(operator, "Operand must be an array.")
	}
	return sl, nil
}

func isEqual(left, right interface{}) (bool, error) {
	if left == nil && right == nil {
		return true, nil
	}
	if left == nil {
		return false, nil
	}

	switch l := left.(type) {
	case float64:
		if r, ok := right.(float64); ok {
			return l == r, nil
		}
	case bool:
		if r, ok := right.(bool); ok {
			return l == r, nil
		}
	case string:
		if r, ok := right.(string); ok {
			return l == r, nil
		}
	}

	return false, nil
}

func newError(token *lexer.Token, message string) *RuntimeError {
	return &RuntimeError{Token: token, Message: message}
}

func stringify(inter interface{}) string {
	if inter == nil {
		return "null"
	}

	return fmt.Sprintf("%v", inter)
}
