package copier_test

import (
	"testing"

	"github.com/gontainer/gontainer-helpers/copier"
	"github.com/stretchr/testify/assert"
)

func TestForceCopy(t *testing.T) {
	t.Run("[]int to []interface{}", func(t *testing.T) {
		var (
			from = []int{1, 2, 3}
			to   []interface{}
		)

		err := copier.ForceCopy(from, &to)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{1, 2, 3}, to)
	})
	t.Run("[]interface{} to []int", func(t *testing.T) {
		var (
			from = []interface{}{1, 2, 3}
			to   []int
		)

		err := copier.ForceCopy(from, &to)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, to)
	})
	t.Run("[]int to [N]int", func(t *testing.T) {
		var (
			from = []int{1, 2, 3}
			to   [3]int
		)

		err := copier.ForceCopy(from, &to)
		assert.NoError(t, err)
		assert.Equal(t, [3]int{1, 2, 3}, to)
	})
	t.Run("[]int to [N]int #2", func(t *testing.T) {
		var (
			from = []int{1, 2, 3}
			to   [2]int
		)

		err := copier.ForceCopy(from, &to)
		assert.EqualError(t, err, "cannot cast `[]int` (length 3) to `[2]int`")
		assert.Empty(t, to)
	})
	t.Run("[N]int to [N-1]int", func(t *testing.T) {
		var (
			from = [3]int{1, 2, 3}
			to   [2]int
		)

		err := copier.ForceCopy(from, &to)
		assert.EqualError(t, err, "cannot cast `[3]int` to `[2]int`")
		assert.Empty(t, to)
	})
	t.Run("[N]int to [N+1]int", func(t *testing.T) {
		var (
			from = [3]int{1, 2, 3}
			to   [4]int
		)

		err := copier.ForceCopy(from, &to)
		assert.NoError(t, err)
		assert.Equal(t, [4]int{1, 2, 3, 0}, to)
	})
	t.Run("[N]interface{} to [N+1]interface{}", func(t *testing.T) {
		var (
			from = [3]interface{}{6, 7, 8}
			to   [4]interface{}
		)

		err := copier.ForceCopy(from, &to)
		assert.NoError(t, err)
		assert.Equal(t, [4]interface{}{6, 7, 8, nil}, to)
	})
	t.Run("[N]int to []int", func(t *testing.T) {
		var (
			from = [3]int{1, 2, 3}
			to   []int
		)

		err := copier.ForceCopy(from, &to)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, to)
	})
	t.Run("[N]int to [N]uint", func(t *testing.T) {
		var (
			from = [3]int{1, 2, 3}
			to   [3]uint
		)

		err := copier.ForceCopy(from, &to)
		assert.NoError(t, err)
		assert.Equal(t, [3]uint{1, 2, 3}, to)
	})
}
