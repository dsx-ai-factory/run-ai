/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managers

import (
	"context"
	"errors"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/issue"
	issue2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/events"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/resources"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sort"
	"strings"
)

type UpdateStrategy int

const (
	Merge UpdateStrategy = iota
	Override
	Ignored

	clientErrorHttpCodeMin   = 400
	clientErrorHttpCodeMax   = 500
	objectHasBeenModifiedMsg = "the object has been modified; please apply your changes to the latest version and try again"

	// event recording
	CreateSucceededReason = "SuccessfulCreateResource"
	CreateFailedReason    = "FailedCreateResource"

	CauseFailedResourceRequestNotFound = "the server could not find the requested resource"
)

type MetadataStrategyConfig struct {
	Labels      UpdateStrategy
	Annotations UpdateStrategy
	Finalizers  UpdateStrategy
}

// CreateOrUpdateIfNeeded creates or updates the given resource if needed.
// It returns true if the resource was created or updated, and false otherwise.
func UpsertIfNeededCheckOwnerUID(ctx context.Context, k8sClient client.Client, apiReadClient client.Reader, resource client.Object,
	ownerUID types.UID, jobName string, createOptions client.CreateOptions,
	updateOptions client.UpdateOptions, shouldUpdate bool, rg issue.ResourceGrouper) (bool, *issue.Issue) {

	return UpsertWithUpdateStrategy(ctx, k8sClient, apiReadClient, resource, ownerUID, jobName, createOptions, updateOptions, shouldUpdate,
		&MetadataStrategyConfig{
			Labels:      Merge,
			Annotations: Merge,
			Finalizers:  Ignored,
		}, rg)
}

func UpsertWithUpdateStrategy(ctx context.Context, k8sClient client.Client, apiReadClient client.Reader,
	resource client.Object, ownerUID types.UID,
	jobName string, createOptions client.CreateOptions,
	updateOptions client.UpdateOptions, shouldUpdate bool, config *MetadataStrategyConfig, rg issue.ResourceGrouper) (bool, *issue.Issue) {
	log := logf.FromContext(ctx)
	resourceType := resources.GetResourceType(resource)

	existingResource, found, err := getExistingResource(ctx, k8sClient, resource)

	if err != nil {
		return false, issue.New(resource, err, false, "", rg)
	}

	if found {
		return handleExistingResource(ctx, k8sClient, resource, existingResource, ownerUID, jobName, updateOptions, shouldUpdate, log, resourceType, config, rg)
	}

	created, issueItem := createResource(ctx, k8sClient, resource, createOptions, log, resourceType, rg)
	if issueItem != nil && issueItem.GetGroup() == issue.Network && apierrors.IsAlreadyExists(issueItem.GetError()) {
		return handleDuplicateNetworkResource(ctx, k8sClient, apiReadClient, resource, ownerUID, jobName, updateOptions, shouldUpdate, log, resourceType, config, rg, issueItem)
	}
	return created, issueItem
}

func handleDuplicateNetworkResource(ctx context.Context, k8sClient client.Client, apiReadClient client.Reader, resource client.Object, ownerUID types.UID, jobName string,
	updateOptions client.UpdateOptions, shouldUpdate bool, log logr.Logger, resourceType string, config *MetadataStrategyConfig, rg issue.ResourceGrouper, issueItem *issue.Issue) (bool, *issue.Issue) {
	log.Info(fmt.Sprintf("resource [%s/%s/%s] already exists, re-fetching to apply ownership check",
		resource.GetNamespace(), resourceType, resource.GetName()))
	existingResource, found, err := getExistingResourceWithReader(ctx, apiReadClient, resource)
	if err != nil {
		return false, issue.New(resource, err, false, "", rg)
	}
	if found {
		return handleExistingResource(ctx, k8sClient, resource, existingResource,
			ownerUID, jobName, updateOptions, shouldUpdate, log, resourceType, config, rg)
	}
	// should not happen, in case it do happens better to not failed the workload
	return false, issue.New(resource, issueItem.GetError(), false, "", rg)
}

func getExistingResource(ctx context.Context, k8sClient client.Client, resource client.Object) (client.Object, bool, error) {
	return getExistingResourceWithReader(ctx, k8sClient, resource)
}

