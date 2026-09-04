/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"

	apipage "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/page"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

const (
	workloadPage = "workload_page"
)

type Paginator interface {
	List(ctx context.Context, limit, offset int32, filter []string, deleted bool) (*openapi.GetWorkloads200Response, error)
}

type Pager struct {
	workloadClient Paginator
}

func NewPager(workloadClient Paginator) *Pager {
	return &Pager{workloadClient: workloadClient}
}

func (p *Pager) GetWorkloads(txn *transaction.Txn, options *list.PageOptions, fb *filter.Builder, deleted bool) ([]openapi.Workload, *structs.NextToken, error) {
	if options.NoPagination {
		return p.singlePage(txn, options, fb, deleted)
	}
	return p.multiPages(txn, options, fb, deleted)
}

func (p *Pager) multiPages(txn *transaction.Txn, pageOptions *list.PageOptions, fb *filter.Builder, deleted bool) ([]openapi.Workload, *structs.NextToken, error) {
	var (
		workloadList []openapi.Workload
		current      time.Time
		singleNext   *structs.NextToken
		err          error
	)

	pa := apipage.NewAppender[openapi.Workload](int(pageOptions.MaxItems))
	for !pa.ReachedLimit() {
		workloadList, singleNext, err = p.singlePage(txn, pageOptions, fb, deleted)
		if err != nil {
			return nil, nil, err
		}

		// add entries
		err = pa.AddPage(workloadList)
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

func (p *Pager) singlePage(txn *transaction.Txn, pageOptions *list.PageOptions, fb *filter.Builder, deleted bool) ([]openapi.Workload, *structs.NextToken, error) {
	var (
		next *structs.NextToken
	)

	// single page fetch time
	timer := txn.AddTimer(workloadPage)
	defer txn.StopTimer(timer)

	if fb == nil {
		return nil, nil, apipage.ErrNilFilter
	}

	if pageOptions == nil || pageOptions.PageSize < 1 {
		return nil, nil, apipage.ErrPageSizeNotPositive
	}

	page, err := p.page(txn.GetCtx(), pageOptions, fb, deleted)
	if err != nil {
		return nil, nil, err
	}

	// page sanity
	if page == nil {
		return nil, nil, apipage.ErrNilPage
	}

	pa := apipage.NewAppender[openapi.Workload](int(pageOptions.PageSize))
	err = pa.AddPage(page.Workloads)
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

func (p *Pager) page(ctx context.Context, pageOptions *list.PageOptions, fb *filter.Builder, deleted bool) (*openapi.GetWorkloads200Response, error) {
	page, err := p.workloadClient.List(ctx, pageOptions.PageSize+1, pageOptions.Offset, fb.StringList(), deleted)
	if err != nil {
		return nil, err
	}

	return page, nil
}
