package reports

import "fmt"

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testData = map[string][]string{
	"a,c,b": {"1", "1.2"},
	"c,b,a": {"1", "1.2"},
	"a,b,c": {"1", "1.2"},
	"c,a,b": {"1", "1.2"},
}

func TestSortedMapIterator_SimpleIteration(t *testing.T) {
	expectedKeysOrder := []string{
		"a,b,c",
		"a,c,b",
		"c,a,b",
		"c,b,a",
	}

	i := 0
	for key := range iterateSortedResultMap(testData) {
		assert.Equal(t, expectedKeysOrder[i], key)
		i++
	}

	assert.Equal(t, len(expectedKeysOrder), i)
}

func TestSortedMapIterator_EmptyMap(t *testing.T) {
	data := map[string][]string{}

	called := false
	for range iterateSortedResultMap(data) {
		called = true
	}
	assert.False(t, called)
}

func TestSortedMapIterator_SingleElement(t *testing.T) {
	data := map[string][]string{
		"a": {"val1", "val2"},
	}

	iterator := iterateSortedResultMap(data)

	i := 0
	for key, values := range iterator {
		assert.Equal(t, "a", key)
		assert.Equal(t, data["a"], values)
		i++
	}

	assert.Equal(t, len(data), i)
}

func TestSortKeys_WorksAsExpected(t *testing.T) {
	val1 := "10ff8bcb-1bd9-4df0-908d-484a113c517d"
	val2 := "20ff8bcb-1bd9-4df0-908d-484a113c517d"
	label1_1 := fmt.Sprintf("%s_dep1", val1)
	label1_2 := fmt.Sprintf("%s_dep2", val1)
	label2_1 := fmt.Sprintf("%s_dep1", val2)
	label2_2 := fmt.Sprintf("%s_dep2", val2)

	data := map[string][]string{
		label2_2: {},
		label1_1: {},
		label1_2: {},
		label2_1: {},
	}
	expected := []string{label1_1, label1_2, label2_1, label2_2}

	res := sortKeys(data)
	assert.Equal(t, expected, res)
}
