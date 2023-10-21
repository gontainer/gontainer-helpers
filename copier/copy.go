package copier

import (
	"fmt"
	"reflect"

	intReflect "github.com/gontainer/gontainer-helpers/v2/internal/reflect"
)

/*
Copy copies a value of `from` to `to`.

	from := 5
	b := 0
	Copy(from, &to, false)
	fmt.Println(to) // 5
*/
func Copy(from any, to any, convert bool) error {
	t := reflect.ValueOf(to)

	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("expected %s, %T given", reflect.Ptr.String(), to)
	}

	f, err := intReflect.ValueOf(from, t.Elem().Type(), convert)
	if err != nil {
		return err
	}

	t.Elem().Set(f)
	return nil
}
