/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/keepalive"

	mocks "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/io"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TwoWayProxy", func() {
	var (
		ctrl       *gomock.Controller
		connection *mocks.MockReadWriteCloser
		otherIO    *mocks.MockReadWriteCloser
		proxyOBj   *TwoWayProxy
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		connection = mocks.NewMockReadWriteCloser(ctrl)
		otherIO = mocks.NewMockReadWriteCloser(ctrl)
		proxyOBj = NewTwoWayProxy(connection, otherIO, &keepalive.NoKeepalive{})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Start", func() {
		Context("when the context is not canceled", func() {
			It("should copy data between the connection and otherIO", func() {
				connection.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (int, error) {
					copy(p, "Hello")
					return 5, io.EOF
				})
				otherIO.EXPECT().Write([]byte("Hello")).Return(5, nil)

				otherIO.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (int, error) {
					copy(p, "World")
					return 5, io.EOF
				})
				connection.EXPECT().Write([]byte("World")).Return(5, nil)
				connection.EXPECT().Close().Return(nil).AnyTimes()
				otherIO.EXPECT().Close().Return(nil).AnyTimes()
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				err := proxyOBj.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the context is canceled", func() {
			It("should return an error", func() {
				connection.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (int, error) {
					copy(p, "Hello")
					return 5, io.EOF
				})
				otherIO.EXPECT().Write([]byte("Hello")).Return(5, nil)

				otherIO.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (int, error) {
					copy(p, "World")
					return 5, io.EOF
				})
				connection.EXPECT().Write([]byte("World")).Return(5, nil)
				connection.EXPECT().Close().Return(nil).AnyTimes()
				otherIO.EXPECT().Close().Return(nil).AnyTimes()
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				err := proxyOBj.Start(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("context canceled"))
			})
		})

		Context("when the reader returns an error", func() {
			It("should return the error", func() {
				connection.EXPECT().Read(gomock.Any()).Return(0, errors.New("read error"))
				otherIO.EXPECT().Read(gomock.Any()).Return(0, errors.New("read error"))
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				connection.EXPECT().Close().Return(nil).AnyTimes()
				otherIO.EXPECT().Close().Return(nil).AnyTimes()
				err := proxyOBj.Start(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("read error"))
			})
		})

		Context("when the context times out", func() {
			It("should return a context deadline exceeded error", func() {
				connection.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (int, error) {
					time.Sleep(300 * time.Millisecond)
					return 0, io.EOF
				})
				otherIO.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (int, error) {
					time.Sleep(300 * time.Millisecond)
					return 0, io.EOF
				})
				connection.EXPECT().Close().Return(nil).AnyTimes()
				otherIO.EXPECT().Close().Return(nil).AnyTimes()

				ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
				defer cancel()

				err := proxyOBj.Start(ctx)
				Expect(err).To(Equal(context.DeadlineExceeded))
			})
		})
	})
	Context("proxy", func() {
		const defaultTimeShutdown = 300 * time.Millisecond
		Context("when the context is not canceled", func() {
			It("should copy data from the reader to the writer", func() {
				ctx := context.Background()
				simpleFrom := &SimpleReaderWriterCloser{
					reader: *strings.NewReader("Hello, World!"),
				}
				var to strings.Builder
				writerCloser := SimpleReaderWriterCloser{
					writer: &to,
				}

				err := proxy(ctx, simpleFrom, &writerCloser, &keepalive.NoKeepalive{}, defaultTimeShutdown)

				Expect(err).NotTo(HaveOccurred())
				Expect(to.String()).To(Equal("Hello, World!"))
			})
		})

		Context("when the context is canceled", func() {
			It("should return an error", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				simpleFrom := &SimpleReaderWriterCloser{
					reader: *strings.NewReader("Hello, World!"),
				}
				var to strings.Builder
				writerCloser := SimpleReaderWriterCloser{
					writer: &to,
				}

				err := proxy(ctx, simpleFrom, &writerCloser, &keepalive.NoKeepalive{}, defaultTimeShutdown)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the reader returns an error", func() {
			It("should return the error", func() {
				ctx := context.Background()
				from := &errorReader{}
				var to strings.Builder
				writerCloser := SimpleReaderWriterCloser{
					writer: &to,
				}

				err := proxy(ctx, from, &writerCloser, &keepalive.NoKeepalive{}, defaultTimeShutdown)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("read error"))
			})
		})

		Context("when the context times out", func() {
			It("should return a context deadline exceeded error", func() {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				from := &slowReader{delay: 200 * time.Millisecond}
				var to strings.Builder
				writerCloser := SimpleReaderWriterCloser{
					writer: &to,
				}

				err := proxy(ctx, from, &writerCloser, &keepalive.NoKeepalive{}, 100*time.Millisecond)

				Expect(err).To(Equal(context.DeadlineExceeded))
			})
		})
	})
})

type SimpleReaderWriterCloser struct {
	reader strings.Reader
	writer *strings.Builder
	closed bool
}

func (r *SimpleReaderWriterCloser) Close() error {
	r.closed = true
	return nil
}

func (r *SimpleReaderWriterCloser) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *SimpleReaderWriterCloser) Write(p []byte) (n int, err error) {
	return r.writer.Write(p)
}

// errorReader is a custom reader that always returns an error
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func (r *errorReader) Close() error {
	return nil
}

type slowReader struct {
	delay  time.Duration
	closed bool
}

func (r *slowReader) Close() error {
	r.closed = true
	return nil
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	time.Sleep(r.delay)
	return len(p), nil
}
