/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	mocks2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/connection/mocks"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/attach"
	mocks "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/attach"
	mocksws "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/websocket"

	workloadDomain "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	HeaderCommand = "Command"
)

var _ = Describe("Handler", func() {
	var (
		ctrl         *gomock.Controller
		mockService  *mocks.MockPodService
		mockUpgrader *mocksws.MockUpgrader
		handler      *attach.Handler
		req          *http.Request
		recorder     *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockService = mocks.NewMockPodService(ctrl)
		mockUpgrader = mocksws.NewMockUpgrader(ctrl)
		handler = attach.NewHandler(mockService, mockUpgrader, time.Second)
		req = httptest.NewRequest(http.MethodGet, "/attach", nil)
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
			mockService.EXPECT().GenerateRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			mockService.EXPECT().AttachRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockUpgrader.EXPECT().Upgrade(gomock.Any(), gomock.Any(), gomock.Any()).Return(mocks2.NewMockWebSocket(), nil)
			req.Header.Set(HeaderCommand, encodeCommandString([]string{"bla"}))

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

	Context("when the GenerateRequest method returns an error", func() {
		BeforeEach(func() {
			ctx := workloadDomain.SetIntoContext(req.Context(), &workloadDomain.Workload{})
			req = req.WithContext(ctx)
			mockService.EXPECT().GenerateRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf(""))
		})

		It("should return an internal server error", func() {
			req.Header.Set(HeaderCommand, encodeCommandString([]string{"bla"}))
			handler.Handle(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
		})
	})

	Context("when upgrading to websocket connection fails", func() {
		BeforeEach(func() {
			ctx := workloadDomain.SetIntoContext(req.Context(), &workloadDomain.Workload{})
			req = req.WithContext(ctx)
			mockService.EXPECT().GenerateRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			mockUpgrader.EXPECT().Upgrade(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("upgrade error"))
		})

		It("should return an internal server error", func() {
			req.Header.Set(HeaderCommand, encodeCommandString([]string{"bla"}))
			handler.Handle(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})

func encodeCommandString(cmd []string) string {
	strByte, err := json.Marshal(cmd)
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(strByte)
}
