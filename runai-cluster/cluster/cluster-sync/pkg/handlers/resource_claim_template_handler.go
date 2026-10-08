/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
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

// ResourceClaimTemplateHandler mirrors the DeviceClass pattern: informer events
// drive per-item POST/PUT/DELETE for low-latency updates, while HandleSyncEvent
// performs a periodic full-set reconciliation that upserts the complete cluster
// inventory and prunes server-side rows that no longer exist in the cluster.
//
// Scope resolution matters here but not for DeviceClass because RCTs are
// namespaced and may be project / department / cluster-scoped. Replicas
// (cluster/department templates copied into project namespaces, project
// templates in unmanaged namespaces, project templates whose Project does not
// expose its id label) are filtered out of every path - per-item and sync -
// so the tracker never sees them.
type ResourceClaimTemplateHandler struct {
	backendClient *runaiBackendClient.BackendClient
	cachedReader  client.Reader
}

func NewResourceClaimTemplateHandler(cachedReader client.Reader, backend *runaiBackendClient.BackendClient) *ResourceClaimTemplateHandler {
	return &ResourceClaimTemplateHandler{cachedReader: cachedReader, backendClient: backend}
}

// HandleUpdateItem is the rate-limited callback used for update events. It
// re-reads the RCT from the cache so the request carries the latest observed
// state rather than whichever revision happened to be queued when the event
// fired.
func (h *ResourceClaimTemplateHandler) HandleUpdateItem(key client.ObjectKey) error {
	rct := &resourcev1.ResourceClaimTemplate{}
	if err := h.cachedReader.Get(context.Background(), key, rct); err != nil {
		return fmt.Errorf("get RCT %s: %w", key, err)
	}
	return h.HandleResourceClaimTemplateUpdate(rct)
}

// HandleResourceClaimTemplateAdd issues the per-item POST for a newly observed
// RCT. Replicas and unresolvable scopes return a nil error after a debug log -
// they must never reach the tracker.
func (h *ResourceClaimTemplateHandler) HandleResourceClaimTemplateAdd(rct *resourcev1.ResourceClaimTemplate) error {
	parsed, skip, err := h.parseForSingleItem(rct)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	zap.S().Infof("Adding ResourceClaimTemplate < Namespace: %s, Name: %s, Scope: %s >",
		rct.Namespace, rct.Name, parsed.ScopeType)

	_, res, httpErr := h.backendClient.K8sObjectTrackerClient.CreateResourceClaimTemplate(
		h.backendClient.Context(), config.Config().ClusterUuid, *parsed)
	if res != nil {
		_ = res.Body.Close()
	}
	return h.backendClient.HandleHttpResponse(res, httpErr, fmt.Sprintf("Added ResourceClaimTemplate %s/%s", rct.Namespace, rct.Name))
}

// HandleResourceClaimTemplateUpdate issues the per-item PUT. Same scope / replica
// filtering as the Add path.
func (h *ResourceClaimTemplateHandler) HandleResourceClaimTemplateUpdate(rct *resourcev1.ResourceClaimTemplate) error {
	parsed, skip, err := h.parseForSingleItem(rct)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	zap.S().Debugf("Updating ResourceClaimTemplate < Namespace: %s, Name: %s, Scope: %s >",
		rct.Namespace, rct.Name, parsed.ScopeType)

	_, res, httpErr := h.backendClient.K8sObjectTrackerClient.UpdateResourceClaimTemplate(
		h.backendClient.Context(), config.Config().ClusterUuid, *parsed)
	if res != nil {
		_ = res.Body.Close()
	}
	return h.backendClient.HandleHttpResponse(res, httpErr, fmt.Sprintf("Updated ResourceClaimTemplate %s/%s", rct.Namespace, rct.Name))
}

// HandleResourceClaimTemplateDelete issues the per-item DELETE. The identity
// (clusterId, scopeType, scopeId, name) travels in the query string. If the
// RCT was skipped on Add (replica / unresolvable project), the DELETE is also
// skipped because the server never stored it in the first place.
func (h *ResourceClaimTemplateHandler) HandleResourceClaimTemplateDelete(rct *resourcev1.ResourceClaimTemplate) error {
	parsed, skip, err := h.parseForSingleItem(rct)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	zap.S().Debugf("Deleting ResourceClaimTemplate < Namespace: %s, Name: %s, Scope: %s >",
		rct.Namespace, rct.Name, parsed.ScopeType)

	res, httpErr := h.backendClient.K8sObjectTrackerClient.DeleteResourceClaimTemplate(
		h.backendClient.Context(), config.Config().ClusterUuid, parsed.ScopeType, parsed.ScopeId, parsed.Name)
	if res != nil {
		_ = res.Body.Close()
	}
	return h.backendClient.HandleHttpResponse(res, httpErr, fmt.Sprintf("Deleted ResourceClaimTemplate %s/%s", rct.Namespace, rct.Name))
}

