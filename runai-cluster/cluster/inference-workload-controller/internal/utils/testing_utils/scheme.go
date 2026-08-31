/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package register

import (
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	v15 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	v2alpha13 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	v2alpha14 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	ipv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	v12 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtime2 "k8s.io/apimachinery/pkg/util/runtime"
	scheme2 "k8s.io/client-go/kubernetes/scheme"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
)

func InitSchema(scheme *runtime.Scheme) {
	runtime2.Must(scheme2.AddToScheme(scheme))
	runtime2.Must(v1.AddToScheme(scheme))
	runtime2.Must(v12.AddToScheme(scheme))
	runtime2.Must(v2alpha13.AddToScheme(scheme)) // inference policy
	runtime2.Must(v2alpha14.AddToScheme(scheme)) // inference workload
	runtime2.Must(v2.AddToScheme(scheme))        // runai project
	runtime2.Must(v2alpha2.AddToScheme(scheme))  // runai department, podgroup
	runtime2.Must(v1alpha1.AddToScheme(scheme))  // nodepools, external workload
	runtime2.Must(v1.AddToScheme(scheme))        // knativeservice
	runtime2.Must(v15.AddToScheme(scheme))
	runtime2.Must(iwv2.AddToScheme(scheme))
	runtime2.Must(ipv2.AddToScheme(scheme))
	runtime2.Must(kaiv2alpha2.AddToScheme(scheme)) // podgroup
}
