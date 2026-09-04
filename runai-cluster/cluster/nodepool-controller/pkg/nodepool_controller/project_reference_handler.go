/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
)

const maxProjectNamesInMessage = 10

// getProjectsReferencingNodePool returns the names of the projects that still hold a queue
// (per-nodepool resources) for the given nodepool. Only spec references block deletion;
func (npc *NodePoolController) getProjectsReferencingNodePool(ctx context.Context, nodePoolName string) ([]string, error) {
	projects := &v1alpha1.ProjectList{}
	if err := npc.Client.List(ctx, projects); err != nil {
		log.Error().Msgf("Failed listing kai projects while checking references to nodepool <%v>, err: %v",
			nodePoolName, err.Error())
		return nil, err
	}

	referencingProjects := []string{}
	for i := range projects.Items {
		project := &projects.Items[i]
		for _, queue := range project.Spec.Queues {
			if queue.Nodepool == nodePoolName {
				referencingProjects = append(referencingProjects, project.Name)
				break
			}
		}
	}

	sort.Strings(referencingProjects)
	return referencingProjects, nil
}

// addOrUpdateProjectsReferenceCondition - in case there are referencing projects - adds new or updates existing condition.
// In case there are none - updates an existing condition to False; does not add a new False condition.
func (npc *NodePoolController) addOrUpdateProjectsReferenceCondition(nodePool *v1alpha1.NodePool, referencingProjects []string) {
	hasReferences := len(referencingProjects) > 0
	condition := v1alpha1.NodePoolCondition{
		Reason: v1alpha1.ProjectReferencesExistReason,
		Type:   v1alpha1.ProjectReferencesExist,
	}
	if hasReferences {
		condition.Message = getProjectsReferenceMessage(referencingProjects)
	}
	condition.SetConditionStatusValue(hasReferences)
	nodePool.SetNodePoolCondition(condition)
}

func getProjectsReferenceMessage(referencingProjects []string) string {
	if len(referencingProjects) == 0 {
		return ""
	}

	projectNames := referencingProjects
	suffix := ""
	if len(projectNames) > maxProjectNamesInMessage {
		suffix = fmt.Sprintf(" (and %d more)", len(projectNames)-maxProjectNamesInMessage)
		projectNames = projectNames[:maxProjectNamesInMessage]
	}
	return fmt.Sprintf(common.ProjectsReferencingNodePoolMessage, strings.Join(projectNames, ", ")+suffix)
}
