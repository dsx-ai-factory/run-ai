/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"cmp"
	"context"
	"encoding/json"
	"slices"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateApplicationStatus(
	ctx context.Context,
	client ctrlClient.Client,
	application *v1alpha1.Application,
	applicationPhase *string,
	provider ApplicationSourceProvider,
) error {
	log.Debug().Msgf("Updating application %s status ResourceVersion %s", application.Name, application.ResourceVersion)

	resources := []v1alpha1.ApplicationResource{}
	// get all resources in the release - currently support external-workload, pvc, secret, service
	workloadsForApplication, resourcesForApplication, err := getResourcesForApplication(ctx, client, application, provider)
	if err != nil {
		return err
	}
	for _, ew := range workloadsForApplication {
		// Validate resource belongs to this application
		if !provider.BelongsTo(&ew, application.Name) {
			continue
		}
		resources = append(resources, createResourceFromEw(&ew))
		// for workloads only - add label of application-id and application-name for sync (only add/update these labels)
		err = addApplicationLabelToExternalWorkload(ctx, client, ew, resolveApplicationId(application), application.Name)
		if err != nil {
			return err
		}
	}
	for _, resource := range resourcesForApplication {
		// Validate resource belongs to this application
		if !provider.BelongsTo(resource, application.Name) {
			continue
		}
		resources = append(resources, createResourceFromObject(resource))
	}

	err = getRefreshedApplication(ctx, client, application)
	if err != nil {
		return err
	}

	// Check if status actually changed before updating to reduce conflicts
	oldPhase := application.Status.K8sPhase
	oldResources := application.Status.Resources

	isEqual := compareResources(oldResources, resources)
	// Only update if something changed
	if (applicationPhase == nil || oldPhase == *applicationPhase) && isEqual {
		log.Debug().Msgf("Application %s status unchanged, skipping update", application.Name)
		return nil
	}

	if applicationPhase != nil {
		application.Status.K8sPhase = *applicationPhase
	}
	application.Status.Resources = resources

	err = client.Status().Update(ctx, application)
	if err != nil {
		if apierrors.IsConflict(err) {
			// Conflict errors are expected when multiple reconciliations happen concurrently
			// The controller-runtime will automatically retry, so just log and return the error
			log.Debug().Msgf("Conflict updating application %s status, will retry: %v", application.Name, err)
		}
		return err
	}
	log.Debug().Msgf("Updating application %s status done", application.Name)
	return nil
}

func compareResources(a, b []v1alpha1.ApplicationResource) bool {
	// 1. Sort both slices by Name
	// Note: This modifies the order of 'a' and 'b' in memory!
	slices.SortFunc(a, func(i, j v1alpha1.ApplicationResource) int {
		return cmp.Compare(i.UID, j.UID)
	})

	slices.SortFunc(b, func(i, j v1alpha1.ApplicationResource) int {
		return cmp.Compare(i.UID, j.UID)
	})

	// 2. Compare the sorted slices
	return slices.Equal(a, b)
}

func getRefreshedApplication(ctx context.Context, cachedClient ctrlClient.Reader, application *v1alpha1.Application) (err error) {
	applicationNamespacedName := types.NamespacedName{Name: application.Name, Namespace: application.Namespace}
	err = cachedClient.Get(ctx, applicationNamespacedName, application)
	if err != nil {
		return err
	}
	return nil
}

// resolveApplicationId returns the control plane application ID from the annotation
// if present (run:ai-managed applications), otherwise falls back to the Kubernetes CR UID.
func resolveApplicationId(application *v1alpha1.Application) string {
	if application.Annotations != nil {
		if cpId, ok := application.Annotations[constants.ControlPlaneApplicationIdAnnotation]; ok && cpId != "" {
			return cpId
		}
	}
	return string(application.UID)
}

func addApplicationLabelToExternalWorkload(ctx context.Context, client ctrlClient.Client, ew v1alpha1.ExternalWorkload, applicationId string, applicationName string) error {
	labelsMap := ew.GetLabels()
	if labelsMap == nil {
		labelsMap = map[string]string{}
	}
	existingAppId, appIdExists := labelsMap[constants.ApplicationIdLabel]
	existingAppName, appNameExists := labelsMap[constants.ApplicationNameLabel]

	// Check if both labels already exist with correct values
	if appIdExists && existingAppId == applicationId && appNameExists && existingAppName == applicationName {
		return nil
	}

	labelsToPatch := map[string]string{
		constants.ApplicationIdLabel:   applicationId,
		constants.ApplicationNameLabel: applicationName,
	}
	return patchLabels(ctx, client, labelsToPatch, ew)
}

func patchLabels(ctx context.Context, client ctrlClient.Client, labelsToPatch map[string]string, ew v1alpha1.ExternalWorkload) error {
	patchBytes, err := getPatchBytes(labelsToPatch)
	if err != nil {
		return err
	}

	patch := ctrlClient.RawPatch(types.MergePatchType, patchBytes)
	err = client.Patch(ctx, &ew, patch)
	if err != nil {
		if apierrors.IsConflict(err) {
			// Conflict errors are expected when the ExternalWorkload is being updated concurrently
			// The controller-runtime will automatically retry, so just log and return the error
			log.Debug().Msgf("Conflict patching ExternalWorkload %s/%s, will retry: %v", ew.Namespace, ew.Name, err)
		}
		return err
	}
	return nil
}

func getPatchBytes(labels map[string]string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"labels": labels,
	}})
}

