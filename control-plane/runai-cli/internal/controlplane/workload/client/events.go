package client

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/options"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

const (
	defaultPageSizeEvents int32 = 100

	EventsDefaultSortOrder    = Asc
	unableToParseSortOrderMsg = "unable to parse sort order '%s'"
)

type SortOrder string

const (
	Asc  SortOrder = "asc"
	Desc SortOrder = "desc"
)

func (so SortOrder) String() string {
	return string(so)
}

func GetEventSortOrderOptions() []string {
	return []string{Asc.String(), Desc.String()}
}

func ParseSortOrder(sortOrderStr string) (SortOrder, error) {
	switch sortOrderStr {
	case Asc.String():
		return Asc, nil
	case Desc.String():
		return Desc, nil
	default:
		return "", fmt.Errorf(unableToParseSortOrderMsg, sortOrderStr)
	}
}

type BackendWorkloadEventClient struct {
	apiClient *openapi.APIClient
	sortOrder SortOrder
}

func NewBackendWorkloadEventsClient(urlStr, token string, sortOrderStr string) (*BackendWorkloadEventClient, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}

	sortOrder, err := ParseSortOrder(sortOrderStr)
	if err != nil {
		return nil, err
	}

	return &BackendWorkloadEventClient{apiClient: openapi.NewAPIClient(conf), sortOrder: sortOrder}, nil
}

func (bsc *BackendWorkloadEventClient) List(ctx context.Context,
	workloadID string, pagination options.Paginationer) (*openapi.GetWorkloadEvents200Response, error) {
	request := bsc.apiClient.EventsAPI.GetWorkloadEvents(ctx, workloadID)
	if pagination != nil {
		request = request.Limit(pagination.GetLimit())
		request = request.Offset(pagination.GetOffset())
	}

	request = request.SortOrder(bsc.sortOrder.String())

	workloadList, response, err := request.Execute()
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	defer func() {
		if response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if !httputils.CheckResponseStatusCodeOK(response) {
		return nil, fmt.Errorf(consts.ErrorWorkloadEventsListAPI, response.Status)
	}

	return workloadList, nil
}

func (bsc *BackendWorkloadEventClient) ListAll(ctx context.Context, workloadID string) ([]openapi.Event, error) {
	var allEvents []openapi.Event
	var offset int32
	pageSize := defaultPageSizeEvents

	for {
		pagination := options.Pagination{

			Limit:  ptr.To(pageSize),
			Offset: offset,
		}

		events, err := bsc.List(ctx, workloadID, &pagination)
		if err != nil {
			return nil, err
		}
		allEvents = append(allEvents, events.Events...)
		if !events.HasNext() {
			break
		}
		offset = events.GetNext()
	}

	return allEvents, nil
}

func (bsc *BackendWorkloadEventClient) ListByLimit(ctx context.Context, workloadID string, limit int32) ([]openapi.Event, error) {
	if limit <= NoLimit {
		return bsc.ListAll(ctx, workloadID)
	}
	var allEvents []openapi.Event
	var offset int32
	pageSize := defaultPageSizeEvents
	if limit < defaultPageSizeEvents {
		pageSize = limit
	}

	for {
		pagination := options.Pagination{

			Limit:  ptr.To(pageSize),
			Offset: offset,
		}

		events, err := bsc.List(ctx, workloadID, &pagination)
		if err != nil {
			return nil, err
		}
		allEvents = append(allEvents, events.Events...)
		if !events.HasNext() {
			break
		}
		offset = events.GetNext()
		lenAllEvents := int32(len(allEvents)) //nolint:gosec // limit is set max to 100
		if limit-lenAllEvents <= defaultPageSizeEvents {
			pageSize = limit - lenAllEvents
		}

		if pageSize <= 0 {
			break
		}
	}

	return allEvents, nil
}
