/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"encoding/base64"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workload composer validations", func() {

	BeforeEach(func() {
	})

	Context("no annotations", func() {
		It("no complimentary info", func() {
			workload := workloadsApi.SupersetSpec{
				Image: *workloadsApi.NewNullableString(convutils.Ref("ubuntu")),
			}

			info, err := LoadComplimentaryAssetInfo(workload.Annotations, false)
			Expect(err).To(BeNil())
			Expect(info).To(BeNil())
		})
	})

	Context("annotations with no info (=non assets workload)", func() {
		It("no complimentary info", func() {
			workload := workloadsApi.SupersetSpec{
				Image: *workloadsApi.NewNullableString(convutils.Ref("ubuntu")),
				Annotations: []workloadsApi.Annotation{
					{
						Name:  *workloadsApi.NewNullableString(convutils.Ref("hello")),
						Value: *workloadsApi.NewNullableString(convutils.Ref("world")),
					},
				},
			}

			info, err := LoadComplimentaryAssetInfo(workload.Annotations, false)
			Expect(err).To(BeNil())
			Expect(info).To(BeNil())
		})
	})

	Context("annotations with non-encoded info", func() {
		It("no complimentary info", func() {
			workload := workloadsApi.SupersetSpec{
				Image: *workloadsApi.NewNullableString(convutils.Ref("ubuntu")),
				Annotations: []workloadsApi.Annotation{
					{
						Name:  *workloadsApi.NewNullableString(convutils.Ref("hello")),
						Value: *workloadsApi.NewNullableString(convutils.Ref("world")),
					},
					{
						Name:  *workloadsApi.NewNullableString(convutils.Ref("run.ai/assets-info")),
						Value: *workloadsApi.NewNullableString(convutils.Ref(`{ "environment" : { "name" : "env-name" } }`)),
					},
				},
			}

			info, err := LoadComplimentaryAssetInfo(workload.Annotations, false)
			Expect(err).To(BeNil())
			Expect(info).NotTo(BeNil())
			Expect(info.Environment.Name).To(Equal("env-name"))
		})
	})

	Context("annotations with malformed encoded info", func() {
		It("should error", func() {
			workload := workloadsApi.SupersetSpec{
				Image: *workloadsApi.NewNullableString(convutils.Ref("ubuntu")),
				Annotations: []workloadsApi.Annotation{
					{
						Name:  *workloadsApi.NewNullableString(convutils.Ref("run.ai/assets-info")),
						Value: *workloadsApi.NewNullableString(convutils.Ref("e[pjrf[pijef[prej")),
					},
				},
			}

			_, err := LoadComplimentaryAssetInfo(workload.Annotations, true)
			Expect(err.Message).To(Equal("failed to base64 decode run.ai/assets-info annotation"))
			Expect(err).NotTo(BeNil())
		})
	})

	Context("annotations with malformed encoded info", func() {
		It("should error", func() {
			encodedString := base64.StdEncoding.EncodeToString([]byte("malformed-data"))

			workload := workloadsApi.SupersetSpec{
				Image: *workloadsApi.NewNullableString(convutils.Ref("ubuntu")),
				Annotations: []workloadsApi.Annotation{
					{
						Name:  *workloadsApi.NewNullableString(convutils.Ref("run.ai/assets-info")),
						Value: *workloadsApi.NewNullableString(convutils.Ref(encodedString)),
					},
				},
			}

			_, err := LoadComplimentaryAssetInfo(workload.Annotations, true)
			Expect(err.Message).To(Equal("failed to parse run.ai/assets-info: malformed-data"))
			Expect(err).NotTo(BeNil())
		})
	})

	Context("annotations with encoded info", func() {
		It("no complimentary info", func() {
			expectedInfo := ComplimentaryAssetInfo{
				Environment: workloadsApi.Environment{
					Name: "my-env",
					Id:   "my-id",
					Connections: []workloadsApi.Connection{
						{
							Name:           "my-connection",
							ToolType:       "my-tool-type",
							ConnectionType: "my-conn-type",
							Url:            convutils.Ref("my-url"),
							AuthorizedUsers: []string{
								"my-user-1",
								"my-user-2",
							},
						},
						{
							Name:           "your-connection",
							ToolType:       "your-tool-type",
							ConnectionType: "your-conn-type",
							AuthorizedGroups: []string{
								"my-group-1",
								"my-group-2",
							},
						},
						{
							Name:           "open-connection",
							ToolType:       "your-tool-type",
							ConnectionType: "your-conn-type",
						},
					},
				},
				Datasources: MapAssetKindToResources{
					"my-storage": []workloadsApi.Datasource{
						{
							Name: "storage-name",
							Id:   "storage-id",
							Type: convutils.Ref("storage-type"),
						},
						{
							Name: "no-type-name",
							Id:   "no-type-id",
						},
					},
				},
			}

			expectedInfoJson, jsonErr := gojsonq.LoadJson(expectedInfo)
			Expect(jsonErr).To(BeNil())

			expectedInfoText := gojsonq.ToBytes(expectedInfoJson)

			encodedInfo := base64.StdEncoding.EncodeToString(expectedInfoText)

			workload := workloadsApi.SupersetSpec{
				Image: *workloadsApi.NewNullableString(convutils.Ref("ubuntu")),
				Annotations: []workloadsApi.Annotation{
					{
						Name:  *workloadsApi.NewNullableString(convutils.Ref("hello")),
						Value: *workloadsApi.NewNullableString(convutils.Ref("world")),
					},
					{
						Name:  *workloadsApi.NewNullableString(convutils.Ref("run.ai/assets-info")),
						Value: *workloadsApi.NewNullableString(&encodedInfo),
					},
				},
			}

			info, err := LoadComplimentaryAssetInfo(workload.Annotations, true)
			Expect(err).To(BeNil())
			Expect(info).NotTo(BeNil())
			Expect(*info).To(Equal(expectedInfo))
		})
	})

})
