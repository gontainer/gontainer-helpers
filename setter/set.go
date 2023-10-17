package setter

// Deprecated: in the new major version this func won't convert the value.
// Use ConvertAndSet.
var Set = ConvertAndSet

// TODO
//func Set(strct interface{}, field string, val interface{}) error {
//	return set(strct, field, val, false)
//}

// ConvertAndSet assigns the value `val` to the field `field` on the struct `strct`.
// Unexported fields are supported.
func ConvertAndSet(strct interface{}, field string, val interface{}) error {
	return set(strct, field, val, true)
}
