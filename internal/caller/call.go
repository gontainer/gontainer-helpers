package caller

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gontainer/gontainer-helpers/grouperror"
	helpersReflect "github.com/gontainer/gontainer-helpers/internal/reflect"
)

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

// CallFunc calls the given func.
//
// fn.Kind() MUST BE equal to [reflect.Func]
func CallFunc(fn reflect.Value, params []interface{}, convertParams bool) (res []interface{}, err error) {
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
		v, errC := helpersReflect.ValueOf(p, convertTo, convertParams)
		if errC != nil {
			errs = append(errs, grouperror.Prefix(fmt.Sprintf("arg%d: ", i), errC))
		}
		paramsRef[i] = v
	}
	if len(errs) > 0 {
		return nil, grouperror.Join(errs...)
	}

	var result []interface{}
	for _, v := range fn.Call(paramsRef) {
		result = append(result, v.Interface())
	}

	return result, nil
}
