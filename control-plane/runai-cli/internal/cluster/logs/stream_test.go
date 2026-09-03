/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs_test

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/network"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
	clusterErrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/logs"
	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("WorkloadLogsStreamer", func() {
	var (
		ctrl             *gomock.Controller
		mockDialer       *mocks.MockDialer
		workloadStreamer *logs.WorkloadLogsStreamer
		ctx              context.Context
		workload         *apis.Workload
		options          *logs.Options
		maxRetryDuration time.Duration
		timeBetweenTries time.Duration

		expectedConn     network.Conn
		expectedResponse *http.Response
		errExpected      error
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDialer = mocks.NewMockDialer(ctrl)
		maxRetryDuration = 500 * time.Millisecond
		timeBetweenTries = 100 * time.Millisecond
		workloadStreamer = logs.NewWorkloadLogsStreamer(mockDialer, "https://example.com", maxRetryDuration, timeBetweenTries, nil)
		ctx = context.Background()
		workload = &apis.Workload{
			Project: "example-project",
			Type:    api.Training,
			Name:    "example-workload",
		}
		options = logs.NewOptions("example-pod", "example-container", "2023-05-29T12:00:00Z", 100, true, true, true, 1024, 0)
		expectedConn = mocks.NewMockConn(ctrl)
		expectedResponse = &http.Response{}
		errExpected = nil
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Stream", func() {
		It("returns the logs reader when streaming is successful", func() {
			mockDialer.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedConn, expectedResponse, errExpected)

			reader, err := workloadStreamer.Stream(ctx, workload, options)

			Expect(err).NotTo(HaveOccurred())
			Expect(reader).To(Equal(expectedConn))
		})

		It("returns an error when building the URL fails", func() {
			invalidOptions := &logs.Options{
				SinceTime: "invalid-time",
				Since:     300,
			}

			reader, err := workloadStreamer.Stream(ctx, workload, invalidOptions)

			Expect(err).To(HaveOccurred())
			Expect(reader).To(BeNil())
		})

		It("returns an error when streaming fails with a non-retryable error", func() {
			errExpected = errors.New("non-retryable error")
			mockDialer.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, errExpected)

			reader, err := workloadStreamer.Stream(ctx, workload, options)

			Expect(err.Error()).To(ContainSubstring(errExpected.Error()))
			Expect(reader).To(BeNil())
		})

		It("retries and returns an error when streaming fails with a retryable error and exceeds the maximum retry duration", func() {
			errExpected = clusterErrors.NewNotReadyError("")
			res := &http.Response{
				StatusCode:    http.StatusPreconditionFailed,
				Body:          nil,
				ContentLength: 1,
			}
			mockDialer.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, res, errExpected).AnyTimes()

			startTime := time.Now()
			reader, err := workloadStreamer.Stream(ctx, workload, options)
			duration := time.Since(startTime)

			Expect(err.Error()).To(ContainSubstring(errExpected.Error()))
			Expect(reader).To(BeNil())
			Expect(duration).To(BeNumerically(">=", maxRetryDuration))
		})
	})
})
