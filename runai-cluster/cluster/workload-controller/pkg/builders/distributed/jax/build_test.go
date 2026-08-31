/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jax

import (
	"testing"

	mnnvlapi "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	distributedtesting "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/testing"
)

func TestBuildersJAXJob(t *testing.T) {
	toTest := []common.TestConfig{
		{TestName: "basic", TestEnvVars: &common.TestEnvVars{ADSEncryptionKey: common.ADSEncryptionKey}},
		{TestName: "more_params"},
		{TestName: "limits"},
		{TestName: "volumes"},
		{TestName: "ports"},
		{TestName: "ephemeral"},
		{TestName: "command_args"},
		{TestName: "pod_affinity"},
		{TestName: "basic_policy"},
		{TestName: "s3_bucket"},
		{TestName: "restart_policy"},
		{TestName: "image_pull_secrets", TestEnvVars: &common.TestEnvVars{ADSEncryptionKey: common.ADSEncryptionKey}},
	}
	for _, curTest := range toTest {
		Init()
		common.Run(t, &curTest, nil, nil, nil,
			distributedtesting.GetDistributedBuilder)
	}
}

func TestBuildersMNNVL(t *testing.T) {
	toTest := []common.TestConfig{
		{TestName: "basic_mnnvl_worker_required", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModeRequired,
			MasterMode: mnnvlapi.ModeNone,
		}},
		{TestName: "basic_mnnvl_worker_preferred", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModePreferred,
			MasterMode: mnnvlapi.ModeNone,
		}},
		{TestName: "node_affinity_with_mnnvl_required", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModeRequired,
			MasterMode: mnnvlapi.ModeNone,
		}},
		{TestName: "node_affinity_with_mnnvl_preferred", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModePreferred,
			MasterMode: mnnvlapi.ModeNone,
		}},
	}

	for _, curTest := range toTest {
		Init()
		common.Run(t, &curTest, nil, nil, nil,
			distributedtesting.GetDistributedBuilder)
	}
}
