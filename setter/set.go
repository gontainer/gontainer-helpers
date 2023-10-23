package setter

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/gontainer/gontainer-helpers/v2/grouperror"
	intReflect "github.com/gontainer/gontainer-helpers/v2/internal/reflect"
)

/*
Set assigns the value `val` to the field `field` on the struct `strct`.
Unexported fields are supported.

	type Person struct {
		Name string
	}
	p := Person{}
	_ = setter.Set(&p, "Name", "Jane", false)
	fmt.Println(p) // {Jane}
*/
func Set(strct any, field string, val any, convert bool) (err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("set (%T).%+q: ", strct, field), err)
		}
	}()

	if field == "_" {
		return fmt.Errorf(`"_" is not supported`)
	}

	chain, err := valueToKindChain(strct)
	if err != nil {
		return err
	}

	/*
		removes prepending duplicate Ptr & Interface elements
		e.g.:
			s := &struct{ val int }{}
			Set(&s, ... // chain == {Ptr, Ptr, Struct}

		or:
			var s any = &struct{ val int }{}
			var s2 any = &s
			var s3 any = &s
			Set(&s3, ... // chain == {Ptr, Interface, Ptr, Interface, Ptr, Interface, Struct}
	*/
	reflectVal := reflect.ValueOf(strct)
	for {
		switch {
		case len(chain) >= 2 && chain[0] == reflect.Ptr && chain[1] == reflect.Ptr:
			reflectVal = reflectVal.Elem()
			chain = chain[1:]
			continue
		case len(chain) >= 3 && chain[0] == reflect.Ptr && chain[1] == reflect.Interface && chain[2] == reflect.Ptr:
			reflectVal = reflectVal.Elem().Elem()
			chain = chain[2:]
			continue
		}

		break
	}

	switch {
	// s := struct{ val int }{}
	// Set(&s...
	case chain.equalTo(reflect.Ptr, reflect.Struct):
		return setOnValue(
			reflectVal.Elem(),
			field,
			val,
			convert,
		)

	// var s any = struct{ val int }{}
	// Set(&s...
	case chain.equalTo(reflect.Ptr, reflect.Interface, reflect.Struct):
		v := reflectVal.Elem()
		tmp := reflect.New(v.Elem().Type()).Elem()
		tmp.Set(v.Elem())
		if err := setOnValue(tmp, field, val, convert); err != nil {
			return err
		}
		v.Set(tmp)
		return nil

	default:
		return fmt.Errorf("expected pointer to struct, %T given", strct)
	}
}

func setOnValue(strct reflect.Value, field string, val any, convert bool) error {
	f := strct.FieldByName(field)
	if !f.IsValid() {
		return fmt.Errorf("field %+q does not exist", field)
	}

	v, err := intReflect.ValueOf(val, f.Type(), convert)
	if err != nil {
		return err
	}

	if !f.CanSet() { // handle unexported fields
		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	}
	f.Set(v)
	return nil
}

type kindChain []reflect.Kind

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

func valueToKindChain(val any) (kindChain, error) {
	var r kindChain
	v := reflect.ValueOf(val)
	ptrs := make(map[string]struct{})
	for {
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			ptr := fmt.Sprintf("%p", v.Interface())
			if _, ok := ptrs[ptr]; ok {
				return nil, errors.New("unexpected pointer loop")
			}
			ptrs[ptr] = struct{}{}
		}
		r = append(r, v.Kind())
		if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
			continue
		}
		break
	}
	return r, nil
}
