package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type typ struct{ value int }

func (a typ) Compare(b typ) int {
	if a.value == b.value {
		return 0
	}
	return 1
}

func TestDistinct(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		actual := Distinct([]string{"a", "a", "b", "b"})
		assert.Equal(t, []string{"a", "b"}, actual)
	})

	t.Run("integer", func(t *testing.T) {
		actual := Distinct([]int{1, 1, 3, 3, 2})
		assert.Equal(t, []int{1, 3, 2}, actual)
	})

	t.Run("object", func(t *testing.T) {
		actual := Distinct([]typ{{1}, {1}, {2}})
		assert.Equal(t, []typ{{1}, {2}}, actual)
	})
}

func TestUnion(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		actual := Union([]string{"a", "a", "b"}, []string{"b", "c"})
		assert.Equal(t, []string{"a", "b", "c"}, actual)
	})

	t.Run("integer", func(t *testing.T) {
		actual := Union([]int{1, 1, 2, 3}, []int{2, 2, 3, 4}, []int{3, 4, 5})
		assert.Equal(t, []int{1, 2, 3, 4, 5}, actual)
	})

	t.Run("integer_order", func(t *testing.T) {
		actual := Union([]int{1, 2, 2, 3}, []int{10, 10, 3, 6}, []int{4, 2, 8})
		assert.Equal(t, []int{1, 2, 3, 10, 6, 4, 8}, actual)
	})

	t.Run("object", func(t *testing.T) {
		actual := Union([]typ{{1}, {1}, {2}}, []typ{{1}, {3}})
		assert.Equal(t, []typ{{1}, {2}, {3}}, actual)
	})
}

func TestIntersect(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		actual := Intersect([]string{"a", "a", "b"}, []string{"b", "c"})
		assert.Equal(t, []string{"b"}, actual)
	})

	t.Run("integer", func(t *testing.T) {
		actual := Intersect([]int{1, 1, 3, 2}, []int{2, 10, 3, 4}, []int{2, 3, 4, 5})
		assert.Equal(t, []int{3, 2}, actual)
	})

	t.Run("object", func(t *testing.T) {
		actual := Intersect([]typ{{1}, {1}, {2}}, []typ{{1}, {3}})
		assert.Equal(t, []typ{{1}}, actual)
	})
}

func TestDifference(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		actual := Difference([]string{"a", "a", "b"}, []string{"b", "c"})
		assert.Equal(t, []string{"a", "c"}, actual)
	})

	t.Run("integer", func(t *testing.T) {
		actual := Difference([]int{1, 1, 3, 2}, []int{2, 10, 3, 4}, []int{2, 3, 4, 5})
		assert.Equal(t, []int{1, 10, 5}, actual)
	})

	t.Run("object", func(t *testing.T) {
		actual := Difference([]typ{{1}, {1}, {2}}, []typ{{1}, {3}})
		assert.Equal(t, []typ{{2}, {3}}, actual)
	})
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name   string
		slice  []int
		remove []int
		expect []int
	}{
		{"common", []int{1, 2, 3, 4, 5}, []int{2, 4}, []int{1, 3, 5}},
		{"partial-non-exist", []int{1, 2, 3, 4, 5}, []int{2, 6}, []int{1, 3, 4, 5}},
		{"all-non-exist", []int{1, 2, 3}, []int{6}, []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Delete(tt.slice, tt.remove...)
			assert.Equal(t, tt.expect, actual)
		})
	}
}
