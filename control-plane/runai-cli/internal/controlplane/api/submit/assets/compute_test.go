/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package assets

import (
	"fmt"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	cpmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/compute"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

const (
	computeTestFilePath = "compute_asset_test"

	ComputeAssetName = "my-compute-resource"
	ComputeAssetId   = "compute-uuid-123"

	TrainingWithComputeName      = "tr-with-compute"
	TrainingWithComputeImageName = "compute-test-image"
)

var submitWithComputeList = []string{TrainingWithComputeName}

var _ = Describe("Compute Asset Submit Tests", Ordered, func() {
	var (
		configPath        string
		err               error
		envMap            map[string]string
		mockCtrl          *gomock.Controller
		mockComputeLister *cpmocks.MockComputeDescribeLister
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, computeTestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl = gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)

		// Mock compute lister
		mockComputeLister = cpmocks.NewMockComputeDescribeLister(mockCtrl)
		compute.SetClient(mockComputeLister)

		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Training{}, submitWithComputeList)
	})

	AfterAll(func() {
		compute.SetClient(nil)
	})

	BeforeEach(func() {
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("cluster set failed, error: " + err.Error())
		}
	})

	Context("Submit with compute asset", Ordered, func() {
		It("should resolve compute asset name to ID and submit successfully", func() {
			// Mock compute list response
			computeAsset := newTestComputeAsset(ComputeAssetName, ComputeAssetId)
			listResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{computeAsset},
			}
			mockComputeLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil).Times(1)

			commandStr := fmt.Sprintf("training submit %s -i %s -p p301 --compute %s",
				TrainingWithComputeName, TrainingWithComputeImageName, ComputeAssetName)
			output, err := cli.RunCommand(commandStr, envMap)

			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf("Creating training %s", TrainingWithComputeName)))
		})

	})
})

func newTestComputeAsset(name, id string) openapi.ComputeAsset {
	return openapi.ComputeAsset{
		Meta: openapi.AssetMeta{
			Id:   id,
			Name: name,
		},
	}
}
