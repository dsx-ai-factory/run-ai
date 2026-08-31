/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"encoding/json"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/scopes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

type EnvVar struct {
	Name  string
	Value string
}

type ExtendedResource struct {
	Resource string
	Quantity string
}

type UnsupportedStruct struct {
	Id    string
	Value string
}

var _ = Describe("merging test", func() {
	Context("itemized merging", func() {
		It("should not merge identical keys", func() {
			departmentPolicy := openapi.SupersetDefaults{
				Image: *openapi.NewNullableString(convutils.Ref("image-dept")),
				EnvironmentVariables: *openapi.NewNullableEnvironmentVariablesDefaults(&openapi.EnvironmentVariablesDefaults{
					Instances: []openapi.EnvironmentVariable{
						{
							Name:  *openapi.NewNullableString(convutils.Ref("var1")),
							Value: *openapi.NewNullableString(convutils.Ref("val1-dept")),
						},
						{
							Name:  *openapi.NewNullableString(convutils.Ref("var2")),
							Value: *openapi.NewNullableString(convutils.Ref("val2-dept")),
						},
					},
				}),
				Security: *openapi.NewNullableSupersetDefaultsAllOfSecurity(&openapi.SupersetDefaultsAllOfSecurity{
					RunAsUid: *openapi.NewNullableInt64(convutils.Ref(int64(123))),
				}),
				Compute: *openapi.NewNullableSupersetDefaultsAllOfCompute(&openapi.SupersetDefaultsAllOfCompute{
					ExtendedResources: *openapi.NewNullableExtendedResourcesDefaults(&openapi.ExtendedResourcesDefaults{
						Instances: []openapi.ExtendedResource{
							{
								Resource: *openapi.NewNullableString(convutils.Ref("a")),
								Quantity: *openapi.NewNullableString(convutils.Ref("10M")),
							},
							{
								Resource: *openapi.NewNullableString(convutils.Ref("b")),
								Quantity: *openapi.NewNullableString(convutils.Ref("20M")),
							},
						},
					}),
				}),
				NodePools: []string{
					"dept-np-1",
					"dept-np-2",
				},
			}

			projectPolicy := openapi.SupersetDefaults{
				Image: *openapi.NewNullableString(convutils.Ref("image-proj")),
				EnvironmentVariables: *openapi.NewNullableEnvironmentVariablesDefaults(&openapi.EnvironmentVariablesDefaults{
					Instances: []openapi.EnvironmentVariable{
						{
							Name:  *openapi.NewNullableString(convutils.Ref("var1")),
							Value: *openapi.NewNullableString(convutils.Ref("val1-proj")),
						},
						{
							Name:  *openapi.NewNullableString(convutils.Ref("var3")),
							Value: *openapi.NewNullableString(convutils.Ref("val3-proj")),
						},
					},
				}),
				Compute: *openapi.NewNullableSupersetDefaultsAllOfCompute(&openapi.SupersetDefaultsAllOfCompute{
					ExtendedResources: *openapi.NewNullableExtendedResourcesDefaults(&openapi.ExtendedResourcesDefaults{
						Instances: []openapi.ExtendedResource{
							{
								Resource: *openapi.NewNullableString(convutils.Ref("a")),
								Quantity: *openapi.NewNullableString(convutils.Ref("30M")),
							},
							{
								Resource: *openapi.NewNullableString(convutils.Ref("c")),
								Quantity: *openapi.NewNullableString(convutils.Ref("40M")),
							},
						},
					}),
				}),
				NodePools: []string{
					"proj-np-1",
					"proj-np-2",
				},
			}

			expected := openapi.SupersetDefaults{
				Image: *openapi.NewNullableString(convutils.Ref("image-proj")),
				EnvironmentVariables: *openapi.NewNullableEnvironmentVariablesDefaults(&openapi.EnvironmentVariablesDefaults{
					Instances: []openapi.EnvironmentVariable{
						{
							Name:  *openapi.NewNullableString(convutils.Ref("var1")),
							Value: *openapi.NewNullableString(convutils.Ref("val1-proj")),
						},
						{
							Name:  *openapi.NewNullableString(convutils.Ref("var2")),
							Value: *openapi.NewNullableString(convutils.Ref("val2-dept")),
						},
						{
							Name:  *openapi.NewNullableString(convutils.Ref("var3")),
							Value: *openapi.NewNullableString(convutils.Ref("val3-proj")),
						},
					},
				}),
				Compute: *openapi.NewNullableSupersetDefaultsAllOfCompute(&openapi.SupersetDefaultsAllOfCompute{
					ExtendedResources: *openapi.NewNullableExtendedResourcesDefaults(&openapi.ExtendedResourcesDefaults{
						Instances: []openapi.ExtendedResource{
							{
								Resource: *openapi.NewNullableString(convutils.Ref("a")),
								Quantity: *openapi.NewNullableString(convutils.Ref("30M")),
							},
							{
								Resource: *openapi.NewNullableString(convutils.Ref("b")),
								Quantity: *openapi.NewNullableString(convutils.Ref("20M")),
							},
							{
								Resource: *openapi.NewNullableString(convutils.Ref("c")),
								Quantity: *openapi.NewNullableString(convutils.Ref("40M")),
							},
						},
					}),
				}),
				NodePools: []string{
					"proj-np-1",
					"proj-np-2",
				},
				Security: *openapi.NewNullableSupersetDefaultsAllOfSecurity(&openapi.SupersetDefaultsAllOfSecurity{
					RunAsUid: *openapi.NewNullableInt64(convutils.Ref(int64(123))),
				}),
			}

			departmentMap, err := gojsonq.LoadJson(departmentPolicy)
			Expect(err).To(BeNil())

			projectMap, err := gojsonq.LoadJson(projectPolicy)
			Expect(err).To(BeNil())

			departmentJsonMap := dao_utils.JsonBMap(departmentMap)
			projectJsonMap := dao_utils.JsonBMap(projectMap)

			effectiveDept := dbmodel.EffectivePolicy{
				DefaultsV2: departmentJsonMap,
			}
			effectiveProj := dbmodel.Policy{
				DefaultsV2: projectJsonMap,
			}

			orgUnitTenant, _ := scopes.MakeOrgUnit(api.SCOPETYPE_TENANT, 1, nil, nil, nil)

			ht := HierarchyTree{}
			effectiveDept.DefaultsV2, err = ht.MergeDefaults(effectiveDept.DefaultsV2, effectiveProj.DefaultsV2, effectiveDept.RulesV2, orgUnitTenant, MergeDefaultsOptions{})
			Expect(err).To(BeNil())

			bytes, err := json.Marshal(departmentMap)
			Expect(err).To(BeNil())

			var result openapi.SupersetDefaults
			err = json.Unmarshal(bytes, &result)
			Expect(result).To(Equal(expected))
		})
	})

	Context("empty target", func() {
		It("should get everything from the source", func() {
			emptyPolicy := openapi.SupersetDefaults{
				EnvironmentVariables: *openapi.NewNullableEnvironmentVariablesDefaults(&openapi.EnvironmentVariablesDefaults{}),
				NodePools:            []string{},
			}

			projectPolicy := openapi.SupersetDefaults{
				EnvironmentVariables: *openapi.NewNullableEnvironmentVariablesDefaults(&openapi.EnvironmentVariablesDefaults{
					Instances: []openapi.EnvironmentVariable{
						{
							Name:  *openapi.NewNullableString(convutils.Ref("var1")),
							Value: *openapi.NewNullableString(convutils.Ref("val1-proj")),
						},
						{
							Name:  *openapi.NewNullableString(convutils.Ref("var3")),
							Value: *openapi.NewNullableString(convutils.Ref("val3-proj")),
						},
					},
				}),
				NodePools: []string{
					"proj-np-1",
					"proj-np-2",
				},
			}

			emptyMap, err := gojsonq.LoadJson(emptyPolicy)
			Expect(err).To(BeNil())

			projectMap, err := gojsonq.LoadJson(projectPolicy)
			Expect(err).To(BeNil())

			emptyJsonMap := dao_utils.JsonBMap(emptyMap)
			projectJsonMap := dao_utils.JsonBMap(projectMap)

			emptyDept := dbmodel.EffectivePolicy{
				DefaultsV2: emptyJsonMap,
			}
			effectiveProj := dbmodel.Policy{
				DefaultsV2: projectJsonMap,
			}

			orgUnitTenant, _ := scopes.MakeOrgUnit(api.SCOPETYPE_TENANT, 1, nil, nil, nil)

			ht := HierarchyTree{}
			emptyDept.DefaultsV2, err = ht.MergeDefaults(emptyDept.DefaultsV2, effectiveProj.DefaultsV2, emptyDept.RulesV2, orgUnitTenant, MergeDefaultsOptions{})
			Expect(err).To(BeNil())

			bytes, err := json.Marshal(emptyJsonMap)
			Expect(err).To(BeNil())

			var result openapi.SupersetDefaults
			err = json.Unmarshal(bytes, &result)
			Expect(result).To(Equal(projectPolicy))
		})
	})
})
