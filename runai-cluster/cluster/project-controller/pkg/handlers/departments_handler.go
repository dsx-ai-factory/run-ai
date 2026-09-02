/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	schedv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The DepartmentHandler acquires the default limit ranges Config Map (under our 'runai' Namespace)
// and applies it as a LimitRange resource under the Project's Namespace
type DepartmentHandler struct {
	common.WithLoggerAndCli
}

func NewDepartmentHandler(client client.Client) DepartmentHandler {
	return DepartmentHandler{
		WithLoggerAndCli: common.WithLoggerAndCli{
			Client: client,
			Log:    ctrl.Log.WithName("handlers").WithName(common.LogDepartmentTag),
		},
	}
}

func (h *DepartmentHandler) Handle(ctx context.Context, department *kaiv1alpha1.Department) error {
	var err error
	reconciledQueues := map[string]bool{}
	for _, departmentQueue := range department.Spec.Queues {
		queueObject := h.buildQueueFromSpec(departmentQueue, department)
		innerErr := h.reconcileQueueObject(ctx, queueObject, department, departmentQueue.Nodepool)
		if innerErr != nil {
			h.Log.Error(innerErr, "Failed reconciling queue object for department",
				common.LogQueueTag, queueObject.Name, common.LogDepartmentTag, department.Name)
			err = multierror.Append(err, innerErr)
			continue
		}

		reconciledQueues[queueObject.Name] = true
	}

	if err != nil {
		h.Log.Info("Error occurred during reconcile of queues - skipping deletion of unnecessary queues",
			common.LogDepartmentTag, department.Name)
		return err
	}

	err = deleteUnnecessaryQueues(ctx, h.Client, h.Log, reconciledQueues,
		common.LogDepartmentTag, department.Name,
		config.Get().QueueDepartmentNameLabelKey, department.Name, common.DepartmentKind)
	if err != nil {
		h.Log.Error(err, "Failed deleting unnecessary queues for department",
			common.LogDepartmentTag, department.Name)
		return err
	}

	h.Log.Info("Successfully reconciled queues for department", common.LogDepartmentTag, department.Name)
	return nil
}

func (h *DepartmentHandler) reconcileQueueObject(ctx context.Context,
	queueObject *schedv2.Queue, department *kaiv1alpha1.Department, nodepoolName string) error {
	departmentName := department.Name
	existingQueue, err := h.getExistingQueueOfDepartment(ctx, department, nodepoolName, queueObject.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			queueObject.Name = h.generateQueueNameOfDepartment(ctx, departmentName, queueObject.Name)
			return createQueue(ctx, h.Client, h.Log,
				queueObject, common.LogDepartmentTag, departmentName)
		}
		h.Log.Error(err, "Error encountered while trying to get queue for department",
			common.LogQueueTag, queueObject.Name, common.LogDepartmentTag, departmentName)
		return err
	}

	queueObject.Name = existingQueue.Name
	return h.handleExisting(ctx, queueObject, existingQueue, departmentName)
}

func (h *DepartmentHandler) handleExisting(ctx context.Context,
	expectedQueueObject, existingQueue *schedv2.Queue, departmentName string) error {
	if areQueuesEqual(expectedQueueObject, existingQueue) {
		h.Log.Info("Existing Queue identical to expected in department spec, skipping",
			common.LogQueueTag, existingQueue.Name, common.LogDepartmentTag, departmentName)
		return nil
	}

	return updateExistingQueue(ctx, h.Client, h.Log,
		expectedQueueObject, existingQueue, common.LogDepartmentTag, departmentName)
}

// buildQueueFromSpec - the returned queue spec will have the suggested queue name.
// this queue name might already be taken by other resources, the caller needs to validate it.
func (h *DepartmentHandler) buildQueueFromSpec(queueDepartmentSpec kaiv1alpha1.QueueConfig,
	department *kaiv1alpha1.Department) *schedv2.Queue {
	queue := &schedv2.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name: GetQueueName(queueDepartmentSpec, department.Name),
			Labels: map[string]string{
				config.Get().QueueDepartmentNameLabelKey: department.Name,
			},
			OwnerReferences: []metav1.OwnerReference{{
				// Set from constants (not department.TypeMeta, which is often empty at runtime)
				// so the owner Kind reliably identifies this as a department-owned queue.
				APIVersion:         kaiv1alpha1.GroupVersion.Identifier(),
				Kind:               common.DepartmentKind,
				Name:               department.Name,
				UID:                department.UID,
				Controller:         &common.TrueRef,
				BlockOwnerDeletion: &common.TrueRef,
			}},
		},
	}
	populateQueueSpec(queueDepartmentSpec.Resources, department.Name, "", queueDepartmentSpec.Priority, &queue.Spec)
	addNodePoolLabelIfNeeded(queue.Labels, config.Get().NodePoolLabelKey, queueDepartmentSpec.Nodepool)
	return queue
}

// getExistingQueueOfDepartment finds the department's own queue for the nodepool by OwnerReference
// (Department UID). Using the owner - not the department-name label - makes this rename-safe
// (the UID is stable) and immune to project queues that share the department-name label.
func (h *DepartmentHandler) getExistingQueueOfDepartment(ctx context.Context,
	department *kaiv1alpha1.Department, nodepoolName, suggestedQueueName string) (*schedv2.Queue, error) {
	queues, err := listQueuesForNodepool(ctx, h.Client, nodepoolName)
	if err != nil {
		h.Log.Error(err, "Failed listing queues for department", common.LogDepartmentTag, department.Name,
			"nodepool", nodepoolName)
		return nil, err
	}
	for i := range queues {
		if isOwnedByUID(queues[i].OwnerReferences, department.UID) {
			return &queues[i], nil
		}
	}
	return nil, errors.NewNotFound(schedv2.Resource("queue"),
		fmt.Sprintf("%s/%s", department.Name, nodepoolName))
}

func (h *DepartmentHandler) generateQueueNameOfDepartment(ctx context.Context,
	departmentName, suggestedQueueName string) string {
	return generateQueueNameForResource(ctx, h.Client, "",
		departmentName, suggestedQueueName)
}
