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
		"Binary : Left Expr, Operator *lexer.Token, Right Expr",
		"Grouping : Expression Expr",
		"Literal : Value interface{}",
		"Unary : Operator *lexer.Token, Right Expr",
	}

	statements := []string{}

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
	Accept(ExprVisitor) interface{}
}

type Stmt interface {
	Accept(StmtVisitor) interface{}
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

	return nil
}

func defineType(file *os.File, ndType, structName, fieldStr string) error {
	checkWrite(file, "type %s struct {", structName)

	fields := strings.Split(fieldStr, ", ")
	for _, field := range fields {
		checkWrite(file, "\t%s", strings.TrimSpace(field))
	}

	checkWrite(file, "}\n")

	ptr := strings.ToLower(structName[0:1])
	return checkWrite(file, "func (%s *%s) Accept(visitor %sVisitor) interface{} { return visitor.Visit%[2]s(%[1]s) }", ptr, structName, ndType)
}

func defineVisitor(file *os.File, ndType string, types []string) error {
	checkWrite(file, "\ntype %sVisitor interface {", ndType)

	for _, ty := range types {
		lower := strings.ToLower(ty)
		checkWrite(file, "\tVisit%s(%s *%[1]s) interface{}", ty, lower)
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
