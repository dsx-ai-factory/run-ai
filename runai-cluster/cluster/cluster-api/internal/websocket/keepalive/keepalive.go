package keepalive

import (
	"time"
)

type Keepalive interface {
	GetType() string
	Execute() error
	Channel() <-chan time.Time
}
