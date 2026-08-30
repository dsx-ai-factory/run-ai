package sampler_actions

import (
	"io"
	"time"
)

func BestEffortRead(reader io.Reader) string {
	buf := make([]byte, 4096)
	n, _ := reader.Read(buf)
	return string(buf[:n])
}

func MaxTime(t1, t2 *time.Time) *time.Time {
	if t1.After(*t2) {
		return t1
	}
	return t2
}
