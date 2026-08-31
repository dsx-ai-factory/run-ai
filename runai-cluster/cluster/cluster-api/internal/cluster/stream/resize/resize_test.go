/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resize_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/stream"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/stream/resize"
	"k8s.io/client-go/tools/remotecommand"
)

func TestLogs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resize Suite")
}

var _ = Describe("Terminal Resize", func() {
	Describe("TerminalSizeQueue", func() {
		var queue *resize.TerminalSizeQueue

		BeforeEach(func() {
			queue = resize.NewTerminalSizeQueue()
		})

		AfterEach(func() {
			queue.Stop()
		})

		It("should handle resize events", func() {
			// Add a resize event
			queue.Resize(80, 24)

			// Get the next size
			size := queue.Next()
			Expect(size).NotTo(BeNil())
			Expect(size.Width).To(Equal(uint16(80)))
			Expect(size.Height).To(Equal(uint16(24)))
		})

		It("should return nil when stopped", func() {
			queue.Stop()
			size := queue.Next()
			Expect(size).To(BeNil())
		})

		It("should handle multiple resize events", func() {
			// Add multiple resize events
			queue.Resize(80, 24)
			queue.Resize(120, 30)

			// Get the first size
			size1 := queue.Next()
			Expect(size1).NotTo(BeNil())
			Expect(size1.Width).To(Equal(uint16(80)))
			Expect(size1.Height).To(Equal(uint16(24)))

			// Get the second size
			size2 := queue.Next()
			Expect(size2).NotTo(BeNil())
			Expect(size2.Width).To(Equal(uint16(120)))
			Expect(size2.Height).To(Equal(uint16(30)))
		})

		It("should handle resize message JSON", func() {
			resizeMsg := resize.TerminalSize{
				Width:  100,
				Height: 50,
			}

			data, err := json.Marshal(resizeMsg)
			Expect(err).NotTo(HaveOccurred())

			err = queue.HandleResizeMessage(data)
			Expect(err).NotTo(HaveOccurred())

			// Get the size
			size := queue.Next()
			Expect(size).NotTo(BeNil())
			Expect(size.Width).To(Equal(uint16(100)))
			Expect(size.Height).To(Equal(uint16(50)))
		})

		It("should handle invalid JSON gracefully", func() {
			invalidJSON := []byte(`{"invalid": "json"`)
			err := queue.HandleResizeMessage(invalidJSON)
			Expect(err).To(HaveOccurred())
		})

		It("should drop old events when buffer is full", func() {
			// Fill the buffer beyond capacity
			for i := 0; i < 15; i++ {
				queue.Resize(uint16(80+i), uint16(24+i))
			}

			// The queue should still work and return the most recent events
			var sizes []*remotecommand.TerminalSize
			for i := 0; i < 10; i++ {
				select {
				case size := <-make(chan *remotecommand.TerminalSize):
					sizes = append(sizes, size)
				default:
					// Use Next() to get the size
					size := queue.Next()
					if size != nil {
						sizes = append(sizes, size)
					}
				}
			}

			// Should have received some events
			Expect(len(sizes)).To(BeNumerically(">", 0))
			// the first 5 are dropped
			Expect(sizes[0].Width).To(Equal(uint16(85)))
			Expect(sizes[0].Height).To(Equal(uint16(29)))
			Expect(sizes[9].Width).To(Equal(uint16(94)))
			Expect(sizes[9].Height).To(Equal(uint16(38)))
		})
	})

	Describe("Handler with Terminal Resize", func() {
		It("should create terminal size queue for TTY interactive sessions", func() {
			options := stream.Options{
				Stdin:  true,
				Stdout: true,
				Stderr: true,
				TTY:    true,
			}

			// Create a mock websocket connection
			mockConn := &stream.MockWebSocket{}
			handler := stream.NewHandler(mockConn, options)

			streamOptions := handler.GetStreamOptions()
			Expect(streamOptions.TerminalSizeQueue).NotTo(BeNil())
			Expect(streamOptions.Tty).To(BeTrue())

			handler.Close()
		})

		It("should not create terminal size queue for non-TTY sessions", func() {
			options := stream.Options{
				Stdin:  true,
				Stdout: true,
				Stderr: true,
				TTY:    false,
			}

			// Create a mock websocket connection
			mockConn := &stream.MockWebSocket{}
			handler := stream.NewHandler(mockConn, options)

			streamOptions := handler.GetStreamOptions()
			Expect(streamOptions.TerminalSizeQueue).To(BeNil())

			handler.Close()
		})
	})
})
