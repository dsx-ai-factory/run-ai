/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nim_workloads

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestNimServiceComposer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "nim service composer tests suite")
}
