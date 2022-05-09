package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	assert.Equal(t, Add(1, 2), 3)
	assert.Equal(t, Add(-1, 2), 1)
	assert.Equal(t, Add(-1.5, 2.0), 0.5)
}

func TestMin(t *testing.T) {
	assert.Equal(t, Min(1, 2), 1)
	assert.Equal(t, Min(2, -2), -2)
	assert.Equal(t, Min(2.5, -2.0), -2.0)
}

func TestMax(t *testing.T) {
	assert.Equal(t, Max(1, 2), 2)
	assert.Equal(t, Max(2, -2), 2)
	assert.Equal(t, Max(2.5, -2.0), 2.5)
}
