/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package os

type Type string

const (
	Windows Type = "windows"
	Linux   Type = "linux"
	Darwin  Type = "darwin"
	Unknown Type = "unknown"
)

func (t Type) String() string {
	return string(t)
}

type Arch string

const (
	AMD64       Arch = "amd64"
	ARM64       Arch = "arm64"
	X86         Arch = "386"
	UnknownArch Arch = "unknown"
)

func (a Arch) String() string {
	return string(a)
}
