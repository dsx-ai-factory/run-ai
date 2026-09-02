/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/config"
)

func TestDatabase(t *testing.T) {
	RegisterFailHandler(Fail)

	config.MockConfig()

	RunSpecs(t, "Database Suite")
}
