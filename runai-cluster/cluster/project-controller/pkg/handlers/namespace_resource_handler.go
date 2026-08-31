/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NamespaceResourceHandler struct {
	common.WithLoggerAndCli
	isOpenshift      bool
	createNamespaces bool
}

func NewNamespaceResourceHandler(client client.Client, isOpenshift, createNamespaces bool) NamespaceResourceHandler {
	return NamespaceResourceHandler{
		WithLoggerAndCli: common.WithLoggerAndCli{
			Client: client,
			Log:    ctrl.Log.WithName("resource_handlers").WithName(common.LogNamespaceTag),
		},
		isOpenshift:      isOpenshift,
		createNamespaces: createNamespaces,
	}
}

func (handler NamespaceResourceHandler) HandleResource(
	project kaiv1alpha1.Project,
) ([]kaiv1alpha1.ProjectCondition, error) {
	statusMsg, reason, err := handler.handleResourceInner(project)
	return []kaiv1alpha1.ProjectCondition{{
		Type:    kaiv1alpha1.NamespaceReady,
		Status:  GetStatusFromError(err),
		Reason:  getReasonFromError(err, reason),
		Message: statusMsg,
	}}, err
}

func (handler NamespaceResourceHandler) handleResourceInner(project kaiv1alpha1.Project,
) (string, ProjectConditionReason, error) {
	if !handler.createNamespaces {
		// just validate the namespace exists
		_, err := handler.KaiProjectToNamespace(&project)
		if err != nil {
			return common.ProjectsNamespaceNotFoundStatusMsg, NamespaceNotFound, err
		}
		return "", "", nil
	}

	if project.Spec.Namespace != "" {
		return handler.handleUserExternalNamespaceForProject(&project)
	}

	namespaceName := DefaultProjectNamespaceName(&project)
	namespaceObjToCreate := GetNamespaceObject(project, namespaceName, handler.isOpenshift)

	existingNamespace, err := handler.GetNamespace(namespaceName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return handler.handleNonExistingNamespace(namespaceObjToCreate, project.Name)
		}

		handler.Log.Error(err, "Error during Namespace retrieval",
			common.LogNamespaceTag, namespaceName, common.LogProjectTag, project.Name)
		return common.FailedGetNamespaceStatusMsg, NamespaceGetFailed, err
	}

	return handler.handleExistingNamespace(existingNamespace, *namespaceObjToCreate, project.Name)
}

func (handler NamespaceResourceHandler) handleExistingNamespace(
	existingNamespace corev1.Namespace, newNamespace corev1.Namespace, projectName string,
) (statusMsg string, reason ProjectConditionReason, err error) {
	if common.IsManuallyOverridden(&existingNamespace) {
		handler.Log.Info("Namespace already exists and is marked as manually overridden, it will not be updated",
			common.LogNamespaceTag, existingNamespace.Name, common.LogProjectTag, projectName)
		return "", "", nil
	}

	if NamespacesEqual(newNamespace, existingNamespace) {
		// New and existing are identical, no need to act.
		handler.Log.V(4).Info("Existing Namespace identical to incoming Project spec, skipping",
			common.LogNamespaceTag, existingNamespace.Name, common.LogProjectTag, projectName)
		return "", "", nil
	}

	// Required for updates
	newNamespace.ResourceVersion = existingNamespace.ResourceVersion
	newNamespace.UID = existingNamespace.UID
	newNamespace.Finalizers = existingNamespace.Finalizers

	// Carry over annotations set by anyone else on the namespace, but never the ones this
	// handler owns: newNamespace already holds their value as derived from the Project spec,
	// and copying the existing value back would undo the very change we are about to write.
	if newNamespace.Annotations == nil {
		newNamespace.Annotations = map[string]string{}
	}
	for key, value := range existingNamespace.Annotations {
		if key == config.Get().EnforceSchedulerAnnotationKey {
			continue
		}
		newNamespace.Annotations[key] = value
	}

	// New and existing differ, need to update
	handler.Log.Info("Existing Namespace differs from Project spec, updating Namespace.",
		common.LogNamespaceTag, existingNamespace.Name, common.LogProjectTag, projectName)
	if err = handler.Client.Update(context.Background(), &newNamespace); err != nil {
		handler.Log.Error(err, "Error updating Namespace",
			common.LogNamespaceTag, existingNamespace.Name, common.LogProjectTag, projectName)
		return common.FailedUpdateNamespaceStatusMsg, NamespaceUpdateFailed, err
	}
	return "", "", nil
}

func (handler NamespaceResourceHandler) handleNonExistingNamespace(
	newNamespace *corev1.Namespace, projectName string,
) (statusMsg string, reason ProjectConditionReason, err error) {
	if err = handler.Client.Create(context.Background(), newNamespace); err != nil && !apierrors.IsAlreadyExists(err) {
		handler.Log.Error(err, "Error creating Namespace",
			common.LogNamespaceTag, newNamespace.Name, common.LogProjectTag, projectName)
		return common.FailedCreateNamespaceStatusMsg, NamespaceCreateFailed, err
	}
	handler.Log.Info("Successfully created namespace for project",
		common.LogNamespaceTag, newNamespace.Name, common.LogProjectTag, projectName)
	return "", "", nil
}

