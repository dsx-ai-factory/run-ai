/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test_deprecated

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/logger"
	"go.uber.org/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClusterSync(t *testing.T) {
	logger.InitLogger()
	zap.S().Infof("Started service tests")
	RegisterFailHandler(Fail)
	MockEnvDeprecated(t)
	RunSpecs(t, "cluster-sync service Test Suite")
}
