/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mpi

import (
	"testing"

	mnnvlapi "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"

	distributedtesting "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/testing"
)

func TestBuildersMpiNvidia(t *testing.T) {
	toTest := []common.TestConfig{
		{TestName: "basic"},
		{TestName: "basic-data-volume"},
		{TestName: "more_params"},
		{TestName: "limits"},
		{TestName: "volumes"},
		{TestName: "ports"},
		{TestName: "ephemeral"},
		{TestName: "command_args"},
		{TestName: "pod_affinity"},
		{TestName: "probes"},
		{TestName: "master_basic", TestEnvVars: &common.TestEnvVars{ADSEncryptionKey: common.ADSEncryptionKey}},
		{TestName: "master_more_params"},
		{TestName: "master_only_params"},
		{TestName: "master_only_privileged_true"},
		{TestName: "master_only_privileged_false"},
		{TestName: "master_volumes"},
		{TestName: "basic_policy"},
		{TestName: "master_with_policy"},

		{TestName: "slots_per_worker"},
		{TestName: "s3_bucket"},
		{TestName: "s3_master_bucket"},
		{TestName: "restart_policy"},
		{TestName: "launcher_create_policy"},
		{TestName: "image_pull_secrets", TestEnvVars: &common.TestEnvVars{ADSEncryptionKey: common.ADSEncryptionKey}},
		{TestName: "basic_with_topology"},
	}

	for _, curTest := range toTest {
		Init()
		common.Run(t, &curTest, nil, nil, nil, distributedtesting.GetDistributedBuilder)
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
		{TestName: "basic_mnnvl_worker_required_with_topology", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModeRequired,
			MasterMode: mnnvlapi.ModeNone,
		}},
	}

	for _, curTest := range toTest {
		Init()
		common.Run(t, &curTest, nil, nil, nil,
			distributedtesting.GetDistributedBuilder)
	}
}

func TestBuildersMpiPodLabelsAndAnnotations(t *testing.T) {
	toTest := []common.TestConfig{
		{TestName: "pod-labels-annotations"},
	}

	labels := map[string]string{
		"label": "value",
	}
	annotations := map[string]string{
		"annotation": "value",
	}

	for _, curTest := range toTest {
		Init()
		common.Run(t, &curTest, nil, labels, annotations, distributedtesting.GetDistributedBuilder)
	}
}
