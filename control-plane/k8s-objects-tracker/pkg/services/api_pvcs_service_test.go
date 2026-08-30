package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/clients/authorization"
	orgUnit "github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/dbtesting"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

var router chi.Router

const includeAncestors = "&includeAncestors=true"
const includeDescendants = "&includeDescendants=true"

func NewPvcInt(name string, origin *api.Origin) api.Pvc {
	return api.Pvc{
		Name:              name,
		Origin:            origin,
		StorageClassName:  convutils.Ref(fmt.Sprintf("sc-%s", name)),
		Phase:             convutils.Ref(api.PvcPhaseOptional("Bound")),
		ActualCapacity:    convutils.Ref("10Gi"),
		RequestedCapacity: convutils.Ref("12Gi"),
	}
}

func ToPvcCreationRequest(pvc api.Pvc) api.PvcCreationRequest {
	return api.PvcCreationRequest{
		Origin:            pvc.Origin,
		StorageClassName:  convutils.Ref(fmt.Sprintf("sc-%s", pvc.Name)),
		Phase:             convutils.Ref(api.PvcPhaseOptional("Bound")),
		ActualCapacity:    convutils.Ref("10Gi"),
		RequestedCapacity: convutils.Ref("12Gi"),
	}
}

func comparePvcs(a, b api.PvcListInfo) bool {
	aAsJson, _ := json.Marshal(a)
	bAsJson, _ := json.Marshal(b)
	return strings.Compare(string(aAsJson), string(bAsJson)) < 0
}

func NewExistingPvc(name string) api.Pvc {
	return NewPvcInt(name, &originExisting)
}

func NewAssetsPvc(name string) api.Pvc {
	return NewPvcInt(name, &originAssets)
}

