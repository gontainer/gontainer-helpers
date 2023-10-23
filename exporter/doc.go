/*
Package exporter provides sets of function to export variables to a GO code.

	s, _ := exporter.Export([3]any{nil, 1.5, "hello world"})
	fmt.Println(s)
	// Output: [3]interface{}{nil, float64(1.5), "hello world"}
*/
package exporter