func createResourceFromObject(resource ctrlClient.Object) v1alpha1.ApplicationResource {
	uid := string(resource.GetUID())
	return v1alpha1.ApplicationResource{
		Name:             resource.GetName(),
		Kind:             resource.GetObjectKind().GroupVersionKind().Kind,
		UID:              uid,
		ControlPlaneId:   uid,
		ControlPlaneName: resource.GetName(),
	}
}

func createResourceFromEw(resource *v1alpha1.ExternalWorkload) v1alpha1.ApplicationResource {
	uid := string(resource.GetUID())
	return v1alpha1.ApplicationResource{
		Name:             resource.GetName(),
		Kind:             resource.GetObjectKind().GroupVersionKind().Kind,
		UID:              uid,
		ControlPlaneId:   string(resource.Spec.OriginalWorkloadUID),
		ControlPlaneName: resource.Spec.OriginalWorkloadName,
	}
}

func IsApplicationNamespace(ctx context.Context, client ctrlClient.Client, namespaceName string) bool {
	namespace := corev1.Namespace{}
	err := client.Get(ctx, ctrlClient.ObjectKey{Name: namespaceName}, &namespace)
	if err != nil {
		log.Error().Msgf("Failed to get namespace %s: %v", namespaceName, err)
		return false
	}
	_, hasRunaiQueueLabel := namespace.Labels[constants.RunaiQueueLabel]
	if !hasRunaiQueueLabel {
		return false
	}
	enforceRunaiScheduler, ok := namespace.Annotations[constants.EnforceSchedulerAnnotationName]
	if !ok || enforceRunaiScheduler != "true" {
		log.Warn().Msgf("namespace %s has %s label but does not have %s annotation: val- %s", namespace.Name, constants.RunaiQueueLabel, constants.EnforceSchedulerAnnotationName, enforceRunaiScheduler)
		return false
	}
	return true
}

func getResourcesForApplication(ctx context.Context, client ctrlClient.Client, application *v1alpha1.Application, provider ApplicationSourceProvider) ([]v1alpha1.ExternalWorkload, []ctrlClient.Object, error) {
	workloadsForApplication, err := getWorkloads(ctx, client, application, provider)
	if err != nil {
		return nil, nil, err
	}

	resourcesForApplication, err := getPvcs(ctx, client, application, provider)
	if err != nil {
		return nil, nil, err
	}
	cms, err := getCms(ctx, client, application, provider)
	if err != nil {
		return nil, nil, err
	}
	resourcesForApplication = append(resourcesForApplication, cms...)
	secrets, err := getSecrets(ctx, client, application, provider)
	if err != nil {
		return nil, nil, err
	}
	resourcesForApplication = append(resourcesForApplication, secrets...)
	services, err := getServices(ctx, client, application, provider)
	if err != nil {
		return nil, nil, err
	}
	resourcesForApplication = append(resourcesForApplication, services...)
	return workloadsForApplication, resourcesForApplication, nil
}

func getWorkloads(ctx context.Context, client ctrlClient.Client, application *v1alpha1.Application, provider ApplicationSourceProvider) ([]v1alpha1.ExternalWorkload, error) {
	workloads := &v1alpha1.ExternalWorkloadList{}
	err := client.List(ctx, workloads, provider.GetListOptions(application.Namespace))
	if err != nil {
		return nil, err
	}
	return workloads.Items, nil
}

func getPvcs(ctx context.Context, client ctrlClient.Client, application *v1alpha1.Application, provider ApplicationSourceProvider) ([]ctrlClient.Object, error) {
	pvcs := &corev1.PersistentVolumeClaimList{}
	err := client.List(ctx, pvcs, provider.GetListOptions(application.Namespace))
	if err != nil {
		return nil, err
	}
	result := make([]ctrlClient.Object, len(pvcs.Items))
	for i := range pvcs.Items {
		result[i] = &pvcs.Items[i]
	}
	return result, nil
}

func getCms(ctx context.Context, client ctrlClient.Client, application *v1alpha1.Application, provider ApplicationSourceProvider) ([]ctrlClient.Object, error) {
	cms := &corev1.ConfigMapList{}
	err := client.List(ctx, cms, provider.GetListOptions(application.Namespace))
	if err != nil {
		return nil, err
	}
	result := make([]ctrlClient.Object, len(cms.Items))
	for i := range cms.Items {
		result[i] = &cms.Items[i]
	}
	return result, nil
}

func getSecrets(ctx context.Context, client ctrlClient.Client, application *v1alpha1.Application, provider ApplicationSourceProvider) ([]ctrlClient.Object, error) {
	secrets := &corev1.SecretList{}
	err := client.List(ctx, secrets, provider.GetListOptions(application.Namespace))
	if err != nil {
		return nil, err
	}
	result := make([]ctrlClient.Object, len(secrets.Items))
	for i := range secrets.Items {
		result[i] = &secrets.Items[i]
	}
	return result, nil
}

func getServices(ctx context.Context, client ctrlClient.Client, application *v1alpha1.Application, provider ApplicationSourceProvider) ([]ctrlClient.Object, error) {
	services := &corev1.ServiceList{}
	err := client.List(ctx, services, provider.GetListOptions(application.Namespace))
	if err != nil {
		return nil, err
	}
	result := make([]ctrlClient.Object, len(services.Items))
	for i := range services.Items {
		result[i] = &services.Items[i]
	}
	return result, nil
}
