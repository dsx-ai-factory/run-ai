/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package defaulter

// Defaulter is an interface types must implement to default their fields
type Defaulter interface {
	// SetDefaultsWhereNeeded iterates over all struct fields and sets defaults for missing values
	SetDefaultsWhereNeeded()
}
