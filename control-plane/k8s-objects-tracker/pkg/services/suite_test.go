/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/clusters"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	cache_testing "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	runaictx_http "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"

	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorizationApi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorizationUtils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dbtesting"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var DefaultCheckerClusterConfig = &checker.ClustersCacheConfig{ExpirationSec: 600, CleanupSec: 10}

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	config.MockConfig()
	config.Init()

	cleanup := dbtesting.CreateTables()
	t.Cleanup(cleanup)

	cache_testing.EnableCache()

	authorizationClient.SetMockAuthorizationClient(auth.SimpleMockAuthService{})

	clusters.SetMock(clusters.NewSimpleMock())

	checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

	RunSpecs(t, "DB tests")
	t.Cleanup(func() {
		dbtesting.CleanTables()
		cache_testing.DisableCache()
	})
}

// we must use the same tenant ID throughout all tests
var randomTenantId = rand.Int31n(1000)

// we hold a thread-safe counter that increases every time we want new random number
var (
	counter      = rand.Int31n(100)
	counterMutex sync.Mutex
)

func nextRandom() int32 {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	counter += rand.Int31n(99) + 1
	return counter
}

// RandomScopes represents a set of org unit parameters
type RandomScopes struct {
	TenantId     int32
	TenantIdStr  string
	ClusterId    string
	DepartmentId string
	ProjectId    int32
	ProjectIdStr string
	IsForbidden  bool // is the requesting user forbidden from seeing this?
	Namespace    *string
}

func NewRandomsSet() *RandomScopes {
	result := RandomScopes{
		TenantId:     randomTenantId,
		ClusterId:    uuid.New().String(),
		DepartmentId: fmt.Sprintf("%d", nextRandom()),
		ProjectId:    nextRandom(),
	}

	result.TenantIdStr = fmt.Sprintf("%v", result.TenantId)
	result.ProjectIdStr = fmt.Sprintf("%v", result.ProjectId)

	return &result
}

// NewRandomsSetWithDifferentTenant returns a RandomScopes with random values
// and a TenantId incremented by 1 to ensure it's different from randomTenantId.
func NewRandomsSetWithDifferentTenant() *RandomScopes {
	result := RandomScopes{
		TenantId:     randomTenantId,
		ClusterId:    uuid.New().String(),
		DepartmentId: fmt.Sprintf("%d", nextRandom()),
		ProjectId:    nextRandom(),
	}

	result.TenantId = randomTenantId + 1
	result.TenantIdStr = fmt.Sprintf("%v", result.TenantId)
	result.ProjectIdStr = fmt.Sprintf("%v", result.ProjectId)
	return &result
}

func (rs *RandomScopes) Clone() *RandomScopes {
	return &RandomScopes{
		TenantId:     rs.TenantId,
		TenantIdStr:  rs.TenantIdStr,
		ClusterId:    rs.ClusterId,
		DepartmentId: rs.DepartmentId,
		ProjectId:    rs.ProjectId,
		ProjectIdStr: rs.ProjectIdStr,
		IsForbidden:  rs.IsForbidden,
	}
}

func (rs *RandomScopes) DifferentCluster() *RandomScopes {
	result := rs.Clone()
	result.ClusterId = uuid.New().String()
	return result
}

func (rs *RandomScopes) DifferentDepartment() *RandomScopes {
	result := rs.Clone()
	result.DepartmentId = fmt.Sprintf("%d", nextRandom())
	return result.DifferentProject()
}

func (rs *RandomScopes) DifferentProject() *RandomScopes {
	namespace := fmt.Sprintf("runai-%d", rs.ProjectId)
	result := rs.Clone()
	result.ProjectId = nextRandom()
	result.ProjectIdStr = fmt.Sprintf("%d", result.ProjectId)
	result.Namespace = &namespace
	return result
}

func (rs *RandomScopes) Forbidden() *RandomScopes {
	result := rs.Clone()
	result.IsForbidden = true
	return result
}

func (rs *RandomScopes) PvcsClusterUrl(name string) string {
	return fmt.Sprintf("/api/v2/pvcs/%s%s", rs.ClusterId, name)
}

func (rs *RandomScopes) PvcsDepartmentUrl(name string) string {
	return fmt.Sprintf("/api/v2/pvcs/%s/departments/%s%s", rs.ClusterId, rs.DepartmentId, name)
}

func (rs *RandomScopes) PvcsProjectUrl(name string) string {
	return fmt.Sprintf("/api/v2/pvcs/%s/namespaces/runai-%s%s", rs.ClusterId, rs.ProjectIdStr, name)
}

