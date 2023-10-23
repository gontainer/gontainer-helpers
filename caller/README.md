# Caller

This package provides functions that allow calling other functions with unknown arguments.

```go
sum := func(a, b int) int {
    return a + b
}

returns, _ := caller.Call(sum, []any{2, 3}, false)
fmt.Println(returns) // [5]
```

See [examples](examples_test.go).
