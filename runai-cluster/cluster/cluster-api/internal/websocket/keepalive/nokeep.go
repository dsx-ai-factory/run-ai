package keepalive

import "time"

const TypeNoKeepalive = "no-keepalive"

type NoKeepalive struct{}

func (nk *NoKeepalive) GetType() string {
	return TypeNoKeepalive
}

func (nk *NoKeepalive) Execute() error {
	return nil
}

func (nk *NoKeepalive) Channel() <-chan time.Time {
	return nil
}
