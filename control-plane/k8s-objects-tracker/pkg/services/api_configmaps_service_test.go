/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/authorization"
	orgUnit "github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dbtesting"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("configmaps test", func() {
	var randomScopes *RandomScopes
	var anotherCluster, anotherDept, anotherProject, dept1, dept2, dept1proj1, dept1proj2, dept2proj1, dept2proj2 *RandomScopes

	BeforeEach(func() {
		router = CreateRouter()

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

		backendMock := NewRandomsBasedBackendMockClient(randomScopes, dept1, dept2, dept1proj1, dept1proj2, dept2proj1, dept2proj2,
			anotherCluster, anotherDept, anotherProject)
		orgUnit.SetMock(backendMock)
		authorization.SetMock(backendMock)

		dbtesting.CleanTables()
	})

	Context("permission issue when creating configmap", func() {
		It("should be rejected", func() {
			createHttpReq, err := http.NewRequest("POST", randomScopes.ConfigMapClusterUrl("/cm-1"), nil)
			createHttpReq.Header = *headerWithMissingInfo()
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("apply in multiple clusters", func() {
		It("each cluster shows its own configmap", func() {
			cmCluster1 := []api.ConfigMap{newConfigMap("cm-a"), newConfigMap("cm-c"), newConfigMap("cm-e")}
			cmCluster2 := []api.ConfigMap{newConfigMap("cm-b"), newConfigMap("cm-e")}

			createConfigMaps(randomScopes, randomScopes.ConfigMapClusterUrl(""), cmCluster1)
			createConfigMaps(anotherCluster, anotherCluster.ConfigMapClusterUrl(""), cmCluster2)

			expectedCluster1 := randomScopes.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, cmCluster1)
			// List by cluster id
			actualCluster1 := listConfigMaps(randomScopes, randomScopes.ConfigMapsListUrlByCluster())
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")
			// List by cluster scope
			actualCluster1 = listConfigMaps(randomScopes, randomScopes.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_CLUSTER, ""))
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1 by cluster scope")

			expectedCluster2 := anotherCluster.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, cmCluster2)
			actualCluster2 := listConfigMaps(anotherCluster, anotherCluster.ConfigMapsListUrlByCluster())
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")
			actualCluster2 = listConfigMaps(anotherCluster, anotherCluster.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_CLUSTER, ""))
			Expect(actualCluster2).To(Equal(actualCluster2), "cluster2 by cluster scope")
		})
	})
	Context("delete in multiple clusters", func() {
		It("each cluster has its own configmap", func() {
			configMaps := []api.ConfigMap{newConfigMap("cm-a"), newConfigMap("cm-b")}

			createConfigMaps(randomScopes, randomScopes.ConfigMapClusterUrl(""), configMaps)
			createConfigMaps(anotherCluster, anotherCluster.ConfigMapClusterUrl(""), configMaps)

			expectedCluster1 := randomScopes.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, configMaps)
			actualCluster1 := listConfigMaps(randomScopes, randomScopes.ConfigMapsListUrlByCluster())
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			expectedCluster2 := anotherCluster.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, configMaps)
			actualCluster2 := listConfigMaps(anotherCluster, anotherCluster.ConfigMapsListUrlByCluster())
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")

			//
			//   deletion of cm-a from cluster1
			//
			deleteConfigMap(randomScopes, randomScopes.ConfigMapClusterUrl("/cm-a"))

			expectedCluster1 = randomScopes.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, []api.ConfigMap{newConfigMap("cm-b")})
			actualCluster1 = listConfigMaps(randomScopes, randomScopes.ConfigMapsListUrlByCluster())
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			actualCluster2 = listConfigMaps(anotherCluster, anotherCluster.ConfigMapsListUrlByCluster())
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")

			//
			//   deletion of bbb from cluster2
			//
			deleteConfigMap(anotherCluster, anotherCluster.ConfigMapClusterUrl("/cm-b"))

			actualCluster1 = listConfigMaps(randomScopes, randomScopes.ConfigMapsListUrlByCluster())
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			expectedCluster2 = anotherCluster.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, []api.ConfigMap{newConfigMap("cm-a")})
			actualCluster2 = listConfigMaps(anotherCluster, anotherCluster.ConfigMapsListUrlByCluster())
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")
		})
	})
	Context("create and delete in multiple scopes", func() {
		It("each scope shows its own configmap", func() {
			cmCluster := []api.ConfigMap{newConfigMap("cm-e"), newConfigMap("cm-f")}
			cmDepts := [][]api.ConfigMap{
				{newConfigMap("cm-a"), newConfigMap("cm-c"), newConfigMap("cm-e")},
				{newConfigMap("cm-b"), newConfigMap("cm-c")},
			}
			cmProjects := [][]api.ConfigMap{
				{newConfigMap("cm-a"), newConfigMap("cm-d"), newConfigMap("cm-g"), newConfigMap("cm-h")},
				{},
				{newConfigMap("cm-h"), newConfigMap("cm-k"), newConfigMap("cm-l"), newConfigMap("cm-m")},
				{newConfigMap("cm-n")},
			}

			createConfigMaps(randomScopes, randomScopes.ConfigMapClusterUrl(""), cmCluster)
			createConfigMaps(dept1, dept1.ConfigMapDepartmentUrl(""), cmDepts[0])
			createConfigMaps(dept2, dept2.ConfigMapDepartmentUrl(""), cmDepts[1])
			createConfigMaps(dept1proj1, dept1proj1.ConfigMapProjectUrl(""), cmProjects[0])
			createConfigMaps(dept1proj2, dept1proj2.ConfigMapProjectUrl(""), cmProjects[1])
			createConfigMaps(dept2proj1, dept2proj1.ConfigMapProjectUrl(""), cmProjects[2])
			createConfigMaps(dept2proj2, dept2proj2.ConfigMapProjectUrl(""), cmProjects[3])

			expectedCluster := randomScopes.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, cmCluster)
			actualCluster := listConfigMaps(randomScopes, randomScopes.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_CLUSTER, ""))
			Expect(actualCluster).To(Equal(expectedCluster), "cluster")

			for idx, curDept := range []*RandomScopes{dept1, dept2} {
				expected := curDept.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, cmDepts[idx])
				actual := listConfigMaps(curDept, curDept.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_DEPARTMENT, ""))
				Expect(actual).To(Equal(expected), fmt.Sprintf("department-%v", idx))
			}

			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2, dept2proj1, dept2proj2} {
				expected := curProj.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_PROJECT, cmProjects[idx])
				actual := listConfigMaps(curProj, curProj.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_PROJECT, ""))
				Expect(actual).To(Equal(expected), fmt.Sprintf("project-%v", idx))
			}

			// Time to delete some config maps
			deleteConfigMap(dept1, dept1.ConfigMapDepartmentUrl("/cm-e"))
			cmDepts[0] = cmDepts[0][:len(cmDepts[0])-1]
			deleteConfigMap(dept2proj1, dept2proj1.ConfigMapProjectUrl("/cm-m"))
			cmProjects[2] = cmProjects[2][:len(cmProjects[2])-1]

			expectedCluster = randomScopes.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, cmCluster)
			actualCluster = listConfigMaps(randomScopes, randomScopes.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_CLUSTER, ""))
			Expect(actualCluster).To(Equal(expectedCluster), "round 2 cluster")

			for idx, curDept := range []*RandomScopes{dept1, dept2} {
				expected := curDept.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, cmDepts[idx])
				actual := listConfigMaps(curDept, curDept.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_DEPARTMENT, ""))
				Expect(actual).To(Equal(expected), fmt.Sprintf("round 2 department-%v", idx))
			}

			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2, dept2proj1, dept2proj2} {
				expected := curProj.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_PROJECT, cmProjects[idx])
				actual := listConfigMaps(curProj, curProj.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_PROJECT, ""))
				Expect(actual).To(Equal(expected), fmt.Sprintf("round 2 project-%v", idx))
			}

			namedExpected := randomScopes.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, []api.ConfigMap{newConfigMap("cm-e")})
			namedActual := listConfigMaps(randomScopes, randomScopes.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_CLUSTER, "cm-e"))
			Expect(namedActual).To(Equal(namedExpected), "filter by name")
		})
	})
	Context("sync test", func() {
		It("result is cm-a and cm-c (cm-b is removed)", func() {
			configMaps := []api.ConfigMap{newConfigMap("cm-a"), newConfigMap("cm-b")}
			createConfigMaps(dept1, dept1.ConfigMapDepartmentUrl(""), configMaps)

			expectedConfigMaps := dept1.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, configMaps)
			actualConfigMaps := listConfigMaps(dept1, dept1.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_DEPARTMENT, ""))
			Expect(actualConfigMaps).To(Equal(expectedConfigMaps), "department round 1")

			appliedConfigMaps := api.ConfigMaps{
				Configmaps: []api.ConfigMap{newConfigMap("cm-a"), newConfigMap("cm-c")},
			}
			applyConfigMaps(dept1, dept1.ConfigMapDepartmentUrl(""), appliedConfigMaps)

			expectedConfigMaps = dept1.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, appliedConfigMaps.Configmaps)
			actualConfigMaps = listConfigMaps(dept1, dept1.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_DEPARTMENT, ""))
			Expect(actualConfigMaps).To(Equal(expectedConfigMaps), "department after sync")
		})
	})

	Context("re-create same configmap", func() {
		It("should not change", func() {
			configMaps := []api.ConfigMap{newConfigMap("cm-a"), newConfigMap("cm-c"), newConfigMap("cm-e")}

			createConfigMaps(randomScopes, randomScopes.ConfigMapClusterUrl(""), configMaps)

			expectedCluster := randomScopes.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, configMaps)
			actualCluster := listConfigMaps(randomScopes, randomScopes.ConfigMapsListUrlByCluster())
			Expect(actualCluster).To(Equal(expectedCluster), "round 1")

			// Recreate the same configMaps
			createConfigMaps(randomScopes, randomScopes.ConfigMapClusterUrl(""), configMaps)

			expectedCluster = randomScopes.NewConfigMapsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, configMaps)
			actualCluster = listConfigMaps(randomScopes, randomScopes.ConfigMapsListUrlByCluster())
			Expect(expectedCluster).To(Equal(actualCluster), "round 2")
		})
	})
	Context("list + rbac", func() {
		It("auth rules must apply", func() {
			createConfigMaps(anotherCluster, anotherCluster.ConfigMapDepartmentUrl(""), []api.ConfigMap{newConfigMap("cm-a")})
			createConfigMaps(anotherDept, anotherDept.ConfigMapDepartmentUrl(""), []api.ConfigMap{newConfigMap("cm-a")})
			createConfigMaps(anotherProject, anotherProject.ConfigMapDepartmentUrl(""), []api.ConfigMap{newConfigMap("cm-a")})

			//
			//  forbidden cluster
			//
			urlCluster := randomScopes.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_CLUSTER, "")
			listClusterReq, err := http.NewRequest("GET", urlCluster, nil)
			listClusterReq.Header = *headerForNonAdmin(randomScopes)
			Expect(err).To(BeNil())

			listClusterResp := httptest.NewRecorder()
			router.ServeHTTP(listClusterResp, listClusterReq)
			Expect(listClusterResp.Code).To(Equal(http.StatusForbidden), urlCluster)

			//
			//  forbidden department
			//
			urlDept := randomScopes.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_DEPARTMENT, "")
			listDeptReq, err := http.NewRequest("GET", urlDept, nil)
			listDeptReq.Header = *headerForNonAdmin(randomScopes)
			Expect(err).To(BeNil())

			listDeptResp := httptest.NewRecorder()
			router.ServeHTTP(listDeptResp, listDeptReq)
			Expect(listClusterResp.Code).To(Equal(http.StatusForbidden), urlDept)

			//
			//  forbidden project
			//
			urlProject := randomScopes.ConfigMapsListUrlByScope(orgUnitApi.SCOPETYPE_PROJECT, "")
			listProjectReq, err := http.NewRequest("GET", urlProject, nil)
			listProjectReq.Header = *headerForNonAdmin(randomScopes)
			Expect(err).To(BeNil())

			listProjectResp := httptest.NewRecorder()
			router.ServeHTTP(listProjectResp, listProjectReq)
			Expect(listClusterResp.Code).To(Equal(http.StatusForbidden), urlProject)
		})
	})
	Context("project does not exist", func() {
		It("rejected", func() {
			nonExistProject := randomScopes.DifferentProject()

			createHttpReq, err := http.NewRequest("POST", nonExistProject.ConfigMapProjectUrl("/cm-a"), makePayload(newConfigMap("cm-a")))
			createHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			namespace := fmt.Sprintf("runai-%s", nonExistProject.ProjectIdStr)
			Expect(createHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("ProjectId for namespace %s not found", namespace)))

			deleteHttpReq, err := http.NewRequest("DELETE", nonExistProject.ConfigMapProjectUrl("/cm-a"), nil)
			deleteHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("ProjectId for namespace %s not found", namespace)))
		})
	})

	Context("missing both scopeType and clusterId", func() {
		It("should be rejected", func() {
			listHttpReq, err := http.NewRequest("GET", "/api/v1/config-maps", nil)
			listHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(listHttpResp.Body.String()).To(ContainSubstring("either scopeType or clusterId must be provided"))
		})

		Context("empty scope id when scopeType is provided", func() {
			It("should be rejected", func() {
				listHttpReq, err := http.NewRequest("GET", "/api/v1/config-maps?scopeType=cluster", nil)
				listHttpReq.Header = *headerForAdmin(randomScopes)
				Expect(err).To(BeNil())
				listHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listHttpResp, listHttpReq)
				Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(listHttpResp.Body.String()).To(ContainSubstring("scopeId is mandatory when scopeType is provided"))
			})
		})

		Context("invalid scope type", func() {
			It("should reject SQL injection payloads", func() {
				listHttpReq, err := http.NewRequest("GET", "/api/v1/config-maps?scopeType=z%27%29%20OR%20TRUE%20--&scopeId=x", nil)
				listHttpReq.Header = *headerForAdmin(randomScopes)
				Expect(err).To(BeNil())
				listHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listHttpResp, listHttpReq)
				Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(listHttpResp.Body.String()).To(ContainSubstring("invalid scopeType"))
			})
		})
	})
})