func (handler NamespaceResourceHandler) handleUserExternalNamespaceForProject(project *kaiv1alpha1.Project,
) (string, ProjectConditionReason, error) {
	namespaceName := project.Spec.Namespace

	existingNamespace, err := handler.GetNamespace(namespaceName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			handler.Log.Error(err, "Can't find external namespace that should exist for project",
				common.LogNamespaceTag, namespaceName, common.LogProjectTag, project.Name)
			return common.ProjectsNamespaceNotFoundStatusMsg, NamespaceNotFound, err
		}

		handler.Log.Error(err, "Error during Namespace retrieval - external namespace should exist for project",
			common.LogNamespaceTag, namespaceName, common.LogProjectTag, project.Name)
		return common.FailedGetNamespaceStatusMsg, NamespaceGetFailed, err
	}

	return handler.labelExternalNamespaceForProject(project, existingNamespace)
}

func (handler NamespaceResourceHandler) labelExternalNamespaceForProject(project *kaiv1alpha1.Project, existingNamespace corev1.Namespace) (string, ProjectConditionReason, error) {
	updated := false
	if existingNamespace.Labels == nil {
		existingNamespace.Labels = map[string]string{}
	}

	// Labels the namespace to mark it as kai namespace
	projectNameLabel, found := existingNamespace.Labels[config.Get().NamespaceProjectLabelKey]
	if !found {
		existingNamespace.Labels[config.Get().NamespaceProjectLabelKey] = project.Name
		updated = true
	}
	if projectNameLabel != "" && projectNameLabel != project.Name {
		// If the label exists but with another project name,
		// we want to avoid an infinite live lock that every project update the label again and again
		handler.Log.Error(nil, "External Namespace already has another project label, cant set more than one project to a namespace",
			common.LogNamespaceTag, existingNamespace.Name, common.LogProjectTag, projectNameLabel, common.LogProjectTag, project.Name)
		statusMsg := fmt.Sprintf("'%s' %s", config.Get().NamespaceProjectLabelKey, common.QueueLabelMissingFromNamespaceStatusMsg)
		return statusMsg, NamespaceLabelMissing, errors.New(statusMsg)
	}

	if existingNamespace.Annotations == nil {
		existingNamespace.Annotations = map[string]string{}
	}

	// Annotate the namespace with enforce scheduler value from project
	enforceAnnotation, found := existingNamespace.Annotations[config.Get().EnforceSchedulerAnnotationKey]
	if !found || enforceAnnotation != strconv.FormatBool(project.Spec.EnforceKaiScheduler) {
		existingNamespace.Annotations[config.Get().EnforceSchedulerAnnotationKey] = strconv.FormatBool(project.Spec.EnforceKaiScheduler)
		updated = true
	}

	if updated {
		handler.Log.Info("External Namespace missing labels or annotations, updating Namespace.",
			common.LogNamespaceTag, existingNamespace.Name, common.LogProjectTag, project.Name)
		if err := handler.Client.Update(context.Background(), &existingNamespace); err != nil {
			handler.Log.Error(err, "Error updating Namespace",
				common.LogNamespaceTag, existingNamespace.Name, common.LogProjectTag, project.Name)
			return common.FailedUpdateNamespaceStatusMsg, NamespaceUpdateFailed, err
		}
	}

	return "", "", nil
}

func GetNamespaceObject(project kaiv1alpha1.Project, namespaceName string, isOpenshift bool) *corev1.Namespace {
	newNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: common.CoreV1ApiVersion,
			Kind:       common.NamespaceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				config.Get().NamespaceProjectLabelKey: project.Name,
				config.Get().NamespaceVersionLabelKey: common.AgentNamespaceVersion,
			},
			Annotations: map[string]string{
				config.Get().EnforceSchedulerAnnotationKey: strconv.FormatBool(project.Spec.EnforceKaiScheduler),
			},
		},
	}

	if isOpenshift {
		newNamespace.Labels[common.OcpClusterMonitoringLabel] = "true"
	}

	return newNamespace
}

func NamespacesEqual(left corev1.Namespace, right corev1.Namespace) bool {
	return reflect.DeepEqual(left.Kind, right.Kind) &&
		reflect.DeepEqual(left.APIVersion, right.APIVersion) &&
		reflect.DeepEqual(left.Name, right.Name) &&
		left.Labels != nil && right.Labels != nil &&
		reflect.DeepEqual(left.Labels[config.Get().NamespaceProjectLabelKey], right.Labels[config.Get().NamespaceProjectLabelKey]) &&
		left.Annotations[config.Get().EnforceSchedulerAnnotationKey] ==
			right.Annotations[config.Get().EnforceSchedulerAnnotationKey]
}

// DefaultProjectNamespaceName is relevant ONLY when creating a new namespace and naming it in the default
// naming scheme of config.NamespacePrefix()-<project>. For any other case, KaiProjectToNamespace must be used
// to derive namespace from the queue label.
func DefaultProjectNamespaceName(project *kaiv1alpha1.Project) string {
	prefix := config.NamespacePrefix()
	return fmt.Sprintf("%s%s", prefix, project.Name)
}
