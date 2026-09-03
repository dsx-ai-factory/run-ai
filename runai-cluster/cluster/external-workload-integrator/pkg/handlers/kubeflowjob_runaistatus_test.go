/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	statuscontract "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/status/contract"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kubeflowv2beta1 "github.com/kubeflow/mpi-operator/pkg/apis/kubeflow/v2beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("KubeflowJobTopOwnerHandler.applyRunaiDistributedStatus", func() {
	const (
		ns      = "team-a"
		jobName = "pt-job"
		wlrName = "wlr-1"
	)
	jobUID := types.UID("job-uid-abc")

	newEW := func(runnerName ...string) *v1alpha1.ExternalWorkload {
		labels := map[string]string{}
		if len(runnerName) > 0 && runnerName[0] != "" {
			labels[commonconstants.WorkloadRunnerNameLabel] = runnerName[0]
		}
		return &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "ew-1",
				Namespace:   ns,
				Annotations: map[string]string{},
				Labels:      labels,
			},
		}
	}

	It("clears stale annotation and pod counts when the CR is not WLR-owned", func() {
		job := makeJob(jobName, ns, jobUID)
		c := fakeclient.NewClientBuilder().WithScheme(newTestScheme()).Build()
		ew := newEW()
		ew.Annotations[commonconstants.EwiStatusAnnotation] = "true"
		staleFields := `{"readyPods":3,"runningPods":3,"otherKey":"preserved"}`
		ew.Spec.AdditionalFields = &staleFields

		h := &KubeflowJobTopOwnerHandler{cachedClient: c}
		h.applyRunaiDistributedStatus(context.Background(), job,
			&KubeflowJobFieldsForParsing{jobStatus: &kubeflowv1.JobStatus{}}, ew)

		Expect(ew.Annotations).NotTo(HaveKey(commonconstants.EwiStatusAnnotation),
			"non-WLR-owned CRs must not get the authoritative annotation")
		Expect(decodedAdditionalFields(ew)).To(Equal(map[string]any{"otherKey": "preserved"}))
	})

	It("clears ewi-status annotation on pod-list error", func() {
		job := makeJob(jobName, ns, jobUID)
		c := fakeclient.NewClientBuilder().WithScheme(newTestScheme()).WithInterceptorFuncs(interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return apierrors.NewServiceUnavailable("pod list failed")
			},
		}).Build()
		ew := newEW(wlrName)
		ew.Annotations[commonconstants.EwiStatusAnnotation] = "true"
		staleFields := `{"readyPods":3,"runningPods":3,"otherKey":"preserved"}`
		ew.Spec.AdditionalFields = &staleFields

		h := &KubeflowJobTopOwnerHandler{cachedClient: c}
		h.applyRunaiDistributedStatus(context.Background(), job,
			&KubeflowJobFieldsForParsing{jobStatus: &kubeflowv1.JobStatus{}}, ew)

		Expect(ew.Annotations).NotTo(HaveKey(commonconstants.EwiStatusAnnotation),
			"pod-list failure must clear the stale annotation so cluster-sync falls back to WLR phase")
		Expect(decodedAdditionalFields(ew)).To(Equal(map[string]any{"otherKey": "preserved"}))
	})

	It("clears ewi-status annotation on WorkloadRunner fetch error", func() {
		job := makeJob(jobName, ns, jobUID)
		c := fakeclient.NewClientBuilder().WithScheme(newTestScheme()).WithInterceptorFuncs(interceptor.Funcs{
			Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
				return apierrors.NewServiceUnavailable("wlr get failed")
			},
		}).Build()
		ew := newEW(wlrName)
		ew.Annotations[commonconstants.EwiStatusAnnotation] = "true"
		staleFields := `{"readyPods":3,"runningPods":3,"otherKey":"preserved"}`
		ew.Spec.AdditionalFields = &staleFields

		h := &KubeflowJobTopOwnerHandler{cachedClient: c}
		h.applyRunaiDistributedStatus(context.Background(), job,
			&KubeflowJobFieldsForParsing{jobStatus: &kubeflowv1.JobStatus{}}, ew)

		Expect(ew.Annotations).NotTo(HaveKey(commonconstants.EwiStatusAnnotation))
		Expect(decodedAdditionalFields(ew)).To(Equal(map[string]any{"otherKey": "preserved"}))
	})

	It("sets Running phase, annotation, and pod counts in additionalFields (happy path)", func() {
		job := makeJob(jobName, ns, jobUID)
		pod := readyPod("pod-0", ns, jobUID)
		wlr := emptyWLR(wlrName, ns)

		c := fakeclient.NewClientBuilder().
			WithScheme(newTestScheme()).
			WithObjects(wlr, &pod).
			Build()
		ew := newEW(wlrName)

		h := &KubeflowJobTopOwnerHandler{cachedClient: c}
		h.applyRunaiDistributedStatus(context.Background(), job,
			&KubeflowJobFieldsForParsing{jobStatus: &kubeflowv1.JobStatus{}}, ew)

		Expect(ew.Annotations[commonconstants.EwiStatusAnnotation]).To(Equal("true"))
		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWRunningPhase))

		Expect(ew.Spec.AdditionalFields).NotTo(BeNil())
		fields := map[string]interface{}{}
		Expect(json.Unmarshal([]byte(*ew.Spec.AdditionalFields), &fields)).To(Succeed())
		Expect(fields).To(HaveKey(statuscontract.AdditionalFieldReadyPods))
	})

	It("maps Suspended legacy phase to ExWStoppedPhase", func() {
		suspend := true
		job := makeJob(jobName, ns, jobUID)
		wlr := emptyWLR(wlrName, ns)

		c := fakeclient.NewClientBuilder().
			WithScheme(newTestScheme()).
			WithObjects(wlr).
			Build()
		ew := newEW(wlrName)

		fields := &KubeflowJobFieldsForParsing{
			jobStatus: &kubeflowv1.JobStatus{
				Conditions: []kubeflowv1.JobCondition{
					{Type: kubeflowv1.JobSuspended, Status: corev1.ConditionTrue},
				},
			},
			runPolicy: &kubeflowv1.RunPolicy{Suspend: &suspend},
		}

		h := &KubeflowJobTopOwnerHandler{cachedClient: c}
		h.applyRunaiDistributedStatus(context.Background(), job, fields, ew)

		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWStoppedPhase))
		Expect(ew.Annotations[commonconstants.EwiStatusAnnotation]).To(Equal("true"))
	})

	It("adds a runtime WorkloadIssue condition from failed pods when the WLR has no critical issue", func() {
		job := makeJob(jobName, ns, jobUID)
		job.Status.Conditions = []kubeflowv1.JobCondition{
			{Type: kubeflowv1.JobFailed, Status: corev1.ConditionTrue},
		}
		pod := failedPod("pod-0", ns, jobUID)
		wlr := emptyWLR(wlrName, ns)

		c := fakeclient.NewClientBuilder().
			WithScheme(newTestScheme()).
			WithObjects(wlr, &pod).
			Build()
		ew := newEW(wlrName)

		h := &KubeflowJobTopOwnerHandler{cachedClient: c}
		h.applyRunaiDistributedStatus(context.Background(), job,
			&KubeflowJobFieldsForParsing{jobStatus: &job.Status}, ew)

		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWFailedPhase))
		Expect(ew.Status.ObservedStatus.Message).To(Equal("container 'main' failed: boom (pod: pod-0)"))
		Expect(ew.Status.ObservedStatus.Conditions).To(ContainElement(v1alpha1.ObservedCondition{
			Type:    commonconstants.WorkloadIssueConditionType,
			Status:  corev1.ConditionTrue,
			Reason:  "Compute",
			Message: "container 'main' failed: boom (pod: pod-0)",
		}))
	})

	It("sets Deleting when the Kubeflow CR is terminating", func() {
		now := metav1.Now()
		job := makeJob(jobName, ns, jobUID)
		job.DeletionTimestamp = &now
		wlr := emptyWLR(wlrName, ns)

		c := fakeclient.NewClientBuilder().
			WithScheme(newTestScheme()).
			WithObjects(wlr).
			Build()
		ew := newEW(wlrName)

		h := &KubeflowJobTopOwnerHandler{cachedClient: c}
		h.applyRunaiDistributedStatus(context.Background(), job,
			&KubeflowJobFieldsForParsing{jobStatus: &kubeflowv1.JobStatus{}}, ew)

		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWDeletingPhase))
		Expect(ew.Annotations[commonconstants.EwiStatusAnnotation]).To(Equal("true"))
	})

	It("sets workload-controller-style conditions when required MNNVL is not ready", func() {
		job := makeJob(jobName, ns, jobUID)
		pod := initializingPod("pod-0", ns, jobUID)
		wlr := emptyWLR(wlrName, ns)
		wlr.Spec.Instructions = &dromosv1alpha1.Instructions{
			GpuOptimizations: &dromosv1alpha1.GpuOptimizations{
				Mnnvl: &dromosv1alpha1.MnnvlConfig{Mode: dromosv1alpha1.MnnvlModeRequired},
			},
		}
		wlr.Status.Plugins = map[string]*dromosv1alpha1.PluginStatus{
			"mnnvl": {Phase: dromosv1alpha1.PluginPhaseCreating},
		}

		c := fakeclient.NewClientBuilder().
			WithScheme(newTestScheme()).
			WithObjects(wlr, &pod).
			Build()
		ew := newEW(wlrName)

		h := &KubeflowJobTopOwnerHandler{cachedClient: c}
		h.applyRunaiDistributedStatus(context.Background(), job,
			&KubeflowJobFieldsForParsing{jobStatus: &job.Status}, ew)

		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWInitializingPhase))
		Expect(ew.Status.ObservedStatus.Message).To(Equal(statuscontract.MsgInitializingSinglePod))
		Expect(ew.Status.ObservedStatus.Conditions).To(ContainElement(v1alpha1.ObservedCondition{
			Type:    statuscontract.ConditionTypeReady,
			Status:  corev1.ConditionFalse,
			Reason:  statuscontract.ReasonMnnvlNotReady,
			Message: statuscontract.MsgMnnvlNotReady,
		}))
		Expect(ew.Status.ObservedStatus.Conditions).To(ContainElement(v1alpha1.ObservedCondition{
			Type:    statuscontract.ConditionTypeGPUNetworkAccelerated,
			Status:  corev1.ConditionFalse,
			Reason:  statuscontract.ReasonMnnvlNotReady,
			Message: statuscontract.MsgMnnvlNotReady,
		}))
	})

	It("sets GPU network acceleration true when required MNNVL is ready", func() {
		job := makeJob(jobName, ns, jobUID)
		pod := readyPod("pod-0", ns, jobUID)
		wlr := emptyWLR(wlrName, ns)
		wlr.Spec.Instructions = &dromosv1alpha1.Instructions{
			GpuOptimizations: &dromosv1alpha1.GpuOptimizations{
				Mnnvl: &dromosv1alpha1.MnnvlConfig{Mode: dromosv1alpha1.MnnvlModeRequired},
			},
		}
		wlr.Status.Plugins = map[string]*dromosv1alpha1.PluginStatus{
			"mnnvl": {Phase: dromosv1alpha1.PluginPhaseReady},
		}

		c := fakeclient.NewClientBuilder().
			WithScheme(newTestScheme()).
			WithObjects(wlr, &pod).
			Build()
		ew := newEW(wlrName)

		h := &KubeflowJobTopOwnerHandler{cachedClient: c}
		h.applyRunaiDistributedStatus(context.Background(), job,
			&KubeflowJobFieldsForParsing{jobStatus: &job.Status}, ew)

		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWRunningPhase))
		Expect(ew.Status.ObservedStatus.Message).To(BeEmpty())
		Expect(ew.Status.ObservedStatus.Conditions).To(ContainElement(v1alpha1.ObservedCondition{
			Type:   statuscontract.ConditionTypeGPUNetworkAccelerated,
			Status: corev1.ConditionTrue,
		}))
	})

	It("uses a generic WorkloadIssue when a failed job has no pod-level message", func() {
		job := makeJob(jobName, ns, jobUID)
		job.Status.Conditions = []kubeflowv1.JobCondition{
			{Type: kubeflowv1.JobFailed, Status: corev1.ConditionTrue},
		}
		wlr := emptyWLR(wlrName, ns)

		c := fakeclient.NewClientBuilder().
			WithScheme(newTestScheme()).
			WithObjects(wlr).
			Build()
		ew := newEW(wlrName)

		h := &KubeflowJobTopOwnerHandler{cachedClient: c}
		h.applyRunaiDistributedStatus(context.Background(), job,
			&KubeflowJobFieldsForParsing{jobStatus: &job.Status}, ew)

		Expect(ew.Status.ObservedStatus.Conditions).To(ContainElement(v1alpha1.ObservedCondition{
			Type:    commonconstants.WorkloadIssueConditionType,
			Status:  corev1.ConditionTrue,
			Reason:  "Compute",
			Message: statuscontract.MsgGenericJobFailed,
		}))
	})
})

