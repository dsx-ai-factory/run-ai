/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validation_test

import (
	"context"
	"fmt"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/k8s/validation"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
)

const (
	TestFilePath = "cluster_validator"
)

var _ = Describe("ClusterValidator", Ordered, func() {
	var (
		mockCtrl       *gomock.Controller
		mockKubeClient *mocks.MockKubeClienter
		validator      *validation.ClusterValidator
		ctx            context.Context
	)

	BeforeAll(func() {
		ctx = context.Background()
		// Create a minimal config file for testing
		err := os.WriteFile("./config.json", []byte("{}"), 0644)
		Expect(err).To(BeNil())
		err = config.GetInstance().Initializing(config.Options{Debug: true, ConfigPath: "./", ConfigFile: "config.json"})
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockKubeClient = mocks.NewMockKubeClienter(mockCtrl)
		validator = validation.NewClusterValidator(mockKubeClient)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	AfterAll(func() {
		os.Remove("./config.json")
	})

	Describe("NewClusterValidator", func() {
		It("should create a new ClusterValidator instance", func() {
			validator := validation.NewClusterValidator(mockKubeClient)
			Expect(validator).ToNot(BeNil())
		})
	})

	Describe("ValidateConnectivity", func() {
		Context("when kube client connectivity check succeeds", func() {
			It("should return no error", func() {
				mockKubeClient.EXPECT().CheckConnectivity(ctx).Return(nil)

				err := validator.ValidateConnectivity(ctx)
				Expect(err).To(BeNil())
			})
		})

		Context("when kube client connectivity check fails", func() {
			It("should return the connectivity error", func() {
				expectedErr := fmt.Errorf("connection failed")
				mockKubeClient.EXPECT().CheckConnectivity(ctx).Return(expectedErr)

				err := validator.ValidateConnectivity(ctx)
				Expect(err).To(Equal(expectedErr))
			})
		})
	})

	Describe("ValidateClusterIdentity", func() {
		Context("when cluster domain and server URL hosts match", func() {
			It("should return true with no error", func() {
				mockKubeClient.EXPECT().GetServerUrl(ctx).Return("http://localhost.runailabs.com:123")
				err := config.GetInstance().SetCluster("", "", "http://localhost.runailabs.com")
				fmt.Printf("err: %v\n", err)
				Expect(err).To(BeNil())

				isValid, err := validator.ValidateClusterIdentity(ctx)
				Expect(err).To(BeNil())
				Expect(isValid).To(BeTrue())
			})

			It("should handle URLs with different paths but same host", func() {
				mockKubeClient.EXPECT().GetServerUrl(ctx).Return("http://localhost.runailabs.com/different/path")
				err := config.GetInstance().SetCluster("", "", "http://localhost.runailabs.com/abc/def")
				fmt.Printf("err: %v\n", err)
				Expect(err).To(BeNil())

				isValid, err := validator.ValidateClusterIdentity(ctx)
				Expect(err).To(BeNil())
				Expect(isValid).To(BeTrue())
			})
		})

		Context("when cluster domain and server URL hosts do not match", func() {
			It("should return false with no error", func() {
				mockKubeClient.EXPECT().GetServerUrl(ctx).Return("http://different-host.com:123")
				config.GetInstance().SetCluster("", "", "http://localhost.runailabs.com")
				isValid, err := validator.ValidateClusterIdentity(ctx)
				Expect(err).To(BeNil())
				Expect(isValid).To(BeFalse())
			})
		})

		Context("when server URL is invalid", func() {
			It("should return error for malformed server URL", func() {
				mockKubeClient.EXPECT().GetServerUrl(ctx).Return("://invalid-server-url")

				isValid, err := validator.ValidateClusterIdentity(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse server url"))
				Expect(isValid).To(BeFalse())
			})
		})

		Context("when cluster domain is invalid", func() {
			It("should return error for malformed cluster domain", func() {
				config.GetInstance().SetCluster("", "", "://invalid-cluster-domain")
				isValid, err := validator.ValidateClusterIdentity(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse cluster domain"))
				Expect(isValid).To(BeFalse())
			})
		})

	})
})
