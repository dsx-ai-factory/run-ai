// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ngcmetadata

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNGCMetadata(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NGC metadata suite")
}
