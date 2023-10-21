package caller

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/grouperror"
	"github.com/gontainer/gontainer-helpers/internal/caller"
)

// TODO rename params to args

// Call calls the given function with the given arguments.
// It returns values returned by the function in a slice.
func Call(fn any, params []any, convertParams bool) (_ []any, err error) {
	defer func() {
		if err != nil {
			// TODO maybe better:
			//err = grouperror.Prefix(fmt.Sprintf("cannot call %T: ", fn), err)
			err = grouperror.Prefix(fmt.Sprintf("cannot call func %T: ", fn), err)
		}
	}()

	v, err := caller.Func(fn)
	if err != nil {
		return nil, err
	}
	return caller.CallFunc(v, params, convertParams)
}

/*
CallProvider works similar to Call with the difference it requires a provider as the first argument.
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
func CallProvider(provider any, params []any, convertParams bool) (_ any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("cannot call provider %T: ", provider), err)
		}
	}()

	fn, err := caller.FuncProvider(provider)
	if err != nil {
		return nil, err
	}

	results, err := caller.CallFunc(fn, params, convertParams)
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

// CallByName works similar to Call with the difference it calls the method by the name over the given receiver.
func CallByName(object any, method string, params []any, convertParams bool) (_ []any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("cannot call method (%T).%+q: ", object, method), err)
		}
	}()

	fn, err := caller.Method(object, method)
	if err != nil {
		return nil, err
	}
	return caller.CallFunc(fn, params, convertParams)
}

/*
CallWitherByName works similar to CallByName with the difference the method must be a wither.

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
func CallWitherByName(object any, wither string, params []any, convertParams bool) (_ any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("cannot call wither (%T).%+q: ", object, wither), err)
		}
	}()

	fn, err := caller.Wither(object, wither)
	if err != nil {
		return nil, err
	}

	r, err := caller.CallFunc(fn, params, convertParams)
	var v any
	if len(r) > 0 {
		v = r[0]
	}
	return v, err
}
