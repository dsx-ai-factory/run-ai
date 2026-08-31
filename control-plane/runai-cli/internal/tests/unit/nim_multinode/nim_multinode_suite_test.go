/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nim_multinode_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNimMultiNode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI NIM Multi-Node Tests Suite")
}
