package common

import (
	"fmt"
	"math"
)

func Int32PositiveCast(number int) (int32, error) {
	if number > math.MaxInt32 {
		// handle int overflow
		return math.MaxInt32, fmt.Errorf("number %d exceeds max value %d", number, math.MaxInt32)
	} else if number < 0 {
		return 0, fmt.Errorf("non valid positive or zero number: %d", number)
	}
	return int32(number), nil
}