func (rs *RandomScopes) PvcsListUrl(scopeType orgUnitApi.ScopeType, optionalName string, optionalOrigin string) string {
	return fmt.Sprintf("/api/v1/pvcs?scopeType=%s&scopeId=%s&name=%s&origin=%s", scopeType, rs.ScopeId(scopeType), optionalName, optionalOrigin)
}

func (rs *RandomScopes) PvcsListByClusterId() string {
	return fmt.Sprintf("/api/v1/pvcs?clusterId=%s", rs.ClusterId)
}

func (rs *RandomScopes) PvcsListResponse(scopeType orgUnitApi.ScopeType, pvcs []api.Pvc) api.PvcListResponse {
	return api.PvcListResponse{
		Pvcs: rs.PvcsListInfos(scopeType, pvcs),
	}
}

func (rs *RandomScopes) PvcsListInfos(scopeType orgUnitApi.ScopeType, pvcs []api.Pvc) []api.PvcListInfo {
	result := make([]api.PvcListInfo, 0, len(pvcs))
	for _, pvc := range pvcs {
		pvcListInfo := api.PvcListInfo{
			ClusterId:         rs.ClusterId,
			ScopeType:         string(scopeType),
			ScopeId:           rs.ScopeId(scopeType),
			Name:              pvc.Name,
			Origin:            pvc.Origin,
			StorageClassName:  pvc.StorageClassName,
			RequestedCapacity: pvc.RequestedCapacity,
			ActualCapacity:    pvc.ActualCapacity,
		}

		if pvc.Phase != nil {
			pvcListInfo.Phase = convutils.Ref(*pvc.Phase)
		}

		result = append(result, pvcListInfo)
	}
	return result
}

func (rs *RandomScopes) SecretsClusterUrl(name string) string {
	return fmt.Sprintf("/api/v1/secrets/%s%s", rs.ClusterId, name)
}

func (rs *RandomScopes) SecretsDepartmentUrl(name string) string {
	return fmt.Sprintf("/api/v1/secrets/%s/departments/%s%s", rs.ClusterId, rs.DepartmentId, name)
}

func (rs *RandomScopes) SecretsProjectUrl(name string) string {
	return fmt.Sprintf("/api/v1/secrets/%s/namespaces/runai-%s%s", rs.ClusterId, rs.ProjectIdStr, name)
}

func (rs *RandomScopes) SecretsListUrl(scopeType orgUnitApi.ScopeType, optionalName string, optionalOrigin string) string {
	return fmt.Sprintf("/api/v1/secrets?scopeType=%s&scopeId=%s&name=%s&origin=%s", scopeType, rs.ScopeId(scopeType), optionalName, optionalOrigin)
}

func (rs *RandomScopes) SecretsListUrlWithType(scopeType orgUnitApi.ScopeType, optionalName string, optionalType api.SecretType) string {
	return fmt.Sprintf("/api/v1/secrets?scopeType=%s&scopeId=%s&name=%s&secretType=%s", scopeType, rs.ScopeId(scopeType), optionalName, optionalType)
}

func (rs *RandomScopes) SecretsListUrlWithPrivate(scopeType orgUnitApi.ScopeType, optionalName string, optionalOrigin string) string {
	return fmt.Sprintf("/api/v1/secrets?scopeType=%s&scopeId=%s&name=%s&origin=%s&includePrivate=true", scopeType, rs.ScopeId(scopeType), optionalName, optionalOrigin)
}

func (rs *RandomScopes) SecretsListByClusterId(optionalOrigin string) string {
	return fmt.Sprintf("/api/v1/secrets?clusterId=%s&origin=%s", rs.ClusterId, optionalOrigin)
}

func (rs *RandomScopes) SecretsListResponse(scopeType orgUnitApi.ScopeType, secrets ...api.Secret) api.SecretListResponse {
	return api.SecretListResponse{
		Secrets: rs.SecretsListInfos(scopeType, secrets...),
	}
}

func (rs *RandomScopes) SecretsListInfos(scopeType orgUnitApi.ScopeType, secrets ...api.Secret) []api.SecretListInfo {
	result := make([]api.SecretListInfo, 0, len(secrets))
	for _, secret := range secrets {
		result = append(result, api.SecretListInfo{
			ClusterId: rs.ClusterId,
			ScopeType: string(scopeType),
			ScopeId:   rs.ScopeId(scopeType),
			Name:      secret.Name,
			Type:      secret.Type,
			Origin:    secret.Origin,
			Private:   secret.Private,
		})
	}
	return result
}

// ConfigMapClusterUrl returns the URL for ConfigMap with cluster id and name
func (rs *RandomScopes) ConfigMapClusterUrl(name string) string {
	return fmt.Sprintf("/api/v1/config-maps/%s%s", rs.ClusterId, name)
}

