package interpreter

import "github.com/butlermatt/glpc/lexer"

func NewClass(name string, methods map[string]*Function) *LoxClass {
	return &LoxClass{Name: name, methods: methods}
}

type LoxClass struct {
	Name    string
	methods map[string]*Function
}

func (lc *LoxClass) String() string {
	return lc.Name
}

func (lc *LoxClass) Call(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return &LoxInstance{klass: lc, fields: make(map[string]interface{})}, nil
}

func (lc *LoxClass) Arity() int {
	return 0
}

func (lc *LoxClass) findMethod(instance *LoxInstance, name string) *Function {
	return lc.methods[name]
}

type LoxInstance struct {
	klass  *LoxClass
	fields map[string]interface{}
}

func (li *LoxInstance) String() string {
	return li.klass.Name + " instance"
}

func (li *LoxInstance) Get(name *lexer.Token) (interface{}, error) {
	if v, ok := li.fields[name.Lexeme]; ok {
		return v, nil
	}

	if m := li.klass.findMethod(li, name.Lexeme); m != nil {
		return m, nil
	}

	return nil, newError(name, "Undefined property '"+name.Lexeme+"'.")
}

func (li *LoxInstance) Set(name *lexer.Token, value interface{}) {
	li.fields[name.Lexeme] = value
}
