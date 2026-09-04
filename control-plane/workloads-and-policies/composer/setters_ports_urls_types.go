/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

type ExposedUrlInfo struct {
	Items                      map[string]*ExposedUrlInfoItem
	ServingPortAuthorizedUsers *string
	WorkloadName               string
	ProjectName                string
	IsJupyterOrTensorboard     bool
	Config                     ClusterStorageConfig

	// ResourcePrefix is prepended to networking K8s resource names (Service,
	// Ingress, ...) for an exposed URL, e.g. "job" → "job-<workloadName>-<nick>".
	// See FieldsComposer.ResourcePrefix.
	ResourcePrefix string

	// ForceSuffixInUrls forces all auto-generated URLs to include the per-item
	// nick suffix - the first URL is NOT the suffix-less ofer-<wl>.<domain>.
	// Set by the distributed composer on the non-primary replica role so its
	// Ingresses don't compete with the primary role's first-URL host.
	ForceSuffixInUrls bool
}

type ExposedUrlInfoItem struct {
	// ExposedUrl is the raw spec entry kept for the networking-plugin setter;
	// the oauth2/username-loader builders read the auth fields below instead.
	ExposedUrl *workloadsApi.ExposedUrl

	Nick          string
	ContainerPort *int32
	Url           string

	OAuthPort        *int32
	AuthorizedUsers  *string
	AuthorizedGroups *string

	// Serving-port authentication is composed even when exposed URLs are disabled.
	BuildWhenExternalAuthURLDisabled bool
	OmitPortDeclaration              bool
	ForceJSONResponse                bool
	EnableSidecarRestartPolicy       *bool
	CookiePrefix                     string
	UpstreamTimeoutSeconds           int64
}

func (info *ExposedUrlInfoItem) CookieAuthorizedUsers() *string {
	if info == nil || !convutils.IsStringHasValue(info.AuthorizedUsers) {
		return nil
	}
	return info.AuthorizedUsers
}

func (info *ExposedUrlInfoItem) CookieAuthorizedGroups() *string {
	if info == nil || !convutils.IsStringHasValue(info.AuthorizedGroups) {
		return nil
	}
	return info.AuthorizedGroups
}

// PortsInfo carries the items to convert plus the per-workload metadata the
// networking setters need (workload name, resource prefix, cluster config).
// It mirrors ExposedUrlInfo so the context travels with the data rather than
// out-of-band via the cookie.
type PortsInfo struct {
	Items          map[string]*workloadsApi.Port
	WorkloadName   string
	ResourcePrefix string
	Config         ClusterStorageConfig
}
