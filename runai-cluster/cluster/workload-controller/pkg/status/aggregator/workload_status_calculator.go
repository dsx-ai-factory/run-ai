/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregator

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	job "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	workloadCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	jobInterface "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/status"
	openshiftV1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	JobFailed                    = "job failed"
	JobFailedReason              = "JobFailed"
	UnknownPendingPods           = -1
	workloadIsSuspending         = "workload is suspending"
	JobSuspended                 = "job suspended"
	JobComplete                  = "job completed"
	JobNotRunning                = "job is not running"
	JobNotExistsReason           = "JobNotExists"
	JobNotExistsMessage          = "Job does not exists"
	InitializingSinglePodMessage = "Workload is initializing, the pod is still not ready"
	InitializingMultiPodsMessage = "Workload is initializing, minimum ready pods not reached, waiting for %d pods to be ready"
	GenericJobFailedMessage      = "job reached failed state"

	// issue related
	WorkloadHasIssuesMessage         = "Workload has issues"
	failedDueToIssueSingularPhaseMsg = "Workload failed due to a %s issue"
	failedDueToIssuesPluralPhaseMsg  = "Workload failed due to %s issues"

	degradedDueToIssueSingularPhaseMsg = "Workload is in degraded state due to a %s issue"
	degradedDueToIssuesPluralPhaseMsg  = "Workload is in degraded state due to %s issues"

	// MNNVL
	MnnvlNotReadyMessage     = "MNNVL support is not ready"
	MnnvlNotReadyReason      = "MNNVLSupportNotReady"
	MnnvlNotRequestedMessage = "MNNVL support is not requested"
	MnnvlNotRequestedReason  = "MNNVLNotRequested"
)

var log = logf.Log.WithName("status-calculator")

var (
	defaultDelay    = time.Minute
	suspendingDelay = 30 * time.Second
)

func GetDefaultDelay() time.Duration {
	return defaultDelay
}

func GetSuspendingDelay() time.Duration {
	return suspendingDelay
}

type WorkloadConditionsMap map[common.WorkloadConditionType]*common.WorkloadCondition

type StatusCalculator struct {
	mnnvlSupportedInCluster bool
	workload                *job.JobWorkload
	job                     jobInterface.JobAdapter
	services                map[string]*v1.Service
	routes                  map[string]*openshiftV1.Route
	ingresses               map[string]*network.Ingress
	httpRoutes              map[string]*gatewayv1.HTTPRoute
	secrets                 map[string]*v1.Secret
	computeDomain           *nvapi.ComputeDomain
	expectedResources       []common.Resource
	workloadIssues          WorkloadIssues
}

func NewStatusCalculator(mnnvlSupportedInCluster bool) *StatusCalculator {
	return &StatusCalculator{
		mnnvlSupportedInCluster: mnnvlSupportedInCluster,
		services:                make(map[string]*v1.Service),
		routes:                  make(map[string]*openshiftV1.Route),
		ingresses:               make(map[string]*network.Ingress),
		httpRoutes:              make(map[string]*gatewayv1.HTTPRoute),
		secrets:                 make(map[string]*v1.Secret),
		expectedResources:       make([]common.Resource, 0),
		workloadIssues:          NewWorkloadIssues(),
	}
}

func (sc *StatusCalculator) WithResources(objects []client.Object) {
	for _, resource := range objects {
		switch resourceType := resource.(type) {
		case *v1.Service:
			service := resourceType
			sc.services[service.Name] = service
		case *openshiftV1.Route:
			route := resourceType
			sc.routes[route.Name] = route
		case *network.Ingress:
			ingress := resourceType
			sc.ingresses[ingress.Name] = ingress
		case *gatewayv1.HTTPRoute:
			sc.httpRoutes[resourceType.Name] = resourceType
		case *nvapi.ComputeDomain:
			sc.computeDomain = resourceType
		case *v1.Secret:
			secret := resourceType
			sc.secrets[secret.Name] = resourceType
		default:
			// current log framework does not support logs level in convent way
			// for now it am adding it direct to string // remove this wheN moving to `zap`
			log.Info("Warning supported resource type", "resource", resource)
			continue
		}
	}
}

