/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	"go.uber.org/zap"
	"k8s.io/utils/ptr"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RevisionHandler struct {
	CachedClient  cachedclient.CachedClient
	BackendClient *backendclient.BackendClient
	parser        *parsers.RevisionParser
}

func NewRevisionHandler(cachedClient cachedclient.CachedClient, backendClient *backendclient.BackendClient) *RevisionHandler {
	return &RevisionHandler{CachedClient: cachedClient, BackendClient: backendClient, parser: parsers.NewRevisionParser()}
}

func (h *RevisionHandler) HandleUpdateItem(key client.ObjectKey) error {
	rev := &knv1.Revision{}
	err := h.CachedClient.Get(context.Background(), key, rev)
	if err != nil {
		return err
	}

	return h.HandleRevisionUpdate(rev)
}

func (h *RevisionHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start revision Sync")

	UpdateProjectHandlerWrapper(h.handleUpdateInProject, h.CachedClient, "Revisions")
}

func (h *RevisionHandler) handleUpdateInProject(project *util.ProjectData) {
	projectId, err := util.GetProjectId(project)
	if err != nil {
		return
	}

	namespaceName := project.Namespace
	if namespaceName == "" {
		zap.S().Warnf("Failed getting namespace name for project <%s> while syncing Revisions",
			project.Name)
		return
	}

	// List all revisions in the namespace
	revisions := &knv1.RevisionList{}
	kubeErr := h.CachedClient.List(context.Background(), revisions, &client.ListOptions{Namespace: namespaceName})
	if kubeErr != nil {
		zap.S().Errorf("%+v", kubeErr)
		return
	}

	revisionsInProject, idsToIgnore := h.parseRevisionsForSync(revisions.Items, project)

	request := h.BackendClient.WorkloadsClient.RevisionsAPI.SyncRevisions(h.BackendClient.Context()).
		RevisionsSyncRequest(openapi.RevisionsSyncRequest{
			ProjectId:   projectId,
			Revisions:   revisionsInProject,
			From:        nil, // TODO add pagination
			To:          nil, // TODO add pagination
			IdsToIgnore: idsToIgnore,
		})
	res, err := request.Execute()
	if res != nil {
		_ = res.Body.Close()

	}
	err = h.BackendClient.HandleHttpResponse(res, err, "Sync revisions")
	if err != nil {
		zap.S().Errorf("Had an error while syncing revisions: %+v", err)
	}
}

func (h *RevisionHandler) parseRevisionsForSync(revisions []knv1.Revision, project *util.ProjectData) (
	[]openapi.RevisionSyncFields, []string) {
	parsedRevisions := []openapi.RevisionSyncFields{}
	idsToIgnore := []string{}

	for i := range revisions {
		revision := &revisions[i]
		if !commonconstants.IsManagedInferenceRevision(revision.Labels) {
			// Full sync treats every omitted ID as deleted. Preserve revisions that
			// are outside the current predicate so legacy rows are not tombstoned.
			idsToIgnore = append(idsToIgnore, string(revision.UID))
			continue
		}

		workloadID, err := h.resolveWorkloadID(revision)
		if err != nil {
			zap.S().Warnf("Failed resolving workload ID for Revision %s: %v", revision.UID, err)
			idsToIgnore = append(idsToIgnore, string(revision.UID))
			continue
		}

		parsedRevision, err := h.parser.ParseRevisionToServerScheme(revision, workloadID, project)
		if err != nil {
			zap.S().Errorf(
				"Failed parsing Revision <%s/%s> (uid: %s, workload kind: %q, workload name: %q) to server schema: %v",
				revision.Namespace,
				revision.Name,
				revision.UID,
				revision.Labels[commonconstants.ControlPlaneWorkloadKindLabelForWorkload],
				revision.Labels[commonconstants.ControlPlaneWorkloadNameLabelForWorkload],
				err,
			)
			idsToIgnore = append(idsToIgnore, string(revision.UID))
			continue
		}

		parsedRevisions = append(parsedRevisions, *parsedRevision)
	}

	return parsedRevisions, idsToIgnore
}

