package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterEq(t *testing.T) {
	assert.Equal(t, "clusterId==abc", FilterEq("clusterId", "abc"))
}

func TestValidateNoCommas(t *testing.T) {
	assert.NoError(t, ValidateNoCommas("clusterId", "abc"))
	assert.NoError(t, ValidateNoCommas("clusterId", ""))

	err := ValidateNoCommas("clusterId", "a,b")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "clusterId", "message names the offending field")
}
