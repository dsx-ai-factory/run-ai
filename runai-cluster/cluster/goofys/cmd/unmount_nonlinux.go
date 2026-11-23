// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package main

import "github.com/kahing/goofys/internal"

// tryUnmount retries normal unmounts on macOS and other non-Linux systems.
func tryUnmount(mountPoint string) error {
	return internal.TryUnmount(mountPoint)
}
