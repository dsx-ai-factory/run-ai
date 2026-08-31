/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodeexclusion

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNodeExclusion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node Exclusion Suite")
}