var _ = Describe("convertMPIReplicaStatuses", func() {
	It("skips nil MPI v2beta1 replica status entries", func() {
		active := int32(2)

		result := convertMPIReplicaStatuses(map[kubeflowv2beta1.MPIReplicaType]*kubeflowv2beta1.ReplicaStatus{
			kubeflowv2beta1.MPIReplicaTypeLauncher: nil,
			kubeflowv2beta1.MPIReplicaTypeWorker:   {Active: active},
		})

		Expect(result).NotTo(HaveKey(kubeflowv1.ReplicaType(kubeflowv2beta1.MPIReplicaTypeLauncher)))
		Expect(result).To(HaveKeyWithValue(kubeflowv1.ReplicaType(kubeflowv2beta1.MPIReplicaTypeWorker),
			&kubeflowv1.ReplicaStatus{Active: active}))
	})
})

func newTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = v1alpha1.AddToScheme(s)
	_ = dromosv1alpha1.AddToScheme(s)
	_ = kubeflowv1.AddToScheme(s)
	return s
}

func emptyWLR(name, namespace string) *dromosv1alpha1.WorkloadRunner {
	wlr := &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}
	wlr.SetGroupVersionKind(dromosv1alpha1.SchemeGroupVersion.WithKind(dromosv1alpha1.WorkloadRunnerKind))
	return wlr
}

