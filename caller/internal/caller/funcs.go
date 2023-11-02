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

package caller

import (
	"errors"
	"fmt"
	"reflect"

	intReflect "github.com/gontainer/gontainer-helpers/v2/internal/reflect"
)

var (
	errorInterface = reflect.TypeOf((*error)(nil)).Elem()
)

func Func(fn any) (reflect.Value, error) {
	v := reflect.ValueOf(fn)
	if !v.IsValid() {
		return reflect.Value{}, fmt.Errorf("invalid func: %T", fn)
	}
	if v.Kind() != reflect.Func {
		return reflect.Value{}, fmt.Errorf("expected %s, %T given", reflect.Func.String(), fn)
	}
	return v, nil
}

func Provider(fn any) (reflect.Value, error) {
	v, err := Func(fn)
	if err != nil {
		return reflect.Value{}, err
	}
	if v.Type().NumOut() == 0 || v.Type().NumOut() > 2 {
		return reflect.Value{}, fmt.Errorf("provider must return 1 or 2 values, given function returns %d values", v.Type().NumOut())
	}
	if v.Type().NumOut() == 2 && !v.Type().Out(1).Implements(errorInterface) {
		return reflect.Value{}, errors.New("second value returned by provider must implement error interface")
	}
	return v, nil
}

func Method(object any, method string) (reflect.Value, error) {
	obj := reflect.ValueOf(object)
	if !obj.IsValid() {
		return reflect.Value{}, fmt.Errorf("invalid method receiver: %T", object)
	}
	fn := obj.MethodByName(method)
	_, err := intReflect.ValueToKindChain(obj)
	if err != nil {
		return reflect.Value{}, err
	}
	for !fn.IsValid() && (obj.Kind() == reflect.Ptr || obj.Kind() == reflect.Interface) {
		obj = obj.Elem()
		fn = obj.MethodByName(method)
	}
	if !fn.IsValid() {
		return reflect.Value{}, fmt.Errorf("invalid func (%T).%+q", object, method)
	}
	return fn, nil
}

func Wither(object any, method string) (reflect.Value, error) {
	fn, err := Method(object, method)
	if err != nil {
		return reflect.Value{}, err
	}
	if fn.Type().NumOut() != 1 {
		return reflect.Value{}, fmt.Errorf("wither must return 1 value, given function returns %d values", fn.Type().NumOut())
	}
	return fn, nil
}
