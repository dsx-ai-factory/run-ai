/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	"time"

	workloadissue "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/issue"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"
	apijobs "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	jobInterface "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/modules"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/events"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaijob "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	builderCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	workloadCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/finalizers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/managers"
	statusaggregator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/status/aggregator"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/status/update"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/status"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

type ResourcesManager interface {
	// GetAllResources get all resource
	GetAllResources(ctx context.Context, aggr *resource.Aggregator, resources []common.Resource) ([]client.Object, error)

	UpsertResources(ctx context.Context, workload *apijobs.JobWorkload, workloadSubResources []client.Object, options managers.ResourceOptions) []*issue.Issue
}

type CleanupManager interface {
	DefaultCleanup(ctx context.Context, aggr *resource.Aggregator, fetchWorkloadSubResourcesFn managers.FetchWorkloadSubResourcesFn) error
	ForceFailureCleanup(ctx context.Context, aggr *resource.Aggregator, fetchWorkloadSubResourcesFn managers.FetchWorkloadSubResourcesFn) error
	SuspendedCleanup(ctx context.Context, aggr *resource.Aggregator, fetchWorkloadSubResourcesFn managers.FetchWorkloadSubResourcesFn) error
}

type JobController struct {
	StatusAggregator              statusaggregator.WorkloadStatusCalculater
	computeBuilder                builderCommon.ComputeBuilder
	commonResourcesBuilder        *builderCommon.SharedResourcesBuilder
	ResourcesManager              ResourcesManager
	statusUpdater                 *update.StatusUpdaterManager
	finalizer                     finalizers.RemoveFinalizer
	k8sClient                     client.Client // k8s client
	options                       managers.ResourceOptions
	eventRecorder                 record.EventRecorder
	subResourceRuntimeManager     *managers.SubResourcesRuntimeManager
	cleanupManager                CleanupManager
	enableDanglingResourceCleanup bool
}

func NewJobController(
	k8sClient client.Client,
	computeBuilder builderCommon.ComputeBuilder,
	commonResourcesBuilder *builderCommon.SharedResourcesBuilder,
	statusUpdater *update.StatusUpdaterManager,
	finalizer finalizers.RemoveFinalizer,
	statusAggregator statusaggregator.WorkloadStatusCalculater,
	resourceManager ResourcesManager,
	options managers.ResourceOptions,
	eventRecorder record.EventRecorder,
	cleanupManager CleanupManager,
	enableDanglingResourceCleanup bool,
) *JobController {
	return &JobController{
		k8sClient:                     k8sClient,
		computeBuilder:                computeBuilder,
		commonResourcesBuilder:        commonResourcesBuilder,
		statusUpdater:                 statusUpdater,
		options:                       options,
		ResourcesManager:              resourceManager,
		StatusAggregator:              statusAggregator,
		finalizer:                     finalizer,
		subResourceRuntimeManager:     managers.NewSubResourcesRuntimeManager(k8sClient),
		eventRecorder:                 eventRecorder,
		cleanupManager:                cleanupManager,
		enableDanglingResourceCleanup: enableDanglingResourceCleanup,
	}
}

