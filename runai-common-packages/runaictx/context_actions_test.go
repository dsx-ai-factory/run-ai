package runaictx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToStringArray(t *testing.T) {
	var group1 interface{} = "group-1"
	var group2 interface{} = "group-2"
	arr := []interface{}{group1, group2}
	groups := toStringArray(arr)
	assert.Equal(t, groups, []string{"group-1", "group-2"})
}

func TestExtractCtxFieldValueStringVal(t *testing.T) {
	tests := map[string]struct {
		ctxMap    map[string]interface{}
		fieldName string
		expect    any
	}{
		`string val`: {
			ctxMap: map[string]interface{}{
				"tenant_id": "12345",
			},
			fieldName: "tenant_id",
			expect:    "12345",
		},
		`string arr`: {
			ctxMap: map[string]interface{}{
				"arr": []string{"first", "second"},
			},
			fieldName: "arr",
			expect:    []string{"first", "second"},
		},
		`interface arr`: {
			ctxMap: map[string]interface{}{
				"arr": []interface{}{interface{}("first"), interface{}("second")},
			},
			fieldName: "arr",
			expect:    []string{"first", "second"},
		},
		`bad field name`: {
			ctxMap: map[string]interface{}{
				"arr": []interface{}{interface{}("first"), interface{}("second")},
			},
			fieldName: "another",
			expect:    nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			value := extractCtxFieldValue(test.ctxMap, test.fieldName)
			assert.Equal(t, test.expect, value)
		})
	}
}
