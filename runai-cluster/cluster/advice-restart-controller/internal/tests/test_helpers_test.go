/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"time"

	"k8s.io/client-go/tools/record"

	heimdallv1alpha1 "github.com/dsx-ai-factory/nvidia-recovery/advice-operator/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/advice-restart-controller/internal/workloadtype"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	timeout             = 10 * time.Second
	consistentlyTimeout = 5 * time.Second
	interval            = 250 * time.Millisecond
)

func createNamespace(ctx context.Context, c client.Client, name string) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	ExpectWithOffset(1, c.Create(ctx, ns)).To(Succeed())
	return ns
}

func createAdvice(ctx context.Context, c client.Client, name, ns, jobName string,
	state heimdallv1alpha1.JobState, exclusions ...string) *heimdallv1alpha1.Advice {
	var nodeConstraints *heimdallv1alpha1.NodeConstraints
	if len(exclusions) > 0 {
		nodeConstraints = &heimdallv1alpha1.NodeConstraints{
			Exclusions: exclusions,
		}
	}

	advice := &heimdallv1alpha1.Advice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: heimdallv1alpha1.AdviceSpec{
			Job: heimdallv1alpha1.AdviceJobReference{
				HeimdallID:   "heimdall-" + name,
				APIVersion:   "kubeflow.org/v1",
				Kind:         "PyTorchJob",
				Name:         jobName,
				JobID:        "scheduler-id-" + name,
				AttemptIndex: 0,
			},
			Advice: heimdallv1alpha1.AdviceRecommendation{
				State:           &state,
				NodeConstraints: nodeConstraints,
			},
		},
	}
	ExpectWithOffset(1, c.Create(ctx, advice)).To(Succeed())

	created := &heimdallv1alpha1.Advice{}
	EventuallyWithOffset(1, func() error {
		return c.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, created)
	}, timeout, interval).Should(Succeed())

	return created
}

// createPyTorchJob creates a typed PyTorchJob and returns it with the server-assigned UID.
func createPyTorchJob(ctx context.Context, c client.Client, name, ns string,
	ownerName, ownerUID string) *kubeflowv1.PyTorchJob {
	job := &kubeflowv1.PyTorchJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: kubeflowv1.PyTorchJobSpec{
			PyTorchReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				kubeflowv1.PyTorchJobReplicaTypeMaster: {
					Replicas: ptr.To[int32](1),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "pytorch", Image: "test"}},
						},
					},
				},
				kubeflowv1.PyTorchJobReplicaTypeWorker: {
					Replicas: ptr.To[int32](1),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "pytorch", Image: "test"}},
						},
					},
				},
			},
		},
	}

	if ownerName != "" {
		job.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion: workloadtype.GVKWorkloadRunner.GroupVersion().String(),
				Kind:       workloadtype.GVKWorkloadRunner.Kind,
				Name:       ownerName,
				UID:        types.UID(ownerUID),
			},
		})
	}

	ExpectWithOffset(1, c.Create(ctx, job)).To(Succeed())

	created := &kubeflowv1.PyTorchJob{}
	EventuallyWithOffset(1, func() error {
		return c.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, created)
	}, timeout, interval).Should(Succeed())

	return created
}

// createWLR creates a WorkloadRunner and returns it with the server UID
func createWLR(
	ctx context.Context,
	c client.Client,
	name, ns, phase string,
	nodeAffinity *corev1.NodeAffinity,
) *dromosv1alpha1.WorkloadRunner {
	wlr := &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: runtime.RawExtension{
				Raw: fmt.Appendf(nil, `{"apiVersion":"kubeflow.org/v1","kind":"PyTorchJob","metadata":{"name":%q}}`, name),
			},
		},
	}
	if nodeAffinity != nil {
		wlr.Spec.Instructions = &dromosv1alpha1.Instructions{
			NodeAffinity: &dromosv1alpha1.NodeAffinityConfig{NodeAffinity: nodeAffinity},
		}
	}
	ExpectWithOffset(1, c.Create(ctx, wlr)).To(Succeed())

	created := &dromosv1alpha1.WorkloadRunner{}
	EventuallyWithOffset(1, func() error {
		return c.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, created)
	}, timeout, interval).Should(Succeed())

	if phase != "" {
		created.Status.Phase = dromosv1alpha1.WorkloadRunnerPhase(phase)
		ExpectWithOffset(1, c.Status().Update(ctx, created)).To(Succeed())
	}

	return created
}

func createRestartRequestWithStatus(ctx context.Context, c client.Client, name, ns string,
	labels map[string]string, phase runv1alpha1.RestartRequestPhase) *runv1alpha1.RestartRequest {
	rr := &runv1alpha1.RestartRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: runv1alpha1.RestartRequestSpec{
			WorkloadRef: runv1alpha1.WorkloadReference{
				GVK: metav1.GroupVersionKind{
					Group:   "run.ai",
					Version: "v1alpha1",
					Kind:    "WorkloadRunner",
				},
				Name: "dummy-workload",
			},
		},
	}
	ExpectWithOffset(1, c.Create(ctx, rr)).To(Succeed())

	if phase != "" {
		EventuallyWithOffset(1, func() error {
			if err := c.Get(ctx, types.NamespacedName{Name: rr.Name, Namespace: ns}, rr); err != nil {
				return err
			}
			rr.Status.Phase = phase
			return c.Status().Update(ctx, rr)
		}, timeout, interval).Should(Succeed())
	}

	return rr
}

func drainEvents(recorder *record.FakeRecorder) {
	for len(recorder.Events) > 0 {
		<-recorder.Events
	}
}

func deleteAllAdvices(ctx context.Context, c client.Client, ns string) {
	ExpectWithOffset(1, c.DeleteAllOf(ctx, &heimdallv1alpha1.Advice{}, client.InNamespace(ns))).To(Succeed())
}

func deleteAllRestartRequests(ctx context.Context, c client.Client, ns string) {
	ExpectWithOffset(1, c.DeleteAllOf(ctx, &runv1alpha1.RestartRequest{}, client.InNamespace(ns))).To(Succeed())
}

func deleteAllPyTorchJobs(ctx context.Context, c client.Client, ns string) {
	ExpectWithOffset(1, c.DeleteAllOf(ctx, &kubeflowv1.PyTorchJob{}, client.InNamespace(ns))).To(Succeed())
}

func deleteAllWLRs(ctx context.Context, c client.Client, ns string) {
	ExpectWithOffset(1, c.DeleteAllOf(ctx, &dromosv1alpha1.WorkloadRunner{}, client.InNamespace(ns))).To(Succeed())
}
