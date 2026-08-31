/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ReadyCondition = "LimitRangeReady"

// The LimitRangeResourceHandler acquires the default limit ranges Config Map (under Scheduler Namespace) and applies it as a LimitRange resource under the Project's Namespace
type LimitRangeResourceHandler struct {
	common.WithLoggerAndCli
	LimitRangeName string
}

func NewLimitRangeResourceHandler(client client.Client) LimitRangeResourceHandler {
	return LimitRangeResourceHandler{
		WithLoggerAndCli: common.WithLoggerAndCli{
			Client: client,
			Log:    ctrl.Log.WithName("resource_handlers").WithName(common.LogLimitRangeTag),
		},
		LimitRangeName: config.Get().LimitRangeName,
	}
}

func (handler LimitRangeResourceHandler) HandleResource(
	project kaiv1alpha1.Project,
) ([]kaiv1alpha1.ProjectCondition, error) {
	err := handler.handleResourceInner(project)
	return []kaiv1alpha1.ProjectCondition{{
		Type:    ReadyCondition,
		Status:  GetStatusFromError(err),
		Reason:  getReasonFromError(err, LimitRangeHandlerFailed),
		Message: GetMessageFromError(err),
	}}, err
}

func (handler LimitRangeResourceHandler) handleResourceInner(project kaiv1alpha1.Project) error {
	var defaultLimitRangeConfigMap = &corev1.ConfigMap{}
	if getErr := handler.Client.Get(context.Background(),
		client.ObjectKey{Namespace: config.Get().InstallNamespace, Name: common.DefaultLimitRangeConfigMapName},
		defaultLimitRangeConfigMap,
	); getErr != nil {
		return handler.handleGetConfigMapError(getErr)
	}
	return handler.handleLimitRange(defaultLimitRangeConfigMap, project)
}

func (handler LimitRangeResourceHandler) handleLimitRange(
	defaultLimitRangeConfigMap *corev1.ConfigMap, project kaiv1alpha1.Project,
) (err error) {
	var (
		limitRange   corev1.LimitRange
		shouldCreate bool
	)
	if limitRange, shouldCreate, err = handler.GetNamespacedLimitRange(project); err != nil {
		handler.Log.Error(err, "Failed to retrieve Limit Range for project", common.LogLimitRangeTag, handler.LimitRangeName, common.LogNamespaceTag, project.Name)
		return err
	} else if common.IsManuallyOverridden(limitRange.GetObjectMeta()) {
		handler.Log.Info(
			"LimitRange already exists and is marked as manually overridden, it will not be updated",
			common.LogLimitRangeTag, limitRange.Name, common.LogNamespaceTag,
			limitRange.Namespace, common.LogProjectTag, project.Name)
		return nil
	}
	handler.populateLimitRange(defaultLimitRangeConfigMap, &limitRange)
	return handler.createOrUpdateLimitRange(shouldCreate, &limitRange)
}

func (handler LimitRangeResourceHandler) GetNamespacedLimitRange(
	project kaiv1alpha1.Project,
) (limitRange corev1.LimitRange, shouldCreate bool, err error) {
	limitRangeName := handler.LimitRangeName
	namespace, err := handler.KaiProjectToNamespace(&project)
	if err != nil {
		handler.Log.Error(err, "Failed to derive namespace, cannot create", common.LogProjectTag, project.Name)
		shouldCreate = false
		return
	}

	if err = handler.Client.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: limitRangeName}, &limitRange); err != nil {
		if errors.IsNotFound(err) {
			// Namespaced Limit Range doesn't exist yet - We'll create it
			handler.Log.V(4).Info("Namespaced LimitRange not found, it will be created.", common.LogLimitRangeTag, limitRangeName, common.LogNamespaceTag, namespace)
			err = nil
			shouldCreate = true
			limitRange = buildEmptyLimitRange(project, namespace, handler.LimitRangeName)
		}
	}
	return limitRange, shouldCreate, err
}

func (handler LimitRangeResourceHandler) populateLimitRange(
	defaultLimitRangeConfigMap *corev1.ConfigMap, limitRange *corev1.LimitRange,
) {
	resetLimits(limitRange)
	handler.reconcileLimitRangeWithDefaultConfigMap(defaultLimitRangeConfigMap, limitRange)
}

