/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/utils/ptr"

	"github.com/go-logr/logr"
	multierror "github.com/hashicorp/go-multierror"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	schedv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	queueNameRandSuffixLen     = 4
	maxNameLen                 = 58 // with buffer for "-" and suffix for max 63 chars
	queueNameRandSuffixLongLen = 20
	maxNameLenWithLongSuffix   = 42 // 63 - 20 -1
	generateQueueNameTries     = 10
	DefaultQueuePriority       = 100
)

// GetQueueName derives the suggested queue name from a queue spec, shared by the project and
// department handlers: the spec's Name is used as-is when set; otherwise it falls back to
// "<ownerName>-<nodepool>" (or just "<ownerName>" for the default nodepool). The result is a
// suggestion - the caller still validates it against queues taken by other resources and may
// append a random suffix (see generateQueueNameForResource).
func GetQueueName(queue kaiv1alpha1.QueueConfig, ownerName string) string {
	if queue.Name != "" {
		return queue.Name
	}
	if queue.Nodepool == config.Get().DefaultNodepoolName {
		return ownerName
	}
	return fmt.Sprintf("%s-%s", ownerName, queue.Nodepool)
}

func populateQueueSpec(resources *kaiv1alpha1.QueueResourcesConfig, displayName, parentQueue string, priority *int32, outputQueueSpec *schedv2.QueueSpec) {
	priorityValue := ptr.To(DefaultQueuePriority)
	if priority != nil {
		priorityValue = ptr.To(int(*priority))
	}
	*outputQueueSpec = schedv2.QueueSpec{
		DisplayName: displayName,
		ParentQueue: parentQueue,
		Resources: &schedv2.QueueResources{
			GPU: schedv2.QueueResource{
				Quota:           resources.GPU.Deserved,
				OverQuotaWeight: resources.GPU.OverQuotaWeight,
				Limit:           resources.GPU.Limit,
			},
			CPU: schedv2.QueueResource{
				Quota:           resources.CPU.Deserved,
				OverQuotaWeight: resources.CPU.OverQuotaWeight,
				Limit:           resources.CPU.Limit,
			},
			Memory: schedv2.QueueResource{
				Quota:           resources.Memory.Deserved,
				OverQuotaWeight: resources.Memory.OverQuotaWeight,
				Limit:           resources.Memory.Limit,
			},
		},
		Priority: priorityValue,
	}
}

func addNodePoolLabelIfNeeded(labels map[string]string, nodepoolLabelKey, nodepoolName string) {
	if nodepoolName != config.Get().DefaultNodepoolName {
		labels[nodepoolLabelKey] = nodepoolName
	}
}

func areQueuesEqual(left, right *schedv2.Queue) bool {
	return reflect.DeepEqual(left.Name, right.Name) &&
		reflect.DeepEqual(left.Labels, right.Labels) &&
		reflect.DeepEqual(left.OwnerReferences, right.OwnerReferences) &&
		reflect.DeepEqual(left.Spec, right.Spec)
}

func updateExistingQueueWithExpectedValues(expectedQueueObject, existingQueue *schedv2.Queue) {
	existingQueue.Labels = expectedQueueObject.Labels
	existingQueue.OwnerReferences = expectedQueueObject.OwnerReferences

	expectedQueueObject.Spec.DeepCopyInto(&existingQueue.Spec)
}

func createQueue(ctx context.Context, k8sClient client.Client, log logr.Logger,
	queueObject *schedv2.Queue, objectType, objectName string) error {
	log.Info(fmt.Sprintf("Creating queue for %s", objectType),
		common.LogQueueTag, queueObject.Name, objectType, objectName)

	err := k8sClient.Create(ctx, queueObject)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Error(err, fmt.Sprintf("Failed creating queue for %s", objectType),
			common.LogQueueTag, queueObject.Name, objectType, objectName)
		return err
	}

	return nil
}

func updateExistingQueue(ctx context.Context, k8sClient client.Client, log logr.Logger,
	expectedQueueObject, existingQueue *schedv2.Queue, objectType, objectName string) error {
	log.Info(fmt.Sprintf("Existing Queue differs from expected in %s spec, updating. Resources changes: %s",
		objectType, printQueueResourcesChanges(existingQueue, expectedQueueObject)),
		common.LogQueueTag, existingQueue.Name, objectType, objectName)

	updateExistingQueueWithExpectedValues(expectedQueueObject, existingQueue)
	err := k8sClient.Update(ctx, existingQueue)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed updating queue for %s", objectType),
			common.LogQueueTag, expectedQueueObject.Name, objectType, objectName)
		return err
	}
	return nil
}

