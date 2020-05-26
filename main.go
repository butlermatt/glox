package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

import (
	"github.com/butlermatt/glox/vm"
)

func main() {
	if len(os.Args) == 1 {
		repl()
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		fmt.Fprintf(os.Stderr, "Usage: %s [path]\n", os.Args[0])
		os.Exit(64)
	}
}

func repl() {
	v := vm.New()
	b := bufio.NewScanner(os.Stdin)

	fmt.Printf("> ")
	for b.Scan() {
		fmt.Printf("> ")
		err := v.Interpret(b.Text())
		if err != vm.InterpretOk {
			fmt.Println(err)
		}
	}

	v.Free()
}

func runFile(path string) {
	v := vm.New()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %+v", err)
		os.Exit(1)
	}

	res := v.Interpret(string(data))
	if res != vm.InterpretOk {
		fmt.Fprintf(os.Stderr, "%v\n", res)
		os.Exit(70)
	}

	v.Free()
}
