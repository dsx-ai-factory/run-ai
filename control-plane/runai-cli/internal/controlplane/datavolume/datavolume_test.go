/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolume

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cpmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
)

func TestDatavolume(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Datavolume Suite")
}

var _ = Describe("GetNameToIdMap", func() {
	var (
		mockCtrl   *gomock.Controller
		mockLister *cpmocks.MockDatavolumeLister
		txn        *transaction.Txn
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockLister = cpmocks.NewMockDatavolumeLister(mockCtrl)
		SetClient(mockLister)
		txn = transaction.NewTxn()
	})

	AfterEach(func() {
		mockCtrl.Finish()
		SetClient(nil)
	})

	Describe("Empty names list", func() {
		It("should return nil map without calling API", func() {
			nameToIdMap, err := GetNameToIdMap(txn, "http://localhost", "token", "p1", []string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(nameToIdMap).To(BeNil())
		})
	})

	Describe("Single datavolume", func() {
		It("should resolve single datavolume ID when found", func() {
			listResponse := &openapi.GetDatavolumes200Response{
				Datavolumes: []openapi.DatavolumeNoSharedScopes{
					newTestDatavolume("dv1", "dv-uuid-1"),
				},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any()).Return(listResponse, nil)

			nameToIdMap, err := GetNameToIdMap(txn, "http://localhost", "token", "p1", []string{"dv1"})

			Expect(err).ToNot(HaveOccurred())
			Expect(nameToIdMap["dv1"]).To(Equal("dv-uuid-1"))
			Expect(len(nameToIdMap)).To(Equal(1))
		})
	})

	Describe("Multiple datavolumes across hierarchy", func() {
		It("should resolve datavolumes from project, department, and cluster scopes", func() {
			listResponse := &openapi.GetDatavolumes200Response{
				Datavolumes: []openapi.DatavolumeNoSharedScopes{
					newTestDatavolume("dv-project", "dv-project-id"),
					newTestDatavolume("dv-department", "dv-dept-id"),
					newTestDatavolume("dv-cluster", "dv-cluster-id"),
				},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any()).Return(listResponse, nil)

			nameToIdMap, err := GetNameToIdMap(txn, "http://localhost", "token", "p1", []string{
				"dv-project",
				"dv-department",
				"dv-cluster",
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(nameToIdMap["dv-project"]).To(Equal("dv-project-id"))
			Expect(nameToIdMap["dv-department"]).To(Equal("dv-dept-id"))
			Expect(nameToIdMap["dv-cluster"]).To(Equal("dv-cluster-id"))
			Expect(len(nameToIdMap)).To(Equal(3))
		})

		It("should handle mixed existing and non-existing datavolumes", func() {
			listResponse := &openapi.GetDatavolumes200Response{
				Datavolumes: []openapi.DatavolumeNoSharedScopes{
					newTestDatavolume("dv1", "dv-uuid-1"),
					newTestDatavolume("dv2", "dv-uuid-2"),
				},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any()).Return(listResponse, nil)

			nameToIdMap, err := GetNameToIdMap(txn, "http://localhost", "token", "p1", []string{"dv1", "dv2", "dv-missing"})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("dv-missing"))
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(nameToIdMap).To(BeNil())
		})
	})

	Describe("Datavolume not found", func() {
		It("should return error when datavolume not found", func() {
			listResponse := &openapi.GetDatavolumes200Response{
				Datavolumes: []openapi.DatavolumeNoSharedScopes{},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any()).Return(listResponse, nil)

			nameToIdMap, err := GetNameToIdMap(txn, "http://localhost", "token", "p1", []string{"missing"})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(nameToIdMap).To(BeNil())
		})

		It("should return error for specific missing datavolume among many", func() {
			listResponse := &openapi.GetDatavolumes200Response{
				Datavolumes: []openapi.DatavolumeNoSharedScopes{
					newTestDatavolume("dv1", "dv-uuid-1"),
				},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any()).Return(listResponse, nil)

			nameToIdMap, err := GetNameToIdMap(txn, "http://localhost", "token", "p1", []string{"dv1", "dv-nonexistent"})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("dv-nonexistent"))
			Expect(nameToIdMap).To(BeNil())
		})
	})

	Describe("API error", func() {
		It("should return error when API fails", func() {
			mockLister.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))

			nameToIdMap, err := GetNameToIdMap(txn, "http://localhost", "token", "p1", []string{"dv1"})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to list datavolumes"))
			Expect(nameToIdMap).To(BeNil())
		})
	})

	Describe("Duplicate names handling", func() {
		It("should handle duplicate names in input list", func() {
			listResponse := &openapi.GetDatavolumes200Response{
				Datavolumes: []openapi.DatavolumeNoSharedScopes{
					newTestDatavolume("dv1", "dv-uuid-1"),
				},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any()).Return(listResponse, nil)

			nameToIdMap, err := GetNameToIdMap(txn, "http://localhost", "token", "p1", []string{"dv1", "dv1"})

			Expect(err).ToNot(HaveOccurred())
			Expect(nameToIdMap["dv1"]).To(Equal("dv-uuid-1"))
			Expect(len(nameToIdMap)).To(Equal(1))
		})
	})
})

func newTestDatavolume(name, id string) openapi.DatavolumeNoSharedScopes {
	return openapi.DatavolumeNoSharedScopes{
		Id:            id,
		Name:          name,
		OriginPvcName: "pvc-test",
		ProjectId:     "p1",
		ProjectName:   "p1",
		Status: openapi.DatavolumeStatus{
			Phase: openapi.DATAVOLUMEPHASE_READY,
		},
	}
}
