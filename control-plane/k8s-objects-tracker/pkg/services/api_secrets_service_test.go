package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"

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

var existingPassword = api.SecretCreationRequest{
	Type:   api.SECRETTYPE_PASSWORD,
	Origin: &originExisting,
}

func compareSecrets(a, b api.SecretListInfo) bool {
	aAsJson, _ := json.Marshal(a)
	bAsJson, _ := json.Marshal(b)
	return strings.Compare(string(aAsJson), string(bAsJson)) < 0
}

func NewSecretReq(secretType api.SecretType, origin *api.Origin) api.SecretCreationRequest {
	return api.SecretCreationRequest{Type: secretType, Origin: origin}
}

var _ = Describe("secrets test", func() {
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

	Context("permission issue when creating secret", func() {
		It("should be rejected", func() {
			createHttpReq, err := http.NewRequest("POST", randomScopes.SecretsClusterUrl("/aaa"), makePayload(existingPassword))
			createHttpReq.Header = *headerWithMissingInfo()
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("apply in multiple clusters", func() {
		It("each cluster shows its own secrets", func() {
			secretsCluster1 := api.Secrets{
				Secrets: []api.Secret{
					{Name: "aaa", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting},
					{Name: "ccc", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting},
					{Name: "eee", Type: api.SECRETTYPE_DOCKER_REGISTRY, Origin: &originExisting},
				},
			}
			secretsCluster2 := api.Secrets{
				Secrets: []api.Secret{
					{Name: "bbb", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting},
					{Name: "ccc", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting},
				},
			}

			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), secretsCluster1)
			applySecrets(anotherCluster, anotherCluster.SecretsClusterUrl(""), secretsCluster2)

			expectedCluster1 := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretsCluster1.Secrets...)
			actualCluster1 := listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster1).To(Equal(expectedCluster1), randomScopes.SecretsClusterUrl("cluster1"))

			expectedCluster2 := anotherCluster.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretsCluster2.Secrets...)
			actualCluster2 := listSecrets(anotherCluster, anotherCluster.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")
		})
	})

	Context("create in multiple clusters", func() {
		It("each cluster shows its own secret", func() {
			secretCluster1 := api.Secret{Name: "aaa", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting}
			secretCluster2 := api.Secret{Name: "bbb", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting}

			createSecret(randomScopes, randomScopes.SecretsClusterUrl("/aaa"), NewSecretReq(secretCluster1.Type, secretCluster1.Origin))
			createSecret(anotherCluster, anotherCluster.SecretsClusterUrl("/bbb"), NewSecretReq(secretCluster2.Type, secretCluster2.Origin))

			expectedCluster1 := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretCluster1)
			actualCluster1 := listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			expectedCluster2 := anotherCluster.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretCluster2)
			actualCluster2 := listSecrets(anotherCluster, anotherCluster.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")
		})
	})

	Context("delete in multiple clusters", func() {
		It("each cluster has its own secrets", func() {
			aaa := newRandomSecret("aaa")
			bbb := newRandomSecret("bbb")

			bothAaaAndBbb := api.Secrets{
				Secrets: []api.Secret{aaa, bbb},
			}

			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), bothAaaAndBbb)
			applySecrets(anotherCluster, anotherCluster.SecretsClusterUrl(""), bothAaaAndBbb)

			expectedCluster1 := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, bothAaaAndBbb.Secrets...)
			actualCluster1 := listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			expectedCluster2 := anotherCluster.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, bothAaaAndBbb.Secrets...)
			actualCluster2 := listSecrets(anotherCluster, anotherCluster.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")

			//
			//   deletion of aaa from cluster1
			//
			deleteSecret(randomScopes, randomScopes.SecretsClusterUrl("/aaa"))

			expectedCluster1 = randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, bbb)
			actualCluster1 = listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			actualCluster2 = listSecrets(anotherCluster, anotherCluster.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")

			//
			//   deletion of bbb from cluster2
			//
			deleteSecret(anotherCluster, anotherCluster.SecretsClusterUrl("/bbb"))

			actualCluster1 = listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster1).To(Equal(expectedCluster1), "cluster1")

			expectedCluster2 = anotherCluster.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, aaa)
			actualCluster2 = listSecrets(anotherCluster, anotherCluster.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster2).To(Equal(expectedCluster2), "cluster2")
		})
	})

	Context("create in multiple scopes", func() {
		It("each scope shows its own secrets", func() {
			aaa := newRandomSecret("aaa")
			bbb := newRandomSecret("bbb")
			ccc := newRandomSecret("ccc")
			ddd := newRandomSecret("ddd")
			eee := newRandomSecret("eee")
			fff := newRandomSecret("fff")
			ggg := newRandomSecret("ggg")
			hhh := newRandomSecret("hhh")

			secretsCluster := api.Secrets{
				Secrets: []api.Secret{eee, fff},
			}
			secretsOfDept := []api.Secrets{
				{Secrets: []api.Secret{aaa, ccc, eee}},
				{Secrets: []api.Secret{bbb, ccc}},
			}
			secretsOfProject := []api.Secrets{
				{Secrets: []api.Secret{aaa, ddd, ggg, hhh}},
				{Secrets: []api.Secret{}},
				{Secrets: []api.Secret{ddd}},
				{Secrets: []api.Secret{eee, hhh}},
			}

			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), secretsCluster)
			applySecrets(dept1, dept1.SecretsDepartmentUrl(""), secretsOfDept[0])
			applySecrets(dept2, dept2.SecretsDepartmentUrl(""), secretsOfDept[1])
			applySecrets(dept1proj1, dept1proj1.SecretsProjectUrl(""), secretsOfProject[0])
			applySecrets(dept1proj2, dept1proj2.SecretsProjectUrl(""), secretsOfProject[1])
			applySecrets(dept2proj1, dept2proj1.SecretsProjectUrl(""), secretsOfProject[2])
			applySecrets(dept2proj2, dept2proj2.SecretsProjectUrl(""), secretsOfProject[3])

			expectedCluster := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretsCluster.Secrets...)
			actualCluster := listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster).To(Equal(expectedCluster), "cluster")

			expectedCluster = randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretsCluster.Secrets...)
			actualCluster = listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins)+includeAncestors)
			Expect(actualCluster).To(Equal(expectedCluster), "cluster-ancestors")

			for idx, curProj := range []*RandomScopes{dept1, dept2} {
				expected := curProj.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, secretsOfDept[idx].Secrets...)
				actual := listSecrets(curProj, curProj.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.DepartmentId)

				expectedCluster = randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretsCluster.Secrets...)
				expected.Secrets = append(expected.Secrets, expectedCluster.Secrets...)
				sort.Slice(expected.Secrets, func(i, j int) bool {
					return compareSecrets(expected.Secrets[i], expected.Secrets[j])
				})

				actual = listSecrets(curProj, curProj.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins)+includeAncestors)
				sort.Slice(actual.Secrets, func(i, j int) bool {
					return compareSecrets(actual.Secrets[i], actual.Secrets[j])
				})

				Expect(actual).To(Equal(expected), curProj.DepartmentId)
			}

			for idx, curProj := range []*RandomScopes{dept1proj1, dept1proj2, dept2proj1, dept2proj2} {
				expected := curProj.SecretsListResponse(orgUnitApi.SCOPETYPE_PROJECT, secretsOfProject[idx].Secrets...)
				actual := listSecrets(curProj, curProj.SecretsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins))
				Expect(actual).To(Equal(expected), curProj.ProjectIdStr)
			}
		})
	})

	Context("create in multiple origins", func() {
		It("filtering flags should work well", func() {
			eee := newRandomSecret("eee")
			fff := newRandomSecret("fff")
			ggg := newRandomSecret("ggg")
			hhh := newRandomSecret("hhh")

			eee.Origin = &originExisting
			fff.Origin = &originAssets
			ggg.Origin = &originAssets
			hhh.Origin = &originExisting

			secretsExisting := api.Secrets{
				Secrets: []api.Secret{eee, hhh},
			}
			secretsAssets := api.Secrets{
				Secrets: []api.Secret{fff, ggg},
			}
			secretsBoth := api.Secrets{
				Secrets: []api.Secret{eee, fff, ggg, hhh},
			}

			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), secretsBoth)

			// by default - only existing are displayed, this is needed for backward compatibility
			expectedCluster := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretsExisting.Secrets...)
			actualCluster := listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", ""))
			Expect(actualCluster).To(Equal(expectedCluster), "default existing")

			// existing only
			expectedCluster = randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretsExisting.Secrets...)
			actualCluster = listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", string(api.ORIGIN_EXISTING)))
			Expect(actualCluster).To(Equal(expectedCluster), "existing")

			// assets only
			expectedCluster = randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretsAssets.Secrets...)
			actualCluster = listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", string(api.ORIGIN_ASSETS)))
			Expect(actualCluster).To(Equal(expectedCluster), "assets")

			// all only
			expectedCluster = randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretsBoth.Secrets...)
			actualCluster = listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actualCluster).To(Equal(expectedCluster), "all")
		})
	})

	Context("create aaa + create bbb + sync aaa ccc ", func() {
		It("result is aaa and ccc (bbb is removed)", func() {
			aaa := newRandomSecret("aaa")
			bbb := newRandomSecret("bbb")
			ccc := newRandomSecret("ccc")

			createSecret(dept1, dept1.SecretsDepartmentUrl("/aaa"), NewSecretReq(aaa.Type, aaa.Origin))
			createSecret(dept1, dept1.SecretsDepartmentUrl("/bbb"), NewSecretReq(bbb.Type, bbb.Origin))

			createdSecrets := api.Secrets{
				Secrets: []api.Secret{aaa, bbb},
			}

			expectedAfterCreate := dept1.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, createdSecrets.Secrets...)
			actualAfterCreate := listSecrets(dept1, dept1.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterCreate).To(Equal(expectedAfterCreate))

			appliedSecrets := api.Secrets{
				Secrets: []api.Secret{aaa, ccc},
			}
			applySecrets(dept1, dept1.SecretsDepartmentUrl(""), appliedSecrets)

			expectedAfterApply := dept1.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, appliedSecrets.Secrets...)
			actualAfterSync := listSecrets(dept1, dept1.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterSync).To(Equal(expectedAfterApply))
		})
	})

	Context("create aaa ccc + delete aaa", func() {
		It("result is bbb", func() {
			aaa := newRandomSecret("aaa")
			ccc := newRandomSecret("ccc")

			appliedSecrets := api.Secrets{
				Secrets: []api.Secret{aaa, ccc},
			}
			applySecrets(dept1, dept1.SecretsDepartmentUrl(""), appliedSecrets)

			expectedAfterSync := dept1.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, appliedSecrets.Secrets...)
			actualAfterSync := listSecrets(dept1, dept1.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterSync).To(Equal(expectedAfterSync))

			deleteSecret(dept1, dept1.SecretsDepartmentUrl("/aaa"))

			expectedAfterDelete := dept1.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, ccc)
			actualAfterDelete := listSecrets(dept1, dept1.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterDelete).To(Equal(expectedAfterDelete))
		})
	})

	Context("search by name and type", func() {
		It("filtering done correctly", func() {
			aaa := api.Secret{Name: "aaa", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting}
			ccc := api.Secret{Name: "ccc", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}

			appliedSecrets := api.Secrets{
				Secrets: []api.Secret{aaa, ccc},
			}
			applySecrets(dept1, dept1.SecretsDepartmentUrl(""), appliedSecrets)

			expectedAfterSync := dept1.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, appliedSecrets.Secrets...)
			actualAfterSync := listSecrets(dept1, dept1.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterSync).To(Equal(expectedAfterSync))

			//
			//   filter by name
			//
			expectedWithNameFilter := dept1.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, ccc)
			actualWithNameFilter := listSecrets(dept1, dept1.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "ccc", dao.AllOrigins))
			Expect(actualWithNameFilter).To(Equal(expectedWithNameFilter))

			//
			//   filter by name and type
			//
			expectedWithNameAndTypeFilter := dept1.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, ccc)
			actualWithNameAndTypeFilter := listSecrets(dept1, dept1.SecretsListUrlWithType(orgUnitApi.SCOPETYPE_DEPARTMENT, "ccc", ccc.Type))
			Expect(actualWithNameAndTypeFilter).To(Equal(expectedWithNameAndTypeFilter))

			//
			//   mismatch filter
			//
			expectedWithMismatchFilter := dept1.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT)
			actualWithMismatchFilter := listSecrets(dept1, dept1.SecretsListUrlWithType(orgUnitApi.SCOPETYPE_DEPARTMENT, "ccc", aaa.Type))
			Expect(actualWithMismatchFilter).To(Equal(expectedWithMismatchFilter))
		})
	})

	Context("re-create VS type change", func() {
		It("should upsert the secret", func() {
			aaa := api.Secret{Name: "aaa", Type: api.SECRETTYPE_PASSWORD, Origin: &originExisting}

			createSecret(dept1, dept1.SecretsDepartmentUrl("/aaa"), NewSecretReq(aaa.Type, aaa.Origin))

			expectedAfterCreate := dept1.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, aaa)
			actualAfterCreate := listSecrets(dept1, dept1.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterCreate).To(Equal(expectedAfterCreate))

			//
			//   re-create same secret, no change
			//
			createSecret(dept1, dept1.SecretsDepartmentUrl("/aaa"), NewSecretReq(aaa.Type, aaa.Origin))
			actualAfterReCreate := listSecrets(dept1, dept1.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterReCreate).To(Equal(expectedAfterCreate))

			//
			//   change the type of secret
			//
			aaaNewType := api.Secret{Name: "aaa", Type: api.SECRETTYPE_ACCESS_KEY, Origin: &originExisting}

			createSecret(dept1, dept1.SecretsDepartmentUrl("/aaa"), NewSecretReq(aaaNewType.Type, aaaNewType.Origin))
			expectedAfterNewType := dept1.SecretsListResponse(orgUnitApi.SCOPETYPE_DEPARTMENT, aaaNewType)
			actualAfterNewType := listSecrets(dept1, dept1.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins))
			Expect(actualAfterNewType).To(Equal(expectedAfterNewType))

			//
			//   sync changes back to original type
			//
			ccc := api.Secret{Name: "aaa", Type: api.SECRETTYPE_DOCKER_REGISTRY, Origin: &originExisting}
			applySecrets(dept1, dept1.SecretsClusterUrl(""), api.Secrets{Secrets: []api.Secret{aaa, ccc}})

			expectedAfterSync := expectedAfterNewType
			actualAfterSync := listSecrets(dept1, dept1.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "aaa", dao.AllOrigins))
			Expect(actualAfterSync).To(Equal(expectedAfterSync))
		})
	})

	Context("list + rbac", func() {
		It("auth rules must apply", func() {
			createSecret(anotherCluster, anotherCluster.SecretsDepartmentUrl("/aaa"), existingPassword)
			createSecret(anotherDept, anotherDept.SecretsDepartmentUrl("/aaa"), existingPassword)
			createSecret(anotherProject, anotherProject.SecretsDepartmentUrl("/aaa"), existingPassword)

			//
			//  forbidden cluster
			//
			urlCluster := randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins)
			listClusterReq, err := http.NewRequest("GET", urlCluster, nil)
			listClusterReq.Header = *headerForNonAdmin(randomScopes)
			Expect(err).To(BeNil())

			listClusterResp := httptest.NewRecorder()
			router.ServeHTTP(listClusterResp, listClusterReq)
			Expect(listClusterResp.Code).To(Equal(http.StatusForbidden), urlCluster)

			//
			//  forbidden department
			//
			urlDept := randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_DEPARTMENT, "", dao.AllOrigins)
			listDeptReq, err := http.NewRequest("GET", urlDept, nil)
			listDeptReq.Header = *headerForNonAdmin(randomScopes)
			Expect(err).To(BeNil())

			listDeptResp := httptest.NewRecorder()
			router.ServeHTTP(listDeptResp, listDeptReq)
			Expect(listClusterResp.Code).To(Equal(http.StatusForbidden), urlDept)

			//
			//  forbidden project
			//
			urlProject := randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_PROJECT, "", dao.AllOrigins)
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

			aaaCreationRequest := api.SecretCreationRequest{
				Type:   api.SECRETTYPE_PASSWORD,
				Origin: &originExisting,
			}

			createHttpReq, err := http.NewRequest("POST", nonExistProject.SecretsProjectUrl("/aaa"), makePayload(aaaCreationRequest))
			createHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			namespace := fmt.Sprintf("runai-%s", nonExistProject.ProjectIdStr)
			Expect(createHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("ProjectId for namespace %s not found", namespace)))

			deleteHttpReq, err := http.NewRequest("DELETE", nonExistProject.SecretsProjectUrl("/aaa"), nil)
			deleteHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("ProjectId for namespace %s not found", namespace)))

			syncPayload := api.Secrets{Secrets: []api.Secret{newRandomSecret("aaa")}}
			syncHttpReq, err := http.NewRequest("POST", nonExistProject.SecretsProjectUrl(""), makePayload(syncPayload))
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
			aaa := newRandomSecret("aaa")
			bbb := newRandomSecret("bbb")

			clusterAaa := api.Secrets{
				Secrets: []api.Secret{aaa},
			}
			clusterBbb := api.Secrets{
				Secrets: []api.Secret{bbb},
			}

			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), clusterAaa)
			applySecrets(anotherCluster, anotherCluster.SecretsClusterUrl(""), clusterBbb)

			expectedClusterAaa := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterAaa.Secrets...)
			actualClusterAaa := listSecrets(randomScopes, randomScopes.SecretsListByClusterId(dao.AllOrigins))
			Expect(actualClusterAaa).To(Equal(expectedClusterAaa), "clusterAaa")

			expectedClusterBbb := anotherCluster.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, clusterBbb.Secrets...)
			actualClusterBbb := listSecrets(anotherCluster, anotherCluster.SecretsListByClusterId(dao.AllOrigins))
			Expect(actualClusterBbb).To(Equal(expectedClusterBbb), "clusterBbb")
		})
	})

	Context("empty secret name", func() {
		It("should be rejected", func() {
			xxx := newRandomSecret("xxx")
			yyy := newRandomSecret("yyy")
			emptyName := newRandomSecret("")

			syncPayload := api.Secrets{Secrets: []api.Secret{xxx, emptyName, yyy}}
			syncHttpReq, err := http.NewRequest("POST", dept1proj1.SecretsProjectUrl(""), makePayload(syncPayload))
			syncHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			syncHttpResp := httptest.NewRecorder()
			router.ServeHTTP(syncHttpResp, syncHttpReq)
			Expect(syncHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(syncHttpResp.Body.String()).To(ContainSubstring("mandatory secret name is not provided"))
		})
	})

	Context("empty secret type", func() {
		It("should be rejected", func() {
			xxx := newRandomSecret("xxx")
			yyy := newRandomSecret("yyy")
			emptyType := api.Secret{Name: "zzz"}

			syncPayload := api.Secrets{Secrets: []api.Secret{xxx, emptyType, yyy}}
			syncHttpReq, err := http.NewRequest("POST", dept1proj1.SecretsProjectUrl(""), makePayload(syncPayload))
			syncHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			syncHttpResp := httptest.NewRecorder()
			router.ServeHTTP(syncHttpResp, syncHttpReq)
			Expect(syncHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(syncHttpResp.Body.String()).To(ContainSubstring("required field 'type' is zero value"))
		})
	})

	Context("missing both scopeType and clusterId", func() {
		It("should be rejected", func() {
			listHttpReq, err := http.NewRequest("GET", "/api/v1/secrets", nil)
			listHttpReq.Header = *headerForAdmin(randomScopes)
			Expect(err).To(BeNil())
			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(listHttpResp.Body.String()).To(ContainSubstring("either scopeType or clusterId must be provided"))
		})

		Context("empty scope id when scopeType is provided", func() {
			It("should be rejected", func() {
				listHttpReq, err := http.NewRequest("GET", "/api/v1/secrets?scopeType=cluster", nil)
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
				listHttpReq, err := http.NewRequest("GET", "/api/v1/secrets?scopeType=z%27%29%20OR%20TRUE%20--&scopeId=x", nil)
				listHttpReq.Header = *headerForAdmin(randomScopes)
				Expect(err).To(BeNil())
				listHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listHttpResp, listHttpReq)
				Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(listHttpResp.Body.String()).To(ContainSubstring("invalid scopeType"))
			})
		})
	})

	Context("includePrivate flag", func() {
		It("should only list non-private secrets by default", func() {
			secretPrivate := api.Secret{Name: "private", Type: api.SECRETTYPE_GENERIC, Origin: &originExisting, Private: convutils.Ref(true)}
			secretPublic := api.Secret{Name: "public", Type: api.SECRETTYPE_GENERIC, Origin: &originExisting, Private: convutils.Ref(false)}

			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), api.Secrets{Secrets: []api.Secret{secretPrivate, secretPublic}})

			// By default, only non-private is listed
			expected := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretPublic)
			actual := listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actual).To(Equal(expected))
		})

		It("should list both private and non-private secrets when includePrivate=true", func() {
			secretPrivate := api.Secret{Name: "private", Type: api.SECRETTYPE_GENERIC, Origin: &originExisting, Private: convutils.Ref(true)}
			secretPublic := api.Secret{Name: "public", Type: api.SECRETTYPE_GENERIC, Origin: &originExisting, Private: convutils.Ref(false)}

			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), api.Secrets{Secrets: []api.Secret{secretPrivate, secretPublic}})

			url := randomScopes.SecretsListUrlWithPrivate(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins)
			expected := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretPublic, secretPrivate)
			actual := listSecrets(randomScopes, url)
			// Sort for comparison
			sort.Slice(actual.Secrets, func(i, j int) bool { return compareSecrets(actual.Secrets[i], actual.Secrets[j]) })
			sort.Slice(expected.Secrets, func(i, j int) bool { return compareSecrets(expected.Secrets[i], expected.Secrets[j]) })
			Expect(actual).To(Equal(expected))
		})

		It("should return no secrets by default if only private secrets exist", func() {
			secretPrivate := api.Secret{Name: "private", Type: api.SECRETTYPE_GENERIC, Origin: &originExisting, Private: convutils.Ref(true)}

			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), api.Secrets{Secrets: []api.Secret{secretPrivate}})

			expected := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER)
			actual := listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actual).To(Equal(expected))
		})

		It("should return private secrets if includePrivate=true and only private secrets exist", func() {
			secretPrivate := api.Secret{Name: "private", Type: api.SECRETTYPE_GENERIC, Origin: &originExisting, Private: convutils.Ref(true)}

			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), api.Secrets{Secrets: []api.Secret{secretPrivate}})

			url := randomScopes.SecretsListUrlWithPrivate(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins)
			expected := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretPrivate)
			actual := listSecrets(randomScopes, url)
			Expect(actual).To(Equal(expected))
		})

		It("should list secrets when Private field is not set", func() {
			secretNoPrivate := api.Secret{Name: "no-private", Type: api.SECRETTYPE_GENERIC, Origin: &originExisting}

			applySecrets(randomScopes, randomScopes.SecretsClusterUrl(""), api.Secrets{Secrets: []api.Secret{secretNoPrivate}})

			// Should be listed by default since Private is not set
			expected := randomScopes.SecretsListResponse(orgUnitApi.SCOPETYPE_CLUSTER, secretNoPrivate)
			actual := listSecrets(randomScopes, randomScopes.SecretsListUrl(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins))
			Expect(actual).To(Equal(expected))

			// Should also be listed when includePrivate=true
			url := randomScopes.SecretsListUrlWithPrivate(orgUnitApi.SCOPETYPE_CLUSTER, "", dao.AllOrigins)
			actual = listSecrets(randomScopes, url)
			Expect(actual).To(Equal(expected))
		})
	})
})

