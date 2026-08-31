/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/getters"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller"
	resourcemanager "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/resources"
	inferenceaggregator "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/aggregator"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/issue"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/utils/spec_checksum"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	workloadissue "github.com/run-ai/runai/runai-cluster/cluster/workload-common/issue"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	builderCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	workloadcontroller "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/controller"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/finalizers"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/managers"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/status/aggregator"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/status/update"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

type Controller struct {
	StatusAggregator       aggregator.WorkloadStatusCalculater
	knativeBuilder         builderCommon.ComputeBuilder
	commonResourcesBuilder *builderCommon.SharedResourcesBuilder
	resourcesManager       workloadcontroller.ResourcesManager
	statusUpdater          *update.StatusUpdaterManager
	finalizer              finalizers.RemoveFinalizer
	k8sClient              client.Client // k8s client
	options                managers.ResourceOptions
	knativeService         *knv1.Service
}

func NewController(
	k8sClient client.Client,
	knativeBuilder builderCommon.ComputeBuilder,
	commonResourcesBuilder *builderCommon.SharedResourcesBuilder,
	statusUpdater *update.StatusUpdaterManager,
	finalizer finalizers.RemoveFinalizer,
	statusAggregator inferenceaggregator.InferenceWorkloadStatusCalculater,
	resourceManager workloadcontroller.ResourcesManager,
	options managers.ResourceOptions,
) *Controller {
	InferenceController := Controller{
		knativeBuilder:         knativeBuilder,
		k8sClient:              k8sClient,
		commonResourcesBuilder: commonResourcesBuilder,
		statusUpdater:          statusUpdater,
		options:                options,
		resourcesManager:       resourceManager,
		StatusAggregator:       statusAggregator,
		finalizer:              finalizer,
		knativeService:         &knv1.Service{},
	}
	return &InferenceController
}

