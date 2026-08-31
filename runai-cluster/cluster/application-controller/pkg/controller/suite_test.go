/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	ginkgoTesting "testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCrud(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller tests")
}