func (jobController *JobController) Reconcile(ctx context.Context, _ ctrl.Request, workload *apijobs.JobWorkload, owner client.Object) (ctrl.Result, error) {
	logging := logger.FromContext(ctx)
	// deepcopy of the original workload status before modification
	originalWorkload := apijobs.JobWorkload{}
	workload.Status.DeepCopyInto(&originalWorkload.Status)

	aggr := resource.NewAggregator(workload)
	jobController.StatusAggregator.Clear()
	jobController.StatusAggregator.WithWorkload(workload)

	if workload.IsMarkedForDeletion() {
		// for backward compatibility, we need to remove the resources and finalizer of pre 2.13.0 version workloads
		if finalizers.HasWorkloadFinalizer(workload.ObjectMeta) {
			logging.Info("Workload is marked for deletion, removing workload Finalizer", "workload", workload.Name)
			err := finalizers.RemoveResources(ctx, workload.Status.WorkloadStatus, jobController.k8sClient)
			if err != nil {
				return ctrl.Result{}, nil
			}
			err = jobController.finalizer.RemoveFinalizer(ctx, workload)
			if err != nil {
				return jobController.FailedReconcileStatus(ctx, aggr, err)
			}
		}

		logging.Info("Workload is marked for deletion", "workload", workload.Name)
		// the resources and subresource will be deleted by k8s
		return ctrl.Result{}, nil
	}
	// training workloads does not support distributed mpi jobs
	if workload.TypeMeta.Kind == constants.KindTrainingWorkload && workload.Spec.Mpi.Deref() {
		logging.Info("Training workload does not support distributed mpi jobs anymore, skipping reconcile", "workload", workload.Name)
		// we don't have anything to do with workload used as template. k8s stores it in etcd, and we're done
		// support only delete mpi jobs
		return ctrl.Result{}, nil
	}

	// we don't have anything to do with workload used as template. k8s stores it in etcd and we're done
	if workload.Spec.Usage == common.UsedAsTemplate {
		logging.Info("Workload is used as template, skipping reconcile", "workload", workload.Name)
		// we don't have anything to do with workload used as template. k8s stores it in etcd and we're done
		return ctrl.Result{}, nil
	}

	// initialize the reconcile context or die
	ctx, err := jobController.initContext(ctx, workload, owner)
	if err != nil {
		// if it's the workload's first reconcile, fail it to prevent infinite fail loop
		if workload.IsFirstTime() {
			handleFirstAcknowledgement(ctx, workload)
			aggr.AddIssues(workloadissue.NewIssue(aggr.GetWorkload(), err, true, ""))
			jobController.StatusAggregator.CalculatePhaseAndConditions(ctx, aggr)

			failErr := jobController.FailWorkload(ctx, aggr, &originalWorkload, owner)
			if failErr != nil {
				logging.Error(failErr, fmt.Sprintf("failed to fail workload %s", workload.Name))
			}

			// no additional reconcile required
			return jobController.SuccessReconcile(ctx, aggr, &originalWorkload)
		}

		return jobController.FailedReconcileStatus(ctx, aggr, err)
	}

	if workload.IsCompleted() {
		err = jobController.updateStatusOfCompletedWorkload(ctx, aggr)
		if err != nil {
			return jobController.FailedReconcileStatus(ctx, aggr, err)
		}
		return jobController.SuccessReconcile(ctx, aggr, &originalWorkload)
	}

	if workload.Status.OwnedResources == nil {
		workload.Status.OwnedResources = make([]common.ObservedResource, 0)
	}

	if workload.Annotations == nil {
		workload.Annotations = make(map[string]string)
	}

	if workload.Labels == nil {
		workload.Labels = make(map[string]string)
	}

	if workload.IsFirstTime() {
		handleFirstAcknowledgement(ctx, workload)
	}

	jobAdapter, dangling, err := jobController.generateWorkloadStatus(ctx, aggr)
	if err != nil {
		return jobController.FailedReconcileStatus(ctx, aggr, err)
	}

	if workload.IsCompleted() {
		logging.Info("Workload is completed, skipping reconcile", "workload", workload.Name)

		cleanupErr := jobController.cleanupManager.DefaultCleanup(ctx, aggr, jobController.fetchWorkloadSubResources)
		if cleanupErr != nil {
			return jobController.FailedReconcileStatus(ctx, aggr, cleanupErr)
		}

		return jobController.SuccessReconcile(ctx, aggr, &originalWorkload)
	}

	// if the workload is fully suspended, and the desired phase is Suspended - we can stop the reconcile loop and perform cleanup for suspended phase
	if statusaggregator.IsJobSuspended(jobAdapter) && workload.Spec.Active.Deref() == false {
		logging.Info("Workload is fully suspended, skipping reconcile", "workload", workload.Name)
		cleanupErr := jobController.cleanupManager.SuspendedCleanup(ctx, aggr, jobController.fetchWorkloadSubResources)
		if cleanupErr != nil {
			return jobController.FailedReconcileStatus(ctx, aggr, cleanupErr)
		}

		return jobController.SuccessReconcile(ctx, aggr, &originalWorkload)
	}

	if len(dangling) != 0 && jobController.enableDanglingResourceCleanup {
		// TODO add metric increment to remove functionality in the future.
		err = modules.ClearDanglingResources(ctx, jobController.k8sClient, dangling)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	logging.Info(fmt.Sprintf("Workload %s has changed, rebulding workload", workload.Name))
	rebuildErr := jobController.rebuildWorkload(ctx, aggr, owner, jobAdapter)

	// make sure that all issues are taken into account, including runtime
	var reconcileErr error
	if rebuildErr != nil {
		reconcileErr = rebuildErr
	} else if aggr.HasIssues() {
		reconcileErr = aggr.GetIssuesList()[0]
	}

	if reconcileErr != nil {
		updatedStatus := jobController.StatusAggregator.CalculatePhaseAndConditions(ctx, aggr)
		if updatedStatus.WorkloadCommonStatus.Phase == common.PhaseFailed.String() {
			logging.Info(fmt.Sprintf("failed to rebuild workload %s", workload.Name))
			failErr := jobController.FailWorkload(ctx, aggr, &originalWorkload, owner)
			if failErr != nil {
				logging.Error(failErr, fmt.Sprintf("failed to fail workload %s", workload.Name))
			}
		}

		// this also returns the original error to trigger requeue
		return jobController.FailedReconcileStatus(ctx, aggr, reconcileErr)
	}

	_, err = jobController.SuccessReconcile(ctx, aggr, &originalWorkload)
	if err != nil {
		return ctrl.Result{}, err
	}
	if requeue, delay := shouldRequeue(workload); requeue {
		return ctrl.Result{RequeueAfter: delay}, nil
	}
	return ctrl.Result{}, nil
}

func shouldRequeue(workload *apijobs.JobWorkload) (bool, time.Duration) {
	conditions := workload.Status.Conditions
	for _, condition := range conditions {
		if condition.Type == common.Suspended && condition.Status == v1.ConditionFalse && condition.Reason == common.PodsStillRunningReason {
			return true, statusaggregator.GetSuspendingDelay()
		}
	}

	return false, statusaggregator.GetDefaultDelay()
}

// initContext prepares the context to the reconcile loop for the current workload
func (jobController *JobController) initContext(ctx context.Context, workload *apijobs.JobWorkload, owner client.Object) (context.Context, error) {
	ctx, err := workloadproject.SetProjectInContextFromNamespace(ctx, jobController.k8sClient, workload.Namespace, workload.Name, workload.Spec.Usage)
	if err != nil {
		return ctx, err
	}

	mnnvlStatus, err := mnnvlutils.GenerateStatus(ctx, jobController.k8sClient, workload)
	if err != nil {
		return ctx, err
	}
	ctx = mnnvlutils.ToContext(ctx, mnnvlStatus)

	workloadRecorder := events.NewWorkloadRecorder(owner, jobController.eventRecorder)
	ctx = events.ToContext(ctx, workloadRecorder)

	return ctx, nil
}

func (jobController *JobController) updateStatusOfCompletedWorkload(ctx context.Context, aggr *resource.Aggregator) error {
	workload := aggr.GetWorkload()
	logging := logger.FromContext(ctx)
	logging.Info("Workload is completed, only updating workload status", "workload", workload.Name)
	_, _, err := jobController.generateWorkloadStatus(ctx, aggr)
	return err
}

func (jobController *JobController) generateWorkloadStatus(ctx context.Context, aggr *resource.Aggregator) (jobInterface.JobAdapter, []client.Object, error) {
	workload := aggr.GetWorkload()
	logging := logger.FromContext(ctx)

	logging.Info("Updating workload status", "workload", workload.Name)
	setWorkloadGeneration(ctx, workload)

	dangling, err := jobController.fetchWorkloadSubResources(ctx, aggr)
	if err != nil {
		return nil, nil, err
	}

	var jobAdapter jobInterface.JobAdapter
	job := aggr.GetJob()
	if job != nil {
		// get the job adapter
		jobAdapterFactory, factoryErr := jobs.GetJobAdapterFactory(job, jobController.k8sClient)
		if factoryErr != nil {
			return nil, nil, factoryErr
		}
		jobAdapter, err = jobAdapterFactory.CreateJobAdapter(job)
		if err != nil {
			return nil, nil, err
		}
	}

	logging.V(4).Info("Calculate workload status", "workload", workload.Name)

	// generate runtime issue list and add to the aggr
	runtimeIssues := jobController.subResourceRuntimeManager.CheckSubResources(ctx, aggr)
	aggr.AddIssues(runtimeIssues...)

	// aggregate the resources status
	err = jobController.aggregateResourcesStatus(ctx, jobAdapter, aggr)
	if err != nil {
		return nil, nil, err
	}
	return jobAdapter, dangling, nil
}

func (jobController *JobController) fetchWorkloadSubResources(ctx context.Context, aggr *resource.Aggregator) ([]client.Object, error) {
	workload := aggr.GetWorkload()
	logging := logger.FromContext(ctx)
	workloadSpecResources, err := jobController.getWorkloadSubResourcesPreview(ctx, workload)
	if err != nil {
		logging.Error(err, "Failed to generate workload sub resources preview", "workload", workload.Name)
		return nil, err
	}
	logging.V(4).Info("Retrieving workload sub resources from cluster", "workload", workload.Name)
	return jobController.ResourcesManager.GetAllResources(ctx, aggr, workloadSpecResources)
}

func (jobController *JobController) rebuildStalledResources(ctx context.Context, job jobInterface.JobAdapter, owner client.Object, aggr *resource.Aggregator) (ctrl.Result, error) {
	workload := aggr.GetWorkload()
	logging := logger.FromContext(ctx)
	logging.Info("Checking if there are stalled resources", "workload", workload.Name)
	resourcesMap := createSubresourceMap(job, nil, nil)
	var rebuildResources []common.ObservedResource
	rebuildResources = jobController.getChangedSubResource(resourcesMap, workload)
	rebuildResources = append(rebuildResources, jobController.getDeletedSubResource(resourcesMap, workload)...)

	if len(rebuildResources) > 0 {
		return ctrl.Result{}, jobController.recreateMismatchedResource(ctx, aggr, owner, rebuildResources, job)
	}
	return ctrl.Result{}, nil
}

func (jobController *JobController) getWorkloadSubResourcesPreview(ctx context.Context, workload *apijobs.JobWorkload) ([]common.Resource, error) {
	logR := logger.FromContext(ctx)
	logR.V(4).Info("generating workload sub resources", "workload", workload.Name)
	computeResources, err := jobController.computeBuilder.GetPreview(ctx, workload)
	if err != nil {
		return nil, fmt.Errorf("failed to get job resources: %v", err)
	}
	commonResources, err := jobController.commonResourcesBuilder.GetPreview(ctx, workload)
	if err != nil {
		return nil, fmt.Errorf("failed to get common resources: %v", err)
	}
	workloadSpecResources := append(computeResources, commonResources...)
	jobController.StatusAggregator.WithExpectedResources(workloadSpecResources)
	return workloadSpecResources, nil
}

// buildResources builds the resources of the workload
func (jobController *JobController) buildResourceStructs(ctx context.Context, aggr *resource.Aggregator, owner client.Object, existingJob jobInterface.JobAdapter) error {
	workload := aggr.GetWorkload()
	options := builderCommon.ResourceBuilderOptions{Owner: owner}

	// get services
	services, err := jobController.commonResourcesBuilder.BuildServices(ctx, workload, options)
	if err != nil {
		return err
	}

	var originalObject client.Object
	if existingJob != nil {
		originalObject = existingJob.GetOriginalObject()
	}

	// get external services
	externalsServices, err := jobController.commonResourcesBuilder.BuildExternalUrls(ctx, workload, &builderCommon.ExternalBuilderOptions{
		Owner:       owner,
		DisableAuth: ptr.To(!builderCommon.IsJobSupportsAuth(originalObject)),
	})
	if err != nil {
		return err
	}

	// build job
	computeResources, err := jobController.buildCompute(ctx, workload, builderCommon.ComputeBuilderOptions{Owner: owner}, existingJob)
	if err != nil {
		return err
	}

	// will be deleted in the future
	// pvc does not owned by the workload
	pvcList, err := jobController.commonResourcesBuilder.BuildPVC(ctx, workload, options)
	if err != nil {
		return err
	}

	// set workload resource with the resource list
	for _, service := range services {
		aggr.AddDesired(service)
	}
	aggr.AddDesired(externalsServices...)
	aggr.AddDesired(pvcList...)
	aggr.AddDesired(computeResources.GetAll()...)
	return nil
}

func (jobController *JobController) aggregateResourcesStatus(ctx context.Context, job jobInterface.JobAdapter, aggr *resource.Aggregator) error {
	jobController.StatusAggregator.Clear()
	jobController.StatusAggregator.WithJob(job)
	jobController.StatusAggregator.WithServices(aggr.GetServices())
	jobController.StatusAggregator.WithResources(aggr.GetExternalNetwork())
	jobController.StatusAggregator.WithComputeDomain(aggr.GetComputeDomain())
	jobController.StatusAggregator.WithWorkload(aggr.GetWorkload())
	jobController.StatusAggregator.WithSecrets(aggr.GetSecrets())
	workloadStatus := jobController.StatusAggregator.CalculateAll(ctx, aggr)
	workloadStatus.DeepCopyInto(&aggr.GetWorkload().Status)
	return nil
}

func (jobController *JobController) recreateMismatchedResource(ctx context.Context, aggr *resource.Aggregator, owner client.Object, _ []common.ObservedResource, existingJob jobInterface.JobAdapter) error {
	// currently we only support rebuild the entire workload not just the mismatched resources
	err := jobController.rebuildWorkload(ctx, aggr, owner, existingJob)
	if err != nil {
		return err
	}
	return nil
}

func (jobController *JobController) isWorkloadChanged() bool {
	// currently we only support rebuild the entire workload not just the mismatched resources
	// so, we always return true
	return true
}

// getDeletedSubResource returns a map of sub resources that were deleted
func (jobController *JobController) getDeletedSubResource(subResources map[string]client.Object, workload *apijobs.JobWorkload) []common.ObservedResource {
	var deletedResource []common.ObservedResource
	for _, ownedResource := range workload.Status.OwnedResources {
		if _, ok := subResources[ownedResource.String()]; !ok {
			deletedResource = append(deletedResource, ownedResource)
		}
	}
	return deletedResource
}

func (jobController *JobController) getChangedSubResource(subResources map[string]client.Object, workload *apijobs.JobWorkload) []common.ObservedResource {
	var changedResource []common.ObservedResource
	for _, ownedResource := range workload.Status.OwnedResources {
		if subResource, ok := subResources[ownedResource.String()]; !ok {
			if ownedResource.IsChanged(subResource.GetGeneration()) {
				changedResource = append(changedResource, ownedResource)
			}
		}
	}
	return changedResource
}

func (jobController *JobController) updateWorkloadStatus(ctx context.Context, aggr *resource.Aggregator, originalWorkload *apijobs.JobWorkload, err error) error {
	workload := aggr.GetWorkload()
	setLastReconcileTimeCondition(workload, err)
	setIssueStatus(aggr)
	return jobController.statusUpdater.Update(ctx, workload, originalWorkload)
}

func (jobController *JobController) SuccessReconcile(ctx context.Context, aggr *resource.Aggregator, originalWorkload *apijobs.JobWorkload) (ctrl.Result, error) {
	return ctrl.Result{}, jobController.updateWorkloadStatus(ctx, aggr, originalWorkload, nil)
}

func (jobController *JobController) FailedReconcileStatus(ctx context.Context, aggr *resource.Aggregator, err error) (ctrl.Result, error) {
	logging := logger.FromContext(ctx)
	logging.Error(err, "workload reconcile failed with error")
	updateError := jobController.updateWorkloadStatus(ctx, aggr, nil, err)
	if updateError != nil {
		logging.Error(updateError, "failed to update status")
	}
	return ctrl.Result{}, err
}

func setLastReconcileTimeCondition(workload *apijobs.JobWorkload, err error) {
	conditionType := common.LastReconcileSucceeded
	conditionStatus := v1.ConditionTrue
	conditionReason := ""
	conditionMessage := ""

	if err != nil {
		conditionStatus = v1.ConditionFalse
		conditionReason = common.FailedReconcileReason
		conditionMessage = err.Error()
	}

	status.UpsertCondition(&workload.Status.Conditions, conditionType, conditionStatus, conditionReason, conditionMessage)
}

func setIssueStatus(aggr *resource.Aggregator) {
	workload := aggr.GetWorkload()

	statusPhase, err := common.PhaseStringToEnum(workload.Status.WorkloadCommonStatus.Phase)
	if err != nil {
		return
	}

	// when workload is in finite state or in terminating state, we do not want to modify the issues state
	if statusPhase == common.PhaseCompleted || statusPhase == common.PhaseTerminating {
		return
	}

	issues := aggr.GetIssueStatus()
	// if workload is set to fail, we have issues that cause the failure, we want to keep them.
	if statusPhase == common.PhaseFailed && issues == nil {
		return
	}
	workload.Status.Issues = issues
}

func (jobController *JobController) rebuildWorkload(ctx context.Context, aggr *resource.Aggregator, owner client.Object, existingJob jobInterface.JobAdapter) error {
	workload := aggr.GetWorkload()
	logging := logger.FromContext(ctx)
	logging.Info("rebuilding workload", "workload", workload.Name)
	// generate list of resources that should exist in workload.
	err := jobController.buildResourceStructs(ctx, aggr, owner, existingJob)
	if err != nil {
		aggr.AddIssues(workloadissue.NewIssue(workload, err, true, ""))
		return err
	}

	// create or update resource based on the generated list.
	err = jobController.submitResources(ctx, aggr)
	if err != nil {
		return err
	}
	return nil

}

// buildCompute builds the job based on the job builder supllied at controller creation
func (jobController *JobController) buildCompute(ctx context.Context, workload *apijobs.JobWorkload, options builderCommon.ComputeBuilderOptions, existingJob jobInterface.JobAdapter) (*builderCommon.ComputeResources, error) {
	existingPodSpec := jobController.getPodsSpec(existingJob)
	options.ExitingPodsSpec = existingPodSpec

	computeResources, err := jobController.computeBuilder.Build(ctx, workload, &options)
	if err != nil {
		return nil, err
	}

	return computeResources, nil
}

func (jobController *JobController) getPodsSpec(existingJob jobInterface.JobAdapter) map[workloadCommon.ReplicaType]*v1.PodSpec {
	existingPodsSpec := make(map[workloadCommon.ReplicaType]*v1.PodSpec)
	if existingJob != nil {
		for replicaType, replicaSpec := range existingJob.GetSpec().GetReplicaTypeSpecs() {
			if replicaSpec.GetPodTemplateSpec() != nil {
				existingPodsSpec[replicaType] = &replicaSpec.GetPodTemplateSpec().Spec
			} else {
				existingPodsSpec[replicaType] = nil
			}
		}
	}
	// if the node affinity is empty, we need to set it to nil
	// this will help distinguish between empty node affinity and no node affinity
	if len(existingPodsSpec) == 0 {
		existingPodsSpec = nil
	}
	return existingPodsSpec
}

func createSubresourceMap(job jobInterface.JobAdapter, service []v1.Service, ingressOrRoute []client.Object) map[string]client.Object {
	resourceMap := make(map[string]client.Object)
	resourceMap[common.ObservedResourceFromObject(job).String()] = job
	for _, svc := range service {
		resourceMap[common.ObservedResourceFromObject(&svc).String()] = &svc
	}
	for _, route := range ingressOrRoute {
		resourceMap[common.ObservedResourceFromObject(route).String()] = route
	}
	return resourceMap
}

// Submits the resources to the cluster
func (jobController *JobController) submitResources(ctx context.Context, aggr *resource.Aggregator) error {
	workload := aggr.GetWorkload()
	logging := logger.FromContext(ctx)
	logging.Info("submitting resources to kubernetes  api server", "workload", workload.Name)

	//
	// create or update all resources except runaiJob
	//
	var runaiJobResource *runaijob.RunaiJob
	var OtherResources []client.Object

	for _, obj := range aggr.GetDesiredList() {
		if runaiJob, ok := obj.(*runaijob.RunaiJob); ok {
			runaiJobResource = runaiJob
		} else {
			OtherResources = append(OtherResources, obj)
		}
	}

	issues := jobController.ResourcesManager.UpsertResources(ctx, workload, OtherResources, jobController.options)
	if issues != nil {
		aggr.AddIssues(issues...)
		return fmt.Errorf("upsert workload resources error")
	}

	if runaiJobResource != nil {
		err := builderCommon.UpdateJobUrlWithNodePorts(ctx, OtherResources, jobController.k8sClient, &runaiJobResource.ObjectMeta)
		if err != nil {
			aggr.AddIssues(workloadissue.NewIssue(runaiJobResource, err, false, ""))
			return err
		}
		issues = jobController.ResourcesManager.UpsertResources(ctx, workload, []client.Object{runaiJobResource}, jobController.options)
		if issues != nil {
			aggr.AddIssues(issues...)
			return fmt.Errorf("workload job upsert failed")
		}
	}

	return nil
}

func (jobController *JobController) SendFailureEvent(ctx context.Context, phaseMessage string, owner client.Object) {
	if jobController.eventRecorder == nil {
		logging := logger.FromContext(ctx)
		logging.V(2).Info("could not emit failure event - recorder is nil")
		return
	}

	jobController.eventRecorder.Eventf(owner, v1.EventTypeNormal, common.WorkloadFailedReason, phaseMessage)
}

func (jobController *JobController) FailWorkload(ctx context.Context, aggr *resource.Aggregator, originalWorkload *apijobs.JobWorkload,
	owner client.Object) error {
	workload := aggr.GetWorkload()
	cleanupErr := jobController.cleanupManager.ForceFailureCleanup(ctx, aggr, jobController.fetchWorkloadSubResources)
	if cleanupErr != nil {
		logging := logger.FromContext(ctx)
		msg := fmt.Sprintf("failed to cleanup workload resources. error message: %s", cleanupErr.Error())
		logging.V(2).Error(cleanupErr, msg, "workload", workload.Name)

		jobController.eventRecorder.Eventf(owner, v1.EventTypeWarning, common.CleanupFailedReason, msg)

		aggr.AddIssues(workloadissue.NewIssue(workload, cleanupErr, false, msg))
	}

	_, err := jobController.setFailedWorkloadStatus(ctx, aggr, originalWorkload)
	if err != nil {
		return err
	}

	jobController.SendFailureEvent(ctx, workload.Status.WorkloadCommonStatus.PhaseMessage, owner)

	return cleanupErr
}

func (jobController *JobController) setFailedWorkloadStatus(ctx context.Context, aggr *resource.Aggregator, originalWorkload *apijobs.JobWorkload) (ctrl.Result, error) {
	logging := logger.FromContext(ctx)
	workload := aggr.GetWorkload()

	workload.Status.CreationStatus = common.StatusFailed
	workload.Status.SetCompletionTime(metav1.Now())

	err := jobController.updateWorkloadStatus(ctx, aggr, originalWorkload, nil)
	if err != nil {
		logging.Error(err, "failed to update status", "error", err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func setWorkloadGeneration(ctx context.Context, workload *apijobs.JobWorkload) {
	logging := logger.FromContext(ctx)
	logging.V(4).Info("Setting workload observed generation status", "workload", workload.Name)
	workload.Status.ObservedGeneration = workload.Generation
}

func handleFirstAcknowledgement(ctx context.Context, workload *apijobs.JobWorkload) {
	logging := logger.FromContext(ctx)
	logging.Info("First time the controller is aware of this job", "workload", workload.Name)
	// Set the first acknowledged time
	// this the first time the controller is aware of this job
	workload.Status.SetFirstAcknowledgedTime(time.Now())
	// for backward compatibility, we need to set the created status as well
	// remove this in the next release
	workload.Status.CreationStatus = common.StatusSubmitted
	workload.Status.WorkloadCommonStatus.Phase = common.PhaseCreating.String()
}
