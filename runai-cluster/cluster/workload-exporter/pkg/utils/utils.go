/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/getters"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NodePoolsSeparator = " "
)

type OrgData struct {
	ProjectName    string
	DepartmentName string
}

func GetOrgData(ctx context.Context, namespaceName string, readerClient client.Reader) (*OrgData, error) {
	project, err := GetProjectFromNamespaceName(ctx, namespaceName, readerClient)
	if err != nil {
		return nil, err
	}
	return &OrgData{
		ProjectName:    project.Name,
		DepartmentName: project.Spec.Parent,
	}, nil
}

func GetProjectNameFromNamespaceName(ctx context.Context, namespaceName string, readerClient client.Reader) (string, error) {
	namespace, err := getNamespace(ctx, namespaceName, readerClient)
	if err != nil {
		return "", err
	}

	projectName := GetProjectNameFromNamespace(namespace)
	if projectName == "" {
		return "", fmt.Errorf("namespace <%s> doesn't include project label", namespaceName)
	}

	return projectName, nil
}

func GetProjectFromNamespaceName(ctx context.Context, namespaceName string, readerClient client.Reader) (*kaiorg.Project, error) {
	projectName, err := GetProjectNameFromNamespaceName(ctx, namespaceName, readerClient)
	if err != nil {
		return nil, err
	}

	return getKaiProject(ctx, projectName, readerClient)
}

func GetProjectNameFromNamespace(namespace *corev1.Namespace) string {
	project, found := namespace.Labels[constants.RunaiQueueLabel]
	if !found {
		return ""
	}
	return project
}

func GetQueueName(ctx context.Context, nodePoolName string, projectName string, readerClient client.Reader) (
	string, error) {
	queue, err := getQueue(ctx, nodePoolName, projectName, readerClient)
	if err != nil {
		return "", err
	}

	return queue.Name, nil
}

func GetPodsOfWorkload(ctx context.Context, workloadId, workloadNamespace, workloadName string, readerClient client.Reader) (
	[]corev1.Pod, error) {
	pods, err := getters.GetPodsOfWorkload(ctx, workloadId, workloadNamespace, readerClient)
	if err != nil {
		log.Error().Msgf("Failed listing pods of workload <%s/%s> (id <%s>), err: %s",
			workloadNamespace, workloadName, workloadId, err.Error())
		return nil, err
	}
	return pods, nil
}

func GetPodGroupsOfWorkloadFromPods(ctx context.Context, workloadNamespace string, podsOfWorkload []corev1.Pod,
	readerClient client.Reader) ([]*kaiv2alpha2.PodGroup, error) {
	return getters.GetPodGroupsOfPods(ctx, workloadNamespace, podsOfWorkload, readerClient)
}

func GetPodGroupsOfWorkload(ctx context.Context,
	workloadId, workloadNamespace string,
	podsOfWorkload *[]corev1.Pod,
	readerClient client.Reader) ([]*kaiv2alpha2.PodGroup, error) {
	var podGroups []*kaiv2alpha2.PodGroup
	var err error

	if podsOfWorkload != nil {
		podGroups, err = GetPodGroupsOfWorkloadFromPods(ctx, workloadNamespace, *podsOfWorkload, readerClient)
	} else {
		podGroups, err = getters.GetPodGroupsOfWorkload(ctx, workloadId, workloadNamespace, readerClient)
	}

	return podGroups, err
}

func GetRequestedNodePoolsFromRunaiWorkload(requestedNodePoolsFld *flds.StringField) []string {
	if requestedNodePoolsFld == nil {
		return []string{}
	}

	requestedNodePoolsStr := requestedNodePoolsFld.Deref()
	if requestedNodePoolsStr == "" {
		return []string{}
	}

	return strings.Split(requestedNodePoolsStr, NodePoolsSeparator)
}
