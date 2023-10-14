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
func ConvertAndCopy(from interface{}, to interface{}) error {
	return copyTo(from, to, true)
}

// Copy copies a value of `from` to `to`.
//
//	from := 5
//	b := 0
//	Copy(from, &to)
//	fmt.Println(to) // 5
func Copy(from interface{}, to interface{}) error {
	return copyTo(from, to, false)
}

func copyTo(from interface{}, to interface{}, convert bool) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		err = fmt.Errorf("%s", r)
	}()

	t := reflect.ValueOf(to)

	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer, `%s` given", t.Kind())
	}

	var f reflect.Value
	if convert {
		f, err = internalReflect.Convert(from, t.Elem().Type())
		if err != nil {
			return err
		}
	} else {
		if from == nil && internalReflect.IsNilable(t.Elem().Kind()) {
			f = reflect.Zero(t.Elem().Type())
		} else {
			f = reflect.ValueOf(from)
		}
	}

	t.Elem().Set(f)
	return nil
}
