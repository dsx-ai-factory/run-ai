/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package network

import "net"

// Conn is an interface that extends net.Conn for now only for mocking purposes
type Conn interface {
	net.Conn
}
