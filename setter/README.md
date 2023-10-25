# Setter

Package setter allows for manipulation of a value of a field of any struct.

```go
person := struct {
    name string
}{}
_ = setter.Set(&person, "name", "Mary", false)
fmt.Println(person.name)
// Output: Mary
```

See [examples](examples_test.go).