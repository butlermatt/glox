package interpreter

import (
	"fmt"
	"github.com/butlermatt/glpc/lexer"
	"github.com/butlermatt/glpc/parser"
)

type RuntimeError struct {
	Token   *lexer.Token
	Message string
}

func (re *RuntimeError) Error() string {
	return fmt.Sprintf("[Runtime Error line %d] %s", re.Token.Line, re.Message)
}

type Interpreter struct {
}

func (i *Interpreter) Interpret(expr parser.Expr) (string, error) {
	res, err := i.evaluate(expr)
	if err != nil {
		return "", err
	}

	if res == nil {
		return "null", nil
	}

	return fmt.Sprintf("%v", res), nil
}

func (i *Interpreter) VisitBinary(binary *parser.Binary) (interface{}, error) {
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

func (i *Interpreter) VisitGrouping(grouping *parser.Grouping) (interface{}, error) {
	return i.evaluate(grouping)
}

func (i *Interpreter) VisitLiteral(literal *parser.Literal) (interface{}, error) {
	return literal.Value, nil
}
func (i *Interpreter) VisitUnary(unary *parser.Unary) (interface{}, error) {
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

func (i *Interpreter) evaluate(expr parser.Expr) (interface{}, error) {
	return expr.Accept(i)
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
