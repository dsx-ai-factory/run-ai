/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
)

type DatavolumesClient struct {
	ApiClient *openapi.APIClient
}

func (dc *DatavolumesClient) GetDatavolumeAncestorsSync(ctx context.Context, resourceId string) (*openapi.Ancestors, error) {
	request := dc.ApiClient.DatavolumesAPI.GetDatavolumeAncestors(ctx, resourceId)

	ancestors, _, err := dc.ApiClient.DatavolumesAPI.GetDatavolumeAncestorsExecute(request)

	return ancestors, err
}

func (dc *DatavolumesClient) UpdateDatavolumeStatus(ctx context.Context, id string, datavolumeStatusInternal openapi.DatavolumeStatusInternal) (*http.Response, error) {
	request := dc.ApiClient.DatavolumesAPI.UpdateDatavolumeStatus(
		ctx,
		id,
	)
	request = request.DatavolumeStatusInternal(datavolumeStatusInternal)

	return request.Execute()
}

func (dc *DatavolumesClient) SyncDatavolumes(
	ctx context.Context,
	datavolumes []openapi.DatavolumeStatusForSync,
	clusterId string,
	from *string,
	to *string) (*http.Response, error) {
	request := dc.ApiClient.DatavolumesAPI.SyncDatavolumes(ctx)
	request = request.DatavolumesSyncRequest(openapi.DatavolumesSyncRequest{
		Datavolumes: datavolumes,
		ClusterId:   clusterId,
		From:        from,
		To:          to,
	})

	return request.Execute()
}

func (dc *DatavolumesClient) GetDatavolumes(ctx context.Context, requestType openapi.DatavolumeRequestType, usableInProjectId string,
	limit, offset int32, filter []string, sortBy string, sortOrder string) (*openapi.GetDatavolumes200Response, error) {
	request := dc.ApiClient.DatavolumesAPI.GetDatavolumes(ctx)
	request = request.RequestType(requestType)
	request = request.Limit(limit)
	request = request.Offset(offset)
	if usableInProjectId != "" {
		request = request.UsableInProjectId(usableInProjectId)
	}

	if len(filter) != 0 {
		request = request.FilterBy(filter)
	}

	if len(sortBy) != 0 {
		request = request.SortBy(sortBy)
	}

	if len(sortOrder) != 0 {
		request = request.SortOrder(sortOrder)
	}

	dvList, _, err := request.Execute()
	if err != nil {
		return nil, err
	}

	return dvList, nil
}

func (dc *DatavolumesClient) GetDataVolumesNotPaginated(ctx context.Context, requestType openapi.DatavolumeRequestType, usableInProjectId string,
	filter []string, sortBy string, sortOrder string) ([]openapi.DatavolumeNoSharedScopes, error) {
	request := dc.ApiClient.DatavolumesAPI.GetDatavolumes(ctx)
	request = request.RequestType(requestType)
	if usableInProjectId != "" {
		request = request.UsableInProjectId(usableInProjectId)
	}

	if len(filter) != 0 {
		request = request.FilterBy(filter)
	}

	if len(sortBy) != 0 {
		request = request.SortBy(sortBy)
	}

	if len(sortOrder) != 0 {
		request = request.SortOrder(sortOrder)
	}

	nextOffset := int32(0)
	var dataVolumes []openapi.DatavolumeNoSharedScopes

	for true {
		request = request.Limit(500)
		request = request.Offset(nextOffset)
		resp, _, err := request.Execute()
		if err != nil {
			return nil, err
		}

		dataVolumes = append(dataVolumes, resp.Datavolumes...)

		if resp.GetNext() == 0 {
			break
		}

		if resp.GetNext() <= nextOffset {
			return nil, errors.New(fmt.Sprintf("received next offset %d while current offset is %d", resp.GetNext(), nextOffset))
		}

		nextOffset = resp.GetNext()
	}

	return dataVolumes, nil
}

func (dc *DatavolumesClient) CreateInternalDatavolume(ctx context.Context, request openapi.InternalDatavolumeCreationFields) (*openapi.InternalDatavolume, *http.Response, error) {
	internalDvCreationRequest := dc.ApiClient.DatavolumesAPI.InternalCreateDatavolume(ctx)
	internalDvCreationRequest = internalDvCreationRequest.InternalDatavolumeCreationFields(request)
	return internalDvCreationRequest.Execute()
}

func (dc *DatavolumesClient) DeleteDatavolumeById(ctx context.Context, id string) (*openapi.HttpResponse, *http.Response, error) {
	deleteRequest := dc.ApiClient.DatavolumesAPI.DeleteDatavolume(ctx, id)
	return dc.ApiClient.DatavolumesAPI.DeleteDatavolumeExecute(deleteRequest)
}
