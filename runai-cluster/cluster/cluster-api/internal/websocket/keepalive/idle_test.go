/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package keepalive

import (
	"context"
	"log"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket"
)

var _ = Describe("Idle", func() {
	var (
		conn1 net.Conn
		conn2 net.Conn
		enc   *websocket.ExtendedNetConnection
	)

	BeforeEach(func() {
		conn1, conn2 = net.Pipe()
		enc = &websocket.ExtendedNetConnection{InnerConnection: conn1}
	})

	AfterEach(func() {
		_ = conn1.Close()
		_ = conn2.Close()
	})

	Context("check channel timeout", func() {
		It("Idle connection timeout reached", func() {
			ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*100)
			defer cancel()

			go func() {
				byteArr := make([]byte, 10)
				_, _ = conn2.Read(byteArr)
			}()

			i, err := NewIdle(enc, conn2, time.Millisecond, time.Millisecond)
			Expect(err).ToNot(HaveOccurred())

			select {
			case <-i.Channel():
				eErr := i.Execute()
				Expect(eErr).To(HaveOccurred())
			case <-ctx.Done():
				Expect(ctx.Err()).ToNot(HaveOccurred())
			}
		})

		It("Idle connection timeout missed", func() {
			i, err := NewIdle(enc, conn2, time.Millisecond*10, time.Millisecond*100)
			Expect(err).ToNot(HaveOccurred())
			ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond)
			defer cancel()

			go func() {
				byteArr := make([]byte, 10)
				_, _ = conn2.Read(byteArr)
			}()

			select {
			case <-i.Channel():
				err = i.Execute()
				Expect(err).To(HaveOccurred())
			case <-ctx.Done():
				Expect(ctx.Err()).To(HaveOccurred())
			}
		})

		It("Idle connection execute single time", func() {
			byteArr := make([]byte, 6)
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancel()

			i, err := NewIdle(enc, conn2, time.Millisecond, time.Millisecond*10)
			Expect(err).ToNot(HaveOccurred())

			Expect(conn2).ToNot(BeNil())
			go func() {
				_, wErr := conn2.Write([]byte("blabla"))
				if wErr != nil {
					log.Fatal(wErr)
				}
			}()
			go func() {
				data := make([]byte, 10)
				_, _ = conn2.Read(data)
			}()

			_, rErr := enc.Read(byteArr)
			if rErr != nil {
				log.Fatal(rErr)
			}
			Expect(string(byteArr)).To(Equal("blabla"))

			select {
			case <-i.Channel():
				err = i.Execute()
				Expect(err).ToNot(HaveOccurred())
			case <-ctx.Done():
				Expect(ctx.Err()).ToNot(HaveOccurred())
			}

			go func() {
				data := make([]byte, 10)
				_, _ = conn2.Read(data)
			}()
			time.Sleep(20 * time.Millisecond)

			select {
			case <-i.Channel():
				err = i.Execute()
				Expect(err).To(HaveOccurred())
			case <-ctx.Done():
				Expect(ctx.Err()).ToNot(HaveOccurred())
			}
		})
	})
})
