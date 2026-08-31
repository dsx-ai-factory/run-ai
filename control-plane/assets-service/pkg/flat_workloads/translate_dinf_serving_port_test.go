/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadsService "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var _ = Describe("TranslateToFlatWorkload", func() {
	var (
		ctx                     context.Context
		workloadType            string
		workloadAssets          []dbmodel.AssetsContainedInWorkload
		credentialsAssets       map[string]dbmodel.CredentialsAssetSpecs
		dataSourceOverridesById map[string]*api.DataSourceOverrides
		options                 *TranslateFlatWorkloadOptions
	)

	BeforeEach(func() {
		ctx = context.Background()
		workloadType = string(workloadsService.WORKLOADTYPE_DISTRIBUTED_INFERENCE)
		credentialsAssets = make(map[string]dbmodel.CredentialsAssetSpecs)
		dataSourceOverridesById = make(map[string]*api.DataSourceOverrides)
		options = &TranslateFlatWorkloadOptions{}
	})

	Context("when workload type is DistributedInference", func() {
		BeforeEach(func() {
			envAsset := &dbmodel.EnvironmentAssetRec{
				AssetId: "env-asset-id",
				EnvironmentAssetSpec: api.EnvironmentAssetSpec{
					Image:   convutils.Ref("my-image"),
					Command: convutils.Ref("python"),
					Args:    convutils.Ref("train.py"),
					EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
						{Name: convutils.Ref("MY_ENV"), Value: convutils.Ref("distributed")},
					},
					Connections: []api.Connection{
						{
							InternalToolInfo: &api.InternalToolInfo{
								ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
								ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
								ContainerPort:  8000,
								ServingPortInfo: &api.ServingPortInfo{
									Protocol: convutils.Ref("http"),
								},
							},
						},
					},
				},
			}

			workloadAssets = []dbmodel.AssetsContainedInWorkload{
				{
					AssetMetaRec: dbmodel.AssetMetaRec{
						AssetMeta: dbmodel.AssetMeta{
							Name: "my-env",
							Id:   "env-asset-id",
							Kind: api.ASSETKIND_ENVIRONMENT,
						},
					},
					AssetsEnvironment: envAsset,
				},
			}
		})

		It("should translate serving port correctly", func() {
			payloadMap, _, err := TranslateToFlatWorkload(
				ctx,
				workloadType,
				workloadAssets,
				credentialsAssets,
				nil,
				dataSourceOverridesById,
				options,
			)
			Expect(err).To(BeNil())
			Expect(payloadMap).NotTo(BeNil())

			Expect(payloadMap["image"]).To(Equal("my-image"))
			Expect(payloadMap["command"]).To(Equal("python"))
			Expect(payloadMap["args"]).To(Equal("train.py"))

			servingPortMap, exists := payloadMap["servingPort"]
			Expect(exists).To(BeTrue())

			spMap, ok := servingPortMap.(map[string]interface{})
			Expect(ok).To(BeTrue())

			containerVal, ok := spMap["container"]
			Expect(ok).To(BeTrue())
			Expect(convertToInt(containerVal)).To(Equal(8000))

			protocolVal, ok := spMap["protocol"]
			Expect(ok).To(BeTrue())
			Expect(protocolVal).To(Equal("http"))
		})
	})
})

func convertToInt(val interface{}) int {
	if num, ok := val.(json.Number); ok {
		n, _ := num.Int64()
		return int(n)
	}
	if f, ok := val.(float64); ok {
		return int(f)
	}
	return val.(int)
}
