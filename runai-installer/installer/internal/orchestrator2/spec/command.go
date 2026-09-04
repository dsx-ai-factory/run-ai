/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

type Command string

const (
	CommandInstall    Command = "install"
	CommandUpgrade    Command = "upgrade"
	CommandUninstall  Command = "uninstall"
	CommandAirgapPack Command = "airgap-pack"
	CommandAirgapLoad Command = "airgap-load"
)

func (c Command) IsKnown() bool {
	switch c {
	case CommandInstall, CommandUpgrade, CommandUninstall, CommandAirgapPack, CommandAirgapLoad:
		return true
	}
	return false
}
