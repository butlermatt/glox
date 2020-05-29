package bytecode

type ValueType byte

const (
	// ValueType bool
	ValBool ValueType = iota
	// ValueType nil
	ValNil
	// ValueType Number
	ValNumber
	// ValueType Object
	ValObj
)

type Value interface {
	Type() ValueType
}

// Boolean Value
type BoolValue struct {
	Value bool
}

func (bv BoolValue) Type() ValueType { return ValBool }

var True = BoolValue{Value: true}
var False = BoolValue{Value: false}

func BoolAsValue(b bool) BoolValue {
	if b {
		return True
	}
	return False
}

// Nil Value
type NilValue struct{}
func (nv NilValue) Type() ValueType { return ValNil }

// Do not confuse with nil
var Nil = NilValue{}

// Number Value
type NumberValue struct {
	Value float64
}

func (num NumberValue) Type() ValueType { return ValNumber }


type ValueArray struct {
	Values []Value
}

// NewValueArray returns a new ValueArray for managing constant values.
func NewValueArray() *ValueArray {
	return &ValueArray{}
}

// Write appends a value to the ValueArray.
func (va *ValueArray) Write(v Value) {
	va.Values = append(va.Values, v)
}

// Free clears the ValueArray of any values.
func (va *ValueArray) Free() {
	va.Values = va.Values[:0]
}
