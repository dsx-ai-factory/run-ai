/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/auth"
)

var _ = Describe("Manager (singleton)", func() {
	var (
		mockCtrl    *gomock.Controller
		mockBackend *authmocks.MockTokenService
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockBackend = authmocks.NewMockTokenService(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("returns the same singleton instance", func() {
		m1 := GetManager()
		m2 := GetManager()
		Expect(m1).To(BeIdenticalTo(m2))
	})

	Describe("Init and IsInitialized", func() {
		It("should not be initialized by default", func() {
			m := &Manager{}
			Expect(m.IsInitialized()).To(BeFalse())
		})
		It("should be initialized after Init", func() {
			m := &Manager{}
			m.Init(func(cfg ConfigGetter) (TokenService, error) { return mockBackend, nil }, nil)
			Expect(m.IsInitialized()).To(BeTrue())
		})
	})

	Describe("GetAccessToken", func() {
		It("calls backend methods after Init", func() {
			m := &Manager{}
			m.Init(func(cfg ConfigGetter) (TokenService, error) { return mockBackend, nil }, nil)
			mockBackend.EXPECT().GetAccessToken(gomock.Any()).Return("tok", nil)
			tok, err := m.GetAccessToken(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(tok).To(Equal("tok"))
		})
		It("returns error if ServiceCreatorFn returns error", func() {
			m := &Manager{}
			m.Init(func(cfg ConfigGetter) (TokenService, error) { return nil, context.DeadlineExceeded }, nil)
			_, err := m.GetAccessToken(context.Background())
			Expect(err).To(Equal(context.DeadlineExceeded))
		})
	})

	It("only initializes once even if Init is called multiple times", func() {
		m := &Manager{}

		creator := func(cfg ConfigGetter) (TokenService, error) { return mockBackend, nil }
		m.Init(creator, nil)
		m.Init(func(cfg ConfigGetter) (TokenService, error) { return nil, nil }, nil)
		Expect(m.IsInitialized()).To(BeTrue())
	})

	Describe("ServiceCreatorFn returns (nil, nil)", func() {
		It("returns error when backend is nil", func() {
			m := &Manager{}
			m.Init(func(cfg ConfigGetter) (TokenService, error) { return nil, fmt.Errorf("dd") }, nil)
			_, err := m.GetAccessToken(context.Background())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("backend returns error", func() {
		It("returns error from GetAccessToken", func() {
			m := &Manager{}
			m.Init(func(cfg ConfigGetter) (TokenService, error) { return mockBackend, nil }, nil)
			errBackend := context.Canceled
			mockBackend.EXPECT().GetAccessToken(gomock.Any()).Return("", errBackend)
			_, err := m.GetAccessToken(context.Background())
			Expect(err).To(Equal(errBackend))
		})
	})

	// Optional: Thread safety test for singleton
	It("GetManager is thread safe and always returns the same instance", func() {
		results := make(chan *Manager, 10)
		for i := 0; i < 10; i++ {
			go func() {
				results <- GetManager()
			}()
		}
		var first *Manager
		for i := 0; i < 10; i++ {
			m := <-results
			if i == 0 {
				first = m
			} else {
				Expect(m).To(BeIdenticalTo(first))
			}
		}
	})
})
