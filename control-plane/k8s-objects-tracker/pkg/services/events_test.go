/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/authorization"
	orgUnit "github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dbtesting"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("event test", func() {
	var randomScopes *RandomScopes
	var anotherCluster, anotherDept, anotherProject, dept1, dept2, dept1proj1, dept1proj2, dept2proj1, dept2proj2, tenant1cluster1, tenant2cluster2 *RandomScopes

	BeforeEach(func() {
		randomScopes = NewRandomsSet()

		anotherCluster = randomScopes.DifferentCluster().Forbidden()
		anotherDept = randomScopes.DifferentDepartment().Forbidden()
		anotherProject = randomScopes.DifferentProject().Forbidden()

		dept1 = randomScopes.DifferentDepartment()
		dept1proj1 = dept1.DifferentProject()
		dept1proj2 = dept1proj1.DifferentProject()

		dept2 = randomScopes.DifferentDepartment()
		dept2proj1 = dept2.DifferentProject()
		dept2proj2 = dept2proj1.DifferentProject()

		tenant1cluster1 = randomScopes
		tenant2cluster2 = NewRandomsSetWithDifferentTenant()

		backendMock := NewRandomsBasedBackendMockClient(randomScopes, dept1, dept2, dept1proj1, dept1proj2, dept2proj1, dept2proj2,
			anotherCluster, anotherDept, anotherProject)
		orgUnit.SetMock(backendMock)
		authorization.SetMock(backendMock)

		router = CreateRouter()

		dbtesting.CleanTables()
	})

	Context("test project deletion", func() {
		It("basic flow", func() {
			// Inject config maps
			configMaps := []api.ConfigMaps{
				{Configmaps: []api.ConfigMap{newConfigMap("cm-a"), newConfigMap("cm-b"), newConfigMap("cm-c")}},
				{Configmaps: []api.ConfigMap{newConfigMap("cm-d"), newConfigMap("cm-e"), newConfigMap("cm-f")}},
			}
			applyConfigMaps(dept1proj1, dept1proj1.ConfigMapProjectUrl(""), configMaps[0])
			applyConfigMaps(dept1proj2, dept1proj2.ConfigMapProjectUrl(""), configMaps[1])
			// Verify config maps were injected
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_PROJECT, configMaps[idx].Configmaps)
				actual := listConfigMaps(curProj, curProj.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_PROJECT, ""))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}

			// Inject pvcs
			pvcs := []api.Pvcs{
				{Pvcs: []api.Pvc{NewExistingPvc("aaa"), NewExistingPvc("ddd"), NewExistingPvc("ggg"), NewExistingPvc("hhh")}},
				{Pvcs: []api.Pvc{NewExistingPvc("hhh"), NewExistingPvc("kkk"), NewExistingPvc("lll"), NewExistingPvc("mmm")}},
			}
			applyPvcs(dept1proj1, dept1proj1.PvcsProjectUrl(""), pvcs[0])
			applyPvcs(dept1proj2, dept1proj2.PvcsProjectUrl(""), pvcs[1])
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.PvcsListResponse(orgUnitApi.SCOPETYPE_PROJECT, pvcs[idx].Pvcs)
				actual := listPvcs(curProj, curProj.PvcsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}

			// Inject secret
			secrets := []api.Secrets{
				{Secrets: []api.Secret{{Name: "aaa", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}, {Name: "ccc", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting}, {Name: "eee", Type: api.SECRETTYPE_DOCKER_REGISTRY, Origin: &originExisting}}},
				{Secrets: []api.Secret{{Name: "bbb", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}, {Name: "ddd", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting}, {Name: "fff", Type: api.SECRETTYPE_DOCKER_REGISTRY, Origin: &originExisting}}},
			}
			applySecrets(dept1proj1, dept1proj1.SecretsProjectUrl(""), secrets[0])
			applySecrets(dept1proj2, dept1proj2.SecretsProjectUrl(""), secrets[1])
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.SecretsListResponse(orgUnitApi.SCOPETYPE_PROJECT, secrets[idx].Secrets...)
				actual := listSecrets(curProj, curProj.SecretsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}

			// Call project deleted event
			dept1proj1Ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", dept1proj1.TenantId))
			delErr := HandleProjectDeletedEvent(dept1proj1Ctx, &events.ProjectDeleted{
				ProjectId: fmt.Sprintf("%d", dept1proj1.ProjectId),
			})
			Expect(delErr).To(BeNil())
			// Verify that everything that needed to be deleted was deleted, and only that

			// Verify config maps were deleted
			configMaps[0] = api.ConfigMaps{}
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_PROJECT, configMaps[idx].Configmaps)
				actual := listConfigMaps(curProj, curProj.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_PROJECT, ""))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}

			// Verify PVCS were deleted
			pvcs[0] = api.Pvcs{}
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.PvcsListResponse(orgUnitApi.SCOPETYPE_PROJECT, pvcs[idx].Pvcs)
				actual := listPvcs(curProj, curProj.PvcsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			// Verify Secrets were deleted
			secrets[0] = api.Secrets{}
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.SecretsListResponse(orgUnitApi.SCOPETYPE_PROJECT, secrets[idx].Secrets...)
				actual := listSecrets(curProj, curProj.SecretsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}

		})
	})

	Context("test cluster deletion", func() {
		It("basic flow", func() {
			// Inject config maps
			configMaps := []api.ConfigMaps{
				{Configmaps: []api.ConfigMap{newConfigMap("cm-a"), newConfigMap("cm-b"), newConfigMap("cm-c")}},
				{Configmaps: []api.ConfigMap{newConfigMap("cm-d"), newConfigMap("cm-e"), newConfigMap("cm-f")}},
			}
			applyConfigMaps(dept1proj1, dept1proj1.ConfigMapProjectUrl(""), configMaps[0])
			applyConfigMaps(dept1proj2, dept1proj2.ConfigMapProjectUrl(""), configMaps[1])
			// Verify config maps were injected
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_PROJECT, configMaps[idx].Configmaps)
				actual := listConfigMaps(curProj, curProj.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_PROJECT, ""))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}

			// Inject pvcs in project scope
			projPvcs := []api.Pvcs{
				{Pvcs: []api.Pvc{NewExistingPvc("aaa"), NewExistingPvc("ddd"), NewExistingPvc("ggg"), NewExistingPvc("hhh")}},
				{Pvcs: []api.Pvc{NewExistingPvc("hhh"), NewExistingPvc("kak"), NewExistingPvc("lll"), NewExistingPvc("mmm")}},
			}
			applyPvcs(dept1proj1, dept1proj1.PvcsProjectUrl(""), projPvcs[0])
			applyPvcs(dept1proj2, dept1proj2.PvcsProjectUrl(""), projPvcs[1])
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.PvcsListResponse(orgUnitApi.SCOPETYPE_PROJECT, projPvcs[idx].Pvcs)
				actual := listPvcs(curProj, curProj.PvcsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			// Inject pvcs in cluster scope
			clusterPvcs := []api.Pvcs{
				{Pvcs: []api.Pvc{NewExistingPvc("123"), NewExistingPvc("789")}},
				{Pvcs: []api.Pvc{NewExistingPvc("456"), NewExistingPvc("987")}},
			}
			applyPvcs(randomScopes, randomScopes.PvcsClusterUrl(""), clusterPvcs[0])
			applyPvcs(anotherCluster, anotherCluster.PvcsClusterUrl(""), clusterPvcs[1])
			for idx, curCluster := range []*RandomScopes{randomScopes, anotherCluster} {
				expected := curCluster.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterPvcs[idx].Pvcs)
				actual := listPvcs(curCluster, curCluster.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curCluster.ClusterId)
			}

			// Inject secret in project scope
			projSecrets := []api.Secrets{
				{Secrets: []api.Secret{{Name: "aaa", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}, {Name: "ccc", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting}, {Name: "eee", Type: api.SECRETTYPE_DOCKER_REGISTRY, Origin: &originExisting}}},
				{Secrets: []api.Secret{{Name: "bbb", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}, {Name: "ddd", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting}, {Name: "fff", Type: api.SECRETTYPE_DOCKER_REGISTRY, Origin: &originExisting}}},
			}
			applySecrets(dept1proj1, dept1proj1.SecretsProjectUrl(""), projSecrets[0])
			applySecrets(dept1proj2, dept1proj2.SecretsProjectUrl(""), projSecrets[1])
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.SecretsListResponse(orgUnitApi.SCOPETYPE_PROJECT, projSecrets[idx].Secrets...)
				actual := listSecrets(curProj, curProj.SecretsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			clusterSecrets := []api.Secrets{
				{Secrets: []api.Secret{{Name: "secret-1", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}}},
				{Secrets: []api.Secret{{Name: "secret-2", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}}},
			}
			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), clusterSecrets[0])
			applySecrets(anotherCluster, anotherCluster.SecretsClusterUrl(""), clusterSecrets[1])
			for idx, curCluster := range []*RandomScopes{randomScopes, anotherCluster} {
				expected := curCluster.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterSecrets[idx].Secrets...)
				actual := listSecrets(curCluster, curCluster.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curCluster.ClusterId)
			}

			// Call project deleted event
			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", randomScopes.TenantId))
			delErr := HandleClusterDeleteEvent(ctx, &events.ClusterDelete{
				ClusterId: randomScopes.ClusterId,
			})
			Expect(delErr).To(BeNil())
			// Verify that everything that needed to be deleted was deleted, and only that

			// Verify config maps were deleted
			configMaps[0] = api.ConfigMaps{}
			configMaps[1] = api.ConfigMaps{}
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_PROJECT, configMaps[idx].Configmaps)
				actual := listConfigMaps(curProj, curProj.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_PROJECT, ""))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}

			// Verify PVCS were deleted
			projPvcs[0] = api.Pvcs{}
			projPvcs[1] = api.Pvcs{}
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.PvcsListResponse(orgUnitApi.SCOPETYPE_PROJECT, projPvcs[idx].Pvcs)
				actual := listPvcs(curProj, curProj.PvcsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			clusterPvcs[0] = api.Pvcs{}
			for idx, curCluster := range []*RandomScopes{randomScopes, anotherCluster} {
				expected := curCluster.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterPvcs[idx].Pvcs)
				actual := listPvcs(curCluster, curCluster.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curCluster.ClusterId)
			}

			// Verify Secrets were deleted
			projSecrets[0] = api.Secrets{}
			projSecrets[1] = api.Secrets{}
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.SecretsListResponse(orgUnitApi.SCOPETYPE_PROJECT, projSecrets[idx].Secrets...)
				actual := listSecrets(curProj, curProj.SecretsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			clusterSecrets[0] = api.Secrets{}
			for idx, curCluster := range []*RandomScopes{randomScopes, anotherCluster} {
				expected := curCluster.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterSecrets[idx].Secrets...)
				actual := listSecrets(curCluster, curCluster.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curCluster.ClusterId)
			}

		})
	})
	Context("test tenant deletion", func() {
		It("basic flow", func() {
			// Inject project scoped config maps
			projectConfigMaps := []api.ConfigMaps{
				{Configmaps: []api.ConfigMap{newConfigMap("cm-proj1-1"), newConfigMap("cm-proj1-2")}},
				{Configmaps: []api.ConfigMap{newConfigMap("cm-proj2-1"), newConfigMap("cm-proj2-2")}},
			}
			applyConfigMaps(dept1proj1, dept1proj1.ConfigMapProjectUrl(""), projectConfigMaps[0])
			applyConfigMaps(dept1proj2, dept1proj2.ConfigMapProjectUrl(""), projectConfigMaps[1])
			// Verify config maps were injected
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_PROJECT, projectConfigMaps[idx].Configmaps)
				actual := listConfigMaps(curProj, curProj.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_PROJECT, ""))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			//Inject cluster scoped config maps
			clusterConfigMaps := []api.ConfigMaps{
				{Configmaps: []api.ConfigMap{newConfigMap("cm-tenant1cluster1-1"), newConfigMap("cm-tenant1cluster1-2")}},
				{Configmaps: []api.ConfigMap{newConfigMap("cm-tenant2cluster2-1"), newConfigMap("cm-tenant2cluster2-2")}},
			}
			applyConfigMaps(tenant1cluster1, tenant1cluster1.ConfigMapClusterUrl(""), clusterConfigMaps[0])
			applyConfigMaps(tenant2cluster2, tenant2cluster2.ConfigMapClusterUrl(""), clusterConfigMaps[1])
			// Verify config maps were injected
			for idx, curCluster := range []*RandomScopes{tenant1cluster1, tenant2cluster2} {
				expected := curCluster.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterConfigMaps[idx].Configmaps)
				actual := listConfigMaps(curCluster, curCluster.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_CLUSTER, ""))
				Expect(actual).To(Equal(expected), curCluster.ClusterId)
			}

			// Inject pvcs in project scope
			projPvcs := []api.Pvcs{
				{Pvcs: []api.Pvc{NewExistingPvc("pvc-proj1-1"), NewExistingPvc("pvc-proj1-2")}},
				{Pvcs: []api.Pvc{NewExistingPvc("pvc-proj2-1"), NewExistingPvc("pvc-proj2-2")}},
			}
			applyPvcs(dept1proj1, dept1proj1.PvcsProjectUrl(""), projPvcs[0])
			applyPvcs(dept1proj2, dept1proj2.PvcsProjectUrl(""), projPvcs[1])
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.PvcsListResponse(orgUnitApi.SCOPETYPE_PROJECT, projPvcs[idx].Pvcs)
				actual := listPvcs(curProj, curProj.PvcsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			// Inject pvcs in cluster scope
			clusterPvcs := []api.Pvcs{
				{Pvcs: []api.Pvc{NewExistingPvc("pvc-tenant1cluster1-1"), NewExistingPvc("pvc-tenant1cluster1-2")}},
				{Pvcs: []api.Pvc{NewExistingPvc("pvc-tenant2cluster2-1"), NewExistingPvc("pvc-tenant2cluster2-2")}},
			}
			applyPvcs(tenant1cluster1, tenant1cluster1.PvcsClusterUrl(""), clusterPvcs[0])
			applyPvcs(tenant2cluster2, tenant2cluster2.PvcsClusterUrl(""), clusterPvcs[1])
			for idx, curCluster := range []*RandomScopes{tenant1cluster1, tenant2cluster2} {
				expected := curCluster.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterPvcs[idx].Pvcs)
				actual := listPvcs(curCluster, curCluster.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curCluster.ClusterId)
			}
			// Inject secret in project scope
			projSecrets := []api.Secrets{
				{Secrets: []api.Secret{{Name: "ak-proj1-1", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}, {Name: "ak-proj1-2", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting}}},
				{Secrets: []api.Secret{{Name: "ak-proj2-1", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}, {Name: "ak-proj2-2", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting}}},
			}
			applySecrets(dept1proj1, dept1proj1.SecretsProjectUrl(""), projSecrets[0])
			applySecrets(dept1proj2, dept1proj2.SecretsProjectUrl(""), projSecrets[1])
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				expected := curProj.SecretsListResponse(orgUnitApi.SCOPETYPE_PROJECT, projSecrets[idx].Secrets...)
				actual := listSecrets(curProj, curProj.SecretsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			// Inject secret in cluster scope
			clusterSecrets := []api.Secrets{
				{Secrets: []api.Secret{{Name: "ak-tenant1cluster1", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}}},
				{Secrets: []api.Secret{{Name: "ak-tenant2cluster2", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}}},
			}
			// Inject secret in cluster scope
			applySecrets(tenant1cluster1, tenant1cluster1.SecretsClusterUrl(""), clusterSecrets[0])
			applySecrets(tenant2cluster2, tenant2cluster2.SecretsClusterUrl(""), clusterSecrets[1])
			for idx, curCluster := range []*RandomScopes{tenant1cluster1, tenant2cluster2} {
				expected := curCluster.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterSecrets[idx].Secrets...)
				actual := listSecrets(curCluster, curCluster.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curCluster.ClusterId)
			}

			// Call tenant deleted event
			ctx := context.Background()
			delErr := HandleTenantDeleteEvent(ctx, &events.TenantDelete{
				TenantId: fmt.Sprintf("%d", randomScopes.TenantId),
			})
			Expect(delErr).To(BeNil())
			// Verify that everything that needed to be deleted was deleted, and only that

			// Verify project scoped config maps were deleted
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				projectConfigMaps[idx] = api.ConfigMaps{}
				expected := curProj.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_PROJECT, projectConfigMaps[idx].Configmaps)
				actual := listConfigMaps(curProj, curProj.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_PROJECT, ""))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			//Verify cluster scoped config maps were deleted
			clusterConfigMaps[0] = api.ConfigMaps{}
			for idx, curCluster := range []*RandomScopes{tenant1cluster1, tenant2cluster2} {
				expected := curCluster.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterConfigMaps[idx].Configmaps)
				actual := listConfigMaps(curCluster, curCluster.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_CLUSTER, ""))
				Expect(actual).To(Equal(expected), curCluster.ClusterId)
			}

			// Verify PVCS were deleted
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				projPvcs[idx] = api.Pvcs{}
				expected := curProj.PvcsListResponse(orgUnitApi.SCOPETYPE_PROJECT, projPvcs[idx].Pvcs)
				actual := listPvcs(curProj, curProj.PvcsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			clusterPvcs[0] = api.Pvcs{}
			for idx, curCluster := range []*RandomScopes{tenant1cluster1, tenant2cluster2} {
				expected := curCluster.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterPvcs[idx].Pvcs)
				actual := listPvcs(curCluster, curCluster.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curCluster.ClusterId)
			}

			// Verify Secrets were deleted
			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2} {
				projSecrets[idx] = api.Secrets{}
				expected := curProj.SecretsListResponse(orgUnitApi.SCOPETYPE_PROJECT, projSecrets[idx].Secrets...)
				actual := listSecrets(curProj, curProj.SecretsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
			clusterSecrets[0] = api.Secrets{}
			for idx, curCluster := range []*RandomScopes{tenant1cluster1, tenant2cluster2} {
				expected := curCluster.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterSecrets[idx].Secrets...)
				actual := listSecrets(curCluster, curCluster.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curCluster.ClusterId)
			}
		})
	})
})