func (sc *StatusCalculator) Clear() {
	sc.services = make(map[string]*v1.Service)
	sc.routes = make(map[string]*openshiftV1.Route)
	sc.ingresses = make(map[string]*network.Ingress)
	sc.httpRoutes = make(map[string]*gatewayv1.HTTPRoute)
	sc.secrets = make(map[string]*v1.Secret)
	sc.expectedResources = make([]common.Resource, 0)
	sc.workload = nil
	sc.job = nil
	sc.computeDomain = nil
	sc.workloadIssues = NewWorkloadIssues()
}

func (sc *StatusCalculator) WithServices(services []*v1.Service) {
	for _, service := range services {
		sc.services[service.Name] = service
	}
}

func (sc *StatusCalculator) WithWorkload(workload *job.JobWorkload) {
	sc.workload = workload
}

func (sc *StatusCalculator) WithJob(job jobInterface.JobAdapter) {
	sc.job = job
}

func (sc *StatusCalculator) WithComputeDomain(computeDomain *nvapi.ComputeDomain) {
	sc.computeDomain = computeDomain
}

func (sc *StatusCalculator) WithSecrets(secrets []*v1.Secret) {
	for _, secret := range secrets {
		sc.secrets[secret.Name] = secret
	}
}

func (sc *StatusCalculator) CalculateAll(ctx context.Context, aggr *resource.Aggregator) job.JobStatus {
	if sc.workload == nil {
		sc.workload = &job.JobWorkload{}
	}

	sc.handleMinimumRequiredReplica()

	// in some of the cases where the workload is failed due to issues the job was not created -> no point to calculate again
	if sc.workload.IsFailedByWorkload() && sc.job == nil {
		return sc.workload.Status
	}

	sc.updateServiceStatuses(ctx)

	if sc.job == nil {
		// clear job status and set to Creating by default, but preserve the cumulative
		// RestartCount, which would otherwise be wiped while the job is transiently
		// absent during a restart.
		sc.workload.Status.WorkloadCommonStatus = common.WorkloadCommonStatus{
			Phase:        common.PhaseCreating.String(),
			RestartCount: sc.workload.Status.RestartCount,
		}
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready, v1.ConditionFalse, JobNotExistsReason,
			JobNotExistsMessage)
		return sc.workload.Status
	}

	sc.handleJobUrls()

	sc.jobStatus()

	sc.handleStartTimeStatus()

	return sc.CalculatePhaseAndConditions(ctx, aggr)
}

func (sc *StatusCalculator) CalculatePhaseAndConditions(ctx context.Context, aggr *resource.Aggregator) job.JobStatus {
	if sc.workload == nil {
		sc.workload = &job.JobWorkload{}
	}

	sc.calculateWorkloadIssues(aggr)

	sc.updateConditionStatus(ctx)

	sc.handleWorkloadPhase(ctx)

	return sc.workload.Status
}

func (sc *StatusCalculator) calculateWorkloadIssues(aggr *resource.Aggregator) {
	wi := WorkloadIssues{
		hasIssues: aggr.HasIssues(),
	}

	groups := make(map[string]bool)
	for _, iss := range aggr.GetIssuesList() {
		if iss.IsCritical() {
			wi.hasCritical = true
			groups[iss.GetGroup().String()] = true

			if iss.GetGroup() == issue.Compute {
				wi.hasCriticalCompute = true
			}
		}
	}

	wi.groupsWithCritical = utils.GetMapKeys(groups)
	sc.workloadIssues.Set(wi)
}

// generateFailedByIssuesMessage's output should be consistent as it appears in the status and calculated multiple times
func generateFailedByIssuesMessage(groups []string) string {
	if len(groups) == 0 {
		return ""
	}

	msgTemplate := failedDueToIssuesPluralPhaseMsg
	if len(groups) == 1 {
		msgTemplate = failedDueToIssueSingularPhaseMsg
	}

	sort.Strings(groups)
	return fmt.Sprintf(msgTemplate, strings.Join(groups, ", "))
}

