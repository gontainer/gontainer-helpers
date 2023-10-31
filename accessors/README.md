# Accessors

Package accessors allows for read and write operations on fields of any struct.

```go
person := struct {
    name string
}{}
_ = accessors.Set(&person, "name", "Mary", false)
fmt.Println(person.name)
// Output: Mary
```

See [examples](examples_test.go).
