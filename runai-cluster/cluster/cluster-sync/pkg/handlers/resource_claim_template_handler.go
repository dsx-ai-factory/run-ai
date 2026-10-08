/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"sync"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	trackerclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	v1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"go.uber.org/zap"
	resourcev1 "k8s.io/api/resource/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceClaimTemplateHandler reports every observed RCT in the cluster to the
// collapsed /resource-claim-templates/sync endpoint. The control-plane API only
// exposes a single POST for the full cluster inventory (per-item CRUD was
// explicitly ruled out so a payload carrying every RCT spec would not blow up
// the request/response size), so every informer event funnels through the same
// coalesced resync key.
type ResourceClaimTemplateHandler struct {
	backendClient *runaiBackendClient.BackendClient
	cachedReader  client.Reader
}

func NewResourceClaimTemplateHandler(cachedReader client.Reader, backend *runaiBackendClient.BackendClient) *ResourceClaimTemplateHandler {
	return &ResourceClaimTemplateHandler{cachedReader: cachedReader, backendClient: backend}
}

// HandleUpdateItem is the rate-limited callback the controller feeds. The key
// is ignored on purpose: the tracker only accepts a full-set replace, so any
// change triggers a re-sync of every RCT.
func (h *ResourceClaimTemplateHandler) HandleUpdateItem(_ client.ObjectKey) error {
	return h.sync()
}

// HandleSyncEvent is the periodic drift-correction entry point. Any missed
// informer event is caught up here, and the server-side prune on this endpoint
// removes rows for RCTs that no longer exist in the cluster.
func (h *ResourceClaimTemplateHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start syncing ResourceClaimTemplates")

	if err := h.sync(); err != nil {
		zap.S().Errorf("Had an error while syncing ResourceClaimTemplates: %+v", err)
	}
}

func (h *ResourceClaimTemplateHandler) sync() error {
	rcts := &resourcev1.ResourceClaimTemplateList{}
	if err := h.cachedReader.List(context.Background(), rcts); err != nil {
		zap.S().Errorf("Failed listing ResourceClaimTemplates, error: %s", err.Error())
		return err
	}

	// Build a namespace -> project id map so project-scope RCTs can be
	// reported with the control-plane project id (e.g. "4500002") rather
	// than the raw namespace string ("runai-ofer"). An RCT in a namespace
	// that is not Run:ai-managed, or whose project does not expose its id,
	// is skipped as a replica.
	projectIdByNamespace := h.projectIdsByNamespace()

	// Initialize with a non-nil slice so an empty cluster sends [] rather than null.
	items := make([]trackerclient.ResourceClaimTemplate, 0, len(rcts.Items))
	for i := range rcts.Items {
		rct := &rcts.Items[i]

		parsed, err := parsers.ParseResourceClaimTemplateToServerScheme(rct)
		if err != nil {
			zap.S().Warnf("Skipping RCT %s/%s: %v", rct.Namespace, rct.Name, err)
			continue
		}

		if parsed.Scope.Type == auth_client.SCOPETYPE_PROJECT {
			projectId, projectNs := projectIdByNamespace[rct.Namespace]
			if !projectNs {
				continue // not an original project RCT; it's a replica or in an unmanaged ns.
			}
			parsed.CreationRequest.ScopeId = projectId
		} else if rct.Namespace != v1.DefaultNamespace {
			// cluster / department / tenant-scoped templates only ship when
			// they live in the runai control namespace; copies elsewhere are
			// replicas we must not report.
			continue
		}

		items = append(items, parsed.CreationRequest)
	}

	_, res, httpErr := h.backendClient.K8sObjectTrackerClient.SyncResourceClaimTemplates(
		h.backendClient.Context(), config.Config().ClusterUuid, items)
	if res != nil {
		_ = res.Body.Close()
	}

	return h.backendClient.HandleHttpResponse(res, httpErr, "Sync ResourceClaimTemplate data")
}

// projectIdsByNamespace resolves the control-plane project id for every
// Run:ai-managed project namespace. The underlying parseScope only reaches
// labels on the RCT (and so can give us the namespace at best); the control
// plane needs the project id (e.g. "4500002"), which lives on the Project
// object itself. Projects that lack the id label are silently skipped so
// their RCTs look like replicas and are filtered out by the sync loop.
func (h *ResourceClaimTemplateHandler) projectIdsByNamespace() map[string]string {
	projects, err := util.GetAllProjects(h.cachedReader)
	if err != nil {
		zap.S().Warnf("Failed listing projects while syncing RCTs: %v", err)
		return map[string]string{}
	}

	result := make(map[string]string, len(projects))
	for i := range projects {
		project := &projects[i]
		if project.Namespace == "" || project.ProjectId == "" {
			continue
		}
		result[project.Namespace] = project.ProjectId
	}
	return result
}