// generateDegradedByIssuesMessage's output should be consistent as it appears in the status and calculated multiple times
func generateDegradedByIssuesMessage(groups []string) string {
	if len(groups) == 0 {
		return ""
	}

	msgTemplate := degradedDueToIssuesPluralPhaseMsg
	if len(groups) == 1 {
		msgTemplate = degradedDueToIssueSingularPhaseMsg
	}

	sort.Strings(groups)
	return fmt.Sprintf(msgTemplate, strings.Join(groups, ", "))
}

func (sc *StatusCalculator) handleMinimumRequiredReplica() {
	// only for training and interactive workloads,  when refactor the status calculation, we will handle this in a better way
	if sc.workload.Kind == constants.KindTrainingWorkload || sc.workload.Kind == constants.KindInteractiveWorkload {
		sc.workload.Status.MinRequiredReadyPods = workloadCommon.MinimumReadyReplicas
	}
}

func (sc *StatusCalculator) updateServiceStatuses(ctx context.Context) {
	sc.workload.Status.NetworkStatus.ServiceStatuses = make([]common.ServiceStatus, 0)
	if sc.services == nil {
		return
	}
	for _, service := range sc.services {
		serviceStatus := sc.buildServiceStatus(ctx, service)
		sc.workload.Status.NetworkStatus.ServiceStatuses = append(sc.workload.Status.NetworkStatus.ServiceStatuses, serviceStatus)
	}
}

// buildServiceStatus builds a ServiceStatus for the given service, including ports and LoadBalancer status if applicable.
func (sc *StatusCalculator) buildServiceStatus(ctx context.Context, service *v1.Service) common.ServiceStatus {
	if service == nil {
		return common.ServiceStatus{}
	}
	return common.ServiceStatus{
		Name:         service.Name,
		Type:         service.Spec.Type,
		Ports:        buildServicePorts(service),
		LoadBalancer: buildLoadBalancerStatus(ctx, service),
	}
}

// buildServicePorts returns a list of ExposePort for the given service (no sorting needed for map type).
func buildServicePorts(service *v1.Service) []common.ExposePort {
	if service == nil || service.Spec.Ports == nil {
		return []common.ExposePort{}
	}
	ports := make([]common.ExposePort, 0, len(service.Spec.Ports))
	for _, port := range service.Spec.Ports {
		externalPort := port.Port
		if port.NodePort != 0 {
			externalPort = port.NodePort
		}
		ports = append(ports, common.ExposePort{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       externalPort,
			TargetPort: port.TargetPort,
		})
	}
	return ports
}

// buildLoadBalancerStatus returns a pointer to LoadBalancerStatus for the given service if it is of type LoadBalancer, otherwise returns nil.
func buildLoadBalancerStatus(ctx context.Context, service *v1.Service) *common.LoadBalancerStatus {
	var emptyLoadBalancerStatus common.LoadBalancerStatus
	if service == nil || service.Spec.Type != v1.ServiceTypeLoadBalancer {
		return nil
	}
	if service.Status.LoadBalancer.Ingress == nil {
		return &emptyLoadBalancerStatus
	}
	logStatus := logf.FromContext(ctx).WithName("status-calculator")
	var ingressList []common.LoadBalancerIngress
	for _, ingress := range service.Status.LoadBalancer.Ingress {
		if ingress.IP == "" && ingress.Hostname == "" {
			logStatus.V(4).Info("Ignoring ingress with empty IP and Hostname in LoadBalancer status", "service", service.Name)
			continue
		}
		ingressList = append(ingressList, common.LoadBalancerIngress{
			IP:       ingress.IP,
			Hostname: ingress.Hostname,
		})
	}
	if len(ingressList) == 0 {
		return &emptyLoadBalancerStatus
	}
	// Sort ingress by IP if present, otherwise Hostname
	sort.Slice(ingressList, func(i, j int) bool {
		keyA := ingressList[i].IP
		if keyA == "" {
			keyA = ingressList[i].Hostname
		}
		keyB := ingressList[j].IP
		if keyB == "" {
			keyB = ingressList[j].Hostname
		}
		return keyA < keyB
	})
	return &common.LoadBalancerStatus{Ingress: ingressList}
}

