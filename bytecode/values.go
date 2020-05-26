package bytecode

type Value float64

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
