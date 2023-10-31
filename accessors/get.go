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

package accessors

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/gontainer/gontainer-helpers/v2/grouperror"
)

func Get(strct any, field string) (_ any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("get (%T).%+q: ", strct, field), err)
		}
	}()

	if field == "_" {
		return nil, errors.New(`"_" is not supported`)
	}

	chain, err := valueToKindChain(strct)
	if err != nil {
		return nil, err
	}

	reflectVal := reflect.ValueOf(strct)
	for len(chain) > 1 {
		chain = chain[1:]
		reflectVal = reflectVal.Elem()
	}

	if !reflectVal.IsValid() || reflectVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, %T given", strct)
	}

	f := reflectVal.FieldByName(field)
	if !f.IsValid() {
		return nil, fmt.Errorf("field %+q does not exist", field)
	}

	if !f.CanSet() { // handle unexported fields
		if !f.CanAddr() {
			tmpReflectVal := reflect.New(reflectVal.Type()).Elem()
			tmpReflectVal.Set(reflectVal)
			f = tmpReflectVal.FieldByName(field)
		}
		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	}

	return f.Interface(), nil
}
