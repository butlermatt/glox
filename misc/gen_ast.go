package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <output directory>", os.Args[0])
		os.Exit(1)
	}

	outDir := os.Args[1]

	expressions := []string{
		"Array : Values []Expr",
		"Assign : Name *lexer.Token, Value Expr",
		"Binary : Left Expr, Operator *lexer.Token, Right Expr",
		"Call : Callee Expr, Paren *lexer.Token, Args []Expr",
		"Get : Object Expr, Name *lexer.Token",
		"Grouping : Expression Expr",
		"Literal : Value interface{}",
		"Logical : Left Expr, Operator *lexer.Token, Right Expr",
		"Set : Object Expr, Name *lexer.Token, Value Expr",
		"Super : Keyword *lexer.Token, Method *lexer.Token",
		"This : Keyword *lexer.Token",
		"Unary : Operator *lexer.Token, Right Expr",
		"Variable : Name *lexer.Token",
	}

	statements := []string{
		"Block : Statements []Stmt",
		"Class : Name *lexer.Token, Superclass *VariableExpr, Methods []*FunctionStmt",
		"Expression : Expression Expr",
		"Function : Name *lexer.Token, Parameters []*lexer.Token, Body []Stmt",
		"If : Condition Expr, Then Stmt, Else Stmt",
		"Print : Expression Expr",
		"Return : Keyword *lexer.Token, Value Expr",
		"Var : Name *lexer.Token, Initializer Expr",
		"For : Initializer Stmt, Condition Expr, Body Stmt, Increment Expr",
		"Break : Keyword *lexer.Token",
		"Continue : Keyword *lexer.Token",
	}

	err := defineAst(outDir, expressions, statements)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v", err)
		os.Exit(1)
	}
}

func defineAst(outDir string, expr []string, stmt []string) error {
	file, err := os.Create(outDir + "/ast.go")
	if err != nil {
		return err
	}
	defer file.Close()

	checkWrite(file, "package %s\n", outDir)
	checkWrite(file, "import \"github.com/butlermatt/glpc/lexer\"\n")
	checkWrite(file,
		`type Expr interface {
	Accept(ExprVisitor) (interface{}, error)
}

type Stmt interface {
	Accept(StmtVisitor) error
}`)

	checkWrite(file, "\n")

	var exprNames []string
	for _, ty := range expr {
		structInfo := strings.Split(ty, ":")
		structName := strings.TrimSpace(structInfo[0])
		exprNames = append(exprNames, structName)
		err := defineType(file, "Expr", structName, structInfo[1])
		if err != nil {
			return err
		}
	}

	err = defineVisitor(file, "Expr", exprNames)
	if err != nil {
		return err
	}

	var stmtNames []string
	for _, st := range stmt {
		structInfo := strings.Split(st, ":")
		structName := strings.TrimSpace(structInfo[0])
		stmtNames = append(stmtNames, structName)
		err := defineType(file, "Stmt", structName, structInfo[1])
		if err != nil {
			return err
		}
	}
	err = defineVisitor(file, "Stmt", stmtNames)
	if err != nil {
		return err
	}

	return nil
}

func defineType(file *os.File, ndType, structName, fieldStr string) error {
	checkWrite(file, "type %s struct {", structName+ndType)

	fields := strings.Split(fieldStr, ", ")
	for _, field := range fields {
		checkWrite(file, "\t%s", strings.TrimSpace(field))
	}

	checkWrite(file, "}\n")

	ptr := strings.ToLower(structName[0:1])
	if ndType == "Expr" {
		return checkWrite(file, "func (%s *%s) Accept(visitor %sVisitor) (interface{}, error) { return visitor.Visit%[2]s(%[1]s) }\n", ptr, structName+ndType, ndType)
	} else {
		return checkWrite(file, "func (%s *%s) Accept(visitor %sVisitor) error { return visitor.Visit%[2]s(%[1]s) }\n", ptr, structName+ndType, ndType)
	}
}

func defineVisitor(file *os.File, ndType string, types []string) error {
	checkWrite(file, "\ntype %sVisitor interface {", ndType)

	for _, ty := range types {
		lower := strings.ToLower(ndType)
		if ndType == "Expr" {
			checkWrite(file, "\tVisit%s(%s *%[1]s) (interface{}, error)", ty+ndType, lower)
		} else {
			checkWrite(file, "\tVisit%s(%s *%[1]s) error", ty+ndType, lower)
		}
	}
	return checkWrite(file, "}")
}

var writeError error = nil

func checkWrite(f *os.File, str string, args ...interface{}) error {
	if writeError != nil {
		return writeError
	}

	_, writeError = fmt.Fprintf(f, str+"\n", args...)
	return writeError
}
