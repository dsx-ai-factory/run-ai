/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validators_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"k8s.io/utils/ptr"
)

var _ = Describe("ValidateClusterUUID", func() {
	var (
		ctrl        *gomock.Controller
		mockClient  *cpmocks.MockClusterClient
		txn         *transaction.Txn
		tmpDir      string
		logger      *zap.Logger
		logObserver *observer.ObservedLogs
	)

	const (
		testClusterUUID = "test-cluster-uuid-123"
		testClusterName = "test-cluster"
		testCpURL       = "https://test-cp.example.com"
		testToken       = "test-token"
		testDomain      = "test-domain.example.com"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClient = cpmocks.NewMockClusterClient(ctrl)

		// Set up mock cluster client
		cluster.SetClient(mockClient)

		// Create transaction
		txn = transaction.NewTxn()

		// Create temporary directory for config
		var err error
		tmpDir, err = os.MkdirTemp("", "clusteruuid-test")
		Expect(err).NotTo(HaveOccurred())

		// Create empty config file
		configPath := filepath.Join(tmpDir, "test_config.json")
		file, err := os.Create(configPath)
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()
		_, err = file.WriteString("{}")
		Expect(err).NotTo(HaveOccurred())

		// Set up logger with observer for testing log output - use DebugLevel
		observedZapCore, logs := observer.New(zap.DebugLevel)
		logger = zap.New(observedZapCore)
		logObserver = logs

		// Initialize config with test logger
		err = config.GetInstance().Initializing(config.Options{
			Debug:      true,
			ConfigPath: tmpDir,
			ConfigFile: "test_config.json",
		})
		Expect(err).NotTo(HaveOccurred())

		// Set the test logger
		config.GetInstance().SetLogger(logger)
	})

	AfterEach(func() {
		ctrl.Finish()
		_ = os.RemoveAll(tmpDir)
	})

	Context("when clusterUUID is empty", func() {
		It("should skip validation and return nil", func() {
			err := validators.ValidateClusterUUID(txn, "", testClusterName, testCpURL, testToken)

			Expect(err).To(BeNil())

			// Verify debug log was written - updated message
			debugLogs := logObserver.FilterMessage("clusterUUID is not set, skipping validation").All()
			Expect(debugLogs).To(HaveLen(1))
			Expect(debugLogs[0].Level).To(Equal(zap.DebugLevel))
		})
	})

	Context("when cluster retrieval fails", func() {
		BeforeEach(func() {
			mockClient.EXPECT().
				GetClusterByName(gomock.Any(), testClusterName).
				Return(nil, errors.New("network error")).
				Times(1)
		})

		It("should return an error with reconfigure message", func() {
			err := validators.ValidateClusterUUID(txn, testClusterUUID, testClusterName, testCpURL, testToken)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cluster 'test-cluster' does not exist or is misconfigured"))
			Expect(err.Error()).To(ContainSubstring("to fix this, run `runai cluster` to reconfigure it"))

			// Verify debug log was written - changed from Error to Debug level
			debugLogs := logObserver.FilterMessage("failed to get the cluster by name.").All()
			Expect(debugLogs).To(HaveLen(1))
			Expect(debugLogs[0].Level).To(Equal(zap.DebugLevel))
		})
	})

	Context("when cluster is not found (nil response)", func() {
		BeforeEach(func() {
			mockClient.EXPECT().
				GetClusterByName(gomock.Any(), testClusterName).
				Return(nil, fmt.Errorf("failed to find cluster name: %s", testClusterName)).
				Times(1)
		})

		It("should return an error with reconfigure message", func() {
			err := validators.ValidateClusterUUID(txn, testClusterUUID, testClusterName, testCpURL, testToken)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cluster 'test-cluster' does not exist or is misconfigured"))
			Expect(err.Error()).To(ContainSubstring("to fix this, run `runai cluster` to reconfigure it"))

			// Verify debug log was written - changed from Error to Debug level
			debugLogs := logObserver.FilterMessage("failed to get the cluster by name.").All()
			Expect(debugLogs).To(HaveLen(1))
			Expect(debugLogs[0].Level).To(Equal(zap.DebugLevel))
		})
	})

	Context("when cluster UUID does not match", func() {
		BeforeEach(func() {
			differentUUID := "different-uuid-456"
			mockCluster := &openapi.ClustersMinimalData{
				Uuid:   differentUUID,
				Name:   testClusterName,
				Domain: *openapi.NewNullableString(ptr.To(testDomain)),
			}

			mockClient.EXPECT().
				GetClusterByName(gomock.Any(), testClusterName).
				Return(mockCluster, nil).
				Times(1)
		})

		It("should return an error with UUID mismatch message", func() {
			err := validators.ValidateClusterUUID(txn, testClusterUUID, testClusterName, testCpURL, testToken)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("configured cluster 'test-cluster' has a UUID mismatch"))
			Expect(err.Error()).To(ContainSubstring("to fix this, run `runai cluster` to reconfigure it"))

			// Verify debug log was written - changed from Error to Debug level
			debugLogs := logObserver.FilterMessage("configured cluster UUID does not match the cluster UUID").All()
			Expect(debugLogs).To(HaveLen(1))
			Expect(debugLogs[0].Level).To(Equal(zap.DebugLevel))

			// Check log context fields
			logContext := debugLogs[0].Context
			Expect(logContext).To(ContainElement(zap.String("configuredClusterUUID", testClusterUUID)))
			Expect(logContext).To(ContainElement(zap.String("clusterUUID", "different-uuid-456")))
			Expect(logContext).To(ContainElement(zap.String("clusterName", testClusterName)))
		})
	})

	Context("when cluster UUID matches", func() {
		BeforeEach(func() {
			mockCluster := &openapi.ClustersMinimalData{
				Uuid:   testClusterUUID,
				Name:   testClusterName,
				Domain: *openapi.NewNullableString(ptr.To(testDomain)),
			}

			mockClient.EXPECT().
				GetClusterByName(gomock.Any(), testClusterName).
				Return(mockCluster, nil).
				Times(1)
		})

		It("should return nil (success)", func() {
			err := validators.ValidateClusterUUID(txn, testClusterUUID, testClusterName, testCpURL, testToken)

			Expect(err).To(BeNil())

			// Verify success debug log was written
			debugLogs := logObserver.FilterMessage("cluster UUID match cluster name").All()
			Expect(debugLogs).To(HaveLen(1))
			Expect(debugLogs[0].Level).To(Equal(zap.DebugLevel))

			// Check log context fields
			logContext := debugLogs[0].Context
			Expect(logContext).To(ContainElement(zap.String("clusterUUID", testClusterUUID)))
			Expect(logContext).To(ContainElement(zap.String("clusterName", testClusterName)))
		})
	})

	Context("edge cases", func() {
		Context("when clusterUUID is whitespace only", func() {
			BeforeEach(func() {
				// Whitespace is not considered empty by the function, so it will try to validate
				mockClient.EXPECT().
					GetClusterByName(gomock.Any(), testClusterName).
					Return(nil, errors.New("whitespace UUID")).
					Times(1)
			})

			It("should try to validate and return an error", func() {
				err := validators.ValidateClusterUUID(txn, "   ", testClusterName, testCpURL, testToken)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster 'test-cluster' does not exist or is misconfigured"))
			})
		})

		Context("when all parameters are empty", func() {
			It("should skip validation and return nil", func() {
				err := validators.ValidateClusterUUID(txn, "", "", "", "")

				Expect(err).To(BeNil())

				// Verify debug log was written - updated message
				debugLogs := logObserver.FilterMessage("clusterUUID is not set, skipping validation").All()
				Expect(debugLogs).To(HaveLen(1))
			})
		})

		Context("when cluster name is empty but UUID is provided", func() {
			BeforeEach(func() {
				mockClient.EXPECT().
					GetClusterByName(gomock.Any(), "").
					Return(nil, errors.New("empty cluster name")).
					Times(1)
			})

			It("should return an error", func() {
				err := validators.ValidateClusterUUID(txn, testClusterUUID, "", testCpURL, testToken)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster '' does not exist or is misconfigured"))
			})
		})
	})

	Context("logging verification", func() {
		It("should log with correct fields for successful validation", func() {
			mockCluster := &openapi.ClustersMinimalData{
				Uuid:   testClusterUUID,
				Name:   testClusterName,
				Domain: *openapi.NewNullableString(ptr.To(testDomain)),
			}

			mockClient.EXPECT().
				GetClusterByName(gomock.Any(), testClusterName).
				Return(mockCluster, nil).
				Times(1)

			err := validators.ValidateClusterUUID(txn, testClusterUUID, testClusterName, testCpURL, testToken)

			Expect(err).To(BeNil())

			// Verify all expected log fields are present
			debugLogs := logObserver.FilterMessage("cluster UUID match cluster name").All()
			Expect(debugLogs).To(HaveLen(1))

			logContext := debugLogs[0].Context
			Expect(logContext).To(HaveLen(2))
			Expect(logContext).To(ContainElement(zap.String("clusterUUID", testClusterUUID)))
			Expect(logContext).To(ContainElement(zap.String("clusterName", testClusterName)))
		})

		It("should log with correct fields for UUID mismatch", func() {
			differentUUID := "different-uuid"
			mockCluster := &openapi.ClustersMinimalData{
				Uuid:   differentUUID,
				Name:   testClusterName,
				Domain: *openapi.NewNullableString(ptr.To(testDomain)),
			}

			mockClient.EXPECT().
				GetClusterByName(gomock.Any(), testClusterName).
				Return(mockCluster, nil).
				Times(1)

			err := validators.ValidateClusterUUID(txn, testClusterUUID, testClusterName, testCpURL, testToken)

			Expect(err).To(HaveOccurred())

			// Verify all expected log fields are present
			debugLogs := logObserver.FilterMessage("configured cluster UUID does not match the cluster UUID").All()
			Expect(debugLogs).To(HaveLen(1))

			logContext := debugLogs[0].Context
			Expect(logContext).To(HaveLen(3))
			Expect(logContext).To(ContainElement(zap.String("configuredClusterUUID", testClusterUUID)))
			Expect(logContext).To(ContainElement(zap.String("clusterUUID", differentUUID)))
			Expect(logContext).To(ContainElement(zap.String("clusterName", testClusterName)))
		})
	})
})
