/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	openapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
	"github.com/stretchr/testify/require"
)

type mockedWriter struct {
	writtenLogs []openapi.AuditLog
}

func (m *mockedWriter) Write(log openapi.AuditLog) error {
	m.writtenLogs = append(m.writtenLogs, log)
	return nil
}

func (m *mockedWriter) IsInitialized() bool {
	return true
}

func TestMiddleware(t *testing.T) {
	auditMiddleware, writer := setup()
	respRec := httptest.NewRecorder()
	req := createHttpRequest()

	isInitialized = true
	auditMiddleware.ServeHTTP(respRec, req)

	require.Equal(t, 400, respRec.Code)
	require.Equal(t, "actual handler was called", respRec.Body.String())

	require.Equal(t, 1, len(writer.writtenLogs))
	al := writer.writtenLogs[0]
	require.Equal(t, "Subject", al.Subject)
}

func TestMiddleware_writerNotInitialized(t *testing.T) {
	auditMiddleware, writer := setup()
	respRec := httptest.NewRecorder()
	req := createHttpRequest()

	isInitialized = false
	auditMiddleware.ServeHTTP(respRec, req)

	require.Equal(t, 400, respRec.Code)
	require.Equal(t, "actual handler was called", respRec.Body.String())

	require.Equal(t, 0, len(writer.writtenLogs))
}

func setup() (http.HandlerFunc, *mockedWriter) {
	apiHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cb := AuditFromCtx(request.Context())
		cb.WithSubject("Subject").
			WithSubjectType("SubjectType").
			WithEntityId("EntityId").
			WithEntityName("EntityName").
			WithEntityType("EntityType").
			WithClusterName("ClusterName").
			WithClusterId("ClusterId").
			WithAction("Action").
			WithTenantId("1")

		writer.WriteHeader(400)
		writer.Write([]byte("actual handler was called"))
	})
	auditMiddleware := Middleware(apiHandler)
	writer := &mockedWriter{}
	singleInstance = writer

	return auditMiddleware, writer
}

func createHttpRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "some/url", nil)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "1.1.1.1"
	return req
}
