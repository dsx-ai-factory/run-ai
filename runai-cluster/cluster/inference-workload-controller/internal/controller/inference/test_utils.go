/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference

import (
	"context"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	inferenceaggregator "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/aggregator"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/issue"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	ssv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	builderscommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/managers"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/status/update"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func buildPodWithRevisionLabel(name, namespace, revisionName string, podStatusPhase v1.PodPhase) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
			Labels: map[string]string{
				clustercommonconstants.KnativeRevisionLabel: revisionName,
			},
		},
		Status: v1.PodStatus{
			Phase: podStatusPhase,
		},
	}
}

func getInferenceWorkload(ctx context.Context, client k8sclient.Client, name, namespace string) *infw.InferenceWorkload {
	tmp := &infw.InferenceWorkload{}
	err := client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, tmp)
	Expect(err).ToNot(HaveOccurred())
	return tmp
}

func getKnativeService(ctx context.Context, client k8sclient.Client, name, namespace string) *knv1.Service {
	tmp := &knv1.Service{}
	err := client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, tmp)
	if !errors.IsNotFound(err) {
		Expect(err).ToNot(HaveOccurred())
	}
	return tmp
}

func createController(client k8sclient.Client, updater update.StatusUpdater) *Controller {

	knativeBuilder := builders.NewKnativeBuilder(client)
	pvcBuilder := builderscommon.NewPvcBuilder(client)
	serviceBuilder := builderscommon.NewServiceBuilder(client)
	externalUrlBuilder := builderscommon.NewExternalsUrlsBuilder(client, false)
	credentialsBuilder := builderscommon.NewCredentialsBuilder(client)
	commonResourceBuilder := builderscommon.NewShardResourcesBuilder(serviceBuilder, externalUrlBuilder, pvcBuilder, credentialsBuilder)

	statusManger := update.NewSimpleStatusUpdaterManager(updater)

	statusAggregator := inferenceaggregator.NewStatusCalculator()
	k8sManager := managers.NewResourceManager(client, client, &issue.InferenceResourceGrouper{})

	jobController := NewController(client, knativeBuilder,
		commonResourceBuilder, statusManger, nil, statusAggregator, k8sManager,
		managers.ResourceOptions{})
	return jobController
}

func checkOwnedResource(resources []common.ObservedResource, targetKind string) bool {
	for _, resource := range resources {
		if strings.Contains(resource.Kind.String(), targetKind) {
			return true
		}
	}
	return false
}

func buildBasicInferenceWorkload(workloadName string, namespace string) *infw.InferenceWorkload {
	inferenceWorkload := &infw.InferenceWorkload{
		TypeMeta: metav1.TypeMeta{
			Kind:       "InferenceWorkload",
			APIVersion: "run.ai/v2alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName,
			Namespace: namespace,
		},
		Spec: ssv1.InferenceSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: field_types.NewOptionalStringField(workloadName, ""),
				},
				NetworkingFields: field_groups.NetworkingFields{
					Ports: &field_types.PortsField{
						Value: &map[string]field_types.PortItemAndRules{
							"serving-port": {
								Value: &field_types.PortItemValue{
									ServiceType:   ptr.To(constants.ServiceTypeServingPort),
									ContainerPort: ptr.To(8080),
								},
							},
						},
					},
				},
			},
		},
	}

	return inferenceWorkload
}

func setKnativeService(ctx context.Context, k8sClient k8sclient.Client, workloadName, namespace string, latestReadyRevisionName, latestCreatedRevisionName *string, conditions *duckv1.Conditions) *knv1.Service {
	// Get the updated Knative service from the cluster
	knativeService := getKnativeService(ctx, k8sClient, workloadName, namespace)

	if latestReadyRevisionName != nil {
		knativeService.Status.LatestReadyRevisionName = *latestReadyRevisionName
	}

	if latestCreatedRevisionName != nil {
		knativeService.Status.LatestCreatedRevisionName = *latestCreatedRevisionName
	}

	if conditions != nil {
		knativeService.Status.Conditions = *conditions
	}

	Expect(k8sClient.Update(ctx, knativeService)).To(Succeed())

	// return the updated Knative service if needed after the successful update
	return knativeService
}

func buildKnativeServiceCondition(conditionType apis.ConditionType, status v1.ConditionStatus, reason, message string) duckv1.Conditions {
	return duckv1.Conditions{
		{
			Type:    conditionType,
			Status:  status,
			Reason:  reason,
			Message: message,
		},
	}
}
