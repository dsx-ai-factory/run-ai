/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"io"

	io2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/io"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/keepalive"
	proxy2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/proxy"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logs", func() {
	Describe("ProxyFactory", func() {
		It("should create a new one way proxy", func() {
			proxyFactory := NewOneWayProxyFactory()
			reader, _ := io.Pipe()
			proxy := proxyFactory(reader, &io2.MockReadWriteCloser{}, &keepalive.NoKeepalive{})
			Expect(proxy).To(BeAssignableToTypeOf(&proxy2.OneWayProxy{}))
		})
	})
})
