/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	karta "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/karta/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Dynamo v1beta1 Karta", func() {
	It("reads the complete pod template, including sidecars and init containers", func() {
		component := dynamoV1beta1Component(`[
          {"name":"worker", "replicas":2, "multinode":{"nodeCount":3}, "podTemplate":{
            "metadata":{"labels":{"app":"chatbot"},"annotations":{"example.com/test":"value"}},
            "spec":{
              "schedulerName":"kai-scheduler", "priorityClassName":"inference",
              "nodeSelector":{"pool":"gpu"}, "serviceAccountName":"model-reader",
              "resourceClaims":[{"name":"gpu","resourceClaimTemplateName":"gpu-template"}],
              "containers":[
                {"name":"metrics","image":"metrics:test","resources":{"requests":{"cpu":"100m"}}},
                {"name":"main","image":"dynamo:test","resources":{"limits":{"nvidia.com/gpu":"1"}}}
              ],
              "initContainers":[
                {"name":"download","image":"download:test","resources":{"requests":{"memory":"4Gi"}}},
                {"name":"runtime","image":"dynamo:test","restartPolicy":"Always",
                  "resources":{"requests":{"cpu":"250m"}}}
              ]
            }
          }}
        ]`)
		instances, err := component.GetExtractedInstances(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(instances).To(HaveKey("worker"))
		template := instances["worker"].PodTemplateSpec
		Expect(template).NotTo(BeNil())
		Expect(template.Labels).To(HaveKeyWithValue("app", "chatbot"))
		Expect(template.Annotations).To(HaveKeyWithValue("example.com/test", "value"))
		Expect(template.Spec.SchedulerName).To(Equal("kai-scheduler"))
		Expect(template.Spec.PriorityClassName).To(Equal("inference"))
		Expect(template.Spec.NodeSelector).To(HaveKeyWithValue("pool", "gpu"))
		Expect(template.Spec.ServiceAccountName).To(Equal("model-reader"))
		Expect(template.Spec.ResourceClaims).To(HaveLen(1))
		Expect(template.Spec.Containers).To(HaveLen(2))
		Expect(template.Spec.Containers[0].Resources.Requests.Cpu().String()).To(Equal("100m"))
		Expect(template.Spec.Containers[1].Name).To(Equal("main"))
		Expect(template.Spec.Containers[1].Image).To(Equal("dynamo:test"))
		gpu := template.Spec.Containers[1].Resources.Limits[corev1.ResourceName("nvidia.com/gpu")]
		Expect(gpu.String()).To(Equal("1"))
		Expect(template.Spec.InitContainers).To(HaveLen(2))
		Expect(template.Spec.InitContainers[0].Resources.Requests.Memory().String()).To(Equal("4Gi"))
		Expect(template.Spec.InitContainers[1].RestartPolicy).NotTo(BeNil())
		Expect(*template.Spec.InitContainers[1].RestartPolicy).To(Equal(corev1.ContainerRestartPolicyAlways))
		Expect(*instances["worker"].Scale.Replicas).To(Equal(int32(6)))
	})

	It("keeps component identities aligned when a legacy template or main container is absent", func() {
		component := dynamoV1beta1Component(`[
          {"name":"legacy"},
          {"name":"null-template","podTemplate":null},
          {"name":"metadata-only","podTemplate":{"metadata":{"labels":{"app":"frontend"}}}},
          {"name":"sidecar-only","podTemplate":{"spec":{"containers":[{"name":"sidecar","image":"metrics:test"}]}}},
          {"name":"worker","podTemplate":{"spec":{"containers":[{"name":"main","image":"worker:test"}]}}}
        ]`)
		instances, err := component.GetExtractedInstances(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(instances).To(HaveLen(5))
		Expect(instances["legacy"].PodTemplateSpec).NotTo(BeNil())
		Expect(instances["legacy"].PodTemplateSpec.Spec.Containers).To(BeEmpty())
		Expect(instances["metadata-only"].PodTemplateSpec.Labels).To(HaveKeyWithValue("app", "frontend"))
		Expect(instances["sidecar-only"].PodTemplateSpec.Spec.Containers[0].Name).To(Equal("sidecar"))
		Expect(instances["worker"].PodTemplateSpec.Spec.Containers[0].Image).To(Equal("worker:test"))
	})

	It("writes pod changes back without dropping sidecars or init containers", func() {
		component := dynamoV1beta1Component(`[
          {"name":"worker","podTemplate":{"spec":{
            "containers":[{"name":"main","image":"worker:test"},{"name":"sidecar","image":"metrics:test"}],
            "initContainers":[{"name":"download","image":"download:test"}]
          }}}
        ]`)
		templates, err := component.GetPodTemplateSpec(context.Background())
		Expect(err).NotTo(HaveOccurred())
		template := templates["worker"]
		template.Spec.NodeSelector = map[string]string{"pool": "gpu"}
		templates["worker"] = template
		Expect(component.UpdatePodTemplateSpec(context.Background(), templates)).To(Succeed())
		updated, err := component.GetPodTemplateSpec(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(updated["worker"]).To(Equal(template))
	})

	DescribeTable("reads replicas without inventing autoscaling bounds",
		func(componentJSON string, replicas int32) {
			component := dynamoV1beta1Component("[" + componentJSON + "]")
			scales, err := component.GetScale(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(scales).To(HaveKey("worker"))
			Expect(scales["worker"].Replicas).NotTo(BeNil())
			Expect(*scales["worker"].Replicas).To(Equal(replicas))
			Expect(scales["worker"].MinReplicas).To(BeNil())
			Expect(scales["worker"].MaxReplicas).To(BeNil())
		},
		Entry("omitted replicas", `{"name":"worker"}`, int32(1)),
		Entry("zero replicas", `{"name":"worker","replicas":0}`, int32(0)),
		Entry("multiple replicas", `{"name":"worker","replicas":3}`, int32(3)),
		Entry("default node count when multinode is enabled", `{"name":"worker","replicas":2,"multinode":{}}`, int32(4)),
		Entry("multiple nodes", `{"name":"worker","replicas":2,"multinode":{"nodeCount":4}}`, int32(8)),
		Entry("zero replicas across multiple nodes", `{"name":"worker","replicas":0,"multinode":{"nodeCount":4}}`, int32(0)),
	)
})

func dynamoV1beta1Component(componentsJSON string) *resource.Component {
	component, err := dynamoV1beta1Factory(componentsJSON).GetComponent("component")
	Expect(err).NotTo(HaveOccurred())
	return component
}

func dynamoV1beta1Factory(componentsJSON string) *resource.ComponentFactory {
	var components []any
	Expect(json.Unmarshal([]byte(componentsJSON), &components)).To(Succeed())
	object := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "nvidia.com/v1beta1", "kind": "DynamoGraphDeployment",
		"metadata": map[string]any{"name": "chatbot"},
		"spec":     map[string]any{"components": components},
	}}
	definition := BuildDynamoGraphDeploymentV1beta1()
	k := &karta.Karta{Spec: definition.Spec}
	Expect(karta.NewKartaValidator(k).Validate()).To(Succeed())
	return resource.NewComponentFactoryFromObject(k, object)
}
