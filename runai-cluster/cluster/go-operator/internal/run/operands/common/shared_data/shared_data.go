/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package shared_data

import "sync"

type Key string

const (
	NmcMgmtPrometheusCA Key = "nmc-mgmt-prometheus-ca"
)

// SharedData represents the shared data between operands
// NOTE: This is not thread-safe, and should not be used to
// share data between operands that may run concurrently
type SharedData struct {
	data sync.Map
}

func NewSharedData() *SharedData {
	return &SharedData{}
}

func (s *SharedData) Get(key Key) (interface{}, bool) {
	value, ok := s.data.Load(key)
	return value, ok
}

func (s *SharedData) Set(key Key, value interface{}) {
	s.data.Store(key, value)
}
