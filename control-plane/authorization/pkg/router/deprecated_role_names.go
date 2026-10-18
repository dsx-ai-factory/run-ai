/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const internalCreateAccessRulePath = "/internal/authorization/access-rules"

// deprecatedInternalRoleNames maps the internal role names that were renamed to their current names. Clients built
// before the rename still send the old names, and the generated server rejects any value that is not in the current
// enum, so the old names are rewritten before the request reaches the generated handler. This is deliberately not in
// the generated code or the client: remove this file once no supported client sends the old names.
var deprecatedInternalRoleNames = map[string]string{
	"Cluster application":   "Cluster service account",
	"Installer application": "Installer service account",
}

// deprecatedRoleNamesMiddleware rewrites a deprecated roleName in the body of the internal create-access-rule request.
// Anything it cannot parse is passed through untouched so the generated handler keeps producing its usual errors.
func deprecatedRoleNamesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == internalCreateAccessRulePath && r.Body != nil {
			r.Body = rewriteDeprecatedRoleName(r)
		}
		next.ServeHTTP(w, r)
	})
}

func rewriteDeprecatedRoleName(r *http.Request) io.ReadCloser {
	original, err := io.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		return io.NopCloser(bytes.NewReader(original))
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(original, &fields); err != nil {
		return io.NopCloser(bytes.NewReader(original))
	}

	var roleName string
	if err := json.Unmarshal(fields["roleName"], &roleName); err != nil {
		return io.NopCloser(bytes.NewReader(original))
	}
	newName, deprecated := deprecatedInternalRoleNames[roleName]
	if !deprecated {
		return io.NopCloser(bytes.NewReader(original))
	}

	rawNewName, err := json.Marshal(newName)
	if err != nil {
		return io.NopCloser(bytes.NewReader(original))
	}
	fields["roleName"] = rawNewName
	rewritten, err := json.Marshal(fields)
	if err != nil {
		return io.NopCloser(bytes.NewReader(original))
	}

	logging.Warnf("deprecated role name %q received on %s, using %q", roleName, internalCreateAccessRulePath, newName)
	r.ContentLength = int64(len(rewritten))
	return io.NopCloser(bytes.NewReader(rewritten))
}