// ConfigMapDepartmentUrl returns the URL for ConfigMap with department id, cluster id and name
func (rs *RandomScopes) ConfigMapDepartmentUrl(name string) string {
	return fmt.Sprintf("/api/v1/config-maps/%s/departments/%s%s", rs.ClusterId, rs.DepartmentId, name)
}

// ConfigMapProjectUrl returns the URL for ConfigMap with namespace id, cluster id and name
func (rs *RandomScopes) ConfigMapProjectUrl(name string) string {
	return fmt.Sprintf("/api/v1/config-maps/%s/namespaces/runai-%s%s", rs.ClusterId, rs.ProjectIdStr, name)
}

// ConfigMapsListUrlByScope returns the URL for listing ConfigMap with scope type, id and name as filter param
func (rs *RandomScopes) ConfigMapsListUrlByScope(scopeType orgUnitApi.ScopeType, name string) string {
	return fmt.Sprintf("/api/v1/config-maps?name=%s&scopeType=%s&scopeId=%s", name, scopeType, rs.ScopeId(scopeType))
}

// ConfigMapsListUrlByCluster returns the URL for listing ConfigMap with cluster id
func (rs *RandomScopes) ConfigMapsListUrlByCluster() string {
	return fmt.Sprintf("/api/v1/config-maps?clusterId=%s", rs.ClusterId)
}

// NewConfigMapsListResponse creates a new ConfigMap list responses, base on scope type and list of config map
// used to create expected list results
func (rs *RandomScopes) NewConfigMapsListResponse(scopeType orgUnitApi.ScopeType, configMaps []api.ConfigMap) api.ConfigMapsListResponse {
	return api.ConfigMapsListResponse{
		ConfigMaps: rs.configMapsListInfos(scopeType, configMaps),
	}
}

// configMapsListInfos helps to create a new ConfigMap list responses
func (rs *RandomScopes) configMapsListInfos(scopeType orgUnitApi.ScopeType, configMaps []api.ConfigMap) []api.ConfigMapListInfo {
	result := make([]api.ConfigMapListInfo, 0, len(configMaps))
	for _, c := range configMaps {
		result = append(result, api.ConfigMapListInfo{
			ClusterId: rs.ClusterId,
			ScopeType: string(scopeType),
			ScopeId:   rs.ScopeId(scopeType),
			Name:      c.Name,
		})
	}
	return result
}

func (rs *RandomScopes) ClusterObjectsStatusUrl(resourceType string, scopeId string, scopeType string, secretType string) string {
	url := fmt.Sprintf("/internal/api/v1/cluster-objects/status")
	if resourceType != "" || scopeId != "" || scopeType != "" || secretType != "" {
		url += "?"
	}
	addQuery := ""
	if resourceType != "" {
		url += fmt.Sprintf("resourceType=%s", resourceType)
		addQuery = "&"
	}
	if scopeId != "" {
		url += fmt.Sprintf("%sscopeId=%s", addQuery, scopeId)
		addQuery = "&"
	}
	if scopeType != "" {
		url += fmt.Sprintf("%sscopeType=%s", addQuery, scopeType)
		addQuery = "&"
	}
	if secretType != "" {
		url += fmt.Sprintf("%ssecretType=%s", addQuery, secretType)
	}
	return url
}

func (rs *RandomScopes) ScopeId(scopeType orgUnitApi.ScopeType) string {
	switch scopeType {
	case orgUnitApi.SCOPETYPE_CLUSTER:
		return rs.ClusterId
	case orgUnitApi.SCOPETYPE_DEPARTMENT:
		return rs.DepartmentId
	default:
		return rs.ProjectIdStr
	}
}

func headerWithMissingInfo() *http.Header {
	return &http.Header{}
}

func headerForAdmin(randoms *RandomScopes) *http.Header {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	return runaictx_http.GetHeaders(ctx)
}

func headerForNonAdmin(randoms *RandomScopes) *http.Header {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "just@a.user")
	return runaictx_http.GetHeaders(ctx)
}

// makePayload prepares a go object to be sent as a payload to REST call
func makePayload(payloadObject interface{}) *strings.Reader {
	payloadString, err := json.Marshal(payloadObject)
	Expect(err).To(BeNil())
	return strings.NewReader(string(payloadString))
}

func SetContext(tenantId string, subjectId string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	ctx = context.WithValue(ctx, authorizationUtils.PermittedScopesKey, &authorizationApi.PermittedScopes{
		Tenants: []string{tenantId},
	})
	return ctx
}
