/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	multierror "github.com/hashicorp/go-multierror"

	schedv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type QueueResourceHandler struct {
	common.WithLoggerAndCli
}

func NewQueueResourceHandler(client client.Client) QueueResourceHandler {
	return QueueResourceHandler{
		WithLoggerAndCli: common.WithLoggerAndCli{
			Client: client,
			Log:    ctrl.Log.WithName("resource_handlers").WithName(common.LogQueueTag),
		},
	}
}

func (handler QueueResourceHandler) HandleResource(
	project kaiv1alpha1.Project,
) ([]kaiv1alpha1.ProjectCondition, error) {
	err := handler.handleResourceInner(context.Background(), project)
	return []kaiv1alpha1.ProjectCondition{{
		Type:    kaiv1alpha1.QueuesReady,
		Status:  GetStatusFromError(err),
		Reason:  getReasonFromError(err, QueuesHandlerFailed),
		Message: GetMessageFromError(err),
	}}, err
}

func (handler QueueResourceHandler) handleResourceInner(ctx context.Context, project kaiv1alpha1.Project) error {
	handler.Log.Info("Handling project", common.LogProjectTag, project.Name, common.LogQueueNumTag, len(project.Spec.Queues))

	var err error
	reconciledQueues := map[string]bool{}
	for _, projectQueue := range project.Spec.Queues {
		queueObject, innerErr := handler.buildQueueFromSpec(ctx, projectQueue, &project)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			continue
		}

		innerErr = handler.reconcileQueueObject(ctx, queueObject, project.Name, projectQueue.Nodepool)
		if innerErr != nil {
			handler.Log.Error(innerErr, "Failed reconciling queue object for project",
				common.LogQueueTag, queueObject.Name, common.LogProjectTag, project.Name)
			err = multierror.Append(err, innerErr)
			continue
		}

		reconciledQueues[queueObject.Name] = true
	}

	if err != nil {
		handler.Log.Info("Error occurred during reconcile of queues - skipping deletion of unnecessary queues",
			common.LogProjectTag, project.Name)
		return err
	}

	err = deleteUnnecessaryQueues(ctx, handler.Client, handler.Log, reconciledQueues,
		common.LogProjectTag, project.Name,
		config.Get().ProjectLabelKey, project.Name, common.ProjectKind)
	if err != nil {
		handler.Log.Error(err, "Failed deleting unnecessary queues for project",
			common.LogProjectTag, project.Name)
		return err
	}

	handler.Log.Info("Successfully reconciled queues for project", common.LogProjectTag, project.Name)
	return nil
}

func (handler QueueResourceHandler) reconcileQueueObject(ctx context.Context,
	queueObject *schedv2.Queue, projectName, nodepoolName string) error {
	existingQueue, err := handler.getExistingQueueOfProject(ctx,
		queueObject.Name, projectName, nodepoolName)
	if err != nil {
		if errors.IsNotFound(err) {
			queueObject.Name = handler.generateQueueNameOfProject(ctx, queueObject.Name, projectName)
			queueObject.Spec.DisplayName = queueObject.Name
			return createQueue(ctx, handler.Client, handler.Log,
				queueObject, common.LogProjectTag, projectName)
		}
		handler.Log.Error(err, "Error encountered while trying to get queue for project",
			common.LogQueueTag, queueObject.Name, common.LogProjectTag, projectName)
		return err
	}

	queueObject.Name = existingQueue.Name
	queueObject.Spec.DisplayName = existingQueue.Name
	return handler.handleExisting(ctx, queueObject, existingQueue, projectName)
}

func (handler QueueResourceHandler) handleExisting(ctx context.Context,
	expectedQueueObject, existingQueue *schedv2.Queue, projectName string) error {
	if areQueuesEqual(expectedQueueObject, existingQueue) {
		handler.Log.Info("Existing Queue identical to expected in project spec, skipping",
			common.LogQueueTag, existingQueue.Name, common.LogProjectTag, projectName)
		return nil
	}

	if common.IsManuallyOverridden(existingQueue) {
		handler.Log.Info("Queue already exists and is marked as manually overridden, it will not be updated",
			common.LogQueueTag, existingQueue.Name, common.LogProjectTag, projectName)
	}

	return updateExistingQueue(ctx, handler.Client, handler.Log,
		expectedQueueObject, existingQueue, common.LogProjectTag, projectName)
}

func (handler QueueResourceHandler) buildQueueFromSpec(ctx context.Context, queueProjectSpec kaiv1alpha1.QueueConfig,
	project *kaiv1alpha1.Project) (*schedv2.Queue, error) {
	parentQueueName, err := handler.parentQueueNameForProject(ctx, project, queueProjectSpec.Nodepool)
	if err != nil {
		return nil, err
	}
	suggestedName := GetQueueName(queueProjectSpec, project.Name)
	queue := &schedv2.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name: suggestedName,
			Labels: map[string]string{
				config.Get().ProjectLabelKey:             project.Name,
				config.Get().ProjectIdLabelKey:           string(project.UID),
				config.Get().QueueDepartmentNameLabelKey: project.Spec.Parent,
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         kaiv1alpha1.GroupVersion.Identifier(),
				Kind:               common.ProjectKind,
				Name:               project.Name,
				UID:                project.UID,
				Controller:         &common.TrueRef,
				BlockOwnerDeletion: &common.TrueRef,
			}},
		},
	}

	populateQueueSpec(queueProjectSpec.Resources, suggestedName, parentQueueName, queueProjectSpec.Priority, &queue.Spec)
	addNodePoolLabelIfNeeded(queue.Labels, config.Get().NodePoolLabelKey, queueProjectSpec.Nodepool)
	return queue, nil
}

