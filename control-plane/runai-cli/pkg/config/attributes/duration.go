package attributes

import "time"

const (
	typeName = "duration"
)

type Duration struct {
	duration time.Duration
}

func (d *Duration) String() string {
	return d.duration.String()
}

func (d *Duration) Set(duration string) error {
	var err error
	d.duration, err = time.ParseDuration(duration)
	return err
}

func (d *Duration) Type() string {
	return typeName
}

func (d *Duration) ToSec() int64 {
	return int64(d.duration.Seconds())
}