func (sc *StatusCalculator) jobStatus() {

	if sc.job.GetStatus().GetPendingPods() != UnknownPendingPods {
		sc.workload.Status.PendingPods = sc.job.GetStatus().GetPendingPods()
	} else {
		// if we don't know the number of pending pods, we assume that the job as 0 pending pods
		sc.workload.Status.PendingPods = 0
	}
	sc.workload.Status.ReadyPods = sc.job.GetStatus().GetReadyPods()
	sc.workload.Status.RunningPods = sc.job.GetStatus().GetRunningPods()
	sc.workload.Status.SucceededPods = sc.job.GetStatus().GetSucceededPods()
	sc.workload.Status.FailedPods = sc.job.GetStatus().GetFailedPods()
	sc.workload.Status.TotalAllocatedGPU = sc.job.GetStatus().GetAllocatedGPU()
	sc.workload.Status.TotalAllocatedGPUMemory = sc.job.GetStatus().GetAllocatedGPUMemory()

	// workload's completion time can also be set by us when failing a workload.
	// in this case, the job might not be in a finite state - we don't want to override the time we set with nil from the job
	jobCompletionTime := sc.job.GetStatus().GetCompletionTime()
	if !jobCompletionTime.IsZero() {
		sc.workload.Status.WorkloadCommonStatus.CompletionTime = jobCompletionTime
	}
}

func (sc *StatusCalculator) handleJobUrls() {
	if urls, ok := sc.job.GetAnnotations()[constants.AnnotationJobUrl]; ok {
		urls = strings.Replace(urls, " ", "", -1)
		currentUrls := strings.Split(urls, ",")
		sc.workload.Status.WorkloadCommonStatus.ExternalUrls = make([]string, len(currentUrls))
		copy(sc.workload.Status.WorkloadCommonStatus.ExternalUrls, currentUrls)
	}
}

func (sc *StatusCalculator) handleReadyCondition(ctx context.Context) {
	// check issues before job - when has issues job might not have been created, so if job == nil it's not the actual reason
	if sc.workloadIssues.hasIssues {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready,
			v1.ConditionFalse, "", WorkloadHasIssuesMessage)
		return
	}

	if sc.job == nil {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready,
			v1.ConditionFalse, JobNotExistsReason, JobNotExistsMessage)
		return
	}

	if sc.job.GetStatus().IsJobFinished() {
		if sc.job.GetStatus().IsJobFailed() {
			status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready,
				v1.ConditionFalse, "", JobFailed)
			return
		}
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready,
			v1.ConditionFalse, "", JobComplete)
		return
	}

	if IsJobSuspended(sc.job) {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready,
			v1.ConditionFalse, "", JobSuspended)
		return
	}

	if !isWorkloadActive(sc.workload) {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready,
			v1.ConditionFalse, "", workloadIsSuspending)
		return
	}

	if !sc.job.GetStatus().IsRunning() {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready,
			v1.ConditionFalse, "", JobNotRunning)
		return
	}

	errorList := make([]error, 0)
	if sc.job.GetStatus().IsPending() {
		errorList = append(errorList, fmt.Errorf("still has pending pods (%d)", sc.job.GetStatus().GetPendingPods()))
	}

	for _, resource := range sc.expectedResources {
		var ok bool
		switch resource.Kind.Kind {
		case constants.KindService:
			_, ok = sc.services[resource.Name]
		case constants.KindIngress:
			_, ok = sc.ingresses[resource.Name]
		case constants.KindRoute:
			_, ok = sc.routes[resource.Name]
		case constants.KindHTTPRoute:
			_, ok = sc.httpRoutes[resource.Name]
		case constants.KindComputeDomain:
			ok = sc.computeDomain != nil && sc.computeDomain.Name == resource.Name
		case constants.KindSecret:
			_, ok = sc.secrets[resource.Name]
		default:
			continue
		}

		if !ok {
			errorList = append(errorList, fmt.Errorf("%s %s does not exists", strings.ToLower(resource.Kind.Kind), resource.Name))
		}
	}

	if len(errorList) > 0 {
		message := errors.NewAggregate(errorList).Error()
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready,
			v1.ConditionFalse, "", message)
		return
	}

	mnnvlStatus := mnnvlutils.FromContext(ctx)
	if mnnvlStatus.IsNetworkAccelerated() && !sc.isGpuNetworkAccelerationReady(mnnvlStatus) {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready,
			v1.ConditionFalse, MnnvlNotReadyReason, MnnvlNotReadyMessage)
		return
	}

	status.UpsertCondition(&sc.workload.Status.Conditions, common.Ready,
		v1.ConditionTrue, "", "")
}

