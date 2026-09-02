/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package deletion

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LimitRangeDeletionHandler struct {
	CommonResourceDeletionHandler
	LimitRangeName string
}

func NewLimitRangeDeletionHandler(client client.Client) LimitRangeDeletionHandler {
	return LimitRangeDeletionHandler{
		CommonResourceDeletionHandler: CommonResourceDeletionHandler{
			WithLoggerAndCli: common.WithLoggerAndCli{
				Client: client,
				Log:    ctrl.Log.WithName("deletion_handlers").WithName(common.LogLimitRangeTag),
			},
		},
		LimitRangeName: config.Get().LimitRangeName,
	}
}

func (handler LimitRangeDeletionHandler) OnDelete(project *kaiv1alpha1.Project) ([]kaiv1alpha1.ProjectCondition, error) {
	namespace, err := handler.KaiProjectToNamespace(project)
	if err != nil {
		// we have to log and forgive, otherwise deletion will remain on Terminating forever
		return []kaiv1alpha1.ProjectCondition{}, nil
	}

	err = handler.onDeleteInner(project, namespace)
	return []kaiv1alpha1.ProjectCondition{{
		Type:    "LimitRangeReady",
		Status:  handlers.GetStatusFromError(err),
		Reason:  getDeletionReasonFromError(err, LimitRangeDeletionHandlerFailed),
		Message: handlers.GetMessageFromError(err),
	}}, err
}

func (handler LimitRangeDeletionHandler) onDeleteInner(project *kaiv1alpha1.Project, namespace string) error {
	limitRange := &corev1.LimitRange{}
	if err := handler.Client.Get(context.Background(),
		client.ObjectKey{Namespace: namespace, Name: handler.LimitRangeName}, limitRange); err != nil {
		if errors.IsNotFound(err) {
			handler.Log.V(4).Info("No LimitRange found under namespace, nothing to finalize",
				common.LogNamespaceTag, namespace, common.LogProjectTag, project.Name)
			return nil
		}

		handler.Log.Error(err, "Failed to retrieve LimitRange under Namespace",
			common.LogLimitRangeTag, handler.LimitRangeName, common.LogNamespaceTag, namespace, common.LogProjectTag, project.Name)
		return err
	}
	return handler.RemoveOwnerRef(project, limitRange, common.LogLimitRangeTag)
}
