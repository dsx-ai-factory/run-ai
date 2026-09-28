/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	apipage "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/page"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
)

const (
	TimerList   = "project_list_timer"
	projectPage = "project_page"
)

type ProjectLister interface {
	List(ctx context.Context, limit, offset int32, filters []string) (*openapi.GetProjects200Response, error)
}

type ProjectService struct {
	projectClient ProjectLister
}

func NewProjectService(projectLister ProjectLister) *ProjectService {
	return &ProjectService{projectClient: projectLister}
}

func (ps *ProjectService) List(txn *transaction.Txn, pageOptions *list.PageOptions, fb *filter.Builder) (*Output, error) {
	var (
		projectList []openapi.Project
		next        *structs.NextToken
		err         error
	)

	if txn == nil {
		return nil, apipage.ErrNilTxn
	}

	ptnId := txn.AddTimer(TimerList)
	defer txn.StopTimer(ptnId)

	projectList, next, err = ps.GetProjects(txn, pageOptions, fb)
	if err != nil {
		return nil, err
	}
	return NewOutput(projectList, next)
}

func (ps *ProjectService) GetProjects(txn *transaction.Txn, options *list.PageOptions, fb *filter.Builder) ([]openapi.Project, *structs.NextToken, error) {
	if options.NoPagination {
		return ps.singlePage(txn, options, fb)
	}
	return ps.multiPages(txn, options, fb)
}

func (ps *ProjectService) singlePage(txn *transaction.Txn, pageOptions *list.PageOptions, fb *filter.Builder) ([]openapi.Project, *structs.NextToken, error) {
	var (
		next *structs.NextToken
	)

	// single page fetch time
	timer := txn.AddTimer(projectPage)
	defer txn.StopTimer(timer)

	page, err := ps.page(txn.GetCtx(), pageOptions, fb)
	if err != nil {
		return nil, nil, err
	}

	// page sanity
	if page == nil {
		return nil, nil, err
	}

	pa := apipage.NewAppender[openapi.Project](int(pageOptions.PageSize))
	err = pa.AddPage(page.Projects)
	if err != nil {
		return nil, nil, err
	}

	if pa.ReachedLimit() {
		next, err = pa.GetNextToken()
		if err != nil {
			return nil, nil, err
		}
	}
	return pa.GetList(), next, nil
}

func (ps *ProjectService) multiPages(txn *transaction.Txn, pageOptions *list.PageOptions, fb *filter.Builder) ([]openapi.Project, *structs.NextToken, error) {
	var (
		projects   []openapi.Project
		current    time.Time
		singleNext *structs.NextToken
		err        error
	)

	pa := apipage.NewAppender[openapi.Project](int(pageOptions.MaxItems))
	for !pa.ReachedLimit() {
		projects, singleNext, err = ps.singlePage(txn, pageOptions, fb)
		if err != nil {
			return nil, nil, err
		}

		// add entries
		err = pa.AddPage(projects)
		if err != nil {
			return nil, nil, err
		}

		// no next, no return founded project list
		if singleNext == nil {
			return pa.GetList(), nil, nil
		}

		// check if current and next CreatedAt values
		err = apipage.CheckNextTimeStampAfterCurrent(current, singleNext.GetTimestamp())
		if err != nil {
			return nil, nil, err
		}

		// update current with next
		current = singleNext.GetTimestamp()
		fb.WithCreatedAt(current)
	}

	// Page Appender Next Token
	paNext, err := pa.GetNextToken()
	if err != nil {
		return nil, nil, err
	}
	// If we reached exactly max items, the page appender will not necessary will have a token.
	if paNext != nil && !paNext.GetTimestamp().IsZero() {
		return pa.GetList(), paNext, nil
	}
	return pa.GetList(), singleNext, nil
}

func (ps *ProjectService) page(ctx context.Context, pageOptions *list.PageOptions, fb *filter.Builder) (*openapi.GetProjects200Response, error) {
	page, err := ps.projectClient.List(ctx, apipage.GetPageSize(pageOptions.MaxItems, pageOptions.PageSize)+1, pageOptions.Offset, fb.StringList())
	if err != nil {
		return nil, err
	}

	return page, nil
}
