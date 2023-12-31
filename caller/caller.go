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
	"reflect"

	"github.com/gontainer/gontainer-helpers/v3/caller/internal/caller"
	"github.com/gontainer/gontainer-helpers/v3/grouperror"
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

const (
	providerInternalErrPrefix       = "cannot call provider %T: "
	providerMethodInternalErrPrefix = "cannot call provider (%T).%+q: "
	providerExternalErrPrefix       = "provider returned error: "
)

func callProvider(
	getFn func() (reflect.Value, error),
	args []any,
	convertArgs bool,
	internalErrPrefix func() string,
) (_ any, err error) {
	executedProvider := false
	defer func() {
		if !executedProvider && err != nil {
			err = grouperror.Prefix(internalErrPrefix(), err)
		}
	}()

	fn, err := getFn()
	if err != nil {
		return nil, err
	}

	if err := caller.ValidatorProvider.Validate(fn); err != nil {
		return nil, err
	}

	results, err := caller.CallFunc(fn, args, convertArgs)
	if err != nil {
		return nil, err
	}

	executedProvider = true

	r := results[0]
	var e error
	if len(results) > 1 {
		// do not panic when results[1] == nil
		e, _ = results[1].(error)
	}
	if e != nil {
		e = grouperror.Prefix(providerExternalErrPrefix, newProviderError(e))
	}

	return r, e
}

/*
CallProvider works similar to [Call] with the difference it requires a provider as the first argument.
Provider is a function which returns 1 or 2 values.
The second return value which is optional must be a type of error.
See [ProviderError].

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

	db, err := caller.CallProvider(p, nil, false)
*/
func CallProvider(provider any, args []any, convertArgs bool) (any, error) {
	return callProvider(
		func() (reflect.Value, error) {
			return caller.Func(provider)
		},
		args,
		convertArgs,
		func() string {
			return fmt.Sprintf(providerInternalErrPrefix, provider)
		},
	)
}

/*
CallProviderMethod works similar to [CallProvider], but the provider must be a method on the given object.

	db, _ := sql.Open("mysql", "user:password@/dbname")
	tx, err := caller.CallProviderMethod(db, "Begin", nil, false)
*/
func CallProviderMethod(object any, method string, args []any, convertArgs bool) (any, error) {
	return callProvider(
		func() (reflect.Value, error) {
			return caller.Method(object, method)
		},
		args,
		convertArgs,
		func() string {
			return fmt.Sprintf(providerMethodInternalErrPrefix, object, method)
		},
	)
}

// ForceCallProviderMethod is an extended version of [CallProviderMethod].
// See [ForceCallMethod].
func ForceCallProviderMethod(object any, method string, args []any, convertArgs bool) (any, error) {
	results, err := caller.ValidateAndForceCallMethod(object, method, args, convertArgs, caller.ValidatorProvider)
	if err != nil {
		return nil, grouperror.Prefix(fmt.Sprintf(providerMethodInternalErrPrefix, object, method), err)
	}

	r := results[0]
	var e error
	if len(results) > 1 {
		// do not panic when results[1] == nil
		e, _ = results[1].(error)
	}
	if e != nil {
		e = grouperror.Prefix(providerExternalErrPrefix, newProviderError(e))
	}

	return r, e
}

/*
CallMethod works similar to [Call] with the difference it calls the method by the name over the given receiver.

	type Person struct {
		Name string
	}

	func (p *Person) SetName(n string) {
		p.Name = n
	}

	func main() {
		p := &Person{}
		_, _ = caller.CallMethod(p, "SetName", []any{"Mary"}, false)
		fmt.Println(p.name)
		// Output: Mary
	}
*/
func CallMethod(object any, method string, args []any, convertArgs bool) (_ []any, err error) {
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
ForceCallMethod is an extended version of [CallMethod].

The following code cannot work:

	var p any = person{}
	caller.CallMethod(&p, "SetName", []any{"Jane"}, false)

because `&p` returns a pointer to an interface, not to the `person` type.
The same problem occurs without using that package:

	var tmp any = person{}
	p := &tmp.(person)
	// compiler returns:
	// invalid operation: cannot take address of tmp.(person) (comma, ok expression of type person).

[ForceCallMethod] solves that problem by copying the value and creating a pointer to it using the [reflect] package,
but that solution is slightly slower. In contrast to [CallMethod], it requires a pointer always.
*/
func ForceCallMethod(object any, method string, args []any, convertArgs bool) (_ []any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("cannot call method (%T).%+q: ", object, method), err)
		}
	}()

	return caller.ValidateAndForceCallMethod(object, method, args, convertArgs, caller.DontValidate)
}

/*
CallWither works similar to [CallMethod] with the difference the method must be a wither.

	type Person struct {
	    name string
	}

	func (p Person) WithName(n string) Person {
	    p.Name = n
	    return p
	}

	func main() {
	    p := Person{}
	    p2, _ := caller.CallWither(p, "WithName", "Mary")
	    fmt.Printf("%+v", p2) // {name:Mary}
	}
*/
func CallWither(object any, wither string, args []any, convertArgs bool) (_ any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("cannot call wither (%T).%+q: ", object, wither), err)
		}
	}()

	fn, err := caller.Method(object, wither)
	if err != nil {
		return nil, err
	}

	if err := caller.ValidatorWither.Validate(fn); err != nil {
		return nil, err
	}

	r, err := caller.CallFunc(fn, args, convertArgs)
	if err != nil {
		return nil, err
	}
	return r[0], nil
}

// ForceCallWither calls the given wither (see [CallWither]) using the same approach as [ForceCallMethod].
func ForceCallWither(object any, wither string, args []any, convertArgs bool) (_ any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("cannot call wither (%T).%+q: ", object, wither), err)
		}
	}()

	r, err := caller.ValidateAndForceCallMethod(object, wither, args, convertArgs, caller.ValidatorWither)
	if err != nil {
		return nil, err
	}

	return r[0], nil
}
