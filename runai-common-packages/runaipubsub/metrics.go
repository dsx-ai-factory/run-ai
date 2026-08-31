/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaipubsub

import (
	"strconv"

	"github.com/run-ai/runai/runai-common-packages/metrics"
)

const (
	pubsubEventCount = "pubsub_event_count"
	eventType        = "event_type"
	errorLabel       = "error"
)

var (
	metricLabels = []string{
		eventType,
		errorLabel,
	}
)

type pubsubMetrics struct{}

func newPubsubMetrics() *pubsubMetrics {
	return &pubsubMetrics{}
}

func (m *pubsubMetrics) register() {
	metrics.CounterVec(pubsubEventCount).Register("Count of pubsub events processed by subscriber", metricLabels)
}

func (m *pubsubMetrics) recordEvent(topic string, err error) {
	metrics.CounterVec(pubsubEventCount).Get().WithLabels(map[string]string{
		eventType:  topic,
		errorLabel: strconv.FormatBool(err != nil),
	}).Inc()
}

func (m *pubsubMetrics) registerEvent(topic string) {
	metrics.CounterVec(pubsubEventCount).Get().WithLabels(map[string]string{
		eventType:  topic,
		errorLabel: strconv.FormatBool(false),
	}).Add(0)
	metrics.CounterVec(pubsubEventCount).Get().WithLabels(map[string]string{
		eventType:  topic,
		errorLabel: strconv.FormatBool(true),
	}).Add(0)
}
