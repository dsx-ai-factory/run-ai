/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"strconv"
	stdtesting "testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/topology"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"net/http"
)

func TestScopeInfoForAuthValidation(t *stdtesting.T) {
	const requestingTenantId uint = 1001

	tests := []struct {
		name          string
		scopeType     api.ScopeType
		scopeId       string
		expectedScope api.ScopeType
		expectedId    string
	}{
		{
			name:          "system templates use the requesting tenant",
			scopeType:     api.SCOPETYPE_SYSTEM,
			expectedScope: api.SCOPETYPE_TENANT,
			expectedId:    strconv.FormatUint(uint64(requestingTenantId), 10),
		},
		{
			name:          "tenant scope is preserved",
			scopeType:     api.SCOPETYPE_TENANT,
			scopeId:       "tenant-id",
			expectedScope: api.SCOPETYPE_TENANT,
			expectedId:    "tenant-id",
		},
		{
			name:          "cluster scope is preserved",
			scopeType:     api.SCOPETYPE_CLUSTER,
			scopeId:       "cluster-id",
			expectedScope: api.SCOPETYPE_CLUSTER,
			expectedId:    "cluster-id",
		},
		{
			name:          "department scope is preserved",
			scopeType:     api.SCOPETYPE_DEPARTMENT,
			scopeId:       "department-id",
			expectedScope: api.SCOPETYPE_DEPARTMENT,
			expectedId:    "department-id",
		},
		{
			name:          "project scope is preserved",
			scopeType:     api.SCOPETYPE_PROJECT,
			scopeId:       "project-id",
			expectedScope: api.SCOPETYPE_PROJECT,
			expectedId:    "project-id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *stdtesting.T) {
			template := &UnifiedRequest{
				tenantId: 0,
				template: &api.SupersetTemplate{
					Meta: api.TemplateMetaFields{
						ScopeType: test.scopeType,
						ScopeId:   test.scopeId,
					},
				},
			}

			scope := template.ScopeInfoForAuthValidation(requestingTenantId)

			if string(scope.Scope) != string(test.expectedScope) {
				t.Fatalf("expected scope %q, got %q", test.expectedScope, scope.Scope)
			}
			if scope.ScopeId != test.expectedId {
				t.Fatalf("expected scope ID %q, got %q", test.expectedId, scope.ScopeId)
			}
		})
	}
}

