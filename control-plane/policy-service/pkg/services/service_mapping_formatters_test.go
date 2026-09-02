/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

var (
	urlFormatters = "/api/v1/policies/formatter"
)

func formatterByNameURL(name string) string {
	return fmt.Sprintf("%s/%s", urlFormatters, name)
}

func formatterFieldsURL(name string) string {
	return fmt.Sprintf("%s/%s/fields", urlFormatters, name)
}

func singleArgStringFormatterField() api.FormatterField {
	setExpr := `{"x":$ARGS.named.k}`
	return api.FormatterField{
		SetExpr: &setExpr,
		ArgsInfo: map[string]api.MappingFieldType{
			"k": api.MAPPINGFIELDTYPE_STRING,
		},
	}
}

func formatterWithContainerPortField(portsSet, portsGet string) map[string]api.FormatterField {
	return map[string]api.FormatterField{
		"ports": {
			SetExpr: &portsSet,
			GetExpr: &portsGet,
			ArrayKey: &api.FormatterArrayKey{
				ArgName: "containerPort",
			},
			ArgsInfo: map[string]api.MappingFieldType{
				"containerPort": api.MAPPINGFIELDTYPE_INTEGER,
				"protocol":      api.MAPPINGFIELDTYPE_STRING,
			},
			ArgsEnum: map[string][]string{
				"protocol": {"TCP", "UDP"},
			},
		},
	}
}

