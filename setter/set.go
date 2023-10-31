// Copyright (c) 2023 Bartłomiej Krukowski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is furnished
// to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
If the third argument equals true, it converts the type whenever it is possible.
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
		case chain.prefixed(reflect.Ptr, reflect.Ptr):
			reflectVal = reflectVal.Elem()
			chain = chain[1:]
			continue
		case chain.prefixed(reflect.Ptr, reflect.Interface, reflect.Ptr):
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

func (c kindChain) prefixed(kinds ...reflect.Kind) bool {
	if len(c) < len(kinds) {
		return false
	}

	return c[:len(kinds)].equalTo(kinds...)
}

func valueToKindChain(val any) (kindChain, error) {
	var r kindChain
	v := reflect.ValueOf(val)
	ptrs := make(map[uintptr]struct{})
	for {
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			ptr := v.Elem().UnsafeAddr()
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