var _ = Describe("pvcs test", func() {
	var randomScopes *RandomScopes
	var anotherCluster, anotherDept, anotherProject, dept1, dept2, dept1proj1, dept1proj2, dept2proj1, dept2proj2 *RandomScopes

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

		backendMock := NewRandomsBasedBackendMockClient(randomScopes, dept1, dept2, dept1proj1, dept1proj2, dept2proj1, dept2proj2,
			anotherCluster, anotherDept, anotherProject)
		orgUnit.SetMock(backendMock)
		authorization.SetMock(backendMock)

		router = CreateRouter()

		dbtesting.CleanTables()
	})

	Context("permission issue when creating pvc", func() {
		It("should be rejected", func() {
			createHttpReq, err := http.NewRequest("POST", randomScopes.PvcsClusterUrl("/aaa"), makePayload(existingPvc))
			createHttpReq.Header = *headerWithMissingInfo()
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("apply in multiple clusters", func() {
		It("each cluster shows its own pvcs", func() {
			pvcsCluster1 := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("aaa"), NewExistingPvc("ccc"), NewExistingPvc("eee")},
			}
			pvcsCluster2 := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("bbb"), NewExistingPvc("ccc")},
			}

			applyPvcs(randomScopes, randomScopes.PvcsClusterUrl(""), pvcsCluster1)
			applyPvcs(anotherCluster, anotherCluster.PvcsClusterUrl(""), pvcsCluster2)

			expectedCluster1 := randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsCluster1.Pvcs)
			actualCluster1 := listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster1).To(Equal(expectedCluster1), randomScopes.PvcsClusterUrl("cluster1"))

			expectedCluster2 := anotherCluster.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsCluster2.Pvcs)
			actualCluster2 := listPvcs(anotherCluster, anotherCluster.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")
		})
	})

	Context("create in multiple clusters", func() {
		It("each cluster shows its own pvcs", func() {
			existingAaa := NewExistingPvc("aaa")
			existingBbb := NewExistingPvc("bbb")

			createPvc(randomScopes, randomScopes.PvcsClusterUrl("/aaa"), ToPvcCreationRequest(existingAaa))
			createPvc(anotherCluster, anotherCluster.PvcsClusterUrl("/bbb"), ToPvcCreationRequest(existingBbb))

			expectedCluster1 := randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, []api.Pvc{existingAaa})
			actualCluster1 := listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			expectedCluster2 := anotherCluster.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, []api.Pvc{existingBbb})
			actualCluster2 := listPvcs(anotherCluster, anotherCluster.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")
		})
	})

	Context("delete in multiple clusters", func() {
		It("each cluster has its own pvcs", func() {
			bothAaaAndBbb := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("aaa"), NewExistingPvc("bbb")},
			}

			applyPvcs(randomScopes, randomScopes.PvcsClusterUrl(""), bothAaaAndBbb)
			applyPvcs(anotherCluster, anotherCluster.PvcsClusterUrl(""), bothAaaAndBbb)

			expectedCluster1 := randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, bothAaaAndBbb.Pvcs)
			actualCluster1 := listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			expectedCluster2 := anotherCluster.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, bothAaaAndBbb.Pvcs)
			actualCluster2 := listPvcs(anotherCluster, anotherCluster.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")

			//
			//   deletion of aaa from cluster1
			//
			deletePvc(randomScopes, randomScopes.PvcsClusterUrl("/aaa"))

			expectedCluster1 = randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, []api.Pvc{NewExistingPvc("bbb")})
			actualCluster1 = listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			actualCluster2 = listPvcs(anotherCluster, anotherCluster.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")

			//
			//   deletion of bbb from cluster2
			//
			deletePvc(anotherCluster, anotherCluster.PvcsClusterUrl("/bbb"))

			actualCluster1 = listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			expectedCluster2 = anotherCluster.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, []api.Pvc{NewExistingPvc("aaa")})
			actualCluster2 = listPvcs(anotherCluster, anotherCluster.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")
		})
	})

	Context("create in multiple scopes", func() {
		It("each scope shows its own pvcs", func() {
			pvcsCluster := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("eee"), NewExistingPvc("fff")},
			}
			pvcsOfDept := []api.Pvcs{
				{Pvcs: []api.Pvc{NewExistingPvc("aaa"), NewExistingPvc("ccc"), NewExistingPvc("eee")}},
				{Pvcs: []api.Pvc{NewExistingPvc("bbb"), NewExistingPvc("ccc")}},
			}
			pvcsOfProject := []api.Pvcs{
				{Pvcs: []api.Pvc{NewExistingPvc("aaa"), NewExistingPvc("ddd"), NewExistingPvc("ggg"), NewExistingPvc("hhh")}},
				{Pvcs: []api.Pvc{}},
				{Pvcs: []api.Pvc{NewExistingPvc("hhh"), NewExistingPvc("kkk"), NewExistingPvc("lll"), NewExistingPvc("mmm")}},
				{Pvcs: []api.Pvc{NewExistingPvc("nnn")}},
			}

			applyPvcs(randomScopes, randomScopes.PvcsClusterUrl(""), pvcsCluster)
			applyPvcs(dept1, dept1.PvcsDepartmentUrl(""), pvcsOfDept[0])
			applyPvcs(dept2, dept2.PvcsDepartmentUrl(""), pvcsOfDept[1])
			applyPvcs(dept1proj1, dept1proj1.PvcsProjectUrl(""), pvcsOfProject[0])
			applyPvcs(dept1proj2, dept1proj2.PvcsProjectUrl(""), pvcsOfProject[1])
			applyPvcs(dept2proj1, dept2proj1.PvcsProjectUrl(""), pvcsOfProject[2])
			applyPvcs(dept2proj2, dept2proj2.PvcsProjectUrl(""), pvcsOfProject[3])

			expectedCluster := randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsCluster.Pvcs)
			actualCluster := listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster).To(Equal(expectedCluster), "cluster")

			expectedCluster = randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsCluster.Pvcs)
			actualCluster = listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins)+includeAncestors)
			Expect(actualCluster).To(Equal(expectedCluster), "cluster+ancestors")

			for idx, curProj := range []*RandomScopes{dept1, dept2} {
				expected := curProj.PvcsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, pvcsOfDept[idx].Pvcs)
				actual := listPvcs(curProj, curProj.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.DepartmentId)

				//
				//   try with ancestors, expecting to get the cluster pvcs as well
				//
				expectedCluster = curProj.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsCluster.Pvcs)
				expected.Pvcs = append(expected.Pvcs, expectedCluster.Pvcs...)

				sort.Slice(expected.Pvcs, func(i, j int) bool {
					return comparePvcs(expected.Pvcs[i], expected.Pvcs[j])
				})

				actual = listPvcs(curProj, curProj.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins)+includeAncestors)
				sort.Slice(actual.Pvcs, func(i, j int) bool {
					return comparePvcs(actual.Pvcs[i], actual.Pvcs[j])
				})

				Expect(actual).To(Equal(expected), curProj.DepartmentId)
			}

			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2, dept2proj1, dept2proj2} {
				expected := curProj.PvcsListResponse(orgUnitApi.SCOPETYPE_PROJECT, pvcsOfProject[idx].Pvcs)
				actual := listPvcs(curProj, curProj.PvcsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
		})

		It("cluster with includeDescendants should return all pvcs in the cluster", func() {
			pvcsCluster := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("cluster-pvc-1")},
			}
			pvcsDept1 := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("dept1-pvc-1"), NewExistingPvc("dept1-pvc-2")},
			}
			pvcsDept1Proj1 := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("dept1proj1-pvc-1"), NewExistingPvc("dept1proj1-pvc-2")},
			}
			pvcsDept1Proj2 := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("dept1proj2-pvc-1")},
			}

			applyPvcs(randomScopes, randomScopes.PvcsClusterUrl(""), pvcsCluster)
			applyPvcs(dept1, dept1.PvcsDepartmentUrl(""), pvcsDept1)
			applyPvcs(dept1proj1, dept1proj1.PvcsProjectUrl(""), pvcsDept1Proj1)
			applyPvcs(dept1proj2, dept1proj2.PvcsProjectUrl(""), pvcsDept1Proj2)

			// Query cluster without includeDescendants - should only get cluster-scoped PVCs
			expectedCluster := randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsCluster.Pvcs)
			actualCluster := listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster).To(Equal(expectedCluster), "cluster without descendants")

			// Query cluster with includeDescendants - should get ALL PVCs in the cluster
			expectedAllPvcs := randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsCluster.Pvcs)
			expectedDept := dept1.PvcsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, pvcsDept1.Pvcs)
			expectedProj1 := dept1proj1.PvcsListResponse(orgUnitApi.SCOPETYPE_PROJECT, pvcsDept1Proj1.Pvcs)
			expectedProj2 := dept1proj2.PvcsListResponse(orgUnitApi.SCOPETYPE_PROJECT, pvcsDept1Proj2.Pvcs)
			expectedAllPvcs.Pvcs = append(expectedAllPvcs.Pvcs, expectedDept.Pvcs...)
			expectedAllPvcs.Pvcs = append(expectedAllPvcs.Pvcs, expectedProj1.Pvcs...)
			expectedAllPvcs.Pvcs = append(expectedAllPvcs.Pvcs, expectedProj2.Pvcs...)

			sort.Slice(expectedAllPvcs.Pvcs, func(i, j int) bool {
				return comparePvcs(expectedAllPvcs.Pvcs[i], expectedAllPvcs.Pvcs[j])
			})

			actualAllPvcs := listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins)+includeDescendants)
			sort.Slice(actualAllPvcs.Pvcs, func(i, j int) bool {
				return comparePvcs(actualAllPvcs.Pvcs[i], actualAllPvcs.Pvcs[j])
			})

			Expect(actualAllPvcs).To(Equal(expectedAllPvcs), "cluster with descendants (all cluster PVCs)")

			// Query cluster with BOTH includeAncestors and includeDescendants
			// Should return the same as includeDescendants alone (all cluster PVCs)
			actualBothFlags := listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins)+includeAncestors+includeDescendants)
			sort.Slice(actualBothFlags.Pvcs, func(i, j int) bool {
				return comparePvcs(actualBothFlags.Pvcs[i], actualBothFlags.Pvcs[j])
			})

			Expect(actualBothFlags).To(Equal(expectedAllPvcs), "cluster with both ancestors and descendants")

			// Query department with includeDescendants - should have no effect (only cluster scope supported)
			expectedDept2 := dept1.PvcsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, pvcsDept1.Pvcs)
			actualDept := listPvcs(dept1, dept1.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins)+includeDescendants)
			Expect(actualDept).To(Equal(expectedDept2), "department with descendants (no effect)")

			// Query project with includeDescendants - should have no effect (only cluster scope supported)
			expectedProj := dept1proj1.PvcsListResponse(orgUnitApi.SCOPETYPE_PROJECT, pvcsDept1Proj1.Pvcs)
			actualProj := listPvcs(dept1proj1, dept1proj1.PvcsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins)+includeDescendants)
			Expect(actualProj).To(Equal(expectedProj), "project with descendants (no effect)")
		})
	})

	Context("create in multiple origins", func() {
		It("filtering flags should work well", func() {
			pvcsClusterExisting := api.Pvcs{
				Pvcs: []api.Pvc{
					NewExistingPvc("eee"),
				},
			}
			pvcsClusterAssets := api.Pvcs{
				Pvcs: []api.Pvc{
					NewAssetsPvc("fff"),
				},
			}
			pvcsClusterBoth := api.Pvcs{
				Pvcs: append(pvcsClusterExisting.Pvcs, pvcsClusterAssets.Pvcs...),
			}

			applyPvcs(randomScopes, randomScopes.PvcsClusterUrl(""), pvcsClusterBoth)

			// by default - only existing are displayed, this is needed for backward compatibility
			expectedCluster := randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsClusterExisting.Pvcs)
			actualCluster := listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", ""))
			Expect(actualCluster).To(Equal(expectedCluster), "default existing")

			// existing only
			expectedCluster = randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsClusterExisting.Pvcs)
			actualCluster = listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", string(api.ORIGIN_EXISTING)))
			Expect(actualCluster).To(Equal(expectedCluster), "existing")

			// assets only
			expectedCluster = randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsClusterAssets.Pvcs)
			actualCluster = listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", string(api.ORIGIN_ASSETS)))
			Expect(actualCluster).To(Equal(expectedCluster), "assets")

			// all only
			expectedCluster = randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsClusterBoth.Pvcs)
			actualCluster = listPvcs(randomScopes, randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster).To(Equal(expectedCluster), "all")
		})
	})

	Context("create aaa + create bbb + sync aaa ccc ", func() {
		It("result is aaa and ccc (bbb is removed)", func() {
			existingAaa := NewExistingPvc("aaa")
			existingBbb := NewExistingPvc("bbb")
			existingCcc := NewExistingPvc("ccc")

			createPvc(dept1, dept1.PvcsDepartmentUrl("/aaa"), ToPvcCreationRequest(existingAaa))
			createPvc(dept1, dept1.PvcsDepartmentUrl("/bbb"), ToPvcCreationRequest(existingBbb))

			createdPvcs := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("aaa"), NewExistingPvc("bbb")},
			}

			expectedAfterCreate := dept1.PvcsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, createdPvcs.Pvcs)
			actualAfterCreate := listPvcs(dept1, dept1.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterCreate).To(Equal(expectedAfterCreate))

			appliedPvcs := api.Pvcs{
				Pvcs: []api.Pvc{existingAaa, existingCcc},
			}
			applyPvcs(dept1, dept1.PvcsDepartmentUrl(""), appliedPvcs)

			expectedAfterApply := dept1.PvcsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, appliedPvcs.Pvcs)
			actualAfterSync := listPvcs(dept1, dept1.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterSync).To(Equal(expectedAfterApply))
		})
	})

	Context("create aaa ccc + delete aaa", func() {
		It("result is bbb", func() {
			appliedPvcs := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("aaa"), NewExistingPvc("ccc")},
			}
			applyPvcs(dept1, dept1.PvcsDepartmentUrl(""), appliedPvcs)

			expectedAfterSync := dept1.PvcsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, appliedPvcs.Pvcs)
			actualAfterSync := listPvcs(dept1, dept1.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterSync).To(Equal(expectedAfterSync))

			deletePvc(dept1, dept1.PvcsDepartmentUrl("/aaa"))

			expectedAfterDelete := dept1.PvcsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, []api.Pvc{NewExistingPvc("ccc")})
			actualAfterDelete := listPvcs(dept1, dept1.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterDelete).To(Equal(expectedAfterDelete))
		})
	})

	Context("search by name", func() {
		It("filtering done correctly", func() {
			appliedPvcs := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("aaa"), NewExistingPvc("ccc")},
			}
			applyPvcs(dept1, dept1.PvcsDepartmentUrl(""), appliedPvcs)

			expectedAfterSync := dept1.PvcsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, appliedPvcs.Pvcs)
			actualAfterSync := listPvcs(dept1, dept1.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterSync).To(Equal(expectedAfterSync))

			expectedWithFilter := dept1.PvcsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, []api.Pvc{NewExistingPvc("ccc")})
			actualWithFilter := listPvcs(dept1, dept1.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "ccc", dao.AllOrigins))
			Expect(actualWithFilter).To(Equal(expectedWithFilter))
		})
	})

	Context("re-create same PVC", func() {
		It("should not change", func() {
			existingAaa := NewExistingPvc("aaa")
			existingCcc := NewExistingPvc("ccc")

			appliedPvcs := api.Pvcs{
				Pvcs: []api.Pvc{existingAaa, existingCcc},
			}
			applyPvcs(dept1, dept1.PvcsDepartmentUrl(""), appliedPvcs)

			expectedAfterSync := dept1.PvcsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, appliedPvcs.Pvcs)
			actualAfterSync := listPvcs(dept1, dept1.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterSync).To(Equal(expectedAfterSync))

			//
			//   re-create same PVC, no change
			//
			createPvc(dept1, dept1.PvcsDepartmentUrl("/ccc"), ToPvcCreationRequest(existingCcc))

			actualAfterReCreate := listPvcs(dept1, dept1.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterReCreate).To(Equal(expectedAfterSync))
		})
	})

	Context("list + rbac", func() {
		It("auth rules must apply", func() {
			createPvc(anotherCluster, anotherCluster.PvcsDepartmentUrl("/aaa"), existingPvc)
			createPvc(anotherDept, anotherDept.PvcsDepartmentUrl("/aaa"), existingPvc)
			createPvc(anotherProject, anotherProject.PvcsDepartmentUrl("/aaa"), existingPvc)

			//
			//  forbidden cluster
			//
			urlCluster := randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins)
			listClusterReq, err := http.NewRequest("GET", urlCluster, nil)
			listClusterReq.Header = *headerForNonAdmin(randomScopes)
			Expect(err).To(BeNil())

			listClusterResp := httptest.NewRecorder()
			router.ServeHTTP(listClusterResp, listClusterReq)
			Expect(listClusterResp.Code).To(Equal(http.StatusForbidden), urlCluster)

			//
			//  forbidden department
			//
			urlDept := randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins)
			listDeptReq, err := http.NewRequest("GET", urlDept, nil)
			listDeptReq.Header = *headerForNonAdmin(randomScopes)
			Expect(err).To(BeNil())

			listDeptResp := httptest.NewRecorder()
			router.ServeHTTP(listDeptResp, listDeptReq)
			Expect(listClusterResp.Code).To(Equal(http.StatusForbidden), urlDept)

			//
			//  forbidden project
			//
			urlProject := randomScopes.PvcsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins)
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

			createHttpReq, err := http.NewRequest("POST", nonExistProject.PvcsProjectUrl("/aaa"), makePayload(existingPvc))
			createHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			namespace := fmt.Sprintf("runai-%s", nonExistProject.ProjectIdStr)
			Expect(createHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("ProjectId for namespace %s not found", namespace)))

			deleteHttpReq, err := http.NewRequest("DELETE", nonExistProject.PvcsProjectUrl("/aaa"), nil)
			deleteHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("ProjectId for namespace %s not found", namespace)))

			syncPayload := api.Pvcs{Pvcs: []api.Pvc{NewExistingPvc("aaa")}}
			syncHttpReq, err := http.NewRequest("POST", nonExistProject.PvcsProjectUrl(""), makePayload(syncPayload))
			syncHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			syncHttpResp := httptest.NewRecorder()
			router.ServeHTTP(syncHttpResp, syncHttpReq)
			Expect(syncHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(syncHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("ProjectId for namespace %s not found", namespace)))
		})
	})

	Context("list with cluster id", func() {
		It("correctly filter by cluster", func() {
			pvcsCluster1 := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("aaa"), NewExistingPvc("bbb")},
			}
			pvcsCluster2 := api.Pvcs{
				Pvcs: []api.Pvc{NewExistingPvc("bbb"), NewExistingPvc("ccc")},
			}

			applyPvcs(randomScopes, randomScopes.PvcsClusterUrl(""), pvcsCluster1)
			applyPvcs(anotherCluster, anotherCluster.PvcsClusterUrl(""), pvcsCluster2)

			expectedCluster1 := randomScopes.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsCluster1.Pvcs)
			actualCluster1 := listPvcs(randomScopes, randomScopes.PvcsListByClusterId())
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			expectedCluster2 := anotherCluster.PvcsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, pvcsCluster2.Pvcs)
			actualCluster2 := listPvcs(anotherCluster, anotherCluster.PvcsListByClusterId())
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")
		})
	})

	Context("empty PVC name", func() {
		It("should be rejected", func() {
			syncPayload := api.Pvcs{Pvcs: []api.Pvc{NewExistingPvc("xxx"), NewExistingPvc(""), NewExistingPvc("yyy")}}
			syncHttpReq, err := http.NewRequest("POST", dept1proj1.PvcsProjectUrl(""), makePayload(syncPayload))
			syncHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			syncHttpResp := httptest.NewRecorder()
			router.ServeHTTP(syncHttpResp, syncHttpReq)
			Expect(syncHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(syncHttpResp.Body.String()).To(ContainSubstring("mandatory PVC name is not provided"))
		})
	})

	Context("missing both scopeType and clusterId", func() {
		It("should be rejected", func() {
			listHttpReq, err := http.NewRequest("GET", "/api/v1/pvcs", nil)
			listHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(listHttpResp.Body.String()).To(ContainSubstring("either scopeType or clusterId must be provided"))
		})

		Context("empty scope id when scopeType is provided", func() {
			It("should be rejected", func() {
				listHttpReq, err := http.NewRequest("GET", "/api/v1/pvcs?scopeType=cluster", nil)
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
				listHttpReq, err := http.NewRequest("GET", "/api/v1/pvcs?scopeType=z%27%29%20OR%20TRUE%20--&scopeId=x", nil)
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////

func createPvc(randomScopes *RandomScopes, url string, request api.PvcCreationRequest) {
	createHttpReq, err := http.NewRequest("POST", url, makePayload(request))
	createHttpReq.Header = *headerForAdmin(randomScopes)

	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)

	Expect(createHttpResp.Code).To(Equal(http.StatusOK), url)
}

func deletePvc(randomScopes *RandomScopes, url string) {
	deleteHttpReq, err := http.NewRequest("DELETE", url, nil)
	deleteHttpReq.Header = *headerForAdmin(randomScopes)

	Expect(err).To(BeNil())

	deleteHttpResp := httptest.NewRecorder()
	router.ServeHTTP(deleteHttpResp, deleteHttpReq)

	Expect(deleteHttpResp.Code).To(Equal(http.StatusNoContent), url)
}

func applyPvcs(randomScopes *RandomScopes, url string, payload api.Pvcs) {
	syncHttpReq, err := http.NewRequest("POST", url, makePayload(payload))
	syncHttpReq.Header = *headerForAdmin(randomScopes)

	Expect(err).To(BeNil())

	syncHttpResp := httptest.NewRecorder()
	router.ServeHTTP(syncHttpResp, syncHttpReq)

	Expect(syncHttpResp.Code).To(Equal(http.StatusOK), url)
}

func listPvcs(randomScopes *RandomScopes, url string) api.PvcListResponse {
	listHttpReq, err := http.NewRequest("GET", url, nil)
	listHttpReq.Header = *headerForAdmin(randomScopes)

	Expect(err).To(BeNil(), url)

	listHttpResp := httptest.NewRecorder()
	router.ServeHTTP(listHttpResp, listHttpReq)

	Expect(listHttpResp.Code).To(Equal(http.StatusOK), url)

	var result api.PvcListResponse
	err = json.Unmarshal(listHttpResp.Body.Bytes(), &result)

	Expect(err).To(BeNil(), url)

	return result
}
