/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managers

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/go-logr/logr"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/convertors"
	resourceobject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/status/resource"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	CleaningResourcesMessage          = "Cleaning resources"
	CleaningResourcesOnSuspendMessage = "Cleaning resources on suspend"

	CleanupFailedReason    = "CleanupFailed"
	CleanupSucceededReason = "CleanupSuccess"
)

type FetchWorkloadSubResourcesFn func(ctx context.Context, aggr *resource.Aggregator) ([]client.Object, error)
type FailureCleanupPolicy string

const (
	KeepFailing FailureCleanupPolicy = "KeepFailing"
	All         FailureCleanupPolicy = "All"
	None        FailureCleanupPolicy = "None"
	Undefined   FailureCleanupPolicy = ""
)

var policies = []string{KeepFailing.String(), All.String(), None.String()}

func (cm FailureCleanupPolicy) String() string {
	return string(cm)
}

func ParseFailureCleanupPolicy(policyStr string) FailureCleanupPolicy {
	if !slices.Contains(policies, policyStr) {
		return Undefined
	}

	return FailureCleanupPolicy(policyStr)
}

func GetDefaultFailureCleanupPolicy() FailureCleanupPolicy {
	return All
}

type deletionPredicate func(resource client.Object) bool

type CleanupManager struct {
	policy                 FailureCleanupPolicy
	k8sClient              client.Client
	eventRecorder          record.EventRecorder
	jobToWorkloadConverter *convertors.GeneralWorkloadConverter
}

func NewCleanupManager(k8sClient client.Client, cleanupPolicy FailureCleanupPolicy, eventRecorder record.EventRecorder) *CleanupManager {
	return &CleanupManager{
		policy:                 cleanupPolicy,
		k8sClient:              k8sClient,
		eventRecorder:          eventRecorder,
		jobToWorkloadConverter: convertors.NewGeneralWorkloadConverter(),
	}
}

// DefaultCleanup is for graceful workload cleanups, aka finite state (does not delete the job)
func (fm *CleanupManager) DefaultCleanup(ctx context.Context, aggr *resource.Aggregator, fetchWorkloadSubResourcesFn FetchWorkloadSubResourcesFn) error {
	predicate := fm.getDefaultPredicateFn()
	return fm.cleanup(ctx, aggr, fetchWorkloadSubResourcesFn, predicate, false)
}

// ForceFailureCleanup is for cases where the workload is failed by the workload, aka when there are issues (does delete the job). It also takes the FailureCleanupPolicy into account
func (fm *CleanupManager) ForceFailureCleanup(ctx context.Context, aggr *resource.Aggregator, fetchWorkloadSubResourcesFn FetchWorkloadSubResourcesFn) error {
	predicate := fm.getCleanupPolicyPredicateFn(aggr)
	return fm.cleanup(ctx, aggr, fetchWorkloadSubResourcesFn, predicate, true)
}

// SuspendedCleanup is for cases where the workload suspended
func (fm *CleanupManager) SuspendedCleanup(ctx context.Context, aggr *resource.Aggregator, fetchWorkloadSubResourcesFn FetchWorkloadSubResourcesFn) error {
	workload := aggr.GetWorkload()
	log := logger.FromContext(ctx)
	log.Info(CleaningResourcesOnSuspendMessage, "workload", workload.Name)

	workloadObject := fm.jobToWorkloadConverter.Convert(workload)
	if workloadObject == nil {
		return fmt.Errorf("failed to perform cleanup: unknown workload kind %s", workload.Kind)
	}

	_, err := fetchWorkloadSubResourcesFn(ctx, aggr)
	if err != nil {
		return err
	}
	log.V(4).Info("Fetched resources to cleanup", "workload", workload.Name)

	computeDomain := aggr.GetComputeDomain()
	if computeDomain == nil {
		return nil
	}
	if !computeDomain.DeletionTimestamp.IsZero() {
		log.V(4).Info(fmt.Sprintf("Already deleted %s", constants.KindComputeDomain), "workload", workload.Name)
		return nil
	}

	deletedObjects := sync.Map{}
	err = deleteObject(ctx, computeDomain, fm.k8sClient, &deletedObjects)
	if err != nil {
		fm.recordFailEvent(computeDomain.Kind, computeDomain.Name, err, workloadObject)
		return err
	}

	fm.recordSuccessEvent(computeDomain.Kind, computeDomain.Name, workloadObject)

	updateOwnedResources(workload, &deletedObjects, log)

	return nil
}