// parentQueueNameForProject resolves the name of the parent (department) queue for a project's
// queue in the given nodepool. It fetches the project's parent Department (project.Spec.Parent
// is the department name), then:
//   - derives the queue name from the department spec (GetQueueName) and, if a queue with that
//     name exists and is owned by the department, uses it; otherwise
//   - falls back to finding the department-owned queue for the nodepool by OwnerReference (so a
//     collision-suffixed name is still honored).
//
// The parent is never located via the runai/department-* identity labels - only via the
// department object and OwnerReference. Returns "" when the project has no parent, the parent
// Department does not exist, or the department has no owned queue for the nodepool.
func (handler QueueResourceHandler) parentQueueNameForProject(ctx context.Context,
	project *kaiv1alpha1.Project, nodepoolName string) (string, error) {
	if project.Spec.Parent == "" {
		return "", nil
	}

	department := &kaiv1alpha1.Department{}
	if err := handler.Client.Get(ctx, client.ObjectKey{Name: project.Spec.Parent}, department); err != nil {
		if errors.IsNotFound(err) {
			handler.Log.Info("Parent department for project not found, continuing without a parent queue",
				common.LogProjectTag, project.Name, common.LogDepartmentTag, project.Spec.Parent, "nodepool", nodepoolName)
			return "", nil
		}
		handler.Log.Error(err, "Failed to get parent department for project",
			common.LogProjectTag, project.Name, common.LogDepartmentTag, project.Spec.Parent)
		return "", err
	}

	// First: take the name from the department spec and validate the queue exists and is owned
	// by the department.
	candidateName := ""
	for i := range department.Spec.Queues {
		if department.Spec.Queues[i].Nodepool != nodepoolName {
			continue
		}
		candidateName = GetQueueName(department.Spec.Queues[i], department.Name)
		candidate := &schedv2.Queue{}
		err := handler.Client.Get(ctx, client.ObjectKey{Name: candidateName}, candidate)
		if err == nil && isOwnedByUID(candidate.OwnerReferences, department.UID) {
			return candidateName, nil
		}
		if err != nil && !errors.IsNotFound(err) {
			return "", err
		}
		break // spec entry found but not a valid owned queue - fall back to the owner search
	}
	if candidateName == "" {
		// no queue for this nodepool
		return "", nil
	}
	// Fallback: find the queue owned by the department for this nodepool.
	return handler.parentQueueByOwner(ctx, department, nodepoolName)
}

// parentQueueByOwner returns the name of the queue owned by the given department for the
// nodepool, located by OwnerReference (not by the department identity labels).
func (handler QueueResourceHandler) parentQueueByOwner(ctx context.Context,
	department *kaiv1alpha1.Department, nodepoolName string) (string, error) {
	queues, err := listQueuesForNodepool(ctx, handler.Client, nodepoolName)
	if err != nil {
		handler.Log.Error(err, "Failed listing queues to find parent queue by owner",
			common.LogDepartmentTag, department.Name, "nodepool", nodepoolName)
		return "", err
	}
	for i := range queues {
		if isOwnedByUID(queues[i].OwnerReferences, department.UID) {
			return queues[i].Name, nil
		}
	}
	return "", nil
}

// isOwnedByUID reports whether any of the owner references points to the given owner UID.
func isOwnedByUID(ownerRefs []metav1.OwnerReference, ownerUID types.UID) bool {
	for _, ref := range ownerRefs {
		if ref.UID == ownerUID {
			return true
		}
	}
	return false
}

// isOwnedByKind reports whether any of the owner references is of the given Kind. Used to tell a
// department-owned queue (Kind=Department) apart from a project-owned queue (Kind=Project), since
// both can carry the same department-name label.
func isOwnedByKind(ownerRefs []metav1.OwnerReference, kind string) bool {
	for _, ref := range ownerRefs {
		if ref.Kind == kind {
			return true
		}
	}
	return false
}

func (handler QueueResourceHandler) getExistingQueueOfProject(ctx context.Context,
	queueNameFomSpec, projectName, nodepoolName string) (*schedv2.Queue, error) {
	return getExistingQueueOfResource(ctx, handler.Client, handler.Log, nodepoolName, projectName,
		"", queueNameFomSpec)
}

func (handler QueueResourceHandler) generateQueueNameOfProject(ctx context.Context,
	queueNameFomSpec, projectName string) string {
	return generateQueueNameForResource(ctx, handler.Client, projectName,
		"", queueNameFomSpec)
}
