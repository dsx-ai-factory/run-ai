/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package fips

import (
	"crypto/fips140"

	"github.com/run-ai/runai/runai-common-packages/logging"
)

func init() {
	logging.Init(nil)
	logging.Infof("FIPS 140-3 status: enabled=%t enforced=%t version=%q",
		fips140.Enabled(), fips140.Enforced(), fips140.Version())
}