////////////////////////////////////////////////////////////////////////////////////////////////////////////

func createSecret(randomScopes *RandomScopes, url string, request api.SecretCreationRequest) {
	createHttpReq, err := http.NewRequest("POST", url, makePayload(request))
	createHttpReq.Header = *headerForAdmin(randomScopes)
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusOK), url)
}

func deleteSecret(randomScopes *RandomScopes, url string) {
	deleteHttpReq, err := http.NewRequest("DELETE", url, nil)
	deleteHttpReq.Header = *headerForAdmin(randomScopes)

	Expect(err).To(BeNil())

	deleteHttpResp := httptest.NewRecorder()
	router.ServeHTTP(deleteHttpResp, deleteHttpReq)
	Expect(deleteHttpResp.Code).To(Equal(http.StatusNoContent), url)
}

func applySecrets(randomScopes *RandomScopes, url string, payload api.Secrets) {
	syncHttpReq, err := http.NewRequest("POST", url, makePayload(payload))
	syncHttpReq.Header = *headerForAdmin(randomScopes)

	Expect(err).To(BeNil())

	syncHttpResp := httptest.NewRecorder()
	router.ServeHTTP(syncHttpResp, syncHttpReq)
	Expect(syncHttpResp.Code).To(Equal(http.StatusOK), url)
}

func listSecrets(randomScopes *RandomScopes, url string) api.SecretListResponse {
	listHttpReq, err := http.NewRequest("GET", url, nil)
	listHttpReq.Header = *headerForAdmin(randomScopes)
	Expect(err).To(BeNil(), url)

	listHttpResp := httptest.NewRecorder()
	router.ServeHTTP(listHttpResp, listHttpReq)
	Expect(listHttpResp.Code).To(Equal(http.StatusOK), url)

	var result api.SecretListResponse
	err = json.Unmarshal(listHttpResp.Body.Bytes(), &result)
	Expect(err).To(BeNil(), url)

	return result
}

func newRandomSecret(name string) api.Secret {
	types := []api.SecretType{api.SECRETTYPE_ACCESS_KEY, api.SECRETTYPE_PASSWORD, api.SECRETTYPE_DOCKER_REGISTRY}
	origins := []*api.Origin{&originAssets, &originExisting}

	return api.Secret{
		Name:   name,
		Type:   types[rand.Int31n(int32(len(types)))],
		Origin: origins[rand.Int31n(int32(len(origins)))],
	}
}
