package setter

// Set assigns the value `val` to the field `field` on the struct `strct`.
// Unexported fields are supported.
func Set(strct any, field string, val any, convert bool) error {
	return set(strct, field, val, convert)
}
