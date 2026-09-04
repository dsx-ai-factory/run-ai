/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pcreds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testRunaiNS = "runai"

func newValidatorPlugin(seeded ...string) *Plugin {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	objs := make([]client.Object, 0, len(seeded))
	for _, name := range seeded {
		objs = append(objs, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testRunaiNS},
		})
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	return NewPlugin(c, testRunaiNS)
}

func runnerWithAnnotation(value string) *dromosv1alpha1.WorkloadRunner {
	r := &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{Name: "test-runner", Namespace: "test-ns"},
	}
	if value != "" {
		r.Annotations = map[string]string{pcreds.Annotation: value}
	}
	return r
}

var _ = Describe("WebhookValidate", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	DescribeTable("rejects bad input",
		func(seeded []string, annotation string, expectedType field.ErrorType, expectedSubstring string) {
			plugin := newValidatorPlugin(seeded...)
			_, errs := plugin.WebhookValidate(ctx, nil, runnerWithAnnotation(annotation))

			Expect(errs).NotTo(BeEmpty())
			Expect(errs[0].Type).To(Equal(expectedType))
			Expect(errs.ToAggregate().Error()).To(ContainSubstring(expectedSubstring))
		},

		Entry("malformed yaml",
			nil, "not: a: list", field.ErrorTypeInvalid, "invalid yaml"),

		Entry("empty list",
			nil, "[]", field.ErrorTypeInvalid, "at least one entry"),

		Entry("entry missing sourceSecret",
			nil, "- otherField: foo\n", field.ErrorTypeInvalid, "sourceSecret is required"),

		Entry("source Secret missing in runai ns",
			[]string{"present"},
			"- sourceSecret: present\n- sourceSecret: missing\n",
			field.ErrorTypeNotFound, "missing"),
	)

	DescribeTable("admits valid input",
		func(seeded []string, annotation string) {
			plugin := newValidatorPlugin(seeded...)
			_, errs := plugin.WebhookValidate(ctx, nil, runnerWithAnnotation(annotation))
			Expect(errs).To(BeEmpty())
		},

		Entry("annotation absent",
			nil, ""),

		Entry("single source resolves",
			[]string{"my-hf-token-abc1234567"},
			"- sourceSecret: my-hf-token-abc1234567\n"),

		Entry("multiple sources all resolve",
			[]string{"a", "b", "c"},
			"- sourceSecret: a\n- sourceSecret: b\n- sourceSecret: c\n"),
	)
})
