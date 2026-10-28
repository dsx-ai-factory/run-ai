/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	workloadsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/workloads"
	ourConfig "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/config"
	sharedWorkloadsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type retryingWorkloadActionsClient struct {
	*workloadsClient.SimpleMock
	mu       sync.Mutex
	failures int
	requests []workloadsApi.CreateWorkloadActionsRequest
	tokens   []any
}

type capturingWorkloadActionRecorder struct {
	actions []workloadAction
}

func (r *capturingWorkloadActionRecorder) Record(_ context.Context, actions ...workloadAction) {
	r.actions = append(r.actions, actions...)
}

func (r *capturingWorkloadActionRecorder) reset() {
	r.actions = nil
}

func (r *capturingWorkloadActionRecorder) recorded() []workloadAction {
	return slices.Clone(r.actions)
}

func (c *retryingWorkloadActionsClient) CreateWorkloadActions(ctx context.Context, request workloadsApi.CreateWorkloadActionsRequest) *workloadsApi.Error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requests = append(c.requests, request)
	c.tokens = append(c.tokens, ctx.Value(workloadsApi.ContextAccessToken))
	if c.failures > 0 {
		c.failures--
		return workloadsApi.NewError(http.StatusPreconditionFailed, "Workload registration is pending; retry action delivery")
	}
	return nil
}

var _ = Describe("Workload action delivery", func() {
	It("retries pending registration with the same record IDs and token", func() {
		client := &retryingWorkloadActionsClient{SimpleMock: workloadsClient.NewSimpleMock(), failures: 1}
		recorder := &asyncWorkloadActionRecorder{
			client:     func() workloadsClient.ClientInterface { return client },
			newBackOff: func() backoff.BackOff { return backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Millisecond), 2) },
		}
		ctx := context.WithValue(context.Background(), middlewares.ContextJwtToken, "token")
		request := workloadsApi.CreateWorkloadActionsRequest{
			Actions: []workloadsApi.WorkloadActionRecordInput{{Id: "record-1"}, {Id: "record-2"}},
		}

		Expect(recorder.deliver(ctx, request)).To(BeNil())
		Expect(client.requests).To(Equal([]workloadsApi.CreateWorkloadActionsRequest{request, request}))
		Expect(client.tokens).To(ConsistOf("token", "token"))
	})

	// The action was accepted, so its history should still be recorded if the request is canceled.
	It("continues after request cancellation", func() {
		client := &retryingWorkloadActionsClient{SimpleMock: workloadsClient.NewSimpleMock()}
		recorder := &asyncWorkloadActionRecorder{
			client:     func() workloadsClient.ClientInterface { return client },
			newBackOff: func() backoff.BackOff { return backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Millisecond), 1) },
		}
		ctx, cancel := context.WithCancel(context.Background())
		recorder.Record(ctx, workloadAction{ClusterID: "cluster", WorkloadID: "workload", WorkloadName: "name", Action: workloadsApi.WORKLOADACTIONTYPE_CREATE, AcceptedAt: time.Now()})
		cancel()

		Eventually(func() int {
			client.mu.Lock()
			defer client.mu.Unlock()
			return len(client.requests)
		}).Should(Equal(1))
	})

	It("stops retrying the internal history API after reaching the retry limit", func() {
		// Retry saving the action history, without repeating the original workload operation.
		// Allow one retry and verify that processing stops after two failed attempts.
		client := &retryingWorkloadActionsClient{SimpleMock: workloadsClient.NewSimpleMock(), failures: 3}
		recorder := &asyncWorkloadActionRecorder{
			client:     func() workloadsClient.ClientInterface { return client },
			newBackOff: func() backoff.BackOff { return backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Millisecond), 1) },
		}

		apiErr := recorder.deliver(context.Background(), workloadsApi.CreateWorkloadActionsRequest{})

		Expect(apiErr).To(HaveOccurred())
		Expect(client.requests).To(HaveLen(2))
	})
	DescribeTable("bounds delivery by the retry budget", func(blockResponse bool) {
		originalConfig := *config.Get[*ourConfig.Server]()
		serverConfig := originalConfig
		serverConfig.RetryMaxIntervalSec = 1
		Expect(config.Set(&serverConfig)).To(Succeed())
		DeferCleanup(func() { Expect(config.Set(&originalConfig)).To(Succeed()) })

		release := make(chan struct{})
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			if blockResponse {
				select {
				case <-r.Context().Done():
				case <-release:
				}
				return
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		DeferCleanup(server.Close)
		DeferCleanup(func() { close(release) })
		recorder := &asyncWorkloadActionRecorder{
			client:     func() workloadsClient.ClientInterface { return sharedWorkloadsClient.NewWorkloadsClient(server.URL) },
			newBackOff: func() backoff.BackOff { return backoff.NewConstantBackOff(time.Minute) },
		}
		result := make(chan *workloadsApi.Error, 1)
		go func() { result <- recorder.deliver(context.Background(), workloadsApi.CreateWorkloadActionsRequest{}) }()

		var apiErr *workloadsApi.Error
		Eventually(result, 3*time.Second).Should(Receive(&apiErr))
		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Message).To(ContainSubstring("deadline exceeded"))
		Expect(requests.Load()).To(Equal(int32(1)))
	},
		Entry("while waiting for an HTTP response", true),
		Entry("while waiting to retry", false),
	)

})
