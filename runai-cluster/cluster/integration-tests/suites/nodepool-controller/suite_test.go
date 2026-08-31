/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNodePoolController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nodepool Controller Integration Tests Suite")
}
