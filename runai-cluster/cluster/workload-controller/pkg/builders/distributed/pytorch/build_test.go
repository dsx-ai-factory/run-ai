/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pytorch

import (
	"context"
	"testing"

	mnnvlapi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	distp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	distw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	common2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/common"
	v1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"

	distributedtesting "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/testing"
)

func TestBuildersPytorchNvidia(t *testing.T) {
	toTest := []common.TestConfig{
		{TestName: "basic"},
		{TestName: "more_params"},
		{TestName: "limits"},
		{TestName: "volumes"},
		{TestName: "ports"},
		{TestName: "ephemeral"},
		{TestName: "command_args"},
		{TestName: "pod_affinity"},
		{TestName: "master_basic"},
		{TestName: "master_more_params"},
		{TestName: "master_only_params"},
		{TestName: "master_only_privileged_true"},
		{TestName: "master_only_privileged_false"},
		{TestName: "master_volumes"},
		{TestName: "no_master"},
		{TestName: "basic_policy"},
		{TestName: "master_with_policy"},
		{TestName: "priority_class"},

		{TestName: "elastic"},
		{TestName: "elastic_default_replicas"},
		{TestName: "configmap_volumes"},
		{TestName: "configmap_master_volumes"},
		{TestName: "s3_bucket"},
		{TestName: "image_pull_secrets"},
		{TestName: "basic_with_topology"},
	}

	for _, curTest := range toTest {
		Init()
		common.Run(t, &curTest, nil, nil, nil,
			distributedtesting.GetDistributedBuilder)
	}
}
func TestBuildersMNNVL(t *testing.T) {
	toTest := []common.TestConfig{
		{TestName: "basic_mnnvl", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModePreferred,
			MasterMode: mnnvlapi.ModePreferred,
		}},
		{TestName: "basic_mnnvl_required", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModeRequired,
			MasterMode: mnnvlapi.ModeRequired,
		}},
		{TestName: "basic_mnnvl_worker_required", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModeRequired,
			MasterMode: mnnvlapi.ModeNone,
		}},
		{TestName: "basic_mnnvl_worker_preferred", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModePreferred,
			MasterMode: mnnvlapi.ModeNone,
		}},
		{TestName: "no_master_mnnvl_required", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModeRequired,
			MasterMode: mnnvlapi.None,
		}},
		{TestName: "node_affinity_with_mnnvl_required", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModeRequired,
			MasterMode: mnnvlapi.ModeNone,
		}},
		{TestName: "node_affinity_with_mnnvl_preferred", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModePreferred,
			MasterMode: mnnvlapi.ModeNone,
		}},
		{TestName: "basic_mnnvl_with_topology", AddCtxFn: distributedtesting.AddMnnvlToContextWrapper, ValueToCtx: mnnvlutils.Status{
			WorkerMode: mnnvlapi.ModeRequired,
			MasterMode: mnnvlapi.ModeNone,
		}}}

	for _, curTest := range toTest {
		Init()
		common.Run(t, &curTest, nil, nil, nil,
			distributedtesting.GetDistributedBuilder)
	}
}

func TestElasticValidation(t *testing.T) {
	workload := distw.DistributedWorkload{
		Spec: distp.DistributedSpec{
			DistributedFields: field_groups.DistributedFields{
				JobType: constants.DistributedPyTorchJob,
			},
		},
	}

	distBuilder := Builder{
		DistributedBuilder: common2.DistributedBuilder{
			MergedWorkload: &workload,
		},
		PyTorch: &v1.PyTorchJob{},
	}

	//
	//   replicas is 2 -> the minReplica <= 2 and maxReplicas >= 2
	//
	workload.Spec.Workers = field_types.NewOptionalIntField(ptr.To(2))
	workload.Spec.PyTorchJob = &field_types.PyTorchField{
		Values: &field_types.PyTorchMembers{
			MinReplicas: field_types.NewOptionalIntField(ptr.To(3)),
		},
	}

	err := distBuilder.BuildPytorchElasticSpec(context.Background())
	assert.ErrorContains(t, err, "minReplicas (3) must be no greater than replicas (2)")

	workload.Spec.Workers = field_types.NewOptionalIntField(ptr.To(2))
	workload.Spec.PyTorchJob = &field_types.PyTorchField{
		Values: &field_types.PyTorchMembers{
			MaxReplicas: field_types.NewOptionalIntField(ptr.To(1)),
		},
	}

	err = distBuilder.BuildPytorchElasticSpec(context.Background())
	assert.ErrorContains(t, err, "maxReplicas (1) must be no less than replicas (2)")
}
