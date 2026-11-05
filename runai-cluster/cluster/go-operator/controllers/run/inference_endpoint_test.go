/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run_test

import (
	"context"

	run "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/controllers/run"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Inference endpoint registration", func() {
	It("is opt-in, runs on the workload cluster once enabled, and honors explicit disablement", func() {
		config := &runv1.RunaiConfig{}
		config.SetDefaultsWhereNeeded()
		Expect(config.Spec.InferenceEndpointController.Enabled).To(HaveValue(BeFalse()))
		// Gateway API and GIE CRDs are prerequisites the chart does not install.
		config.Spec.InferenceEndpointController.Enabled = ptr.To(true)
		client := fake.NewClientBuilder().Build()
		hybrid := kc.New(client, client, nil)
		managed := run.NewManagedOperands(context.Background(), hybrid, hybrid, config)
		var remoteNames, localNames []string
		for _, operand := range managed.GetRemoteClusterEnabled() {
			remoteNames = append(remoteNames, operand.Name())
		}
		for _, operand := range managed.GetLocalClusterOperands() {
			localNames = append(localNames, operand.Name())
		}
		Expect(remoteNames).To(ContainElement("inference-endpoint-controller"))
		Expect(localNames).NotTo(ContainElement("inference-endpoint-controller"))
		config.Spec.InferenceEndpointController.Enabled = ptr.To(false)
		remoteNames = nil
		for _, operand := range managed.GetRemoteClusterEnabled() {
			remoteNames = append(remoteNames, operand.Name())
		}
		Expect(remoteNames).NotTo(ContainElement("inference-endpoint-controller"))
	})
})
