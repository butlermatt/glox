package main

import (
	bc "github.com/butlermatt/glox/bytecode"
	//"github.com/butlermatt/glox/debug"
	"github.com/butlermatt/glox/vm"
)

func main() {
	chunk := bc.NewChunk()
	cloc := chunk.AddConstant(bc.Value(1.2))
	chunk.WriteOp(123, bc.OpConstant, cloc)

	cloc = chunk.AddConstant(3.4)
	chunk.WriteOp(123, bc.OpConstant, cloc)

	chunk.WriteOp(123, bc.OpAdd)

	cloc = chunk.AddConstant(5.6)
	chunk.WriteOp(123, bc.OpConstant, cloc)

	chunk.WriteOp(123, bc.OpDivide)
	chunk.WriteOp(123, bc.OpNegate)
	chunk.WriteOp(123, bc.OpReturn)

	//debug.DisassembleChunk(chunk, "Test chunk")
	v := vm.New()
	v.Interpret(chunk)
	chunk.Free()
	v.Free()
}
