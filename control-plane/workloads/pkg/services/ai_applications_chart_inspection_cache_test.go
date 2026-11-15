// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	chartinspection "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/chart_inspection"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	cachetesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chart inspection cache sharing", func() {
	const asset = "11111111-1111-4111-8111-111111111111"
	const otherAsset = "22222222-2222-4222-8222-222222222222"
	const credential = "<test-only-key>"
	var archive []byte
	var request api.AiApplicationChartInspectionRequest
	caller := func(tenant, user string) context.Context {
		ctx := inspectionTestContext(tenant, user)
		ctx = context.WithValue(ctx, runaictx.CfClientId, "client-"+user)
		ctx = context.WithValue(ctx, runaictx.CfGroups, []string{"group-" + user})
		return context.WithValue(ctx, runaictx.CfSupplementalGroups, []string{"extra-" + user})
	}
	BeforeEach(func() {
		cachetesting.EnableCache()
		archive = inspectionTestArchive(GinkgoT())
		request = api.AiApplicationChartInspectionRequest{
			ChartRef: api.AiApplicationSpecChartRef{Repo: "https://helm.ngc.nvidia.com/nvidia", ChartName: "example", ChartVersion: "1.0.0"},
			Auth:     api.AiApplicationAuth{AssetId: asset},
		}
	})
	AfterEach(func() { cachetesting.DisableCache() })

	It("returns metadata and typed errors before HTTP response conversion", func() {
		denied := false
		accessErr := api.NewError(http.StatusForbidden, "credential access denied", nil)
		c := &applicationChartInspector{
			inspector: chartinspection.NewInspector(func(context.Context, chartinspection.Selection, string) ([]byte, *api.Error) {
				return archive, nil
			}, 0),
			loadKey: func(context.Context, string) (string, *api.Error) {
				if denied {
					return "", accessErr
				}
				return credential, nil
			},
		}
		request.Auth = api.AiApplicationAuth{AssetId: asset}
		result, err := c.inspect(caller("1", "alice"), request)
		Expect(err).To(BeNil())
		Expect(result).NotTo(BeNil())
		Expect(result.ChartName).To(Equal(request.ChartRef.ChartName))
		denied = true
		result, err = c.inspect(caller("1", "alice"), request)
		Expect(result).To(BeNil())
		Expect(err).To(BeIdenticalTo(accessErr))
	})

	DescribeTable("reuses only equivalent authorized selections", func(tenant, user, firstAsset, nextAsset, nextKey string, personal bool, wantDownloads int32) {
		var downloads, lookups atomic.Int32
		key := credential
		c := &applicationChartInspector{
			inspector: chartinspection.NewInspector(func(context.Context, chartinspection.Selection, string) ([]byte, *api.Error) {
				downloads.Add(1)
				return archive, nil
			}, 0),
			loadKey: func(context.Context, string) (string, *api.Error) { lookups.Add(1); return key, nil },
		}
		request.Auth = api.AiApplicationAuth{AssetId: firstAsset, UserCredential: convutils.Ref(personal)}
		first, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(caller("1", "alice"), request)
		Expect(err).NotTo(HaveOccurred())
		Expect(first.Code).To(Equal(http.StatusOK))
		key = nextKey
		request.Auth.AssetId = nextAsset
		second, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(caller(tenant, user), request)
		Expect(err).NotTo(HaveOccurred())
		Expect(second.Code).To(Equal(http.StatusOK))
		Expect(second.Body).To(Equal(first.Body))
		Expect(downloads.Load()).To(Equal(wantDownloads))
		expectedLookups := int32(0)
		if firstAsset != "" {
			expectedLookups++
		}
		if nextAsset != "" {
			expectedLookups++
		}
		Expect(lookups.Load()).To(Equal(expectedLookups))
	},
		Entry("shared credential across users, clients and groups", "1", "bob", asset, asset, credential, false, int32(1)),
		Entry("different tenant", "2", "alice", asset, asset, credential, false, int32(2)),
		Entry("different asset with identical key material", "1", "alice", asset, otherAsset, credential, false, int32(2)),
		Entry("rotated key material", "1", "alice", asset, asset, "<rotated-test-key>", false, int32(2)),
		Entry("personal credentials remain caller isolated", "1", "bob", asset, asset, credential, true, int32(2)),
		Entry("same personal caller reuses its result", "1", "alice", asset, asset, credential, true, int32(1)),
		Entry("public charts share within the tenant", "1", "bob", "", "", "", false, int32(1)),
		Entry("public charts remain tenant isolated", "2", "alice", "", "", "", false, int32(2)),
		Entry("public and credentialed requests remain separate", "1", "alice", asset, "", "", false, int32(2)),
	)

	DescribeTable("denies another user's personal asset even when marked shared", func(flag *bool) {
		var downloads atomic.Int32
		c := &applicationChartInspector{
			inspector: chartinspection.NewInspector(func(context.Context, chartinspection.Selection, string) ([]byte, *api.Error) {
				downloads.Add(1)
				return archive, nil
			}, 0),
			// Assets enforces ownership independently of the request's userCredential flag.
			loadKey: func(ctx context.Context, _ string) (string, *api.Error) {
				if *runaictx.GetString(ctx, runaictx.CfSubjectID) != "alice" {
					return "", api.NewError(http.StatusNotFound, "credential not found", nil)
				}
				return credential, nil
			},
		}
		request.Auth.UserCredential = flag
		first, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(caller("1", "alice"), request)
		Expect(err).NotTo(HaveOccurred())
		Expect(first.Code).To(Equal(http.StatusOK))
		second, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(caller("1", "bob"), request)
		Expect(err).NotTo(HaveOccurred())
		Expect(second.Code).To(Equal(http.StatusNotFound))
		Expect(downloads.Load()).To(Equal(int32(1)))
	}, Entry("personal", convutils.Ref(true)), Entry("shared", convutils.Ref(false)), Entry("omitted", (*bool)(nil)))

	DescribeTable("uses the same scope for in-flight work", func(tenant, user, nextAsset, nextKey string, personal bool, wantDownloads int32) {
		// Disable metadata storage so the second caller must join the running worker.
		cachetesting.DisableCache()
		var downloads, lookups atomic.Int32
		release := make(chan struct{})
		var once sync.Once
		unblock := func() { once.Do(func() { close(release) }) }
		defer unblock()
		c := &applicationChartInspector{
			inspector: chartinspection.NewInspector(func(ctx context.Context, _ chartinspection.Selection, _ string) ([]byte, *api.Error) {
				downloads.Add(1)
				select {
				case <-release:
					return archive, nil
				case <-ctx.Done():
					return nil, api.NewError(http.StatusRequestTimeout, "chart inspection request cancelled or expired", nil)
				}
			}, 0),
			loadKey: func(ctx context.Context, _ string) (string, *api.Error) {
				lookups.Add(1)
				if *runaictx.GetString(ctx, runaictx.CfSubjectID) == "denied" {
					return "", api.NewError(http.StatusForbidden, "credential access denied", nil)
				}
				if *runaictx.GetString(ctx, runaictx.CfSubjectID) == "alice" {
					return credential, nil
				}
				return nextKey, nil
			},
		}
		request.Auth.UserCredential = convutils.Ref(personal)
		if nextAsset == "" {
			request.Auth.AssetId = ""
		}
		responses := make(chan api.ImplResponse, 2)
		inspect := func(callCtx context.Context, req api.AiApplicationChartInspectionRequest) {
			defer GinkgoRecover()
			out, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(callCtx, req)
			Expect(err).NotTo(HaveOccurred())
			responses <- out
		}
		// Tie callers to cleanup without losing their trusted identities.
		firstCtx, firstCancel := context.WithCancel(caller("1", "alice"))
		defer firstCancel()
		go inspect(firstCtx, request)
		Eventually(downloads.Load).Should(Equal(int32(1)))
		nextRequest := request
		nextRequest.Auth.AssetId = nextAsset
		secondCtx, secondCancel := context.WithCancel(caller(tenant, user))
		defer secondCancel()
		go inspect(secondCtx, nextRequest)
		if nextAsset != "" {
			Eventually(lookups.Load).Should(Equal(int32(2)))
		}
		if wantDownloads > 1 {
			Eventually(downloads.Load).Should(Equal(wantDownloads))
		} else {
			Consistently(downloads.Load, 100*time.Millisecond).Should(Equal(int32(1)))
		}
		deniedCtx, deniedCancel := context.WithTimeout(caller("1", "denied"), time.Second)
		defer deniedCancel()
		deniedRequest := request
		deniedRequest.Auth.AssetId = asset
		denied, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(deniedCtx, deniedRequest)
		Expect(err).NotTo(HaveOccurred())
		Expect(denied.Code).To(Equal(http.StatusForbidden))
		unblock()
		for n := 0; n < 2; n++ {
			var out api.ImplResponse
			Eventually(responses).Should(Receive(&out))
			Expect(out.Code).To(Equal(http.StatusOK))
		}
		Expect(downloads.Load()).To(Equal(wantDownloads))
	},
		Entry("shared credential", "1", "bob", asset, credential, false, int32(1)),
		Entry("public chart", "1", "bob", "", "", false, int32(1)),
		Entry("different tenant", "2", "bob", asset, credential, false, int32(2)),
		Entry("different asset", "1", "bob", otherAsset, credential, false, int32(2)),
		Entry("rotated credential", "1", "bob", asset, "<rotated-test-key>", false, int32(2)),
		Entry("personal credential", "1", "bob", asset, credential, true, int32(2)),
	)
})
