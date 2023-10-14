// Package assert provides tool to test error groups:
//
//	func TestMyCode(t *testing.T) {
//		err := errors.Group(fmt.Errorf("error 1"), fmt.Errorf("error 2"))
//		assert.EqualErrorGroup(t, err, []string{"error 1", "error 2"})
//	}
//
// This package requires https://github.com/stretchr/testify.
package assert
