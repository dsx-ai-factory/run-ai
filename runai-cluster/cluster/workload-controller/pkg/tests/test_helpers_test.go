/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"time"

	. "github.com/onsi/gomega"
	distw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	mpi "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

var webhookCertDir string

func drainEvents(recorder *record.FakeRecorder) {
	for len(recorder.Events) > 0 {
		<-recorder.Events
	}
}

func deleteAllForType[T client.ObjectList](ctx context.Context, k8sClient client.Client, objs T) {
	Eventually(func() error {
		return k8sClient.List(ctx, objs)
	}, timeout, interval).Should(Succeed())
	items, err := meta.ExtractList(objs)
	Expect(err).NotTo(HaveOccurred())
	for _, obj := range items {
		clientObj, ok := obj.(client.Object)
		Expect(ok).To(BeTrue())
		_ = k8sClient.Delete(ctx, clientObj)
	}
}

func deleteAllTestResources(ctx context.Context, k8sClient client.Client) {
	deleteAllForType(ctx, k8sClient, &distw.DistributedWorkloadList{})
	deleteAllForType(ctx, k8sClient, &mpi.MPIJobList{})
}
