package lock

import (
	"time"
)

type Record struct {
	LockedAt time.Time
	LockedBy string
}
