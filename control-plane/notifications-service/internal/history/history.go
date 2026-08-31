/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package history

import (
	"fmt"
	"time"
)

var ErrEventNotFound = fmt.Errorf("event not found")

type Event struct {
	Id             string                 `json:"id"`
	SourceListener string                 `json:"source_listener"`
	Properties     map[string]interface{} `json:"properties"`
	CreatedAt      time.Time              `json:"created_at"`
	Read           bool                   `json:"read"`
}

type EventFilter struct {
	TenantId string
	UserId   string
	Read     *bool
}

type CleanupFilter struct {
	MaxEventsPerUser *int
	MaxAge           *time.Duration
}

func (cf *CleanupFilter) WithMaxEventsPerUser(n int) *CleanupFilter {
	cf.MaxEventsPerUser = &n
	return cf
}

func (cf *CleanupFilter) WithMaxAge(d time.Duration) *CleanupFilter {
	cf.MaxAge = &d
	return cf
}
