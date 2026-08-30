package testing

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAnyArg(t *testing.T) {
	arg := AnyArg()
	values := []driver.Value{int64(1), "hello", time.Now(), nil, 3.14}
	for _, v := range values {
		require.True(t, arg.Match(v))
	}
}

func TestAnyTime(t *testing.T) {
	arg := AnyTime()

	timeValues := []driver.Value{time.Now(), time.Now().Add(1 * time.Hour), time.Now().Add(2 * time.Hour)}
	for _, v := range timeValues {
		require.True(t, arg.Match(v))
	}

	nonTimeValues := []driver.Value{int64(1), "hello", nil, 3.14}
	for _, v := range nonTimeValues {
		require.False(t, arg.Match(v))
	}
}
