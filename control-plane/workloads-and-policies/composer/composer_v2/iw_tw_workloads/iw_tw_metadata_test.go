/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package iw_tw_workloads

import (
	"context"
	"encoding/json"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workspace and Training pod metadata", func() {
	for _, workloadType := range []string{composer.Workspace, composer.Training} {
		Context(workloadType, func() {
			var spec workloadsApi.SupersetSpec

			BeforeEach(func() {
				spec = workloadsApi.SupersetSpec{}
				spec.SetImage("ubuntu")
			})

			compose := func() *BatchJobTarget {
				GinkgoHelper()
				c := NewInterTrainJobComposer(workloadType, "example", "test", "runai-test", "tester", nil, composer.FieldsComposerOptions{}, nil)
				target := NewMinimalBatchJobTarget()
				Expect(c.Compose(context.Background(), &spec, target)()).To(BeNil())
				return target
			}

			It("propagates custom labels and annotations to the Job and pod template", func() {
				Expect(json.Unmarshal([]byte(`{
					"labels": [{"name": "app.kubernetes.io/part-of", "value": "example"}],
					"annotations": [{"name": "example.com/note", "value": "test"}]
				}`), &spec)).To(Succeed())

				target := compose()
				for _, labels := range []map[string]string{target.Job.Labels, target.Job.Spec.Template.Labels} {
					Expect(labels).To(HaveKeyWithValue("app.kubernetes.io/part-of", "example"))
				}
				Expect(target.Job.Annotations).To(HaveKeyWithValue("example.com/note", "test"))
				Expect(target.Job.Spec.Template.Annotations).To(HaveKeyWithValue("example.com/note", "test"))
			})

			It("preserves empty label values and omits excluded labels", func() {
				Expect(json.Unmarshal([]byte(`[
					{"name": "example.com/empty", "value": ""},
					{"name": "example.com/no-value"},
					{"name": "example.com/excluded", "value": "unused", "exclude": true}
				]`), &spec.Labels)).To(Succeed())

				target := compose()
				for _, labels := range []map[string]string{target.Job.Labels, target.Job.Spec.Template.Labels} {
					Expect(labels).To(HaveKeyWithValue("example.com/empty", ""))
					Expect(labels).To(HaveKeyWithValue("example.com/no-value", ""))
					Expect(labels).NotTo(HaveKey("example.com/excluded"))
				}
			})

			DescribeTable("preserves generated pod identity metadata",
				func(withConflictingLabels bool) {
					expected := map[string]string{
						"app":          "job",
						"createdBy":    "Job",
						"project":      "test",
						"release":      "example",
						"workloadKind": workloadType,
						"workloadName": "example",
					}
					if withConflictingLabels {
						for key := range expected {
							label := workloadsApi.Label{}
							label.SetName(key)
							label.SetValue("user-value")
							spec.Labels = append(spec.Labels, label)
						}
					}
					annotation := workloadsApi.Annotation{}
					annotation.SetName("user")
					annotation.SetValue("user-value")
					spec.Annotations = []workloadsApi.Annotation{annotation}

					target := compose()
					Expect(target.Job.Spec.Template.Labels).To(Equal(expected))
					Expect(target.Job.Spec.Template.Annotations).To(HaveKeyWithValue("user", "tester"))
				},
				Entry("without user labels", false),
				Entry("with conflicting user labels", true),
			)
		})
	}
})
