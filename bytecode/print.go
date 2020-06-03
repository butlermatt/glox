package bytecode

import "fmt"

func PrintValue(value Value) {
	switch value.Type() {
	case ValBool:
		if value == True {
			fmt.Printf("true")
		} else {
			fmt.Printf("false")
		}
	case ValNil:
		fmt.Printf("nil")
	case ValNumber:
		fmt.Printf("%g", value.(NumberValue).Value)
	case ValObj:
		printObject(value.(ObjValue).Value)
	}
}

func printObject(value Obj) {
	switch v := value.(type) {
	case *StringObj:
		fmt.Printf("%s", v.Value)
	}
}
