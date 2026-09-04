/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTree(t *testing.T) {
	RegisterFailHandler(Fail)

	config.MockConfig()

	RunSpecs(t, "tree suite")
}
