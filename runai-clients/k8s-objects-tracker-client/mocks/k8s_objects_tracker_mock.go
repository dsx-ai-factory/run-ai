/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mocks

import (
	"context"
	"net/http"

	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
)

// StorageClassConfigMock supports fixed set of pvcs and possible error
type K8SObjectTrackerMock struct {
	// in reality the key is tenantId + clusterId + name, but we only track the names.
	StorageClassConfig map[string]k8sObjectsTracker.StorageClassConfiguration
	Secrets            []k8sObjectsTracker.SecretListInfo
	// ResourceClaimTemplates is keyed by cluster id because tracker rows do not carry it.
	ResourceClaimTemplates      map[string][]k8sObjectsTracker.ResourceClaimTemplate
	ResourceClaimTemplatesError *k8sObjectsTracker.Error
}

func NewK8SObjectTrackerMock() K8SObjectTrackerMock {
	return K8SObjectTrackerMock{
		StorageClassConfig: make(map[string]k8sObjectsTracker.StorageClassConfiguration),
	}
}

func (s K8SObjectTrackerMock) GetStorageClassConfigInt(ctx context.Context, clusterId string, storageClassName string) (*k8sObjectsTracker.StorageClassConfiguration, *k8sObjectsTracker.Error) {
	result, ok := s.StorageClassConfig[storageClassName]
	if !ok || result.ScopeId != clusterId {
		return nil, k8sObjectsTracker.NewError(http.StatusNotFound, "not found")
	}
	return &result, nil
}

func (s K8SObjectTrackerMock) GetStorageClassesConfigInt(ctx context.Context, clusterId string, storageClassNames []string) (map[string]k8sObjectsTracker.StorageClassConfiguration, *k8sObjectsTracker.Error) {
	configs := make(map[string]k8sObjectsTracker.StorageClassConfiguration)

	for _, name := range storageClassNames {
		result, ok := s.StorageClassConfig[name]
		if ok && result.ScopeId == clusterId {
			configs[name] = result
		}
	}

	return configs, nil
}

func (s K8SObjectTrackerMock) WithStorageClassConfig(name string, config k8sObjectsTracker.StorageClassConfiguration) K8SObjectTrackerMock {
	s.StorageClassConfig[name] = config
	return s
}

func (s K8SObjectTrackerMock) ListSecretsByScope(ctx context.Context, scopeType string, scopeId string) (*k8sObjectsTracker.SecretListResponse, *k8sObjectsTracker.Error) {
	return &k8sObjectsTracker.SecretListResponse{
		Secrets: s.Secrets,
	}, nil
}

func (s K8SObjectTrackerMock) WithSecrets(secrets []k8sObjectsTracker.SecretListInfo) K8SObjectTrackerMock {
	s.Secrets = secrets
	return s
}

// ListResourceClaimTemplatesByName mirrors the tracker's filtering: exact name match across
// every scope of the cluster, and a nil row origin counted as "existing".
func (s K8SObjectTrackerMock) ListResourceClaimTemplatesByName(ctx context.Context, clusterId string, name string, origin string) (*k8sObjectsTracker.ResourceClaimTemplates, *k8sObjectsTracker.Error) {
	if s.ResourceClaimTemplatesError != nil {
		return nil, s.ResourceClaimTemplatesError
	}

	items := []k8sObjectsTracker.ResourceClaimTemplate{}
	for _, rct := range s.ResourceClaimTemplates[clusterId] {
		if rct.Name != name {
			continue
		}
		if origin != "" && origin != "all" {
			rowOrigin := string(k8sObjectsTracker.ORIGIN_EXISTING)
			if o := rct.Origin.Get(); o != nil {
				rowOrigin = string(*o)
			}
			if rowOrigin != origin {
				continue
			}
		}
		items = append(items, rct)
	}
	return &k8sObjectsTracker.ResourceClaimTemplates{Items: items}, nil
}

func (s K8SObjectTrackerMock) WithResourceClaimTemplates(clusterId string, rcts []k8sObjectsTracker.ResourceClaimTemplate) K8SObjectTrackerMock {
	if s.ResourceClaimTemplates == nil {
		s.ResourceClaimTemplates = make(map[string][]k8sObjectsTracker.ResourceClaimTemplate)
	}
	s.ResourceClaimTemplates[clusterId] = rcts
	return s
}

func (s K8SObjectTrackerMock) WithResourceClaimTemplatesError(err *k8sObjectsTracker.Error) K8SObjectTrackerMock {
	s.ResourceClaimTemplatesError = err
	return s
}
