/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"

	workloadTypeConfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/workload-type-config"
	dinfw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics/data"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	workloadInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "workload_info",
			Help:      "Workload info",
		}, []string{
			workloadNameLabel, WorkloadIdLabel,
			workloadTypeLabel,
			workloadCategoryLabel,
			queueLabel, nodepoolLabel,
			projectLabel, departmentLabel,
			phaseLabel, detailedStatusLabel, statusLabel,
			userLabel,
			// internal labels so we can reset the metric on deletion -
			// object name can be different from workload name for external workloads.
			// at the time of deletion, we only have the object name, namespace, and source.
			workloadObjectNameLabel, workloadNamespaceLabel, workloadSourceLabel,
		},
	)
)

func SetExternalWorkloadMetrics(ctx context.Context, workload *v1alpha1.ExternalWorkload, client client.Client) error {
	workloadId := data.GetWorkloadId(workload.Spec.OriginalWorkloadUID, workload.Annotations)

	err := setWorkloadMetrics(ctx, workload.Spec.OriginalWorkloadName, workload.Namespace, workloadId,
		workload.Spec.Type, workload.Spec.Group, string(workload.Status.ObservedStatus.Phase),
		"",
		workload.Spec.RequestedNodePools,
		nil, ptr.To(workload.Spec.PriorityClassName),
		workload.Name, ExternalSource, workload.Labels, client)
	if err != nil {
		return err
	}

	log.Info().Msgf("Updated ExternalWorkload %s/%s metrics successfully - phase was: %s",
		workload.Namespace, workload.Name, string(workload.Status.ObservedStatus.Phase))
	return nil
}

func SetTrainingWorkloadMetrics(ctx context.Context, workload *twv2a1.TrainingWorkload, client client.Client) error {
	workloadId := data.GetWorkloadId(workload.UID, workload.Annotations)
	user := data.GetUserOfWorkload(&workload.Spec.WorkloadFields)

	err := setWorkloadMetrics(ctx, workload.Name, workload.Namespace, workloadId,
		workload.Kind, workload.GroupVersionKind().Group, workload.Status.Phase, user,
		utils.GetRequestedNodePoolsFromRunaiWorkload(workload.Spec.NodePools),
		nil, nil,
		workload.Name, TrainingSource, workload.Labels, client)
	if err != nil {
		return err
	}

	log.Info().Msgf("Updated TrainingWorkload %s/%s metrics successfully - phase was: %s",
		workload.Namespace, workload.Name, workload.Status.Phase)
	return nil
}

func SetInteractiveWorkloadMetrics(ctx context.Context, workload *iwv2a1.InteractiveWorkload, client client.Client) error {
	workloadId := data.GetWorkloadId(workload.UID, workload.Annotations)
	user := data.GetUserOfWorkload(&workload.Spec.WorkloadFields)

	err := setWorkloadMetrics(ctx, workload.Name, workload.Namespace, workloadId,
		workload.Kind, workload.GroupVersionKind().Group, workload.Status.Phase, user,
		utils.GetRequestedNodePoolsFromRunaiWorkload(workload.Spec.NodePools),
		ptr.To(workload.Spec.Preemptible.Deref()), nil,
		workload.Name, InteractiveSource, workload.Labels, client)
	if err != nil {
		return err
	}

	log.Info().Msgf("Updated InteractiveWorkload %s/%s metrics successfully - phase was: %s",
		workload.Namespace, workload.Name, workload.Status.Phase)
	return nil
}

func SetDistributedWorkloadMetrics(ctx context.Context, workload *diwv2.DistributedWorkload, client client.Client) error {
	workloadId := data.GetWorkloadId(workload.UID, workload.Annotations)
	user := data.GetUserOfWorkload(&workload.Spec.WorkloadFields)

	err := setWorkloadMetrics(ctx, workload.Name, workload.Namespace, workloadId,
		workload.Kind, workload.GroupVersionKind().Group, workload.Status.Phase, user,
		utils.GetRequestedNodePoolsFromRunaiWorkload(workload.Spec.NodePools),
		nil, nil,
		workload.Name, DistributedSource, workload.Labels, client)
	if err != nil {
		return err
	}

	log.Info().Msgf("Updated DistributedWorkload %s/%s metrics successfully - phase was: %s",
		workload.Namespace, workload.Name, workload.Status.Phase)
	return nil
}

