/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package revision

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	testingclient "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller/testing/client"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	register "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	resourceUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	commonconstants2 "github.com/run-ai/runai/runai-cluster/common/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

const (
	projectName = "test-project"
	namespace   = "runai-test-project"
	name        = "rev-1"
)

type RevisionStatusTestEntry struct {
	name                 string
	revisionStatus       servingv1.RevisionStatus
	pods                 []v1.Pod
	expectedPhase        *string
	expectedPhaseMessage *string
}

var _ = Describe("Reconcile revisions", Ordered, func() {
	var (
		k8sClient    k8sclient.Client
		reconciler   *Reconciler
		ctx          context.Context
		project      *runaiv2.Project
		namespaceObj *v1.Namespace
	)

	scheme := runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeAll(func() {
		project = resourceUtils.CreateProject(projectName)
		namespaceObj = resourceUtils.CreateNamespace(namespace, projectName)
		k8sClient = testingclient.NewFakeClientWithConflictsErrors(scheme, []k8sclient.Object{&servingv1.Revision{}}, project, namespaceObj)
		ctx = workloadproject.SetProjectInContext(context.Background(), projectName)
		err := os.Setenv(workloadconfig.EnvVarClusterDomain, "https://cluster-domain.mock")
		Expect(err).ToNot(HaveOccurred())

		reconciler = NewSimpleReconcile(k8sClient)
	})

	RunRevisionStatusTest := func(entry RevisionStatusTestEntry) {
		namespacedName := types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}

		sourceRevision := &servingv1.Revision{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Status: entry.revisionStatus,
		}
		err := k8sClient.Create(ctx, sourceRevision)
		Expect(err).To(BeNil())
		// update right after creation since create only store spec without status
		err = k8sClient.Update(ctx, sourceRevision)
		Expect(err).To(BeNil())

		for _, pod := range entry.pods {
			err := k8sClient.Create(ctx, &pod)
			Expect(err).To(BeNil())
		}

		_, err = reconciler.Reconcile(ctx, controllerruntime.Request{
			NamespacedName: namespacedName,
		})
		Expect(err).To(BeNil())

		revision := &servingv1.Revision{}
		err = k8sClient.Get(ctx, namespacedName, revision)
		Expect(err).To(BeNil())

		phase, ok := revision.Status.Annotations[commonconstants.RunaiPhaseAnnotationKey]
		Expect(ok).To(BeTrue())
		Expect(phase).ToNot(Equal(""))

		phaseVersion, ok := revision.Status.Annotations[commonconstants.RunaiPhaseVersionAnnotationKey]
		Expect(ok).To(BeTrue())

		phaseMsg, ok := revision.Status.Annotations[commonconstants.RunaiPhaseMessageAnnotationKey]
		Expect(ok).To(BeTrue())

		somePodRunning := false
		somePodPending := false
		for _, pod := range entry.pods {
			if pod.Status.Phase == v1.PodRunning {
				somePodRunning = true
			}
			if pod.Status.Phase == v1.PodPending {
				somePodPending = true
			}
		}

		if somePodRunning {
			isSourceReady := sourceRevision.IsReady() && !somePodPending
			isResultReady := phase == workloadcommon.PhaseRunning.String()
			Expect(isSourceReady).To(Equal(isResultReady))
		}

		if len(entry.pods) == 0 {
			isSourceFailed := sourceRevision.IsFailed()
			isResultFailed := phase == workloadcommon.PhaseFailed.String()
			Expect(isSourceFailed).To(Equal(isResultFailed))
		}

		if entry.expectedPhase != nil {
			Expect(*entry.expectedPhase).Should(Equal(phase), fmt.Sprintf("test '%s' should have correct phase %s", entry.name, *entry.expectedPhase))
		}
		if entry.expectedPhaseMessage != nil {
			Expect(*entry.expectedPhaseMessage).Should(Equal(phaseMsg), fmt.Sprintf("test '%s' should have correct phase message %s", entry.name, *entry.expectedPhaseMessage))
		}

		_, err = reconciler.Reconcile(ctx, controllerruntime.Request{
			NamespacedName: namespacedName,
		})
		Expect(err).To(BeNil())

		// expect to do nothing since phase is the same
		revision = &servingv1.Revision{}
		err = k8sClient.Get(ctx, namespacedName, revision)
		Expect(err).To(BeNil())

		newPhaseVersion, ok := revision.Status.Annotations[commonconstants.RunaiPhaseVersionAnnotationKey]
		Expect(ok).To(BeTrue())
		Expect(newPhaseVersion).To(Equal(phaseVersion))

		err = k8sClient.Delete(ctx, sourceRevision)
		Expect(err).To(BeNil())

		for _, pod := range entry.pods {
			err := k8sClient.Delete(ctx, &pod)
			Expect(err).To(BeNil())
		}
	}
	var testEntries = []RevisionStatusTestEntry{
		{
			name:           "empty status",
			revisionStatus: servingv1.RevisionStatus{},
			expectedPhase:  ptr.To(workloadcommon.PhaseCreating.String()),
		},
		{
			name: "has resources only",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionResourcesAvailable,
							Status: v1.ConditionTrue,
						},
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseInitializing.String()),
		},
		{
			name: "not ready",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseFailed.String()),
		},
		{
			name: "unknown ready",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhasePending.String()),
		},
		{
			name: "ready",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{{
						Type:   servingv1.RevisionConditionReady,
						Status: v1.ConditionTrue,
					}},
				},
			},
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%d)", name, 0),
						Namespace: namespace,
						Labels: map[string]string{
							commonconstants2.KnativeRevisionLabel: name,
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%d)", name, 1),
						Namespace: namespace,
						Labels: map[string]string{
							commonconstants2.KnativeRevisionLabel: name,
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseRunning.String()),
		},
		{
			name: "ready with no pods",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{{
						Type:   servingv1.RevisionConditionReady,
						Status: v1.ConditionTrue,
					}},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseStopped.String()),
		},
		{
			name: "ready and has resources",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{{
						Type:   servingv1.RevisionConditionResourcesAvailable,
						Status: v1.ConditionTrue,
					}, {
						Type:   servingv1.RevisionConditionReady,
						Status: v1.ConditionTrue,
					}},
				},
			},
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%d)", name, 1),
						Namespace: namespace,
						Labels: map[string]string{
							commonconstants2.KnativeRevisionLabel: name,
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseRunning.String()),
		},
		{
			name: "not ready and has resources",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionResourcesAvailable,
							Status: v1.ConditionTrue,
						},
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseFailed.String()),
		},
		{
			name: "starting to assign resource",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionResourcesAvailable,
							Status: v1.ConditionUnknown,
						},
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseInitializing.String()),
		},
		{
			name: "resources assigned dont have healthy condition yet",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionResourcesAvailable,
							Status: v1.ConditionTrue,
						},
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseInitializing.String()),
		},
		{
			name: "failed to assign conditions - ready not set to false yet",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionResourcesAvailable,
							Status: v1.ConditionFalse,
							Reason: "Failed to assign conditions",
						},
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseInitializing.String()),
		},
		{
			name: "failed to assign conditions",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionResourcesAvailable,
							Status: v1.ConditionFalse,
							Reason: "Failed to assign conditions",
						},
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionFalse,
							Reason: "Failed to assign conditions",
						},
					},
				},
			},
			expectedPhase:        ptr.To(workloadcommon.PhaseFailed.String()),
			expectedPhaseMessage: ptr.To("Revision failed: Failed to assign conditions"),
		},
		{
			name: "resources assigned not healthy - ready not set to false yet",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionResourcesAvailable,
							Status: v1.ConditionTrue,
						},
						{
							Type:   servingv1.RevisionConditionContainerHealthy,
							Status: v1.ConditionFalse,
							Reason: "Container is not healthy",
						},
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseInitializing.String()),
		},
		{
			name: "resources assigned not healthy",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionResourcesAvailable,
							Status: v1.ConditionTrue,
						},
						{
							Type:   servingv1.RevisionConditionContainerHealthy,
							Status: v1.ConditionFalse,
							Reason: "Container is not healthy",
						},
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionFalse,
							Reason: "Container is not healthy",
						},
					},
				},
			},
			expectedPhase:        ptr.To(workloadcommon.PhaseFailed.String()),
			expectedPhaseMessage: ptr.To("Revision failed: Container is not healthy"),
		},
		{
			name: "resources assigned and healthy",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionResourcesAvailable,
							Status: v1.ConditionTrue,
						},
						{
							Type:   servingv1.RevisionConditionContainerHealthy,
							Status: v1.ConditionTrue,
						},
						{
							Type:   apis.ConditionType(servingv1.RevisionConditionReady),
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhaseInitializing.String()),
		},
		{
			name: "test pending",
			revisionStatus: servingv1.RevisionStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   servingv1.RevisionConditionResourcesAvailable,
							Status: v1.ConditionTrue,
						},
						{
							Type:   servingv1.RevisionConditionContainerHealthy,
							Status: v1.ConditionTrue,
						},
						{
							Type:   servingv1.RevisionConditionReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%d)", name, 0),
						Namespace: namespace,
						Labels: map[string]string{
							commonconstants2.KnativeRevisionLabel: name,
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodPending,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%d)", name, 1),
						Namespace: namespace,
						Labels: map[string]string{
							commonconstants2.KnativeRevisionLabel: name,
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},
			expectedPhase: ptr.To(workloadcommon.PhasePending.String()),
		},
	}

	entries := []TableEntry{}
	for _, testEntry := range testEntries {
		entries = append(entries, Entry(testEntry.name, testEntry))
	}

	DescribeTable("Running tests", RunRevisionStatusTest, entries)

	It("should update phase version", func() {
		namespacedName := types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}

		revision := &servingv1.Revision{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		revision.Status.Annotations = map[string]string{}
		err := k8sClient.Create(ctx, revision)
		Expect(err).To(BeNil())

		err = reconciler.UpdateRevisionPhase(ctx, revision)
		Expect(err).To(BeNil())

		revision = &servingv1.Revision{}
		err = k8sClient.Get(ctx, namespacedName, revision)
		Expect(err).To(BeNil())

		phaseVersion, ok := revision.Status.Annotations[commonconstants.RunaiPhaseVersionAnnotationKey]
		Expect(ok).To(BeTrue())

		time.Sleep(time.Second)

		err = reconciler.UpdateRevisionPhase(ctx, revision)
		Expect(err).To(BeNil())

		revision = &servingv1.Revision{}
		err = k8sClient.Get(ctx, namespacedName, revision)
		Expect(err).To(BeNil())

		newPhaseVersion, ok := revision.Status.Annotations[commonconstants.RunaiPhaseVersionAnnotationKey]
		Expect(ok).To(BeTrue())

		phaseVersionInt, err := strconv.Atoi(phaseVersion)
		Expect(err).To(BeNil())

		newPhaseVersionInt, err := strconv.Atoi(newPhaseVersion)
		Expect(err).To(BeNil())

		Expect(newPhaseVersionInt).To(BeNumerically(">", phaseVersionInt))
	})

	It("should manage V2 inference revision phase annotations", func() {
		revisionName := "v2-revision"
		revision := &servingv1.Revision{
			ObjectMeta: metav1.ObjectMeta{
				Name:      revisionName,
				Namespace: namespace,
				Labels: map[string]string{
					commonconstants2.ControlPlaneWorkloadKindLabelForWorkload: commonconstants2.V2InferenceWorkloadKind,
					commonconstants2.ControlPlaneWorkloadNameLabelForWorkload: revisionName,
				},
			},
		}
		Expect(revisionLabelPredicate().Create(event.CreateEvent{Object: revision})).To(BeTrue())
		Expect(k8sClient.Create(ctx, revision)).To(Succeed())

		_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
			NamespacedName: types.NamespacedName{Namespace: namespace, Name: revisionName},
		})
		Expect(err).NotTo(HaveOccurred())

		updated := &servingv1.Revision{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: revisionName}, updated)).To(Succeed())
		Expect(updated.Status.Annotations).To(HaveKey(commonconstants.RunaiPhaseAnnotationKey))
		Expect(updated.Status.Annotations).To(HaveKey(commonconstants.RunaiPhaseMessageAnnotationKey))
		Expect(updated.Status.Annotations).To(HaveKey(commonconstants.RunaiPhaseVersionAnnotationKey))
		Expect(k8sClient.Delete(ctx, updated)).To(Succeed())
	})

	It("should exclude unmanaged V2-like Knative revisions", func() {
		revision := &servingv1.Revision{ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				commonconstants2.ControlPlaneWorkloadKindLabelForWorkload: commonconstants2.V2InferenceWorkloadKind,
			},
		}}
		Expect(revisionLabelPredicate().Create(event.CreateEvent{Object: revision})).To(BeFalse())
	})
})
