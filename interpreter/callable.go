package interpreter

type CallFn func(interpreter *Interpreter, args []interface{}) (interface{}, error)

type Callable interface {
	// Arity is the number of expected arguments
	Arity() int
	Call(interpreter *Interpreter, args []interface{}) (interface{}, error)
}

type BuiltIn struct {
	arity  int
	callFn CallFn
}

func (b *BuiltIn) Arity() int { return b.arity }
func (b *BuiltIn) Call(interp *Interpreter, args []interface{}) (interface{}, error) {
	return b.callFn(interp, args)
}
