package debug

import (
	"fmt"
	bc "github.com/butlermatt/glox/bytecode"
)

func DisassembleChunk(c *bc.Chunk, name string) {
	fmt.Printf("== %s ==\n", name)

	for offset := 0; offset < len(c.Code); {
		offset = DisassembleInstruction(c, offset)
	}
}

func DisassembleInstruction(c *bc.Chunk, offset int) int {
	fmt.Printf("%04d ", offset)

	if offset > 0 && c.Lines[offset] == c.Lines[offset-1] {
		fmt.Printf("   | ")
	} else {
		fmt.Printf("%4d ", c.Lines[offset])
	}

	instruction := bc.OpCode(c.Code[offset])
	switch instruction {
	case bc.OpConstant:
		return constantInstruction("OP_CONSTANT", c, offset)
	case bc.OpNil:
		return simpleInstruction("OP_NIL", offset)
	case bc.OpTrue:
		return simpleInstruction("OP_TRUE", offset)
	case bc.OpFalse:
		return simpleInstruction("OP_FALSE", offset)
	case bc.OpPop:
		return simpleInstruction("OP_POP", offset)
	case bc.OpGetGlobal:
		return constantInstruction("OP_GET_GLOBAL", c, offset)
	case bc.OpDefineGlobal:
		return constantInstruction("OP_DEFINE_GLOBAL", c, offset)
	case bc.OpEqual:
		return simpleInstruction("OP_EQUAL", offset)
	case bc.OpGreater:
		return simpleInstruction("OP_GREATER", offset)
	case bc.OpLess:
		return simpleInstruction("OP_LESS", offset)
	case bc.OpAdd:
		return simpleInstruction("OP_ADD", offset)
	case bc.OpSubtract:
		return simpleInstruction("OP_SUBTRACT", offset)
	case bc.OpMultiply:
		return simpleInstruction("OP_MULTIPLY", offset)
	case bc.OpDivide:
		return simpleInstruction("OP_DIVIDE", offset)
	case bc.OpNot:
		return simpleInstruction("OP_NOT", offset)
	case bc.OpNegate:
		return simpleInstruction("OP_NEGATE", offset)
	case bc.OpPrint:
		return simpleInstruction("OP_PRINT", offset)
	case bc.OpReturn:
		return simpleInstruction("OP_RETURN", offset)
	default:
		fmt.Printf("Unknown OpCode %d\n", instruction)
		return offset + 1
	}
}

func simpleInstruction(name string, offset int) int {
	fmt.Println(name)
	return offset + 1
}

func constantInstruction(name string, c *bc.Chunk, offset int) int {
	cOff := c.Code[offset+1]
	fmt.Printf("%-16s %4d '", name, cOff)
	bc.PrintValue(c.Constants.Values[cOff])
	fmt.Println("'")
	return offset + 2
}

