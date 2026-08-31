/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"strings"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	nodepoolsSpecSeparator   = " "
	nodepoolNotFoundTemplate = "Nodepool '%s' does not exist, please create it first"
	failedToGetNodepools     = "Failed to get nodepools, please try again"
)

type ExistingNodepoolsValidator struct {
	k8sClient client.Client
}

func NewExistingNodepoolsValidator(k8sClient client.Client) *ExistingNodepoolsValidator {
	return &ExistingNodepoolsValidator{
		k8sClient: k8sClient,
	}
}

func (v ExistingNodepoolsValidator) Validate(ctx context.Context, workload *jobs.JobWorkload) error {
	log := logf.FromContext(ctx)

	nodepoolsFromSpec := getNodepoolNamesFromWorkloadSpec(workload)
	if len(nodepoolsFromSpec) == 0 {
		log.V(4).Info("got no nodepools to validate")
		return nil
	}

	log.Info("checking if expected nodepools exist", "workload", workload.Name)

	existingNodepoolSet, err := v.existingNodepoolNames(ctx)
	if err != nil {
		log.Error(err, "failed to list existing nodepools")
		return errors.New(failedToGetNodepools)
	}

	for _, nodepoolFromSpec := range nodepoolsFromSpec {
		if _, ok := existingNodepoolSet[nodepoolFromSpec]; !ok {
			log.Info("nodepool does not exist", "nodepool", nodepoolFromSpec)
			return fmt.Errorf(nodepoolNotFoundTemplate, nodepoolFromSpec)
		}
	}

	log.Info("all expected nodepools exist", "workload", workload.Name)
	return nil
}

// existingNodepoolNames returns the set of node pool names present in the cluster.
func (v ExistingNodepoolsValidator) existingNodepoolNames(ctx context.Context) (map[string]bool, error) {
	var existingNodepoolList kaiv1alpha1.NodePoolList
	if err := v.k8sClient.List(ctx, &existingNodepoolList); err != nil {
		return nil, err
	}

	existingNodepoolSet := make(map[string]bool)
	for _, np := range existingNodepoolList.Items {
		existingNodepoolSet[np.Name] = true
	}
	return existingNodepoolSet, nil
}

func getNodepoolNamesFromWorkloadSpec(workload *jobs.JobWorkload) []string {
	specNodepools := workload.Spec.NodePools.Deref()
	specNodepool := workload.Spec.NodePool.Deref()

	if len(specNodepools) == 0 {
		if len(specNodepool) > 0 {
			return []string{specNodepool}
		} else {
			return []string{}
		}
	}

	nodepoolList := strings.Split(specNodepools, nodepoolsSpecSeparator)

	nodepoolSet := make(map[string]bool)
	for _, np := range nodepoolList {
		nodepoolSet[np] = true
	}

	return maps.Keys(nodepoolSet)
}