func (handler LimitRangeResourceHandler) createOrUpdateLimitRange(shouldCreate bool, limitRange *corev1.LimitRange) (err error) {
	if shouldCreate {
		err = handler.Client.Create(context.Background(), limitRange)
	} else {
		err = handler.Client.Update(context.Background(), limitRange)
	}
	return err
}

func (handler LimitRangeResourceHandler) reconcileLimitRangeWithDefaultConfigMap(defaultLimitRangeConfigMap *corev1.ConfigMap, limitRange *corev1.LimitRange) {
	handler.assignLimitValueToResourceList(defaultLimitRangeConfigMap.Data, limitRange.Spec.Limits[0].DefaultRequest, common.CpuDefaultRequest, common.CpuResourceName)
	handler.assignLimitValueToResourceList(defaultLimitRangeConfigMap.Data, limitRange.Spec.Limits[0].DefaultRequest, common.MemoryDefaultRequest, common.MemoryResourceName)

	handler.assignLimitValueToResourceList(defaultLimitRangeConfigMap.Data, limitRange.Spec.Limits[0].Default, common.CpuDefaultLimit, common.CpuResourceName)
	handler.assignLimitValueToResourceList(defaultLimitRangeConfigMap.Data, limitRange.Spec.Limits[0].Default, common.MemoryDefaultLimit, common.MemoryResourceName)

	handler.assignLimitValueToResourceList(defaultLimitRangeConfigMap.Data, limitRange.Spec.Limits[0].Max, common.CpuMaxLimit, common.CpuResourceName)
	handler.assignLimitValueToResourceList(defaultLimitRangeConfigMap.Data, limitRange.Spec.Limits[0].Max, common.MemoryMaxLimit, common.MemoryResourceName)
}

func (handler LimitRangeResourceHandler) assignLimitValueToResourceList(limitsMap map[string]string, resourceList corev1.ResourceList, field string, resourceName corev1.ResourceName) {
	if unit, exists := limitsMap[field]; exists {
		if quantity, err := resource.ParseQuantity(unit); err != nil {
			handler.Log.Error(err, "Failed to parse unit for resource from limit range field", "Field", field, "Unit", unit, "Resource", resourceName)
		} else {
			resourceList[resourceName] = quantity
		}
	}
}

func (handler LimitRangeResourceHandler) handleGetConfigMapError(getErr error) error {
	if errors.IsNotFound(getErr) {
		// nothing to do
		handler.Log.Info("Default LimitRange Config Map does not exist, skipping LimitRange reconciliation.")
		return nil
	} else {
		handler.Log.Error(getErr, "Failed to retrieve the default limit range Config Map under the install namespace")
		return getErr
	}
}

func buildEmptyLimitRange(project kaiv1alpha1.Project, namespace string, limitRangeName string) (limitRange corev1.LimitRange) {
	limitRange = corev1.LimitRange{
		TypeMeta: metav1.TypeMeta{
			Kind:       common.LimitRangeKind,
			APIVersion: common.CoreV1ApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      limitRangeName,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         kaiv1alpha1.GroupVersion.Identifier(),
				Kind:               common.ProjectKind,
				Name:               project.Name,
				UID:                project.UID,
				Controller:         &common.TrueRef,
				BlockOwnerDeletion: &common.TrueRef,
			}},
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{{
				Type:           corev1.LimitTypeContainer,
				Max:            make(map[corev1.ResourceName]resource.Quantity),
				Default:        make(map[corev1.ResourceName]resource.Quantity),
				DefaultRequest: make(map[corev1.ResourceName]resource.Quantity),
			}},
		},
	}
	return limitRange
}

func resetLimits(limitRange *corev1.LimitRange) {
	limitRange.Spec.Limits = []corev1.LimitRangeItem{{
		Type:           corev1.LimitTypeContainer,
		Max:            make(map[corev1.ResourceName]resource.Quantity),
		Default:        make(map[corev1.ResourceName]resource.Quantity),
		DefaultRequest: make(map[corev1.ResourceName]resource.Quantity),
	}}
}
