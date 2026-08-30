package metrics

import (
	"encoding/json"
	"time"
)

type Timer struct {
	name     string
	cmdStart int64
	start    int64
	end      int64
}

type ReportTimer struct {
	Name  string `json:"name"`
	Start int64  `json:"start"`
	End   int64  `json:"end"`
}

func NewTimer(name string, cmdStart int64) *Timer {
	return &Timer{name: name, cmdStart: cmdStart, start: GetUnixTime() - cmdStart}
}

func (t *Timer) ResetStartTime() {
	t.start = t.calcTimeDiff()
}

func (t *Timer) TickEndTime() {
	t.end = t.calcTimeDiff()
}

func (t *Timer) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, t)
}

func (t *Timer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ReportTimer{
		Name:  t.name,
		Start: t.start,
		End:   t.end,
	})
}

func (t *Timer) calcTimeDiff() int64 {
	return GetUnixTime() - t.cmdStart
}

func GetUnixTime() int64 {
	return time.Now().UTC().UnixMilli()
}
