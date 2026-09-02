/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package deletion

import (
	"context"
	"os"
	"strings"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const DeleteNamespaceFeatureFlag = "RUN_AI_PROJECT_CONTROLLER_FEATURE_FINALIZE_NS"

type NamespaceDeletionHandler struct {
	CommonResourceDeletionHandler
	deleteNamespaceFeatureEnabled  bool
	createNamespacesFeatureEnabled bool
}

func NewNamespaceDeletionHandler(client client.Client, createNamespacesFeatureEnabled bool) NamespaceDeletionHandler {
	featureFlag := strings.ToLower(os.Getenv(DeleteNamespaceFeatureFlag))
	deleteNamespaceFeatureEnabled := featureFlag == "true" || featureFlag == "enabled"

	return NamespaceDeletionHandler{
		deleteNamespaceFeatureEnabled: deleteNamespaceFeatureEnabled,
		CommonResourceDeletionHandler: CommonResourceDeletionHandler{
			WithLoggerAndCli: common.WithLoggerAndCli{
				Client: client,
				Log:    ctrl.Log.WithName("deletion_handlers").WithName(common.LogNamespaceTag),
			},
		},
		createNamespacesFeatureEnabled: createNamespacesFeatureEnabled,
	}
}

func (handler NamespaceDeletionHandler) OnDelete(project *kaiv1alpha1.Project) ([]kaiv1alpha1.ProjectCondition, error) {
	err := handler.onDeleteInner(project)
	return []kaiv1alpha1.ProjectCondition{{
		Type:    kaiv1alpha1.NamespaceReady,
		Status:  handlers.GetStatusFromError(err),
		Reason:  getDeletionReasonFromError(err, NamespaceDeletionHandlerFailed),
		Message: handlers.GetMessageFromError(err),
	}}, err
}

func (handler NamespaceDeletionHandler) onDeleteInner(project *kaiv1alpha1.Project) error {
	if !handler.createNamespacesFeatureEnabled {
		return nil
	}

	namespaceName, err := handler.KaiProjectToNamespace(project)
	if err != nil {
		return err
	}

	namespace := &corev1.Namespace{}
	if project.Spec.Namespace == "" && handler.deleteNamespaceFeatureEnabled {
		return handler.deleteNamespace(project.Name, namespaceName)
	}

	if err = handler.Client.Get(context.Background(), client.ObjectKey{Name: namespaceName}, namespace); err != nil {
		handler.Log.Error(err, "Error retrieving Namespace during finalization",
			common.LogNamespaceTag, namespaceName, common.LogProjectTag, project.Name)
		return err
	}

	return handler.RemoveOwnerRef(project, namespace, common.LogNamespaceTag)
}

func (handler NamespaceDeletionHandler) deleteNamespace(projectName string, namespaceName string) error {
	handler.Log.Info("Deleting namespace and Namespaced resources",
		common.LogNamespaceTag, namespaceName, common.LogProjectTag, projectName)
	// This will cause all secondary namespaced resources to get deleted
	return handler.DeleteResourceIfNeeded(&corev1.Namespace{}, namespaceName, "",
		common.LogNamespaceTag, projectName)
}
