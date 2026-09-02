/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networkstatus

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	testDistributedInferenceNamespace = "runai-team-a"
	testDistributedInferenceRunner    = "model-runner"
	testDistributedInferenceService   = "dinfw-model-sp-0"
	testDistributedInferenceURL       = "http://dinfw-model-sp-0.runai-team-a.svc.cluster.local"
)

func distributedInferenceExternalWorkload(urls ...string) *v1alpha1.ExternalWorkload {
	return &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ew-model",
			Namespace: testDistributedInferenceNamespace,
			Labels: map[string]string{
				commonconstants.WorkloadRunnerNameLabel: testDistributedInferenceRunner,
			},
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			OriginalWorkloadUID: types.UID("model-uid"),
			Type:                distributedInferenceWorkloadType,
			Urls:                urls,
		},
	}
}

func distributedInferenceEndpointWorkloadRunner(externalAccess bool) *dromosv1alpha1.WorkloadRunner {
	endpoint := endpoints.Endpoint{Services: []endpoints.Service{{
		Create: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: testDistributedInferenceService}},
	}}}
	if externalAccess {
		endpoint.ExternalAccess = &endpoints.ExternalAccess{}
	}
	workloadRunner := &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{Name: testDistributedInferenceRunner, Namespace: testDistributedInferenceNamespace},
	}
	if err := endpoints.Set(workloadRunner, &endpoints.Config{Endpoints: []endpoints.Endpoint{endpoint}}); err != nil {
		panic(err)
	}
	return workloadRunner
}

func distributedInferenceWorkloadRunner(externalAccess bool) *dromosv1alpha1.WorkloadRunner {
	networking := &dromosv1alpha1.NetworkingConfig{
		Services: []dromosv1alpha1.NetworkingServiceItem{{
			Create: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: testDistributedInferenceService}},
		}},
	}
	if externalAccess {
		networking.ExternalAccess = &dromosv1alpha1.NetworkingExternalAccessConfig{}
	}
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{Name: testDistributedInferenceRunner, Namespace: testDistributedInferenceNamespace},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Instructions: &dromosv1alpha1.Instructions{Networking: networking},
		},
	}
}

func distributedInferenceService(port int32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: testDistributedInferenceService, Namespace: testDistributedInferenceNamespace},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"app": "model"},
			Ports: []corev1.ServicePort{{
				Name: distributedInferenceServicePort,
				Port: port,
			}},
		},
	}
}

func distributedInferencePod() corev1.Pod {
	return corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: "model-0", Namespace: testDistributedInferenceNamespace, Labels: map[string]string{"app": "model"},
	}}
}

