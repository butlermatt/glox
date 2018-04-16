package main

import (
	"bufio"
	"fmt"
	"github.com/butlermatt/glpc/lexer"
	"github.com/butlermatt/glpc/parser"
	"io/ioutil"
	"os"
)

func main() {
	fmt.Println("This is a simple interface for debugging GLPC.")

	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [script]", os.Args[0])
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		runPrompt()
	}
}

func runFile(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %+v", err)
		os.Exit(1)
	}

	run(string(data))
}

func runPrompt() {
	b := bufio.NewScanner(os.Stdin)

	fmt.Printf("> ")
	for b.Scan() {
		fmt.Printf("> ")
		run(b.Text())
	}
}

func run(input string) {
	l := lexer.New(input)
	p := parser.New(l)

	exp := p.Parse()
	errs := p.Errors()
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Printf("[Line %d] Error %s: %s\n", e.Line, e.Where, e.Msg)
		}
		return
	}

	ap := &parser.AstPrinter{}
	fmt.Println(ap.Print(exp))
}
