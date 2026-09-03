/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type Verbosity int

const (
	SlimVerbosity Verbosity = iota
	MetaDataVerbosity
	VerboseVerbosity
)
