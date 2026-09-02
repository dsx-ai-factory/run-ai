/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package errors

import (
	"fmt"
	"net/http"

	metrics_api "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/database/dao"
)

type BusinessErrorHandler interface {
	DaoErrorToBusinessError(err *dao.DaoError, entity Entity) *BusinessError
	MetricsAPIErrorToBusinessError(err *metrics_api.Error, entity Entity) *BusinessError
}
type BusinessErrorHandlerImpl struct{}

func NewBusinessErrorHandler() BusinessErrorHandler {
	return &BusinessErrorHandlerImpl{}
}

func (bu *BusinessErrorHandlerImpl) DaoErrorToBusinessError(daoErr *dao.DaoError, entity Entity) *BusinessError {
	if daoErr == nil {
		return nil
	}
	switch daoErr.Reason {
	case dao.Conflict:
		return &BusinessError{Reason: Conflict, Err: fmt.Errorf("%s already exists. err: %v", entity, daoErr)}
	case dao.NotFound:
		if entity == Tenant {
			return &BusinessError{Reason: TenantNotFound, Err: fmt.Errorf("tenant not found. err: %v", daoErr)}
		}
		return &BusinessError{Reason: NotFound, Err: fmt.Errorf("%s not found. err: %v", entity, daoErr)}
	default:
		return &BusinessError{Reason: GeneralError, Err: fmt.Errorf("failed to handle %s. err: %v", entity, daoErr)}
	}
}

func (bu *BusinessErrorHandlerImpl) MetricsAPIErrorToBusinessError(err *metrics_api.Error, entity Entity) *BusinessError {
	if err == nil {
		return nil
	}
	switch err.Code {
	case http.StatusConflict:
		return &BusinessError{Reason: Conflict, Err: fmt.Errorf("%s already exists. err: %v", entity, err)}
	case http.StatusNotFound:
		return &BusinessError{Reason: NotFound, Err: fmt.Errorf("%s not found. err: %v", entity, err)}
	default:
		return &BusinessError{Reason: GeneralError, Err: fmt.Errorf("failed to handle %d. err: %v", entity, err)}
	}
}
