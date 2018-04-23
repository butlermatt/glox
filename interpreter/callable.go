package interpreter

import "github.com/butlermatt/glpc/parser"

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

type Function struct {
	declaration   *parser.FunctionStmt
	closure       *Environment
	isInitializer bool
}

func (f *Function) Arity() int     { return len(f.declaration.Parameters) }
func (f *Function) String() string { return "<fn " + f.declaration.Name.Lexeme + ">" }
func (f *Function) Call(interp *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != f.Arity() {
		return nil, newError(f.declaration.Name, "Incorrect number of arguments passed.")
	}

	env := NewEnclosedEnvironment(f.closure)
	for i, p := range f.declaration.Parameters {
		env.Define(p, args[i])
	}

	err := interp.executeBlock(f.declaration.Body, env)
	if err != nil {
		if e, ok := err.(*ReturnError); ok {
			return e.Value, nil
		}
		return nil, err
	}

	if f.isInitializer {
		return f.closure.m["this"], nil
	}

	return nil, nil
}

func (f *Function) Bind(instance *LoxInstance) *Function {
	env := NewEnclosedEnvironment(f.closure)
	env.m["this"] = instance
	return NewFunction(f.declaration, env, f.isInitializer)
}

func NewFunction(declaration *parser.FunctionStmt, environment *Environment, isInit bool) *Function {
	return &Function{declaration: declaration, closure: environment, isInitializer: isInit}
}
