package interpreter

import (
	"github.com/butlermatt/glpc/lexer"
)

type Environment struct {
	m map[string]interface{}
}

func NewEnvironment() *Environment {
	return &Environment{m: make(map[string]interface{})}
}

func (e *Environment) Define(name *lexer.Token, value interface{}) error {
	if _, ok := e.m[name.Lexeme]; ok {
		return newError(name, "Variable '"+name.Lexeme+"' has already been declared.")
	}
	e.m[name.Lexeme] = value
	return nil
}

func (e *Environment) Get(name *lexer.Token) (interface{}, error) {
	v, ok := e.m[name.Lexeme]
	if ok {
		return v, nil
	}
	return nil, newError(name, "Undefined variable '"+name.Lexeme+"'.")
}

func (e *Environment) Assign(name *lexer.Token, value interface{}) error {
	if _, ok := e.m[name.Lexeme]; !ok {
		return newError(name, "Undefined variable '"+name.Lexeme+"'.")
	}

	e.m[name.Lexeme] = value
	return nil
}
