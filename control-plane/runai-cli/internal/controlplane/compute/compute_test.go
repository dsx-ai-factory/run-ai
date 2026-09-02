/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package compute

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
)

func TestCompute(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compute Suite")
}

var _ = Describe("GetId", func() {
	var (
		mockCtrl   *gomock.Controller
		mockLister *cpmocks.MockComputeDescribeLister
		txn        *transaction.Txn
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockLister = cpmocks.NewMockComputeDescribeLister(mockCtrl)
		SetClient(mockLister)
		txn = transaction.NewTxn()
	})

	AfterEach(func() {
		mockCtrl.Finish()
		SetClient(nil)
	})

	Describe("Compute found by name", func() {
		It("should return compute ID when found", func() {
			compute := newTestCompute("my-compute", "compute-uuid-123")
			listResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{compute},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "my-compute", "cluster-1", "123")

			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal("compute-uuid-123"))
		})

		It("should work with cluster filter", func() {
			compute := newTestCompute("cluster-compute", "compute-uuid-456")
			listResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{compute},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Nil(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "cluster-compute", "cluster-1", "")

			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal("compute-uuid-456"))
		})
	})

	Describe("Compute not found", func() {
		It("should return error when compute not found", func() {
			listResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "nonexistent", "cluster-1", "123")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(id).To(BeEmpty())
		})
	})

	Describe("API error", func() {
		It("should return error when API fails", func() {
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))

			id, err := GetId(txn, "http://localhost", "token", "my-compute", "cluster-1", "123")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to search for compute"))
			Expect(id).To(BeEmpty())
		})
	})
})

func newTestCompute(name, id string) openapi.ComputeAsset {
	return openapi.ComputeAsset{
		Meta: openapi.AssetMeta{
			Id:   id,
			Name: name,
		},
	}
}
