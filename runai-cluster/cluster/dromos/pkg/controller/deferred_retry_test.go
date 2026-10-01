/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

type deferredRetryManager struct {
	ctrl.Manager
	scheme *runtime.Scheme
}

func (m deferredRetryManager) GetScheme() *runtime.Scheme { return m.scheme }

type deferredRetryPlugin struct {
	name           string
	failure        error
	postPhaseError error
	calls          *[]string
}

func (p *deferredRetryPlugin) Name() string { return p.name }

func (p *deferredRetryPlugin) HandleAdditionalResources(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner,
	_ *unstructured.Unstructured, _ *kartav1alpha1.Karta, _ client.Client, state *reconcile.State,
) error {
	*p.calls = append(*p.calls, p.name)
	if p.failure != nil {
		_ = state.AddPluginIssue(ctx, p, runner, p.failure, false, "additional resource unavailable", issue.Internal)
		state.RequestRequeueAfter(time.Minute)
	}
	return nil
}

func (p *deferredRetryPlugin) HandlePostPhase(_ context.Context, _ *dromosv1alpha1.WorkloadRunner,
	_ *unstructured.Unstructured, _ *kartav1alpha1.Karta, _ client.Client,
) error {
	*p.calls = append(*p.calls, p.name+"-post")
	return p.postPhaseError
}

var _ = Describe("deferred plugin retries", func() {
	var (
		ctx             context.Context
		controller      *WorkloadRunnerController
		k8sClient       client.WithWatch
		runner          *dromosv1alpha1.WorkloadRunner
		retrying, later *deferredRetryPlugin
		calls           []string
		request         ctrl.Request
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme := runtime.NewScheme()
		Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(kartav1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(jobsetv1alpha2.AddToScheme(scheme)).To(Succeed())
		jobset := newJobSet()
		jobset.Name = "training"
		template, err := json.Marshal(jobset)
		Expect(err).NotTo(HaveOccurred())
		runner = newWorkloadRunner("training-runner", "default", runtime.RawExtension{Raw: template})
		runner.UID = "training-runner-uid"
		request = ctrl.Request{NamespacedName: client.ObjectKeyFromObject(runner)}
		karta := test.NewValidJobSetKarta()
		karta.Status.Conditions = []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue}}
		k8sClient = fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(runner).WithObjects(runner, karta).Build()
		calls = nil
		retrying = &deferredRetryPlugin{name: "a-retrying", calls: &calls, failure: errors.New("temporary failure")}
		later = &deferredRetryPlugin{name: "b-later", calls: &calls}
		fw := framework.NewFramework(framework.NewRegistry(retrying, later), nil)
		controller = NewWorkloadRunnerController(k8sClient, fw, record.NewFakeRecorder(10))
		controller.manager = deferredRetryManager{scheme: scheme}
		controller.activeWatches[jobset.GroupVersionKind()] = true
	})

	It("applies the workload and runs remaining hooks before retrying, then clears the recovered issue", func() {
		result, err := controller.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(Equal(time.Minute))
		Expect(calls).To(Equal([]string{"a-retrying", "b-later", "a-retrying-post", "b-later-post"}))

		attached := &jobsetv1alpha2.JobSet{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: runner.Namespace, Name: "training"}, attached)).To(Succeed())
		Expect(metav1.GetControllerOf(attached)).NotTo(BeNil())
		Expect(metav1.GetControllerOf(attached).UID).To(Equal(runner.UID))
		Expect(k8sClient.Get(ctx, request.NamespacedName, runner)).To(Succeed())
		Expect(runner.Status.Plugins[retrying.Name()].Issues).To(HaveLen(1))
		Expect(runner.Status.Plugins[retrying.Name()].Issues[0].IsCritical).To(BeFalse())

		retrying.failure = nil
		calls = nil
		result, err = controller.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
		Expect(calls).To(Equal([]string{"a-retrying", "b-later", "a-retrying-post", "b-later-post"}))
		Expect(k8sClient.Get(ctx, request.NamespacedName, runner)).To(Succeed())
		Expect(runner.Status.Plugins[retrying.Name()].Issues).To(BeEmpty())
	})

	It("preserves a post-phase error instead of returning the deferred retry", func() {
		later.postPhaseError = errors.New("post-phase failed")
		result, err := controller.Reconcile(ctx, request)
		Expect(err).To(MatchError(ContainSubstring(later.postPhaseError.Error())))
		Expect(result).To(Equal(ctrl.Result{}))
		Expect(calls).To(ContainElement("b-later-post"))
	})

	It("does not retry when the runner disappears before its status is applied", func() {
		controller.client = interceptor.NewClient(k8sClient, interceptor.Funcs{
			SubResourcePatch: func(_ context.Context, _ client.Client, _ string, _ client.Object,
				_ client.Patch, _ ...client.SubResourcePatchOption,
			) error {
				return apierrors.NewNotFound(dromosv1alpha1.GroupVersion.WithResource("workloadrunners").GroupResource(), runner.Name)
			},
		})
		result, err := controller.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
		Expect(calls).To(Equal([]string{"a-retrying", "b-later"}))
	})
})
