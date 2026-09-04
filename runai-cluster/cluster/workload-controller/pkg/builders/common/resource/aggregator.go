/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource

import (
	"fmt"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/issue"
	api "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils"
	openshiftV1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// IssuesMap holds the map of the object kind and name, mapped to issue
// eg:
// [Service:Name] --> Issue
type IssuesMap map[string]*issue.Issue

type subResource struct {
	job           client.Object
	services      []*corev1.Service
	ingress       []*networkingv1.Ingress
	routes        []*openshiftV1.Route
	httpRoutes    []*gatewayv1.HTTPRoute
	pvcs          []*corev1.PersistentVolumeClaim
	secrets       []*corev1.Secret
	computeDomain *nvapi.ComputeDomain
}

type Aggregator struct {
	workload    *jobs.JobWorkload
	issueMap    IssuesMap
	hasCritical bool
	current     subResource
	desiredList []client.Object
}

func NewAggregator(workload *jobs.JobWorkload) *Aggregator {
	return &Aggregator{
		workload: workload,
		issueMap: make(IssuesMap),
	}
}

func (a *Aggregator) AddIssues(issues ...*issue.Issue) {
	for _, iss := range issues {
		if iss == nil {
			continue
		}
		key := generateKey(iss.Kind(), iss.Name())
		if old, ok := a.issueMap[key]; !ok {
			a.updateIssueMap(key, iss)
		} else if shouldReplaceIssues(old, iss) {
			a.updateIssueMap(key, iss)
		}
	}
}

func shouldReplaceIssues(exist *issue.Issue, iss *issue.Issue) bool {
	// if exist is a critical issue, don't replace
	if exist.IsCritical() {
		return false
	}

	// if the new issue is critical, replace
	if iss.IsCritical() {
		return true
	}

	// any other use case can be ignored
	return false
}

func generateKey(kind string, name string) string {
	return fmt.Sprintf("%s:%s", kind, name)
}

func (a *Aggregator) HasIssues() bool {
	return len(a.issueMap) != 0
}

func (a *Aggregator) HasCritical() bool {
	return a.hasCritical
}

func (a *Aggregator) GetIssueStatus() *api.Issues {
	if !a.HasIssues() {
		return nil
	}
	var statusMap api.Issues
	for _, iss := range a.issueMap {
		group := iss.GetGroup()
		apiIssue := api.Issue{
			Kind:        iss.Kind(),
			Name:        iss.Name(),
			Description: iss.Message(),
		}
		switch group {
		case issue.Network:
			statusMap.Network = append(statusMap.Network, apiIssue)
		case issue.Compute:
			statusMap.Compute = append(statusMap.Compute, apiIssue)
		case issue.Storage:
			statusMap.Storage = append(statusMap.Storage, apiIssue)
		default:
			statusMap.Internal = append(statusMap.Internal, apiIssue)
		}
	}
	return &statusMap
}

func (a *Aggregator) GetIssuesList() []*issue.Issue {
	var issueList []*issue.Issue
	for _, iss := range a.issueMap {
		issueList = append(issueList, iss)
	}
	return issueList
}

func (a *Aggregator) GetIssueGroups() []issue.Group {
	groups := make(map[issue.Group]bool)
	for _, iss := range a.issueMap {
		groups[iss.GetGroup()] = true
	}
	return utils.GetMapKeys(groups)
}

func (a *Aggregator) GetIssue(resourceKind, name string) *issue.Issue {
	key := generateKey(resourceKind, name)
	if iss, ok := a.issueMap[key]; ok {
		return iss
	}
	return nil
}

func (a *Aggregator) GetWorkload() *jobs.JobWorkload {
	return a.workload
}

func (a *Aggregator) SetJob(object client.Object) {
	a.current.job = object
}

func (a *Aggregator) GetJob() client.Object {
	return a.current.job
}

func (a *Aggregator) SetComputeDomain(computeDomain *nvapi.ComputeDomain) {
	a.current.computeDomain = computeDomain
}

func (a *Aggregator) GetComputeDomain() *nvapi.ComputeDomain {
	return a.current.computeDomain
}

func (a *Aggregator) AddServices(service ...*corev1.Service) {
	a.current.services = append(a.current.services, service...)
}

func (a *Aggregator) GetServices() []*corev1.Service {
	return a.current.services
}

func (a *Aggregator) AddIngresses(ingress ...*networkingv1.Ingress) {
	a.current.ingress = append(a.current.ingress, ingress...)
}

func (a *Aggregator) AddRoutes(route ...*openshiftV1.Route) {
	a.current.routes = append(a.current.routes, route...)
}

func (a *Aggregator) AddHTTPRoutes(httpRoute ...*gatewayv1.HTTPRoute) {
	a.current.httpRoutes = append(a.current.httpRoutes, httpRoute...)
}

func (a *Aggregator) GetHTTPRoutes() []*gatewayv1.HTTPRoute {
	return a.current.httpRoutes
}

func (a *Aggregator) AddPvcs(pvc ...*corev1.PersistentVolumeClaim) {
	a.current.pvcs = append(a.current.pvcs, pvc...)
}

func (a *Aggregator) GetPvcs() []*corev1.PersistentVolumeClaim {
	return a.current.pvcs
}

func (a *Aggregator) AddSecrets(secret ...*corev1.Secret) {
	a.current.secrets = append(a.current.secrets, secret...)
}

func (a *Aggregator) GetSecrets() []*corev1.Secret {
	return a.current.secrets
}

func (a *Aggregator) GetExternalNetwork() []client.Object {
	var clientList []client.Object
	for _, ingress := range a.current.ingress {
		clientList = append(clientList, ingress)
	}
	for _, route := range a.current.routes {
		clientList = append(clientList, route)
	}
	for _, httpRoute := range a.current.httpRoutes {
		clientList = append(clientList, httpRoute)
	}
	return clientList
}

func (a *Aggregator) AddDesired(objList ...client.Object) {
	a.desiredList = append(a.desiredList, objList...)
}

func (a *Aggregator) GetDesiredList() []client.Object {
	return a.desiredList
}

func (a *Aggregator) updateResult(iss *issue.Issue) {
	if a.hasCritical {
		// if hasCritical is true result cannot be mutated
		return
	}

	// if issue is critical and controller decide to fail workload, no need to requeue (termination will trigger reconcile).
	// if issue is critical but controller decide not to fail workload, controller will requeue any way.
	if iss.IsCritical() {
		a.hasCritical = true
	}
}

func (a *Aggregator) updateIssueMap(key string, iss *issue.Issue) {
	a.issueMap[key] = iss
	a.updateResult(iss)
}
