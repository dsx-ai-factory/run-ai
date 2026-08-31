/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package keepalive

import "time"

const TypeNoKeepalive = "no-keepalive"

type NoKeepalive struct{}

func (nk *NoKeepalive) GetType() string {
	return TypeNoKeepalive
}

func (nk *NoKeepalive) Execute() error {
	return nil
}

func (nk *NoKeepalive) Channel() <-chan time.Time {
	return nil
}
