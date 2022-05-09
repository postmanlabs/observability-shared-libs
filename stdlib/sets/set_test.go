package sets

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicOperations(t *testing.T) {
	s := NewSet[int]()
	assert.Equal(t, len(s), 0)
	assert.Equal(t, map[int]struct{}(s), map[int]struct{}{})

	s.Insert(1)
	assert.Equal(t, s, NewSet(1))

	s.Intersect(NewSet(1, 2))
	assert.Equal(t, s, NewSet(1))

	s.Union(NewSet(1, 2))
	assert.Equal(t, s, NewSet(1, 2))

	s.Delete(1)
	assert.Equal(t, s, NewSet(2))
}

func TestJson(t *testing.T) {
	s := NewSet[int](3, 2, 1)

	bs, err := json.Marshal(s)
	assert.NoError(t, err)

	var deserialized Set[int]
	err = json.Unmarshal(bs, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, deserialized, s, "s == unmarshal(marshal(s))")
}

func TestJsonOrdering(t *testing.T) {
	bs1, err := json.Marshal(NewSet(1, 2, 3))
	assert.NoError(t, err)

	bs2, err := json.Marshal(NewSet(3, 2, 1))
	assert.NoError(t, err)

	assert.Equal(t, string(bs1), string(bs2), "marshal(s) == marshal(s)")
}