func SetInferenceWorkloadMetrics(ctx context.Context, workload *infwv2a1.InferenceWorkload, client client.Client) error {
	workloadId := data.GetWorkloadId(workload.UID, workload.Annotations)
	user := data.GetUserOfWorkload(&workload.Spec.WorkloadFields)

	err := setWorkloadMetrics(ctx, workload.Name, workload.Namespace, workloadId,
		workload.Kind, workload.GroupVersionKind().Group, workload.Status.Phase, user,
		utils.GetRequestedNodePoolsFromRunaiWorkload(workload.Spec.NodePools),
		nil, nil,
		workload.Name, InferenceSource, workload.Labels, client)
	if err != nil {
		return err
	}

	log.Info().Msgf("Updated InferenceWorkload %s/%s metrics successfully - phase was: %s",
		workload.Namespace, workload.Name, workload.Status.Phase)
	return nil
}

func SetDistributedInferenceWorkloadMetrics(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, client client.Client) error {
	workloadId := data.GetWorkloadId(workload.UID, workload.Annotations)
	user := workload.Spec.UserName.Deref()

	err := setWorkloadMetrics(ctx, workload.Name, workload.Namespace, workloadId,
		workload.Kind, workload.GroupVersionKind().Group, workload.Status.Phase, user,
		utils.GetRequestedNodePoolsFromRunaiWorkload(workload.Spec.NodePools),
		nil, nil,
		workload.Name, DistributedInferenceSource, workload.Labels, client)
	if err != nil {
		return err
	}

	log.Info().Msgf("Updated DistributedInferenceSource %s/%s metrics successfully - phase was: %s",
		workload.Namespace, workload.Name, workload.Status.Phase)
	return nil
}

func setWorkloadMetrics(ctx context.Context, name, namespace, uuid, workloadKind, workloadGroup, workloadStatus, user string,
	requestedNodePools []string,
	isPreemptible *bool, priorityClassName *string,
	objectName string, source WorkloadSource, labels map[string]string,
	client client.Client) error {

	orgData, err := utils.GetOrgData(ctx, namespace, client)
	if err != nil {
		return err
	}
	projectName := orgData.ProjectName
	departmentName := orgData.DepartmentName

	var podsOfWorkload *[]corev1.Pod = nil
	nodePoolName := ""
	if len(requestedNodePools) == 1 {
		nodePoolName = requestedNodePools[0]
	} else {
		podsOfWorkloadInner, err := utils.GetPodsOfWorkload(ctx, uuid, namespace, name, client)
		if err != nil {
			return err
		}

		nodePoolName, err = data.GetNodePoolOfWorkload(ctx, podsOfWorkloadInner, client)
		if err != nil {
			return err
		}
		podsOfWorkload = &podsOfWorkloadInner
	}

	queueName := ""
	if nodePoolName != "" {
		queueName, err = utils.GetQueueName(ctx, nodePoolName, projectName, client)
		if err != nil {
			return err
		}
	}

	workloadTypeMetricLabelValue := data.GetWorkloadTypeMetricLabelValueBackwards(ctx,
		workloadKind, uuid, namespace, name,
		isPreemptible, priorityClassName,
		podsOfWorkload, client)

	workloadCategory, err := data.GetWorkloadCategory(ctx, client, workloadKind, workloadGroup, labels)
	if err != nil {
		log.Error().Msgf("Could not get Workload category: %v", err)
	}
	if validationErr := workloadTypeConfig.ValidateWorkloadCategory(ctx, client, workloadCategory); validationErr != nil {
		log.Error().Msgf("Invalid Workload category <%s>, error: %s", workloadCategory, validationErr.Error())
		workloadCategory = ""
	}

	phase := data.GetPhaseFromWorkloadStatus(workloadStatus)

	ResetWorkloadMetrics(objectName, namespace, source)

	workloadInfo.WithLabelValues(
		// workloadNameLabel, WorkloadIdLabel
		name, uuid,
		// workloadTypeLabel
		string(workloadTypeMetricLabelValue),
		// workloadCategoryLabel
		workloadCategory,
		// queueLabel, nodepoolLabel
		queueName, nodePoolName,
		// projectLabel, departmentLabel
		projectName, departmentName,
		// phaseLabel, detailedStatusLabel, statusLabel
		string(phase), workloadStatus, workloadStatus,
		// userLabel
		user,
		// workloadObjectNameLabel, workloadNamespaceLabel, workloadSourceLabel
		objectName, namespace, string(source)).Set(1)
	return nil
}

func ResetWorkloadMetrics(workloadObjectName, namespace string, source WorkloadSource) {
	workloadIdentifier := prometheus.Labels{
		workloadObjectNameLabel: workloadObjectName,
		workloadNamespaceLabel:  namespace,
		workloadSourceLabel:     string(source),
	}
	workloadInfo.DeletePartialMatch(workloadIdentifier)
}

func GetWorkloadInfoMetric() *prometheus.GaugeVec {
	return workloadInfo
}
