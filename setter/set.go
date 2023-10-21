package setter

/*
Set assigns the value `val` to the field `field` on the struct `strct`.
Unexported fields are supported.

	type Person struct {
		Name string
	}
	p := Person{}
	_ = setter.Set(&p, "Name", "Jane", false)
	fmt.Println(p) // {Jane}
*/
func Set(strct any, field string, val any, convert bool) error {
	return set(strct, field, val, convert)
}
