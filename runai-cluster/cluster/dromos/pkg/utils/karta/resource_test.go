/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package karta

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	internaltest "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/internaltest"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	kartaName = "jobset-karta"
)

var (
	jobsetGvk = schema.GroupVersionKind{
		Group:   "jobset.x-k8s.io",
		Version: "v1alpha2",
		Kind:    "JobSet",
	}
	testGVK2 = schema.GroupVersionKind{
		Group:   "another.group",
		Version: "v1",
		Kind:    "AnotherKind",
	}
)

func TestRIUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Karta Utils Suite")
}

var _ = Describe("GetKartaForGVK", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("finds matching Karta", func() {
		karta := newKarta(kartaName, jobsetGvk, nil)
		anotherRI := newKarta("another-karta", testGVK2, nil)
		kartaUtils := NewFactory(newFakeClient(karta, anotherRI))

		result, err := kartaUtils.GetKartaForGVK(ctx, jobsetGvk)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		Expect(result.Name).To(Equal(kartaName))
	})

	It("returns error when not found", func() {
		kartaUtils := NewFactory(newFakeClient())

		result, err := kartaUtils.GetKartaForGVK(ctx, jobsetGvk)

		Expect(err).To(HaveOccurred())
		Expect(errors.IsNotFound(err)).To(BeTrue())
		Expect(result).To(BeNil())
	})
})

var _ = Describe("GetRootComponentGVK", func() {
	It("extracts GVK correctly", func() {
		karta := newKarta(kartaName, jobsetGvk, nil)

		result, err := GetRootComponentGVK(karta)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		Expect(*result).To(Equal(jobsetGvk))
	})

	It("returns error when Kind is nil", func() {
		karta := &kartav1alpha1.Karta{
			ObjectMeta: metav1.ObjectMeta{Name: kartaName},
			Spec: kartav1alpha1.KartaSpec{
				StructureDefinition: kartav1alpha1.StructureDefinition{
					RootComponent: kartav1alpha1.ComponentDefinition{
						Name: "root",
						Kind: nil,
					},
				},
			},
		}

		result, err := GetRootComponentGVK(karta)

		Expect(err).To(HaveOccurred())
		Expect(result).To(BeNil())
	})
})

var _ = Describe("IsReady", func() {
	It("returns true when the Ready condition is True", func() {
		karta := newKarta(kartaName, jobsetGvk, []metav1.Condition{
			{Type: string(Ready), Status: metav1.ConditionTrue},
		})

		Expect(IsReady(karta)).To(BeTrue())
	})

	It("returns false when the Ready condition is False", func() {
		karta := newKarta(kartaName, jobsetGvk, []metav1.Condition{
			{Type: string(Ready), Status: metav1.ConditionFalse},
		})

		Expect(IsReady(karta)).To(BeFalse())
	})

	It("returns false when the Ready condition is missing", func() {
		Expect(IsReady(newKarta(kartaName, jobsetGvk, nil))).To(BeFalse())
	})
})

