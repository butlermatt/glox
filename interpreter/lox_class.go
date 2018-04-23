package interpreter

import "github.com/butlermatt/glpc/lexer"

type LoxClass struct {
	Name string
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

	return nil, newError(name, "Undefined property '"+name.Lexeme+"'.")
}
