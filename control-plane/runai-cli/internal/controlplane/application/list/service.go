/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"context"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	apipage "github.com/run-ai/runai/control-plane/runai-cli/internal/api/page"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	timerList      = "application_list_timer"
	timerPageFetch = "application_page"
	timerGetByName = "application_get_by_name_timer"

	getByNameLimit = 2
)

type AIApplicationLister interface {
	List(ctx context.Context, limit, offset int32, filters []string) (*openapi.ListApplications200Response, error)
}

type AIApplicationDeleter interface {
	Delete(ctx context.Context, aiApplicationId string) error
}

type AIApplicationCreator interface {
	Create(ctx context.Context, request openapi.AIApplicationCreationRequest) (*openapi.AIApplication, error)
}

type AIApplicationService struct {
	client AIApplicationLister
}

func NewAIApplicationService(lister AIApplicationLister) *AIApplicationService {
	return &AIApplicationService{client: lister}
}

func (s *AIApplicationService) List(txn *transaction.Txn, pageOptions *list.PageOptions, fb *filter.Builder) (*Output, error) {
	if txn == nil {
		return nil, apipage.ErrNilTxn
	}

	ptnId := txn.AddTimer(timerList)
	defer txn.StopTimer(ptnId)

	apps, next, err := s.getApplications(txn, pageOptions, fb)
	if err != nil {
		return nil, err
	}
	return NewOutput(apps, next)
}

// GetByName fetches a single application matching the exact name within the
// given filter scope. Returns nil (without an error) when no match exists.
func (s *AIApplicationService) GetByName(txn *transaction.Txn, fb *filter.Builder, name string) (*openapi.AIApplication, error) {
	if txn == nil {
		return nil, apipage.ErrNilTxn
	}

	ptnId := txn.AddTimer(timerGetByName)
	defer txn.StopTimer(ptnId)

	page, err := s.client.List(txn.GetCtx(), getByNameLimit, 0, fb.StringList())
	if err != nil {
		return nil, err
	}
	if page == nil {
		return nil, nil
	}

	for i := range page.Applications {
		if page.Applications[i].GetName() == name {
			return &page.Applications[i], nil
		}
	}
	return nil, nil
}

func (s *AIApplicationService) getApplications(txn *transaction.Txn, options *list.PageOptions, fb *filter.Builder) ([]openapi.AIApplication, *structs.NextToken, error) {
	if options.NoPagination {
		return s.singlePage(txn, options, fb)
	}
	return s.multiPages(txn, options, fb)
}

func (s *AIApplicationService) singlePage(txn *transaction.Txn, pageOptions *list.PageOptions, fb *filter.Builder) ([]openapi.AIApplication, *structs.NextToken, error) {
	timer := txn.AddTimer(timerPageFetch)
	defer txn.StopTimer(timer)

	page, err := s.page(txn.GetCtx(), pageOptions, fb)
	if err != nil {
		return nil, nil, err
	}
	if page == nil {
		return nil, nil, nil
	}

	pa := apipage.NewAppender[openapi.AIApplication](int(pageOptions.PageSize))
	if err = pa.AddPage(page.Applications); err != nil {
		return nil, nil, err
	}

	var next *structs.NextToken
	if pa.ReachedLimit() {
		next, err = pa.GetNextToken()
		if err != nil {
			return nil, nil, err
		}
	}
	return pa.GetList(), next, nil
}

func (s *AIApplicationService) multiPages(txn *transaction.Txn, pageOptions *list.PageOptions, fb *filter.Builder) ([]openapi.AIApplication, *structs.NextToken, error) {
	var (
		apps       []openapi.AIApplication
		current    time.Time
		singleNext *structs.NextToken
		err        error
	)

	pa := apipage.NewAppender[openapi.AIApplication](int(pageOptions.MaxItems))
	for !pa.ReachedLimit() {
		apps, singleNext, err = s.singlePage(txn, pageOptions, fb)
		if err != nil {
			return nil, nil, err
		}

		if err = pa.AddPage(apps); err != nil {
			return nil, nil, err
		}

		if singleNext == nil {
			return pa.GetList(), nil, nil
		}

		if err = apipage.CheckNextTimeStampAfterCurrent(current, singleNext.GetTimestamp()); err != nil {
			return nil, nil, err
		}

		current = singleNext.GetTimestamp()
		fb.WithCreatedAt(current)
	}

	paNext, err := pa.GetNextToken()
	if err != nil {
		return nil, nil, err
	}
	if paNext != nil && !paNext.GetTimestamp().IsZero() {
		return pa.GetList(), paNext, nil
	}
	return pa.GetList(), singleNext, nil
}

func (s *AIApplicationService) page(ctx context.Context, pageOptions *list.PageOptions, fb *filter.Builder) (*openapi.ListApplications200Response, error) {
	return s.client.List(ctx, apipage.GetPageSize(pageOptions.MaxItems, pageOptions.PageSize)+1, pageOptions.Offset, fb.StringList())
}
