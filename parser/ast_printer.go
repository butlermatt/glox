package parser

//
//import (
//	"bytes"
//	"fmt"
//	"log"
//)
//
//type AstPrinter struct {
//}
//
//func (ap *AstPrinter) Print(expr Expr) string {
//	res := expr.Accept(ap)
//	result, ok := res.(string)
//	if !ok {
//		log.Fatalf("failed to convert accepted return to string. got=%T", res)
//	}
//
//	return result
//}
//
//func (ap *AstPrinter) VisitBinary(bin *Binary) interface{} {
//	return ap.parenthesize(bin.Operator.Lexeme, bin.Left, bin.Right)
//}
//
//func (ap *AstPrinter) VisitGrouping(group *Grouping) interface{} {
//	return ap.parenthesize("group", group.Expression)
//}
//
//func (ap *AstPrinter) VisitLiteral(literal *Literal) interface{} {
//	if literal.Value == nil {
//		return "Null"
//	}
//	return fmt.Sprintf("%v", literal.Value)
//}
//
//func (ap *AstPrinter) VisitUnary(unary *Unary) interface{} {
//	return ap.parenthesize(unary.Operator.Lexeme, unary.Right)
//}
//
//func (ap *AstPrinter) parenthesize(name string, exprs ...Expr) string {
//	var out bytes.Buffer
//	out.WriteByte('(')
//	out.WriteString(name)
//	for _, e := range exprs {
//		out.WriteByte(' ')
//		res := e.Accept(ap)
//		rest, ok := res.(string)
//		if !ok {
//			log.Fatalf("error parsing result of accept. On lexeme %q, expected=string, got=%T", name, res)
//		}
//		out.WriteString(rest)
//	}
//
//	out.WriteByte(')')
//
//	return out.String()
//}