func deleteQueue(ctx context.Context, k8sClient client.Client, log logr.Logger,
	queue schedv2.Queue, objectType, objectName string) error {
	log.Info(fmt.Sprintf("Deleting queue for %s as it no longer exists in spec", objectType),
		common.LogQueueTag, queue.Name, objectType, objectName)
	err := k8sClient.Delete(ctx, &queue)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed deleting queue for %s", objectType),
			common.LogQueueTag, queue.Name, objectType, objectName)
		return err
	}
	return nil
}

func ListProjectQueues(ctx context.Context, k8sClient client.Client, projectName string) ([]schedv2.Queue, error) {
	return listQueuesWithLabelSelector(ctx, k8sClient, config.Get().ProjectLabelKey, projectName)
}

func listQueuesWithLabelSelector(ctx context.Context, k8sClient client.Client,
	labelKey, labelValue string) ([]schedv2.Queue, error) {
	return listQueuesWithLabelSelectors(ctx, k8sClient, client.MatchingLabels(map[string]string{labelKey: labelValue}))
}

// listQueuesForNodepool lists queues for the given nodepool without any owner/project/department
// identity filter. It is used to locate a queue by its OwnerReference rather than by the
// runai/department-* identity labels.
func listQueuesForNodepool(ctx context.Context, k8sClient client.Client, nodepoolName string) ([]schedv2.Queue, error) {
	var nodePoolRequirement *labels.Requirement
	var err error
	if nodepoolName == config.Get().DefaultNodepoolName {
		nodePoolRequirement, err = labels.NewRequirement(
			config.Get().NodePoolLabelKey, selection.DoesNotExist, []string{})
	} else {
		nodePoolRequirement, err = labels.NewRequirement(
			config.Get().NodePoolLabelKey, selection.DoubleEquals, []string{nodepoolName})
	}
	if err != nil {
		return nil, err
	}
	selector := labels.NewSelector().Add(*nodePoolRequirement)
	return listQueuesWithLabelSelectors(ctx, k8sClient, &client.ListOptions{LabelSelector: selector})
}

func listQueuesWithLabelSelectors(ctx context.Context, k8sClient client.Client,
	labelSelector client.ListOption) ([]schedv2.Queue, error) {
	queues := &schedv2.QueueList{}
	err := k8sClient.List(ctx, queues, labelSelector)
	if err != nil {
		return []schedv2.Queue{}, err
	}
	return queues.Items, nil
}

// deleteUnnecessaryQueues - in case a nodepool was deleted, we need to delete its queue.
// ownerKind scopes the deletion to queues owned by the reconciling resource's kind (Project /
// Department): the department-name label is shared by a department's queues AND its
// projects' queues, so listing by that label alone is not enough - only queues owned by
// ownerKind are candidates for deletion, so a department reconcile never deletes its projects'
// queues.
func deleteUnnecessaryQueues(ctx context.Context, k8sClient client.Client, log logr.Logger,
	reconciledQueues map[string]bool,
	objectType, objectName string,
	labelKey, labelValue, ownerKind string) error {
	existingQueues, err := listQueuesWithLabelSelector(ctx, k8sClient, labelKey, labelValue)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed listing all %s queues", objectType),
			objectType, objectName)
		return err
	}

	for _, existingQueue := range existingQueues {
		if !isOwnedByKind(existingQueue.OwnerReferences, ownerKind) {
			// not owned by the reconciling resource kind (e.g. a project queue during a
			// department reconcile) - never delete it here.
			continue
		}
		if _, found := reconciledQueues[existingQueue.Name]; found {
			continue
		}

		innerErr := deleteQueue(ctx, k8sClient, log, existingQueue, objectType, objectName)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}
	}

	return err
}

