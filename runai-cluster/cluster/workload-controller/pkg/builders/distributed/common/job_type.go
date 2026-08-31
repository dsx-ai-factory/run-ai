/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"errors"
	"fmt"

	v1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder return a builder based on the job
func GetBuilder(options factory.BuilderOptions, k8sClient client.Client, self *v2alpha1.DistributedWorkload,
	podLabels map[string]string, podAnnotations map[string]string) (factory.Builder, error) {
	//
	//  determine the distributed job we're about to build
	//
	var builderType string

	var toObject interface{}

	if self.Spec.JobType == "" {
		return nil, fmt.Errorf("mandatory field jobType is not provided")
	}

	switch self.Spec.JobType {
	case constants.DistributedMpiJob:
		builderType = factory.MpiBuilder
		if self.Spec.PyTorchJob != nil {
			return nil, fmt.Errorf("pyTorchJob is forbidden for distributed workload of job type %s", self.Spec.JobType)
		}
		toObject = &v2beta1.MPIJob{}
	case constants.DistributedPyTorchJob:
		builderType = factory.PyTorchBuilder
		if self.Spec.MpiJob != nil {
			return nil, fmt.Errorf("mpiJob is forbidden for distributed workload of job type %s", self.Spec.JobType)
		}
		toObject = &v1.PyTorchJob{}
	case constants.DistributedTfJob:
		builderType = factory.TFJobBuilder
		if self.Spec.MpiJob != nil {
			return nil, fmt.Errorf("mpiJob is forbidden for distributed workload of job type %s", self.Spec.JobType)
		}
		if self.Spec.PyTorchJob != nil {
			return nil, fmt.Errorf("pyTorchJob is forbidden for distributed workload of job type %s", self.Spec.JobType)
		}
		toObject = &v1.TFJob{}
	case constants.DistributedXGBoostJob:
		builderType = factory.XGBoostJobBuilder
		if self.Spec.MpiJob != nil {
			return nil, fmt.Errorf("mpiJob is forbidden for distributed workload of job type %s", self.Spec.JobType)
		}
		if self.Spec.PyTorchJob != nil {
			return nil, fmt.Errorf("pyTorchJob is forbidden for distributed workload of job type %s", self.Spec.JobType)
		}
		toObject = &v1.XGBoostJob{}
	case constants.KindJAXJob:
		builderType = factory.JAXJobBuilder
		if self.Spec.MpiJob != nil {
			return nil, fmt.Errorf("mpiJob is forbidden for distributed workload of job type %s", self.Spec.JobType)
		}
		if self.Spec.PyTorchJob != nil {
			return nil, fmt.Errorf("pyTorchJob is forbidden for distributed workload of job type %s", self.Spec.JobType)
		}
		toObject = &v1.JAXJob{}
	}

	builderConstructor := factory.BuildersRegistry[builderType]
	if builderConstructor != nil {
		return builderConstructor(self, toObject, k8sClient, options, podLabels, podAnnotations), nil
	}

	return nil, errors.New("failed to allocate builder, this can only happen during testing/debugging")
}
