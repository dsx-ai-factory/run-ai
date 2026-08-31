/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package environment

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cpmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

func TestEnvironment(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Environment Suite")
}

var _ = Describe("GetId", func() {
	var (
		mockCtrl   *gomock.Controller
		mockLister *cpmocks.MockEnvironmentDescribeLister
		txn        *transaction.Txn
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockLister = cpmocks.NewMockEnvironmentDescribeLister(mockCtrl)
		SetClient(mockLister)
		txn = transaction.NewTxn()
	})

	AfterEach(func() {
		mockCtrl.Finish()
		SetClient(nil)
	})

	Describe("Empty environment name", func() {
		It("should return empty string without calling API", func() {
			id, err := GetId(txn, "http://localhost", "token", "", "cluster-1", "123", spectype.Workspace)
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(BeEmpty())
		})
	})

	Describe("Environment found by name", func() {
		It("should return environment ID when found", func() {
			env := newTestEnvironment("my-env", "env-uuid-123")
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{env},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "my-env", "cluster-1", "123", spectype.Workspace)

			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal("env-uuid-123"))
		})

		It("should work with cluster filter", func() {
			env := newTestEnvironment("cluster-env", "env-uuid-456")
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{env},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "cluster-env", "cluster-1", "", spectype.Training)

			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal("env-uuid-456"))
		})
	})

	Describe("Environment not found (describe - no workload type)", func() {
		It("should return 'not found' error when environment doesn't exist", func() {
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			// Empty workload type simulates describe command
			id, err := GetId(txn, "http://localhost", "token", "nonexistent", "cluster-1", "123", "")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("environment 'nonexistent' not found"))
			Expect(id).To(BeEmpty())
		})
	})

	Describe("Environment does not support workload type (submit)", func() {
		It("should return 'does not support' error for workspace", func() {
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "training-only-env", "cluster-1", "123", spectype.Workspace)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the environment 'training-only-env' does not support workspace workloads"))
			Expect(id).To(BeEmpty())
		})

		It("should return 'does not support' error for training", func() {
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "workspace-only-env", "cluster-1", "123", spectype.Training)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the environment 'workspace-only-env' does not support training workloads"))
			Expect(id).To(BeEmpty())
		})

		It("should return 'does not support' error for distributed training", func() {
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "my-env", "cluster-1", "123", spectype.TrainingDistributed)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the environment 'my-env' does not support distributed training workloads"))
			Expect(id).To(BeEmpty())
		})
	})

	Describe("API error", func() {
		It("should return error when API fails", func() {
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))

			id, err := GetId(txn, "http://localhost", "token", "my-env", "cluster-1", "123", spectype.Workspace)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to search for environment"))
			Expect(id).To(BeEmpty())
		})
	})
})

func newTestEnvironment(name, id string) openapi.EnvironmentAsset {
	return openapi.EnvironmentAsset{
		Meta: openapi.AssetMeta{
			Id:   id,
			Name: name,
		},
	}
}
