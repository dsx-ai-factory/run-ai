package concurrency

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestMap(t *testing.T) {
	set := NewConcurrentMap[int](5)
	assert.Equal(t, set.Size(), 0)

	set.Store("key1", 2)
	set.Store("key2", 3)
	assert.Equal(t, set.Size(), 2)

	set.Store("key2", 22)
	assert.Equal(t, set.Size(), 2)

	key1, found := set.Load("key1")
	assert.True(t, found)
	assert.Equal(t, key1, 2)

	key2, found := set.Load("key2")
	assert.True(t, found)
	assert.Equal(t, key2, 22)

	dflt, found := set.Load("keyx")
	assert.False(t, found)
	assert.Equal(t, dflt, 5)

	keys := make([]string, 0)
	values := make([]int, 0)

	set.Range(func(key string, value int) bool {
		keys = append(keys, key)
		values = append(values, value)
		return true
	})

	sort.Strings(keys)
	sort.Ints(values)

	assert.Equal(t, keys, []string{
		"key1",
		"key2",
	})

	assert.Equal(t, values, []int{
		2,
		22,
	})
	
	set.Delete("key1")
	assert.Equal(t, set.Size(), 1)

	_, found = set.Load("key1")
	assert.Equal(t, found, false)

	_, found = set.Load("key2")
	assert.Equal(t, found, true)
}
