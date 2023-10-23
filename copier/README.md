# Copier

Package copier allows for copying a value to a variable with an unknown type.

**Copy value**

```go
var (
	from = 5 // the type of the variable `to` can be different from the type of the variable `from`
	to   any // as long as the value of the `from` is assignable to the `to`
)
_ = copier.Copy(from, &to, false)
fmt.Println(to)
// Output: 5
```

**Convert & copy value**

```go
var (
	from = int(5) // uint is not assignable to int,
	to   uint     // but [copier.Copy] can convert the type
)
_ = copier.Copy(from, &to, true)
fmt.Println(to)
// Output: 5
```
