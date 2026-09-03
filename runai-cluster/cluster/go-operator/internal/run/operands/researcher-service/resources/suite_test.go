/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	RegistryTestValue = "gcr.io/run-ai-prod"
)

func TestOperand(t *testing.T) {
	t.Setenv(common.DefaultRegistryEnvVarName, RegistryTestValue)

	RegisterFailHandler(Fail)
	RunSpecs(t, "researcher service")
}
