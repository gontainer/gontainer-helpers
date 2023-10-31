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
	"fmt"

	"github.com/gontainer/gontainer-helpers/v2/caller/internal/caller"
	"github.com/gontainer/gontainer-helpers/v2/grouperror"
)

// Call calls the given function with the given arguments.
// It returns values returned by the function in a slice.
// If the third argument equals true, it converts types whenever it is possible.
func Call(fn any, args []any, convertArgs bool) (_ []any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("cannot call %T: ", fn), err)
		}
	}()

	v, err := caller.Func(fn)
	if err != nil {
		return nil, err
	}
	return caller.CallFunc(v, args, convertArgs)
}

/*
CallProvider works similar to [Call] with the difference it requires a provider as the first argument.
Provider is a function which returns 1 or 2 values.
The second return value which is optional must be a type of error.

	p := func() (any, error) {
	    db, err := sql.Open("mysql", "user:password@/dbname")
	    if err != nil {
	         return nil, err
	    }

	    db.SetConnMaxLifetime(time.Minute * 3)
	    db.SetMaxOpenConns(10)
	    db.SetMaxIdleConns(10)

	    return db, nil
	}

	mysql, err := CallProvider(p)
*/
func CallProvider(provider any, args []any, convertArgs bool) (_ any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("cannot call provider %T: ", provider), err)
		}
	}()

	fn, err := caller.Provider(provider)
	if err != nil {
		return nil, err
	}

	results, err := caller.CallFunc(fn, args, convertArgs)
	if err != nil {
		return nil, err
	}

	r := results[0]
	var e error
	if len(results) > 1 {
		// do not panic when results[1] == nil
		e, _ = results[1].(error)
	}

	return r, e
}

/*
CallByName works similar to [Call] with the difference it calls the method by the name over the given receiver.

	type Person struct {
		Name string
	}

	func (p *Person) SetName(n string) {
		p.Name = n
	}

	func main() {
		p := &Person{}
		_, _ = caller.CallByName(p, "SetName", []any{"Mary"}, false)
		fmt.Println(p.name)
		// Output: Mary
	}
*/
func CallByName(object any, method string, args []any, convertArgs bool) (_ []any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("cannot call method (%T).%+q: ", object, method), err)
		}
	}()

	fn, err := caller.Method(object, method)
	if err != nil {
		return nil, err
	}
	return caller.CallFunc(fn, args, convertArgs)
}

/*
CallWitherByName works similar to [CallByName] with the difference the method must be a wither.

	type Person struct {
	    name string
	}

	func (p Person) WithName(n string) Person {
	    p.Name = n
	    return p
	}

	func main() {
	    p := Person{}
	    p2, _ := caller.CallWitherByName(p, "WithName", "Mary")
	    fmt.Printf("%+v", p2) // {name:Mary}
	}
*/
func CallWitherByName(object any, wither string, args []any, convertArgs bool) (_ any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("cannot call wither (%T).%+q: ", object, wither), err)
		}
	}()

	fn, err := caller.Wither(object, wither)
	if err != nil {
		return nil, err
	}

	r, err := caller.CallFunc(fn, args, convertArgs)
	var v any
	if len(r) > 0 {
		v = r[0]
	}
	return v, err
}
