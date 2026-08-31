/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

type MyCredentialAlignedRow struct {
	Type        string
	Name        string
	Description string
}

type MyCredentialAlignedData struct {
	Credentials []MyCredentialAlignedRow
}
