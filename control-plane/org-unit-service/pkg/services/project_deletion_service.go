/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	async_events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/versions"
	datavolumes_service_client "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
	workloads_client "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"k8s.io/utils/ptr"
)

type ProjectDeletionService struct {
	clusterClient     cluster_service_client.IClusterExternalClient
	workloadsClient   workloads_client.IWorkloadsClient
	datavolumesClient datavolumes_service_client.IDatavolumesClient
}

func NewProjectDeletionService(
	clusterClient cluster_service_client.IClusterExternalClient,
	workloadsClient workloads_client.IWorkloadsClient,
	datavolumesClient datavolumes_service_client.IDatavolumesClient,
) ProjectDeletionService {
	return ProjectDeletionService{
		clusterClient:     clusterClient,
		workloadsClient:   workloadsClient,
		datavolumesClient: datavolumesClient,
	}
}

const ProjectDeletionMinVersion = "2.20"

func (pds *ProjectDeletionService) DeleteProject(ctx context.Context, projectId int64, tenantId string, subjectId *string, deletionType *api.DeletionTypes, force *bool,
) (projectName string, departmentId int64, syncDeletion bool, err error) {
	projectMeta, err := dao.GetProjectById(ctx, projectId, tenantId, models.MetaDataVerbosity)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get project metadata: %v", err)
		return "", -1, false, err
	}
	if projectMeta.Phase != nil && *projectMeta.Phase == string(api.PROJECTPHASE_DELETING) && (force == nil || !*force) {
		return projectMeta.Name, convutils.Deref(projectMeta.ParentID, -1), false, &ProjectIsAlreadyDeletingErr{}
	}

	if deletionType == nil || (force != nil && *force) {
		return deleteProjectInDb(ctx, projectMeta, subjectId, deletionType, force)
	}
	departmentId = convutils.Deref(projectMeta.ParentID, -1)
	clusterSupport, err := pds.isProjectDeletionSupportedInCluster(ctx, projectMeta.ClusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to check cluster support of gracefull project deletion. err: %v", err)
		clusterSupport = false
	}

	switch *deletionType {
	case api.DELETIONTYPES_ORPHAN:
		if clusterSupport {
			projectName, syncDeletion, err := setProjectAsDeleting(ctx, projectMeta, subjectId, deletionType, force)
			return projectName, departmentId, syncDeletion, err
		}
		return deleteProjectInDb(ctx, projectMeta, subjectId, deletionType, force)
	case api.DELETIONTYPES_CASCADE:
		if clusterSupport {
			projectName, syncDeletion, err := setProjectAsDeleting(ctx, projectMeta, subjectId, deletionType, force)
			return projectName, departmentId, syncDeletion, err
		}
		return projectMeta.Name, departmentId, false, &ClusterDoesNotSupportProjectDeletionErr{}
	case api.DELETIONTYPES_IF_EMPTY:
		err = pds.validateProjectEmpty(ctx, projectMeta)
		if err != nil {
			return projectMeta.Name, departmentId, false, err
		}

		return deleteProjectInDb(ctx, projectMeta, subjectId, deletionType, force)
	}

	return deleteProjectInDb(ctx, projectMeta, subjectId, deletionType, force)
}

func (pds *ProjectDeletionService) DeleteProjectSyncFromCluster(ctx context.Context, projectId int64, tenantId string) (err error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Got a delete project %d request from the cluster.", projectId)

	projectMeta, err := dao.GetProjectById(ctx, projectId, tenantId, models.MetaDataVerbosity)
	if err != nil {
		if err.Error() == runaihttp.ErrTextNotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Got a delete project %d request from the cluster, but it is already deleted.", projectId)
			return nil
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get project metadata: %v", err)
		return err
	}

	if projectMeta.Phase == nil || *projectMeta.Phase != string(api.PROJECTPHASE_DELETING) {
		return fmt.Errorf("cluster can not delete a project that is not in Deleting phase. Current phase: %s", ptr.Deref(projectMeta.Phase, ""))
	}

	_, _, _, err = deleteProjectInDb(ctx, projectMeta, nil, nil, nil)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project %d named %s: %v", projectId, projectMeta.Name, err)
		return err
	}

	return nil
}

func (pds *ProjectDeletionService) validateProjectEmpty(ctx context.Context, projectMeta *models.Project) error {
	notEmptyErr := &ProjectIsNotEmptyErr{}

	err := pds.validateNoWorkloads(ctx, projectMeta, notEmptyErr)
	if err != nil {
		return err
	}

	err = pds.validateNoDvs(ctx, projectMeta, notEmptyErr)
	if err != nil {
		return err
	}

	// TODO: validate more resources

	if notEmptyErr.IsEmpty() {
		return nil
	}

	return notEmptyErr
}

func (pds *ProjectDeletionService) validateNoDvs(ctx context.Context, projectMeta *models.Project, notEmptyErr *ProjectIsNotEmptyErr) error {
	dvs, err := pds.datavolumesClient.GetDatavolumes(
		ctx,
		openapi.DATAVOLUMEREQUESTTYPE_ORIGINATED,
		"",
		100,
		0,
		client_utils.NewFilterBuilder().AddFilter("projectId", client_utils.EqualOp, fmt.Sprintf("%d", projectMeta.ID)).Get(),
		"",
		"",
	)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error from datavolumes service. error: %v", err)
		return err
	}
	for _, dv := range dvs.Datavolumes {
		notEmptyErr.Datavolumes = append(notEmptyErr.Datavolumes, dv.Name)
	}

	return nil
}

