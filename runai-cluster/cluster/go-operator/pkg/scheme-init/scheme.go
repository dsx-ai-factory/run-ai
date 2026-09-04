/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scheme_init

import (
	"fmt"

	cmgrv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	runv2alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2alpha1"
	schedulingv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v1"
	schedulingv2alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha1"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	ocpconf "github.com/openshift/api/config/v1"
	ocpv1 "github.com/openshift/api/operator/v1"
	osv1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	nodev1beta1 "k8s.io/api/node/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var (
	scheme *runtime.Scheme
)

func init() {
	scheme = runtime.NewScheme()

	f := []func(s *runtime.Scheme) error{
		clientgoscheme.AddToScheme,
		runaiv1.AddToScheme,
		schedulingv1.AddToScheme,
		cmgrv1.AddToScheme,
		monitoringv1.AddToScheme,
		osv1.AddToScheme,
		runv2alpha1.AddToScheme,
		runv1alpha1.AddToScheme,
		dromosv1alpha1.AddToScheme,
		runaiv2.AddToScheme,
		schedulingv2alpha1.AddToScheme,
		nodev1beta1.AddToScheme,
		admissionv1beta1.AddToScheme,
		schedulingv2alpha2.AddToScheme,
		networkingv1.AddToScheme,
		apiextensionsv1.AddToScheme,
		kaiv2alpha2.AddToScheme,
		ocpv1.AddToScheme,
		kaiv2.AddToScheme,
		kaiv1.AddToScheme,
		kaiv1alpha1.AddToScheme,
		kaiorg.AddToScheme,
		ocpconf.AddToScheme,
		vpav1.AddToScheme,
		gwv1.Install,
		//+kubebuilder:scaffold:scheme
	}
	for _, f := range f {
		err := f(scheme)
		if err != nil {
			panic(fmt.Errorf("could not add types to scheme: %v", err))
		}
	}
}

func Scheme() *runtime.Scheme {
	return scheme
}
