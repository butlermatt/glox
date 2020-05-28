package vm

import (
	"fmt"
	bc "github.com/butlermatt/glox/bytecode"
	"github.com/butlermatt/glox/scanner"
	"os"
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
		case bc.OpNil:
			vm.push(bc.Nil)
		case bc.OpTrue:
			vm.push(bc.True)
		case bc.OpFalse:
			vm.push(bc.False)
		case bc.OpEqual:
			r := vm.pop()
			l := vm.pop()
			vm.push(bc.BoolAsValue(valuesEqual(l, r)))
		case bc.OpGreater, bc.OpLess:
			fallthrough
		case bc.OpAdd, bc.OpSubtract, bc.OpMultiply, bc.OpDivide:
			err := vm.binaryOp(inst)
			if err != InterpretOk {
				return err
			}
		case bc.OpNot:
			vm.push(bc.BoolAsValue(isFalsey(vm.pop())))
		case bc.OpNegate:
			ty := vm.peek(0).Type()
			if ty != bc.ValNumber {
				vm.runtimeError("Operand must be a number.")
				return InterpretRuntimeError
			}
			obj := vm.pop().(bc.NumberValue)
			vm.push(bc.NumberValue{Value: -(obj.Value)})
		case bc.OpReturn:
			debug.PrintValue(vm.pop())
			fmt.Println()
			return InterpretOk
		}
	}
}

func (vm *VM) peek(distance int) bc.Value {
	return vm.Stack[vm.sTop-1-distance]
}

func (vm *VM) binaryOp(op bc.OpCode) InterpretResult {
	rType := vm.peek(0).Type()
	lType := vm.peek(1).Type()

	if lType != rType && lType != bc.ValNumber {
		vm.runtimeError("Operands must be numbers.")
		return InterpretRuntimeError
	}

	right := vm.pop().(bc.NumberValue)
	left := vm.pop().(bc.NumberValue)

	switch op {
	case bc.OpGreater:
		vm.push(bc.BoolAsValue(left.Value > right.Value))
	case bc.OpLess:
		vm.push(bc.BoolAsValue(left.Value < right.Value))
	case bc.OpAdd:
		vm.push(bc.NumberValue{Value: left.Value + right.Value})
	case bc.OpSubtract:
		vm.push(bc.NumberValue{Value: left.Value - right.Value})
	case bc.OpMultiply:
		vm.push(bc.NumberValue{Value: left.Value * right.Value})
	case bc.OpDivide:
		vm.push(bc.NumberValue{Value: left.Value / right.Value})
	}

	return InterpretOk
}

func (vm *VM) runtimeError(msg string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, msg, args...)
	line := vm.Chunk.Lines[vm.ip-1]
	_, _ = fmt.Fprintf(os.Stderr, "[line %d] in script\n", line)
}

func isFalsey(value bc.Value) bool {
	if value.Type() == bc.ValNil {
		return true
	}

	if value.Type() == bc.ValBool {
		b := value.(bc.BoolValue)
		return !b.Value
	}

	return false
}

func valuesEqual(l, r bc.Value) bool {
	if l.Type() != r.Type() {
		return false
	}

	switch lv := l.(type) {
	case bc.BoolValue:
		rv := r.(bc.BoolValue)
		return lv.Value == rv.Value
	case bc.NilValue:
		return true
	case bc.NumberValue:
		rv := r.(bc.NumberValue)
		return lv.Value == rv.Value
	default:
		return false // Should not reach here
	}
}
