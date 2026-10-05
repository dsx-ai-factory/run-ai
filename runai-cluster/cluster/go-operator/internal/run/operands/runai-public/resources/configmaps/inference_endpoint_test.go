/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configmaps

import (
	"context"

	kcfake "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client/fake"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Inference endpoint capability publication", func() {
	DescribeTable("advertises the transport only when all participating operands are enabled", func(disabled string) {
		config := getRunaiConfig()
		config.Spec.Agent.Enabled = ptr.To(disabled != "agent")
		config.Spec.AnyworkloadController.Enabled = ptr.To(disabled != "anyworkload-controller")
		config.Spec.InferenceEndpointController.Enabled = ptr.To(disabled != "inference-endpoint-controller")
		kbClient := createKubeClientMock(inferenceCapabilityDeployments(&config)...)
		kc := kcfake.New(kbClient, kbClient, createDynamicClient())
		obj, err := NewRunaiPublic("test", kc, config).Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		var public struct {
			Configuration map[string]interface{} `yaml:"configuration"`
		}
		Expect(yaml.Unmarshal([]byte(obj.(*corev1.ConfigMap).Data["runai-public"]), &public)).To(Succeed())
		if disabled == "" {
			Expect(public.Configuration).To(HaveKeyWithValue("inferenceEndpointSchemaVersion", "v1"))
		} else {
			Expect(public.Configuration).NotTo(HaveKey("inferenceEndpointSchemaVersion"))
		}
	}, Entry("all enabled", ""), Entry("agent disabled", "agent"), Entry("anyworkload disabled", "anyworkload-controller"), Entry("endpoint controller disabled", "inference-endpoint-controller"))

	DescribeTable("withholds capability until every desired component has rolled out", func(index int, change func(*appsv1.Deployment)) {
		config := getRunaiConfig()
		config.Spec.Agent.Enabled = ptr.To(true)
		config.Spec.AnyworkloadController.Enabled = ptr.To(true)
		config.Spec.InferenceEndpointController.Enabled = ptr.To(true)
		objects := inferenceCapabilityDeployments(&config)
		change(objects[index].(*appsv1.Deployment))
		kbClient := createKubeClientMock(objects...)
		kc := kcfake.New(kbClient, kbClient, createDynamicClient())
		obj, err := NewRunaiPublic("test", kc, config).Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		var public struct {
			Configuration map[string]interface{} `yaml:"configuration"`
		}
		Expect(yaml.Unmarshal([]byte(obj.(*corev1.ConfigMap).Data["runai-public"]), &public)).To(Succeed())
		Expect(public.Configuration).NotTo(HaveKey("inferenceEndpointSchemaVersion"))
	},
		Entry("Agent still uses the previous image", 0, func(d *appsv1.Deployment) { d.Spec.Template.Spec.Containers[0].Image = "old-agent:v0" }),
		Entry("Agent is available only through old pods", 0, func(d *appsv1.Deployment) { d.Status.UpdatedReplicas = 0 }),
		Entry("Anyworkload generation has not been observed", 1, func(d *appsv1.Deployment) { d.Status.ObservedGeneration-- }),
		Entry("Anyworkload still has an old replica", 1, func(d *appsv1.Deployment) { d.Status.Replicas++ }),
		Entry("endpoint controller is unavailable", 2, func(d *appsv1.Deployment) { d.Status.AvailableReplicas = 0 }),
		Entry("Agent scaled to zero", 0, func(d *appsv1.Deployment) {
			d.Spec.Replicas = ptr.To(int32(0))
			d.Status = appsv1.DeploymentStatus{ObservedGeneration: d.Generation}
		}),
	)
})

func inferenceCapabilityDeployments(config *runv1.RunaiConfig) []client.Object {
	objects := []client.Object{}
	for i, name := range []string{operands.RunaiAgentName, operands.RunaiAnyworkloadControllerName, operands.RunaiInferenceEndpointControllerName} {
		images := []*string{formatImageRef(config.Spec.Agent.Image), formatImageRef(config.Spec.AnyworkloadController.Image), formatImageRef(config.Spec.InferenceEndpointController.Image)}
		objects = append(objects, &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: config.Namespace, Generation: 2},
			Spec:       appsv1.DeploymentSpec{Replicas: ptr.To(int32(1)), Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: name, Image: *images[i]}}}}},
			Status: appsv1.DeploymentStatus{ObservedGeneration: 2, Replicas: 1, UpdatedReplicas: 1, AvailableReplicas: 1, ReadyReplicas: 1,
				Conditions: []appsv1.DeploymentCondition{{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue}}},
		})
	}
	return append(objects, getTrainingOperatorCm())
}