func makeJob(name, namespace string, uid types.UID) *kubeflowv1.PyTorchJob {
	return &kubeflowv1.PyTorchJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       uid,
		},
		Status: kubeflowv1.JobStatus{
			Conditions: []kubeflowv1.JobCondition{
				{Type: kubeflowv1.JobRunning, Status: corev1.ConditionTrue},
			},
		},
	}
}

func readyPod(name, namespace string, ownerUID types.UID) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    topOwnerPodLabels(ownerUID),
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
		},
	}
}

func initializingPod(name, namespace string, ownerUID types.UID) corev1.Pod {
	pod := readyPod(name, namespace, ownerUID)
	pod.Status.Conditions[0].Status = corev1.ConditionFalse
	return pod
}

func failedPod(name, namespace string, ownerUID types.UID) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    topOwnerPodLabels(ownerUID),
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "main",
				State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
					ExitCode: 1,
					Reason:   "Error",
					Message:  "boom",
				}},
			}},
		},
	}
}

func topOwnerPodLabels(ownerUID types.UID) map[string]string {
	return map[string]string{
		commonconstants.TopOwnerWorkloadIdLabel: string(ownerUID),
		"run.ai/top-owner-uid":                  string(ownerUID),
	}
}

func decodedAdditionalFields(ew *v1alpha1.ExternalWorkload) map[string]any {
	if ew.Spec.AdditionalFields == nil {
		return nil
	}
	fields := map[string]any{}
	Expect(json.Unmarshal([]byte(*ew.Spec.AdditionalFields), &fields)).To(Succeed())
	return fields
}
