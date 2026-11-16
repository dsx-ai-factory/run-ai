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
	"github.com/run-ai/karta/pkg/instructions"
	"github.com/run-ai/karta/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Dynamo v1beta1 Karta", func() {
	It("reads role templates while preserving component identities", func() {
		manifest := `[
   {"name":"worker","replicas":2,"multinode":{"nodeCount":3},"roles":[
    {"name":"leader","podTemplate":{"spec":{"containers":[{"name":"main","image":"leader:test"}]}}},
    {"name":"worker","podTemplate":{"spec":{"containers":[{"name":"main","image":"worker:test"},{"name":"metrics","image":"metrics:test"}]}}}
   ]},
   {"name":"frontend","podTemplate":{"spec":{"containers":[{"name":"main","image":"frontend:test"}]}}}
  ]`
		ordinary, err := dynamoV1beta1NamedComponent(manifest, "component").GetInstanceIds(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(ordinary).To(Equal([]string{"worker", "frontend"}))
		roles := dynamoV1beta1NamedComponent(manifest, "role")
		instances, err := roles.GetExtractedInstances(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(instances).To(HaveLen(2))
		Expect(*instances["worker/leader"].Scale.Replicas).To(Equal(int32(2)))
		Expect(*instances["worker/worker"].Scale.Replicas).To(Equal(int32(4)))
		Expect(instances["worker/worker"].PodTemplateSpec.Spec.Containers).To(HaveLen(2))
	})

	It("preserves pod matching and gang scheduling without requiring role labels", func() {
		factory := dynamoV1beta1Factory(`[
          {"name":"frontend"},
          {"name":"worker","replicas":2,"multinode":{"nodeCount":3},"roles":[
            {"name":"leader","podTemplate":{"spec":{"containers":[]}}}
          ]}
        ]`)
		definition := factory.GetKarta()
		members := definition.Spec.Instructions.GangScheduling.PodGroups[0].Members
		Expect(members).To(HaveLen(1))
		Expect(members[0].ComponentName).To(Equal("component"))
		summary, err := instructions.NewStructureSummary(definition)
		Expect(err).NotTo(HaveOccurred())
		roles, err := factory.GetComponent("role")
		Expect(err).NotTo(HaveOccurred())
		for _, tc := range []struct {
			name, roleLabel string
			replicas        int32
		}{
			{name: "frontend", replicas: 1},
			{name: "worker", replicas: 6},
			{name: "worker", roleLabel: "leader", replicas: 6},
		} {
			pod := &corev1.Pod{}
			pod.Labels = map[string]string{"nvidia.com/dynamo-component": tc.name}
			if tc.roleLabel != "" {
				pod.Labels["run.ai/dynamo-role"] = tc.roleLabel
			}
			querier := resource.NewPodQuerier(pod)
			matchesRole, err := querier.MatchesComponentType(context.Background(), roles.Definition().PodSelector.ComponentTypeSelector)
			Expect(err).NotTo(HaveOccurred())
			Expect(matchesRole).To(BeFalse())
			actual, err := instructions.InferPodComponent(context.Background(), querier, summary)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal("component"))
			instance, err := instructions.InferPodComponentInstance(context.Background(), querier, actual, factory)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance).NotTo(BeNil())
			Expect(*instance).To(Equal(tc.name))
			group, err := instructions.GetPodGroupingEffectiveComponent(context.Background(), querier, actual, summary)
			Expect(err).NotTo(HaveOccurred())
			Expect(group).NotTo(BeNil())
			Expect(group.MemberDefinition.ComponentName).To(Equal("component"))
			count, err := instructions.CalculateSubtreeScale(context.Background(), actual, instance, factory, summary)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(tc.replicas))
		}
	})

	It("updates a role template without creating a component template", func() {
		ctx := context.Background()
		factory := dynamoV1beta1Factory(`[
          {"name":"worker","roles":[
            {"name":"leader","podTemplate":{"spec":{"containers":[{"name":"main","image":"leader:test"}]}}},
            {"name":"worker","podTemplate":{"spec":{"containers":[{"name":"main","image":"worker:test"},{"name":"metrics","image":"metrics:test"}]}}}
          ]},
          {"name":"frontend","podTemplate":{"spec":{"containers":[{"name":"main","image":"frontend:test"}]}}}
        ]`)
		roles, err := factory.GetComponent("role")
		Expect(err).NotTo(HaveOccurred())
		templates, err := roles.GetPodTemplateSpec(ctx)
		Expect(err).NotTo(HaveOccurred())
		worker := templates["worker/worker"]
		worker.Spec.SchedulerName = "kai-scheduler"
		templates["worker/worker"] = worker
		Expect(roles.UpdatePodTemplateSpec(ctx, templates)).To(Succeed())

		updated, err := factory.GetResource()
		Expect(err).NotTo(HaveOccurred())
		components, _, err := unstructured.NestedSlice(updated.(*unstructured.Unstructured).Object, "spec", "components")
		Expect(err).NotTo(HaveOccurred())
		Expect(components[0].(map[string]any)).NotTo(HaveKey("podTemplate"))
		frontend := components[1].(map[string]any)
		Expect(frontend["podTemplate"]).To(Equal(map[string]any{"spec": map[string]any{"containers": []any{map[string]any{"name": "main", "image": "frontend:test"}}}}))
		instances, err := roles.GetExtractedInstances(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instances["worker/worker"].PodTemplateSpec).To(Equal(&worker))
		Expect(instances["worker/leader"].PodTemplateSpec.Spec.SchedulerName).To(BeEmpty())
	})

	DescribeTable("reads role replica defaults",
		func(componentJSON string, leader, worker int32) {
			roles := dynamoV1beta1NamedComponent("["+componentJSON+"]", "role")
			instances, err := roles.GetExtractedInstances(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(*instances["worker/leader"].Scale.Replicas).To(Equal(leader))
			Expect(*instances["worker/worker"].Scale.Replicas).To(Equal(worker))
		},
		Entry("omitted counts", `{"name":"worker","multinode":{},"roles":[{"name":"leader","podTemplate":{"spec":{"containers":[]}}},{"name":"worker","podTemplate":{"spec":{"containers":[]}}}]}`, int32(1), int32(1)),
		Entry("explicit role counts", `{"name":"worker","replicas":2,"multinode":{"nodeCount":3},"roles":[{"name":"leader","replicas":1,"podTemplate":{"spec":{"containers":[]}}},{"name":"worker","replicas":2,"podTemplate":{"spec":{"containers":[]}}}]}`, int32(2), int32(4)),
		Entry("zero component replicas", `{"name":"worker","replicas":0,"multinode":{"nodeCount":3},"roles":[
            {"name":"leader","replicas":1,"podTemplate":{"spec":{"containers":[]}}},
            {"name":"worker","replicas":2,"podTemplate":{"spec":{"containers":[]}}}
        ]}`, int32(0), int32(0)),
	)

})

func dynamoV1beta1NamedComponent(componentsJSON, name string) *resource.Component {
	component, err := dynamoV1beta1Factory(componentsJSON).GetComponent(name)
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
