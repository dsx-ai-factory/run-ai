package services

import (
	"net/http"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func NewErrorResponse(apiError *api.Error) api.ImplResponse {
	return api.ImplResponse{
		Code: int(apiError.Code),
		Body: *apiError,
	}
}

func NewSubmissionErrorResponse(err *workloadsApi.SubmissionErrorV2) api.ImplResponse {
	return api.ImplResponse{
		Code: int(err.Code),
		Body: *err,
	}
}

func NewDeletionResponse() api.ImplResponse {
	return api.ImplResponse{
		Code: http.StatusAccepted,
		Body: &api.Error{
			Code:    http.StatusAccepted,
			Message: "Asset deletion request has been accepted.",
		},
	}
}
