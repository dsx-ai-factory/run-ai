/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package keepalive

import (
	"time"
)

type Keepalive interface {
	GetType() string
	Execute() error
	Channel() <-chan time.Time
}
