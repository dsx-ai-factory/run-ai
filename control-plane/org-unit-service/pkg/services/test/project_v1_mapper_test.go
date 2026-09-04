/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"strconv"
	"time"

	clusteropenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	v1 "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/v1"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Projects 1 mapper test", Ordered, func() {
	Context("test project mapper, cpu ft off, oq ft off", func() {
		cpuFt := false
		oqFt := false
		var override bool

		It("map model to response without jobs", func() {
			override = false
			project := createProject(override, cpuFt, oqFt)
			api, err := v1.MapProjectModelToV1Response(context.Background(), *project, nil, false, oqFt, cpuFt)
			Expect(err).To(BeNil())
			assertV1ApiProjectDefaults(api, override)
			AssertV1ApiResourcesProj(api.NodePoolsResources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			override = true
			project := createProject(override, cpuFt, oqFt)
			api, err := v1.MapProjectModelToV1Response(context.Background(), *project, nil, false, oqFt, cpuFt)
			Expect(err).To(BeNil())
			assertV1ApiProjectDefaults(api, override)
			AssertV1ApiResourcesProj(api.NodePoolsResources, oqFt, cpuFt, false)
		})
	})

	Context("test project mapper, cpu ft on, oq ft off", func() {
		cpuFt := true
		oqFt := false
		var override bool

		It("map model to response null limit", func() {
			override = false
			project := createProject(override, cpuFt, oqFt)
			project.OrgUnitData.Resources[0].CpuLimit = nil
			project.OrgUnitData.Resources[1].CpuLimit = nil
			project.OrgUnitData.Resources[0].MemoryLimit = nil
			project.OrgUnitData.Resources[1].MemoryLimit = nil
			api, err := v1.MapProjectModelToV1Response(context.Background(), *project, nil, false, oqFt, cpuFt)
			Expect(err).To(BeNil())
			assertV1ApiProjectDefaults(api, override)
			Expect(len(api.NodePoolsResources)).To(Equal(2))
			defaultResources := api.NodePoolsResources[0]
			validateV1ApiNodepoolResourcesUnlimitedProj(defaultResources, oqFt, cpuFt, float32(deserved), float32(limit), float32(oqWeight), false)
			poolResources := api.NodePoolsResources[1]
			validateV1ApiNodepoolResourcesUnlimitedProj(poolResources, oqFt, cpuFt, float32(deserved2), float32(limit2), float32(oqWeight2), false)
		})

		It("map model to response without jobs", func() {
			override = false
			project := createProject(override, cpuFt, oqFt)
			api, err := v1.MapProjectModelToV1Response(context.Background(), *project, nil, false, oqFt, true)
			Expect(err).To(BeNil())
			assertV1ApiProjectDefaults(api, override)
			AssertV1ApiResourcesProj(api.NodePoolsResources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			override = true
			project := createProject(override, cpuFt, oqFt)
			api, err := v1.MapProjectModelToV1Response(context.Background(), *project, nil, false, oqFt, true)
			Expect(err).To(BeNil())
			assertV1ApiProjectDefaults(api, override)
			AssertV1ApiResourcesProj(api.NodePoolsResources, oqFt, cpuFt, false)
		})
	})

	Context("test project mapper, cpu ft off, oq ft on", func() {
		cpuFt := false
		oqFt := true
		var override bool

		It("map model to response without jobs", func() {
			override = false
			project := createProject(override, cpuFt, oqFt)
			api, err := v1.MapProjectModelToV1Response(context.Background(), *project, nil, false, oqFt, cpuFt)
			Expect(err).To(BeNil())
			assertV1ApiProjectDefaults(api, override)
			AssertV1ApiResourcesProj(api.NodePoolsResources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			override = true
			project := createProject(override, cpuFt, oqFt)
			api, err := v1.MapProjectModelToV1Response(context.Background(), *project, nil, false, oqFt, cpuFt)
			Expect(err).To(BeNil())
			assertV1ApiProjectDefaults(api, override)
			AssertV1ApiResourcesProj(api.NodePoolsResources, oqFt, cpuFt, false)
		})
	})

	Context("test project mapper, cpu ft on, oq ft on", func() {
		cpuFt := true
		oqFt := true
		var override bool

		It("map model to response without jobs", func() {
			override = false
			project := createProject(override, cpuFt, oqFt)
			api, err := v1.MapProjectModelToV1Response(context.Background(), *project, nil, false, oqFt, true)
			Expect(err).To(BeNil())
			assertV1ApiProjectDefaults(api, override)
			AssertV1ApiResourcesProj(api.NodePoolsResources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			override = true
			project := createProject(override, cpuFt, oqFt)
			api, err := v1.MapProjectModelToV1Response(context.Background(), *project, nil, false, oqFt, true)
			Expect(err).To(BeNil())
			assertV1ApiProjectDefaults(api, override)
			AssertV1ApiResourcesProj(api.NodePoolsResources, oqFt, cpuFt, false)
		})
	})

	Context("test users sort", func() {
		It("test with !excludePermissions", func() {
			projId := int64(11)
			project := models.Project{
				IntBaseModel: models.IntBaseModel{
					ID:        projId,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				TenantId:      "1",
				ClusterId:     uuid.New(),
				Name:          "proj",
				ParentID:      ptr.To(int64(1)),
				OrgUnitDataID: uuid.New(),
				OrgUnitData: models.OrgUnitData{
					Resources: []models.Resources{{
						NodePoolId:         "1",
						NodePoolName:       "default",
						GpuDeserved:        ptr.To(float64(5)),
						GpuLimit:           ptr.To(float64(-1)),
						GpuOverQuotaWeight: ptr.To(float64(5)),
					}},
				},
			}
			accessRules := make(map[string]*[]openapi.AccessRule)
			accessRules["11"] = &[]openapi.AccessRule{{
				SubjectId:   "3",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_USER,
			}, {
				SubjectId:   "17",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_USER,
			}, {
				SubjectId:   "4",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_GROUP,
			}, {
				SubjectId:   "5",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_GROUP,
			}, {
				SubjectId:   "1",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_GROUP,
			}, {
				SubjectId:   "2",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_USER,
			}, {
				SubjectId:   "31",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_USER,
			}, {
				SubjectId:   "2",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_GROUP,
			}, {
				SubjectId:   "1",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_USER,
			}}
			projects, err := v1.MapProjectsModelToV1Response(context.Background(), []*models.Project{&project}, accessRules, true, map[string]clusteropenapi.NodepoolMinimal{"1": {Name: "default", Id: "1"}}, false)

			Expect(err).To(BeNil())
			Expect(len(projects)).To(Equal(1))

			projectV1 := projects[0]
			Expect(projectV1.Permissions.Users).NotTo(BeNil())
			Expect(len(*projectV1.Permissions.Users)).To(Equal(5))
			Expect((*projectV1.Permissions.Users)[0]).To(Equal("1"))
			Expect((*projectV1.Permissions.Users)[1]).To(Equal("17"))
			Expect((*projectV1.Permissions.Users)[2]).To(Equal("2"))
			Expect((*projectV1.Permissions.Users)[3]).To(Equal("3"))
			Expect((*projectV1.Permissions.Users)[4]).To(Equal("31"))
			Expect(projectV1.Permissions.Groups).NotTo(BeNil())
			Expect(len(*projectV1.Permissions.Groups)).To(Equal(4))
			Expect((*projectV1.Permissions.Groups)[0]).To(Equal("1"))
			Expect((*projectV1.Permissions.Groups)[1]).To(Equal("2"))
			Expect((*projectV1.Permissions.Groups)[2]).To(Equal("4"))
			Expect((*projectV1.Permissions.Groups)[3]).To(Equal("5"))
		})

		It("test with excludePermissions", func() {
			projId := int64(11)
			project := models.Project{
				IntBaseModel: models.IntBaseModel{
					ID:        projId,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				TenantId:      "1",
				ClusterId:     uuid.New(),
				Name:          "proj",
				ParentID:      ptr.To(int64(1)),
				OrgUnitDataID: uuid.New(),
				OrgUnitData: models.OrgUnitData{
					Resources: []models.Resources{{
						NodePoolId:         "1",
						NodePoolName:       "default",
						GpuDeserved:        ptr.To(float64(5)),
						GpuLimit:           ptr.To(float64(-1)),
						GpuOverQuotaWeight: ptr.To(float64(5)),
					}},
				},
			}
			accessRules := make(map[string]*[]openapi.AccessRule)
			accessRules["11"] = &[]openapi.AccessRule{{
				SubjectId:   "3",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_USER,
			}, {
				SubjectId:   "17",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_USER,
			}, {
				SubjectId:   "4",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_GROUP,
			}, {
				SubjectId:   "5",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_GROUP,
			}, {
				SubjectId:   "1",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_GROUP,
			}, {
				SubjectId:   "2",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_USER,
			}, {
				SubjectId:   "31",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_USER,
			}, {
				SubjectId:   "2",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_GROUP,
			}, {
				SubjectId:   "1",
				SubjectType: openapi.SUBJECTTYPEEXTENDED_USER,
			}}
			projects, err := v1.MapProjectsModelToV1Response(context.Background(), []*models.Project{&project}, accessRules, true, map[string]clusteropenapi.NodepoolMinimal{"1": {Name: "default", Id: "1"}}, true)

			Expect(err).To(BeNil())
			Expect(len(projects)).To(Equal(1))

			projectV1 := projects[0]
			Expect(projectV1.Permissions.Users).NotTo(BeNil())
			Expect(len(*projectV1.Permissions.Users)).To(Equal(0))
			Expect(projectV1.Permissions.Groups).NotTo(BeNil())
			Expect(len(*projectV1.Permissions.Groups)).To(Equal(0))
		})
	})
})

func assertV1ApiProjectDefaults(requestToApi api.ProjectV1, override bool) {
	Expect(requestToApi.Name).To(Equal(projectName))
	Expect(requestToApi.DefaultNodePools).To(Equal([]string{nodePoolName, nodePool2Name}))
	Expect(requestToApi.Phase).To(Equal(phase))
	Expect(requestToApi.Status).To(Equal(api.ProjectV1AllOfStatus{
		Namespace:     statusNamespace,
		Message:       statusMessage,
		Phase:         phase,
		QuotaStatuses: []api.QuotaStatusV1Inner{},
		Conditions: []map[string]string{
			{
				"type":               "NamespaceReady",
				"status":             "True",
				"lastTransitionTime": "2024-02-15T14:58:19Z",
				"lastProbeTime":      "",
			},
		},
	}))
	departmentId, _ := strconv.ParseInt(parentId, 10, 32)
	Expect(requestToApi.DepartmentId).To(Equal(int32(departmentId)))
	if override {
		Expect(requestToApi.InteractiveJobMaxIdleDurationSecs).To(Equal(float32(jobsLimit)))
		Expect(requestToApi.InteractiveJobTimeLimitSecs).To(Equal(float32(jobsLimit)))
		Expect(requestToApi.TrainingJobMaxIdleDurationSecs).To(Equal(float32(jobsLimit)))
		Expect(requestToApi.TrainingJobMaxIdleDurationSecs).To(Equal(float32(jobsLimit)))
		Expect(requestToApi.InteractivePreemptibleJobMaxIdleDurationSecs).To(Equal(float32(0)))
		Expect(*requestToApi.NodeAffinity).To(Equal(api.ProjectV1NodeAffinityResponse{
			Train: api.ProjectV1NodeAffinityResponseTrain{
				AffinityType: v1.OnlySelectedAffinityType,
				SelectedTypes: []api.ProjectV1NodeAffinityResponseTrainSelectedTypesInner{{
					Id:   0,
					Name: nodeType,
				}},
			},
			Interactive: api.ProjectV1NodeAffinityResponseInteractive{
				AffinityType:  v1.AffinityTypeNoLimit,
				SelectedTypes: []api.ProjectV1NodeAffinityResponseTrainSelectedTypesInner{},
			},
		}))
	} else {
		Expect(requestToApi.InteractiveJobMaxIdleDurationSecs).To(Equal(float32(0)))
		Expect(requestToApi.InteractiveJobTimeLimitSecs).To(Equal(float32(0)))
		Expect(requestToApi.TrainingJobMaxIdleDurationSecs).To(Equal(float32(0)))
		Expect(requestToApi.TrainingJobMaxIdleDurationSecs).To(Equal(float32(0)))
		Expect(requestToApi.InteractivePreemptibleJobMaxIdleDurationSecs).To(Equal(float32(0)))
		Expect(*requestToApi.NodeAffinity).To(Equal(api.ProjectV1NodeAffinityResponse{
			Train: api.ProjectV1NodeAffinityResponseTrain{
				AffinityType:  v1.AffinityTypeNoLimit,
				SelectedTypes: []api.ProjectV1NodeAffinityResponseTrainSelectedTypesInner{},
			},
			Interactive: api.ProjectV1NodeAffinityResponseInteractive{
				AffinityType:  v1.AffinityTypeNoLimit,
				SelectedTypes: []api.ProjectV1NodeAffinityResponseTrainSelectedTypesInner{},
			},
		}))
	}
}