func (h *RevisionHandler) HandleRevisionAdd(revision *knv1.Revision) error {
	zap.S().Debugf("Handling Revision <%s/%s> add", revision.Namespace, revision.Name)

	return h.handleRevisionAddUpdate(revision)
}

func (h *RevisionHandler) HandleRevisionUpdate(revision *knv1.Revision) error {
	zap.S().Debugf("Handling Revision <%s/%s> update", revision.Namespace, revision.Name)
	return h.handleRevisionAddUpdate(revision)
}

func (h *RevisionHandler) handleRevisionAddUpdate(revision *knv1.Revision) error {
	if !commonconstants.IsManagedInferenceRevision(revision.Labels) {
		return nil
	}

	workloadID, err := h.resolveWorkloadID(revision)
	if err != nil {
		return err
	}

	project, err := util.GetProjectFromNamespaceName(revision.Namespace, h.CachedClient)
	if err != nil {
		return err
	}

	revisionSyncFields, err := h.parser.ParseRevisionToServerScheme(revision, workloadID, project)
	if err != nil {
		return fmt.Errorf(
			"failed parsing revision <%s/%s> (uid: %s, workload kind: %q, workload name: %q) to server schema: %w",
			revision.Namespace,
			revision.Name,
			revision.UID,
			revision.Labels[commonconstants.ControlPlaneWorkloadKindLabelForWorkload],
			revision.Labels[commonconstants.ControlPlaneWorkloadNameLabelForWorkload],
			err,
		)
	}

	projectId, err := util.GetProjectId(project)
	if err != nil {
		return err
	}

	return h.handleUpsert(*revisionSyncFields, projectId, string(revision.UID))
}

func (h *RevisionHandler) resolveWorkloadID(revision *knv1.Revision) (string, error) {
	if len(revision.OwnerReferences) == 0 {
		return "", fmt.Errorf("revision <%s/%s> does not have an owner",
			revision.Namespace, revision.Name)
	}

	workloadID, err := topowner.GetTopOwnerWorkloadId(revision, h.CachedClient)
	if err != nil {
		return "", fmt.Errorf("failed resolving top owner for Revision <%s/%s>: %w",
			revision.Namespace, revision.Name, err)
	}
	if workloadID == "" {
		return "", fmt.Errorf("resolved empty workload ID for Revision <%s/%s>",
			revision.Namespace, revision.Name)
	}
	return workloadID, nil
}

func (h *RevisionHandler) handleUpsert(revisionSyncFields openapi.RevisionSyncFields, projectId string, revisionId string) error {
	request := h.BackendClient.WorkloadsClient.RevisionsAPI.SyncRevisions(h.BackendClient.Context()).
		RevisionsSyncRequest(openapi.RevisionsSyncRequest{
			ProjectId: projectId,
			ClusterId: ptr.To(config.Config().ClusterUuid),
			Revisions: []openapi.RevisionSyncFields{revisionSyncFields},
			From:      ptr.To(revisionId),
			To:        ptr.To(revisionId),
		})
	res, err := request.Execute()
	if res != nil {
		_ = res.Body.Close()

	}
	return h.BackendClient.HandleHttpResponse(res, err, "Upsert revision")
}

func (h *RevisionHandler) HandleRevisionDelete(object client.Object) error {
	zap.S().Debugf("Handling %s <%s/%s> delete", object.GetObjectKind().GroupVersionKind().Kind, object.GetNamespace(), object.GetName())

	res, err := h.BackendClient.WorkloadsClient.RevisionsAPI.DeleteRevision(
		h.BackendClient.Context(),
		string(object.GetUID()),
	).Execute()
	if res != nil {
		_ = res.Body.Close()

	}
	if res != nil && res.StatusCode == http.StatusNotFound {
		zap.S().Infof("revision %s was already deleted", object.GetUID())
		return nil
	}

	return h.BackendClient.HandleHttpResponse(
		res,
		err,
		fmt.Sprintf("Deleted revision %s", object.GetUID()),
	)
}
