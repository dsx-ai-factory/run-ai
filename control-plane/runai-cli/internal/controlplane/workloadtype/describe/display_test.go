/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"encoding/json"
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workloadtype/describe"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"gopkg.in/yaml.v3"
)

func captureStdout(fn func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	return string(out)
}

var _ = Describe("Display", func() {
	var (
		wt      *openapi.WorkloadTypeConfig
		created time.Time
		updated time.Time
	)

	BeforeEach(func() {
		created = time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
		updated = time.Date(2025, 1, 16, 12, 0, 0, 0, time.UTC)
		wt = &openapi.WorkloadTypeConfig{
			Name:           "Training",
			Id:             strPtr("wt-123"),
			CategoryName:   strPtr("batch"),
			PriorityName:   strPtr("high"),
			Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
			Group:          "run.ai",
			CreatedAt:      &created,
			CreatedBy:      strPtr("admin"),
			UpdatedAt:      &updated,
			UpdatedBy:      strPtr("admin"),
			Kartas: []openapi.WorkloadTypeKartasWithVersionInner{
				{
					Version:   strPtr("v1"),
					UpdatedAt: &updated,
					Karta: map[string]interface{}{
						"spec": map[string]interface{}{
							"replicas": "integer",
						},
					},
				},
			},
		}
	})

	Describe("AsJSON", func() {
		It("should output valid JSON with expected fields", func() {
			output := captureStdout(func() {
				displayer := describe.NewWorkloadTypeDisplayer(wt)
				err := displayer.AsJSON()
				Expect(err).NotTo(HaveOccurred())
			})

			var result describe.BriefWorkloadType
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Name).To(Equal("Training"))
			Expect(result.ID).To(Equal("wt-123"))
			Expect(result.Category).To(Equal("batch"))
			Expect(result.Priority).To(Equal("high"))
			Expect(result.Preemptibility).To(Equal("preemptible"))
			Expect(result.Group).To(Equal("run.ai"))
			Expect(result.Kartas).To(HaveLen(1))
			Expect(result.Kartas[0].GetVersion()).To(Equal("v1"))
			Expect(result.Kartas[0].Karta).To(HaveKey("spec"))
			Expect(result.ResourceInterfaces).To(BeEmpty())
		})

		It("should fall back to ResourceInterfaces when Kartas is empty", func() {
			wt.Kartas = nil
			wt.ResourceInterfaces = []openapi.WorkloadTypeResourceInterfacesWithVersionInner{
				{
					Version:   strPtr("v1"),
					UpdatedAt: &updated,
					ResourceInterface: map[string]interface{}{
						"spec": map[string]interface{}{
							"replicas": "integer",
						},
					},
				},
			}
			output := captureStdout(func() {
				displayer := describe.NewWorkloadTypeDisplayer(wt)
				err := displayer.AsJSON()
				Expect(err).NotTo(HaveOccurred())
			})

			var result describe.BriefWorkloadType
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Kartas).To(BeEmpty())
			Expect(result.ResourceInterfaces).To(HaveLen(1))
			Expect(result.ResourceInterfaces[0].GetVersion()).To(Equal("v1"))
		})
	})

	Describe("AsYAML", func() {
		It("should output valid YAML with expected fields", func() {
			output := captureStdout(func() {
				displayer := describe.NewWorkloadTypeDisplayer(wt)
				err := displayer.AsYAML()
				Expect(err).NotTo(HaveOccurred())
			})

			var result describe.BriefWorkloadType
			err := yaml.Unmarshal([]byte(output), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Name).To(Equal("Training"))
			Expect(result.ID).To(Equal("wt-123"))
			Expect(result.Group).To(Equal("run.ai"))
		})
	})

	Describe("AsTable", func() {
		It("should output metadata fields and kartas", func() {
			output := captureStdout(func() {
				displayer := describe.NewWorkloadTypeDisplayer(wt)
				err := displayer.AsTable()
				Expect(err).NotTo(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("Training"))
			Expect(output).To(ContainSubstring("wt-123"))
			Expect(output).To(ContainSubstring("batch"))
			Expect(output).To(ContainSubstring("high"))
			Expect(output).To(ContainSubstring("preemptible"))
			Expect(output).To(ContainSubstring("run.ai"))
			Expect(output).To(ContainSubstring("Kartas"))
			Expect(output).NotTo(ContainSubstring("Resource Interfaces"))
			Expect(output).To(ContainSubstring("v1"))
			Expect(output).To(ContainSubstring("replicas"))
		})

		It("should fall back to ResourceInterfaces under the Kartas header when Kartas is empty", func() {
			wt.Kartas = nil
			wt.ResourceInterfaces = []openapi.WorkloadTypeResourceInterfacesWithVersionInner{
				{
					Version:   strPtr("v1"),
					UpdatedAt: &updated,
					ResourceInterface: map[string]interface{}{
						"spec": map[string]interface{}{
							"replicas": "integer",
						},
					},
				},
			}
			output := captureStdout(func() {
				displayer := describe.NewWorkloadTypeDisplayer(wt)
				err := displayer.AsTable()
				Expect(err).NotTo(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("Kartas"))
			Expect(output).NotTo(ContainSubstring("Resource Interfaces"))
			Expect(output).To(ContainSubstring("v1"))
			Expect(output).To(ContainSubstring("replicas"))
		})

		It("should handle workload type without kartas or resource interfaces", func() {
			wt.Kartas = nil
			wt.ResourceInterfaces = nil
			output := captureStdout(func() {
				displayer := describe.NewWorkloadTypeDisplayer(wt)
				err := displayer.AsTable()
				Expect(err).NotTo(HaveOccurred())
			})

			Expect(output).To(ContainSubstring("Training"))
			Expect(output).NotTo(ContainSubstring("Kartas"))
			Expect(output).NotTo(ContainSubstring("Resource Interfaces"))
		})
	})

	Describe("RenderOutput", func() {
		It("should render table format", func() {
			output := captureStdout(func() {
				err := describe.RenderOutput(wt, "table")
				Expect(err).NotTo(HaveOccurred())
			})
			Expect(output).To(ContainSubstring("Training"))
		})

		It("should render json format", func() {
			output := captureStdout(func() {
				err := describe.RenderOutput(wt, "json")
				Expect(err).NotTo(HaveOccurred())
			})

			var result describe.BriefWorkloadType
			err := json.Unmarshal([]byte(output), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Name).To(Equal("Training"))
		})

		It("should render yaml format", func() {
			output := captureStdout(func() {
				err := describe.RenderOutput(wt, "yaml")
				Expect(err).NotTo(HaveOccurred())
			})

			var result describe.BriefWorkloadType
			err := yaml.Unmarshal([]byte(output), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Name).To(Equal("Training"))
		})

		It("should return error for unsupported format", func() {
			err := describe.RenderOutput(wt, "xml")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unsupported output format"))
		})
	})
})