func getExistingResourceWithReader(ctx context.Context, reader client.Reader, resource client.Object) (client.Object, bool, error) {
	existingResource := reflect.New(reflect.TypeOf(resource).Elem()).Interface().(client.Object)
	err := reader.Get(ctx, types.NamespacedName{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}, existingResource)

	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return existingResource, true, nil
}

func handleExistingResource(ctx context.Context, k8sClient client.Client, resource, existingResource client.Object, ownerUID types.UID, jobName string,
	updateOptions client.UpdateOptions, shouldUpdate bool, log logr.Logger, resourceType string, config *MetadataStrategyConfig, rg issue.ResourceGrouper) (bool, *issue.Issue) {

	if !isResourceOwnedByJob(existingResource, ownerUID, jobName) {
		log.V(4).Info("Resource already exists and is not owned by the job", "resource", resourceType, "name", resource.GetName())
		return false, issue.New(resource, apierrors.NewAlreadyExists(schema.GroupResource{}, resource.GetName()), false, "", rg)
	}

	if !shouldUpdate {
		log.V(4).Info("Resource already exists, skipping update", "resource", resourceType, "name", resource.GetName())
		resource.SetUID(existingResource.GetUID())
		return false, nil
	}

	handleMetadataStrategy(resource, existingResource, config)

	return updateExistingResource(ctx, k8sClient, resource, existingResource, updateOptions, log, resourceType, rg)
}

func handleMetadataStrategy(newResource, existingResource client.Object, config *MetadataStrategyConfig) {

	// Handle Labels
	switch config.Labels {
	case Override:
		// Do nothing, the labels are already set
	case Merge:
		mergeLabels(existingResource, newResource)
	case Ignored:
		newResource.SetLabels(existingResource.GetLabels())
	}

	// Handle Annotations
	switch config.Annotations {
	case Override:
		// Do nothing
	case Merge:
		mergeAnnotations(existingResource, newResource)
	case Ignored:
		newResource.SetAnnotations(existingResource.GetAnnotations())
	}

	// Handle Finalizers
	switch config.Finalizers {
	case Override:
		// Do nothing
	case Merge:
		mergeFinalizers(existingResource, newResource)
	case Ignored:
		newResource.SetFinalizers(existingResource.GetFinalizers())
	}
}

func mergeLabels(existingResource, newResource client.Object) {
	existingLabels := existingResource.GetLabels()
	if existingLabels == nil {
		existingLabels = make(map[string]string)
	}
	for k, v := range newResource.GetLabels() {
		existingLabels[k] = v
	}
	newResource.SetLabels(existingLabels)
}

func mergeAnnotations(existingResource, newResource client.Object) {
	existingAnnotations := existingResource.GetAnnotations()
	if existingAnnotations == nil {
		existingAnnotations = make(map[string]string)
	}
	for k, v := range newResource.GetAnnotations() {
		existingAnnotations[k] = v
	}
	newResource.SetAnnotations(existingAnnotations)
}

func mergeFinalizers(existingResource, newResource client.Object) {
	existingFinalizers := existingResource.GetFinalizers()
	newFinalizers := newResource.GetFinalizers()
	if len(newFinalizers) == 0 {
		if len(existingFinalizers) == 0 {
			return
		}
		newResource.SetFinalizers(existingFinalizers)
		return
	}

	// Create a map to easily check for existing finalizers
	finalizersMap := make(map[string]struct{})
	for _, finalizer := range existingFinalizers {
		finalizersMap[finalizer] = struct{}{}
	}

	// Merge existing and new finalizers
	mergedFinalizers := make([]string, 0)

	// First, add all new finalizers
	for _, newFinalizer := range newFinalizers {
		mergedFinalizers = append(mergedFinalizers, newFinalizer)
		delete(finalizersMap, newFinalizer)
	}

	// Then, add any remaining existing finalizers that were not in the new set
	for existingFinalizer := range finalizersMap {
		mergedFinalizers = append(mergedFinalizers, existingFinalizer)
	}
	// sort the finalizers to make sure the order is deterministic
	// this is important for the finalizers to be removed in the correct order
	sort.Slice(mergedFinalizers, func(i, j int) bool {
		return strings.Compare(mergedFinalizers[i], mergedFinalizers[j]) < 0
	})
	// Set the merged finalizers to the new resource
	newResource.SetFinalizers(mergedFinalizers)
}

func isResourceOwnedByJob(resource client.Object, ownerUID types.UID, jobName string) bool {
	return resources.BackwardCompatibilityIsObjectCreatedBy(resource, jobName) || resources.IsObjectOwnedBy(resource, ownerUID)
}

