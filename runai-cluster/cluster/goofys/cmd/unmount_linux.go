// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import "syscall"

// tryUnmount detaches busy Linux mounts so pod cleanup can finish.
func tryUnmount(mountPoint string) error {
	if err := unmount(mountPoint, 0); err != nil {
		return unmount(mountPoint, syscall.MNT_DETACH)
	}
	return nil
}
