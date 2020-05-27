package vm

import (
	"fmt"
	bc "github.com/butlermatt/glox/bytecode"
	"github.com/butlermatt/glox/scanner"
)
import "github.com/butlermatt/glox/debug"

type InterpretResult byte

const (
	InterpretOk InterpretResult = iota
	InterpretCompileError
	InterpretRuntimeError
)

const DEBUG_TRACE = true
const STACK_MAX = 256

type VM struct {
	Chunk    *bc.Chunk
	ip       int
	Stack    [STACK_MAX]bc.Value
	sTop     int
	compiler *scanner.Compiler
}

func New() *VM {
	return &VM{compiler: scanner.NewCompiler()}
}

func (vm *VM) Free() {
	vm.Chunk.Free()
	vm.Chunk = nil
	vm.ip = 0
	vm.sTop = 0
}

func (vm *VM) Interpret(source string) InterpretResult {
	c := bc.NewChunk()

	if !vm.compiler.Compile(source, c) {
		c.Free()
		return InterpretCompileError
	}

	vm.Chunk = c
	vm.ip = 0

	result := vm.run()

	vm.Chunk.Free()
	return result
}

func (vm *VM) push(value bc.Value) {
	// TODO: Add error checking that we're not adding past the top of the stack
	vm.Stack[vm.sTop] = value
	vm.sTop++
}

func (vm *VM) pop() bc.Value {
	vm.sTop--
	return vm.Stack[vm.sTop]
}

func (vm *VM) run() InterpretResult {
	for {
		inst := bc.OpCode(vm.Chunk.Code[vm.ip])

		if DEBUG_TRACE {
			fmt.Printf("          ")
			for i := 0; i < vm.sTop; i++ {
				fmt.Printf("[ ")
				debug.PrintValue(vm.Stack[i])
				fmt.Printf(" ]")
			}
			fmt.Println()
			debug.DisassembleInstruction(vm.Chunk, vm.ip)
		}

		vm.ip++

		switch inst {
		case bc.OpConstant:
			con := vm.Chunk.Constants.Values[vm.Chunk.Code[vm.ip]]
			vm.ip++
			vm.push(con)
		case bc.OpAdd, bc.OpSubtract, bc.OpMultiply, bc.OpDivide:
			vm.binaryOp(inst)
		case bc.OpNegate:
			vm.push(-vm.pop())
		case bc.OpReturn:
			debug.PrintValue(vm.pop())
			fmt.Println()
			return InterpretOk
		}
	}
}

func (vm *VM) binaryOp(op bc.OpCode) {
	right := vm.pop()
	left := vm.pop()

	switch op {
	case bc.OpAdd:
		vm.push(left + right)
	case bc.OpSubtract:
		vm.push(left - right)
	case bc.OpMultiply:
		vm.push(left * right)
	case bc.OpDivide:
		vm.push(left / right)
	}
}
