/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
)

// WorkloadNotReadyError internal error for retrying
type WorkloadNotReadyError struct {
	message string
}

func NewNotReadyError(message string) *WorkloadNotReadyError {
	return &WorkloadNotReadyError{message: fmt.Sprintf(consts.WorkloadLogsStreamerNotReadyError, message)}
}
func (e *WorkloadNotReadyError) Error() string {
	return e.message
}

func IgnoreNotReadyError(err error) error {
	var notReadyError *WorkloadNotReadyError
	ok := errors.As(err, &notReadyError)
	if ok {
		return nil
	}
	return err
}
func HandleError(response *http.Response, err error) error {
	if response == nil {
		return fmt.Errorf(consts.GenericStreamLogsError, err)
	}
	if response.ContentLength < 1 {
		return fmt.Errorf(consts.ContentLengthInvalid, response.ContentLength)
	}
	var data = make([]byte, response.ContentLength)
	body := response.Body
	var errorMsg error
	if body != nil {
		defer body.Close()
		_, readError := body.Read(data)
		if readError != nil {
			return fmt.Errorf(consts.GenericStreamLogsError, readError)
		}
	}
	errorMsg = errors.New(string(data)) //nolint:govet // error extracted from body
	if response.StatusCode == http.StatusPreconditionFailed {
		return NewNotReadyError(errorMsg.Error())
	}
	if response.StatusCode == http.StatusNotFound {
		return fmt.Errorf(consts.WorkloadLogsStreamerNotFoundError, errorMsg)
	}
	return fmt.Errorf(consts.GenericStreamLogsError, errorMsg)
}
