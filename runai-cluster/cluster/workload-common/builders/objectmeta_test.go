/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Labels tests", func() {
	var (
		ctx          context.Context
		projectName  string
		workloadKind string
		jobType      string
		workloadName string
		objectMeta   *metav1.ObjectMeta
	)

	BeforeEach(func() {
		// Setup common test variables
		ctx = context.Background()
		projectName = "test-project"
		ctx = project.SetProjectInContext(ctx, projectName)

		workloadKind = constants.KindInteractiveWorkload
		jobType = "interactive"
		workloadName = "test-workload"

		objectMeta = &metav1.ObjectMeta{
			Name:      workloadName,
			Namespace: "default",
			Labels:    make(map[string]string),
		}
	})

	Context("SetObjectMetaLabels", func() {
		It("should set all required labels", func() {
			err := SetObjectMetaLabels(ctx, workloadKind, jobType, workloadName, objectMeta)
			Expect(err).NotTo(HaveOccurred())

			// Verify all expected labels are set
			Expect(objectMeta.Labels[constants.LabelWorkloadKind]).To(Equal(workloadKind))
			Expect(objectMeta.Labels[constants.LabelRelease]).To(Equal(workloadName))
			Expect(objectMeta.Labels[constants.LabelWorkloadName]).To(Equal(workloadName))
			Expect(objectMeta.Labels[constants.LabelApp]).To(Equal(strings.ToLower(jobType)))
			Expect(objectMeta.Labels[constants.LabelCreatedBy]).To(Equal(jobType))
			Expect(objectMeta.Labels[constants.RunaiProjectLabel]).To(Equal(projectName))
		})

		It("should initialize the labels map if nil", func() {
			// Set labels to nil
			objectMeta.Labels = nil

			err := SetObjectMetaLabels(ctx, workloadKind, jobType, workloadName, objectMeta)
			Expect(err).NotTo(HaveOccurred())

			// Verify the map was initialized and labels were set
			Expect(objectMeta.Labels).NotTo(BeNil())
			Expect(objectMeta.Labels[constants.LabelWorkloadKind]).To(Equal(workloadKind))
		})

		It("should return error when project not found in context", func() {
			// Create a new context without project
			emptyCtx := context.Background()

			err := SetObjectMetaLabels(emptyCtx, workloadKind, jobType, workloadName, objectMeta)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("SetPriorityClassLabel", Ordered, func() {
		var (
			k8sClient          client.Client
			priorityClassNames = []string{constants.InteractivePreemptiblePriorityClass, constants.InteractivePriorityClass}
		)

		BeforeAll(func() {
			scheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			Expect(schedulingv1.AddToScheme(scheme)).To(Succeed())
			k8sClient = initFakeClientWithObjectWithStatus(scheme, nil)

			createPriorityClasses(k8sClient, priorityClassNames)
		})

		AfterAll(func() {
			deletePriorityClasses(k8sClient, priorityClassNames)
		})

		It("should set correct priority class name label", func() {
			err := SetPriorityClassLabel(ctx, k8sClient, constants.InteractivePreemptiblePriorityClass, objectMeta)
			Expect(err).NotTo(HaveOccurred())

			Expect(objectMeta.Labels[constants.LabelPriorityClassName]).To(Equal(constants.InteractivePreemptiblePriorityClass))
		})

		It("should set correct priority class name label", func() {
			err := SetPriorityClassLabel(ctx, k8sClient, constants.InteractivePriorityClass, objectMeta)
			Expect(err).NotTo(HaveOccurred())

			Expect(objectMeta.Labels[constants.LabelPriorityClassName]).To(Equal(constants.InteractivePriorityClass))
		})

		It("should initialize the labels map if nil", func() {
			// Set labels to nil
			objectMeta.Labels = nil

			err := SetPriorityClassLabel(ctx, k8sClient, constants.InteractivePreemptiblePriorityClass, objectMeta)
			Expect(err).NotTo(HaveOccurred())

			// Verify the map was initialized and label was set
			Expect(objectMeta.Labels).NotTo(BeNil())
			Expect(objectMeta.Labels[constants.LabelPriorityClassName]).To(Equal(constants.InteractivePreemptiblePriorityClass))
		})

		It("priorityClass doesn't exist", func() {
			err := SetPriorityClassLabel(ctx, k8sClient, "non-existent-priority-class", objectMeta)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})

	Context("SetPreemptibilityLabel", Ordered, func() {

		BeforeAll(func() {
			scheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			Expect(schedulingv1.AddToScheme(scheme)).To(Succeed())
		})

		It("should set correct preemptibility name label", func() {
			SetPreemptibilityLabel(string(field_groups.Preemptible), objectMeta)
			Expect(objectMeta.Labels[constants.LabelPreemptibilityName]).To(Equal(string(field_groups.Preemptible)))
		})
	})

	Context("AddUserDefinedAnnotationsAndLabels", func() {
		var containerFields *field_groups.ContainerFields

		BeforeEach(func() {
			// Initialize containerFields with some labels and annotations
			containerFields = &field_groups.ContainerFields{
				Labels: &flds.DictionaryField{
					Value: &map[string]flds.DictionaryItemAndRules{
						"custom-label": {
							Value: ptr.To("custom-value"),
						},
					},
				},
				Annotations: &flds.DictionaryField{
					Value: &map[string]flds.DictionaryItemAndRules{
						"custom-annotation": {
							Value: ptr.To("custom-annotation-value"),
						},
					},
				},
			}

			// Ensure objectMeta has initialized maps
			objectMeta.Labels = make(map[string]string)
			objectMeta.Annotations = make(map[string]string)
		})

		It("should add user defined labels and annotations", func() {
			AddUserDefinedAnnotationsAndLabels(objectMeta, containerFields)

			// Verify labels and annotations were added
			Expect(objectMeta.Labels["custom-label"]).To(Equal("custom-value"))
			Expect(objectMeta.Annotations["custom-annotation"]).To(Equal("custom-annotation-value"))
		})

		It("should handle nil containerFields", func() {
			// Set containerFields to nil
			containerFields = nil

			// This should not panic
			AddUserDefinedAnnotationsAndLabels(objectMeta, containerFields)
			// No assertions needed - we're just checking it doesn't panic
		})

		It("should handle nil values in labels and annotations", func() {
			// Create fields with nil Values
			containerFields.Labels = &flds.DictionaryField{
				Value: nil,
			}
			containerFields.Annotations = &flds.DictionaryField{
				Value: nil,
			}

			// This should not panic
			AddUserDefinedAnnotationsAndLabels(objectMeta, containerFields)
			// No assertions needed - we're just checking it doesn't panic
		})

		It("should handle nil values in field items", func() {
			// Create labels and annotations with nil values in items
			containerFields.Labels = &flds.DictionaryField{
				Value: &map[string]flds.DictionaryItemAndRules{
					"nil-label": {
						Value: nil,
					},
				},
			}
			containerFields.Annotations = &flds.DictionaryField{
				Value: &map[string]flds.DictionaryItemAndRules{
					"nil-annotation": {
						Value: nil,
					},
				},
			}

			AddUserDefinedAnnotationsAndLabels(objectMeta, containerFields)

			// Values should be set to empty strings
			Expect(objectMeta.Labels["nil-label"]).To(Equal(""))
			Expect(objectMeta.Annotations["nil-annotation"]).To(Equal(""))
		})

		It("should preserve existing labels and annotations when adding new ones", func() {
			// Set up existing labels and annotations
			objectMeta.Labels["existing-label"] = "existing-value"
			objectMeta.Annotations["existing-annotation"] = "existing-annotation-value"

			// Set up new labels and annotations in containerFields
			containerFields.Labels = &flds.DictionaryField{
				Value: &map[string]flds.DictionaryItemAndRules{
					"new-label": {
						Value: ptr.To("new-value"),
					},
				},
			}
			containerFields.Annotations = &flds.DictionaryField{
				Value: &map[string]flds.DictionaryItemAndRules{
					"new-annotation": {
						Value: ptr.To("new-annotation-value"),
					},
				},
			}

			// Call the function
			AddUserDefinedAnnotationsAndLabels(objectMeta, containerFields)

			// Verify both existing and new labels/annotations are present
			Expect(objectMeta.Labels["existing-label"]).To(Equal("existing-value"))
			Expect(objectMeta.Labels["new-label"]).To(Equal("new-value"))
			Expect(objectMeta.Annotations["existing-annotation"]).To(Equal("existing-annotation-value"))
			Expect(objectMeta.Annotations["new-annotation"]).To(Equal("new-annotation-value"))
		})
	})
})

func createPriorityClasses(k8sClient client.Client, priorityClassNames []string) {
	for _, pcName := range priorityClassNames {
		pc := &schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: pcName,
			},
			Value: 50, // doesn't matter, for the purposes of this test
		}

		err := k8sClient.Create(context.Background(), pc)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to create PriorityClass %s", pcName)
	}
}

func deletePriorityClasses(k8sClient client.Client, priorityClassNames []string) {
	for _, pcName := range priorityClassNames {
		pc := &schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: pcName,
			},
		}

		err := k8sClient.Delete(context.Background(), pc)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to delete PriorityClass %s", pcName)
	}
}
