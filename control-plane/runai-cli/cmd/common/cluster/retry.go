/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"time"
)

const (
	SleepTimeDivider = 10
)

func GetIterationMinimalTime(defaultSleep, waitTimeout time.Duration) time.Duration {
	newSleepTime := waitTimeout / SleepTimeDivider
	if newSleepTime < defaultSleep {
		return newSleepTime
	}
	return defaultSleep
}