func (sc *StatusCalculator) WithExpectedResources(resources []common.Resource) {
	sc.expectedResources = append(sc.expectedResources, resources...)
}

func (sc *StatusCalculator) handleStartTimeStatus() {

	if sc.job.GetStatus().IsRunning() && sc.workload.Status.WorkloadCommonStatus.StartTime.IsZero() {
		sc.workload.Status.WorkloadCommonStatus.StartTime = &metav1.Time{Time: time.Now()}
	} else if !sc.job.GetStatus().IsRunning() && !sc.workload.Status.WorkloadCommonStatus.StartTime.IsZero() {
		sc.workload.Status.WorkloadCommonStatus.StartTime = nil
	}
}

func (sc *StatusCalculator) updateSucceededCondition() {
	if sc.job != nil && sc.job.GetStatus().IsJobFinished() && !sc.job.GetStatus().IsJobFailed() {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Succeeded, v1.ConditionTrue, "", "")
		return
	}

	status.UpsertCondition(&sc.workload.Status.Conditions, common.Succeeded, v1.ConditionFalse, "", "")
}

func (sc *StatusCalculator) updateFailedCondition() {
	/*
		updateFailedCondition sets the reason according to the source of the failure:
		JobFailedReason - failed because of the job
		WorkloadFailedReason - failed by the workload (issues)
	*/

	if sc.job != nil && sc.job.GetStatus().IsJobFailed() {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Failed, v1.ConditionTrue,
			JobFailedReason, sc.job.GetStatus().GetFailedInfo())
		return
	}

	// critical compute issue OR
	// critical, non-compute issue in a non-running phase
	if sc.workloadIssues.hasCriticalCompute ||
		(sc.workloadIssues.hasCritical && (sc.job == nil || !sc.job.GetStatus().IsRunning())) {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Failed, v1.ConditionTrue,
			common.WorkloadFailedReason, sc.getFailedMessages())
		return
	}

	status.UpsertCondition(&sc.workload.Status.Conditions, common.Failed, v1.ConditionFalse, "", "")
}

func (sc *StatusCalculator) updateConditionStatus(ctx context.Context) {
	sc.updateSuspendCondition()
	sc.updateFailedCondition()
	sc.updateSucceededCondition()
	sc.updatePendingCondition()
	sc.handleReadyCondition(ctx)
	sc.updateInitializedCondition()
	sc.updateDegradedCondition()
	sc.updateGpuNetworkAcceleratedCondition(ctx)
}

func (sc *StatusCalculator) updateGpuNetworkAcceleratedCondition(ctx context.Context) {
	if !sc.mnnvlSupportedInCluster {
		return
	}

	mnnvlStatus := mnnvlutils.FromContext(ctx)
	if !mnnvlStatus.IsNetworkAccelerated() {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.GPUNetworkAccelerated, v1.ConditionFalse, MnnvlNotRequestedReason, MnnvlNotRequestedMessage)
	} else if sc.isGpuNetworkAccelerationReady(mnnvlStatus) {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.GPUNetworkAccelerated, v1.ConditionTrue, "", "")
	} else {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.GPUNetworkAccelerated, v1.ConditionFalse, MnnvlNotReadyReason, MnnvlNotReadyMessage)
	}
}

