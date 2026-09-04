/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/testutils"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDao(t *testing.T) {
	RegisterFailHandler(Fail)
	defer runai_config.FlushAll()

	testutils.RegisterPostgresSuite(config.Schema, "file://../../database/migrations")

	RunSpecs(t, "DAO Suite")
}
