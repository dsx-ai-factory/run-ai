/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	ws "nhooyr.io/websocket"

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

		It("uses separate ping and idle intervals", func() {
			cfg := GetInstance()
			DeferCleanup(cfg.SetPingTime, cfg.GetPingTime())
			DeferCleanup(cfg.SetIdleTime, cfg.GetIdleTime())
			// Allow the connection to stay idle longer than the ping interval.
			cfg.SetPingTime(10 * time.Millisecond)
			cfg.SetIdleTime(time.Second)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				peer, err := ws.Accept(w, r, nil)
				Expect(err).ToNot(HaveOccurred())
				// Keep the connection open without sending output.
				peer.CloseRead(ctx)
			}))
			defer server.Close()
			client, _, err := ws.Dial(ctx, server.URL, nil)
			Expect(err).ToNot(HaveOccurred())
			enc := websocket.NewConnection(client, ctx)
			defer func() { _ = enc.Close() }()

			ka, err := NewHandler(enc, Options{}).GetKeepAlive()
			Expect(err).ToNot(HaveOccurred())

			// A ping should not close the connection.
			Eventually(ka.Channel()).Should(Receive())
			Expect(ka.Execute()).To(Succeed())
			// The connection should close after the idle timeout.
			Eventually(ka.Execute, 2*time.Second, 10*time.Millisecond).Should(MatchError("idle reach"))
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
