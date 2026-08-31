/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	mocks2 "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/connection/mocks"

	mocksws "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/websocket"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/keepalive"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/proxy"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/logs"

	"github.com/golang/mock/gomock"
	ioMocks "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/io"
	mocks "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/logs"
	workloadDomain "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/workload"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	var (
		ctrl         *gomock.Controller
		mockService  *mocks.MockPodLogsService
		mockUpgrader *mocksws.MockUpgrader
		handler      *logs.Handler
		req          *http.Request
		recorder     *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockService = mocks.NewMockPodLogsService(ctrl)
		mockUpgrader = mocksws.NewMockUpgrader(ctrl)
		handler = logs.NewHandler(mockService, mockUpgrader, func(oneIO io.ReadCloser, otherIO io.ReadWriteCloser, keepalive keepalive.Keepalive) proxy.ProxyHandler {
			return &ioMocks.MockProxy{}
		}, time.Second)
		req = httptest.NewRequest(http.MethodGet, "/logs", nil)
		recorder = httptest.NewRecorder()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("when the request is valid", func() {
		BeforeEach(func() {
			ctx := workloadDomain.SetIntoContext(req.Context(), &workloadDomain.Workload{})
			req = req.WithContext(ctx)
		})

		It("should return a successful response", func() {
			mockService.EXPECT().LogsReader(gomock.Any(), gomock.Any(), logs.NewOptions()).Return(&ioMocks.MockReadCloser{}, nil)
			mockUpgrader.EXPECT().Upgrade(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks2.NewMockWebSocket(), nil)

			handler.Handle(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	})

	Context("when the workload is not found in the context", func() {
		It("should return an internal server error", func() {
			handler.Handle(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Context("when creating the Options struct fails", func() {
		BeforeEach(func() {
			req = httptest.NewRequest("GET", "/logs?follow=param", nil)
			ctx := workloadDomain.SetIntoContext(req.Context(), &workloadDomain.Workload{})
			req = req.WithContext(ctx)
		})

		It("should return a bad request error", func() {
			handler.Handle(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("when the LogsReader method returns an error", func() {
		BeforeEach(func() {
			ctx := workloadDomain.SetIntoContext(req.Context(), &workloadDomain.Workload{})
			req = req.WithContext(ctx)
			mockService.EXPECT().LogsReader(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("logs reader error"))
		})

		It("should return an internal server error", func() {
			handler.Handle(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
		})
	})

	Context("when upgrading to websocket connection fails", func() {
		BeforeEach(func() {
			ctx := workloadDomain.SetIntoContext(req.Context(), &workloadDomain.Workload{})
			req = req.WithContext(ctx)
			mockService.EXPECT().LogsReader(gomock.Any(), gomock.Any(), gomock.Any()).Return(&ioMocks.MockReadCloser{}, nil)
			mockUpgrader.EXPECT().Upgrade(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("upgrade error"))
		})

		It("should return an internal server error", func() {
			handler.Handle(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
