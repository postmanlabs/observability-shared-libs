package go_ast

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akitasoftware/akita-libs/visitors"
)

type Counter struct {
	count int
}

func countInt(c visitors.Context, v interface{}, x interface{}) bool {
	counter := v.(*Counter)
	if _, ok := x.(int); ok {
		counter.count++
	}
	return true
}

func extendContext(c visitors.Context, v interface{}, x interface{}) visitors.Context {
	return c
}

func TestApply(t *testing.T) {
	counter := new(Counter)
	vm := visitors.NewVisitorManager(visitors.NewContext(), counter, countInt, extendContext)
	d := []int{1, 2, 3}
	Apply(PREORDER, vm, d)
	assert.Equal(t, 3, counter.count)
}
