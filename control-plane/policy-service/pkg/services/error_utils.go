package services

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

func NewErrorResponse(apiError *api.Error) api.ImplResponse {
	return api.ImplResponse{
		Code: int(apiError.Code),
		Body: *apiError,
	}
}
