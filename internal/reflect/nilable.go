package reflect

import (
	"reflect"
)

func IsNilable(k reflect.Kind) bool {
	switch k {
	case
		reflect.Chan,
		reflect.Func,
		reflect.Map,
		reflect.Ptr,
		reflect.UnsafePointer,
		reflect.Interface,
		reflect.Slice:
		return true
	}
	return false
}
