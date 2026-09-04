/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package portforward_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mocks2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/connection/mocks"
	"k8s.io/apimachinery/pkg/util/httpstream"

	mockspf "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/portforward"
	mocks "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/websocket"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/portforward"

	workloadDomain "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLogs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Portforward Suite")
}

const (
	ValidPFWithPort = "/port-forward/?port=1234"

	InvalidJustPod = "/port-forward/?pod=name"
	InvalidNoQuery = "/port-forward"
)

var _ = Describe("Handler", func() {
	var (
		ctrl         *gomock.Controller
		mockService  *mockspf.MockPodPFService
		mockUpgrader *mocks.MockUpgrader
		handler      *portforward.Handler
		req          *http.Request
		recorder     *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockService = mockspf.NewMockPodPFService(ctrl)
		mockUpgrader = mocks.NewMockUpgrader(ctrl)
		handler = portforward.NewHandler(mockService, mockUpgrader, time.Second)
		recorder = httptest.NewRecorder()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("when the request is invalid", func() {
		It("should return a successful response", func() {
			req = generateRequest(InvalidJustPod)
			handler.Handle(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return a successful response", func() {
			req = generateRequest(InvalidNoQuery)
			handler.Handle(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("when the workload is not found in the context", func() {
		It("should return an internal server error", func() {
			req := httptest.NewRequest(http.MethodGet, ValidPFWithPort, nil)

			handler.Handle(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Context("when creating the Options struct fails", func() {
		BeforeEach(func() {
			req = httptest.NewRequest("GET", "/port-forward?port=notport", nil)
			ctx := workloadDomain.SetIntoContext(req.Context(), &workloadDomain.Workload{})
			req = req.WithContext(ctx)
		})

		It("should return a bad request error", func() {
			handler.Handle(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("when the GetConnection method returns an error", func() {
		BeforeEach(func() {
			mockService.EXPECT().GetConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("portforward reader error"))
		})

		It("should return an internal server error", func() {
			req = generateRequest(ValidPFWithPort)

			handler.Handle(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
		})
	})

	Context("when the TunnelConnection method returns an error", func() {
		BeforeEach(func() {
			mockService.EXPECT().GetConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&MockConnection{fmt.Errorf("bla")}, nil)
			mockUpgrader.EXPECT().Upgrade(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks2.NewMockWebSocket(), nil)
		})

		It("should return an internal server error", func() {
			req = generateRequest(ValidPFWithPort)

			handler.Handle(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
		})
	})

	Context("when upgrading to websocket connection fails", func() {
		BeforeEach(func() {
			mockService.EXPECT().GetConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(&MockConnection{}, nil)
			mockUpgrader.EXPECT().Upgrade(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("upgrade error"))
		})

		It("should return an internal server error", func() {
			req = generateRequest(ValidPFWithPort)

			handler.Handle(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})

func generateRequest(path string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	ctx := workloadDomain.SetIntoContext(req.Context(), &workloadDomain.Workload{})
	req = req.WithContext(ctx)

	return req
}

type MockConnection struct {
	err error
}

func (mc *MockConnection) CreateStream(_ http.Header) (httpstream.Stream, error) {
	return nil, mc.err
}

func (mc *MockConnection) Close() error {
	return nil
}

func (mc *MockConnection) CloseChan() <-chan bool {
	var mock chan bool
	return mock
}

func (mc *MockConnection) SetIdleTimeout(_ time.Duration) {
}

func (mc *MockConnection) RemoveStreams(_ ...httpstream.Stream) {
}
