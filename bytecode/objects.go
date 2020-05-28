package bytecode

type ObjType byte
const(
	ObjString ObjType = iota
)

type Obj interface {
	Type() ObjType
}

// Object Value
type ObjValue struct {
	Value Obj
	Next  *ObjValue
}
func (o ObjValue) Type() ValueType { return ValObj }
func NewObjValue(obj Obj) ObjValue {
	o := ObjValue{Value: obj, Next: Objects}
	Objects = &o
	return o
}

var Objects *ObjValue = nil
func FreeObjects() {
	obj := Objects
	for obj != nil {
		n := obj.Next
		freeObject(obj.Value)
		obj.Value = nil
		obj.Next = nil
		obj = n
	}
}

func freeObject(obj Obj) {
	switch o := obj.(type) {
	case *StringObj:
		o.Value = ""
	}
}

type StringObj struct {
	Value string
}
func StringAsValue(value string) ObjValue {
	return NewObjValue(&StringObj{Value: value})
}

func (so *StringObj) Type() ObjType { return ObjString }

func IsString(value Value) bool {
	if value.Type() != ValObj {
		return false
	}

	vObj := value.(ObjValue).Value
	return vObj.Type() == ObjString
}