var _ = Describe("FormatterTest", Ordered, func() {
	router := CreateRouter()

	var backendMock *BackendMockClient
	var clusterMock cluster.SimpleMockClient

	var randoms = testutils.NewRandomValues()
	var tenantId = int32(randoms.TenantId)
	var formatterName = "api-test-formatter"

	var portsSet = `
			{
				"ports":[
					{
						"containerPort":$ARGS.named.containerPort,
						"protocol":$ARGS.named.protocol
					}
				]
			}`

	var portsGet = `.spec.containers[0].ports`

	var setLabels = `
			{
				"labels":{
					($ARGS.named.key): $ARGS.named.value
				}
			}`

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()

		backendMock = NewBackendMockClient(2, 2)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))
		authorization.SetMock(NewPoliciesAuthClient(backendMock))

		clusterMock = cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)
	})

	Context("authentication", func() {
		It("rejects create without auth headers", func() {
			body := api.FormatterCreationRequest{
				Name:   formatterName,
				Fields: formatterWithContainerPortField(portsSet, portsGet),
			}

			request, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(body))
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("tenant formatter APIs", func() {
		It("returns conflict when creating the same formatter name twice", func() {
			dupName := "fmt-dup-create-test"
			createBody := api.FormatterCreationRequest{
				Name:        dupName,
				Description: convutils.Ref("v1"),
				Fields:      formatterWithContainerPortField(portsSet, portsGet),
			}

			req1, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
			Expect(err).To(BeNil())
			req1.Header = *headerForAdmin(tenantId)
			resp1 := httptest.NewRecorder()
			router.ServeHTTP(resp1, req1.WithContext(tenantContextForTest(tenantId)))
			Expect(resp1.Code).To(Equal(http.StatusOK), resp1.Body.String())

			req2, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
			Expect(err).To(BeNil())
			req2.Header = *headerForAdmin(tenantId)
			resp2 := httptest.NewRecorder()
			router.ServeHTTP(resp2, req2.WithContext(tenantContextForTest(tenantId)))
			Expect(resp2.Code).To(Equal(http.StatusConflict), resp2.Body.String())
		})

		It("creates, lists, gets, patches, deletes a field, and deletes the formatter", func() {
			createBody := api.FormatterCreationRequest{
				Name:        formatterName,
				Description: convutils.Ref("v1 test formatter"),
				Fields:      formatterWithContainerPortField(portsSet, portsGet),
			}

			// Create
			reqPost, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantId)

			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantId)))
			Expect(respPost.Code).To(Equal(http.StatusOK), respPost.Body.String())

			var created api.Formatter
			Expect(json.Unmarshal(respPost.Body.Bytes(), &created)).To(Succeed())
			Expect(created.Name).To(Equal(formatterName))
			Expect(created.Description).ToNot(BeNil())
			Expect(*created.Description).To(Equal("v1 test formatter"))
			Expect(created.Fields).To(HaveKey("ports"))

			// List
			reqList, err := http.NewRequest(http.MethodGet, urlFormatters+"?offset=0&limit=50", nil)
			Expect(err).To(BeNil())
			reqList.Header = *headerForAdmin(tenantId)

			respList := httptest.NewRecorder()
			router.ServeHTTP(respList, reqList.WithContext(tenantContextForTest(tenantId)))
			Expect(respList.Code).To(Equal(http.StatusOK), respList.Body.String())

			var listResp api.FormatterListResponse
			Expect(json.Unmarshal(respList.Body.Bytes(), &listResp)).To(Succeed())
			found := false
			for _, item := range listResp.Items {
				if item.Name == formatterName {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())

			// Get by name
			reqGet, err := http.NewRequest(http.MethodGet, formatterByNameURL(formatterName), nil)
			Expect(err).To(BeNil())
			reqGet.Header = *headerForAdmin(tenantId)

			respGet := httptest.NewRecorder()
			router.ServeHTTP(respGet, reqGet.WithContext(tenantContextForTest(tenantId)))
			Expect(respGet.Code).To(Equal(http.StatusOK), respGet.Body.String())

			var got api.Formatter
			Expect(json.Unmarshal(respGet.Body.Bytes(), &got)).To(Succeed())
			Expect(got.Name).To(Equal(formatterName))
			Expect(got.Fields).To(HaveKey("ports"))

			// Patch — merge labels field and update description
			patchBody := api.FormatterUpdateRequest{
				Description: convutils.Ref("v2 test formatter"),
				Fields: map[string]api.FormatterField{
					"labels": {
						SetExpr: &setLabels,
						ArgsInfo: map[string]api.MappingFieldType{
							"value": api.MAPPINGFIELDTYPE_QUANTITY,
						},
					},
				},
			}

			reqPatch, err := http.NewRequest(http.MethodPatch, formatterByNameURL(formatterName), makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantId)

			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantId)))
			Expect(respPatch.Code).To(Equal(http.StatusOK), respPatch.Body.String())

			var patched api.Formatter
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &patched)).To(Succeed())
			Expect(patched.Fields).To(HaveKey("ports"))
			Expect(patched.Fields).To(HaveKey("labels"))
			Expect(patched.Description).ToNot(BeNil())
			Expect(*patched.Description).To(Equal("v2 test formatter"))

			// Delete fields (labels only)
			delFieldsBody := api.FormatterDeleteFieldsRequest{Fields: []string{"labels"}}
			reqDelFields, err := http.NewRequest(http.MethodDelete, formatterFieldsURL(formatterName), makePayload(delFieldsBody))
			Expect(err).To(BeNil())
			reqDelFields.Header = *headerForAdmin(tenantId)

			respDelFields := httptest.NewRecorder()
			router.ServeHTTP(respDelFields, reqDelFields.WithContext(tenantContextForTest(tenantId)))
			Expect(respDelFields.Code).To(Equal(http.StatusOK), respDelFields.Body.String())
			var afterDelFields api.Formatter
			Expect(json.Unmarshal(respDelFields.Body.Bytes(), &afterDelFields)).To(Succeed())
			Expect(afterDelFields.Fields).To(HaveKey("ports"))
			Expect(afterDelFields.Fields).ToNot(HaveKey("labels"))

			reqGet2, err := http.NewRequest(http.MethodGet, formatterByNameURL(formatterName), nil)
			Expect(err).To(BeNil())
			reqGet2.Header = *headerForAdmin(tenantId)

			respGet2 := httptest.NewRecorder()
			router.ServeHTTP(respGet2, reqGet2.WithContext(tenantContextForTest(tenantId)))
			Expect(respGet2.Code).To(Equal(http.StatusOK))
			Expect(json.Unmarshal(respGet2.Body.Bytes(), &got)).To(Succeed())
			Expect(got.Fields).To(HaveKey("ports"))
			Expect(got.Fields).ToNot(HaveKey("labels"))

			// Delete formatter
			reqDel, err := http.NewRequest(http.MethodDelete, formatterByNameURL(formatterName), nil)
			Expect(err).To(BeNil())
			reqDel.Header = *headerForAdmin(tenantId)

			respDel := httptest.NewRecorder()
			router.ServeHTTP(respDel, reqDel.WithContext(tenantContextForTest(tenantId)))
			Expect(respDel.Code).To(Equal(http.StatusNoContent), respDel.Body.String())

			reqGet404, err := http.NewRequest(http.MethodGet, formatterByNameURL(formatterName), nil)
			Expect(err).To(BeNil())
			reqGet404.Header = *headerForAdmin(tenantId)

			respGet404 := httptest.NewRecorder()
			router.ServeHTTP(respGet404, reqGet404.WithContext(tenantContextForTest(tenantId)))
			Expect(respGet404.Code).To(Equal(http.StatusNotFound))
			Expect(respGet404.Body.String()).To(ContainSubstring(fmt.Sprintf("formatter named %s not found", formatterName)))
		})

		It("returns 404 when getting an unknown formatter", func() {
			req, err := http.NewRequest(http.MethodGet, formatterByNameURL("unknown-formatter-zz"), nil)
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantId)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
			Expect(resp.Code).To(Equal(http.StatusNotFound))
			Expect(resp.Body.String()).To(ContainSubstring("formatter named unknown-formatter-zz not found"))
		})

		It("returns validation error when deleting a field that does not exist", func() {
			createBody := api.FormatterCreationRequest{
				Name:   formatterName,
				Fields: formatterWithContainerPortField(portsSet, portsGet),
			}

			reqPost, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantId)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantId)))
			Expect(respPost.Code).To(Equal(http.StatusOK))

			delFieldsBody := api.FormatterDeleteFieldsRequest{Fields: []string{"no-such-field"}}
			reqDelFields, err := http.NewRequest(http.MethodDelete, formatterFieldsURL(formatterName), makePayload(delFieldsBody))
			Expect(err).To(BeNil())
			reqDelFields.Header = *headerForAdmin(tenantId)

			respDelFields := httptest.NewRecorder()
			router.ServeHTTP(respDelFields, reqDelFields.WithContext(tenantContextForTest(tenantId)))
			Expect(respDelFields.Code).To(Equal(http.StatusBadRequest), respDelFields.Body.String())
			Expect(respDelFields.Body.String()).To(ContainSubstring("field no-such-field does not exist in formatter " + formatterName))
		})

		Context("field map key validation (MappingFieldName)", func() {
			DescribeTable("rejects POST create with invalid field name key",
				func(badKey string) {
					body := api.FormatterCreationRequest{
						Name:   "fmt-bad-field-key",
						Fields: map[string]api.FormatterField{badKey: singleArgStringFormatterField()},
					}

					reqPost, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(body))
					Expect(err).To(BeNil())
					reqPost.Header = *headerForAdmin(tenantId)

					respPost := httptest.NewRecorder()
					router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantId)))
					Expect(respPost.Code).To(Equal(http.StatusBadRequest), respPost.Body.String())
					var apiErr api.Error
					Expect(json.Unmarshal(respPost.Body.Bytes(), &apiErr)).To(Succeed())
					Expect(apiErr.Message).To(Equal(fmt.Sprintf(
						`field name %q is invalid: must be alphanumeric with optional internal hyphens, and must not start or end with a hyphen`,
						badKey,
					)))
				},
				Entry("bang in key", "x!x"),
				Entry("trailing hyphen", "x-"),
				Entry("leading hyphen", "-x"),
				Entry("internal space", "bad key"),
			)

			It("rejects PATCH merge with invalid field name key", func() {
				createBody := api.FormatterCreationRequest{
					Name:   formatterName,
					Fields: formatterWithContainerPortField(portsSet, portsGet),
				}
				reqPost, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
				Expect(err).To(BeNil())
				reqPost.Header = *headerForAdmin(tenantId)
				respPost := httptest.NewRecorder()
				router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantId)))
				Expect(respPost.Code).To(Equal(http.StatusOK))

				patchBody := api.FormatterUpdateRequest{
					Fields: map[string]api.FormatterField{
						"x!x": singleArgStringFormatterField(),
					},
				}
				reqPatch, err := http.NewRequest(http.MethodPatch, formatterByNameURL(formatterName), makePayload(patchBody))
				Expect(err).To(BeNil())
				reqPatch.Header = *headerForAdmin(tenantId)

				respPatch := httptest.NewRecorder()
				router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantId)))
				Expect(respPatch.Code).To(Equal(http.StatusBadRequest), respPatch.Body.String())
				var apiErr api.Error
				Expect(json.Unmarshal(respPatch.Body.Bytes(), &apiErr)).To(Succeed())
				Expect(apiErr.Message).To(Equal(
					`field name "x!x" is invalid: must be alphanumeric with optional internal hyphens, and must not start or end with a hyphen`,
				))
			})

			It("rejects DELETE fields when key is invalid (before missing-field check)", func() {
				createBody := api.FormatterCreationRequest{
					Name:   formatterName,
					Fields: formatterWithContainerPortField(portsSet, portsGet),
				}
				reqPost, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
				Expect(err).To(BeNil())
				reqPost.Header = *headerForAdmin(tenantId)
				respPost := httptest.NewRecorder()
				router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantId)))
				Expect(respPost.Code).To(Equal(http.StatusOK))

				delFieldsBody := api.FormatterDeleteFieldsRequest{Fields: []string{"x!x"}}
				reqDelFields, err := http.NewRequest(http.MethodDelete, formatterFieldsURL(formatterName), makePayload(delFieldsBody))
				Expect(err).To(BeNil())
				reqDelFields.Header = *headerForAdmin(tenantId)

				respDelFields := httptest.NewRecorder()
				router.ServeHTTP(respDelFields, reqDelFields.WithContext(tenantContextForTest(tenantId)))
				Expect(respDelFields.Code).To(Equal(http.StatusBadRequest), respDelFields.Body.String())
				var apiErr api.Error
				Expect(json.Unmarshal(respDelFields.Body.Bytes(), &apiErr)).To(Succeed())
				Expect(apiErr.Message).To(Equal(
					`field name "x!x" is invalid: must be alphanumeric with optional internal hyphens, and must not start or end with a hyphen`,
				))
			})
		})

		It("returns validation error for patch with empty fields map", func() {
			createBody := api.FormatterCreationRequest{
				Name:   formatterName,
				Fields: formatterWithContainerPortField(portsSet, portsGet),
			}
			reqPost, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
			Expect(err).To(BeNil())
			reqPost.Header = *headerForAdmin(tenantId)
			respPost := httptest.NewRecorder()
			router.ServeHTTP(respPost, reqPost.WithContext(tenantContextForTest(tenantId)))
			Expect(respPost.Code).To(Equal(http.StatusOK))

			reqPatch, err := http.NewRequest(http.MethodPatch, formatterByNameURL(formatterName), makePayload(api.FormatterUpdateRequest{}))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantId)
			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantId)))
			Expect(respPatch.Code).To(Equal(http.StatusBadRequest), respPatch.Body.String())
			Expect(respPatch.Body.String()).To(ContainSubstring("patch must include at least one field in fields"))
		})

		It("patches formatter successfully when stored fields json is null", func() {
			name := fmt.Sprintf("nil-fields-%d", tenantId)
			now := time.Now().UTC()
			row := &dbmodel.MappingFormatters{
				ID:          uuid.New(),
				Name:        name,
				ScopeType:   api.MAPPINGSCOPE_TENANT,
				TenantId:    tenantId,
				Fields:      map[string]api.FormatterField{"temp": {SetExpr: convutils.Ref(`{"x":1}`)}},
				Description: convutils.Ref("seed"),
				CreatedBy:   "admin",
				CreatedAt:   now,
				UpdatedBy:   "admin",
				UpdatedAt:   now,
			}
			Expect(dao.CreateFormatter(tenantContextForTest(tenantId), row)).To(BeNil())

			// force invalid-but-permitted jsonb value to mimic legacy/corrupt data and hit row.Fields == nil branch
			Expect(runaidb.DB().Exec("update mapping_formatters set fields = 'null'::jsonb where id = ?", row.ID)).Error().To(BeNil())

			patchBody := api.FormatterUpdateRequest{
				Fields: map[string]api.FormatterField{
					"labels": {
						SetExpr: &setLabels,
						ArgsInfo: map[string]api.MappingFieldType{
							"value": api.MAPPINGFIELDTYPE_QUANTITY,
						},
					},
				},
			}
			reqPatch, err := http.NewRequest(http.MethodPatch, formatterByNameURL(name), makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantId)

			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantId)))
			Expect(respPatch.Code).To(Equal(http.StatusOK), respPatch.Body.String())

			var patched api.Formatter
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &patched)).To(Succeed())
			Expect(patched.Fields).To(HaveKey("labels"))
		})

		It("isolates formatter resources between tenants", func() {
			otherTenant := int32(randoms.CloneWithDifferentTenant().TenantId)
			sharedName := fmt.Sprintf("formatter-shared-%d", tenantId)

			createBody := api.FormatterCreationRequest{
				Name:   sharedName,
				Fields: formatterWithContainerPortField(portsSet, portsGet),
			}

			// tenant A creates formatter
			reqCreateA, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
			Expect(err).To(BeNil())
			reqCreateA.Header = *headerForAdmin(tenantId)

			respCreateA := httptest.NewRecorder()
			router.ServeHTTP(respCreateA, reqCreateA.WithContext(tenantContextForTest(tenantId)))
			Expect(respCreateA.Code).To(Equal(http.StatusOK), respCreateA.Body.String())

			// tenant B list should not include tenant A formatter
			reqListB, err := http.NewRequest(http.MethodGet, urlFormatters, nil)
			Expect(err).To(BeNil())
			reqListB.Header = *headerForAdmin(otherTenant)

			respListB := httptest.NewRecorder()
			router.ServeHTTP(respListB, reqListB.WithContext(tenantContextForTest(otherTenant)))
			Expect(respListB.Code).To(Equal(http.StatusOK), respListB.Body.String())

			var listRespB api.FormatterListResponse
			Expect(json.Unmarshal(respListB.Body.Bytes(), &listRespB)).To(Succeed())
			for _, item := range listRespB.Items {
				Expect(item.Name).ToNot(Equal(sharedName))
			}

			// tenant B cannot get tenant A formatter
			reqGetB, err := http.NewRequest(http.MethodGet, formatterByNameURL(sharedName), nil)
			Expect(err).To(BeNil())
			reqGetB.Header = *headerForAdmin(otherTenant)

			respGetB := httptest.NewRecorder()
			router.ServeHTTP(respGetB, reqGetB.WithContext(tenantContextForTest(otherTenant)))
			Expect(respGetB.Code).To(Equal(http.StatusNotFound), respGetB.Body.String())
			Expect(respGetB.Body.String()).To(ContainSubstring(fmt.Sprintf("formatter named %s not found", sharedName)))

			// tenant B cannot update tenant A formatter
			reqPatchB, err := http.NewRequest(
				http.MethodPatch,
				formatterByNameURL(sharedName),
				makePayload(api.FormatterUpdateRequest{
					Fields: map[string]api.FormatterField{
						"labels": {
							SetExpr: &setLabels,
							ArgsInfo: map[string]api.MappingFieldType{
								"value": api.MAPPINGFIELDTYPE_QUANTITY,
							},
						},
					},
				}),
			)
			Expect(err).To(BeNil())
			reqPatchB.Header = *headerForAdmin(otherTenant)

			respPatchB := httptest.NewRecorder()
			router.ServeHTTP(respPatchB, reqPatchB.WithContext(tenantContextForTest(otherTenant)))
			Expect(respPatchB.Code).To(Equal(http.StatusNotFound), respPatchB.Body.String())
			Expect(respPatchB.Body.String()).To(ContainSubstring(fmt.Sprintf("formatter named %s not found", sharedName)))

			// tenant B cannot delete fields from tenant A formatter
			reqDelFieldsB, err := http.NewRequest(
				http.MethodDelete,
				formatterFieldsURL(sharedName),
				makePayload(api.FormatterDeleteFieldsRequest{Fields: []string{"ports"}}),
			)
			Expect(err).To(BeNil())
			reqDelFieldsB.Header = *headerForAdmin(otherTenant)

			respDelFieldsB := httptest.NewRecorder()
			router.ServeHTTP(respDelFieldsB, reqDelFieldsB.WithContext(tenantContextForTest(otherTenant)))
			Expect(respDelFieldsB.Code).To(Equal(http.StatusNotFound), respDelFieldsB.Body.String())
			Expect(respDelFieldsB.Body.String()).To(ContainSubstring(fmt.Sprintf("formatter named %s not found", sharedName)))

			// tenant B can create same name in its own tenant only
			reqCreateB, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
			Expect(err).To(BeNil())
			reqCreateB.Header = *headerForAdmin(otherTenant)

			respCreateB := httptest.NewRecorder()
			router.ServeHTTP(respCreateB, reqCreateB.WithContext(tenantContextForTest(otherTenant)))
			Expect(respCreateB.Code).To(Equal(http.StatusOK), respCreateB.Body.String())
		})
	})

	Context("runai- system prefix guard", func() {
		It("rejects create when name uses runai- prefix", func() {
			body := api.FormatterCreationRequest{
				Name:   "runai-forbidden",
				Fields: formatterWithContainerPortField(portsSet, portsGet),
			}

			req, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(body))
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantId)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring("prefix 'runai-' is not allowed for tenant scoped formatter"))
		})

		It("rejects patch when path name uses runai- prefix", func() {
			patchBody := api.FormatterUpdateRequest{
				Fields: map[string]api.FormatterField{
					"x": {
						SetExpr: convutils.Ref(`{"a":$ARGS.named.k}`),
						ArgsInfo: map[string]api.MappingFieldType{
							"k": api.MAPPINGFIELDTYPE_STRING,
						},
					},
				},
			}

			req, err := http.NewRequest(http.MethodPatch, formatterByNameURL("runai-nope"), makePayload(patchBody))
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantId)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring("prefix 'runai-' is not allowed for tenant scoped formatter"))
		})

		It("rejects delete when path name uses runai- prefix", func() {
			req, err := http.NewRequest(http.MethodDelete, formatterByNameURL("runai-nope"), nil)
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantId)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring("prefix 'runai-' is not allowed for tenant scoped formatter"))
		})

		It("rejects delete-fields when path name uses runai- prefix", func() {
			req, err := http.NewRequest(
				http.MethodDelete,
				formatterFieldsURL("runai-nope"),
				makePayload(api.FormatterDeleteFieldsRequest{Fields: []string{"x"}}),
			)
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantId)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring("prefix 'runai-' is not allowed for tenant scoped formatter"))
		})
	})

	Context("builtin formatter name guard", func() {
		reservedFormatterNames := []string{"KeyValue", "KeyValueMap"}

		It("rejects create when formatter name is reserved", func() {
			for _, name := range reservedFormatterNames {
				body := api.FormatterCreationRequest{
					Name:   name,
					Fields: formatterWithContainerPortField(portsSet, portsGet),
				}
				req, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(body))
				Expect(err).To(BeNil())
				req.Header = *headerForAdmin(tenantId)

				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
				Expect(resp.Code).To(Equal(http.StatusBadRequest), name+": "+resp.Body.String())
				Expect(resp.Body.String()).To(ContainSubstring(fmt.Sprintf("formatter %s is a builtin formatter and cannot be created", name)))
			}
		})

		It("rejects get when path name is reserved", func() {
			for _, name := range reservedFormatterNames {
				req, err := http.NewRequest(http.MethodGet, formatterByNameURL(name), nil)
				Expect(err).To(BeNil())
				req.Header = *headerForAdmin(tenantId)

				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
				Expect(resp.Code).To(Equal(http.StatusBadRequest), name+": "+resp.Body.String())
				Expect(resp.Body.String()).To(ContainSubstring(fmt.Sprintf("formatter %s is a builtin formatter", name)))
			}
		})

		It("rejects patch when path name is reserved", func() {
			patchBody := api.FormatterUpdateRequest{
				Fields: map[string]api.FormatterField{
					"x": singleArgStringFormatterField(),
				},
			}
			for _, name := range reservedFormatterNames {
				req, err := http.NewRequest(http.MethodPatch, formatterByNameURL(name), makePayload(patchBody))
				Expect(err).To(BeNil())
				req.Header = *headerForAdmin(tenantId)

				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
				Expect(resp.Code).To(Equal(http.StatusBadRequest), name+": "+resp.Body.String())
				Expect(resp.Body.String()).To(ContainSubstring(fmt.Sprintf("formatter %s is a builtin formatter", name)))
			}
		})

		It("rejects delete when path name is reserved", func() {
			for _, name := range reservedFormatterNames {
				req, err := http.NewRequest(http.MethodDelete, formatterByNameURL(name), nil)
				Expect(err).To(BeNil())
				req.Header = *headerForAdmin(tenantId)

				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
				Expect(resp.Code).To(Equal(http.StatusBadRequest), name+": "+resp.Body.String())
				Expect(resp.Body.String()).To(ContainSubstring(fmt.Sprintf("formatter %s is a builtin formatter", name)))
			}
		})

		It("rejects delete-fields when path name is reserved", func() {
			for _, name := range reservedFormatterNames {
				req, err := http.NewRequest(
					http.MethodDelete,
					formatterFieldsURL(name),
					makePayload(api.FormatterDeleteFieldsRequest{Fields: []string{"x"}}),
				)
				Expect(err).To(BeNil())
				req.Header = *headerForAdmin(tenantId)

				resp := httptest.NewRecorder()
				router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
				Expect(resp.Code).To(Equal(http.StatusBadRequest), name+": "+resp.Body.String())
				Expect(resp.Body.String()).To(ContainSubstring(fmt.Sprintf("formatter %s is a builtin formatter", name)))
			}
		})
	})

	Context("argsInfo automatic enrichment", func() {
		It("auto-populates argsInfo with string arguments from expressions", func() {
			name := fmt.Sprintf("auto-args-formatter-%d", tenantId)
			setExpr := `{"labels": {"app": $ARGS.named.appName, "version": $ARGS.named.version}}`
			getExpr := `.metadata.labels[$ARGS.named.labelKey]`

			createBody := api.FormatterCreationRequest{
				Name: name,
				Fields: map[string]api.FormatterField{
					"labels": {
						SetExpr: &setExpr,
						GetExpr: &getExpr,
					},
				},
			}

			req, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantId)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
			Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

			var created api.Formatter
			Expect(json.Unmarshal(resp.Body.Bytes(), &created)).To(Succeed())

			Expect(created.Fields).To(HaveKey("labels"))
			labelsField := created.Fields["labels"]
			Expect(labelsField.ArgsInfo).To(HaveLen(3))
			Expect(labelsField.ArgsInfo).To(HaveKeyWithValue("appName", api.MAPPINGFIELDTYPE_STRING))
			Expect(labelsField.ArgsInfo).To(HaveKeyWithValue("version", api.MAPPINGFIELDTYPE_STRING))
			Expect(labelsField.ArgsInfo).To(HaveKeyWithValue("labelKey", api.MAPPINGFIELDTYPE_STRING))
		})

		It("preserves explicitly provided argsInfo types while adding missing string args", func() {
			name := fmt.Sprintf("mixed-args-formatter-%d", tenantId)
			setExpr := `{"port": $ARGS.named.portNum, "name": $ARGS.named.portName}`

			createBody := api.FormatterCreationRequest{
				Name: name,
				Fields: map[string]api.FormatterField{
					"ports": {
						SetExpr: &setExpr,
						ArgsInfo: map[string]api.MappingFieldType{
							"portNum": api.MAPPINGFIELDTYPE_INTEGER,
						},
					},
				},
			}

			req, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(tenantId)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
			Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

			var created api.Formatter
			Expect(json.Unmarshal(resp.Body.Bytes(), &created)).To(Succeed())

			Expect(created.Fields).To(HaveKey("ports"))
			portsField := created.Fields["ports"]
			Expect(portsField.ArgsInfo).To(HaveLen(2))
			Expect(portsField.ArgsInfo).To(HaveKeyWithValue("portNum", api.MAPPINGFIELDTYPE_INTEGER))
			Expect(portsField.ArgsInfo).To(HaveKeyWithValue("portName", api.MAPPINGFIELDTYPE_STRING))
		})

		It("auto-populates argsInfo on patch for new fields", func() {
			name := fmt.Sprintf("patch-auto-args-%d", tenantId)
			initialSetExpr := `{"x": 1}`

			createBody := api.FormatterCreationRequest{
				Name: name,
				Fields: map[string]api.FormatterField{
					"simple": {SetExpr: &initialSetExpr},
				},
			}
			reqCreate, err := http.NewRequest(http.MethodPost, urlFormatters, makePayload(createBody))
			Expect(err).To(BeNil())
			reqCreate.Header = *headerForAdmin(tenantId)
			respCreate := httptest.NewRecorder()
			router.ServeHTTP(respCreate, reqCreate.WithContext(tenantContextForTest(tenantId)))
			Expect(respCreate.Code).To(Equal(http.StatusOK), respCreate.Body.String())

			patchSetExpr := `{"env": $ARGS.named.envName, "value": $ARGS.named.envValue}`
			patchBody := api.FormatterUpdateRequest{
				Fields: map[string]api.FormatterField{
					"envVars": {SetExpr: &patchSetExpr},
				},
			}

			reqPatch, err := http.NewRequest(http.MethodPatch, formatterByNameURL(name), makePayload(patchBody))
			Expect(err).To(BeNil())
			reqPatch.Header = *headerForAdmin(tenantId)

			respPatch := httptest.NewRecorder()
			router.ServeHTTP(respPatch, reqPatch.WithContext(tenantContextForTest(tenantId)))
			Expect(respPatch.Code).To(Equal(http.StatusOK), respPatch.Body.String())

			var patched api.Formatter
			Expect(json.Unmarshal(respPatch.Body.Bytes(), &patched)).To(Succeed())

			Expect(patched.Fields).To(HaveKey("envVars"))
			envVarsField := patched.Fields["envVars"]
			Expect(envVarsField.ArgsInfo).To(HaveLen(2))
			Expect(envVarsField.ArgsInfo).To(HaveKeyWithValue("envName", api.MAPPINGFIELDTYPE_STRING))
			Expect(envVarsField.ArgsInfo).To(HaveKeyWithValue("envValue", api.MAPPINGFIELDTYPE_STRING))
		})
	})
})

