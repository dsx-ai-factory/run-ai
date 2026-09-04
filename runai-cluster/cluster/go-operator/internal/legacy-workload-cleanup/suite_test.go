/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLegacyWorkloadCleanup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Legacy Workload Cleanup Suite")
}
