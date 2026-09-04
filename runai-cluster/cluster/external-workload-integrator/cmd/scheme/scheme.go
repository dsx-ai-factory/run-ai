/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scheme

import (
	"context"

	dynamov1alpha1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1alpha1"
	dynamov1beta1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1beta1"
	"github.com/rs/zerolog/log"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	devv1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	kubeflowv1beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/kubeflownotebook/v1beta1"
	kubeflowv2beta1 "github.com/kubeflow/mpi-operator/pkg/apis/kubeflow/v2beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	openshiftV1 "github.com/openshift/api/route/v1"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	nim "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nim/v1alpha1"
	nvcfv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nvcf/v1"
	rayv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/ray/v1"
	seldonv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/v1"
	seldonv1alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/v1alpha2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	infv1 "sigs.k8s.io/gateway-api-inference-extension/api/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// InitScheme initializes the scheme with all supported types.
// Return params:
// - additionalGVKs - list of additional GVKs that were added to the scheme based on Kartas
// - unregisteredGVKs - list of GVKs that are not registered in the cluster, but should be
func InitScheme(scheme *runtime.Scheme) ([]schema.GroupVersionKind, []schema.GroupVersionKind) {
	utilruntime.Must(kartav1alpha1.AddToScheme(scheme))
	utilruntime.Must(dromosv1alpha1.AddToScheme(scheme)) // WorkloadRunner (read for workloads-v2 phase/message)
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
	utilruntime.Must(kubeflowv1.AddToScheme(scheme)) // pytorchjob, tfjob, xgboostjob, jaxjob
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
	utilruntime.Must(nim.AddToScheme(scheme))
	utilruntime.Must(nvcfv1.AddToScheme(scheme))
	utilruntime.Must(resourcev1.AddToScheme(scheme))
	utilruntime.Must(gwv1.Install(scheme))
	utilruntime.Must(gwv1alpha2.Install(scheme))
	utilruntime.Must(infv1.Install(scheme))
	utilruntime.Must(dynamov1alpha1.AddToScheme(scheme))
	utilruntime.Must(dynamov1beta1.AddToScheme(scheme))

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

	additionalGvks, unregisteredGvks, err := readAdditionalGVKsFromKartas(allTypesThatNeedSupport)
	if err != nil {
		return []schema.GroupVersionKind{}, []schema.GroupVersionKind{}
	}
	for _, gvk := range additionalGvks {
		utils.AddUnstructuredToSchemeIfNeeded(gvk, scheme)
	}

	return additionalGvks, unregisteredGvks
}

// readAdditionalGVKsFromKartas
// returns:
// - additionalGVKs - list of additional GVKs that were added to the scheme based on Kartas
// - unregisteredGVKs - list of GVKs that are not registered in the cluster, but should be
// - error - in case of error
func readAdditionalGVKsFromKartas(nativelySupportedGvks map[schema.GroupVersionKind]bool) ([]schema.GroupVersionKind, []schema.GroupVersionKind, error) {
	kartas, err := listKartas()
	if err != nil {
		return []schema.GroupVersionKind{}, []schema.GroupVersionKind{}, err
	}

	additionalGvks := []schema.GroupVersionKind{}
	unregisteredGvks := []schema.GroupVersionKind{}
	for _, karta := range kartas {
		isEligible, crdExists := isKartaEligibleForHandling(&karta, nativelySupportedGvks)
		gvk := utils.GetGvkOfKarta(&karta)
		if gvk == nil {
			continue
		}

		if !isEligible {
			if crdExists {
				unregisteredGvks = append(unregisteredGvks, *gvk)
			}
			continue
		}

		additionalGvks = append(additionalGvks, *gvk)

		additionalGvks = append(additionalGvks, utils.GetAdditionalGvksOfKarta(&karta)...)
	}

	return additionalGvks, unregisteredGvks, nil
}

func listKartas() ([]kartav1alpha1.Karta, error) {
	clientConfig := ctrl.GetConfigOrDie()

	// an initial scheme that contains only karta to list Kartas
	initialScheme := runtime.NewScheme()
	utilruntime.Must(kartav1alpha1.AddToScheme(initialScheme))

	k8sClient, err := client.New(clientConfig, client.Options{
		Scheme: initialScheme,
	})
	if err != nil {
		log.Fatal().Msgf("Failed creating writerClient client, error: %s", err.Error())
		return nil, err
	}

	kartaList := &kartav1alpha1.KartaList{}
	err = k8sClient.List(context.Background(), kartaList)
	if err != nil {
		log.Error().Msgf("Failed to list Kartas: %v", err)
		return nil, err
	}
	return kartaList.Items, nil
}

// isKartaEligibleForHandling
// returns: allConditionsPassed, crdExists
func isKartaEligibleForHandling(karta *kartav1alpha1.Karta, nativelySupportedGvks map[schema.GroupVersionKind]bool) (bool, bool) {
	if karta == nil {
		return false, false
	}

	gvkOfKarta := utils.GetGvkOfKarta(karta)
	if gvkOfKarta == nil {
		return false, false
	}

	_, isNativelySupported := nativelySupportedGvks[*gvkOfKarta]
	if isNativelySupported {
		// if it's already natively supported - no need to handle it here as additional GVK
		return false, false
	}

	if karta.DeletionTimestamp != nil {
		return false, false
	}

	crdExists := false
	rbacReady := false
	for _, condition := range karta.Status.Conditions {
		if condition.Type == string(runv1alpha1.CRDExists) &&
			condition.Status == metav1.ConditionTrue {
			crdExists = true
			continue
		}
		if condition.Type == string(runv1alpha1.RBACReady) &&
			condition.Status == metav1.ConditionTrue {
			rbacReady = true
			continue
		}
	}

	return crdExists && rbacReady, crdExists
}
