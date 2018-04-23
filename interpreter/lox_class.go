package interpreter

type LoxClass struct {
	Name string
}

func (lc *LoxClass) String() string {
	return lc.Name
}

func (lc *LoxClass) Call(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return &LoxInstance{klass: lc}, nil
}

func (lc *LoxClass) Arity() int {
	return 0
}

type LoxInstance struct {
	klass *LoxClass
}

func (li *LoxInstance) String() string {
	return li.klass.Name + " instance"
}
