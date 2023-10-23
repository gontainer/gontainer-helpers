# Exporter

This package provides sets of function to export variables to a GO code.

```go
s, _ := exporter.Export([3]any{nil, 1.5, "hello world"})
fmt.Println(s)
// Output: [3]interface{}{nil, float64(1.5), "hello world"}
```

See [examples](examples_test.go).
