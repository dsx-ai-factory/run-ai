/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers"
	dwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	ipv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"os"
	"strings"

	devv1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	workflow "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/argo/v1alpha1"
	kubeflowv1beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/kubeflownotebook/v1beta1"
	nvcfv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nvcf/v1"
	rayv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/ray/v1"
	seldonv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/v1"
	seldonv1alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/v1alpha2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	kubeflowv2beta1 "github.com/kubeflow/mpi-operator/pkg/apis/kubeflow/v2beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	openshiftV1 "github.com/openshift/api/route/v1"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	resourcev1beta2 "k8s.io/api/resource/v1beta2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	kubeClient client.Client

	kubeConfig *rest.Config
	scheme     *runtime.Scheme
)

func initConnectivity() error {
	var err error

	if kubeConfig == nil {
		err = initKubeConfig()
		if err != nil {
			return err
		}
	}
	if scheme == nil {
		initScheme()
	}

	if kubeClient == nil {
		kubeClient, err = client.New(kubeConfig, client.Options{
			Scheme: scheme,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func initKubeConfig() error {
	config, err := clientcmd.BuildConfigFromFlags("", getKubeConfigPath())
	if err != nil {
		return err
	}

	// Update throttling parameters matching to the current (v1.25) kubectl
	config.QPS = 50
	config.Burst = 300

	kubeConfig = config
	return nil
}

func getKubeConfigPath() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
	}
	kubeconfig = strings.Split(kubeconfig, ":")[0]

	return kubeconfig
}

func initScheme() {
	scheme = scheme_init.Scheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
	utilruntime.Must(kubeflowv1.AddToScheme(scheme)) // pytorchjob, tfjob, xgboostjob
	utilruntime.Must(kubeflowv2beta1.AddToScheme(scheme))
	utilruntime.Must(kubeflowv1beta1.AddToScheme(scheme)) // notebook
	utilruntime.Must(servingv1.AddToScheme(scheme))       // knativeservice
	utilruntime.Must(openshiftV1.AddToScheme(scheme))
	utilruntime.Must(rayv1alpha1.AddToScheme(scheme))
	utilruntime.Must(rayv1.AddToScheme(scheme))
	utilruntime.Must(seldonv1.AddToScheme(scheme))
	utilruntime.Must(seldonv1alpha2.AddToScheme(scheme))
	utilruntime.Must(kubevirtv1.AddToScheme(scheme))
	utilruntime.Must(devv1alpha2.AddToScheme(scheme))
	utilruntime.Must(workflow.AddToScheme(scheme))
	utilruntime.Must(iwv2.AddToScheme(scheme))
	utilruntime.Must(ipv1.AddToScheme(scheme))
	utilruntime.Must(twv2.AddToScheme(scheme))
	utilruntime.Must(dwv2.AddToScheme(scheme))
	utilruntime.Must(nvcfv1.AddToScheme(scheme))
	utilruntime.Must(infv1.AddToScheme(scheme))
	utilruntime.Must(resourcev1.AddToScheme(scheme))
	utilruntime.Must(resourcev1beta1.AddToScheme(scheme))
	utilruntime.Must(resourcev1beta2.AddToScheme(scheme))

	allTypesThatNeedSupport := supportedhandlers.GetSupportedGVKs()
	allTypesThatNeedSupport[schema.GroupVersionKind{
		Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"}] = true
	allTypesThatNeedSupport[schema.GroupVersionKind{
		Group: "run.ai", Version: "v2alpha1", Kind: "InteractiveWorkload"}] = true
	allTypesThatNeedSupport[schema.GroupVersionKind{
		Group: "run.ai", Version: "v2alpha1", Kind: "DistributedWorkload"}] = true
	allTypesThatNeedSupport[schema.GroupVersionKind{
		Group: "run.ai", Version: "v2alpha1", Kind: "InferenceWorkload"}] = true
	allTypesThatNeedSupport[schema.GroupVersionKind{
		Group: "run.ai", Version: "v2alpha1", Kind: "DistributedInferenceWorkload"}] = true

	for typeToSupport := range allTypesThatNeedSupport {
		utils.AddUnstructuredToSchemeIfNeeded(typeToSupport, scheme)
	}
}
