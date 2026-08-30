package utils

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJsonNumberToInt(t *testing.T) {

	result, err := JsonNumberToInt(int(5))
	assert.NoError(t, err)
	assert.Equal(t, int(5), *result)

	result, err = JsonNumberToInt(int32(5))
	assert.NoError(t, err)
	assert.Equal(t, int(5), *result)

	result, err = JsonNumberToInt(float32(5.0))
	assert.NoError(t, err)
	assert.Equal(t, int(5), *result)

	result, err = JsonNumberToInt("5")
	assert.NoError(t, err)
	assert.Equal(t, int(5), *result)

	result, err = JsonNumberToInt(json.Number("5"))
	assert.NoError(t, err)
	assert.Equal(t, int(5), *result)
}

func TestArgToMap(t *testing.T) {
	k1 := "a.com/key1"
	v1 := "value1"
	k2 := "a.com/key2"
	v2 := "value2"
	args := fmt.Sprintf("%s=%s,%s=%s", k1, v1, k2, v2)
	argsMap := ArgToMap(args)
	assert.Equal(t, v1, argsMap[k1])
	assert.Equal(t, v2, argsMap[k2])
}