var _ = Describe("GetResourcePhase", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	DescribeTable("phase extraction",
		func(riBuilder func() *kartav1alpha1.Karta, resourceConditions []metav1.Condition, expectedPhase dromosv1alpha1.WorkloadRunnerPhase, expectError bool) {
			karta := riBuilder()
			resource := createJobsetWithStatus("test-jobset", resourceConditions)
			factory := NewFactory(nil)

			phase, err := factory.GetResourcePhase(ctx, karta, resource)

			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(phase).To(Equal(expectedPhase))
		},
		Entry("returns Initializing when resource has Initializing=True with extra conditions",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.InitializingStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Initializing", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Running", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{
					Type:   "Initializing",
					Status: metav1.ConditionTrue,
				},
			},
			dromosv1alpha1.WorkloadRunnerPhaseInitializing,
			false,
		),
		Entry("returns Running when resource has Running=True with extra conditions",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Running", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.CompletedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Completed", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{
					Type:   "Running",
					Status: metav1.ConditionTrue,
				},
			},
			dromosv1alpha1.WorkloadRunnerPhaseRunning,
			false,
		),
		Entry("returns Completed when resource has Completed=True with extra conditions",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.CompletedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Completed", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.FailedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Failed", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{
					Type:   "Completed",
					Status: metav1.ConditionTrue,
				},
			},
			dromosv1alpha1.WorkloadRunnerPhaseCompleted,
			false,
		),
		Entry("returns Failed when resource has Failed=True with extra conditions",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.FailedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Failed", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Running", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{
					Type:   "Failed",
					Status: metav1.ConditionTrue,
				},
			},
			dromosv1alpha1.WorkloadRunnerPhaseFailed,
			false,
		),
		Entry("returns Unknown when resource has no status",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Running", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{},
			dromosv1alpha1.WorkloadRunnerPhaseUnknown,
			false,
		),
		Entry("returns Unknown when resource conditions don't match any StatusMatcher",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Running", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{
					Type:   "Pending",
					Status: metav1.ConditionTrue,
				},
			},
			dromosv1alpha1.WorkloadRunnerPhaseUnknown,
			false,
		),
		Entry("returns Failed over Running when both conditions match",
			func() *kartav1alpha1.Karta {
				builder := internaltest.NewJobSetKartaBuilder("test-karta").WithDefaults()
				builder.WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "Running", Status: ptr.To("True")},
				})
				builder.WithStatusMatcher(kartav1alpha1.FailedStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "Failed", Status: ptr.To("True")},
				})
				return builder.Build()
			},
			[]metav1.Condition{
				{
					Type:   "Running",
					Status: metav1.ConditionTrue,
				},
				{
					Type:   "Failed",
					Status: metav1.ConditionTrue,
				},
			},
			dromosv1alpha1.WorkloadRunnerPhaseFailed,
			false,
		),
		Entry("returns Suspended when resource has Suspended=True condition",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.SuspendedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Suspended", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.InitializingStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Initializing", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{Type: "Suspended", Status: metav1.ConditionTrue},
				{Type: "Initializing", Status: metav1.ConditionTrue},
			},
			dromosv1alpha1.WorkloadRunnerPhaseSuspended,
			false,
		),
		Entry("returns Suspended over Initializing (priority check)",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.SuspendedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Suspended", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.InitializingStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Initializing", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{Type: "Suspended", Status: metav1.ConditionTrue},
				{Type: "Initializing", Status: metav1.ConditionTrue},
			},
			dromosv1alpha1.WorkloadRunnerPhaseSuspended,
			false,
		),
		Entry("returns Suspending when suspending matcher matches",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.SuspendingStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Suspending", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Running", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{Type: "Suspending", Status: metav1.ConditionTrue},
			},
			dromosv1alpha1.WorkloadRunnerPhaseSuspending,
			false,
		),
		Entry("returns Resuming when resuming matcher matches",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.ResumingStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Resuming", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.SuspendedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Suspended", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{Type: "Resuming", Status: metav1.ConditionTrue},
			},
			dromosv1alpha1.WorkloadRunnerPhaseResuming,
			false,
		),
		Entry("returns Resuming over Suspended (priority check)",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.ResumingStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Resuming", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.SuspendedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Suspended", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{Type: "Resuming", Status: metav1.ConditionTrue},
				{Type: "Suspended", Status: metav1.ConditionTrue},
			},
			dromosv1alpha1.WorkloadRunnerPhaseResuming,
			false,
		),
		Entry("returns Failed over Suspending (priority check)",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.FailedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Failed", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.SuspendingStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Suspending", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{Type: "Failed", Status: metav1.ConditionTrue},
				{Type: "Suspending", Status: metav1.ConditionTrue},
			},
			dromosv1alpha1.WorkloadRunnerPhaseFailed,
			false,
		),
		Entry("returns Failed over Suspended (priority check)",
			func() *kartav1alpha1.Karta {
				return internaltest.NewJobSetKartaBuilder("test-karta").
					WithDefaults().
					WithStatusMatcher(kartav1alpha1.FailedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Failed", Status: ptr.To("True")},
					}).
					WithStatusMatcher(kartav1alpha1.SuspendedStatus, []kartav1alpha1.ExpectedCondition{
						{Type: "Suspended", Status: ptr.To("True")},
					}).
					Build()
			},
			[]metav1.Condition{
				{Type: "Failed", Status: metav1.ConditionTrue},
				{Type: "Suspended", Status: metav1.ConditionTrue},
			},
			dromosv1alpha1.WorkloadRunnerPhaseFailed,
			false,
		),
		Entry("returns Completed over Running when both conditions match",
			func() *kartav1alpha1.Karta {
				builder := internaltest.NewJobSetKartaBuilder("test-karta").WithDefaults()
				builder.WithStatusMatcher(kartav1alpha1.RunningStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "Running", Status: ptr.To("True")},
				})
				builder.WithStatusMatcher(kartav1alpha1.CompletedStatus, []kartav1alpha1.ExpectedCondition{
					{Type: "Completed", Status: ptr.To("True")},
				})
				return builder.Build()
			},
			[]metav1.Condition{
				{
					Type:   "Running",
					Status: metav1.ConditionTrue,
				},
				{
					Type:   "Completed",
					Status: metav1.ConditionTrue,
				},
			},
			dromosv1alpha1.WorkloadRunnerPhaseCompleted,
			false,
		),
	)
})

func newKarta(name string, gvk schema.GroupVersionKind, conditions []metav1.Condition) *kartav1alpha1.Karta {
	karta := &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				KartaGroupLabel:   gvk.Group,
				KartaVersionLabel: gvk.Version,
				KartaKindLabel:    gvk.Kind,
			},
		},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "root",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   gvk.Group,
						Version: gvk.Version,
						Kind:    gvk.Kind,
					},
				},
			},
		},
		Status: kartav1alpha1.KartaStatus{
			Conditions: conditions,
		},
	}

	return karta
}

func newFakeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	_ = kartav1alpha1.AddToScheme(scheme)
	_ = dromosv1alpha1.AddToScheme(scheme)

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func createJobsetWithStatus(name string, conditions []metav1.Condition) *unstructured.Unstructured {
	jobset := &unstructured.Unstructured{}
	jobset.SetGroupVersionKind(jobsetGvk)
	jobset.SetName(name)
	jobset.SetNamespace("default")

	if len(conditions) > 0 {
		conditionsSlice := make([]interface{}, len(conditions))
		for i, c := range conditions {
			conditionsSlice[i] = map[string]interface{}{
				"type":               c.Type,
				"status":             string(c.Status),
				"lastTransitionTime": metav1.Now().Format(time.RFC3339),
				"reason":             c.Type,
				"message":            "",
			}
		}
		_ = unstructured.SetNestedSlice(jobset.Object, conditionsSlice, "status", "conditions")
	}

	return jobset
}
