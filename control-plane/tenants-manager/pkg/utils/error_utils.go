package utils

import (
	"fmt"
	"net/http"

	berrors "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/errors"

	"github.com/run-ai/runai/control-plane/tenants-manager/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
)

//go:generate mockery
type ErrorUtilser interface {
	TenantErrorToResponse(err *dao.DaoError) *api.Error
	GetTenantErrorToResponseFromSubEntity(err *dao.DaoError) *api.Error
	BusinessErrorToApiErr(err *berrors.BusinessError, entity berrors.Entity) *api.Error
	ErrorToApiError(err error, statusCode int32) *api.Error
}

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Err     error  `json:"err,omitempty"`
}

func (e Error) Error() string {
	return e.Message
}

func NewErrorUtils() ErrorUtilser {
	return &ServiceUtils{}
}

type ServiceUtils struct{}

func (su *ServiceUtils) TenantErrorToResponse(err *dao.DaoError) *api.Error {
	if err == nil {
		return nil
	}
	switch err.Reason {
	case dao.Conflict:
		return &api.Error{Code: http.StatusConflict, Message: "Tenant name already exists"}
	case dao.NotFound:
		return &api.Error{Code: http.StatusNotFound, Message: berrors.TenantNotFoundMsg}
	case dao.General:
		return &api.Error{Code: http.StatusInternalServerError, Message: "Tenant operation Failed"}
	case dao.NoContent:
		return &api.Error{Code: http.StatusNoContent, Message: "No content"}
	default:
		return nil
	}
}

func (su *ServiceUtils) GetTenantErrorToResponseFromSubEntity(err *dao.DaoError) *api.Error {
	if err == nil {
		return nil
	}
	switch err.Reason {
	case dao.NotFound:
		return &api.Error{Code: http.StatusBadRequest, Message: berrors.TenantNotFoundMsg}
	default:
		return &api.Error{Code: http.StatusInternalServerError, Message: "Failed to get tenant"}
	}
}

func (su *ServiceUtils) BusinessErrorToApiErr(err *berrors.BusinessError, entity berrors.Entity) *api.Error {
	if err == nil {
		return nil
	}
	if err.Reason == berrors.Conflict {
		return &api.Error{Code: http.StatusConflict, Message: fmt.Sprintf("%s already exists", entity.String())}
	}
	if err.Reason == berrors.NotFound {
		return &api.Error{Code: http.StatusNotFound, Message: fmt.Sprintf("%s not found", entity.String())}
	}
	if err.Reason == berrors.TenantNotFound && entity != berrors.Tenant {
		return &api.Error{Code: http.StatusBadRequest, Message: berrors.TenantNotFoundMsg}
	}
	if err.Reason == berrors.TenantNotFound {
		return &api.Error{Code: http.StatusNotFound, Message: berrors.TenantNotFoundMsg}
	}
	return &api.Error{Code: http.StatusInternalServerError, Message: fmt.Sprintf("%s operation Failed", entity.String())}
}

func (su *ServiceUtils) ErrorToApiError(err error, statusCode int32) *api.Error {
	if err == nil {
		return nil
	}
	return &api.Error{Code: statusCode, Message: err.Error()}
}