func (sc *StatusCalculator) updateDegradedCondition() {
	// degraded is subset of running
	if sc.job != nil && sc.job.GetStatus().IsRunning() && sc.workloadIssues.hasCritical {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Degraded, v1.ConditionTrue, common.DegradedReason.String(), "")
		return
	}

	status.UpsertCondition(&sc.workload.Status.Conditions, common.Degraded, v1.ConditionFalse, "", "")
}

func (sc *StatusCalculator) updateInitializedCondition() {
	if sc.job == nil {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Initialized,
			v1.ConditionFalse, JobNotExistsReason, "")
		return
	}

	var reason, message string
	if sc.job.GetStatus().GetReadyPods() > 0 {
		status.UpsertCondition(&sc.workload.Status.Conditions, common.Initialized, v1.ConditionTrue, "", "")
		return
	}

	if sc.job.GetStatus().IsInitializing() {
		reason, message = sc.getInitializingReasonMessages()
	} else if sc.job.GetStatus().IsCreating() {
		reason, message = common.CreatingReason.String(), ""
	} else if sc.job.GetStatus().IsPending() {
		reason, message = common.PendingReason.String(), ""
	} else if sc.job.GetStatus().IsSuspended() {
		reason, message = common.SuspendedReason.String(), ""
	}
	status.UpsertCondition(&sc.workload.Status.Conditions, common.Initialized, v1.ConditionFalse, reason, message)
}

func (sc *StatusCalculator) getFailedMessages() string {
	if sc.job != nil && sc.job.GetStatus().IsJobFailed() {
		info := sc.job.GetStatus().GetFailedInfo()
		if info != "" {
			return info
		}
		return GenericJobFailedMessage
	}

	if len(sc.workloadIssues.groupsWithCritical) > 0 {
		return generateFailedByIssuesMessage(sc.workloadIssues.groupsWithCritical)
	}

	return ""
}

func (sc *StatusCalculator) getInitializingReasonMessages() (string, string) {
	if sc.workload.Kind == constants.KindInteractiveWorkload ||
		(sc.workload.Kind == constants.KindTrainingWorkload &&
			sc.workload.Spec.Parallelism.Deref32default(1) == 1) {
		return common.ReplicaNotReadyReason.String(), InitializingSinglePodMessage
	}
	return common.MinimumReplicasNotReadyReason.String(),
		fmt.Sprintf(InitializingMultiPodsMessage, workloadCommon.MinimumReadyReplicas)
}

func (sc *StatusCalculator) updateSuspendCondition() {
	if sc.job != nil && sc.job.GetStatus().IsSuspended() {
		if sc.job.GetStatus().GetReadyPods() == 0 {
			status.UpsertCondition(&sc.workload.Status.Conditions, common.Suspended, v1.ConditionTrue,
				"", sc.job.GetStatus().GetSuspendInfo())
		} else {
			status.UpsertCondition(&sc.workload.Status.Conditions, common.Suspended, v1.ConditionFalse,
				common.PodsStillRunningReason, sc.job.GetStatus().GetSuspendInfo())
		}

		return
	}

	status.UpsertCondition(&sc.workload.Status.Conditions, common.Suspended, v1.ConditionFalse,
		"", "")
}

func (sc *StatusCalculator) updatePendingCondition() {
	if sc.job != nil && sc.job.GetStatus().IsPending() {
		status.UpsertCondition(&sc.workload.Status.WorkloadCommonStatus.Conditions,
			common.Pending, v1.ConditionTrue, "", "")
		return
	}

	status.UpsertCondition(&sc.workload.Status.Conditions, common.Pending, v1.ConditionFalse, "", "")
}

