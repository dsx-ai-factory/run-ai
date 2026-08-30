package retry

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	attempts := 2
	tries := 0

	err := Retry(context.Background(), func(uint) error {
		tries++
		return nil
	}, attempts)

	assert.NoError(t, err)
	assert.Equal(t, tries, 1)
}

func TestRetry_Fails(t *testing.T) {
	attempts := 2
	tries := 0

	err := Retry(context.Background(), func(uint) error {
		tries++
		return fmt.Errorf("error")
	}, attempts)

	assert.Equal(t, err.Error(), "error")
	assert.Equal(t, tries, attempts)
}

func TestRetryStreamErrors_FailsAndNotStuck(t *testing.T) {
	attempts := 2
	tries := 0
	errs := make(chan error, 1)

	err := RetryStreamErrors(context.Background(), errs, func(uint) error {
		tries++
		return fmt.Errorf("error RetryStreamError")
	}, attempts)

	assert.Equal(t, err.Error(), "error RetryStreamError")
	assert.Equal(t, (<-errs).Error(), "error RetryStreamError")
	assert.Equal(t, tries, attempts)
}
