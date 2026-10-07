/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServicesV1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services V1 Suite")
}

// rolesHTTPClient sends the roles calls through the real generated client so the specs see the actual request.
type rolesHTTPClient struct {
	auth_mocks.IAuthorizationClient
	rolesAPI openapi.RolesAPI
}

func (c *rolesHTTPClient) GetRoles(ctx context.Context) openapi.ApiGetRolesRequest {
	return c.rolesAPI.GetRoles(ctx)
}

func (c *rolesHTTPClient) GetRolesExecute(r openapi.ApiGetRolesRequest) (*openapi.GetRoles200Response, *http.Response, error) {
	return c.rolesAPI.GetRolesExecute(r)
}

var _ = Describe("DepartmentsV1ApiService", func() {
	Describe("getDepartmentRole", func() {
		var (
			service      *DepartmentsV1ApiService
			gotURL       *url.URL
			statusCode   int
			responseBody string
		)

		BeforeEach(func() {
			gotURL = nil
			statusCode = http.StatusOK
			responseBody = `{"roles":[{"id":42,"name":"Department administrator","description":"","createdBy":"Run:ai","custom":false,"effectiveEnabled":true,"permissions":[]}]}`

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotURL = r.URL
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(statusCode)
				_, _ = w.Write([]byte(responseBody))
			}))
			DeferCleanup(server.Close)

			config := openapi.NewConfiguration()
			config.Servers = openapi.ServerConfigurations{{URL: server.URL}}
			service = &DepartmentsV1ApiService{
				authorizationClient: &rolesHTTPClient{rolesAPI: openapi.NewAPIClient(config).RolesAPI},
			}
		})

		It("looks the role up by exact name on the v2 roles API", func() {
			role, err := service.getDepartmentRole(context.Background())
			Expect(err).ToNot(HaveOccurred())

			Expect(gotURL).ToNot(BeNil())
			Expect(gotURL.Path).To(Equal("/api/v2/authorization/roles"))
			Expect(gotURL.Query()["filterBy"]).To(Equal([]string{"name==Department administrator"}))
			Expect(role.Id).To(Equal(int32(42)))
			Expect(role.Name).To(Equal(DepartmentAdminRole))
		})

		It("picks the role whose name matches exactly", func() {
			responseBody = `{"roles":[{"id":7,"name":"Department administrator (Legacy)","permissions":[]},{"id":42,"name":"Department administrator","permissions":[]}]}`

			role, err := service.getDepartmentRole(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(role.Id).To(Equal(int32(42)))
		})

		It("returns an error when the roles call fails", func() {
			statusCode = http.StatusInternalServerError
			responseBody = `{"code":500,"message":"internal error"}`

			role, err := service.getDepartmentRole(context.Background())
			Expect(err).To(HaveOccurred())
			Expect(role).To(BeNil())
		})
	})
})
