/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

// auth command
const (
	// cobra
	AuthUse      = "auth"
	AuthShortMsg = "Authentication related commands"
)

// get-token command
const (
	// cobra
	GetTokenUse   = "get-token"
	GetTokenShort = "Print an access token for the logged-in user"

	// flags
	FlagNameOutput    = "output"
	FlagShortOutput   = "o"
	FlagDefaultOutput = "kubeconfig"
	FlagHelpOutput    = "Output format: plaintext or kubeconfig"

	// errors
	FailedGetTokenMsg = "Failed to fetch token, try logout and in again"
)
