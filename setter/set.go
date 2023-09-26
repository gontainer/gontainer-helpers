package setter

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	internalReflect "github.com/gontainer/gontainer-helpers/internal/reflect"
)

type kindChain []reflect.Kind

func set(strct reflect.Value, field string, val interface{}) error {
	f := strct.FieldByName(field)
	if !f.IsValid() {
		return fmt.Errorf("field `%s` does not exist", field)
	}

	v, err := internalReflect.Convert(val, f.Type())
	if err != nil {
		return err
	}

	if !f.CanSet() { // handle unexported fields
		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	}
	f.Set(v)
	return nil
}

// Set assigns the value `val` to the field `field` on the struct `strct`.
// Unexported fields are supported.
func Set(strct interface{}, field string, val interface{}) error {
	if field == "_" {
		return fmt.Errorf(`"_" is not supported`)
	}

	chain := valueToKindChain(strct)

	// removes prepending duplicate Ptr elements
	// e.g.
	// s := &struct{ val int }{}
	// Set(&s... // chain == {Ptr, Ptr, Struct}
	reflectVal := reflect.ValueOf(strct)
	for len(chain) >= 2 && chain[0] == reflect.Ptr && chain[1] == reflect.Ptr {
		reflectVal = reflectVal.Elem()
		chain = chain[1:]
	}

	wrap := func(err error) error {
		if err == nil {
			return nil
		}
		return fmt.Errorf("set `%T`.%+q: %w", strct, field, err)
	}

	switch {
	// s := struct{ val int }{}
	// Set(&s...
	case chain.equalTo(reflect.Ptr, reflect.Struct):
		return wrap(set(
			reflectVal.Elem(),
			field,
			val,
		))

	// case chain.equalTo(reflect.Ptr, reflect.Interface, reflect.Ptr (, reflect.Ptr...), reflect.Struct):
	// var s interface{} = &struct{ val int }{}
	// Set(&s...
	case chain.isInterfaceOverPointerChain():
		elem := reflectVal.Elem()
		for i := 0; i < len(chain)-2; i++ {
			elem = elem.Elem()
		}
		return wrap(set(elem, field, val))

	// var s interface{} = struct{ val int }{}
	// Set(&s...
	case chain.equalTo(reflect.Ptr, reflect.Interface, reflect.Struct):
		v := reflectVal.Elem()
		tmp := reflect.New(v.Elem().Type()).Elem()
		tmp.Set(v.Elem())
		if err := wrap(set(tmp, field, val)); err != nil {
			return err
		}
		v.Set(tmp)
		return nil

	default:
		return fmt.Errorf("expected pointer to struct, %s given", chain.String())
	}
}

func (c kindChain) equalTo(kinds ...reflect.Kind) bool {
	if len(c) != len(kinds) {
		return false
	}

	for i := 0; i < len(c); i++ {
		if c[i] != kinds[i] {
			return false
		}
	}

	return true
}

// isInterfaceOverPointerChain is equivalent to:
// chain.equalTo(reflect.Ptr, reflect.Interface, reflect.Ptr (, reflect.Ptr...), reflect.Struct)
func (c kindChain) isInterfaceOverPointerChain() bool {
	if len(c) < 4 {
		return false
	}
	if c[0] != reflect.Ptr {
		return false
	}
	if c[1] != reflect.Interface {
		return false
	}
	if c[len(c)-1] != reflect.Struct {
		return false
	}

	for _, curr := range c[2 : len(c)-1] {
		if curr != reflect.Ptr {
			return false
		}
	}

	return true
}

func (c kindChain) String() string {
	parts := make([]string, len(c))
	for i, k := range c {
		parts[i] = k.String()
	}
	return strings.Join(parts, ".")
}

func valueToKindChain(val interface{}) kindChain {
	var r kindChain
	v := reflect.ValueOf(val)
	for {
		r = append(r, v.Kind())
		if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
			continue
		}
		break
	}
	return r
}