func updateExistingResource(ctx context.Context, k8sClient client.Client, resource, existingResource client.Object,
	updateOptions client.UpdateOptions, log logr.Logger, resourceType string, rg issue.ResourceGrouper) (bool, *issue.Issue) {

	log.Info(fmt.Sprintf("updating resource [%s/%s/%s]",
		resource.GetNamespace(),
		resourceType,
		resource.GetName()))

	resource.SetResourceVersion(existingResource.GetResourceVersion())

	err := k8sClient.Update(ctx, resource, &updateOptions)
	if err != nil {
		return false, generateUpsertIssue(err, resource, log, rg)
	}

	return existingResource.GetGeneration() != resource.GetGeneration(), nil
}

func createResource(ctx context.Context, k8sClient client.Client, resource client.Object,
	createOptions client.CreateOptions, log logr.Logger, resourceType string, rg issue.ResourceGrouper) (bool, *issue.Issue) {

	log.Info(fmt.Sprintf("creating resource [%s/%s/%s]",
		resource.GetNamespace(),
		resourceType,
		resource.GetName()))

	err := k8sClient.Create(ctx, resource, &createOptions)
	if err != nil {
		recordFailCreateEventIfNecessary(ctx, resource, resourceType, err)
		return false, generateUpsertIssue(err, resource, log, rg)
	}

	recordSuccessCreateEventIfNecessary(ctx, resource, resourceType)
	return true, nil
}

func generateUpsertIssue(err error, resource client.Object, log logr.Logger, rg issue.ResourceGrouper) *issue.Issue {
	isCritical := true
	if !isCriticalError(err, resource) {
		isCritical = false
		log.V(4).Error(err, "upsert resource encountered non-critical error", "resourceName", resource.GetName())
	}

	return issue.New(resource, err, isCritical, "", rg)
}

func isCriticalError(err error, resource client.Object) bool {
	var (
		code    int32
		reason  metav1.StatusReason
		details *metav1.StatusDetails
	)

	if statusErr, ok := err.(apierrors.APIStatus); ok || errors.As(err, &statusErr) {
		code = statusErr.Status().Code
		reason = statusErr.Status().Reason
		details = statusErr.Status().Details
	}

	rg := issue2.ComputeResourceGrouper{}
	issueGroup := rg.GetIssueGroupByResource(resource)
	// currently, we don't want to generate critical compute issues that are not strictly job failures.
	// these critical compute issues are causing workload failure, and we want to have high confidence when failing
	if issueGroup == issue.Compute {
		return isComputeError(details)
	}

	// handle specific case of k8s conflict where object has already been modified - not a critical conflict
	if reason == metav1.StatusReasonConflict && strings.Contains(err.Error(), objectHasBeenModifiedMsg) {
		return false
	}

	// the error is critical if it's a client error
	return isClientError(code)
}

func isComputeError(details *metav1.StatusDetails) bool {
	if details == nil {
		return false
	}
	for _, cause := range details.Causes {
		// failed calling webhook will fail the workload even in case of compute resource
		if strings.Contains(cause.Message, CauseFailedResourceRequestNotFound) {
			return true
		}
	}
	return false
}

func isClientError(statusCode int32) bool {
	// the error is a client error if code is 4xx (besides 429)
	return statusCode >= clientErrorHttpCodeMin && statusCode < clientErrorHttpCodeMax && statusCode != http.StatusTooManyRequests
}

func recordSuccessCreateEventIfNecessary(ctx context.Context, resource client.Object, resourceType string) {
	if !shouldRecordEventForKind(resourceType) {
		return
	}

	recorder := events.FromContext(ctx)
	recorder.Eventf(corev1.EventTypeNormal, CreateSucceededReason, "Successfully created %s %s", resourceType, resource.GetName())
}

func recordFailCreateEventIfNecessary(ctx context.Context, resource client.Object, resourceType string, err error) {
	if !shouldRecordEventForKind(resourceType) {
		return
	}

	recorder := events.FromContext(ctx)
	recorder.Eventf(corev1.EventTypeWarning, CreateFailedReason, "Failed to create %s %s. Error: %v", resourceType, resource.GetName(), err)
}

func shouldRecordEventForKind(kind string) bool {
	switch kind {
	case constants.KindComputeDomain:
		return true
	default:
		return false
	}
}
