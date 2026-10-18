/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeprecatedRoleNamesMiddleware(t *testing.T) {
	for _, tc := range []struct {
		name     string
		method   string
		path     string
		body     string
		wantBody string
	}{
		{
			name:     "deprecated cluster role is rewritten",
			method:   http.MethodPost,
			path:     internalCreateAccessRulePath,
			body:     `{"roleName":"Cluster application","subjectId":"agent"}`,
			wantBody: `{"roleName":"Cluster service account","subjectId":"agent"}`,
		},
		{
			name:     "deprecated installer role is rewritten",
			method:   http.MethodPost,
			path:     internalCreateAccessRulePath,
			body:     `{"roleName":"Installer application","subjectId":"installer"}`,
			wantBody: `{"roleName":"Installer service account","subjectId":"installer"}`,
		},
		{
			name:     "current role name is untouched",
			method:   http.MethodPost,
			path:     internalCreateAccessRulePath,
			body:     `{"roleName":"Cluster service account"}`,
			wantBody: `{"roleName":"Cluster service account"}`,
		},
		{
			name:     "other role is untouched",
			method:   http.MethodPost,
			path:     internalCreateAccessRulePath,
			body:     `{"roleName":"System administrator"}`,
			wantBody: `{"roleName":"System administrator"}`,
		},
		{
			name:     "invalid json is passed through",
			method:   http.MethodPost,
			path:     internalCreateAccessRulePath,
			body:     `{"roleName":`,
			wantBody: `{"roleName":`,
		},
		{
			name:     "missing role name is passed through",
			method:   http.MethodPost,
			path:     internalCreateAccessRulePath,
			body:     `{"subjectId":"agent"}`,
			wantBody: `{"subjectId":"agent"}`,
		},
		{
			name:     "other path is untouched",
			method:   http.MethodPost,
			path:     "/internal/authorization/scope-access-rules",
			body:     `{"roleName":"Cluster application"}`,
			wantBody: `{"roleName":"Cluster application"}`,
		},
		{
			name:     "other method is untouched",
			method:   http.MethodPut,
			path:     internalCreateAccessRulePath,
			body:     `{"roleName":"Cluster application"}`,
			wantBody: `{"roleName":"Cluster application"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotBody string
			handler := deprecatedRoleNamesMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				gotBody = string(b)
			}))

			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			handler.ServeHTTP(httptest.NewRecorder(), req)

			assert.Equal(t, tc.wantBody, gotBody, "body seen by the next handler")
		})
	}
}
