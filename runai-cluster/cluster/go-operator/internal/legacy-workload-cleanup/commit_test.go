/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	advicerestartcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/advice-restart-controller"
	inferencecontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/inference-workload-controller"
	jobcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/job-controller"
	workloadcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-controller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("Atomic flags-off commit", func() {
	It("disables all three legacy controllers in one update", func() {
		k8sClient := runaiConfigClient(enabledRunaiConfig())
		Expect(commitControllerDisable(context.Background(), k8sClient)).To(Succeed())
		expectControllerState(k8sClient, controllersDisabled)
	})

	It("treats an already committed update as success", func() {
		config := enabledRunaiConfig()
		setControllerFlags(config, false)
		k8sClient := runaiConfigClient(config)
		Expect(commitControllerDisable(context.Background(), k8sClient)).To(Succeed())
	})

	It("retries an optimistic-concurrency conflict", func() {
		base := runaiConfigClient(enabledRunaiConfig())
		conflicting := &conflictOnceClient{Client: base}
		Expect(commitControllerDisable(context.Background(), conflicting)).To(Succeed())
		Expect(conflicting.updateCalls).To(Equal(2))
		expectControllerState(base, controllersDisabled)
	})

	It("leaves every flag enabled when the update fails", func() {
		failed := runaiConfigClientWithInterceptor(enabledRunaiConfig(), interceptor.Funcs{
			Update: func(context.Context, client.WithWatch, client.Object, ...client.UpdateOption) error {
				return errors.New("update failed")
			},
		})
		Expect(commitControllerDisable(context.Background(), failed)).To(MatchError(ContainSubstring("update failed")))
		expectControllerState(failed, controllersEnabled)
	})

	It("does not repeat the Operator readiness gate during the flags-off commit", func() {
		config := enabledRunaiConfig()
		config.Status.ObservedGeneration = config.Generation - 1
		k8sClient := runaiConfigClient(config)
		Expect(commitControllerDisable(context.Background(), k8sClient)).To(Succeed())
		expectControllerState(k8sClient, controllersDisabled)
	})

	It("does not use Advice Restart Controller state as a commit gate", func() {
		config := enabledRunaiConfig()
		config.Spec.AdviceRestartController.Enabled = ptr.To(true)
		k8sClient := runaiConfigClient(config)
		Expect(commitControllerDisable(context.Background(), k8sClient)).To(Succeed())
		expectControllerState(k8sClient, controllersDisabled)
	})

})

type conflictOnceClient struct {
	client.Client
	updateCalls int
}

func (c *conflictOnceClient) Update(ctx context.Context, object client.Object, options ...client.UpdateOption) error {
	c.updateCalls++
	if c.updateCalls == 1 {
		return apierrors.NewConflict(schema.GroupResource{Group: "run.ai", Resource: "runaiconfigs"},
			object.GetName(), errors.New("conflict"))
	}
	return c.Client.Update(ctx, object, options...)
}

func enabledRunaiConfig() *runaiv1.RunaiConfig {
	return &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{Name: runaiv1.DefaultName, Namespace: runaiv1.DefaultNamespace, Generation: 1},
		Spec: runaiv1.RunaiConfigSpec{
			WorkloadController:          &workloadcontroller.WorkloadController{Enabled: ptr.To(true)},
			InferenceWorkloadController: &inferencecontroller.InferenceWorkloadController{Enabled: ptr.To(true)},
			JobController:               &jobcontroller.JobController{Enabled: ptr.To(true)},
			AdviceRestartController:     &advicerestartcontroller.AdviceRestartController{Enabled: ptr.To(false)},
		},
		Status: readyRunaiConfig(1).Status,
	}
}

func runaiConfigClient(config *runaiv1.RunaiConfig) client.Client {
	return runaiConfigClientWithInterceptor(config, interceptor.Funcs{})
}

func runaiConfigClientWithInterceptor(config *runaiv1.RunaiConfig, funcs interceptor.Funcs) client.Client {
	scheme := runtime.NewScheme()
	Expect(runaiv1.AddToScheme(scheme)).To(Succeed())
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(config).WithInterceptorFuncs(funcs).Build()
}

func getRunaiConfig(k8sClient client.Client) *runaiv1.RunaiConfig {
	config := &runaiv1.RunaiConfig{}
	Expect(k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: runaiv1.DefaultNamespace, Name: runaiv1.DefaultName,
	}, config)).To(Succeed())
	return config
}

func expectControllerState(k8sClient client.Client, expected controllerState) {
	state, err := classifyControllerState(getRunaiConfig(k8sClient))
	Expect(err).NotTo(HaveOccurred())
	Expect(state).To(Equal(expected))
}
