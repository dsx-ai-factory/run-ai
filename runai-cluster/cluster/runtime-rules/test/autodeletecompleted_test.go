/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/controllers"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/kubeConfig"
	repoconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func sleepToLetControllerHandleEvent() {
	time.Sleep(200 * time.Millisecond)
}

var _ = Describe("runtime-rules AutoDeleteCompleted controller — WorkloadRunner", Ordered, func() {
	var (
		ctx, cancel = context.WithCancel(context.TODO())
		stopper     chan struct{}
		k8sClient   client.WithWatch
	)

	// buildRunnerWithChild builds a WorkloadRunner whose Spec.Template embeds a
	// uniquely-named wrapped Job manifest stamped with the given labels and
	// annotations — where the runtime-rules reads cast/workloadKind and auto-delete/GPU
	// metadata from. It also returns the materialized child Job (same name) so the
	// test can seed it for owner-ref cascade realism. completionTime (when non-nil)
	// is written as the run.ai/completion-time annotation the converter reads off
	// the runner.
	buildRunnerWithChild := func(name string, phase dromosv1alpha1.WorkloadRunnerPhase, completionTime *time.Time,
		templateLabels, templateAnnotations map[string]string) (*dromosv1alpha1.WorkloadRunner, *unstructured.Unstructured) {
		childName := name + "-job"
		runner := BuildWorkloadRunner(testNamespaceName, name, phase)
		runner.Spec.Template = apiruntime.RawExtension{Raw: buildJobTemplateRaw(childName, templateLabels, templateAnnotations)}
		if completionTime != nil {
			runner.Annotations = map[string]string{constants.CompletionTimeAnnotation: completionTime.UTC().Format(time.RFC3339)}
		}

		child := &unstructured.Unstructured{}
		child.SetAPIVersion("batch/v1")
		child.SetKind("Job")
		child.SetName(childName)
		child.SetNamespace(testNamespaceName)
		if templateAnnotations != nil {
			child.SetAnnotations(templateAnnotations)
		}
		return runner, child
	}

	// buildCastRunnerWithChild is the in-scope (v1-cast, Training) shorthand.
	buildCastRunnerWithChild := func(name string, phase dromosv1alpha1.WorkloadRunnerPhase, completionTime *time.Time, templateAnnotations map[string]string) (*dromosv1alpha1.WorkloadRunner, *unstructured.Unstructured) {
		return buildRunnerWithChild(name, phase, completionTime, labelsForCastRunner("Training"), templateAnnotations)
	}

	runnerExists := func(name string) bool {
		_, err := kubeConfig.GetWLR(ctx, k8sClient, testNamespaceName, name)
		return err == nil
	}

	BeforeAll(func() {
		ns := BuildNamespace(testNamespaceName, map[string]string{RunaiQueueLabel: testProjectName})
		projects := BuildProjectFixtures(testProjectName, testNamespaceName, 5, 5)

		scheme := initSchemaWithRunner()
		k8sClient = fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(append(projects, ns)...).Build()

		stopper = make(chan struct{})
		// KindWorkloadRunner present in supportedJobKinds so the controller watches WorkloadRunners.
		autoDeleteCompletedController := controllers.NewAutoDeleteCompletedController(ctx, k8sClient, record.NewFakeRecorder(10),
			map[string]client.Object{constants.KindWorkloadRunner: &dromosv1alpha1.WorkloadRunner{}})
		go autoDeleteCompletedController.Start(ctx, stopper)

		time.Sleep(1 * time.Second)
	})

	AfterAll(func() {
		cancel()
		close(stopper)
	})

	It("auto-delete = 0, Completed: runner is deleted promptly (child cascades via owner-ref)", func() {
		now := time.Now()
		runner, child := buildCastRunnerWithChild("ad-completed", dromosv1alpha1.WorkloadRunnerPhaseCompleted, &now,
			map[string]string{constants.AutoDeletionTimeAfterCompletionAnnotation: "0"})
		Expect(k8sClient.Create(ctx, child)).To(Succeed())
		Expect(k8sClient.Create(ctx, runner)).To(Succeed())

		sleepToLetControllerHandleEvent()
		Expect(runnerExists("ad-completed")).To(BeFalse())
	})

	It("Failed is treated like Completed (delay 0 → immediate delete)", func() {
		now := time.Now()
		runner, child := buildCastRunnerWithChild("ad-failed", dromosv1alpha1.WorkloadRunnerPhaseFailed, &now,
			map[string]string{constants.AutoDeletionTimeAfterCompletionAnnotation: "0"})
		Expect(k8sClient.Create(ctx, child)).To(Succeed())
		Expect(k8sClient.Create(ctx, runner)).To(Succeed())

		sleepToLetControllerHandleEvent()
		Expect(runnerExists("ad-failed")).To(BeFalse())
	})

	It("auto-delete > 0: survives until the delay elapses, then is deleted", func() {
		now := time.Now()
		runner, child := buildCastRunnerWithChild("ad-delayed", dromosv1alpha1.WorkloadRunnerPhaseCompleted, &now,
			map[string]string{constants.AutoDeletionTimeAfterCompletionAnnotation: "2"})
		Expect(k8sClient.Create(ctx, child)).To(Succeed())
		Expect(k8sClient.Create(ctx, runner)).To(Succeed())

		time.Sleep(1 * time.Second)
		Expect(runnerExists("ad-delayed")).To(BeTrue())

		time.Sleep(1 * time.Second)
		sleepToLetControllerHandleEvent()
		Expect(runnerExists("ad-delayed")).To(BeFalse())
	})

	It("unset auto-delete annotation: runner is never deleted (legacy unset parity)", func() {
		now := time.Now()
		runner, child := buildCastRunnerWithChild("ad-unset", dromosv1alpha1.WorkloadRunnerPhaseCompleted, &now, nil)
		Expect(k8sClient.Create(ctx, child)).To(Succeed())
		Expect(k8sClient.Create(ctx, runner)).To(Succeed())

		time.Sleep(1 * time.Second)
		sleepToLetControllerHandleEvent()
		Expect(runnerExists("ad-unset")).To(BeTrue())
	})

	It("Suspended runner is never auto-deleted, even with a stale completion-time annotation", func() {
		now := time.Now()
		runner, child := buildCastRunnerWithChild("ad-suspended", dromosv1alpha1.WorkloadRunnerPhaseSuspended, &now,
			map[string]string{constants.AutoDeletionTimeAfterCompletionAnnotation: "0"})
		Expect(k8sClient.Create(ctx, child)).To(Succeed())
		Expect(k8sClient.Create(ctx, runner)).To(Succeed())

		time.Sleep(1 * time.Second)
		sleepToLetControllerHandleEvent()
		Expect(runnerExists("ad-suspended")).To(BeTrue())
	})

	It("non-cast runner is never auto-deleted (scope gate consistent with the time-limit path)", func() {
		now := time.Now()
		// Template labels without the v1-cast marker — this runner is out of scope.
		nonCastLabels := labelsForCastRunner("Training")
		delete(nonCastLabels, repoconstants.ConvertedFromNativeLabel)
		runner, child := buildRunnerWithChild("ad-noncast", dromosv1alpha1.WorkloadRunnerPhaseCompleted, &now,
			nonCastLabels, map[string]string{constants.AutoDeletionTimeAfterCompletionAnnotation: "0"})
		Expect(k8sClient.Create(ctx, child)).To(Succeed())
		Expect(k8sClient.Create(ctx, runner)).To(Succeed())

		time.Sleep(1 * time.Second)
		sleepToLetControllerHandleEvent()
		Expect(runnerExists("ad-noncast")).To(BeTrue())
	})
})
