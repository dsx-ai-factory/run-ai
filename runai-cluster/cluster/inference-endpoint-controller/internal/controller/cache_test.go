/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	inferencev1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/inference-endpoint-controller/api/v1alpha1"
)

var _ = Describe("Manager cache restricted to the rendered children", Ordered, func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		mgr    ctrl.Manager
		ns     string
	)

	// The controller name must be unique per process, so one manager serves both specs.
	BeforeAll(func() {
		ctx, cancel = context.WithCancel(context.Background())

		var err error
		mgr, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  scheme.Scheme,
			Cache:   ChildCacheOptions(),
			Metrics: metricsserver.Options{BindAddress: "0"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect((&InferenceEndpointReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), MetricsNamespace: metricsNamespace}).SetupWithManager(mgr)).To(Succeed())
		go func() {
			defer GinkgoRecover()
			Expect(mgr.Start(ctx)).To(Succeed())
		}()
		Expect(mgr.GetCache().WaitForCacheSync(ctx)).To(BeTrue())
	})

	AfterAll(func() { cancel() })

	BeforeEach(func() {
		ns = fmt.Sprintf("cache-%d", time.Now().UnixNano())
		Expect(k8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})).To(Succeed())
	})

	It("does not cache objects the controller did not render", func() {
		unrelated := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: ns}}
		Expect(k8sClient.Create(ctx, unrelated)).To(Succeed())

		key := types.NamespacedName{Name: "unrelated", Namespace: ns}
		Consistently(func() bool {
			return errors.IsNotFound(mgr.GetClient().Get(ctx, key, &corev1.ConfigMap{}))
		}, 2*time.Second, 200*time.Millisecond).Should(BeTrue())
		Expect(mgr.GetAPIReader().Get(ctx, key, &corev1.ConfigMap{})).To(Succeed())
	})

	It("still repairs a deleted child through the restricted watches", func() {
		endpoint := &inferencev1alpha1.InferenceEndpoint{
			ObjectMeta: metav1.ObjectMeta{Name: endpointName, Namespace: ns},
			Spec: inferencev1alpha1.InferenceEndpointSpec{
				Workload: inferencev1alpha1.WorkloadSpec{
					Selector: inferencev1alpha1.WorkloadSelector{MatchLabels: map[string]string{testAppLabel: "vllm-llama3"}},
					Port:     8000,
				},
				Routing: inferencev1alpha1.RoutingSpec{
					GatewayRef: inferencev1alpha1.GatewayReference{Name: gatewayName},
					Hostnames:  []string{"llama3.test"},
				},
				Epp: inferencev1alpha1.EppSpec{Profile: inferencev1alpha1.EppProfileGIEV15},
			},
		}
		Expect(k8sClient.Create(ctx, endpoint)).To(Succeed())

		key := types.NamespacedName{Name: eppName, Namespace: ns}
		configMap := &corev1.ConfigMap{}
		Eventually(func() error { return k8sClient.Get(ctx, key, configMap) }, 30*time.Second, 200*time.Millisecond).Should(Succeed())
		firstUID := configMap.UID

		Expect(k8sClient.Delete(ctx, configMap)).To(Succeed())
		Eventually(func() (types.UID, error) {
			recreated := &corev1.ConfigMap{}
			err := k8sClient.Get(ctx, key, recreated)
			return recreated.UID, err
		}, 30*time.Second, 200*time.Millisecond).ShouldNot(Equal(firstUID))
	})
})
