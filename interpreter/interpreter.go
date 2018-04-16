package interpreter

import (
	"github.com/butlermatt/glpc/lexer"
	"github.com/butlermatt/glpc/parser"
)

type Interpreter struct {
}

func (i *Interpreter) VisitBinary(binary *parser.Binary) interface{} { return nil }
func (i *Interpreter) VisitGrouping(grouping *parser.Grouping) interface{} {
	return i.evaluate(grouping)
}
func (i *Interpreter) VisitLiteral(literal *parser.Literal) interface{} { return literal.Value }
func (i *Interpreter) VisitUnary(unary *parser.Unary) interface{} {
	right := i.evaluate(unary.Right)

	switch unary.Operator.Type {
	case lexer.Minus:
		val, ok := right.(float64)
		if !ok {
			return nil
		}
		return -val
	}

	return nil
}
func (i *Interpreter) evaluate(expr parser.Expr) interface{} {
	return expr.Accept(i)
}
