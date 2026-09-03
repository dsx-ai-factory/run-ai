/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package issue

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/issue"
	mpiv2beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ComputeResourceGrouper struct {
	issue.DefaultResourceGrouper
}

func (rg *ComputeResourceGrouper) GetObjectResource(resource client.Object) (metav1.GroupVersionKind, error) {
	defaultResourceGrouper := issue.DefaultResourceGrouper{}

	if resource == nil {
		return defaultResourceGrouper.GetObjectResource(resource)
	}

	var objectResource metav1.GroupVersionKind
	switch resource.(type) {
	case *runaiJobV1.RunaiJob:
		objectResource.Kind = constants.KindRunaiJob
	case *kubeflowv1.PyTorchJob:
		objectResource.Kind = constants.KindPyTorchJob
	case *mpiv2beta1.MPIJob:
		objectResource.Kind = constants.KindMpiJob
	case *kubeflowv1.TFJob:
		objectResource.Kind = constants.KindTFJob
	case *kubeflowv1.XGBoostJob:
		objectResource.Kind = constants.KindXGBoostJob
	case *kubeflowv1.JAXJob:
		objectResource.Kind = constants.KindJAXJob
	default:
		return defaultResourceGrouper.GetObjectResource(resource)
	}

	return objectResource, nil
}

func (rg *ComputeResourceGrouper) GetIssueGroupByGvk(gvk metav1.GroupVersionKind) issue.Group {
	switch gvk.Kind {
	case constants.KindRunaiJob, constants.KindPyTorchJob, constants.KindMpiJob, constants.KindTFJob, constants.KindXGBoostJob, constants.KindJAXJob:
		return issue.Compute
	default:
		return rg.DefaultResourceGrouper.GetIssueGroupByGvk(gvk)
	}
}

func (rg *ComputeResourceGrouper) GetIssueGroupByResource(resource client.Object) issue.Group {
	gvk, err := rg.GetObjectResource(resource)
	if err != nil {
		return ""
	}

	return rg.GetIssueGroupByGvk(gvk)
}

func NewIssue(object client.Object, err error, isCritical bool, desc string) *issue.Issue {
	return issue.New(object, err, isCritical, desc, &ComputeResourceGrouper{})
}