// HandleSyncEvent is the periodic drift-correction entry point. It lists every
// RCT the agent can see and ships the full inventory to /sync; the server
// upserts all and prunes anything no longer in the list. This is the only
// path that can delete rows for RCTs whose per-item DeleteFunc was missed
// (informer backlog / agent restart).
func (h *ResourceClaimTemplateHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start syncing ResourceClaimTemplates")

	rcts := &resourcev1.ResourceClaimTemplateList{}
	if err := h.cachedReader.List(context.Background(), rcts); err != nil {
		zap.S().Errorf("Failed listing ResourceClaimTemplates, error: %s", err.Error())
		return
	}

	// Build a namespace -> project id map once per sync so project-scope
	// RCTs can carry the control-plane project id (e.g. "4500002") rather
	// than the raw namespace. The per-item CRUD path uses a lighter
	// per-item lookup; both must agree on which RCTs are replicas.
	//
	// If listing projects fails we MUST NOT fall back to an empty map:
	// every project-scope RCT would then look unresolvable and be dropped
	// from the payload, and the server-side prune would then delete all
	// project-scope rows on our behalf. Abort the sync instead so the
	// next tick retries from a consistent view.
	projectIdByNamespace, err := h.projectIdsByNamespace()
	if err != nil {
		zap.S().Errorf("Aborting ResourceClaimTemplate sync: %v",
			fmt.Errorf("list projects: %w", err))
		return
	}

	items := make([]trackerclient.ResourceClaimTemplate, 0, len(rcts.Items))
	for i := range rcts.Items {
		rct := &rcts.Items[i]

		parsed, skip, err := h.resolveForServer(rct, func(ns string) (string, bool) {
			id, ok := projectIdByNamespace[ns]
			return id, ok
		})
		if err != nil {
			// Same correctness constraint as the GetAllProjects
			// failure above: dropping a single RCT from a /sync
			// payload is semantically "the cluster no longer has
			// it", so the server-side prune would delete its row.
			// Abort instead and let the next tick retry.
			zap.S().Errorf("Aborting ResourceClaimTemplate sync: failed to resolve %s/%s: %v",
				rct.Namespace, rct.Name, err)
			return
		}
		if skip {
			continue
		}
		items = append(items, *parsed)
	}

	_, res, httpErr := h.backendClient.K8sObjectTrackerClient.SyncResourceClaimTemplates(
		h.backendClient.Context(), config.Config().ClusterUuid, items)
	if res != nil {
		_ = res.Body.Close()
	}

	if err := h.backendClient.HandleHttpResponse(res, httpErr, "Sync ResourceClaimTemplate data"); err != nil {
		zap.S().Errorf("Had an error while syncing ResourceClaimTemplates: %+v", err)
	}
}

// parseForSingleItem resolves the server-shaped RCT for the per-item path.
// Returns (payload, skip, err): a non-nil error aborts the caller; skip=true
// means the RCT is a replica or otherwise unreportable and must be dropped
// silently so the tracker is never asked to touch it.
func (h *ResourceClaimTemplateHandler) parseForSingleItem(rct *resourcev1.ResourceClaimTemplate) (*trackerclient.ResourceClaimTemplate, bool, error) {
	return h.resolveForServer(rct, func(ns string) (string, bool) {
		// Mirror what the shared handlers do (pod/pvc/revision): look
		// the project up via the namespace label chain. If anything is
		// missing (namespace not Run:ai managed, Project without id)
		// the RCT is treated as a replica.
		id, err := util.GetProjectIdFromNamespaceName(ns, h.cachedReader)
		if err != nil || id == "" {
			return "", false
		}
		return id, true
	})
}

// resolveForServer is the shared scope / origin / project-id resolver used by
// both the per-item and the sync paths. The resolveProjectId callback is the
// one place the two paths differ: per-item does a single namespace lookup,
// sync does one bulk list and consults the resulting map.
func (h *ResourceClaimTemplateHandler) resolveForServer(
	rct *resourcev1.ResourceClaimTemplate,
	resolveProjectId func(namespace string) (string, bool),
) (*trackerclient.ResourceClaimTemplate, bool, error) {
	parsed, err := parsers.ParseResourceClaimTemplateToServerScheme(rct)
	if err != nil {
		return nil, false, fmt.Errorf("parse RCT %s/%s: %w", rct.Namespace, rct.Name, err)
	}

	if parsed.Scope.Type == auth_client.SCOPETYPE_PROJECT {
		projectId, ok := resolveProjectId(rct.Namespace)
		if !ok {
			return nil, true, nil
		}
		parsed.CreationRequest.ScopeId = projectId
	} else if rct.Namespace != v1.DefaultNamespace {
		// cluster / department / tenant-scoped templates only ship when
		// they live in the runai control namespace; copies elsewhere
		// are replicas the control plane must not see.
		return nil, true, nil
	}

	return &parsed.CreationRequest, false, nil
}

// projectIdsByNamespace resolves the control-plane project id for every
// Run:ai-managed project namespace. Projects whose id label is empty are
// skipped so their RCTs look like replicas and are filtered out above.
//
// The error from GetAllProjects is propagated rather than swallowed: an
// empty map here would make every project-scope RCT look like a replica and
// trigger a server-side prune of all project-scope rows via the next /sync
// call. The caller is responsible for aborting the sync in that case.
func (h *ResourceClaimTemplateHandler) projectIdsByNamespace() (map[string]string, error) {
	projects, err := util.GetAllProjects(h.cachedReader)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	result := make(map[string]string, len(projects))
	for i := range projects {
		project := &projects[i]
		if project.Namespace == "" || project.ProjectId == "" {
			continue
		}
		result[project.Namespace] = project.ProjectId
	}
	return result, nil
}
