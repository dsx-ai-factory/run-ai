/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/authorization"
	orgUnit "github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dbtesting"
)

var _ = Describe("cluster objects status test", func() {
	var randomScopes *RandomScopes
	var anotherCluster, anotherDept, anotherProject, dept1, dept2, dept1proj1, dept1proj2, dept2proj1, dept2proj2 *RandomScopes

	lostPhaseStr := string(api.CLUSTEROBJECTISSUETYPE_LOST_PHASE)
	replicationErrorStr := string(api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR)

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

	Context("PVC", func() {
		Context("Cluster scope", func() {
			Context("create with issues", func() {
				It("should get the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr), NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)}
					clusterScopePvc := existingPvc
					clusterScopePvc.Issues = issues

					createPvc(randomScopes, randomScopes.PvcsClusterUrl("/aaa"), clusterScopePvc)
					pvcs := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{
						{
							ScopeType: "cluster",
							ScopeId:   randomScopes.ClusterId,
							Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
						},
						{
							ScopeType: "project",
							ScopeId:   dept1proj2.ProjectIdStr,
							Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
						},
					}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))
				})
				It("should update the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)}
					differentIssues := []api.Issue{NewClusterResourceIssue("runai", replicationErrorStr)}

					clusterScopePvc := existingPvc
					clusterScopePvc.Issues = issues
					createPvc(randomScopes, randomScopes.PvcsClusterUrl("/aaa"), clusterScopePvc)
					pvcs := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{{
						ScopeType: "cluster",
						ScopeId:   randomScopes.ClusterId,
						Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
					}}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))

					clusterScopePvc.Issues = differentIssues
					createPvc(randomScopes, randomScopes.PvcsClusterUrl("/aaa"), clusterScopePvc)
					pvcs = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))

					expectedIssues = []api.ClusterObjectIssue{{
						ScopeType: "cluster",
						ScopeId:   randomScopes.ClusterId,
						Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
					}}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))
				})
				It("should have only the issues related to each resource", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					clusterScopePvc1 := api.Pvc{
						Name:   "aaa",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					clusterScopePvc2 := api.Pvc{
						Name:   "bbb",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					pvcs := api.Pvcs{Pvcs: []api.Pvc{clusterScopePvc1, clusterScopePvc2}}
					applyPvcs(randomScopes, randomScopes.PvcsClusterUrl(""), pvcs)
					pvcsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}

					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))
				})
				It("should update the issues according to the resources", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					clusterScopePvc1 := api.Pvc{
						Name:   "aaa",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					clusterScopePvc2 := api.Pvc{
						Name:   "bbb",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					pvcs := api.Pvcs{Pvcs: []api.Pvc{clusterScopePvc1, clusterScopePvc2}}
					applyPvcs(randomScopes, randomScopes.PvcsClusterUrl(""), pvcs)
					pvcsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}
					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))

					clusterScopePvc1.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, replicationErrorStr)}
					clusterScopePvc2.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, lostPhaseStr)}
					pvcs = api.Pvcs{Pvcs: []api.Pvc{clusterScopePvc1, clusterScopePvc2}}
					applyPvcs(randomScopes, randomScopes.PvcsClusterUrl(""), pvcs)
					pvcsList = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))
					expectedResponse = []api.ClusterObjectStatus{
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
					}
					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))
				})
			})
		})
		Context("Department scope", func() {
			Context("create with issues", func() {
				It("should get the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr), NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)}
					departmentScopePvc := existingPvc
					departmentScopePvc.Issues = issues

					createPvc(randomScopes, randomScopes.PvcsDepartmentUrl("/aaa"), departmentScopePvc)
					pvcs := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{
						{
							ScopeType: "department",
							ScopeId:   randomScopes.DepartmentId,
							Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
						},
						{
							ScopeType: "project",
							ScopeId:   dept1proj2.ProjectIdStr,
							Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
						},
					}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))
				})
				It("should update the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)}
					differentIssues := []api.Issue{NewClusterResourceIssue("runai", replicationErrorStr)}

					departmentScopePvc := existingPvc
					departmentScopePvc.Issues = issues
					createPvc(randomScopes, randomScopes.PvcsDepartmentUrl("/aaa"), departmentScopePvc)
					pvcs := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{{
						ScopeType: "department",
						ScopeId:   randomScopes.DepartmentId,
						Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
					}}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))

					departmentScopePvc.Issues = differentIssues
					createPvc(randomScopes, randomScopes.PvcsDepartmentUrl("/aaa"), departmentScopePvc)
					pvcs = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))

					expectedIssues = []api.ClusterObjectIssue{{
						ScopeType: "department",
						ScopeId:   randomScopes.DepartmentId,
						Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
					}}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))
				})
				It("should have only the issues related to each resource", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					clusterScopePvc1 := api.Pvc{
						Name:   "aaa",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					clusterScopePvc2 := api.Pvc{
						Name:   "bbb",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					pvcs := api.Pvcs{Pvcs: []api.Pvc{clusterScopePvc1, clusterScopePvc2}}
					applyPvcs(randomScopes, randomScopes.PvcsDepartmentUrl(""), pvcs)
					pvcsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}

					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))
				})
				It("should update the issues according to the resources", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					departmentScopePvc1 := api.Pvc{
						Name:   "aaa",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					departmentScopePvc2 := api.Pvc{
						Name:   "bbb",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					pvcs := api.Pvcs{Pvcs: []api.Pvc{departmentScopePvc1, departmentScopePvc2}}
					applyPvcs(randomScopes, randomScopes.PvcsDepartmentUrl(""), pvcs)
					pvcsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}
					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))

					departmentScopePvc1.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, replicationErrorStr)}
					departmentScopePvc2.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, lostPhaseStr)}
					pvcs = api.Pvcs{Pvcs: []api.Pvc{departmentScopePvc1, departmentScopePvc2}}
					applyPvcs(randomScopes, randomScopes.PvcsDepartmentUrl(""), pvcs)
					pvcsList = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))
					expectedResponse = []api.ClusterObjectStatus{
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
					}
					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))
				})
			})
		})
		Context("Project scope", func() {
			Context("create with issues", func() {
				It("should get the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai-"+randomScopes.ProjectIdStr, lostPhaseStr)}
					projectScopePvc := existingPvc
					projectScopePvc.Issues = issues

					createPvc(randomScopes, randomScopes.PvcsProjectUrl("/aaa"), projectScopePvc)
					pvcs := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{{
						ScopeType: "project",
						ScopeId:   randomScopes.ProjectIdStr,
						Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
					}}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))
				})
				It("should update the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai-"+randomScopes.ProjectIdStr, lostPhaseStr)}
					differentIssues := []api.Issue{NewClusterResourceIssue("runai-"+randomScopes.ProjectIdStr, replicationErrorStr)}
					projectScopePvc := existingPvc
					projectScopePvc.Issues = issues

					createPvc(randomScopes, randomScopes.PvcsProjectUrl("/aaa"), projectScopePvc)
					pvcs := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{{
						ScopeType: "project",
						ScopeId:   randomScopes.ProjectIdStr,
						Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
					}}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))

					projectScopePvc.Issues = differentIssues
					createPvc(randomScopes, randomScopes.PvcsProjectUrl("/aaa"), projectScopePvc)
					pvcs = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))

					expectedIssues = []api.ClusterObjectIssue{{
						ScopeType: "project",
						ScopeId:   randomScopes.ProjectIdStr,
						Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
					}}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))
				})
				It("should have only the issues related to each resource", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					projectScopePvc1 := api.Pvc{
						Name:   "aaa",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					projectScopePvc2 := api.Pvc{
						Name:   "bbb",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					pvcs := api.Pvcs{Pvcs: []api.Pvc{projectScopePvc1, projectScopePvc2}}
					applyPvcs(randomScopes, randomScopes.PvcsProjectUrl(""), pvcs)
					pvcsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}
					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))
				})
				It("should update the issues according to the resources", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					projectScopePvc1 := api.Pvc{
						Name:   "aaa",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					projectScopePvc2 := api.Pvc{
						Name:   "bbb",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					pvcs := api.Pvcs{Pvcs: []api.Pvc{projectScopePvc1, projectScopePvc2}}
					applyPvcs(randomScopes, randomScopes.PvcsProjectUrl(""), pvcs)
					pvcsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}

					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))

					projectScopePvc1.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, replicationErrorStr)}
					projectScopePvc2.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, lostPhaseStr)}
					pvcs = api.Pvcs{Pvcs: []api.Pvc{projectScopePvc1, projectScopePvc2}}
					applyPvcs(randomScopes, randomScopes.PvcsProjectUrl(""), pvcs)
					pvcsList = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))
					expectedResponse = []api.ClusterObjectStatus{
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_PVC,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
					}

					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))
				})
			})
		})
	})
	Context("Config Map", func() {
		Context("Project scope", func() {
			Context("create with issues", func() {
				It("should get the resource with the issues related to it", func() {
					projectScopeConfigMap := api.ConfigMap{
						Name:   "aaa",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+randomScopes.ProjectIdStr, replicationErrorStr)},
					}

					createConfigMaps(randomScopes, randomScopes.ConfigMapProjectUrl(""), []api.ConfigMap{projectScopeConfigMap})
					configMaps := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("config-map", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{{
						ScopeType: "project",
						ScopeId:   randomScopes.ProjectIdStr,
						Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
					}}
					Expect(configMaps[0].Issues).To(Equal(expectedIssues))
				})
				It("should update the resource with the issues related to it", func() {
					projectScopeConfigMap := api.ConfigMap{
						Name:   "aaa",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+randomScopes.ProjectIdStr, replicationErrorStr)},
					}

					createConfigMaps(randomScopes, randomScopes.ConfigMapProjectUrl(""), []api.ConfigMap{projectScopeConfigMap})
					configMaps := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("config-map", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{{
						ScopeType: "project",
						ScopeId:   randomScopes.ProjectIdStr,
						Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
					}}
					Expect(configMaps[0].Issues).To(Equal(expectedIssues))

					projectScopeConfigMap.Issues = []api.Issue{}
					createConfigMaps(randomScopes, randomScopes.ConfigMapProjectUrl(""), []api.ConfigMap{projectScopeConfigMap})
					configMaps = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("config-map", "", "", ""))

					Expect(len(configMaps[0].Issues)).To(Equal(0))
				})
				It("should have only the issues related to each resource", func() {
					projectScopeConfigMap1 := api.ConfigMap{
						Name:   "aaa",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
					}
					projectScopeConfigMap2 := api.ConfigMap{
						Name:   "bbb",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
					}

					configMaps := api.ConfigMaps{Configmaps: []api.ConfigMap{projectScopeConfigMap1, projectScopeConfigMap2}}
					applyConfigMaps(randomScopes, randomScopes.ConfigMapProjectUrl(""), configMaps)
					pvcsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("config-map", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_CONFIG_MAP,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_CONFIG_MAP,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}
					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))
				})
				It("should update the issues according to the resources", func() {
					projectScopeConfigMap1 := api.ConfigMap{
						Name:   "aaa",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
					}
					projectScopeConfigMap2 := api.ConfigMap{
						Name:   "bbb",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
					}

					configMaps := api.ConfigMaps{Configmaps: []api.ConfigMap{projectScopeConfigMap1, projectScopeConfigMap2}}
					applyConfigMaps(randomScopes, randomScopes.ConfigMapProjectUrl(""), configMaps)
					configMapList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("config-map", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_CONFIG_MAP,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_CONFIG_MAP,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}

					sortListsForEqualityCheck(configMapList, expectedResponse)
					Expect(configMapList).To(Equal(expectedResponse))

					projectScopeConfigMap1.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, replicationErrorStr)}
					projectScopeConfigMap2.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, lostPhaseStr)}
					configMaps = api.ConfigMaps{Configmaps: []api.ConfigMap{projectScopeConfigMap1, projectScopeConfigMap2}}
					applyConfigMaps(randomScopes, randomScopes.ConfigMapProjectUrl(""), configMaps)
					configMapList = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("config-map", "", "", ""))
					expectedResponse = []api.ClusterObjectStatus{
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_CONFIG_MAP,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_CONFIG_MAP,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
					}

					sortListsForEqualityCheck(configMapList, expectedResponse)
					Expect(configMapList).To(Equal(expectedResponse))
				})
			})
		})
	})
	Context("Secret", func() {
		Context("Cluster scope", func() {
			Context("create with issues", func() {
				It("should get the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr), NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)}
					clusterScopeSecret := existingPassword
					clusterScopeSecret.Issues = issues

					createSecret(randomScopes, randomScopes.SecretsClusterUrl("/aaa"), clusterScopeSecret)
					pvcs := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{
						{
							ScopeType: "cluster",
							ScopeId:   randomScopes.ClusterId,
							Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
						},
						{
							ScopeType: "project",
							ScopeId:   dept1proj2.ProjectIdStr,
							Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
						},
					}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))
				})
				It("should update the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)}
					differentIssues := []api.Issue{NewClusterResourceIssue("runai", replicationErrorStr)}

					clusterScopeSecret := existingPassword
					clusterScopeSecret.Issues = issues
					createSecret(randomScopes, randomScopes.SecretsClusterUrl("/aaa"), clusterScopeSecret)
					pvcs := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{{
						ScopeType: "cluster",
						ScopeId:   randomScopes.ClusterId,
						Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
					}}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))

					clusterScopeSecret.Issues = differentIssues
					createSecret(randomScopes, randomScopes.SecretsClusterUrl("/aaa"), clusterScopeSecret)
					pvcs = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))

					expectedIssues = []api.ClusterObjectIssue{{
						ScopeType: "cluster",
						ScopeId:   randomScopes.ClusterId,
						Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
					}}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))
				})
				It("should have only the issues related to each resource", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					clusterScopeSecret1 := api.Secret{
						Name:   "aaa",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					clusterScopeSecret2 := api.Secret{
						Name:   "bbb",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					secrets := api.Secrets{Secrets: []api.Secret{clusterScopeSecret1, clusterScopeSecret2}}
					applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), secrets)
					pvcsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}

					sortListsForEqualityCheck(pvcsList, expectedResponse)
					Expect(pvcsList).To(Equal(expectedResponse))
				})
				It("should update the issues according to the resources", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					clusterScopeSecret1 := api.Secret{
						Name:   "aaa",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					clusterScopeSecret2 := api.Secret{
						Name:   "bbb",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					secrets := api.Secrets{Secrets: []api.Secret{clusterScopeSecret1, clusterScopeSecret2}}
					applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), secrets)
					secretList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}
					sortListsForEqualityCheck(secretList, expectedResponse)
					Expect(secretList).To(Equal(expectedResponse))

					clusterScopeSecret1.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, replicationErrorStr)}
					clusterScopeSecret2.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, lostPhaseStr)}
					secrets = api.Secrets{Secrets: []api.Secret{clusterScopeSecret1, clusterScopeSecret2}}
					applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), secrets)
					secretList = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
					expectedResponse = []api.ClusterObjectStatus{
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
						{
							ScopeType:    "cluster",
							ScopeId:      randomScopes.ClusterId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
					}
					sortListsForEqualityCheck(secretList, expectedResponse)
					Expect(secretList).To(Equal(expectedResponse))
				})
			})
		})
		Context("Department scope", func() {
			Context("create with issues", func() {
				It("should get the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr), NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)}
					departmentScopeSecret := existingPassword
					departmentScopeSecret.Issues = issues

					createSecret(randomScopes, randomScopes.SecretsDepartmentUrl("/aaa"), departmentScopeSecret)
					secrets := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{
						{
							ScopeType: "department",
							ScopeId:   randomScopes.DepartmentId,
							Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
						},
						{
							ScopeType: "project",
							ScopeId:   dept1proj2.ProjectIdStr,
							Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
						},
					}
					Expect(secrets[0].Issues).To(Equal(expectedIssues))
				})
				It("should update the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)}
					differentIssues := []api.Issue{NewClusterResourceIssue("runai", replicationErrorStr)}

					departmentScopeSecret := existingPassword
					departmentScopeSecret.Issues = issues
					createSecret(randomScopes, randomScopes.SecretsDepartmentUrl("/aaa"), departmentScopeSecret)
					secrets := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{{
						ScopeType: "department",
						ScopeId:   randomScopes.DepartmentId,
						Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
					}}
					Expect(secrets[0].Issues).To(Equal(expectedIssues))

					departmentScopeSecret.Issues = differentIssues
					createSecret(randomScopes, randomScopes.SecretsDepartmentUrl("/aaa"), departmentScopeSecret)
					secrets = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))

					expectedIssues = []api.ClusterObjectIssue{{
						ScopeType: "department",
						ScopeId:   randomScopes.DepartmentId,
						Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
					}}
					Expect(secrets[0].Issues).To(Equal(expectedIssues))
				})
				It("should have only the issues related to each resource", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					clusterScopeSecret1 := api.Secret{
						Name:   "aaa",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					clusterScopeSecret2 := api.Secret{
						Name:   "bbb",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					secrets := api.Secrets{Secrets: []api.Secret{clusterScopeSecret1, clusterScopeSecret2}}
					applySecrets(randomScopes, randomScopes.SecretsDepartmentUrl(""), secrets)
					secretsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}

					sortListsForEqualityCheck(secretsList, expectedResponse)
					Expect(secretsList).To(Equal(expectedResponse))
				})
				It("should update the issues according to the resources", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					departmentScopeSecret1 := api.Secret{
						Name:   "aaa",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					departmentScopeSecret2 := api.Secret{
						Name:   "bbb",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					secrets := api.Secrets{Secrets: []api.Secret{departmentScopeSecret1, departmentScopeSecret2}}
					applySecrets(randomScopes, randomScopes.SecretsDepartmentUrl(""), secrets)
					secretList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}
					sortListsForEqualityCheck(secretList, expectedResponse)
					Expect(secretList).To(Equal(expectedResponse))

					departmentScopeSecret1.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, replicationErrorStr)}
					departmentScopeSecret2.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, lostPhaseStr)}
					secrets = api.Secrets{Secrets: []api.Secret{departmentScopeSecret1, departmentScopeSecret2}}
					applySecrets(randomScopes, randomScopes.SecretsDepartmentUrl(""), secrets)
					secretList = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
					expectedResponse = []api.ClusterObjectStatus{
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
						{
							ScopeType:    "department",
							ScopeId:      randomScopes.DepartmentId,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
					}
					sortListsForEqualityCheck(secretList, expectedResponse)
					Expect(secretList).To(Equal(expectedResponse))
				})
			})
		})
		Context("Project scope", func() {
			Context("create with issues", func() {
				It("should get the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai-"+randomScopes.ProjectIdStr, lostPhaseStr)}
					projectScopeSecret := existingPassword
					projectScopeSecret.Issues = issues

					createSecret(randomScopes, randomScopes.SecretsProjectUrl("/aaa"), projectScopeSecret)
					pvcs := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{{
						ScopeType: "project",
						ScopeId:   randomScopes.ProjectIdStr,
						Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
					}}
					Expect(pvcs[0].Issues).To(Equal(expectedIssues))
				})
				It("should update the resource with the issues related to it", func() {
					issues := []api.Issue{NewClusterResourceIssue("runai-"+randomScopes.ProjectIdStr, lostPhaseStr)}
					differentIssues := []api.Issue{NewClusterResourceIssue("runai-"+randomScopes.ProjectIdStr, replicationErrorStr)}
					projectScopeSecret := existingPassword
					projectScopeSecret.Issues = issues

					createSecret(randomScopes, randomScopes.SecretsProjectUrl("/aaa"), projectScopeSecret)
					secrets := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))

					expectedIssues := []api.ClusterObjectIssue{{
						ScopeType: "project",
						ScopeId:   randomScopes.ProjectIdStr,
						Issue:     api.CLUSTEROBJECTISSUETYPE_LOST_PHASE,
					}}
					Expect(secrets[0].Issues).To(Equal(expectedIssues))

					projectScopeSecret.Issues = differentIssues
					createSecret(randomScopes, randomScopes.SecretsProjectUrl("/aaa"), projectScopeSecret)
					secrets = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))

					expectedIssues = []api.ClusterObjectIssue{{
						ScopeType: "project",
						ScopeId:   randomScopes.ProjectIdStr,
						Issue:     api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
					}}
					Expect(secrets[0].Issues).To(Equal(expectedIssues))
				})
				It("should have only the issues related to each resource", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					projectScopeSecret1 := api.Secret{
						Name:   "aaa",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					projectScopeSecret2 := api.Secret{
						Name:   "bbb",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					secrets := api.Secrets{Secrets: []api.Secret{projectScopeSecret1, projectScopeSecret2}}
					applySecrets(randomScopes, randomScopes.SecretsProjectUrl(""), secrets)
					secretsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}
					sortListsForEqualityCheck(secretsList, expectedResponse)
					Expect(secretsList).To(Equal(expectedResponse))
				})
				It("should update the issues according to the resources", func() {
					assetsOrigin := api.ORIGIN_ASSETS
					projectScopeSecret1 := api.Secret{
						Name:   "aaa",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)},
						Origin: &assetsOrigin,
					}
					projectScopeSecret2 := api.Secret{
						Name:   "bbb",
						Type:   "password",
						Issues: []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, replicationErrorStr)},
						Origin: &assetsOrigin,
					}

					secrets := api.Secrets{Secrets: []api.Secret{projectScopeSecret1, projectScopeSecret2}}
					applySecrets(randomScopes, randomScopes.SecretsProjectUrl(""), secrets)
					secretsList := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
					expectedResponse := []api.ClusterObjectStatus{
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
					}

					sortListsForEqualityCheck(secretsList, expectedResponse)
					Expect(secretsList).To(Equal(expectedResponse))

					projectScopeSecret1.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, replicationErrorStr)}
					projectScopeSecret2.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj2.ProjectIdStr, lostPhaseStr)}
					secrets = api.Secrets{Secrets: []api.Secret{projectScopeSecret1, projectScopeSecret2}}
					applySecrets(randomScopes, randomScopes.SecretsProjectUrl(""), secrets)
					secretsList = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
					expectedResponse = []api.ClusterObjectStatus{
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "aaa",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj1.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR}},
						},
						{
							ScopeType:    "project",
							ScopeId:      randomScopes.ProjectIdStr,
							ResourceType: api.CLUSTERRESOURCETYPE_SECRET,
							SecretType:   api.SECRETTYPE_PASSWORD,
							ResourceName: "bbb",
							Issues:       []api.ClusterObjectIssue{{ScopeType: "project", ScopeId: dept1proj2.ProjectIdStr, Issue: api.CLUSTEROBJECTISSUETYPE_LOST_PHASE}},
						},
					}

					sortListsForEqualityCheck(secretsList, expectedResponse)
					Expect(secretsList).To(Equal(expectedResponse))
				})
			})
		})
	})
	Context("List api filters", func() {
		Context("resourceType filter", func() {
			It("should return only the resources with the specified resource type", func() {
				clusterScopePVC := existingPvc
				clusterScopePVC.Issues = []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)}
				createPvc(randomScopes, randomScopes.PvcsClusterUrl("/aaa"), clusterScopePVC)

				clusterScopeSecret := existingPassword
				clusterScopeSecret.Issues = []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)}
				createSecret(randomScopes, randomScopes.SecretsClusterUrl("/aaa"), clusterScopeSecret)

				clusterScopeConfigMap := api.ConfigMap{
					Name:   "aaa",
					Issues: []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)},
				}
				createConfigMaps(randomScopes, randomScopes.ConfigMapClusterUrl(""), []api.ConfigMap{clusterScopeConfigMap})
				//	list resources with resourceType "pvc"
				resources := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", "", "", ""))
				Expect(len(resources)).To(Equal(1))
				Expect(resources[0].ResourceType).To(Equal(api.CLUSTERRESOURCETYPE_PVC))

				//	list resources with resourceType "secret"
				resources = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
				Expect(len(resources)).To(Equal(1))
				Expect(resources[0].ResourceType).To(Equal(api.CLUSTERRESOURCETYPE_SECRET))

				//	list resources with resourceType "config-map"
				resources = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("config-map", "", "", ""))
				Expect(len(resources)).To(Equal(1))
				Expect(resources[0].ResourceType).To(Equal(api.CLUSTERRESOURCETYPE_CONFIG_MAP))
			})
			It("should return all the resources if the resourceType is empty", func() {
				clusterScopePVC := existingPvc
				clusterScopePVC.Issues = []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)}
				createPvc(randomScopes, randomScopes.PvcsClusterUrl("/aaa"), clusterScopePVC)

				clusterScopeSecret := existingPassword
				clusterScopeSecret.Issues = []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)}
				createSecret(randomScopes, randomScopes.SecretsClusterUrl("/aaa"), clusterScopeSecret)

				clusterScopeConfigMap := api.ConfigMap{
					Name:   "aaa",
					Issues: []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)},
				}
				createConfigMaps(randomScopes, randomScopes.ConfigMapClusterUrl(""), []api.ConfigMap{clusterScopeConfigMap})
				//	list resources with resourceType ""
				resources := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("", "", "", ""))
				Expect(len(resources)).To(Equal(3))
			})
		})
		Context("secretType filter", func() {
			It("should return only the resources with the specified secret type", func() {
				passwordSecret := api.Secret{
					Name:   "aaa",
					Type:   "password",
					Issues: []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)},
				}
				accessKeySecret := api.Secret{
					Name:   "bbb",
					Type:   "access-key",
					Issues: []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)},
				}
				dockerRegistrySecret := api.Secret{
					Name:   "ccc",
					Type:   "docker-registry",
					Issues: []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)},
				}
				applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), api.Secrets{Secrets: []api.Secret{passwordSecret, accessKeySecret, dockerRegistrySecret}})

				resources := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", "password"))
				Expect(len(resources)).To(Equal(1))
				Expect(resources[0].SecretType).To(Equal(api.SECRETTYPE_PASSWORD))

				resources = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", "access-key"))
				Expect(len(resources)).To(Equal(1))
				Expect(resources[0].SecretType).To(Equal(api.SECRETTYPE_ACCESS_KEY))

				resources = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", "docker-registry"))
				Expect(len(resources)).To(Equal(1))
				Expect(resources[0].SecretType).To(Equal(api.SECRETTYPE_DOCKER_REGISTRY))
			})
			It("should return all the resources if the secretType is empty", func() {
				passwordSecret := api.Secret{
					Name:   "aaa",
					Type:   "password",
					Issues: []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)},
				}
				accessKeySecret := api.Secret{
					Name:   "bbb",
					Type:   "access-key",
					Issues: []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)},
				}
				dockerRegistrySecret := api.Secret{
					Name:   "ccc",
					Type:   "docker-registry",
					Issues: []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)},
				}
				applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), api.Secrets{Secrets: []api.Secret{passwordSecret, accessKeySecret, dockerRegistrySecret}})
				resources := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("secret", "", "", ""))
				Expect(len(resources)).To(Equal(3))
			})
		})
		Context("scopeType+scopeId filter", func() {
			It("should return only the resources with the specified scopeType and scopeId", func() {
				clusterScopePVC := existingPvc
				clusterScopePVC.Issues = []api.Issue{NewClusterResourceIssue("runai", lostPhaseStr)}
				createPvc(randomScopes, randomScopes.PvcsClusterUrl("/aaa"), clusterScopePVC)

				departmentScopePVC := existingPvc
				departmentScopePVC.Issues = []api.Issue{NewClusterResourceIssue("runai-"+dept1proj1.ProjectIdStr, lostPhaseStr)}
				createPvc(randomScopes, randomScopes.PvcsDepartmentUrl("/aaa"), departmentScopePVC)

				projectScopePVC := existingPvc
				projectScopePVC.Issues = []api.Issue{NewClusterResourceIssue("runai-"+randomScopes.ProjectIdStr, lostPhaseStr)}
				createPvc(randomScopes, randomScopes.PvcsProjectUrl("/aaa"), projectScopePVC)

				//	list resources with scopeType "cluster" and scopeId "clusterId"
				resources := listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", randomScopes.ClusterId, "cluster", ""))
				Expect(len(resources)).To(Equal(1))

				//	list resources with scopeType "department" and scopeId "departmentId"
				resources = listClusterObjectsStatus(randomScopes, randomScopes.ClusterObjectsStatusUrl("pvc", randomScopes.DepartmentId, "department", ""))
				Expect(len(resources)).To(Equal(2))

				sort.Slice(resources, func(i, j int) bool {
					return strings.Compare(resources[i].ScopeType, resources[j].ScopeType) < 0
				})
				Expect(resources[0].ScopeType).To(Equal("cluster"))
				Expect(resources[1].ScopeType).To(Equal("department"))
			})
		})
	})

	Context("invalid scope type", func() {
		It("should reject SQL injection payloads", func() {
			listHttpReq, err := http.NewRequest("GET", "/internal/api/v1/cluster-objects/status?scopeType=z%27%29%20OR%20TRUE%20--&scopeId=x", nil)
			listHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(listHttpResp.Body.String()).To(ContainSubstring("invalid scopeType"))
		})
	})
})

func listClusterObjectsStatus(randomScopes *RandomScopes, url string) []api.ClusterObjectStatus {
	listHttpReq, err := http.NewRequest("GET", url, nil)
	listHttpReq.Header = *headerForAdmin(randomScopes)

	Expect(err).To(BeNil(), url)

	listHttpResp := httptest.NewRecorder()
	router.ServeHTTP(listHttpResp, listHttpReq)

	Expect(listHttpResp.Code).To(Equal(http.StatusOK), url)

	var result api.ClusterObjectsStatus
	err = json.Unmarshal(listHttpResp.Body.Bytes(), &result)

	Expect(err).To(BeNil(), url)

	return result.Entries
}

func NewClusterResourceIssue(namespace string, issue string) api.Issue {
	return api.Issue{
		Namespace: namespace,
		Issue:     issue,
	}
}

func sortListsForEqualityCheck(list1 []api.ClusterObjectStatus, list2 []api.ClusterObjectStatus) {
	sort.Slice(list1, func(i, j int) bool {
		return strings.Compare(list1[i].ResourceName, list1[j].ResourceName) < 0
	})
	sort.Slice(list2, func(i, j int) bool {
		return strings.Compare(list2[i].ResourceName, list2[j].ResourceName) < 0
	})
}
