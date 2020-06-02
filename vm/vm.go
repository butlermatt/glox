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
	strings  *bc.Table
	globals  *bc.Table

	compiler *scanner.Compiler
}

func New() *VM {
	bc.Objects = nil
	return &VM{strings: bc.NewTable(), globals: bc.NewTable(), compiler: scanner.NewCompiler()}
}

func (vm *VM) Free() {
	vm.Chunk.Free()
	vm.Chunk = nil
	vm.ip = 0
	vm.sTop = 0

	bc.FreeObjects()
	vm.strings.Free()
	vm.globals.Free()
}

func (vm *VM) Interpret(source string) InterpretResult {
	c := bc.NewChunk()

	if !vm.compiler.Compile(vm.strings, source, c) {
		c.Free()
		return InterpretCompileError
	}

	vm.Chunk = c
	vm.ip = 0

	result := vm.run()

	vm.Free()
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
				bc.PrintValue(vm.Stack[i])
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
		case bc.OpPop:
			vm.pop()
		case bc.OpGetGlobal:
			name := vm.readString()
			value, ok := vm.globals.Get(name)
			if !ok {
				vm.runtimeError("Undefined variable %q. ", name.Value)
				return InterpretRuntimeError
			}
			vm.push(value)
		case bc.OpDefineGlobal:
			val := vm.readString()
			vm.globals.Set(val, vm.peek(0))
			vm.pop()
		case bc.OpEqual:
			r := vm.pop()
			l := vm.pop()
			vm.push(bc.BoolAsValue(valuesEqual(l, r)))
		case bc.OpGreater, bc.OpLess:
			err := vm.binaryOp(inst)
			if err != InterpretOk {
				return err
			}
		case bc.OpAdd:
			if bc.IsString(vm.peek(0)) && bc.IsString(vm.peek(1)) {
				vm.concatenate()
			} else if vm.peek(0).Type() == bc.ValNumber && vm.peek(1).Type() == bc.ValNumber {
				right := vm.pop().(bc.NumberValue)
				left := vm.pop().(bc.NumberValue)
				vm.push(bc.NumberValue{Value: left.Value + right.Value})
			} else {
				vm.runtimeError("Operands must be two numbers or two strings.")
				return InterpretRuntimeError
			}
		case bc.OpSubtract, bc.OpMultiply, bc.OpDivide:
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
		case bc.OpPrint:
			bc.PrintValue(vm.pop())
			fmt.Println()
		case bc.OpReturn:
			// Exit Interpreter
			return InterpretOk
		}
	}
}

func (vm *VM) readString() *bc.StringObj {
	con := vm.Chunk.Constants.Values[vm.Chunk.Code[vm.ip]]
	vm.ip++
	return con.(bc.ObjValue).Value.(*bc.StringObj)
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
	case bc.OpSubtract:
		vm.push(bc.NumberValue{Value: left.Value - right.Value})
	case bc.OpMultiply:
		vm.push(bc.NumberValue{Value: left.Value * right.Value})
	case bc.OpDivide:
		vm.push(bc.NumberValue{Value: left.Value / right.Value})
	}

	return InterpretOk
}

func (vm *VM) concatenate() {
	right := vm.pop().(bc.ObjValue).Value.(*bc.StringObj)
	left  := vm.pop().(bc.ObjValue).Value.(*bc.StringObj)

	vm.push(bc.StringAsValue(vm.strings, left.Value + right.Value))
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
	case bc.ObjValue:
		lobj := lv.Value.(*bc.StringObj)
		robj := r.(bc.ObjValue).Value.(*bc.StringObj)
		return lobj == robj
	default:
		return false // Should not reach here
	}
}
