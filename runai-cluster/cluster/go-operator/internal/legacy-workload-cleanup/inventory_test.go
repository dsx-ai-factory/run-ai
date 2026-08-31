/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

var _ = Describe("Legacy workload inventory", func() {
	It("discovers only workload APIs whose required version is served", func() {
		client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(),
			servedCRD("trainingworkloads.run.ai", "v2alpha1", true),
			servedCRD("interactiveworkloads.run.ai", "v2alpha1", false),
		)

		served, err := discoverServedWorkloads(context.Background(), client)
		Expect(err).NotTo(HaveOccurred())
		Expect(served).To(ConsistOf(legacyWorkloadResources[0]))
	})

	It("lists legacy workloads in every namespace", func() {
		training := newObject("run.ai/v2alpha1", "TrainingWorkload", "project-a", "training", "training-uid")
		client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
			map[schema.GroupVersionResource]string{legacyWorkloadResources[0].GVR: "TrainingWorkloadList"},
			training,
		)

		objects, err := listObjects(context.Background(), client, legacyWorkloadResources[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(objects).To(HaveLen(1))
		Expect(objects[0].GetUID()).To(Equal(training.GetUID()))
	})
})

func servedCRD(name, version string, served bool) *unstructured.Unstructured {
	crd := &apiextensionsv1.CustomResourceDefinition{
		TypeMeta:   metav1.TypeMeta{APIVersion: apiextensionsv1.SchemeGroupVersion.String(), Kind: "CustomResourceDefinition"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group:    "run.ai",
			Names:    apiextensionsv1.CustomResourceDefinitionNames{Plural: "unused", Kind: "Unused"},
			Scope:    apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{Name: version, Served: served, Storage: served}},
		},
	}
	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(crd)
	Expect(err).NotTo(HaveOccurred())
	return &unstructured.Unstructured{Object: object}
}
