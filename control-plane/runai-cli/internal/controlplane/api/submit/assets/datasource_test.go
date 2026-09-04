/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package assets

import (
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/fake"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datasource"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	datasourceTestFilePath = "datasource_asset_test"

	DatasourceAssetName = "my-pvc-datasource"
	DatasourceAssetId   = "datasource-uuid-123"
	DatasourceAssetType = "pvc"

	TrainingWithDatasourceName      = "tr-with-datasource"
	TrainingWithDatasourceImageName = "datasource-test-image"
)

var submitWithDatasourceList = []string{TrainingWithDatasourceName}

var _ = Describe("Datasource Asset Submit Tests", Ordered, func() {
	var (
		configPath           string
		err                  error
		envMap               map[string]string
		mockCtrl             *gomock.Controller
		mockDatasourceLister *cpmocks.MockDatasourceDescribeLister
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, datasourceTestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl = gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)

		// Mock datasource lister
		mockDatasourceLister = cpmocks.NewMockDatasourceDescribeLister(mockCtrl)
		datasource.SetClient(mockDatasourceLister)

		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Training{}, submitWithDatasourceList)
	})

	AfterAll(func() {
		datasource.SetClient(nil)
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

	Context("Submit with datasource asset", Ordered, func() {
		It("should resolve datasource asset name to ID and submit successfully", func() {
			// Mock datasource list response
			datasourceAsset := newTestDatasourceAsset(DatasourceAssetName, DatasourceAssetId)
			listResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{datasourceAsset},
			}
			mockDatasourceLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil).Times(1)

			commandStr := fmt.Sprintf("training submit %s -i %s -p p301 --datasource type=%s,name=%s",
				TrainingWithDatasourceName, TrainingWithDatasourceImageName, DatasourceAssetType, DatasourceAssetName)
			output, err := cli.RunCommand(commandStr, envMap)

			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf("Creating training %s", TrainingWithDatasourceName)))
		})

	})
})

func newTestDatasourceAsset(name, id string) openapi.DatasourceListResponseEntry {
	return openapi.DatasourceListResponseEntry{
		Meta: openapi.AssetMeta{
			Id:   id,
			Name: name,
		},
	}
}
