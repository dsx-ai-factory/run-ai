/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package create

// Item key names for my-credential create --item (key=value pairs per credential type).
const (
	ItemKeyGenericSecretKey   = "key"
	ItemKeyGenericSecretVal   = "value"
	ItemKeyDockerRegistryUser = "user"
	ItemKeyDockerRegistryPass = "password"
	ItemKeyDockerRegistryURL  = "url"
	ItemKeyNgcApiKey          = "ngc-api-key"
)
