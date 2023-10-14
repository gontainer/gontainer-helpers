package caller

import (
	"fmt"
	"reflect"

	"github.com/gontainer/gontainer-helpers/errors"
	helpersReflect "github.com/gontainer/gontainer-helpers/internal/reflect"
)

func call(fn reflect.Value, params ...interface{}) (res []interface{}, err error) {
	if fn.Kind() != reflect.Func {
		return nil, fmt.Errorf("expected `%s`, `%T` given", reflect.Func.String(), fn.Type().String())
	}

	fnType := reflectType{fn.Type()}

	if len(params) > fnType.NumIn() && !fnType.IsVariadic() {
		return nil, errors.New("too many input arguments")
	}

	minParams := fnType.NumIn()
	if fnType.IsVariadic() {
		minParams--
	}
	if len(params) < minParams {
		return nil, errors.New("too few input arguments")
	}

	paramsRef := make([]reflect.Value, len(params))
	errs := make([]error, 0, len(params))
	for i, p := range params {
		convertTo := fnType.inVariadicAware(i)
		// TODO don't convert for existing funcs, add new funcs prefixed by `Convert`
		v, errC := helpersReflect.Convert(p, convertTo)
		if errC != nil {
			errs = append(errs, errors.PrefixedGroup(fmt.Sprintf("arg%d: ", i), errC))
		}
		paramsRef[i] = v
	}
	if len(errs) > 0 {
		return nil, errors.Group(errs...)
	}

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		err = fmt.Errorf("%s", r)
	}()

	var result []interface{}
	for _, v := range fn.Call(paramsRef) {
		result = append(result, v.Interface())
	}

	return result, nil
}

// Call calls the given function with the given arguments.
// It returns values returned by the function in a slice.
func Call(fn interface{}, params ...interface{}) ([]interface{}, error) {
	return call(reflect.ValueOf(fn), params...)
}

var (
	errorInterface = reflect.TypeOf((*error)(nil)).Elem()
)

// CallProvider works similar to Call with the difference it requires a provider as the first argument.
// Provider is a function which returns 1 or 2 values.
// The second return value which is optional must be a type of error.
//
//	p := func() (interface{}, error) {
//	    db, err := sql.Open("mysql", "user:password@/dbname")
//	    if err != nil {
//	         return nil, err
//	    }
//
//	    db.SetConnMaxLifetime(time.Minute * 3)
//	    db.SetMaxOpenConns(10)
//	    db.SetMaxIdleConns(10)
//
//	    return db, nil
//	}
//
//	mysql, err := CallProvider(p)
func CallProvider(provider interface{}, params ...interface{}) (interface{}, error) {
	t := reflect.TypeOf(provider)
	if t.Kind() != reflect.Func {
		return nil, fmt.Errorf("provider must be kind of `%s`, `%s` given", reflect.Func.String(), t.Kind().String())
	}
	if t.NumOut() == 0 || t.NumOut() > 2 {
		return nil, fmt.Errorf("provider must return 1 or 2 values, given function returns %d values", t.NumOut())
	}
	if t.NumOut() == 2 && !t.Out(1).Implements(errorInterface) {
		return nil, errors.New("second value returned by provider must implement error interface")
	}

	results, err := Call(provider, params...)
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
func CallByName(object interface{}, method string, params ...interface{}) ([]interface{}, error) {
	val := reflect.ValueOf(object)
	fn := val.MethodByName(method)
	for !fn.IsValid() && (val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface) {
		val = val.Elem()
		fn = val.MethodByName(method)
	}

	if !fn.IsValid() {
		return nil, fmt.Errorf("invalid func `%T`.%+q", object, method)
	}
	return call(fn, params...)
}

// CallWitherByName works similar to CallByName with the difference the method must be a wither.
//
//	type Person struct {
//	    name string
//	}
//
//	func (p Person) WithName(n string) Person {
//	    p.Name = n
//	    return p
//	}
//
//	func main() {
//	    p := Person{}
//	    p2, _ := caller.CallWitherByName(p, "WithName", "Mary")
//	    fmt.Printf("%+v", p2) // {name:Mary}
//	}
func CallWitherByName(object interface{}, wither string, params ...interface{}) (interface{}, error) {
	val := reflect.ValueOf(object)
	fn := val.MethodByName(wither)
	for !fn.IsValid() && (val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface) {
		val = val.Elem()
		fn = val.MethodByName(wither)
	}

	if !fn.IsValid() {
		return nil, fmt.Errorf("invalid wither `%T`.%+q", object, wither)
	}

	t := fn.Type()

	if t.NumOut() != 1 {
		return nil, fmt.Errorf("wither must return 1 value, given function returns %d values", t.NumOut())
	}

	r, err := call(fn, params...)
	if r == nil {
		return nil, fmt.Errorf("`%T`.%+q: %w", object, wither, err)
	}
	return r[0], err
}

type reflectType struct {
	reflect.Type
}

// inVariadicAware works almost same as reflect.Type.In,
// but it returns t.In(t.NumIn() - 1).Elem() for t.isVariadic() && i >= t.NumIn().
func (t reflectType) inVariadicAware(i int) reflect.Type {
	last := t.NumIn() - 1
	if i > last {
		i = last
	}
	r := t.In(i)
	if t.IsVariadic() && i == last {
		r = r.Elem()
	}
	return r
}
