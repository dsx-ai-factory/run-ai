package query_types

import (
	"context"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func setContext(scopeId string, subjectId string) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, scopeId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	return ctx
}

var _ = Describe("RBAC Rules Test", func() {
	Describe("negative crud test", func() {
		Context("fail to check permitted scopes", func() {
			It("fail the request", func() {
				orgunit.SetMock(&orgunit.FailingMockClient{
					FailGetOrgTree: true,
				})
				authorization.SetMock(&authorization.FailingMockFactory{})
				queryParams := QueryParams{
					ProjectId: convutils.Ref(int32(1)),
				}
				_, err := ApplyRbacRules(context.Background(), &queryParams)
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("failed to interrogate the organizational tree"))
			})
		})
	})

	Describe("positive CRUD test", func() {
		var (
			authMock *testing.RandomsBasedAuthMockClient
			ctx      context.Context
		)

		type TestParams struct {
			scope               api.Scope
			setAdditionalFields func(assetMeta *api.AssetMeta, randoms *testutils.RandomValues)
			randoms             *testutils.RandomValues
		}

		createTestParams := func(scope api.Scope) TestParams {
			randoms := testutils.NewRandomValues()
			return TestParams{
				scope: scope,
				setAdditionalFields: func(assetMeta *api.AssetMeta, randoms *testutils.RandomValues) {
					switch scope {
					case api.SCOPE_TENANT:
						assetMeta.TenantId = convutils.Ref(int32(randoms.TenantId))
					case api.SCOPE_CLUSTER:
						assetMeta.ClusterId = convutils.Ref(randoms.ClusterId)
					case api.SCOPE_DEPARTMENT:
						assetMeta.DepartmentId = convutils.Ref(randoms.DepartmentId)
					case api.SCOPE_PROJECT:
						assetMeta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
					}
				},
				randoms: randoms,
			}
		}

		DescribeTable("Get asset by id", func(testParams TestParams) {
			authMock = testing.NewRandomsBasedAuthMockClient(testParams.randoms)
			authorization.SetMock(authMock)
			org_unit.SetMock(org_unit.NewSimpleMock(testParams.randoms))
			ctx = setContext(testParams.randoms.TenantIdStr, "test@run.ai")

			assetMeta := api.AssetMeta{
				Id:    uuid.NewString(),
				Name:  "asset-for-rbac-test",
				Kind:  api.ASSETKIND_S3,
				Scope: testParams.scope,
			}
			testParams.setAdditionalFields(&assetMeta, testParams.randoms)
			permittedScopes, _, err := ApplyRbacRulesGetById(ctx, assetMeta)
			Expect(err).To(BeNil())
			Expect(permittedScopes).NotTo(BeNil())
		},
			Entry("Tenant scoped asset", createTestParams(api.SCOPE_TENANT)),
			Entry("Cluster scoped asset", createTestParams(api.SCOPE_CLUSTER)),
			Entry("Department scoped asset", createTestParams(api.SCOPE_DEPARTMENT)),
			Entry("Project scoped asset", createTestParams(api.SCOPE_PROJECT)),
		)
	})

})
