/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"net/http"

	"github.com/run-ai/runai/runai-common-packages/runaihttp"
	"gorm.io/gorm"
)

type Error struct {
	Code int32 `json:"code" gorm:"not null;"`

	Message string `json:"message" gorm:"not null;"`

	Details string `json:"details,omitempty" gorm:"not null;"`
}

func NewError(code int, message string, err error) *Error {
	result := Error{
		Code:    int32(code),
		Message: message,
	}
	runaihttp.LoadErrorDetails(&result.Code, &result.Message, &result.Details, err)
	return &result
}
func (e *Error) WithDetails(err error) *Error {
	runaihttp.LoadErrorDetails(&e.Code, &e.Message, &e.Details, err)
	return e
}
func (e *Error) WithDetailsFromDb(db *gorm.DB) *Error {
	if db.Error != nil {
		runaihttp.LoadErrorDetails(&e.Code, &e.Message, &e.Details, db.Error)
		return e
	}
	if db.RowsAffected == 0 {
		e.Code = http.StatusNotFound
		e.Details = "requested resource does not exist"
		return e
	}
	return e
}
func NewValidationError(message string) *Error {
	result := Error{
		Code:    http.StatusBadRequest,
		Message: message,
	}
	return &result
}
