/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package template

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

const (
	invalidJSON = `{invalid json content`
)

func TestTemplateParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Template Parser Suite")
}

var _ = Describe("Parser", func() {
	templateParser := NewParser()

	Context("with valid template", func() {
		It("parses valid template correctly", func() {
			js := createValidJobSet()
			template := runtime.RawExtension{Raw: marshalJobSet(js)}

			result, err := templateParser.Parse(template)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.GroupVersionKind()).To(Equal(js.GroupVersionKind()))
			Expect(result.GetName()).To(Equal(js.GetName()))
			Expect(result.Object).To(HaveKey("spec"))
		})
	})

	Context("with invalid template", func() {
		It("returns error for empty template", func() {
			template := runtime.RawExtension{Raw: []byte{}}

			result, err := templateParser.Parse(template)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("template is empty"))
			Expect(result).To(BeNil())
		})

		It("returns error for missing apiVersion", func() {
			missingAPIVersion := runtime.RawExtension{Raw: createMissingAPIVersionJSON()}

			result, err := templateParser.Parse(missingAPIVersion)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing required apiVersion or kind"))
			Expect(result).To(BeNil())
		})

		It("returns error for missing kind", func() {
			missingKind := runtime.RawExtension{Raw: createMissingKindJSON()}

			result, err := templateParser.Parse(missingKind)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing required apiVersion or kind"))
			Expect(result).To(BeNil())
		})

		It("returns error for invalid JSON", func() {
			template := runtime.RawExtension{Raw: []byte(invalidJSON)}

			result, err := templateParser.Parse(template)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})
})

func createValidJobSet() *jobsetv1alpha2.JobSet {
	return &jobsetv1alpha2.JobSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "jobset.x-k8s.io/v1alpha2",
			Kind:       "JobSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-jobset",
		},
		Spec: jobsetv1alpha2.JobSetSpec{
			ReplicatedJobs: []jobsetv1alpha2.ReplicatedJob{
				{
					Name:     "workers",
					Replicas: 5,
				},
			},
		},
	}
}

func marshalJobSet(js *jobsetv1alpha2.JobSet) []byte {
	data, err := json.Marshal(js)
	if err != nil {
		panic(err)
	}

	return data
}

func createMissingAPIVersionJSON() []byte {
	js := createValidJobSet()
	data := marshalJobSet(js)

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		panic(err)
	}

	delete(obj, "apiVersion")

	result, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return result
}

func createMissingKindJSON() []byte {
	js := createValidJobSet()
	data := marshalJobSet(js)

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		panic(err)
	}

	delete(obj, "kind")

	result, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return result
}
