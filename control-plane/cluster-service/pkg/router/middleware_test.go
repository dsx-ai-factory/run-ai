/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("markNetworkTopologyIdSent", func() {
	serve := func(method, path, body string) (sent any, receivedBody string) {
		next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			sent = r.Context().Value(utils.NetworkTopologyIdSentKey)
			read, err := io.ReadAll(r.Body)
			Expect(err).To(BeNil())
			receivedBody = string(read)
		})
		markNetworkTopologyIdSent(next).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, path, strings.NewReader(body)))
		return sent, receivedBody
	}

	DescribeTable("marks whether a node pool PATCH sent networkTopologyId and keeps the body for the controller",
		func(body string, wantSent bool) {
			sent, receivedBody := serve(http.MethodPatch, "/api/v1/node-pools/100001", body)
			Expect(sent).To(Equal(wantSent))
			Expect(receivedBody).To(Equal(body))
		},
		Entry("field omitted", `{"description":"d"}`, false),
		Entry("field null", `{"networkTopologyId":null}`, true),
		Entry("field set", `{"networkTopologyId":"7daba6ee-2aa9-4ab3-b735-a3d4bcdff72f"}`, true),
	)

	It("leaves other requests unmarked", func() {
		sent, receivedBody := serve(http.MethodPut, "/api/v1/node-pools/100001", `{"networkTopologyId":null}`)
		Expect(sent).To(BeNil())
		Expect(receivedBody).To(Equal(`{"networkTopologyId":null}`))

		sent, _ = serve(http.MethodPatch, "/api/v1/clusters/abc/config", `{"networkTopologyId":null}`)
		Expect(sent).To(BeNil())
	})

	It("leaves a body that isn't a JSON object unmarked for the controller to reject", func() {
		sent, receivedBody := serve(http.MethodPatch, "/api/v1/node-pools/100001", `not json`)
		Expect(sent).To(BeNil())
		Expect(receivedBody).To(Equal(`not json`))
	})
})
