/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clierrors

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	csc "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	imc "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	cec "github.com/run-ai/runai/runai-clients/cli-exposer-client/generated"
	cac "github.com/run-ai/runai/runai-clients/cluster-api-client/generated"
	nmc "github.com/run-ai/runai/runai-clients/nim-client/generated"
	ousc "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	wsc "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	wmc "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
)

const httpErrFormat = "status: %s, reason: %s"
const (
	InternalInvalidURLErrorMsg = "invalid URI for request"
	UserInvalidURLErrorMsg     = "Invalid URI format. Provide a valid web address (e.g., https://www.example.com)"
)

var enrichmentMap = map[string]string{
	InternalInvalidURLErrorMsg: UserInvalidURLErrorMsg,
}

type ApiGeneralErrorMessage struct {
	Message    string                                      `json:"message"`
	Compliance []wsc.ComplianceIssuesComplianceIssuesInner `json:"complianceIssues"`
	Details    string                                      `json:"details"`
	Code       int                                         `json:"code"`
}

type HttpError struct {
	Message string
	Status  string
}

func newHttpError(status, message string) *HttpError {
	return &HttpError{
		Status:  status,
		Message: strings.Trim(message, "\""),
	}
}

func (e *HttpError) Unwrap() error {
	return fmt.Errorf(httpErrFormat, e.Status, e.Message)
}

func (e *HttpError) Error() string {
	return fmt.Sprintf(httpErrFormat, e.Status, e.Message)
}

func HandleResponseError(response *http.Response, err error) error {
	if response == nil || response.Body == nil {
		return err
	}

	switch errType := err.(type) { //nolint:errorlint // same error type different generated code
	case *csc.GenericOpenAPIError:
		return newHttpError(response.Status, string(errType.Body()))
	case *wsc.GenericOpenAPIError:
		return newHttpError(response.Status, string(errType.Body()))
	case *ousc.GenericOpenAPIError:
		return newHttpError(response.Status, string(errType.Body()))
	case *cec.GenericOpenAPIError:
		return newHttpError(response.Status, string(errType.Body()))
	case *cac.GenericOpenAPIError:
		return newHttpError(response.Status, string(errType.Body()))
	case *imc.GenericOpenAPIError:
		return newHttpError(response.Status, string(errType.Body()))
	case *wmc.GenericOpenAPIError:
		return newHttpError(response.Status, string(errType.Body()))
	case *nmc.GenericOpenAPIError:
		return newHttpError(response.Status, string(errType.Body()))
	default:
		return err
	}
}

func UserEnrichment(err error) error {
	var urlErr *url.Error
	ok := errors.As(err, &urlErr)
	if ok {
		if mappedErr, exists := enrichmentMap[urlErr.Err.Error()]; exists {
			return fmt.Errorf("%s, %w", mappedErr, err)
		}
	}
	return err
}
