/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scale

import (
	"k8s.io/utils/ptr"
	"time"
)

const (
	MinTickerDuration = 2 * time.Minute
)

func GetTimeInUtc(timeVal *time.Time) *int64 {
	if timeVal == nil {
		return nil
	}
	if (*timeVal).IsZero() {
		return nil
	}
	utcTime := (*timeVal).UTC()
	return ptr.To(utcTime.UnixNano())
}

func WaitTime(configuredTimeout time.Duration) time.Duration {
	if MinTickerDuration < configuredTimeout {
		return MinTickerDuration
	}
	return configuredTimeout
}
