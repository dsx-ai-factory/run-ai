/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package register

import (
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	v2alpha15 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	v2alpha16 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	v2alpha13 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	v2alpha14 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	ipv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	ipv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	iwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v1alpha2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	tpv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	v2alpha12 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	routev1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubscheme "k8s.io/client-go/kubernetes/scheme"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func InitSchema(scheme *runtime.Scheme) {
	utilruntime.Must(kubscheme.AddToScheme(scheme))
	utilruntime.Must(v2alpha1.AddToScheme(scheme)) // interactive workload
	utilruntime.Must(servingv1.AddToScheme(scheme))
	utilruntime.Must(networkingv1.AddToScheme(scheme))
	utilruntime.Must(v2alpha12.AddToScheme(scheme)) // training workload
	utilruntime.Must(ipv1.AddToScheme(scheme))      // interactive policy
	utilruntime.Must(tpv1.AddToScheme(scheme))      // training policy
	utilruntime.Must(v2alpha13.AddToScheme(scheme)) // inference policy
	utilruntime.Must(v2alpha14.AddToScheme(scheme)) // inference workload
	utilruntime.Must(v2alpha15.AddToScheme(scheme)) // distributed policy
	utilruntime.Must(v2alpha16.AddToScheme(scheme)) // distrbuted workload
	utilruntime.Must(runv1.AddToScheme(scheme))     // runai job
	utilruntime.Must(runaiv2.AddToScheme(scheme))   // runai project
	utilruntime.Must(v2alpha2.AddToScheme(scheme))  // runai department
	utilruntime.Must(kaiorg.AddToScheme(scheme))    // kai.resources project/department
	utilruntime.Must(v1alpha2.AddKnownTypes(scheme))
	utilruntime.Must(v1alpha2.AddToScheme(scheme))   // mpi job
	utilruntime.Must(v2beta1.AddToScheme(scheme))    // mpi job
	utilruntime.Must(kubeflowv1.AddToScheme(scheme)) // pytorch job, tfjob
	utilruntime.Must(v1alpha1.AddToScheme(scheme))   // nodepools
	utilruntime.Must(servingv1.AddToScheme(scheme))  // knativeservice
	utilruntime.Must(routev1.AddToScheme(scheme))
	utilruntime.Must(iwv2.AddToScheme(scheme))
	utilruntime.Must(ipv2.AddToScheme(scheme))
	utilruntime.Must(nvapi.AddToScheme(scheme)) // computedomain
	utilruntime.Must(gatewayv1.Install(scheme)) // gateway api
}