func (iwc *Controller) Reconcile(ctx context.Context, workload *jobs.JobWorkload, owner client.Object) (ctrl.Result, error) {
	// init logger
	var log = logger.Log.WithName("reconcile.infw")
	logger.IntoContext(ctx, log)

	// deepcopy of the original workload status before modification
	originalWorkload := jobs.JobWorkload{}
	workload.Status.DeepCopyInto(&originalWorkload.Status)
	aggr := resource.NewAggregator(workload)

	if workload.IsMarkedForDeletion() {
		log.Info("Workload is marked for deletion", "workload", workload.Name)

		workload.Status.InferenceStatus.Phase = inferenceaggregator.Deleting

		err := iwc.statusUpdater.Update(ctx, workload, &originalWorkload)
		if err != nil {
			return ctrl.Result{}, err
		}
		// the resources and subresource will be deleted by k8s
		return ctrl.Result{}, nil
	}

	ctx, err := workloadproject.SetProjectInContextFromNamespace(ctx, iwc.k8sClient, workload.Namespace, workload.Name, workload.Spec.Usage)
	if err != nil {
		return iwc.FailedReconcileStatus(ctx, aggr, issue.NewIssue(workload, err), true)
	}

	// initialize workload fields (annotations, labels, owned resources)
	iwc.initializeWorkloadFields(workload)

	if workload.IsFirstTime() {
		iwc.handleFirstAcknowledgement(ctx, workload)
	}

	_, err = iwc.getWorkloadResources(ctx, aggr)
	if err != nil {
		return iwc.FailedReconcileStatus(ctx, aggr, issue.NewIssue(workload, err), true)
	}

	err = iwc.updateWorkloadStatus(ctx, aggr, nil)
	if err != nil {
		return iwc.FailedReconcileStatus(ctx, aggr, issue.NewIssue(workload, err), true)
	}

	shouldRebuildKnativeService, newSpecChecksum, err := iwc.shouldRebuildKnativeService(ctx, workload)
	if err != nil {
		return iwc.FailedReconcileStatus(ctx, aggr, issue.NewIssue(workload, err), true)
	}

	var knativeService client.Object
	if shouldRebuildKnativeService {
		log.Info(fmt.Sprintf("Workload %s has changed, rebuilding knative service", workload.Name))

		var issues []*workloadissue.Issue
		var toRequeue bool
		knativeService, issues, toRequeue = iwc.rebuildWorkloadKnativeService(ctx, workload, owner)
		if issues != nil {
			return iwc.FailedReconcileStatus(ctx, aggr, issues[0], toRequeue)
		}

		isChecksumModified, err := iwc.setSpecChecksumAnnotation(ctx, workload, newSpecChecksum)
		if err != nil {
			issueObject := issue.NewIssue(workload, err)
			return iwc.FailedReconcileStatus(ctx, aggr, issueObject, true)
		}

		isClusterVersionLabelModified := iwc.addClusterVersionLabel(workload)

		if isChecksumModified || isClusterVersionLabelModified {
			workloadForPatch := infw.NewFromJobWorkload(workload)
			err := controller.PatchWorkloadLabelsAndAnnotations(ctx, iwc.k8sClient, workloadForPatch, log)
			if err != nil {
				issueObject := issue.NewIssue(workload, err)
				return iwc.FailedReconcileStatus(ctx, aggr, issueObject, true)
			}
		}
	} else {
		log.Info(fmt.Sprintf("Workload %s has not changed, not rebuilding knative service", workload.Name))

		knativeService, err = iwc.getKnativeService(ctx, workload.Namespace, workload.Name)
		if err != nil {
			return iwc.FailedReconcileStatus(ctx, aggr, issue.NewIssue(workload, err), true)
		}
		if knativeService.GetName() == "" {
			return iwc.FailedReconcileStatus(ctx, aggr, issue.NewIssue(workload,
				fmt.Errorf("knative service %s/%s not found", workload.Namespace, workload.Name)), true)
		}

		// update the owned resources with the knative service
		err = status.PatchOwnedResource(knativeService, &workload.Status.OwnedResources)
		if err != nil {
			return iwc.FailedReconcileStatus(ctx, aggr, issue.NewIssue(workload, err), true)
		}
	}

	// rebuild the other resources
	log.Info(fmt.Sprintf("Rebuilding workload %s resources", workload.Name))
	issues, isBuildError := iwc.rebuildWorkloadResources(ctx, aggr, owner, knativeService)

	if issues != nil {
		// if it is a build error, we do not want to requeue the reconcile
		return iwc.FailedReconcileStatus(ctx, aggr, issues[0], !isBuildError)
	}

	if err := resourcemanager.CleanupDisabledNetworkingResources(ctx, iwc.k8sClient, workload.Namespace,
		map[string]string{constants.LabelRelease: workload.Name}, workloadconfig.IsOpenshift()); err != nil {
		return iwc.FailedReconcileStatus(ctx, aggr, issue.NewIssue(workload, err), true)
	}

	_, err = iwc.SuccessReconcile(ctx, workload, &originalWorkload)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (iwc *Controller) addClusterVersionLabel(workload *jobs.JobWorkload) bool {

	labels := workload.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	clusterVersion := os.Getenv(workloadCommon.EnvVarClusterVersion)
	if clusterVersion == "" {
		return false
	}

	if labels[workloadCommon.CreatedAtClusterVersionLabel] == clusterVersion {
		return false
	}

	labels[workloadCommon.CreatedAtClusterVersionLabel] = clusterVersion
	workload.SetLabels(labels)

	return true
}

func (iwc *Controller) initializeWorkloadFields(workload *jobs.JobWorkload) {
	if workload.Status.OwnedResources == nil {
		workload.Status.OwnedResources = make([]common.ObservedResource, 0)
	}

	if workload.Annotations == nil {
		workload.Annotations = make(map[string]string)
	}

	if workload.Labels == nil {
		workload.Labels = make(map[string]string)
	}

	// Add NIM workload label if using NIM image
	if workload.Spec.Image != nil {
		controller.AddNimWorkloadLabel(workload.Spec.Labels, workload.Spec.Image.Deref())
	}
}

func (iwc *Controller) updateWorkloadStatus(ctx context.Context, aggr *resource.Aggregator, buildError *string) error {
	log := logger.FromContext(ctx)
	workload := aggr.GetWorkload()
	log.Info("Updating workload status", "workload", workload.Name)

	iwc.SetWorkloadGeneration(ctx, workload)

	knativeService, err := iwc.getKnativeService(ctx, workload.Namespace, workload.Name)
	if err != nil {
		return err
	}

	var lastCreatedRevisionPods []corev1.Pod
	if latestCreatedRevisionName := knativeService.Status.LatestCreatedRevisionName; latestCreatedRevisionName != "" {
		lastCreatedRevisionPods, err = getters.GetPodsOfRevision(ctx, latestCreatedRevisionName, workload.GetNamespace(), iwc.k8sClient)
		if err != nil {
			log.Error(err, "Failed to get pods of the workload", "workload", workload.Name)
			return err
		}
	}

	log.Info("Calculate workload status", "workload", workload.Name)
	err = iwc.aggregateResourcesStatus(ctx, aggr, knativeService, workload, lastCreatedRevisionPods, buildError)
	if err != nil {
		return err
	}

	return nil
}

func (iwc *Controller) SetWorkloadGeneration(ctx context.Context, workload *jobs.JobWorkload) {
	log := logger.FromContext(ctx)
	log.Info("Setting workload observed generation status", "workload", workload.Name)
	workload.Status.ObservedGeneration = workload.Generation
}

func (iwc *Controller) handleFirstAcknowledgement(ctx context.Context, workload *jobs.JobWorkload) {
	log := logger.FromContext(ctx)
	log.Info("First time the controller is aware of this job", "workload", workload.Name)
	// Set the first acknowledged time
	// this the first time the controller is aware of this job
	workload.Status.SetFirstAcknowledgedTime(time.Now())

	workload.Status.InferenceStatus.Phase = inferenceaggregator.Creating
}

func (iwc *Controller) getWorkloadSubResource(ctx context.Context, workload *jobs.JobWorkload) ([]common.Resource, error) {
	logR := logger.FromContext(ctx)
	logR.Info("generating workload sub resources", "workload", workload.Name)
	knativeResource, err := iwc.knativeBuilder.GetPreview(ctx, workload)
	if err != nil {
		return nil, fmt.Errorf("failed to get job resources: %v", err)
	}
	commonResources, err := iwc.commonResourcesBuilder.GetPreview(ctx, workload)
	if err != nil {
		return nil, fmt.Errorf("failed to get common resources: %v", err)
	}
	workloadSpecResources := append(knativeResource, commonResources...)
	// for now, we do not have any status condition for inference workload to check if the resources that was expected exists
	// iwc.StatusAggregator.WithExpectedResources(workloadSpecResources)

	return workloadSpecResources, nil
}

func (iwc *Controller) buildKnative(ctx context.Context, workload *jobs.JobWorkload, owner client.Object) (client.Object, error) {
	builderOptions := &builderCommon.ComputeBuilderOptions{
		Owner: owner,
	}

	// build knative service
	computeResources, err :=
		builders.
			NewKnativeBuilder(iwc.k8sClient).
			Build(ctx, workload, builderOptions)
	if err != nil {
		log := logger.FromContext(ctx)
		log.Error(err, "failed to build knative service", "error", err)
		return nil, err
	}
	return computeResources.Job, nil
}

// buildResources builds the resources of the workload (not include the knative service)
func (iwc *Controller) buildResourceStructs(ctx context.Context, aggr *resource.Aggregator, owner client.Object, knativeService client.Object) error {
	workload := aggr.GetWorkload()

	// build services
	resourceBuilderOptions := builderCommon.ResourceBuilderOptions{Owner: owner}
	services, err := iwc.commonResourcesBuilder.BuildServices(ctx, workload, resourceBuilderOptions)
	if err != nil {
		return err
	}

	// build external services
	authPortOffset := 0
	if workload.Spec.ServingPortAuthorizationType.DerefOrDefault(constants.ServingPortAuthorizationTypeDefault) != constants.ServingPortAuthorizationTypePublic {
		authPortOffset = 1
	}
	externalsServices, err := iwc.commonResourcesBuilder.BuildExternalUrls(ctx, workload, &builderCommon.ExternalBuilderOptions{
		Owner:          owner,
		DisableAuth:    ptr.To(!IsKnativeServiceSupportsAuth(knativeService)),
		AuthPortOffset: authPortOffset,
	})
	if err != nil {
		return err
	}

	// will be deleted in the future
	// pvc does not owned by the workload
	pvcList, err := iwc.commonResourcesBuilder.BuildPVC(ctx, workload, resourceBuilderOptions)
	if err != nil {
		return err
	}

	credentialsList, err := iwc.commonResourcesBuilder.BuildCredentials(ctx, workload, resourceBuilderOptions)
	if err != nil {
		return err
	}

	// Build NIM metrics Service if needed (bypasses OAuth2 Proxy for metrics scraping)
	metricsService, err := iwc.buildNIMMetricsServiceIfNeeded(ctx, infw.NewFromJobWorkload(workload), owner)
	if err != nil {
		return err
	}

	for _, service := range services {
		aggr.AddDesired(service)
	}
	if metricsService != nil {
		aggr.AddDesired(metricsService)
	}
	aggr.AddDesired(externalsServices...)
	aggr.AddDesired(pvcList...)
	aggr.AddDesired(credentialsList...)
	return nil
}

func (iwc *Controller) aggregateResourcesStatus(ctx context.Context, aggr *resource.Aggregator, knativeService *knv1.Service, workload *jobs.JobWorkload, lastCreatedRevisionPods []corev1.Pod, buildError *string) error {
	iwc.StatusAggregator.Clear()
	iwc.StatusAggregator.WithServices(aggr.GetServices())
	iwc.StatusAggregator.WithSecrets(aggr.GetSecrets())
	iwc.StatusAggregator.(inferenceaggregator.InferenceWorkloadStatusCalculater).WithKnativeService(knativeService)
	iwc.StatusAggregator.(inferenceaggregator.InferenceWorkloadStatusCalculater).WithLatestCreatedRevisionPods(lastCreatedRevisionPods)
	iwc.StatusAggregator.(inferenceaggregator.InferenceWorkloadStatusCalculater).WithBuildError(buildError)
	iwc.StatusAggregator.WithResources(aggr.GetExternalNetwork())
	iwc.StatusAggregator.WithWorkload(workload)
	workloadStatus := iwc.StatusAggregator.CalculateAll(ctx, aggr)
	workloadStatus.DeepCopyInto(&workload.Status)

	return nil
}

func (iwc *Controller) shouldRebuildKnativeService(ctx context.Context, workload *jobs.JobWorkload) (bool, string, error) {
	// We rebuild the Knative service when the workload spec actually changed (the recorded
	// spec-checksum annotation differs from the current spec) or when the service is missing.
	// Gating on the spec prevents redundant rebuilds on every reconcile: on a cluster upgrade
	// the InferenceWorkload spec is unchanged, so a spec-driven rebuild is a no-op and we avoid
	// creating a new Knative revision that would race the progress deadline.
	//
	// Rebuilding a missing service is safe: there is no live revision to race,
	// so recreating one only restores a deleted service (self-heal). A manually edited service
	// that still exists is not reconciled back to spec, since detecting that would require
	// comparing the rendered output and would reintroduce the redundant-rebuild bug.

	infWorkload := infw.NewFromJobWorkload(workload)

	newSpecChecksum, err := spec_checksum.CalculateSpecChecksum(infWorkload)
	if err != nil {
		return true, "", err
	}

	// self-heal: recreate the Knative service if it was deleted out from under the controller.
	knativeServiceExists, err := iwc.knativeServiceExists(ctx, workload.Namespace, workload.Name)
	if err != nil {
		return true, "", err
	}
	if !knativeServiceExists {
		return true, newSpecChecksum, nil
	}

	// no recorded checksum yet (first reconcile, or created by a controller version that did
	// not record one): rebuild once and persist the checksum.
	currentSpecChecksum := spec_checksum.GetSpecChecksumAnnotation(infWorkload)
	if currentSpecChecksum == "" {
		return true, newSpecChecksum, nil
	}

	if currentSpecChecksum != newSpecChecksum {
		return true, newSpecChecksum, nil
	}

	return false, newSpecChecksum, nil
}

func (iwc *Controller) knativeServiceExists(ctx context.Context, workloadNS string, workloadName string) (bool, error) {
	service := &knv1.Service{}
	err := iwc.k8sClient.Get(ctx, client.ObjectKey{Namespace: workloadNS, Name: workloadName}, service)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (iwc *Controller) SuccessReconcile(ctx context.Context, workload *jobs.JobWorkload, originalWorkload *jobs.JobWorkload) (ctrl.Result, error) {
	err := iwc.statusUpdater.Update(ctx, workload, originalWorkload)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (iwc *Controller) FailedReconcileStatus(ctx context.Context, aggr *resource.Aggregator, err *workloadissue.Issue, toRequeue bool) (ctrl.Result, error) {
	log := logger.FromContext(ctx)
	log.Error(err, "workload reconcile failed with error")

	if !toRequeue {
		_, errGetResource := iwc.getWorkloadResources(ctx, aggr)
		if errGetResource != nil {
			return ctrl.Result{}, errGetResource
		}

		errMessage := err.Message()
		updateError := iwc.updateWorkloadStatus(ctx, aggr, &errMessage)
		if updateError != nil {
			return ctrl.Result{}, updateError
		}

		// clean the error that we return to the reconcile
		err = nil
	}

	workload := aggr.GetWorkload()
	updateError := iwc.statusUpdater.Update(ctx, workload, nil)
	if updateError != nil {
		log.Error(updateError, "failed to update failure status")
	}
	return ctrl.Result{}, err
}

func getBackOffForUpdateKnativeService() *backoff.ExponentialBackOff {
	backOffForUpdateKnativeService := backoff.NewExponentialBackOff()
	backOffForUpdateKnativeService.InitialInterval = 50 * time.Millisecond // sets the initial interval between retries.
	backOffForUpdateKnativeService.MaxInterval = 300 * time.Millisecond    // sets the maximum interval between retries.
	backOffForUpdateKnativeService.MaxElapsedTime = time.Second            // sets the maximum total time for retries.
	return backOffForUpdateKnativeService
}

// rebuildAndSubmitKnative rebuilds the Knative service and submits it to the cluster.
// Returns:
// - issues: a list of issues encountered during the rebuild and submission process.
// - toRequeue: a boolean indicating whether the reconcile should be re-queued.
func (iwc *Controller) rebuildAndSubmitKnative(ctx context.Context, workload *jobs.JobWorkload, owner client.Object) (client.Object, []*workloadissue.Issue, bool) {
	log := logger.FromContext(ctx)
	log.Info("rebuilding knative service", "workload", workload.Name)

	knative, err := iwc.buildKnative(ctx, workload, owner)
	if err != nil {
		return nil, []*workloadissue.Issue{issue.NewIssue(workload, err)}, false
	}

	// create or update knative service with retry mechanism in case of conflict
	// if the error is not a conflict error, return the error
	var issues []*workloadissue.Issue
	attempt := 0
	ticker := backoff.NewTicker(getBackOffForUpdateKnativeService())
	defer ticker.Stop()
	for range ticker.C {
		issues = iwc.submitResources(ctx, workload, []client.Object{knative})

		// if the knative service was created successfully, break the loop
		// or if the error is not a conflict error, return the error
		if issues == nil {
			return knative, nil, false
		}

		if !errors.IsConflict(issues[0].GetError()) {
			return nil, issues, false
		}

		attempt += 1
		log.Info("conflict while updating knative service, retrying", "attempt", attempt, "workload", workload.Name)
	}

	// if we reached here, it means that the knative service was not created successfully
	// if the error is a conflict error, we want to re try and create this resource again in the next reconcile
	return nil, issues, errors.IsConflict(issues[0].GetError())
}

func (iwc *Controller) rebuildWorkloadKnativeService(ctx context.Context, workload *jobs.JobWorkload, owner client.Object) (client.Object, []*workloadissue.Issue, bool) {
	log := logger.FromContext(ctx)
	log.Info("rebuilding workload knative service", "workload", workload.Name)

	// build knative service and submit it
	knativeService, issues, toRequeue := iwc.rebuildAndSubmitKnative(ctx, workload, owner)
	if issues != nil {
		return nil, issues, toRequeue
	}

	return knativeService, nil, false
}

func (iwc *Controller) rebuildWorkloadResources(ctx context.Context, aggr *resource.Aggregator, owner client.Object, knativeService client.Object) ([]*workloadissue.Issue, bool) {
	workload := aggr.GetWorkload()
	log := logger.FromContext(ctx)
	log.Info("rebuilding workload resources", "workload", workload.Name)

	// build the other resources
	err := iwc.buildResourceStructs(ctx, aggr, owner, knativeService)
	if err != nil {
		return []*workloadissue.Issue{issue.NewIssue(workload, err)}, false
	}

	// create or update resource based on the generated list.
	issues := iwc.submitResources(ctx, workload, aggr.GetDesiredList())
	if issues != nil {
		return issues, true
	}

	return nil, false

}

// Submits the resources to the cluster
func (iwc *Controller) submitResources(ctx context.Context, workload *jobs.JobWorkload, objects []client.Object) []*workloadissue.Issue {
	log := logger.FromContext(ctx)
	log.Info("submitting workload resources to kubernetes api server", "workload", workload.Name)

	issues := iwc.resourcesManager.UpsertResources(ctx, workload, objects, iwc.options)
	if len(issues) > 0 {
		return issues
	}
	return nil
}

func (iwc *Controller) setSpecChecksumAnnotation(_ context.Context, workload *jobs.JobWorkload, newSpecChecksum string) (bool, error) {
	infWorkload := infw.NewFromJobWorkload(workload)

	isChecksumModified, err := spec_checksum.SetSpecChecksumAnnotation(infWorkload, newSpecChecksum)
	if err != nil {
		return false, err
	}

	// turn the migration annotation to false, so we will no longer treat it as a migrated one
	isAnnotationModified := false
	if spec_checksum.GetSpecChecksumMigratedAnnotation(infWorkload) {
		isAnnotationModified = spec_checksum.SetSpecChecksumMigratedAnnotation(infWorkload, false)
	}

	infWorkload.CopyTo(workload)

	return isChecksumModified || isAnnotationModified, nil
}

func (iwc *Controller) getKnativeService(ctx context.Context, workloadNS string, workloadName string) (*knv1.Service, error) {
	service := &knv1.Service{}
	err := iwc.k8sClient.Get(ctx, client.ObjectKey{Namespace: workloadNS, Name: workloadName}, service)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
	}

	return service, nil
}

func (iwc *Controller) getWorkloadResources(ctx context.Context, aggr *resource.Aggregator) ([]client.Object, error) {
	workload := aggr.GetWorkload()
	log := logger.FromContext(ctx)
	workloadSpecResources, err := iwc.getWorkloadSubResource(ctx, workload)
	if err != nil {
		log.Error(err, "Failed to generate workload sub resources preview", "workload", workload.Name)
		return nil, err
	}

	log.Info("Retrieving workload sub resources from cluster", "workload", workload.Name)
	return iwc.resourcesManager.GetAllResources(ctx, aggr, workloadSpecResources)
}

func isNIM(workload *infw.InferenceWorkload) bool {
	labels := workload.Spec.Labels.Deref()
	if val, exists := labels[constants.LabelNIMWorkload]; exists && val.Value != nil && *val.Value == constants.BoolTrueAsString {
		return true
	}
	return false
}

// buildNIMMetricsServiceIfNeeded builds a dedicated metrics Service for NIM workloads
// This encapsulates NIM-specific logic for metrics collection
func (iwc *Controller) buildNIMMetricsServiceIfNeeded(
	ctx context.Context,
	workload *infw.InferenceWorkload,
	owner client.Object,
) (*corev1.Service, error) {
	if !isNIM(workload) {
		return nil, nil
	}

	servingPort, err := builders.GetServingPort(ctx, &workload.Spec.NetworkingFields, workload.Kind)
	if err != nil {
		return nil, err
	}

	return builders.BuildNIMMetricsService(
		ctx,
		owner,
		workload.GetChildName(),
		workload.Namespace,
		servingPort,
		iwc.k8sClient,
	)
}

func IsKnativeServiceSupportsAuth(workload client.Object) bool {
	ks, ok := workload.(*knv1.Service)
	if ok && ks != nil {
		return workloadbuilders.IsWorkloadSupportsAuth(ks.Spec.Template.Spec.Containers, ks.Spec.Template.Spec.InitContainers)
	}

	return false
}