func (pds *ProjectDeletionService) validateNoWorkloads(ctx context.Context, projectMeta *models.Project, notEmptyErr *ProjectIsNotEmptyErr) error {
	workloads, err := pds.workloadsClient.GetWorkloads(
		ctx,
		100,
		0,
		client_utils.NewFilterBuilder().AddFilter("projectId", client_utils.EqualOp, fmt.Sprintf("%d", projectMeta.ID)).Get(),
		false,
		"",
		"",
	)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error from workloads service. error: %v", err)
		return err
	}
	for _, workload := range workloads.Workloads {
		notEmptyErr.Workloads = append(notEmptyErr.Workloads, workload.Name)
	}

	return nil
}

func (pds *ProjectDeletionService) isProjectDeletionSupportedInCluster(ctx context.Context, clusterId uuid.UUID) (bool, error) {
	clusterVersion, err := GetClusterVersion(ctx, clusterId, pds.clusterClient)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get cluster version. error: %v", err.Error())
		return false, err
	}
	clusterSupport, err := versions.IsVersionSupported(clusterVersion, ProjectDeletionMinVersion)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get cluster version. error: %v", err.Error())
		return false, err
	}
	return clusterSupport, nil
}

func (pds *ProjectDeletionService) DeleteProjectsSyncFromCluster(ctx context.Context, tenantId string, clusterUuid uuid.UUID, projectsInCluster []api.ProjectInSyncRequest) error {
	clusterProjectsInDb, err := dao.GetProjectsByCluster(ctx, tenantId, clusterUuid, nil, models.MetaDataVerbosity)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list cluster %s projectsInCluster", clusterUuid)
		return err
	}

	projectForDeletion := []*models.Project{}
	for _, projInDb := range clusterProjectsInDb {
		if slices.ContainsFunc(projectsInCluster, func(projectInCluster api.ProjectInSyncRequest) bool {
			return projectInCluster.Id == fmt.Sprintf("%d", projInDb.ID)
		}) {
			continue
		}
		if projInDb.Phase == nil || *projInDb.Phase != string(api.PROJECTPHASE_DELETING) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("project %s is in the DB but not in cluster %s, and its phase is %s, so not deleting", projInDb.Name, clusterUuid, ptr.Deref(projInDb.Phase, ""))

			continue
		}

		projectForDeletion = append(projectForDeletion, projInDb)
	}

	return deleteProjectsInDb(ctx, projectForDeletion)
}

func deleteProjectsInDb(ctx context.Context, projects []*models.Project) (err error) {
	deletedProjects, err := dao.DeleteProjectsInDb(ctx, projects)

	for _, project := range deletedProjects {
		publishProjectDeletedEvent(ctx, project)
	}

	return err
}

func deleteProjectInDb(ctx context.Context, project *models.Project, subjectId *string, deletionType *api.DeletionTypes, force *bool) (projectName string, departmentId int64, syncDeletion bool, err error) {
	projectName, departmentId, err = dao.DeleteProjectInDb(ctx, project, subjectId, ptr.Deref(deletionType, api.DELETIONTYPES_ORPHAN), ptr.Deref(force, false))

	publishProjectDeletedEvent(ctx, project)
	return projectName, departmentId, true, err
}

func setProjectAsDeleting(ctx context.Context, project *models.Project, subjectId *string, deletionType *api.DeletionTypes, force *bool) (projectName string, syncDeletion bool, err error) {
	projectName, err = dao.SetProjectAsDeleting(ctx, project, subjectId, ptr.Deref(deletionType, api.DELETIONTYPES_ORPHAN), ptr.Deref(force, false))

	publishProjectDeletionRequestEvent(ctx, project)
	return projectName, false, err
}

func publishProjectDeletedEvent(ctx context.Context, projectMeta *models.Project) {
	err := runaipubsub.Publish[async_events.ProjectDeleted](
		ctx,
		async_events.NewProjectDeleted(
			projectMeta.ClusterId.String(),
			fmt.Sprintf("%d", projectMeta.ID),
			projectMeta.Name,
			projectMeta.DeletionType != nil && *projectMeta.DeletionType == api.DELETIONTYPES_ORPHAN,
		))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to publish project %s with id %d deletion event. error: %v", projectMeta.Name, projectMeta.ID, err)
	}
}

func publishProjectDeletionRequestEvent(ctx context.Context, projectMeta *models.Project) error {
	err := runaipubsub.Publish[async_events.ProjectDeleteRequest](
		ctx,
		async_events.NewProjectDeleteRequest(
			projectMeta.ClusterId.String(),
			fmt.Sprintf("%d", projectMeta.ID),
			projectMeta.Name,
			projectMeta.DeletionType != nil && *projectMeta.DeletionType == api.DELETIONTYPES_ORPHAN,
		))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to publish project %s with id %d deletion request event. error: %v", projectMeta.Name, projectMeta.ID, err)
	}

	return err
}