func newConfigMap(name string) api.ConfigMap {
	return api.ConfigMap{
		Name: name,
	}
}

func createConfigMaps(randomScopes *RandomScopes, url string, configMaps []api.ConfigMap) {
	for _, c := range configMaps {
		namedUrl := fmt.Sprintf("%s/%s", url, c.Name)
		syncHttpReq, err := http.NewRequest("POST", namedUrl, makePayload(c))
		syncHttpReq.Header = *headerForAdmin(randomScopes)

		Expect(err).To(BeNil())

		syncHttpResp := httptest.NewRecorder()
		router.ServeHTTP(syncHttpResp, syncHttpReq)

		Expect(syncHttpResp.Code).To(Equal(http.StatusOK), url)
	}
}

func applyConfigMaps(randomScopes *RandomScopes, url string, payload api.ConfigMaps) {
	syncHttpReq, err := http.NewRequest("POST", url, makePayload(payload))
	syncHttpReq.Header = *headerForAdmin(randomScopes)

	Expect(err).To(BeNil())

	syncHttpResp := httptest.NewRecorder()
	router.ServeHTTP(syncHttpResp, syncHttpReq)

	Expect(syncHttpResp.Code).To(Equal(http.StatusOK), url)
}

func listConfigMaps(randomScopes *RandomScopes, url string) api.ConfigMapsListResponse {
	listHttpReq, err := http.NewRequest("GET", url, nil)
	listHttpReq.Header = *headerForAdmin(randomScopes)

	Expect(err).To(BeNil(), url)

	listHttpResp := httptest.NewRecorder()
	router.ServeHTTP(listHttpResp, listHttpReq)

	Expect(listHttpResp.Code).To(Equal(http.StatusOK), url)

	var result api.ConfigMapsListResponse
	err = json.Unmarshal(listHttpResp.Body.Bytes(), &result)

	Expect(err).To(BeNil(), url)

	return result
}

func deleteConfigMap(randomScopes *RandomScopes, url string) {
	deleteHttpReq, err := http.NewRequest("DELETE", url, nil)
	deleteHttpReq.Header = *headerForAdmin(randomScopes)

	Expect(err).To(BeNil())

	deleteHttpResp := httptest.NewRecorder()
	router.ServeHTTP(deleteHttpResp, deleteHttpReq)

	Expect(deleteHttpResp.Code).To(Equal(http.StatusNoContent), url)
}
