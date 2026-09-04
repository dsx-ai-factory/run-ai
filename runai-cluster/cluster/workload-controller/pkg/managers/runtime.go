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
	internalJobs "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/jobs"
	issue2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	workloadcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	workloadconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/resources"
	"github.com/go-logr/logr"
	openshiftV1 "github.com/openshift/api/route/v1"
	core1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// Route
	RouteUrlAlreadyExistConditionReason  = "HostAlreadyClaimed"
	RouteUrlAlreadyExistIssueDescription = "A route with the same url already exist"
	NoIPAddressAssignedToService         = "No IP address is assigned to the service"

	reasonAllocationFailed = "AllocationFailed"
)

type HandlerTask struct {
	ResourceObject client.Object
	Aggr           *resource.Aggregator
	Logger         logr.Logger
	Client         client.Client
}

type resourceHandlersMap map[string]func(task HandlerTask) *issue.Issue

type SubResourcesRuntimeManager struct {
	resourceHandlers resourceHandlersMap
	k8sClient        client.Client
}

func NewSubResourcesRuntimeManager(k8sClient client.Client) *SubResourcesRuntimeManager {
	handlers := make(resourceHandlersMap)
	registerHandlers(handlers)

	return &SubResourcesRuntimeManager{
		resourceHandlers: handlers,
		k8sClient:        k8sClient,
	}
}

func registerHandlers(handlers resourceHandlersMap) {
	handlers[constants.KindRoute] = routeHandler

	handlers[constants.KindService] = serviceHandler

	handlers[constants.KindRunaiJob] = jobHandler
	handlers[constants.KindMpiJob] = jobHandler
	handlers[constants.KindPyTorchJob] = jobHandler
	handlers[constants.KindTFJob] = jobHandler
	handlers[constants.KindXGBoostJob] = jobHandler
	handlers[constants.KindJAXJob] = jobHandler
}

func (m *SubResourcesRuntimeManager) CheckSubResources(ctx context.Context, aggr *resource.Aggregator) []*issue.Issue {
	log := logger.FromContext(ctx)

	workloadName := "nil"
	if aggr.GetWorkload() != nil {
		workloadName = aggr.GetWorkload().Name
	}

	log.V(4).Info("starting to check workload's sub-resources", "workloadName", workloadName)

	issueList := checkSubResourcesGroup(m, aggr.GetExternalNetwork(), aggr, log, m.k8sClient)
	issueList = append(issueList, checkSubResourcesGroup(m, []client.Object{aggr.GetJob()}, aggr, log, m.k8sClient)...)
	issueList = append(issueList, checkSubResourcesGroup(m, aggr.GetServices(), aggr, log, m.k8sClient)...)

	// Uncomment when implemented handlers for types
	/*
		issueList = append(issueList, checkSubResourcesGroup(m, aggr.GetPvcs())...)
		issueList = append(issueList, checkSubResourcesGroup(m, aggr.GetSecret())...)
	*/

	log.V(4).Info(fmt.Sprintf("found %d issues in workload's sub-resources", len(issueList)), "workloadName", workloadName)

	return issueList
}

func checkSubResourcesGroup[T client.Object](m *SubResourcesRuntimeManager, subResources []T, aggr *resource.Aggregator, logger logr.Logger, k8sClient client.Client) []*issue.Issue {
	issueList := make([]*issue.Issue, 0)
	for _, subResource := range subResources {
		if kindHandler, exists := m.resourceHandlers[resources.SafeGetKind(subResource)]; exists {
			iss := kindHandler(HandlerTask{ResourceObject: subResource, Aggr: aggr, Logger: logger, Client: k8sClient})
			if iss != nil {
				issueList = append(issueList, iss)
			}
		}
	}

	return issueList
}

func routeHandler(task HandlerTask) *issue.Issue {
	route, ok := task.ResourceObject.(*openshiftV1.Route)
	if !ok {
		task.Logger.V(3).Info(fmt.Sprintf("kind mismatch for %s. expected: %s, actual: %s", task.ResourceObject.GetName(), constants.KindRoute, resources.SafeGetKind(task.ResourceObject)))
		return nil
	}

	if len(route.Status.Ingress) > 0 {
		ingress := route.Status.Ingress[0]
		for _, condition := range ingress.Conditions {
			if condition.Type == openshiftV1.RouteAdmitted &&
				condition.Status == core1.ConditionFalse &&
				condition.Reason == RouteUrlAlreadyExistConditionReason {
				return issue.New(route, errors.New(condition.Message), true, RouteUrlAlreadyExistIssueDescription, &issue2.ComputeResourceGrouper{})
			}
		}
	}

	return nil
}

func serviceHandler(task HandlerTask) *issue.Issue {
	svc, ok := task.ResourceObject.(*core1.Service)
	if !ok {
		task.Logger.V(3).Info(fmt.Sprintf("kind mismatch for %s. expected: %s, actual: %s", task.ResourceObject.GetName(), constants.KindService, resources.SafeGetKind(task.ResourceObject)))
		return nil
	}

	if svc.Spec.Type != core1.ServiceTypeLoadBalancer {
		return nil
	}

	if len(svc.Status.LoadBalancer.Ingress) != 0 {
		return nil
	}

	// List all events of this service in that namespace
	evList := &core1.EventList{}
	err := task.Client.List(context.Background(), evList,
		client.InNamespace(task.ResourceObject.GetNamespace()),
		client.MatchingFields{workloadcommon.InvolvedObjectUidField: string(svc.UID)})
	if err != nil {
		task.Logger.V(4).Error(err, fmt.Sprintf("failed to list events %v", err))
		return nil
	}

	for _, e := range evList.Items {
		if e.Reason == reasonAllocationFailed {
			return issue.New(svc, errors.New(NoIPAddressAssignedToService), true, NoIPAddressAssignedToService, nil)
		}
	}

	task.Logger.V(4).Error(err, fmt.Sprintf("scanned %d events, %s was not found", len(evList.Items), reasonAllocationFailed))

	return nil
}

func jobHandler(task HandlerTask) *issue.Issue {
	jobAdapter, err := getJobAdapter(task.ResourceObject, task.Logger)
	if err != nil {
		return issue2.NewIssue(task.Aggr.GetWorkload(), err, false, "")
	}

	failedInfo := jobAdapter.GetStatus().GetFailedInfo()
	if failedInfo != workloadconstants.EmptyStr {
		return issue2.NewIssue(task.ResourceObject, errors.New(failedInfo), true, "")
	}

	return nil
}

func getJobAdapter(jobObject client.Object, logger logr.Logger) (jobs.JobAdapter, error) {
	// passing nil k8s client since the adapter is not expected to make any calls to the cluster,
	// just to allow access to the job's status in an interfaced matter
	adapterFactory, err := internalJobs.GetJobAdapterFactory(jobObject, nil)
	if err != nil {
		logger.V(4).Error(err, "failed to create job adapter factory")
		return nil, err
	}

	adapter, err := adapterFactory.CreateJobAdapter(jobObject)
	if err != nil {
		logger.V(4).Error(err, "failed to create job adapter")
		return nil, err
	}

	return adapter, nil
}
