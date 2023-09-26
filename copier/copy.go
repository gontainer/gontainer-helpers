package copier

import (
	"fmt"
	"reflect"

	internalReflect "github.com/gontainer/gontainer-helpers/internal/reflect"
)

// Copy copies a value of `from` to `to`.
//
//	from := 5
//	b := 0
//	Copy(from, &to)
//	fmt.Println(to) // 5
func Copy(from interface{}, to interface{}) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		err = fmt.Errorf("%s", r)
	}()

	f := reflect.ValueOf(from)
	t := reflect.ValueOf(to)

	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer, `%s` given", t.Kind())
	}

	if from == nil && internalReflect.IsNilable(t.Elem().Kind()) {
		f = reflect.Zero(t.Elem().Type())
	}

	t.Elem().Set(f)
	return nil
}
