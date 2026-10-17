/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	iamError "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/error"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
)

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e Error) Error() string {
	return e.Message
}

func HandleError(ctx context.Context, entityType string, identifier string, identifierValue string, err error, errMsg string) (api.ImplResponse, error) {
	if IsNotFound(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(`couldn't find %s with %s %s. err: %v`, entityType, identifier, identifierValue, err.Error())
		return api.Response(http.StatusNotFound, nil), Error{Code: http.StatusNotFound, Message: errMsg}
	}
	if IsConflict(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(`%s with %s %s already exists. err: %v`, entityType, identifier, identifierValue, err.Error())
		return api.Response(http.StatusConflict, nil), Error{Code: http.StatusConflict, Message: errMsg}
	}
	if IsUnauthorized(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(`%s. err: %v`, errMsg, err.Error())
		return api.Response(http.StatusUnauthorized, nil), Error{Code: http.StatusUnauthorized, Message: errMsg}
	}
	if IsBadRequest(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(`%s. err: %v`, errMsg, err.Error())
		return api.Response(http.StatusBadRequest, nil), Error{Code: http.StatusBadRequest, Message: errMsg}
	}
	if IsForbidden(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(`%s. err: %v`, errMsg, err.Error())
		return api.Response(http.StatusForbidden, nil), Error{Code: http.StatusForbidden, Message: errMsg}
	}
	if IsInternalServerError(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(`%s. err: %v`, errMsg, err.Error())
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error not recognized. %s. err: %v", errMsg, err.Error())
	}
	return api.Response(http.StatusInternalServerError, nil), Error{Code: http.StatusInternalServerError, Message: ""}
}

func IsNotFound(err error) bool {
	return strings.Contains(err.Error(), "404 Not Found") ||
		strings.Contains(strings.ToLower(err.Error()), "not found") ||
		errors.Is(err, ErrNotFound) || errors.Is(err, iamError.ErrNotFound) ||
		getErrorStatusCode(err) == http.StatusNotFound || parseSettingPkgError(err) == ErrNotFound
}

func IsConflict(err error) bool {
	return strings.Contains(err.Error(), "409 Conflict") ||
		strings.Contains(strings.ToLower(err.Error()), "conflict") ||
		errors.Is(err, ErrConflict) || errors.Is(err, iamError.ErrConflict) ||
		getErrorStatusCode(err) == http.StatusConflict || parseSettingPkgError(err) == ErrConflict
}

func IsBadRequest(err error) bool {
	return strings.Contains(err.Error(), "400") ||
		strings.Contains(strings.ToLower(err.Error()), "bad request") ||
		errors.Is(err, ErrInvalidInput) || errors.Is(err, iamError.ErrInvalidInput) ||
		getErrorStatusCode(err) == http.StatusBadRequest || parseSettingPkgError(err) == ErrInvalidInput
}

func IsUnauthorized(err error) bool {
	return strings.Contains(err.Error(), "401") ||
		errors.Is(err, ErrUnauthorized) || errors.Is(err, iamError.ErrUnauthorized) ||
		getErrorStatusCode(err) == http.StatusUnauthorized
}

func IsForbidden(err error) bool {
	return strings.Contains(err.Error(), "403") ||
		errors.Is(err, ErrForbidden) ||
		getErrorStatusCode(err) == http.StatusForbidden
}

func IsInternalServerError(err error) bool {
	return strings.Contains(err.Error(), "500") ||
		errors.Is(err, ErrInternal) || errors.Is(err, iamError.ErrInternal) ||
		getErrorStatusCode(err) == http.StatusInternalServerError || parseSettingPkgError(err) == ErrInternal
}

func getErrorStatusCode(err error) int {
	var errorHandler Error
	if errors.As(err, &errorHandler) {
		return errorHandler.Code
	}
	return http.StatusInternalServerError
}

func GetHttpStatusCodeAndMessage(err error) (int, string) {
	if err == nil {
		return http.StatusOK, "ok"
	}

	if IsBadRequest(err) {
		return http.StatusBadRequest, "bad request"
	}
	if IsConflict(err) {
		return http.StatusConflict, "already exists"
	}
	if IsNotFound(err) {
		return http.StatusNotFound, "not found"
	}
	if IsUnauthorized(err) {
		return http.StatusUnauthorized, "unauthorized"
	}
	if IsForbidden(err) {
		return http.StatusForbidden, "forbidden"
	}

	return http.StatusInternalServerError, "internal server error"
}

func parseSettingPkgError(err error) error {
	switch err.(type) {
	case setting_pkg.ErrorNotFound:
		return ErrNotFound
	case setting_pkg.ErrorBadRequest:
		return ErrInvalidInput
	case setting_pkg.ErrorConflict:
		return ErrConflict
	case setting_pkg.ErrorInternal:
		return ErrInternal
	}
	return ErrInternal
}

func ParseGormError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrDuplicatedKey) || isPostgresUniqueViolation(err):
		return ErrConflict
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrInvalidTransaction):
		return ErrInvalidInput
	}
	return ErrInternal
}

func LogHandler(ctx context.Context, msg string, err error) {
	if errors.Is(err, ErrInternal) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", msg)
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s", msg)
	}
}

func isPostgresUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint") ||
		strings.Contains(err.Error(), "SQLSTATE 23505") ||
		strings.Contains(err.Error(), "23505")
}
