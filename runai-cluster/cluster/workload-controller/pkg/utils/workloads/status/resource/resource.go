/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource

import (
	"fmt"
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	mpiv2beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

func GetObjectResource(object client.Object) (metav1.GroupVersionKind, error) {
	if object == nil {
		return metav1.GroupVersionKind{}, fmt.Errorf("object is nil %v", object)
	}

	var objectResource metav1.GroupVersionKind
	switch object.(type) {
	case *kubeflowv1.PyTorchJob:
		objectResource.Kind = constants.KindPyTorchJob
	case *networkingv1.Ingress:
		objectResource.Kind = constants.KindIngress
	case *runaiJobV1.RunaiJob:
		objectResource.Kind = constants.KindRunaiJob
	case *mpiv2beta1.MPIJob:
		objectResource.Kind = constants.KindMpiJob
	case *kubeflowv1.TFJob:
		objectResource.Kind = constants.KindTFJob
	case *kubeflowv1.XGBoostJob:
		objectResource.Kind = constants.KindXGBoostJob
	case *kubeflowv1.JAXJob:
		objectResource.Kind = constants.KindJAXJob
	case *v1.Service:
		objectResource.Kind = constants.KindService
	case *servingv1.Service:
		objectResource = metav1.GroupVersionKind{
			Group:   constants.KnativeServiceGroup,
			Version: constants.KnativeServiceVersion,
			Kind:    constants.KnativeServiceKind,
		}
	case *servingv1.Configuration:
		objectResource = metav1.GroupVersionKind{
			Group:   constants.KnativeServiceGroup,
			Version: constants.KnativeServiceVersion,
			Kind:    constants.KnativeConfigurationKind,
		}
	case *servingv1.Revision:
		objectResource = metav1.GroupVersionKind{
			Group:   constants.KnativeServiceGroup,
			Version: constants.KnativeServiceVersion,
			Kind:    constants.KnativeRevisionKind,
		}
	case *openshiftV1.Route:
		objectResource.Kind = constants.KindRoute
	case *gatewayv1.HTTPRoute:
		objectResource.Kind = constants.KindHTTPRoute
	case *gatewayv1.GRPCRoute:
		objectResource.Kind = constants.KindGRPCRoute
	case *v1.PersistentVolumeClaim:
		objectResource.Kind = constants.KindPvc
	case *nvapi.ComputeDomain:
		objectResource.Kind = constants.KindComputeDomain
	case *v1.Secret:
		objectResource.Kind = constants.KindSecret
	case *lwsv1.LeaderWorkerSet:
		objectResource.Kind = constants.KindLeaderWorkerSet
	default:
		return metav1.GroupVersionKind{Kind: object.GetObjectKind().GroupVersionKind().Kind}, fmt.Errorf("unknown resource type %v", object)
	}
	return objectResource, nil
}
