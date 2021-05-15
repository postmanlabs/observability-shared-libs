package go_ast_pair

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/akitasoftware/akita-libs/visitors"
)

type Counter struct {
	enterCount int
	leaveCount int
}

func countIntEnter(c PairContext, v interface{}, left, right interface{}) Cont {
	counter := v.(*Counter)
	_, leftOk := left.(int)
	_, rightOk := right.(int)
	if leftOk || rightOk {
		counter.enterCount++
	}
	return Continue
}

func countIntLeave(c PairContext, v interface{}, left, right interface{}, cont Cont) Cont {
	counter := v.(*Counter)
	_, leftOk := left.(int)
	_, rightOk := right.(int)
	if leftOk || rightOk {
		counter.leaveCount++
	}
	return cont
}

func extendContext(c PairContext, left, right interface{}) {}

func TestApply(t *testing.T) {
	counter := new(Counter)
	vm := NewPairVisitorManager(NewPairContext(), counter, countIntEnter, DefaultVisitChildren, countIntLeave, extendContext)

	left := make(map[interface{}]int)
	left["moo"] = 0
	left[struct{ a int }{a: 1}] = 0

	right := make(map[interface{}]int)
	right[struct{ a int }{a: 1}] = 0
	right[2] = 0

	Apply(vm, left, right)
	assert.Equal(t, 3, counter.enterCount)
	assert.Equal(t, 3, counter.leaveCount)
}
