/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpu_pci_inventory

import (
	"context"
	"testing"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	internalv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestInventory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GPU PCI Inventory")
}

var _ = Describe("GPU PCI inventory operand", func() {
	var config *runv1.RunaiConfig
	var reader kc.HybridClient
	BeforeEach(func() {
		config = &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "custom-runai"}, Spec: runv1.RunaiConfigSpec{Internal: &internalv1.Internal{Global: &internalv1.Global{Version: "test"}}}}
		config.SetDefaultsWhereNeeded()
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(appsv1.AddToScheme(scheme)).To(Succeed())
		c := fake.NewClientBuilder().WithScheme(scheme).Build()
		reader = kc.New(c, c, nil)
	})
	It("deploys a restricted GPU Linux collector using the configured image without backend credentials", func() {
		image := config.Spec.ClusterSync.Image
		image.Registry = ptr.To("registry.test")
		image.Name = ptr.To("custom-sync")
		image.Tag = ptr.To("test-tag")
		image.PullPolicy = corev1.PullAlways
		config.Spec.Internal.Global.AdditionalImagePullSecrets = []string{"pull-secret"}
		config.Spec.Internal.Global.NMC.ServiceAccountName = "management-sa"
		config.Spec.Global.Tolerations = []corev1.Toleration{{Key: "global", Operator: corev1.TolerationOpExists}}
		config.Spec.Global.DaemonsetsTolerations = []corev1.Toleration{{Key: "daemon", Operator: corev1.TolerationOpExists}}
		op := Operand(reader, config)
		Expect(op.Enabled()).To(BeTrue())
		Expect(op.Resources()).To(HaveLen(2))
		obj, err := op.Resources()[0].Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		ds := obj.(*appsv1.DaemonSet)
		pod := ds.Spec.Template.Spec
		c := pod.Containers[0]
		Expect(ds.Name).To(Equal("gpu-pci-inventory"))
		Expect(ds.Namespace).To(Equal("custom-runai"))
		Expect(pod.ServiceAccountName).To(Equal("gpu-pci-inventory"))
		Expect(pod.NodeSelector).To(HaveKeyWithValue("kubernetes.io/os", "linux"))
		terms := pod.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		for _, term := range terms {
			Expect(term.MatchExpressions).To(ContainElement(corev1.NodeSelectorRequirement{Key: config.Spec.Global.NodeAffinity.NFDGPULabel, Operator: corev1.NodeSelectorOpIn, Values: []string{"true"}}))
		}
		Expect(pod.Tolerations).To(ContainElements(
			config.Spec.Global.Tolerations[0], config.Spec.Global.DaemonsetsTolerations[0],
			corev1.Toleration{Key: "nvidia.com/gpu", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
		))
		Expect(pod.ImagePullSecrets).To(Equal([]corev1.LocalObjectReference{{Name: "pull-secret"}}))
		Expect(c.Image).To(Equal("registry.test/custom-sync:test-tag"))
		Expect(c.ImagePullPolicy).To(Equal(corev1.PullAlways))
		Expect(c.Env).To(ContainElements(
			corev1.EnvVar{Name: "GPU_PCI_COLLECTOR", Value: "true"},
			corev1.EnvVar{Name: "NODE_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		))
		for _, env := range c.Env {
			Expect(env.ValueFrom == nil || env.ValueFrom.SecretKeyRef == nil).To(BeTrue())
			Expect(env.Name).NotTo(BeElementOf("ENDPOINT", "TOKEN", "REMOTE_CLUSTER_API_SERVER"))
		}
		Expect(c.Resources.Requests).To(HaveLen(2))
		Expect(c.Resources.Limits).To(HaveLen(2))
		Expect(c.Resources.Requests).To(HaveKey(corev1.ResourceCPU))
		Expect(c.Resources.Requests).To(HaveKey(corev1.ResourceMemory))
		Expect(c.Resources.Limits).To(HaveKey(corev1.ResourceCPU))
		Expect(c.Resources.Limits).To(HaveKey(corev1.ResourceMemory))
		Expect(pod.HostPID || pod.HostNetwork || pod.HostIPC).To(BeFalse())
		Expect(c.SecurityContext.RunAsNonRoot).To(HaveValue(BeTrue()))
		Expect(c.SecurityContext.RunAsUser).To(HaveValue(Equal(int64(65532))))
		Expect(c.SecurityContext.Privileged).To(HaveValue(BeFalse()))
		Expect(c.SecurityContext.AllowPrivilegeEscalation).To(HaveValue(BeFalse()))
		Expect(c.SecurityContext.ReadOnlyRootFilesystem).To(HaveValue(BeTrue()))
		Expect(c.SecurityContext.Capabilities.Drop).To(Equal([]corev1.Capability{"ALL"}))
		Expect(c.SecurityContext.SeccompProfile.Type).To(Equal(corev1.SeccompProfileTypeRuntimeDefault))
		Expect(pod.Volumes).To(HaveLen(2))
		Expect(c.VolumeMounts).To(HaveLen(2))
		for i, path := range []string{"/sys", "/proc/driver/nvidia/gpus"} {
			Expect(pod.Volumes[i].HostPath.Path).To(Equal(path))
			Expect(pod.Volumes[i].HostPath.Type).To(HaveValue(Equal(corev1.HostPathDirectory)))
			Expect(c.VolumeMounts[i].ReadOnly).To(BeTrue())
		}
		Expect(c.VolumeMounts[0].MountPath).To(Equal("/host/sys"))
		Expect(c.VolumeMounts[1].MountPath).To(Equal("/host/nvidia-gpus"))
		sa, err := op.Resources()[1].Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(sa.(*corev1.ServiceAccount).Name).To(Equal(pod.ServiceAccountName))
	})
	It("follows cluster-sync enablement and keeps optional discovery out of installation readiness", func() {
		op := Operand(reader, config)
		Expect(op.Enabled()).To(BeTrue())
		status, err := op.Status(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(status.Ready).To(BeTrue())
		config.Spec.ClusterSync.Enabled = ptr.To(false)
		Expect(op.Enabled()).To(BeFalse())
		Expect(op.Resources()).To(HaveLen(2))
	})
})
