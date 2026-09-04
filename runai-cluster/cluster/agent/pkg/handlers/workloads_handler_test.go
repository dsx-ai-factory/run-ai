/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

// calling the import here to ensure that the helper init takes place before anything else
// see inside the helper package for more info
import (
	"context"
	"errors"
	"sync"

	_ "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/tests/helper"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	workloads_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	agentcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/common"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

type listErrorReader struct {
	client.Reader
	err error
}

func (r listErrorReader) List(context.Context, client.ObjectList, ...client.ListOption) error {
	return r.err
}

var _ = Describe("Legacy workload sunset", func() {
	var (
		handler *WorkloadsHandler
		scheme  *runtime.Scheme
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
		Expect(runaiv2.AddToScheme(scheme)).To(Succeed())
		Expect(twv2.AddToScheme(scheme)).To(Succeed())
		project := &kaiorg.Project{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{runaiv2.ProjectIdLabel: "project-1"}},
			Status:     kaiorg.ProjectStatus{Namespace: "project-namespace"},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(project).Build()

		handler = &WorkloadsHandler{
			CommonHandler: &CommonHandler{
				cachedClient: fakeClient,
				runaiClient:  fakeClient,
				gvk:          twv2.GroupVersion.WithKind("TrainingWorkload"),
			},
			workloadType: agentcommon.WorkloadTypeTraining,
		}
	})

	It("rejects legacy updates without applying them", func() {
		updates := []workloads_openapi.WorkloadsSyncInfo{{
			Workload: &workloads_openapi.SupersetWorkload{
				Name:         "legacy-training",
				ProjectId:    "project-1",
				DesiredPhase: workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			},
		}}

		clusterUpdates, perWorkloadErrors, err := handler.getClusterUpdates(context.Background(), updates)

		Expect(err).NotTo(HaveOccurred())
		Expect(perWorkloadErrors).To(ConsistOf(MatchError("legacy workload legacy-training only supports deletion")))
		Expect(clusterUpdates.ToDelete).To(BeEmpty())
	})

	It("rejects legacy updates before resolving their project", func() {
		updates := []workloads_openapi.WorkloadsSyncInfo{{
			Workload: &workloads_openapi.SupersetWorkload{
				Name:         "legacy-training",
				ProjectId:    "missing-project",
				DesiredPhase: workloads_openapi.WORKLOADDESIREDPHASE_RUNNING,
			},
		}}

		clusterUpdates, perWorkloadErrors, err := handler.getClusterUpdates(context.Background(), updates)

		Expect(err).NotTo(HaveOccurred())
		Expect(perWorkloadErrors).To(ConsistOf(MatchError("legacy workload legacy-training only supports deletion")))
		Expect(clusterUpdates.ToDelete).To(BeEmpty())
	})

	It("reports a project namespace lookup failure for every workload", func() {
		namespaceErr := errors.New("failed to list projects")
		handler.cachedClient = listErrorReader{Reader: handler.cachedClient, err: namespaceErr}
		updates := []workloads_openapi.WorkloadsSyncInfo{
			{Workload: &workloads_openapi.SupersetWorkload{Name: "legacy-training-1"}},
			{Workload: &workloads_openapi.SupersetWorkload{Name: "legacy-training-2"}},
		}

		clusterUpdates, perWorkloadErrors, err := handler.getClusterUpdates(context.Background(), updates)

		Expect(err).To(MatchError(namespaceErr))
		Expect(perWorkloadErrors).To(ConsistOf(MatchError(namespaceErr), MatchError(namespaceErr)))
		Expect(clusterUpdates.ToDelete).To(BeEmpty())
	})

	It("keeps the legacy GVK needed to delete stale workloads", func() {
		updates := []workloads_openapi.WorkloadsSyncInfo{{
			Workload: &workloads_openapi.SupersetWorkload{
				Name:         "legacy-training",
				ProjectId:    "project-1",
				DesiredPhase: workloads_openapi.WORKLOADDESIREDPHASE_DELETED,
			},
		}}

		clusterUpdates, perWorkloadErrors, err := handler.getClusterUpdates(context.Background(), updates)

		Expect(err).NotTo(HaveOccurred())
		Expect(perWorkloadErrors).To(ConsistOf(BeNil()))
		Expect(clusterUpdates.ToDelete).To(ConsistOf(types.NamespacedName{
			Name: "legacy-training", Namespace: "project-namespace",
		}))
		Expect(clusterUpdates.ToDeleteGvk).To(ConsistOf(twv2.GroupVersion.WithKind("TrainingWorkload")))
	})

	It("reports an unknown project instead of treating its workload as deleted", func() {
		updates := []workloads_openapi.WorkloadsSyncInfo{{
			Workload: &workloads_openapi.SupersetWorkload{
				Name:         "legacy-training",
				ProjectId:    "missing-project",
				DesiredPhase: workloads_openapi.WORKLOADDESIREDPHASE_DELETED,
			},
		}}

		clusterUpdates, perWorkloadErrors, err := handler.getClusterUpdates(context.Background(), updates)

		Expect(err).NotTo(HaveOccurred())
		Expect(perWorkloadErrors).To(ConsistOf(MatchError("unknown project missing-project")))
		Expect(clusterUpdates.ToDelete).To(BeEmpty())
	})

	It("reports deletion of an already absent legacy workload as successful", func() {
		updates := []workloads_openapi.WorkloadsSyncInfo{{
			Workload: &workloads_openapi.SupersetWorkload{
				Name:         "already-absent",
				ProjectId:    "project-1",
				DesiredPhase: workloads_openapi.WORKLOADDESIREDPHASE_DELETED,
			},
		}}
		wg := &sync.WaitGroup{}
		wg.Add(1)

		perWorkloadErrors, err := handler.ProcessWorkloads(context.Background(), updates, wg)

		Expect(err).NotTo(HaveOccurred())
		Expect(perWorkloadErrors).To(ConsistOf(BeNil()))
	})

	It("reports a failed legacy workload deletion and creates an event", func() {
		deleteErr := errors.New("delete failed")
		project := &kaiorg.Project{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{runaiv2.ProjectIdLabel: "project-1"}},
			Status:     kaiorg.ProjectStatus{Namespace: "project-namespace"},
		}
		workload := &twv2.TrainingWorkload{ObjectMeta: metav1.ObjectMeta{
			Name: "legacy-training", Namespace: "project-namespace",
		}}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(project, workload).
			WithInterceptorFuncs(interceptor.Funcs{
				Delete: func(_ context.Context, _ client.WithWatch, _ client.Object, _ ...client.DeleteOption) error {
					return deleteErr
				},
			}).Build()
		handler.cachedClient = fakeClient
		handler.runaiClient = fakeClient
		updates := []workloads_openapi.WorkloadsSyncInfo{{
			Workload: &workloads_openapi.SupersetWorkload{
				Name:         "legacy-training",
				ProjectId:    "project-1",
				DesiredPhase: workloads_openapi.WORKLOADDESIREDPHASE_DELETED,
			},
		}}
		wg := &sync.WaitGroup{}
		wg.Add(1)

		perWorkloadErrors, err := handler.ProcessWorkloads(context.Background(), updates, wg)

		Expect(err).To(HaveOccurred())
		Expect(perWorkloadErrors).To(ConsistOf(MatchError(deleteErr)))
		event := &corev1.Event{}
		Expect(fakeClient.Get(context.Background(), client.ObjectKey{
			Name: "legacy-training", Namespace: "project-namespace",
		}, event)).To(Succeed())
		Expect(event.Reason).To(Equal("Deletion failed"))
	})
})
