/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	"fmt"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	runaiJobV1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	mpiv2beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	jobInterface "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/jobs"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetJobAdapterFactory(object client.Object, k8sClient client.Client) (jobInterface.JobAdapterFactory, error) {
	switch object.(type) {
	case *runaiJobV1.RunaiJob:
		return &RunaiJobAdapterFactory{}, nil
	case *kubeflowv1.PyTorchJob:
		return &DistributedJobAdapterFactory{k8sClient: k8sClient}, nil
	case *kubeflowv1.TFJob:
		return &DistributedJobAdapterFactory{k8sClient: k8sClient}, nil
	case *kubeflowv1.XGBoostJob:
		return &DistributedJobAdapterFactory{k8sClient: k8sClient}, nil
	case *mpiv2beta1.MPIJob:
		return &DistributedJobAdapterFactory{k8sClient: k8sClient}, nil
	case *kubeflowv1.JAXJob:
		return &DistributedJobAdapterFactory{k8sClient: k8sClient}, nil

	default:
		return nil, fmt.Errorf("unspported job facrory for type %v", object)
	}
}