func (sc *StatusCalculator) handleWorkloadPhase(ctx context.Context) {
	logging := logger.FromContext(ctx)

	var newPhaseEnum common.Phase
	var phaseMessage string

	conditions := status.GetWorkloadConditionsAsMap(&sc.workload.Status.Conditions)
	switch {
	// Terminating
	case sc.job != nil && sc.job.GetStatus().IsTerminating():
		newPhaseEnum = common.PhaseTerminating
	// Failed
	case isConditionTrue(conditions, common.Failed):
		newPhaseEnum = common.PhaseFailed
		phaseMessage = sc.getFailedMessages()
	// Completed
	case isConditionTrue(conditions, common.Succeeded):
		newPhaseEnum = common.PhaseCompleted
	// suspending
	case !isWorkloadActive(sc.workload) && !isConditionTrue(conditions, common.Suspended):
		newPhaseEnum = common.PhaseSuspending
	case isWorkloadActive(sc.workload) && isConditionTrue(conditions, common.Suspended):
		newPhaseEnum = common.PhaseResuming
	// Suspended
	case isConditionTrue(conditions, common.Suspended):
		newPhaseEnum = common.PhaseSuspended
	// Running (or Degraded)
	case isConditionTrue(conditions, common.Initialized):
		if isConditionTrue(conditions, common.Degraded) {
			newPhaseEnum = common.PhaseDegraded
			phaseMessage = generateDegradedByIssuesMessage(sc.workloadIssues.groupsWithCritical)
		} else {
			newPhaseEnum = common.PhaseRunning
		}
	// Initializing
	case sc.job != nil && sc.job.GetStatus().IsInitializing():
		newPhaseEnum = common.PhaseInitializing
		_, phaseMessage = sc.getInitializingReasonMessages()
	// Pending
	case isConditionTrue(conditions, common.Pending):
		newPhaseEnum = common.PhasePending
	// Creating
	default:
		newPhaseEnum = common.PhaseCreating
	}

	if newPhaseEnum.String() == sc.workload.Status.WorkloadCommonStatus.Phase {
		logging.V(4).Info("Phase no update needed", "workloadName", sc.workload.Name, "workloadNS", sc.workload.Namespace, "workloadPhase", sc.workload.Status.WorkloadCommonStatus.Phase)
	} else {
		logging.Info("Phase update", "workloadName", sc.workload.Name, "workloadNS", sc.workload.Namespace, "workloadPhaseCurrent", sc.workload.Status.WorkloadCommonStatus.Phase, "workloadPhaseUpdate", newPhaseEnum.String())
		sc.workload.Status.WorkloadCommonStatus.Phase = newPhaseEnum.String()
	}
	sc.workload.Status.WorkloadCommonStatus.PhaseMessage = phaseMessage
}

func isConditionTrue(conditions WorkloadConditionsMap, desiredType common.WorkloadConditionType) bool {
	if c, ok := conditions[desiredType]; ok {
		return c.Status == v1.ConditionTrue
	}
	return false
}

// IsJobSuspended checks if the workload is actually suspended, not only based on the job's status
// this check covers the edge case of certain job types that fail to resume if the pods were not fully deleted at the suspend action (running pods with deletion timestamp)
// this check will help block premature resume requests via the api, but applying via kubectl might still cause failure at resume
func IsJobSuspended(job jobInterface.JobAdapter) bool {
	return job != nil && job.GetStatus().IsSuspended() && job.GetStatus().GetReadyPods() == 0
}

func (sc *StatusCalculator) isGpuNetworkAccelerationReady(mnnvlStatus mnnvlutils.Status) bool {
	if !mnnvlStatus.IsNetworkAccelerated() {
		return false
	}

	return sc.computeDomain != nil && sc.computeDomain.Status.Status == constants.ComputeDomainStatusReady && sc.computeDomain.DeletionTimestamp.IsZero()
}

func isWorkloadActive(workload *job.JobWorkload) bool {
	if workload == nil {
		return false
	}
	return !workload.Spec.Active.IsValueInitialized() || workload.Spec.Active.Deref()
}
