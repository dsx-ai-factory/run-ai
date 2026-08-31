/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
)

var fail healthz.Checker = func(_ *http.Request) error { return fmt.Errorf("i'm a failure") }

var _ = Describe("ExternalAccessCheckerServer", func() {
	var (
		mockInternalServer     *mocks.WebhookServer
		externalAccessedServer *ExternalAccessCheckerServer
		externalAddress        string
		externalTimeout        time.Duration
		mockCtrl               *gomock.Controller
	)
	AfterEach(func() {
		mockCtrl.Finish()
	})
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockInternalServer = &mocks.WebhookServer{}
		externalAddress = "127.0.0.1:8080"
		externalTimeout = 100 * time.Millisecond
		externalAccessedServer = NewExternalAccessCheckerServer(mockInternalServer, externalAddress, externalTimeout)
	})

	Describe("StartedChecker", func() {

		Context("when the internal server's StartedChecker fails", func() {
			It("should return an error indicating the internal server is unreachable", func() {

				mockInternalServer.On("StartedChecker").Return(fail)
				checker := externalAccessedServer.StartedChecker()
				req := &http.Request{}
				err := checker(req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("webhook server is not reachable internally"))
				mockInternalServer.AssertCalled(GinkgoT(), "StartedChecker")
			})
		})

		Context("when the external server cannot be connected", func() {
			It("should return an error indicating the external server is unreachable", func() {
				// no external server listening
				mockInternalServer.On("StartedChecker").Return(healthz.Ping)
				checker := externalAccessedServer.StartedChecker()
				req := &http.Request{}
				err := checker(req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("webhook server is not reachable externally"))
				mockInternalServer.AssertCalled(GinkgoT(), "StartedChecker")
			})
		})
	})
})