func (fm *CleanupManager) cleanup(ctx context.Context, aggr *resource.Aggregator, fetchWorkloadSubResourcesFn FetchWorkloadSubResourcesFn, predicate deletionPredicate, deleteJob bool) error {
	workload := aggr.GetWorkload()
	log := logger.FromContext(ctx)
	log.Info(CleaningResourcesMessage, "workload", workload.Name)

	_, err := fetchWorkloadSubResourcesFn(ctx, aggr)
	if err != nil {
		return err
	}
	log.V(4).Info("Fetched resources to cleanup", "workload", workload.Name)

	services := aggr.GetServices()
	en := aggr.GetExternalNetwork()
	pvcs := aggr.GetPvcs() // aggr only contains pvcs that are owned by the workload
	job := aggr.GetJob()
	computeDomain := aggr.GetComputeDomain()
	secrets := aggr.GetSecrets()

	var wg sync.WaitGroup
	errChan := make(chan error, len(services)+len(en)+len(pvcs)+1+1+len(secrets))
	deletedObjects := sync.Map{}

	deleteObjectsInParallel(ctx, predicate, services, fm.k8sClient, &deletedObjects, &wg, errChan)
	deleteObjectsInParallel(ctx, predicate, en, fm.k8sClient, &deletedObjects, &wg, errChan)
	deleteObjectsInParallel(ctx, predicate, pvcs, fm.k8sClient, &deletedObjects, &wg, errChan)
	deleteObjectsInParallel(ctx, predicate, secrets, fm.k8sClient, &deletedObjects, &wg, errChan)
	if deleteJob && job != nil {
		deleteObjectsInParallel(ctx, predicate, []client.Object{job}, fm.k8sClient, &deletedObjects, &wg, errChan)
	}
	if computeDomain != nil {
		deleteObjectsInParallel(ctx, predicate, []client.Object{computeDomain}, fm.k8sClient, &deletedObjects, &wg, errChan)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for e := range errChan {
		errs = append(errs, e)
	}
	deleteErr := errors.Join(errs...) // equals nil if there were no errors

	updateOwnedResources(workload, &deletedObjects, log)

	return deleteErr
}

// getDefaultPredicateFn returns true for every resource
func (fm *CleanupManager) getDefaultPredicateFn() deletionPredicate {
	return func(resource client.Object) bool {
		return true
	}
}

// getCleanupPolicyPredicateFn filters resource to be deleted by the current FailureCleanupPolicy
func (fm *CleanupManager) getCleanupPolicyPredicateFn(aggr *resource.Aggregator) deletionPredicate {
	return func(resource client.Object) bool {
		switch fm.policy {
		case KeepFailing:
			iss := aggr.GetIssue(resource.GetObjectKind().GroupVersionKind().Kind, resource.GetName())
			return iss == nil
		case None:
			return false
		default: // All
			return true
		}
	}
}

func updateOwnedResources(workload *jobs.JobWorkload, deletedObjects *sync.Map, log logr.Logger) {
	log.V(4).Info("Updating owned resources", "workload", workload.Name)
	var ownedResourcesTmp []common.ObservedResource
	for _, owned := range workload.Status.OwnedResources {
		// remove the deleted object only if it's identical to the one stored in OwnedResources
		if deletedObjAny, deleted := deletedObjects.Load(owned.Name); deleted {
			if deletedObj, ok := deletedObjAny.(client.Object); !ok {
				if owned.Kind.Kind == deletedObj.GetObjectKind().GroupVersionKind().Kind &&
					owned.Namespace == deletedObj.GetNamespace() {
					continue
				}
			}
		} else {
			ownedResourcesTmp = append(ownedResourcesTmp, owned)
		}
	}
	workload.Status.OwnedResources = ownedResourcesTmp
}

func deleteObjectsInParallel[T client.Object](ctx context.Context, predicate deletionPredicate, objects []T, k8sClient client.Client, deletedObjects *sync.Map, wg *sync.WaitGroup, errChan chan<- error) {
	log := logger.FromContext(ctx)

	for _, obj := range objects {
		if predicate(obj) {
			wg.Add(1)
			deleteObjectInParallel(ctx, obj, k8sClient, deletedObjects, wg, errChan)
		} else {
			log.V(4).Info("skipped resource deletion", "resourceName", obj.GetName())
		}
	}
}

func deleteObjectInParallel(ctx context.Context, obj client.Object, k8sClient client.Client, deletedObjects *sync.Map, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	err := deleteObject(ctx, obj, k8sClient, deletedObjects)
	if err != nil {
		errChan <- err
		return
	}
}

func deleteObject(ctx context.Context, obj client.Object, k8sClient client.Client, deletedObjects *sync.Map) error {
	log := logger.FromContext(ctx)

	gvk, err := resourceobject.GetObjectResource(obj)
	if err != nil {
		return err
	}

	err = k8sClient.Delete(ctx, obj)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Error(err, "resource does not exist - skipping", "resourceName", obj.GetName(), "kind", gvk.Kind)
			return nil
		}

		log.Error(err, "failed to delete resource", "resourceName", obj.GetName(), "kind", gvk.Kind)
		return err
	}

	deletedObjects.Store(obj.GetName(), obj)

	log.V(4).Info("resource deleted successfully", "resourceName", obj.GetName(), "kind", gvk.Kind)
	return nil
}

func (fm *CleanupManager) recordSuccessEvent(kind, name string, owner client.Object) {
	fm.eventRecorder.Eventf(owner, corev1.EventTypeNormal, CleanupSucceededReason, "Successfully deleted %s %s", kind, name)
}

func (fm *CleanupManager) recordFailEvent(kind, name string, err error, owner client.Object) {
	fm.eventRecorder.Eventf(owner, corev1.EventTypeWarning, CleanupFailedReason, "Failed to delete %s %s. Error: %v", kind, name, err)
}
