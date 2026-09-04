/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/connection"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/connection/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/keepalive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler Test", func() {
	var (
		conn    connection.WebSocket
		options Options
	)

	BeforeEach(func() {
		conn = mocks.NewMockWebSocket()
	})

	Context("Handler", func() {
		It("empty options with connection type miss match - failed", func() {
			handler := NewHandler(conn, options)

			Expect(handler.GetType()).To(Equal(ReadOnly))
			_, err := handler.GetKeepAlive()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fmt.Errorf("idle handler require extended connection")))
		})

		It("empty options with connection type enc - success", func() {
			enc := websocket.ExtendedNetConnection{
				InnerConnection: conn,
				Cancel:          nil,
			}
			handler := NewHandler(&enc, options)

			Expect(handler.GetType()).To(Equal(ReadOnly))
			ka, err := handler.GetKeepAlive()
			Expect(err).ToNot(HaveOccurred())
			Expect(ka.GetType()).To(Equal(keepalive.IdleType))
		})

		It("options with stdin - success (interactive)", func() {
			options.Stdin = true
			handler := NewHandler(conn, options)

			Expect(handler.GetType()).To(Equal(Interactive))
			ka, err := handler.GetKeepAlive()
			Expect(err).ToNot(HaveOccurred())
			Expect(ka.GetType()).To(Equal(keepalive.PingType))
		})

		It("options with tty - success (interactive)", func() {
			options.TTY = true
			handler := NewHandler(conn, options)

			Expect(handler.GetType()).To(Equal(Interactive))
			ka, err := handler.GetKeepAlive()
			Expect(err).ToNot(HaveOccurred())
			Expect(ka.GetType()).To(Equal(keepalive.PingType))
		})
	})
})
