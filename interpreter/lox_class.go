package interpreter

import "github.com/butlermatt/glpc/lexer"

func NewClass(name string, superclass *LoxClass, methods map[string]*Function) *LoxClass {
	return &LoxClass{Name: name, superclass: superclass, methods: methods}
}

type LoxClass struct {
	Name       string
	superclass *LoxClass
	methods    map[string]*Function
}

func (lc *LoxClass) String() string {
	return lc.Name
}

func (lc *LoxClass) Call(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	instance := &LoxInstance{klass: lc, fields: make(map[string]interface{})}
	initializer := lc.methods["init"]
	if initializer != nil {
		_, err := initializer.Bind(instance).Call(interpreter, args)
		if err != nil {
			return nil, err
		}
	}
	return instance, nil
}

func (lc *LoxClass) Arity() int {
	init := lc.methods["init"]
	if init == nil {
		return 0
	}
	return init.Arity()
}

func (lc *LoxClass) findMethod(instance *LoxInstance, name string) *Function {
	method := lc.methods[name]
	if method != nil {
		return method.Bind(instance)
	}

	if lc.superclass != nil {
		return lc.superclass.findMethod(instance, name)
	}

	return nil
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
