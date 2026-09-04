/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"fmt"
	"time"
)

type GaugeMetric struct {
	name      string
	val       float64
	timestamp time.Time
	labels    [][2]any // not using map, since it is not ordered and tests fail
}

func NewGaugeMetric(name string, value float64) *GaugeMetric {
	return &GaugeMetric{name: name, val: value, timestamp: time.Now(), labels: make([][2]any, 0)}
}

func (m *GaugeMetric) WithTime(t time.Time) *GaugeMetric {
	m.timestamp = t
	return m
}

func (m *GaugeMetric) AddLabel(label, val any) *GaugeMetric {
	m.labels = append(m.labels, [2]any{label, val})
	return m
}

func (m *GaugeMetric) String() string {
	return fmt.Sprintf("%s%s %f %d\n", m.name, getPromqlLabels(m.labels), m.val, m.timestamp.Unix())
}
