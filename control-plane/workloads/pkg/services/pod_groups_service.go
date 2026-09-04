/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"net/http"
)

type PodGroupsService struct {
	WorkloadsDao *dao.WorkloadsDao
}

func NewPodGroupsService() api.PodGroupsAPIServicer {
	conf := runai_config.Get[*config.ServerConfig]()
	return &PodGroupsService{
		WorkloadsDao: dao.NewWorkloadsDao(conf),
	}
}

func (p PodGroupsService) SyncPodGroups(ctx context.Context, request api.PodGroupsSyncRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting pod groups sync")

	children := []models.WorkloadChild{}
	revisionChildren := []models.RevisionChild{}
	passedIds := []string{}
	failedIds := []string{}
	for _, pg := range request.PodGroups {
		parseId, err := uuid.Parse(pg.Id)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse pod group id %s", pg.Id)
			failedIds = append(failedIds, pg.Id)
			continue
		}

		parseWorkloadId, err := uuid.Parse(pg.WorkloadId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse workload id %s of pod group %s", pg.WorkloadId, pg.Name)
			failedIds = append(failedIds, pg.Id)
			continue
		}

		passedIds = append(passedIds, pg.Id)
		children = append(children, models.WorkloadChild{
			WorkloadID: parseWorkloadId,
			ChildID:    parseId,
			ChildType:  "PodGroup",
		})

		if pg.RevisionId != nil {
			parseRevisionId, err := uuid.Parse(*pg.RevisionId)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse revision id %s of pod group %s", *pg.RevisionId, pg.Name)
				failedIds = append(failedIds, pg.Id)
				continue
			}

			revisionChildren = append(revisionChildren, models.RevisionChild{
				RevisionID: parseRevisionId,
				ChildID:    parseId,
				ChildType:  "PodGroup",
			})
		}
	}

	apiErr := p.WorkloadsDao.DbUpsertWorkloadsChildren(ctx, children, []string{})
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	apiErr = dao.DbUpsertRevisionsChildren(ctx, revisionChildren, []string{})
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	if len(failedIds) != 0 {
		return api.Response(http.StatusMultiStatus, api.SyncPodGroups207Response{
			SuccessfullySyncedPodGroups: passedIds,
			FailedPodGroups:             failedIds,
		}), nil
	}

	return NoContentResponse(), nil
}
