/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"context"
	"encoding/json"

	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	cpmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Service")
}

var _ = Describe("Cluster Service Test", func() {
	var (
		ctrl   *gomock.Controller
		mocknc *cpmocks.MockClusterClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mocknc = cpmocks.NewMockClusterClient(ctrl)
		version := "cluster-one-version"
		domain := "cluster-one-domain"
		clusterType := "cluster-type"
		kubeVersion := "kube-version"
		state := openapi.CLUSTERSTATE_CONNECTED
		platform := openapi.ClusterReportedStatusPlatform{
			Type:        *openapi.NewNullableString(&clusterType),
			KubeVersion: *openapi.NewNullableString(&kubeVersion),
		}

		clusterList := []openapi.ClustersMinimalData{{
			Uuid:     "cluster-one-uuid",
			Name:     "cluster-one",
			Domain:   *openapi.NewNullableString(&domain),
			Version:  *openapi.NewNullableString(&version),
			State:    &state,
			Platform: *openapi.NewNullableClusterReportedStatusPlatform(&platform),
		}}
		mocknc.EXPECT().GetClusterList(gomock.Any()).Return(clusterList, nil).AnyTimes()
	})

	AfterEach(func() {
		// We add this in order to check that all the registered mocks ware really called
		ctrl.Finish()
	})

	Context("Service get cluster", func() {
		It("Check brief match", func() {
			bns := NewBackendClusterService(mocknc)
			handler, err := bns.GetClusterList(context.Background())
			Expect(err).To(BeNil())

			// Set raw flag to get original API response
			handler.SetRaw(true)
			response, err := handler.GenerateJson()
			Expect(err).To(BeNil())
			var data manipulate.Data
			err = json.Unmarshal([]byte(response), &data)
			Expect(err).To(BeNil())
			clusterList := data.Clusters
			Expect(clusterList[0].Uuid).To(Equal("cluster-one-uuid"))
			Expect(clusterList[0].Name).To(Equal("cluster-one"))
			Expect(clusterList[0].GetDomain()).To(Equal("cluster-one-domain"))
			Expect(clusterList[0].GetVersion()).To(Equal("cluster-one-version"))

			state := clusterList[0].GetState()
			platform := clusterList[0].GetPlatform()
			Expect(platform.GetType()).To(Equal("cluster-type"))
			Expect(platform.GetKubeVersion()).To(Equal("kube-version"))
			Expect(state).To(Equal(openapi.CLUSTERSTATE_CONNECTED))
		})
	})
})
