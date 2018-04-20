package interpreter

import (
	"errors"
	"fmt"
	"github.com/butlermatt/glpc/lexer"
	"github.com/butlermatt/glpc/parser"
	"time"
)

var BreakError = errors.New("Unexpected 'break' outside of loop")
var ContinueError = errors.New("Unexpected 'continue' outside of loop")

type RuntimeError struct {
	Token   *lexer.Token
	Message string
}

func (re *RuntimeError) Error() string {
	return fmt.Sprintf("[Runtime Error line %d] %s", re.Token.Line, re.Message)
}

type Interpreter struct {
	stmts       []parser.Stmt
	globals     *Environment
	environment *Environment
}

func New(statements []parser.Stmt) *Interpreter {
	env := NewEnvironment()
	env.builtin("clock", &BuiltIn{
		arity: 0,
		callFn: func(interp *Interpreter, args []interface{}) (interface{}, error) {
			return float64(time.Now().Unix()), nil
		}},
	)
	return &Interpreter{stmts: statements, globals: env, environment: env}
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
	return i.evaluate(grouping)
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
	return i.environment.Get(expr.Name)
}

func (i *Interpreter) VisitAssignExpr(expr *parser.AssignExpr) (interface{}, error) {
	value, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}

	err = i.environment.Assign(expr.Name, value)
	if err != nil {
		return nil, err
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

func (i *Interpreter) evaluate(expr parser.Expr) (interface{}, error) {
	return expr.Accept(i)
}

func (i *Interpreter) VisitExpressionStmt(stmt *parser.ExpressionStmt) error {
	i.evaluate(stmt.Expression)
	return nil
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
