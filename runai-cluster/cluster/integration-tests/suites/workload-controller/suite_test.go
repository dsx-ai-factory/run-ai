/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestWorkloadController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workload Controller Integration Tests Suite")
}
