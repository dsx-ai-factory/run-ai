package pointer

import "time"

func ToString(s string) *string {
	return &s
}

func ToBool(b bool) *bool {
	return &b
}

func ToTime(t time.Time) *time.Time {
	return &t
}
