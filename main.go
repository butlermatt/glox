package main

import (
	"fmt"
	"github.com/butlermatt/glpc/lexer"
	"github.com/butlermatt/glpc/parser"
)

func main() {
	fmt.Println("This is a simple interface for debugging GLPC.")
	exp := &parser.Binary{
		Left: &parser.Unary{
			Operator: &lexer.Token{lexer.Minus, "-", nil, 1},
			Right:    &parser.Literal{Value: 123},
		},
		Operator: &lexer.Token{lexer.Star, "*", nil, 1},
		Right:    &parser.Grouping{&parser.Literal{Value: 45.67}},
	}

	ap := &parser.AstPrinter{}
	fmt.Println(ap.Print(exp))
}
