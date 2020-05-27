package bytecode

type OpCode byte

const (
	OpConstant OpCode = iota
	OpAdd
	OpSubtract
	OpMultiply
	OpDivide
	OpNegate
	OpReturn
)

type Chunk struct {
	Code      []byte
	Constants *ValueArray
	Lines     []int // TODO: Convert to map of lines and chunk count. (eg: 123: 5 -> 5 bytes are on line 123)
}

// NewChunk returns a new Chunk ready to be written to.
func NewChunk() *Chunk {
	return &Chunk{Constants: NewValueArray()}
}

// Write appends the specified byte b to the Chunk of Code.
func (c *Chunk) Write(line int, b byte) {
	c.Code = append(c.Code, b)
	c.Lines = append(c.Lines, line)
}

// WriteOp is a convenience wrapper around Write() to allow OpCodes to be passed directly rather than manually casting
func (c *Chunk) WriteOp(line int, op OpCode, args ...byte) {
	c.Write(line, byte(op))

	for _, a := range args {
		c.Write(line, a)
	}
}

// Free clears the existing memory from the Chunk.
func (c *Chunk) Free() {
	c.Code = c.Code[:0]
	c.Constants.Free()
}

// AddConstant appends a constant to the Chunk, and returns its current index.
func (c *Chunk) AddConstant(value Value) int {
	c.Constants.Write(value)
	return len(c.Constants.Values) - 1
}
