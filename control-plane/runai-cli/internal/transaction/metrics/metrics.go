/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"encoding/json"
)

type Metrics struct {
	start   int64
	end     int64
	timers  *[]Timer
	counter *map[string]int
}

type ReportMetrics struct {
	Start   int64          `json:"cmd_start"`
	End     int64          `json:"cmd_end"`
	Timers  []Timer        `json:"timers"`
	Counter map[string]int `json:"counter"`
}

func NewMetrics() *Metrics {
	timers := make([]Timer, 0)
	counter := make(map[string]int)
	return &Metrics{start: GetUnixTime(), timers: &timers, counter: &counter}
}

func (m *Metrics) MarshalJSON() ([]byte, error) {
	return json.Marshal(ReportMetrics{
		Start:   m.start,
		End:     m.end,
		Timers:  *m.timers,
		Counter: *m.counter,
	})
}

func (m *Metrics) StopTxnTimer() {
	m.end = m.getUnixTimeFromStart()
}

func (m *Metrics) CreateTimer(name string) int {
	*m.timers = append(*m.timers, *NewTimer(name, m.start))
	return len(*m.timers) - 1
}

func (m *Metrics) StopTimer(id int) {
	(*m.timers)[id].TickEndTime()
}

func (m *Metrics) SetCounter(name string, value int) {
	(*m.counter)[name] = value
}

func (m *Metrics) IncCounter(name string) {
	_, ok := (*m.counter)[name]
	if !ok {
		(*m.counter)[name] = 0
	}
	(*m.counter)[name]++
}

func (m *Metrics) getUnixTimeFromStart() int64 {
	return GetUnixTime() - m.start
}
