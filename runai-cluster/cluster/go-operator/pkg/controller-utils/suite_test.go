/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_utils

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	testEnv          *envtest.Environment
	k8sClient        client.Client
	apiAccessContext context.Context
	suite            = "Controller Utility Functions"
)

func TestControllerUtils(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, suite)
}