var _ = Describe("unified template validations", func() {
	var randoms *testutils.RandomValues
	var ctx context.Context

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		Expect(randoms).NotTo(BeNil())

		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
	})

	Context("validate specs", func() {
		It("rejects segment size for an explicitly non-PyTorch distributed template", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					SegmentSize:          convutils.Ref(int32(2)),
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(validateErr).NotTo(BeNil())
			Expect(int(validateErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(validateErr.Message).To(ContainSubstring("supported only for PyTorch"))
		})

		DescribeTable("allows segment size with elastic replica bounds",
			func(spec api.SupersetSpec) {
				ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{Spec: spec})
				Expect(creErr).To(BeNil())

				_, _, validateErr := ur.ValidateSpecs(ctx)
				Expect(validateErr).To(BeNil())
			},
			Entry("minReplicas", api.SupersetSpec{
				DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
				SegmentSize:          convutils.Ref(int32(2)),
				MinReplicas:          convutils.Ref(int32(1)),
			}),
			Entry("maxReplicas", api.SupersetSpec{
				DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
				SegmentSize:          convutils.Ref(int32(2)),
				MaxReplicas:          convutils.Ref(int32(4)),
			}),
		)

		It("allows segment size for a PyTorch distributed template", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					SegmentSize:          convutils.Ref(int32(2)),
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(validateErr).To(BeNil())
		})

		It("allows segment size in a partial distributed template without a framework", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					Image:       convutils.Ref("ubuntu"),
					SegmentSize: convutils.Ref(int32(2)),
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(validateErr).To(BeNil())
			Expect(ur.template.Spec.DistributedFramework).To(BeNil())

			ur.template.CombinedSpec = ur.template.Spec
			runnable := ur.RunnableInfo()
			Expect(runnable.IsRunnable).To(BeFalse())
			Expect(runnable.Issue).To(Equal(convutils.Ref("distributedFramework is not provided")))
		})

		It("allows an annotation-only segment size in a partial distributed template without a framework", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					Annotations: convutils.Ref([]api.Annotation{
						testTemplateAnnotation(topology.KAISegmentSizeAnnotationKey, "2"),
					}),
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(validateErr).To(BeNil())
		})

		It("rejects annotation-only segment size for a non-PyTorch distributed template", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Annotations: convutils.Ref([]api.Annotation{
						testTemplateAnnotation(topology.KAISegmentSizeAnnotationKey, "2"),
					}),
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(validateErr).NotTo(BeNil())
			Expect(validateErr.Message).To(ContainSubstring("supported only for PyTorch"))
		})

		It("rejects conflicting typed and annotation segment sizes", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					SegmentSize:          convutils.Ref(int32(2)),
					Annotations: convutils.Ref([]api.Annotation{
						testTemplateAnnotation(topology.KAISegmentSizeAnnotationKey, "3"),
					}),
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(validateErr).NotTo(BeNil())
			Expect(validateErr.Message).To(ContainSubstring("must match"))
		})

		It("rejects segment placement without a segment size", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Annotations: convutils.Ref([]api.Annotation{
						testTemplateAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "rack"),
					}),
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(validateErr).NotTo(BeNil())
			Expect(validateErr.Message).To(ContainSubstring("must be provided when segment placement is configured"))
		})

		It("rejects templates that exceed the maximum number of segments", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					NumWorkers:           convutils.Ref(int32(10_001)),
					SegmentSize:          convutils.Ref(int32(1)),
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(validateErr).NotTo(BeNil())
			Expect(validateErr.Message).To(ContainSubstring("exceeding the maximum of 10000"))
		})

		It("rejects segment annotations on a separate master", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Annotations: convutils.Ref([]api.Annotation{
							testTemplateAnnotation(topology.KAISegmentSizeAnnotationKey, "2"),
						}),
					},
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(validateErr).NotTo(BeNil())
			Expect(validateErr.Message).To(ContainSubstring("not on a separate master spec"))
		})

		It("should reject master on non-distributed", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Image: convutils.Ref("ubuntu"),
					},
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(int(validateErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(validateErr.Message).To(ContainSubstring("master can be provided for distributed templates only"))
		})

		It("should reject master-same-as-worker on non-distributed", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_TRAINING, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(int(validateErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(validateErr.Message).To(ContainSubstring("masterSpecSameAsWorker can be provided for distributed templates only"))
		})

		It("should reject master AND master-same-as-worker distributed", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Image: convutils.Ref("ubuntu"),
					},
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(int(validateErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(validateErr.Message).To(ContainSubstring("master should not be provided when masterSpecSameAsWorker=true"))
		})

		It("master is forbidden for JAX", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					Image:                convutils.Ref("ubuntu"),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_JAX),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Image: convutils.Ref("ubuntu"),
					},
				},
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(int(validateErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(validateErr.Message).To(ContainSubstring("master should not be provided when distributedFramework=JAX"))
		})

		It("master-same-as-worker is forbidden for JAX", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				Spec: api.SupersetSpec{
					Image:                convutils.Ref("ubuntu"),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_JAX),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			})
			Expect(creErr).To(BeNil())

			_, _, validateErr := ur.ValidateSpecs(ctx)
			Expect(int(validateErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(validateErr.Message).To(ContainSubstring("master should not be provided when distributedFramework=JAX"))
		})
	})

	Context("runnable info - non distributed", func() {
		It("image is given with port - can run", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_INFERENCE, api.SupersetTemplate{
				CombinedSpec: api.SupersetSpec{
					Image:       convutils.Ref("ubuntu"),
					ServingPort: &api.SupersetServingPort{Container: convutils.Ref(int32(8080))},
				},
			})
			Expect(creErr).To(BeNil())

			result := ur.RunnableInfo()
			Expect(result.IsRunnable).To(Equal(true))
		})
		It("image is given with no port - can not run", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_INFERENCE, api.SupersetTemplate{
				CombinedSpec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			})
			Expect(creErr).To(BeNil())

			result := ur.RunnableInfo()
			Expect(result.IsRunnable).To(Equal(false))
		})

		It("image is not given - cannot run", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_TRAINING, api.SupersetTemplate{
				CombinedSpec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home"),
				},
			})
			Expect(creErr).To(BeNil())

			result := ur.RunnableInfo()
			Expect(result.IsRunnable).To(Equal(false))
			Expect(*result.Issue).To(Equal("image is not provided"))
		})
	})

	Context("runnable info - distributed", func() {
		It("master image is given - can run", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				CombinedSpec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				CombinedMasterSpec: &api.MasterTemplateSpec{
					Image: convutils.Ref("ubuntu"),
				},
			})
			Expect(creErr).To(BeNil())

			result := ur.RunnableInfo()
			Expect(result.IsRunnable).To(Equal(true))
		})

		It("no framework - cannot run", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				CombinedSpec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
				CombinedMasterSpec: &api.MasterTemplateSpec{
					Image: convutils.Ref("ubuntu"),
				},
			})
			Expect(creErr).To(BeNil())

			result := ur.RunnableInfo()
			Expect(result.IsRunnable).To(Equal(false))
			Expect(*result.Issue).To(Equal("distributedFramework is not provided"))
		})

		It("master image is not given - cannot run", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				CombinedSpec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				CombinedMasterSpec: &api.MasterTemplateSpec{
					WorkingDir: convutils.Ref("/home"),
				},
			})
			Expect(creErr).To(BeNil())

			result := ur.RunnableInfo()
			Expect(result.IsRunnable).To(Equal(false))
			Expect(*result.Issue).To(Equal("master.image is not provided"))
		})

		It("no master for MPI - cannot run", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				CombinedSpec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Image:                convutils.Ref("ubuntu"),
				},
			})
			Expect(creErr).To(BeNil())

			result := ur.RunnableInfo()
			Expect(result.IsRunnable).To(Equal(false))
			Expect(*result.Issue).To(Equal("master is not provided for MPI framework"))
		})

		It("no master for PyTorch - CAN run", func() {
			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, api.SupersetTemplate{
				CombinedSpec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
			})
			Expect(creErr).To(BeNil())

			result := ur.RunnableInfo()
			Expect(result.IsRunnable).To(Equal(true))
		})
	})

	Context("update template", func() {
		It("rename + add description", func() {
			origTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name: "old-name",
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, origTemplate)
			Expect(creErr).To(BeNil())

			request := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:        "old-name",
					Description: convutils.Ref("added-description"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			urRequest, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, request)
			Expect(creErr).To(BeNil())

			_, updErr := ur.UpdateTemplateFrom(ctx, urRequest)
			Expect(updErr).To(BeNil())

			// align the fields that are computed
			Expect(ur.template.Meta.UpdatedAt).NotTo(Equal(request.Meta.UpdatedAt))
			Expect(ur.template.Meta.UpdatedBy).NotTo(Equal(request.Meta.UpdatedBy))
			request.CombinedSpec = origTemplate.Spec
			request.Runnable = api.RunnableInfo{IsRunnable: true}
			request.Meta.UpdatedAt = ur.template.Meta.UpdatedAt
			request.Meta.UpdatedBy = convutils.Ref(testing.TestUser)

			testutils.Expect(*ur.template).ToEqual(request)
		})

		It("rename + remove description", func() {
			origTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:        "old-name",
					Description: convutils.Ref("old-description"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, origTemplate)
			Expect(creErr).To(BeNil())

			request := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name: "old-name",
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			urRequest, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, request)
			Expect(creErr).To(BeNil())

			_, updErr := ur.UpdateTemplateFrom(ctx, urRequest)
			Expect(updErr).To(BeNil())

			// align the fields that are computed
			Expect(ur.template.Meta.UpdatedAt).NotTo(Equal(request.Meta.UpdatedAt))
			Expect(ur.template.Meta.UpdatedBy).NotTo(Equal(request.Meta.UpdatedBy))
			request.CombinedSpec = origTemplate.Spec
			request.Runnable = api.RunnableInfo{IsRunnable: true}
			request.Meta.UpdatedAt = ur.template.Meta.UpdatedAt
			request.Meta.UpdatedBy = convutils.Ref(testing.TestUser)

			testutils.Expect(*ur.template).ToEqual(request)
		})

		It("from wroker+master to same as", func() {
			origTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name: "old-name",
				},
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						WorkingDir: convutils.Ref("/home/master"),
					},
				},
			}

			origTemplate.CombinedSpec = origTemplate.Spec
			origTemplate.CombinedMasterSpec = origTemplate.Master.Spec

			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, origTemplate)
			Expect(creErr).To(BeNil())

			runInfo := ur.RunnableInfo()
			Expect(runInfo.IsRunnable).To(Equal(false))
			Expect(*runInfo.Issue).To(ContainSubstring("master.image is not provided"))

			request := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name: "old-name",
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			urRequest, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, request)
			Expect(creErr).To(BeNil())

			_, updErr := ur.UpdateTemplateFrom(ctx, urRequest)
			Expect(updErr).To(BeNil())

			Expect(urRequest.template.CombinedMasterSpec).To(BeNil())
			request.Runnable = api.RunnableInfo{IsRunnable: true}
		})
	})

	Context("patch template", func() {
		It("validates after clearing segment size while changing framework", func() {
			origTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					WorkloadType: api.WORKLOADTYPE_DISTRIBUTED,
					Name:         "old-name",
				},
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
					SegmentSize:          convutils.Ref(int32(2)),
				},
			}
			origTemplate.CombinedSpec = origTemplate.Spec

			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, origTemplate)
			Expect(creErr).To(BeNil())

			request := api.SupersetTemplate{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
				},
			}
			urRequest, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, request)
			Expect(creErr).To(BeNil())

			_, updErr := ur.PatchTemplateFrom(ctx, urRequest, updateOptions{
				patchMode:       true,
				patchMetaFields: &api.TemplatePatchMetaFields{},
				fieldsToClear:   []string{"spec.segmentSize"},
			})
			Expect(updErr).To(BeNil())
			Expect(ur.template.Spec.SegmentSize).To(BeNil())
			Expect(ur.template.Spec.DistributedFramework).To(Equal(convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI)))
		})

		It("rename + add image", func() {
			origTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					WorkloadType: api.WORKLOADTYPE_DISTRIBUTED,
					Name:         "old-name",
					Description:  convutils.Ref("old-description"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						WorkingDir: convutils.Ref("/home/master"),
					},
				},
			}

			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, origTemplate)
			Expect(creErr).To(BeNil())

			runInfo := ur.RunnableInfo()
			Expect(runInfo.IsRunnable).To(Equal(false))
			Expect(*runInfo.Issue).To(ContainSubstring("image is not provided"))

			request := api.SupersetTemplate{
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			urRequest, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, request)
			Expect(creErr).To(BeNil())

			_, updErr := ur.PatchTemplateFrom(ctx, urRequest, updateOptions{
				patchMode: true,
				patchMetaFields: &api.TemplatePatchMetaFields{
					Name: convutils.Ref("new-name"),
				},
			})
			Expect(updErr).To(BeNil())

			Expect(ur.template.Meta.Name).To(Equal("new-name"))
			Expect(*ur.template.Meta.Description).To(Equal("old-description"))
			Expect(ur.template.Runnable.IsRunnable).To(Equal(false))
			Expect(*ur.template.Runnable.Issue).To(Equal("distributedFramework is not provided"))

			//
			//  patch the distributed framework
			//
			request = api.SupersetTemplate{
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
				},
			}

			urRequest, creErr = NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, request)
			Expect(creErr).To(BeNil())

			_, updErr = ur.PatchTemplateFrom(ctx, urRequest, updateOptions{
				patchMode:       true,
				patchMetaFields: &api.TemplatePatchMetaFields{},
			})
			Expect(updErr).To(BeNil())

			Expect(ur.template.Meta.Name).To(Equal("new-name"))
			Expect(*ur.template.Meta.Description).To(Equal("old-description"))
			Expect(ur.template.Runnable.IsRunnable).To(Equal(false))
			Expect(*ur.template.Runnable.Issue).To(Equal("master.image is not provided"))

			//
			//  patch the master image
			//
			request = api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					WorkloadType: api.WORKLOADTYPE_DISTRIBUTED,
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Image: convutils.Ref("ubuntu"),
					},
				},
			}

			urRequest, creErr = NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, request)
			Expect(creErr).To(BeNil())

			_, updErr = ur.PatchTemplateFrom(ctx, urRequest, updateOptions{
				patchMode:       true,
				patchMetaFields: &api.TemplatePatchMetaFields{},
			})
			Expect(updErr).To(BeNil())

			Expect(ur.template.Meta.Name).To(Equal("new-name"))
			Expect(*ur.template.Meta.Description).To(Equal("old-description"))
			Expect(ur.template.Runnable.IsRunnable).To(Equal(true))
		})

	})

	Context("patch distributed template", func() {

		It("with clear fields", func() {
			origTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:        "old-name",
					Description: convutils.Ref("old-description"),
				},
				Spec: api.SupersetSpec{
					Image:      convutils.Ref("worker-image"),
					WorkingDir: convutils.Ref("/home/worker"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(3)),
						CpuCoreRequest:    convutils.Ref(1.4),
						ExtendedResources: &[]api.ExtendedResource{
							{
								Resource: convutils.Ref("worker/a"),
								Quantity: convutils.Ref("5MB"),
							},
							{
								Resource: convutils.Ref("worker/b"),
								Quantity: convutils.Ref("10MB"),
							},
							{
								Resource: convutils.Ref("worker/c"),
								Quantity: convutils.Ref("12MB"),
							},
						},
					},
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Image:      convutils.Ref("master-image"),
						WorkingDir: convutils.Ref("/home/master"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuDevicesRequest: convutils.Ref(int32(5)),
							CpuCoreRequest:    convutils.Ref(1.4),
							ExtendedResources: &[]api.ExtendedResource{
								{
									Resource: convutils.Ref("master/a"),
									Quantity: convutils.Ref("15MB"),
								},
								{
									Resource: convutils.Ref("master/b"),
									Quantity: convutils.Ref("25MB"),
								},
							},
						},
					},
				},
			}

			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, origTemplate)
			Expect(creErr).To(BeNil())

			//
			//  clear fields
			//
			request := api.SupersetTemplate{}

			urRequest, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, request)
			Expect(creErr).To(BeNil())

			_, updErr := ur.PatchTemplateFrom(ctx, urRequest, updateOptions{
				patchMode:       true,
				patchMetaFields: &api.TemplatePatchMetaFields{},
				fieldsToClear: []string{
					"spec.image",
					"spec.compute.extendedResources[1]",
					"master.spec.compute.extendedResources[1].quantity",
				},
			})
			Expect(updErr).To(BeNil())

			expectedTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:        "old-name",
					Description: convutils.Ref("old-description"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(3)),
						CpuCoreRequest:    convutils.Ref(1.4),
						ExtendedResources: &[]api.ExtendedResource{
							{
								Resource: convutils.Ref("worker/a"),
								Quantity: convutils.Ref("5MB"),
							},
							{
								Resource: convutils.Ref("worker/c"),
								Quantity: convutils.Ref("12MB"),
							},
						},
					},
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Image:      convutils.Ref("master-image"),
						WorkingDir: convutils.Ref("/home/master"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuDevicesRequest: convutils.Ref(int32(5)),
							CpuCoreRequest:    convutils.Ref(1.4),
							ExtendedResources: &[]api.ExtendedResource{
								{
									Resource: convutils.Ref("master/a"),
									Quantity: convutils.Ref("15MB"),
								},
								{
									Resource: convutils.Ref("master/b"),
								},
							},
						},
					},
				},
			}

			testutils.Expect(ur.template.CombinedSpec).ToEqual(expectedTemplate.Spec)
			testutils.Expect(ur.template.CombinedMasterSpec).ToEqual(expectedTemplate.Master.Spec)

			//
			//  remove the master spec
			//
			_, updErr = ur.PatchTemplateFrom(ctx, urRequest, updateOptions{
				patchMode:       true,
				patchMetaFields: &api.TemplatePatchMetaFields{},
				fieldsToClear: []string{
					"master.spec",
				},
			})
			Expect(updErr).To(BeNil())

			expectedTemplate = api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:        "old-name",
					Description: convutils.Ref("old-description"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(3)),
						CpuCoreRequest:    convutils.Ref(1.4),
						ExtendedResources: &[]api.ExtendedResource{
							{
								Resource: convutils.Ref("worker/a"),
								Quantity: convutils.Ref("5MB"),
							},
							{
								Resource: convutils.Ref("worker/c"),
								Quantity: convutils.Ref("12MB"),
							},
						},
					},
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{},
				},
			}

			testutils.Expect(ur.template.CombinedSpec).ToEqual(expectedTemplate.Spec)
			testutils.Expect(ur.template.CombinedMasterSpec).ToEqual(expectedTemplate.Master.Spec)

		})

	})

	Context("patch workspace template", func() {
		It("with clear fields", func() {
			origTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:        "old-name",
					Description: convutils.Ref("old-description"),
				},
				Spec: api.SupersetSpec{
					Image:      convutils.Ref("worker-image"),
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			ur, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, origTemplate)
			Expect(creErr).To(BeNil())

			//
			//  add create home directory, remove image
			//
			request := api.SupersetTemplate{
				Spec: api.SupersetSpec{
					CreateHomeDir: convutils.Ref(false),
				},
			}

			urRequest, creErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, request)
			Expect(creErr).To(BeNil())

			_, updErr := ur.PatchTemplateFrom(ctx, urRequest, updateOptions{
				patchMode:       true,
				patchMetaFields: &api.TemplatePatchMetaFields{},
				fieldsToClear: []string{
					"spec.workingDir",
				},
			})
			Expect(updErr).To(BeNil())

			expectedTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:        "old-name",
					Description: convutils.Ref("old-description"),
				},
				Spec: api.SupersetSpec{
					Image:         convutils.Ref("worker-image"),
					CreateHomeDir: convutils.Ref(false),
				},
			}

			testutils.Expect(ur.template.CombinedSpec).ToEqual(expectedTemplate.Spec)
			Expect(ur.template.Runnable.IsRunnable).To(Equal(true))

			//
			//   now remove the image
			//
			_, updErr = ur.PatchTemplateFrom(ctx, urRequest, updateOptions{
				patchMode:       true,
				patchMetaFields: &api.TemplatePatchMetaFields{},
				fieldsToClear: []string{
					"spec.image",
				},
			})
			Expect(updErr).To(BeNil())

			expectedTemplate = api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:        "old-name",
					Description: convutils.Ref("old-description"),
				},
				Spec: api.SupersetSpec{
					CreateHomeDir: convutils.Ref(false),
				},
			}

			testutils.Expect(ur.template.CombinedSpec).ToEqual(expectedTemplate.Spec)
			Expect(ur.template.Runnable.IsRunnable).To(Equal(false))

			//
			//   remove the entire spec
			//
			_, updErr = ur.PatchTemplateFrom(ctx, urRequest, updateOptions{
				patchMode:       true,
				patchMetaFields: &api.TemplatePatchMetaFields{},
				fieldsToClear: []string{
					"spec",
				},
			})
			Expect(updErr).To(BeNil())

			expectedTemplate = api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:        "old-name",
					Description: convutils.Ref("old-description"),
				},
				Spec: api.SupersetSpec{},
			}

			testutils.Expect(ur.template.CombinedSpec).ToEqual(expectedTemplate.Spec)
			Expect(ur.template.Runnable.IsRunnable).To(Equal(false))

			//
			// invalid field
			//
			_, updErr = ur.PatchTemplateFrom(ctx, urRequest, updateOptions{
				patchMode:       true,
				patchMetaFields: &api.TemplatePatchMetaFields{},
				fieldsToClear: []string{
					"spec.oops",
				},
			})
			Expect(updErr).NotTo(BeNil())
			Expect(updErr.Message).To(ContainSubstring("spec.oops: field oops does not exist"))
		})
	})
})

func testTemplateAnnotation(name, value string) api.Annotation {
	return api.Annotation{
		Name:  convutils.Ref(name),
		Value: convutils.Ref(value),
	}
}
