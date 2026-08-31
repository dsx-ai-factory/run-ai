/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	policyApi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"golang.org/x/exp/maps"
)

var _ = Describe("fields derivation test", func() {
	Context("fields derivation", func() {
		It("should produce correct listings", func() {
			rulesFromPayload := *policyApi.NewNullableWorkspacePolicyRulesV2(&policyApi.WorkspacePolicyRulesV2{
				Args: *policyApi.NewNullableStringRules(&policyApi.StringRules{
					Required: *policyApi.NewNullableBool(convutils.Ref(true)),
				}),
				EnvironmentVariables: *policyApi.NewNullableInstancesRules(&policyApi.InstancesRules{
					Instances: *policyApi.NewNullableItemRules(&policyApi.ItemRules{
						CanAdd: *policyApi.NewNullableBool(convutils.Ref(false)),
					}),
				}),
				Storage: *policyApi.NewNullableSupersetRulesAllOfStorage(&policyApi.SupersetRulesAllOfStorage{
					HostPath: *policyApi.NewNullableHostPathsRules(&policyApi.HostPathsRules{
						Attributes: *policyApi.NewNullableHostPathRules(&policyApi.HostPathRules{
							Path: *policyApi.NewNullableStringRules(&policyApi.StringRules{
								Required: *policyApi.NewNullableBool(convutils.Ref(true)),
							}),
							MountPath: *policyApi.NewNullableStringRules(&policyApi.StringRules{
								Required: *policyApi.NewNullableBool(convutils.Ref(true)),
							}),
						}),
						Instances: *policyApi.NewNullableItemRules(&policyApi.ItemRules{
							CanAdd: *policyApi.NewNullableBool(convutils.Ref(false)),
						}),
					}),
				}),
				Ports: *policyApi.NewNullablePortsRules(&policyApi.PortsRules{
					Attributes: *policyApi.NewNullablePortRules(&policyApi.PortRules{
						Container: *policyApi.NewNullableIntegerRules(&policyApi.IntegerRules{
							Min: *policyApi.NewNullableInt32(convutils.Ref(int32(8000))),
						}),
					}),
					Instances: *policyApi.NewNullableItemRules(&policyApi.ItemRules{
						CanAdd: *policyApi.NewNullableBool(convutils.Ref(false)),
					}),
				}),
			})

			rulesAsMap, err := gojsonq.LoadJson(rulesFromPayload)
			Expect(err).To(BeNil())

			//
			//    fields for "source of rules"
			//
			result := GetRulesForSourceOfRules(rulesAsMap)

			actualFieldsForSourceOfRules := datastructs.NewSet(maps.Keys(result)...)

			expectedFieldsForSourceOfRules := datastructs.NewSet(
				"environmentVariables",
				"ports",
				"ports.container",
				"storage.hostPath",
				"storage.hostPath.mountPath",
				"storage.hostPath.path",
				"args",
			)

			Expect(actualFieldsForSourceOfRules).To(Equal(expectedFieldsForSourceOfRules))

			//
			//    fields for "compliance"
			//
			result = GetRulesForCompliance(rulesAsMap)

			actualFieldsForCompliance := datastructs.NewSet(maps.Keys(result)...)

			expectedFieldsForCompliance := datastructs.NewSet(
				"environmentVariables",
				"ports",
				"storage.hostPath",
				"args",
			)

			Expect(actualFieldsForCompliance).To(Equal(expectedFieldsForCompliance))

			//
			//    fields with rules
			//
			fieldsWithRules := FieldsWithRules(rulesAsMap)

			actualFieldsWithRules := datastructs.NewSet(fieldsWithRules...)

			Expect(actualFieldsWithRules).To(Equal(expectedFieldsForSourceOfRules))
		})
	})
})