var _ = Describe("distributed inference internal URL", func() {
	It("publishes the configured internal serving Service through spec URLs and endpoint status", func() {
		scheme := testScheme()
		ew := distributedInferenceExternalWorkload()
		fc := newFakeClientWithIndex(scheme, ew, distributedInferenceWorkloadRunner(false), distributedInferenceService(80))
		updater := NewNetworkStatusUpdater(fc)

		Expect(updater.recomputeExternalWorkloadEndpoints(context.Background(), ew, []corev1.Pod{distributedInferencePod()})).To(Succeed())

		actual := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(context.Background(), types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}, actual)).To(Succeed())
		Expect(actual.Spec.Urls).To(Equal([]string{testDistributedInferenceURL}))
		Expect(actual.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		Expect(actual.Status.NetworkStatus.ExternalEndpoints[0].Url).To(Equal(testDistributedInferenceURL))
	})

	It("retries a failed URL patch without mutating the queued snapshot", func() {
		ctx := context.Background()
		scheme := testScheme()
		ew := distributedInferenceExternalWorkload()
		patchAttempts := 0
		fc := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(ew, distributedInferenceWorkloadRunner(false), distributedInferenceService(80)).
			WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
			WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(ctx context.Context, interceptedClient client.WithWatch, obj client.Object,
					patch client.Patch, opts ...client.PatchOption,
				) error {
					patchAttempts++
					if patchAttempts == 1 {
						return errors.New("transient spec patch failure")
					}
					return interceptedClient.Patch(ctx, obj, patch, opts...)
				},
			}).
			Build()
		updater := NewNetworkStatusUpdater(fc)
		queue := updater.TasksQueue()
		key := types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}
		task := *utils.NewExternalWorkloadTask(
			utils.AddOperation,
			key,
			v1alpha1.SchemeGroupVersion.WithKind("ExternalWorkload"),
		).SetObject(ew.DeepCopy())

		queue.Add(task)
		firstAttempt, shutdown := queue.Get()
		Expect(shutdown).To(BeFalse())
		updater.handleTask(ctx, firstAttempt)
		queue.Done(firstAttempt)

		Expect(patchAttempts).To(Equal(1))
		queuedExternalWorkload, ok := firstAttempt.Object.(*v1alpha1.ExternalWorkload)
		Expect(ok).To(BeTrue())
		Expect(queuedExternalWorkload.Spec.Urls).To(BeEmpty())
		actual := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, key, actual)).To(Succeed())
		Expect(actual.Spec.Urls).To(BeEmpty())

		Eventually(queue.Len).Should(Equal(1))
		retry, shutdown := queue.Get()
		Expect(shutdown).To(BeFalse())
		updater.handleTask(ctx, retry)
		queue.Done(retry)

		Expect(patchAttempts).To(Equal(2))
		Expect(queuedExternalWorkload.Spec.Urls).To(Equal([]string{testDistributedInferenceURL}))
		Expect(fc.Get(ctx, key, actual)).To(Succeed())
		Expect(actual.Spec.Urls).To(Equal([]string{testDistributedInferenceURL}))
		Expect(queue.Len()).To(BeZero())
	})

	It("includes a non-default Service port", func() {
		scheme := testScheme()
		ew := distributedInferenceExternalWorkload()
		fc := newFakeClientWithIndex(scheme, ew, distributedInferenceWorkloadRunner(false), distributedInferenceService(8080))
		updater := NewNetworkStatusUpdater(fc)

		Expect(updater.reconcileDistributedInferenceInternalURL(context.Background(), ew)).To(Succeed())

		actual := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(context.Background(), types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}, actual)).To(Succeed())
		Expect(actual.Spec.Urls).To(Equal([]string{testDistributedInferenceURL + ":8080"}))
	})

	It("publishes the internal URL from generated-native endpoint configuration", func() {
		scheme := testScheme()
		ew := distributedInferenceExternalWorkload()
		fc := newFakeClientWithIndex(scheme, ew, distributedInferenceEndpointWorkloadRunner(false), distributedInferenceService(80))
		updater := NewNetworkStatusUpdater(fc)

		Expect(updater.recomputeExternalWorkloadEndpoints(context.Background(), ew, []corev1.Pod{distributedInferencePod()})).To(Succeed())

		actual := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(context.Background(), types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}, actual)).To(Succeed())
		Expect(actual.Spec.Urls).To(Equal([]string{testDistributedInferenceURL}))
		Expect(actual.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		Expect(actual.Status.NetworkStatus.ExternalEndpoints[0].Url).To(Equal(testDistributedInferenceURL))
	})

	It("removes only the managed internal URL when external access is configured", func() {
		scheme := testScheme()
		externalURL := "https://model.example.com"
		ew := distributedInferenceExternalWorkload(testDistributedInferenceURL, externalURL)
		fc := newFakeClientWithIndex(scheme, ew, distributedInferenceWorkloadRunner(true), distributedInferenceService(80))
		updater := NewNetworkStatusUpdater(fc)

		Expect(updater.reconcileDistributedInferenceInternalURL(context.Background(), ew)).To(Succeed())

		actual := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(context.Background(), types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}, actual)).To(Succeed())
		Expect(actual.Spec.Urls).To(Equal([]string{externalURL}))
	})

	It("removes only the managed internal URL when generated-native external access is configured", func() {
		scheme := testScheme()
		externalURL := "https://model.example.com"
		ew := distributedInferenceExternalWorkload(testDistributedInferenceURL, externalURL)
		fc := newFakeClientWithIndex(scheme, ew, distributedInferenceEndpointWorkloadRunner(true), distributedInferenceService(80))
		updater := NewNetworkStatusUpdater(fc)

		Expect(updater.reconcileDistributedInferenceInternalURL(context.Background(), ew)).To(Succeed())

		actual := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(context.Background(), types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}, actual)).To(Succeed())
		Expect(actual.Spec.Urls).To(Equal([]string{externalURL}))
	})

	It("preserves a user-provided URL instead of adding the internal fallback", func() {
		scheme := testScheme()
		userProvidedURL := "https://custom.example.com/models/model-a"
		ew := distributedInferenceExternalWorkload(testDistributedInferenceURL, userProvidedURL)
		fc := newFakeClientWithIndex(scheme, ew, distributedInferenceWorkloadRunner(false), distributedInferenceService(80))
		updater := NewNetworkStatusUpdater(fc)

		Expect(updater.reconcileDistributedInferenceInternalURL(context.Background(), ew)).To(Succeed())

		actual := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(context.Background(), types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}, actual)).To(Succeed())
		Expect(actual.Spec.Urls).To(Equal([]string{userProvidedURL}))
	})

	It("does not change other workload types", func() {
		scheme := testScheme()
		ew := distributedInferenceExternalWorkload()
		ew.Spec.Type = "Workspace"
		fc := newFakeClientWithIndex(scheme, ew, distributedInferenceWorkloadRunner(false), distributedInferenceService(80))
		updater := NewNetworkStatusUpdater(fc)

		Expect(updater.reconcileDistributedInferenceInternalURL(context.Background(), ew)).To(Succeed())
		Expect(ew.Spec.Urls).To(BeEmpty())
	})
})
