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
	"reflect"
)

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
