package cmd

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestCmdBuilder(t *testing.T) {
	cb := newCmdBuilder("go test").arg("zero", "").
		argNoBlank("one", "", "two").arg("", "three").argNoBlank("", "four")
	assert.Equal(t, cb.cmd, "go", "unexpected cmd")
	assert.Equal(t, len(cb.args), 8, "unexpected number of args")
	assert.Equal(t, cb.args[0], "test", "unexpected value on arg")
	assert.Equal(t, cb.args[1], "zero", "unexpected value on arg")
	assert.Equal(t, cb.args[2], "", "unexpected value on arg")
	assert.Equal(t, cb.args[3], "one", "unexpected value on arg")
	assert.Equal(t, cb.args[4], "two", "unexpected value on arg")
	assert.Equal(t, cb.args[5], "", "unexpected value on arg")
	assert.Equal(t, cb.args[6], "three", "unexpected value on arg")
	assert.Equal(t, cb.args[7], "four", "unexpected value on arg")
}
