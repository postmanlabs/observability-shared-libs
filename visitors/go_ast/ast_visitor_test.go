package go_ast

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/akitasoftware/akita-libs/visitors"
)

type Counter struct {
	enterCount int
	leaveCount int
}

func countIntEnter(c Context, v interface{}, x interface{}) Cont {
	counter := v.(*Counter)
	if _, ok := x.(int); ok {
		counter.enterCount++
	}
	return Continue
}

func countIntLeave(c Context, v interface{}, x interface{}, cont Cont) Cont {
	counter := v.(*Counter)
	if _, ok := x.(int); ok {
		counter.leaveCount++
	}
	return cont
}

func extendContext(c Context, x interface{}) Context {
	return c
}

func TestApply(t *testing.T) {
	counter := new(Counter)
	vm := NewVisitorManager(NewContext(), counter, countIntEnter, DefaultVisitChildren, countIntLeave, extendContext)
	d := []int{1, 2, 3}
	Apply(vm, d)
	assert.Equal(t, 3, counter.enterCount)
	assert.Equal(t, 3, counter.leaveCount)
}
