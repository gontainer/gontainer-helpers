package copier

import (
	"fmt"
	"reflect"

	internalReflect "github.com/gontainer/gontainer-helpers/internal/reflect"
)

// ConvertAndCopy works similar to Copy, but it converts the type whenever it is possible.
//
//	var (
//		from int = 5
//		to   uint
//	)
//	err := copier.ConvertAndCopy(from, &to)
//	fmt.Println(to) // 5
func ConvertAndCopy(from any, to any) error {
	return copyTo(from, to, true)
}

// Copy copies a value of `from` to `to`.
//
//	from := 5
//	b := 0
//	Copy(from, &to)
//	fmt.Println(to) // 5
func Copy(from any, to any) error {
	return copyTo(from, to, false)
}

func copyTo(from any, to any, convert bool) (err error) {
	t := reflect.ValueOf(to)

	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer, `%s` given", t.Kind())
	}

	f, err := internalReflect.ValueOf(from, t.Elem().Type(), convert)
	if err != nil {
		return err
	}

	if !f.Type().AssignableTo(t.Elem().Type()) {
		return fmt.Errorf("value of type %s is not assignable to type %s", f.Type().String(), t.Elem().Type().String())
	}

	t.Elem().Set(f)
	return nil
}
