/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant

import (
	"net/http"
	"strings"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	tenantIdHeader      = "X-Runai-Tenant-Id"
	forwardedHostHeader = "X-Forwarded-Host"
)

type Handler struct {
	tenantsClient *TenantsManagerClient
}

func NewHandler(tenantsManagerURL string) *Handler {
	return &Handler{
		tenantsClient: NewTenantsManagerClient(tenantsManagerURL),
	}
}

// ServeHTTP is a Traefik forwardAuth endpoint: it resolves the tenant ID from
// the forwarded hostname and returns it as a response header for Traefik to
// copy onto the upstream request. It always returns 200 — when the tenant
// cannot be resolved the request passes through without the header, matching
// the behaviour of the removed extract-tenant-id plugin.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logging.WithCtx(runaictx.LoggerHook(ctx))

	hostname := requestHost(r)
	tenantName := extractTenantNameFromHost(hostname)
	if tenantName == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	tenantId, err := h.tenantsClient.GetTenantIdByNameOrAlias(ctx, tenantName)
	if err != nil {
		log.Errorf("failed to resolve tenant ID for %q: %v", tenantName, err)
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set(tenantIdHeader, tenantId)
	w.WriteHeader(http.StatusOK)
}

// requestHost returns the hostname of the original client request. Traefik's
// forwardAuth sub-request carries it in X-Forwarded-Host; the Host header of
// the sub-request itself points at this sidecar.
func requestHost(r *http.Request) string {
	if host := r.Header.Get(forwardedHostHeader); host != "" {
		return host
	}
	if r.Host != "" {
		return r.Host
	}
	return r.URL.Host
}

func extractTenantNameFromHost(hostname string) string {
	if idx := strings.LastIndex(hostname, ":"); idx != -1 {
		hostname = hostname[:idx]
	}
	if idx := strings.Index(hostname, "."); idx != -1 {
		return hostname[:idx]
	}
	return hostname
}
