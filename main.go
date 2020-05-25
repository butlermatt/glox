package main

import bc "github.com/butlermatt/glox/bytecode"
import "github.com/butlermatt/glox/debug"

func main() {
	chunk := bc.NewChunk()
	cloc := chunk.AddConstant(bc.Value(1.2))
	chunk.WriteOp(123, bc.OpConstant, cloc)
	chunk.WriteOp(123, bc.OpReturn)

	debug.DisassembleChunk(chunk, "Test chunk")
	chunk.Free()
}