// isQueueTakenByOtherResource - validates that the queue
// is not taken by other resource (other department's queue name or even other project's queue name).
// returns bool - whether the queue is taken by other resource or not.
// and if the queue exists - returns the queue object.
func isQueueTakenByOtherResource(ctx context.Context, k8sClient client.Client,
	expectedDepartmentName, expectedProjectName, queueName string) (bool, *schedv2.Queue) {
	existingQueue := &schedv2.Queue{}
	err := k8sClient.Get(ctx,
		client.ObjectKey{Name: queueName},
		existingQueue,
	)
	if err != nil {
		// queue doesn't exist - the suggested name is good, not taken by other resource
		return false, nil
	}

	// A queue "belongs" to a department only if it is owned by a Department (Kind), and to a
	// project only if owned by a Project. The department-name label alone is ambiguous
	// (project queues carry their parent department's name too), so ownership Kind is authoritative.
	ownedByDepartment := isOwnedByKind(existingQueue.OwnerReferences, common.DepartmentKind)
	ownedByProject := isOwnedByKind(existingQueue.OwnerReferences, common.ProjectKind)

	departmentNameOnQueue := existingQueue.Labels[config.Get().QueueDepartmentNameLabelKey]
	projectNameOnQueue := existingQueue.Labels[config.Get().ProjectLabelKey]

	if expectedDepartmentName != "" && ownedByDepartment && expectedDepartmentName == departmentNameOnQueue {
		// already owned by the expected department
		return false, existingQueue
	}

	if expectedProjectName != "" && ownedByProject && expectedProjectName == projectNameOnQueue {
		// already owned by the expected project
		return false, existingQueue
	}

	if ownedByDepartment || ownedByProject {
		// owned by some other department/project - taken by another resource
		return true, existingQueue
	}

	// the queue exists but doesn't have any resource claiming it
	return false, existingQueue
}

func getExistingQueueByLabels(ctx context.Context, k8sClient client.Client, log logr.Logger,
	nodepoolName, projectName, departmentName string) (*schedv2.Queue, error) {
	requirements, err := createRequirementsForQueuesList(log, nodepoolName, projectName, departmentName)
	if err != nil {
		return nil, err
	}

	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(requirements...)

	queues, err := listQueuesWithLabelSelectors(ctx, k8sClient, &client.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		log.Error(err, "Failed listing queues with selector",
			common.LogProjectTag, projectName, common.LogDepartmentTag, departmentName)
		return nil, err
	}

	// The department-name label is shared by a department's queues and its projects'
	// queues, so the selector can return both. Return the queue owned by the kind we are looking
	// up (Project when projectName is set, otherwise Department).
	expectedOwnerKind := common.DepartmentKind
	if projectName != "" {
		expectedOwnerKind = common.ProjectKind
	}
	for i := range queues {
		if isOwnedByKind(queues[i].OwnerReferences, expectedOwnerKind) {
			return &queues[i], nil
		}
	}

	notFoundQueueName := ""
	if projectName != "" {
		notFoundQueueName = fmt.Sprintf("%s/%s", projectName, nodepoolName)
	} else {
		notFoundQueueName = fmt.Sprintf("%s/%s", departmentName, nodepoolName)
	}

	return nil, errors.NewNotFound(schedv2.Resource("queue"), notFoundQueueName)
}

func createRequirementsForQueuesList(log logr.Logger,
	nodepoolName, projectName, departmentName string,
) ([]labels.Requirement, error) {
	if projectName == "" && departmentName == "" {
		return []labels.Requirement{},
			fmt.Errorf(
				"wrong params for createRequirementsForQueuesList - project %s, department %s",
				projectName, departmentName)
	}

	result := []labels.Requirement{}

	if projectName != "" {
		err := addRequirement(config.Get().ProjectLabelKey, projectName, &result, log)
		if err != nil {
			return []labels.Requirement{}, err
		}
	} else {
		err := addRequirement(config.Get().QueueDepartmentNameLabelKey, departmentName, &result, log)
		if err != nil {
			return []labels.Requirement{}, err
		}
	}

	var nodePoolRequirement *labels.Requirement
	var err error
	if nodepoolName == config.Get().DefaultNodepoolName {
		nodePoolRequirement, err = labels.NewRequirement(
			config.Get().NodePoolLabelKey, selection.DoesNotExist, []string{})
	} else {
		nodePoolRequirement, err = labels.NewRequirement(
			config.Get().NodePoolLabelKey, selection.DoubleEquals, []string{nodepoolName})
	}
	if err != nil {
		log.Error(err, "Failed creating requirement for nodepool name label selector",
			common.LogDepartmentTag, departmentName, "nodepool", nodepoolName)
		return []labels.Requirement{}, err
	}

	result = append(result, *nodePoolRequirement)
	return result, nil
}

