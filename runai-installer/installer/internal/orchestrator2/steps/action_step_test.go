package steps

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithExtraTimeout(t *testing.T) {
	s := NewActionStep("x")
	assert.Equal(t, time.Duration(0), s.ExtraTimeout(), "no extra timeout by default")

	WithExtraTimeout(s, 90*time.Minute)
	assert.Equal(t, 90*time.Minute, s.ExtraTimeout(), "WithExtraTimeout sets the step's extra timeout")
}
