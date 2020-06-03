package bytecode

type ObjType byte

const (
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
	Hash  uint32
}

func NewStringObj(tbl *Table, value string) *StringObj {
	hash := HashString(value)
	interned := tbl.FindString(value, hash)
	if interned != nil {
		return interned
	}

	so := &StringObj{Value: value, Hash: hash}
	tbl.Set(so, Nil)
	return so
}

func StringAsValue(tbl *Table, value string) ObjValue {
	return NewObjValue(NewStringObj(tbl, value))
}

func (so *StringObj) Type() ObjType { return ObjString }

func IsString(value Value) bool {
	if value.Type() != ValObj {
		return false
	}

	vObj := value.(ObjValue).Value
	return vObj.Type() == ObjString
}

func HashString(key string) uint32 {
	var hash uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= 16777619
	}

	return hash
}
