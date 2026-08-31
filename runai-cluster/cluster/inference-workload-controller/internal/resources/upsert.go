/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/issue"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/status"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ResourceOptions struct {
	DryRun bool // if true, the resource manager will not create the resource
}

type K8sResourceManagerInterface interface {
	UpsertResources(ctx context.Context, workloadObjectMeta metav1.ObjectMeta, workloadSubResources []client.Object,
		ownedResources *[]common.ObservedResource, workloadConditions *[]common.WorkloadCondition, options ResourceOptions) ([]*issue.Issue, bool)
}

type K8sResourceManager struct {
	client client.Client
	rg     issue.ResourceGrouper
}

func NewResourceManager(client client.Client, rg issue.ResourceGrouper) *K8sResourceManager {
	return &K8sResourceManager{
		client: client,
		rg:     rg,
	}
}

func (krm *K8sResourceManager) configureK8sClientOptions(options ResourceOptions) (client.CreateOptions, client.UpdateOptions) {
	createOptions := client.CreateOptions{}
	updateOptions := client.UpdateOptions{}

	if options.DryRun {
		createOptions.DryRun = []string{"All"}
		updateOptions.DryRun = []string{"All"}
	}

	return createOptions, updateOptions
}
func (krm *K8sResourceManager) UpsertResources(ctx context.Context, workloadObjectMeta metav1.ObjectMeta, workloadSubResources []client.Object,
	ownedResources *[]common.ObservedResource, workloadConditions *[]common.WorkloadCondition, options ResourceOptions) ([]*issue.Issue, bool) {
	logger := log.FromContext(ctx)
	logger.V(4).Info("upserting workload sub resources")
	createOptions, updateOptions := krm.configureK8sClientOptions(options)
	var wg sync.WaitGroup
	upgradeResponseCh := make(chan string, len(workloadSubResources))
	issuesChan := make(chan *issue.Issue, len(workloadSubResources))

	for _, subResource := range workloadSubResources {
		// this newResource resolve mutual exclusion issue of pointer
		newResource := subResource
		wg.Add(1)
		go ObjectUpsert(ctx, krm.client, newResource, workloadObjectMeta, ownedResources, createOptions, updateOptions, upgradeResponseCh, issuesChan, &wg, krm.rg)
	}
	// update ready condition if some resources were upgraded
	// if no resources were upgraded, the ready condition will set to False
	// with info about the resources that are not uptodate and need to be upgraded
	wg.Wait()
	close(issuesChan)

	var issues []*issue.Issue
	// Collect and report errors
	for iss := range issuesChan {
		issues = append(issues, iss)
	}

	close(upgradeResponseCh)
	if len(issues) > 0 {
		isConflict := false
		for _, iss := range issues {
			if errors.IsConflict(iss.GetError()) {
				isConflict = true
				break
			}
		}
		// one of the goroutines failed
		logger.V(2).Info("upsert issue found for workload", "name", workloadObjectMeta.Name)
		return issues, isConflict
	}
	krm.checkUpgradedResource(ctx, upgradeResponseCh, workloadConditions)
	logger.V(4).Info("completed upserting workload sub resources", "workload", workloadObjectMeta.Name)

	return nil, false
}

func ObjectUpsert(ctx context.Context, client client.Client, newResource client.Object, workloadObjectMeta metav1.ObjectMeta, ownedResources *[]common.ObservedResource,
	createOptions client.CreateOptions, updateOptions client.UpdateOptions,
	upgradeResponseCh chan<- string, issueChan chan<- *issue.Issue, wg *sync.WaitGroup, rg issue.ResourceGrouper) {
	defer wg.Done()
	logger := log.FromContext(ctx)
	updateResource := true
	resourceType := resources.GetResourceType(newResource)
	if resourceType == constants.KindPvc || resourceType == constants.KindComputeDomain {
		updateResource = false
	}
	upgraded, iss := UpsertIfNeededCheckOwnerUID(ctx, client, newResource, workloadObjectMeta.UID, workloadObjectMeta.Name, createOptions, updateOptions, updateResource, rg)
	if iss != nil {
		logger.Info("failed to upsert resource", "resource", newResource, "resourceType", resourceType, "error", iss)
		issueChan <- iss
		return
	}
	if upgraded {
		upgradeResponseCh <- fmt.Sprintf("%s/%s", resourceType, newResource.GetName())
	}
	err := status.PatchOwnedResource(newResource, ownedResources)
	if err != nil {
		issueChan <- issue.New(newResource, err, false, "", rg)
	}
}

func (krm *K8sResourceManager) checkUpgradedResource(ctx context.Context, workloadUpgradedResourcesChannel chan string, workloadConditions *[]common.WorkloadCondition) {
	var message bytes.Buffer
	log := log.FromContext(ctx)
	log.V(4).Info("checking if some resources were upgraded")
	for upgradedResource := range workloadUpgradedResourcesChannel {
		message.WriteString(fmt.Sprintf("%s ", upgradedResource))
	}
	if message.Len() > 0 {
		log.V(4).Info("some resources were upgraded", "resources", message.String())
		krm.setReadyConditionFalse(workloadConditions, message.String())
		return
	}
	log.V(4).Info("no resources were upgraded")
}

func (krm *K8sResourceManager) setReadyConditionFalse(workloadConditions *[]common.WorkloadCondition, upgradedResources string) {
	if workloadConditions == nil {
		// TODO remove this once conditions are available in inference workloads
		return
	}

	readyCondition := status.GetWorkloadCondition(workloadConditions, common.Ready)
	var message bytes.Buffer
	if readyCondition != nil && readyCondition.Message != "" {
		message.WriteString(fmt.Sprintf("%s \n", readyCondition.Message))
	}
	message.WriteString("the following workload resources are stalled and need to be upgraded: ")
	message.WriteString(upgradedResources)
	status.UpsertCondition(workloadConditions, common.Ready, v1.ConditionFalse, "", message.String())
}
