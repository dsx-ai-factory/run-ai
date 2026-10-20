/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

const (
	LoginWithAccessKeyCreds = "Authenticating with access key credentials"
	FailedAuthMsg           = "authentication server issue"
	EOFAuthErr              = "connection issue, please check your connection to the authentication server and try again"
	FailedSaveTokenMsg      = "failed to save access key token"
	LoginSuccessfullyMsg    = "Authentication was successful"
	SetClusterProperties    = "Cluster %s set successfully"

	AuthenticationSecretFileReadError = "failed to read access key secret file"
)

const (
	IdTokenKey     = "id_token"
	AccessTokenKey = "access_token"
	CLIClientID    = "cli"
)

const (
	AccessKeyOption = "Access key"
	FilePath        = "File Path"
)

var (
	AccessKeyLoginOptions = []string{AccessKeyOption, FilePath}
)
