package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads/pkg/services_utils"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type EventsService struct{}

func NewEventsService() api.EventsAPIServicer {
	return &EventsService{}
}

func (s *EventsService) GetWorkloadHistory(ctx context.Context, workloadId string, offset int32, limit int32) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	limit = setDefaultLimit(limit)

	records, dbErr := dao.DbGetWorkloadHistoryRecords(ctx, tenantId, workloadId, offset, limit)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	apiRecords := []api.HistoryRecord{}
	for _, record := range *records {
		apiRecords = append(apiRecords, HistoryRecordModelToApi(record))
	}

	recordsCount := int32(len(apiRecords))

	response := api.GetWorkloadHistory200Response{
		Records: apiRecords,
	}
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *EventsService) GetRevisionHistory(ctx context.Context, revisionId string, offset int32, limit int32) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	limit = setDefaultLimit(limit)

	records, dbErr := dao.DbGetRevisionHistoryRecords(ctx, tenantId, revisionId, offset, limit)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	apiRecords := []api.HistoryRecord{}
	for _, record := range *records {
		apiRecords = append(apiRecords, RevisionHistoryRecordModelToApi(record))
	}

	recordsCount := int32(len(apiRecords))

	response := api.GetWorkloadHistory200Response{
		Records: apiRecords,
	}
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *EventsService) GetWorkloadEvents(ctx context.Context, workloadId string, offset int32, limit int32, sortOrder string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	limit = setDefaultLimit(limit)

	events, dbErr := dao.DbGetWorkloadEvents(ctx, tenantId, workloadId, offset, limit, sortOrder)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	apiEvents := []api.Event{}
	for _, event := range *events {
		apiEvents = append(apiEvents, EventViewModelToApi(event))
	}

	recordsCount := int32(len(apiEvents))

	response := api.GetWorkloadEvents200Response{
		Events: apiEvents,
	}
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *EventsService) GetRevisionEvents(ctx context.Context, revisionId string, offset int32, limit int32, sortOrder string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	limit = setDefaultLimit(limit)

	events, dbErr := dao.DbGetRevisionEvents(ctx, tenantId, revisionId, offset, limit, sortOrder)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	apiEvents := []api.Event{}
	for _, event := range *events {
		apiEvents = append(apiEvents, RevisionEventViewModelToApi(event))
	}

	recordsCount := int32(len(apiEvents))

	response := api.GetWorkloadEvents200Response{
		Events: apiEvents,
	}
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *EventsService) SyncEvents(ctx context.Context, eventsSyncRequest api.EventsSyncRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting events sync for tenant <%d>", tenantId)

	eventModels := []models.Event{}
	for _, event := range eventsSyncRequest.Events {
		eventModel, err := EventApiToModel(tenantId, event)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Event <%s> of type <%s> source <%s> reason <%s> message <%s> InvolvedObject [%#v] is invalid in sync request of cluster <%s>. Err: %v",
				event.Id, event.Type, event.Source, event.Reason, event.Message, event.InvolvedObject, event.ClusterId, err)
		} else {
			eventModels = append(eventModels, eventModel)
		}
	}

	dbErr := dao.DbSyncEvents(ctx, eventModels)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully synced events for tenant <%d>", tenantId)

	return NoContentResponse(), nil
}