func getExistingQueueOfResource(ctx context.Context, k8sClient client.Client, log logr.Logger,
	nodepoolName, projectName, departmentName, suggestedQueueName string) (*schedv2.Queue, error) {
	existingQueueByLabels, err := getExistingQueueByLabels(ctx, k8sClient, log,
		nodepoolName, projectName, departmentName)
	if err == nil {
		return existingQueueByLabels, nil
	}

	if !errors.IsNotFound(err) {
		return nil, err
	}

	// if a previously created queue with random suffix had its labels removed... well...
	// there's nothing we can do about it.
	// we can only try to get the one with the expected name (without random suffix)
	// and see if it doesn't belong to another resource.

	isTaken, existingQueueWithExpectedName := isQueueTakenByOtherResource(ctx, k8sClient,
		departmentName, projectName, suggestedQueueName)
	if !isTaken && existingQueueWithExpectedName != nil {
		return existingQueueWithExpectedName, nil
	}

	return nil, errors.NewNotFound(schedv2.Resource("queue"), suggestedQueueName)
}

func generateQueueNameForResource(ctx context.Context, k8sClient client.Client,
	projectName, departmentName, suggestedQueueName string) string {
	newSuggestedQueueName := suggestedQueueName

	// limit the queue name to 63 characters, as per Kubernetes naming and label values conventions
	trimNameLen := len(suggestedQueueName)
	if len(suggestedQueueName) > maxNameLen {
		trimNameLen = maxNameLen
		newSuggestedQueueName = fmt.Sprintf("%s-%s",
			suggestedQueueName[:maxNameLen], rand.String(queueNameRandSuffixLen))
	}

	for i := 0; i < generateQueueNameTries; i++ {
		if !isSuggestedQueueTakenByOtherResource(ctx, k8sClient, projectName, departmentName, newSuggestedQueueName) {
			return newSuggestedQueueName
		}

		newSuggestedQueueName = fmt.Sprintf("%s-%s",
			suggestedQueueName[:trimNameLen], rand.String(queueNameRandSuffixLen))
	}

	// the chances of getting here are very very very slim...
	if trimNameLen > maxNameLenWithLongSuffix {
		trimNameLen = maxNameLenWithLongSuffix
	}
	return fmt.Sprintf("%s-%s",
		suggestedQueueName[:trimNameLen], rand.String(queueNameRandSuffixLongLen))
}

func isSuggestedQueueTakenByOtherResource(ctx context.Context, k8sClient client.Client,
	expectedDepartmentName, expectedProjectName, suggestedQueueName string) bool {
	isTaken, _ := isQueueTakenByOtherResource(ctx, k8sClient,
		expectedDepartmentName, expectedProjectName, suggestedQueueName)
	return isTaken
}

func addRequirement(labelKey, labelValue string, result *[]labels.Requirement, log logr.Logger) error {
	if labelValue == "" {
		return nil
	}

	departmentIdRequirement, err := labels.NewRequirement(
		labelKey, selection.DoubleEquals, []string{labelValue})
	if err != nil {
		log.Error(err, "Failed creating requirement for label selector",
			"labelKey", labelKey, "labelValue", labelValue)
		return err
	}

	*result = append(*result, *departmentIdRequirement)
	return nil
}

func printQueueResourcesChanges(existing, expected *schedv2.Queue) string {
	if existing == nil || existing.Spec.Resources == nil {
		if expected == nil || expected.Spec.Resources == nil {
			return "both nil"
		}
		return fmt.Sprintf("existing: nil, new: %+v", expected.Spec.Resources)
	} else if expected == nil || expected.Spec.Resources == nil {
		return fmt.Sprintf("existing: %+v, new: nil", existing.Spec.Resources)
	}

	e := existing.Spec.Resources
	n := expected.Spec.Resources

	return fmt.Sprintf(
		"GPU: quota: %.2f -> %.2f; overQuotaWeight: %.2f -> %.2f; limit: %.2f -> %.2f; "+
			"CPU: quota: %.2f -> %.2f; overQuotaWeight: %.2f -> %.2f; limit: %.2f -> %.2f; "+
			"Memory: quota: %.2f -> %.2f; overQuotaWeight: %.2f -> %.2f; limit: %.2f -> %.2f",
		e.GPU.Quota, n.GPU.Quota, e.GPU.OverQuotaWeight, n.GPU.OverQuotaWeight, e.GPU.Limit, n.GPU.Limit,
		e.CPU.Quota, n.CPU.Quota, e.CPU.OverQuotaWeight, n.CPU.OverQuotaWeight, e.CPU.Limit, n.CPU.Limit,
		e.Memory.Quota, n.Memory.Quota, e.Memory.OverQuotaWeight, n.Memory.OverQuotaWeight, e.Memory.Limit, n.Memory.Limit)
}
