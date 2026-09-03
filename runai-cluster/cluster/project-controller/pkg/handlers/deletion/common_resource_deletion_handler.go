/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package deletion

import (
	"context"
	"fmt"
	"sort"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	multierror "github.com/hashicorp/go-multierror"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectResourceDeletionHandler interface {
	OnDelete(project *kaiv1alpha1.Project) ([]kaiv1alpha1.ProjectCondition, error)
}

type CommonResourceDeletionHandler struct {
	common.WithLoggerAndCli
}

type ProjectConditionDeletionReason string

const (
	DataVolumesDeletionHandlerFailed ProjectConditionDeletionReason = "DataVolumesDeletionHandlerFailed"
	NamespaceDeletionHandlerFailed   ProjectConditionDeletionReason = "NamespaceDeletionHandlerFailed"
	WorkloadsDeletionHandlerFailed   ProjectConditionDeletionReason = "WorkloadsDeletionHandlerFailed"
	LimitRangeDeletionHandlerFailed  ProjectConditionDeletionReason = "LimitRangeDeletionHandlerFailed"
	QueuesDeletionHandlerFailed      ProjectConditionDeletionReason = "QueuesDeletionHandlerFailed"
	SecretsDeletionHandlerFailed     ProjectConditionDeletionReason = "SecretsDeletionHandlerFailed"
	PvcsDeletionHandlerFailed        ProjectConditionDeletionReason = "PvcsDeletionHandlerFailed"
)

type ProjectIsNotEmptyErr struct {
	RemainingItems []client.Object
}

func (e *ProjectIsNotEmptyErr) Error() string {
	resources := ""
	if len(e.RemainingItems) > 0 {
		resourceMap := make(map[string][]string)
		for _, item := range e.RemainingItems {
			kind := item.GetObjectKind().GroupVersionKind().Kind
			resourceMap[kind] = append(resourceMap[kind], item.GetName())
		}
		// Iterate kinds in sorted order so the message is deterministic when a single
		// blocker group spans multiple kinds (e.g. the Workloads group).
		kinds := make([]string, 0, len(resourceMap))
		for kind := range resourceMap {
			kinds = append(kinds, kind)
		}
		sort.Strings(kinds)
		for _, kind := range kinds {
			names := resourceMap[kind]
			if len(names) > 1 {
				resources += fmt.Sprintf("%s %s +%d more\n", kind, names[0], len(names)-1)
			} else {
				resources += fmt.Sprintf("%s %s\n", kind, names[0])
			}
		}
	}

	return fmt.Sprintf("The project couldn't be deleted because the following needs to be deleted first:\n%s", resources)
}

// DeleteResourceIfNeeded - Gets the object about to be deleted first,
// to make sure it's not manually overridden.  Deletes it if not.
func (handler CommonResourceDeletionHandler) DeleteResourceIfNeeded(
	resource client.Object, resourceName string, resourceNamespace string,
	logTag string, project string, deleteOptions ...client.DeleteOption) error {
	if err := handler.Client.Get(context.Background(),
		client.ObjectKey{
			Name:      resourceName,
			Namespace: resourceNamespace,
		},
		resource,
	); err != nil {
		if errors.IsNotFound(err) {
			handler.Log.Info("Resource does not exist, nothing to delete",
				logTag, resourceName, common.LogProjectTag, project)
			return nil
		} else {
			handler.Log.Error(err, "Failed to retrieve Resource for deletion",
				logTag, resourceName, common.LogProjectTag, project)
			return err
		}
	}

	return handler.DeleteExistingResourceIfNeeded(resource, resourceName, logTag, project, deleteOptions...)
}

// DeleteExistingResourceIfNeeded - makes sure it's not manually overridden.  Deletes it if not.
func (handler CommonResourceDeletionHandler) DeleteExistingResourceIfNeeded(
	resource client.Object, resourceName string,
	logTag string, project string, deleteOptions ...client.DeleteOption) error {
	if common.IsManuallyOverridden(resource) {
		handler.Log.Info("Resource is marked as manually overridden, it will not be deleted.",
			logTag, resourceName, common.LogProjectTag, project)
		return nil
	}
	return handler.deleteResource(resource, project, deleteOptions...)
}

func (handler CommonResourceDeletionHandler) deleteResource(
	obj client.Object, project string, deleteOptions ...client.DeleteOption) error {
	handler.Log.Info(fmt.Sprintf("Deleting %s '%s' owned by Project '%s'",
		obj.GetObjectKind(), obj.GetName(), project))
	if err := handler.Client.Delete(context.Background(), obj, deleteOptions...); err != nil && !errors.IsNotFound(err) {
		handler.Log.Error(err, fmt.Sprintf("Error deleting %s '%s' owned by Project '%s'",
			obj.GetObjectKind(), obj.GetName(), project))
		return err
	}
	return nil
}

func (handler CommonResourceDeletionHandler) RemoveOwnerRef(
	project *kaiv1alpha1.Project, obj client.Object, logTag string,
) error {
	var err error
	if indexOfOwner := common.IsProjectOwner(project, obj); indexOfOwner > -1 {
		handler.Log.V(6).Info("Removing Project from OwnerReferences of resource",
			logTag, obj.GetName(), common.LogNamespaceTag, obj.GetNamespace(), common.LogProjectTag, project.Name)
		deleteOwnerRef(obj, indexOfOwner)
		if innerErr := handler.Client.Update(context.Background(), obj); innerErr != nil {
			handler.Log.Error(innerErr, "Error Removing Project from OwnerReferences of resource",
				logTag, obj.GetName(), common.LogNamespaceTag, obj.GetNamespace(), common.LogProjectTag, project.Name)
			err = multierror.Append(err, innerErr)
		}
	}
	return err
}

func deleteOwnerRef(obj client.Object, indexToRemove int) {
	ownerRefs := obj.GetOwnerReferences()
	ownerRefs[indexToRemove] = ownerRefs[len(ownerRefs)-1]
	ownerRefs = ownerRefs[:len(ownerRefs)-1]
	obj.SetOwnerReferences(ownerRefs)
}

func getDeletionReasonFromError(err error, reason ProjectConditionDeletionReason) string {
	if err == nil {
		return ""
	}
	return string(reason)
}