var _ = Describe("Formatter mapping implementation helpers", func() {
	var ms MappingApiService
	var validSet string
	var invalidExpr string

	BeforeEach(func() {
		ms = MappingApiService{}
		validSet = `{"x": $ARGS.named.k}`
		invalidExpr = "{"
	})

	Context("normalizeFormatterListParams", func() {
		It("applies default limit when limit is zero", func() {
			off, lim := normalizeListParams(0, 0, formatterListDefaultLimit, formatterListMaxLimit)
			Expect(off).To(Equal(0))
			Expect(lim).To(Equal(formatterListDefaultLimit))
		})

		It("normalizes negative offset to zero", func() {
			off, lim := normalizeListParams(-7, 10, formatterListDefaultLimit, formatterListMaxLimit)
			Expect(off).To(Equal(0))
			Expect(lim).To(Equal(10))
		})

		It("caps limit at max", func() {
			off, lim := normalizeListParams(3, 9999, formatterListDefaultLimit, formatterListMaxLimit)
			Expect(off).To(Equal(3))
			Expect(lim).To(Equal(formatterListMaxLimit))
		})

		It("passes through valid offset and limit", func() {
			off, lim := normalizeListParams(2, 42, formatterListDefaultLimit, formatterListMaxLimit)
			Expect(off).To(Equal(2))
			Expect(lim).To(Equal(42))
		})
	})

	Context("getFormatter", func() {
		It("rejects blank formatter name", func() {
			row, err := ms.getFormatter(context.Background(), "   ")
			Expect(row).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(Equal("formatter name is required"))
		})
	})

	Context("validateJQExpression", func() {
		It("rejects whitespace-only expression", func() {
			expr := "   "
			q, err := ms.validateJQExpression("f1", "fieldA", "setExpr", &expr)
			Expect(q).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(Equal("formatter f1: empty setExpr for field fieldA"))
		})

		It("rejects invalid jq", func() {
			expr := "{"
			q, err := ms.validateJQExpression("f1", "fieldA", "getExpr", &expr)
			Expect(q).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(HavePrefix("formatter f1: invalid getExpr query for field fieldA: "))
		})

		DescribeTable("rejects unsafe JQ constructs",
			func(expr string, errFragment string) {
				q, err := ms.validateJQExpression("fmt1", "fieldA", "setExpr", &expr)
				Expect(q).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
				Expect(err.Message).To(ContainSubstring("unsafe setExpr query for field fieldA"))
				Expect(err.Message).To(ContainSubstring(errFragment))
			},
			Entry("del function", `del(.x)`, "del"),
			Entry("recursive descent ..", `.. | .x`, "recursive descent"),
			Entry("recurse function", `recurse | .x`, "recurse"),
			Entry("paths function", `paths`, "paths"),
			Entry("range function", `range(1000000)`, "range"),
			Entry("walk function", `walk(.)`, "walk"),
			Entry("repeat function", `first(repeat(. + 1))`, "repeat"),
			Entry("assignment operator", `.x = 1`, "modifying operator"),
			Entry("update operator", `.x |= . + 1`, "modifying operator"),
		)

		It("accepts valid safe JQ expression", func() {
			expr := `.metadata.labels[$ARGS.named.key]`
			q, err := ms.validateJQExpression("fmt1", "fieldA", "getExpr", &expr)
			Expect(err).To(BeNil())
			Expect(q).NotTo(BeNil())
		})
	})

	Context("builtin formatter guard (validateFormatter)", func() {
		It("rejects validation when formatter name is KeyValue", func() {
			err := ms.validateFormatter("KeyValue", map[string]api.FormatterField{"f": {SetExpr: &validSet}})
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(Equal("formatter KeyValue is a builtin formatter"))
		})

		It("rejects validation when formatter name is KeyValueMap", func() {
			err := ms.validateFormatter("KeyValueMap", map[string]api.FormatterField{"f": {SetExpr: &validSet}})
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(Equal("formatter KeyValueMap is a builtin formatter"))
		})
	})

	Context("validateFields", func() {
		It("rejects invalid field map key (MappingFieldName)", func() {
			err := ms.validateFormatter("fmt1", map[string]api.FormatterField{"x!x": {SetExpr: &validSet}})
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(Equal(
				`field name "x!x" is invalid: must be alphanumeric with optional internal hyphens, and must not start or end with a hyphen`,
			))
		})

		It("rejects field map key longer than 63 characters", func() {
			longKey := strings.Repeat("a", 64)
			err := ms.validateFormatter("fmt1", map[string]api.FormatterField{
				longKey: {SetExpr: &validSet, ArgsInfo: map[string]api.MappingFieldType{"k": api.MAPPINGFIELDTYPE_STRING}},
			})
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(Equal(fmt.Sprintf(`field name %q must be between 1 and 63 characters`, longKey)))
		})

		It("rejects invalid setExpr", func() {
			err := ms.validateFormatter("fmt1", map[string]api.FormatterField{"f": {SetExpr: &invalidExpr}})
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(ContainSubstring("invalid setExpr query for field f: "))
		})

		It("rejects invalid getExpr", func() {
			err := ms.validateFormatter("fmt1", map[string]api.FormatterField{"f": {
				SetExpr: &validSet,
				GetExpr: &invalidExpr,
			}})
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(HavePrefix("formatter fmt1: invalid getExpr query for field f: "))
		})

		It("rejects when both setExpr and getExpr are empty", func() {
			err := ms.validateFormatter("fmt1", map[string]api.FormatterField{"f": {}})
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(Equal("formatter fmt1: field f: both setExpr and getExpr are empty"))
		})

		It("rejects argsInfo argument not referenced in jq", func() {
			err := ms.validateFormatter("fmt1", map[string]api.FormatterField{"f": {
				SetExpr: &validSet,
				ArgsInfo: map[string]api.MappingFieldType{
					"missingArg": api.MAPPINGFIELDTYPE_STRING,
				},
			}})
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(Equal("formatter fmt1: field f: argument missingArg is not referenced in the get/set queries"))
		})

		It("rejects array key arg not referenced in jq", func() {
			err := ms.validateFormatter("fmt1", map[string]api.FormatterField{"f": {
				SetExpr:  &validSet,
				ArrayKey: &api.FormatterArrayKey{ArgName: "containerPort"},
			}})
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(Equal("formatter fmt1: field f: array key arg containerPort is not referenced in the get/set queries"))
		})

		It("rejects enumeration on non-string arg", func() {
			err := ms.validateFormatter("fmt1", map[string]api.FormatterField{"f": {
				SetExpr: &validSet,
				ArgsEnum: map[string][]string{
					"enumOnly": {"A", "B"},
				},
			}})
			Expect(err).ToNot(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
			Expect(err.Message).To(Equal("formatter fmt1: field f: enumeration is not allowed for arg enumOnly"))
		})
	})
})
