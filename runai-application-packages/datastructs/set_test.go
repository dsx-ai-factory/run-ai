package datastructs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitializers(t *testing.T) {
	set1 := NewSet[int32]()
	assert.Equal(t, len(set1), 0)

	set2 := NewSet(1, 3, 9)
	assert.Equal(t, len(set2), 3)

	arr := []string{"aaa", "bbb", "ccc", "ddd"}
	set3 := NewSet(arr...)
	assert.Equal(t, len(set3), 4)
}

func TestActions(t *testing.T) {
	set := NewSet(1, 3, 9)
	assert.Equal(t, len(set), 3)
	assert.True(t, set.Has(1))
	assert.False(t, set.Has(2))

	set2 := NewSet(1, 3)
	assert.False(t, set.EqualTo(set2))
	set2.Add(9)
	assert.True(t, set.EqualTo(set2))

	set.Add(2) // set containing 1, 2, 3, 9
	assert.True(t, set.Has(1))
	assert.True(t, set.Has(2))
	assert.Equal(t, len(set), 4)

	set.Add(1)
	set.Add(1)
	set.Add(1)
	assert.True(t, set.Has(1))
	assert.Equal(t, len(set), 4)
	assert.Equal(t, len(set.Keys()), 4)

	set.Delete(1) // set containing 2, 3, 9
	assert.False(t, set.Has(1))
	assert.True(t, set.Has(9))
	assert.Equal(t, len(set), 3)
	assert.Equal(t, len(set.Keys()), 3)

	set.Delete(1) // set still containing 2, 3, 9
	assert.False(t, set.Has(1))
	assert.Equal(t, len(set), 3)
	assert.Equal(t, len(set.Keys()), 3)

	set.Delete(1, 2, 3, 9) // set becomes empty
	assert.Equal(t, len(set), 0)
	assert.Equal(t, len(set.Keys()), 0)

	// empty set equal to empty set
	assert.True(t, set.EqualTo(NewSet[int]()))
	// set is equal to itself
	assert.True(t, set.EqualTo(set))
}
