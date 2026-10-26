/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/application/describe"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
	"k8s.io/utils/ptr"
)

// minimalApp returns an AIApplication with only the required Name and ProjectName fields set.
func minimalApp() *openapi.AIApplication {
	return &openapi.AIApplication{
		Id:          ptr.To("id-1"),
		Name:        ptr.To("my-app"),
		Type:        ptr.To("inference"),
		ProjectName: ptr.To("p301"),
		CreatedAt:   ptr.To(time.Unix(0, 0).UTC()),
	}
}

var _ = Describe("RenderOutput", func() {
	var buf bytes.Buffer

	BeforeEach(func() {
		buf.Reset()
	})

	Context("table format", func() {
		It("renders the GENERAL section with name and project", func() {
			err := describe.RenderOutput(&buf, minimalApp(), "table")
			Expect(err).To(BeNil())
			out := buf.String()
			Expect(out).To(ContainSubstring("GENERAL"))
			Expect(out).To(ContainSubstring("my-app"))
			Expect(out).To(ContainSubstring("p301"))
		})

		It("omits CHART REFERENCE when Spec is nil", func() {
			err := describe.RenderOutput(&buf, minimalApp(), "table")
			Expect(err).To(BeNil())
			Expect(buf.String()).NotTo(ContainSubstring("CHART REFERENCE"))
		})

		It("renders CHART REFERENCE with chart URL when set", func() {
			app := minimalApp()
			app.Spec = &openapi.AIApplicationSpec{
				ChartRef: openapi.AIApplicationSpecChartRef{
					ChartURL: ptr.To("https://charts.example.com/grafana-11.0.0.tgz"),
				},
			}
			err := describe.RenderOutput(&buf, app, "table")
			Expect(err).To(BeNil())
			out := buf.String()
			Expect(out).To(ContainSubstring("CHART REFERENCE"))
			Expect(out).To(ContainSubstring("Chart URL:"))
			Expect(out).To(ContainSubstring("charts.example.com"))
			Expect(out).NotTo(ContainSubstring("Chart Name:"))
		})

		It("renders CHART REFERENCE with name/version/repo when chart URL is absent", func() {
			app := minimalApp()
			app.Spec = &openapi.AIApplicationSpec{
				ChartRef: openapi.AIApplicationSpecChartRef{
					ChartName:    ptr.To("grafana"),
					ChartVersion: ptr.To("11.0.0"),
					Repo:         ptr.To("https://grafana.github.io/helm-charts"),
				},
			}
			err := describe.RenderOutput(&buf, app, "table")
			Expect(err).To(BeNil())
			out := buf.String()
			Expect(out).To(ContainSubstring("Chart Name:"))
			Expect(out).To(ContainSubstring("grafana"))
			Expect(out).To(ContainSubstring("Chart Version:"))
			Expect(out).To(ContainSubstring("11.0.0"))
			Expect(out).To(ContainSubstring("Repository:"))
			Expect(out).NotTo(ContainSubstring("Chart URL:"))
		})

		It("renders REQUESTED RESOURCES using Request value preferring it over Limit", func() {
			app := minimalApp()
			gpu := openapi.RequestResourceCores{}
			gpu.Request.Set(ptr.To(2.0))
			gpu.Limit.Set(ptr.To(4.0)) // Limit is set but Request should win
			res := openapi.AIApplicationRequestResources{}
			res.Gpu.Set(&gpu)
			app.RequestedResources.Set(&res)

			err := describe.RenderOutput(&buf, app, "table")
			Expect(err).To(BeNil())
			out := buf.String()
			Expect(out).To(ContainSubstring("REQUESTED RESOURCES"))
			Expect(out).To(ContainSubstring("GPU:"))
			Expect(out).To(ContainSubstring("2"))
			Expect(out).NotTo(ContainSubstring("4")) // Limit must not appear
		})

		It("falls back to Limit when Request is unset for REQUESTED RESOURCES", func() {
			app := minimalApp()
			gpu := openapi.RequestResourceCores{}
			gpu.Limit.Set(ptr.To(3.0))
			res := openapi.AIApplicationRequestResources{}
			res.Gpu.Set(&gpu)
			app.RequestedResources.Set(&res)

			err := describe.RenderOutput(&buf, app, "table")
			Expect(err).To(BeNil())
			Expect(buf.String()).To(ContainSubstring("3"))
		})

		It("omits REQUESTED RESOURCES section when all values are empty", func() {
			app := minimalApp()
			res := openapi.AIApplicationRequestResources{} // no fields set
			app.RequestedResources.Set(&res)

			err := describe.RenderOutput(&buf, app, "table")
			Expect(err).To(BeNil())
			Expect(buf.String()).NotTo(ContainSubstring("REQUESTED RESOURCES"))
		})

		It("renders ALLOCATED RESOURCES when values are present", func() {
			app := minimalApp()
			alloc := openapi.AIApplicationAllocatedResources{}
			alloc.Gpu.Set(ptr.To(1.5))
			alloc.GpuMemory.Set(ptr.To("3Gi"))
			app.AllocatedResources.Set(&alloc)

			err := describe.RenderOutput(&buf, app, "table")
			Expect(err).To(BeNil())
			out := buf.String()
			Expect(out).To(ContainSubstring("ALLOCATED RESOURCES"))
			Expect(out).To(ContainSubstring("1.5"))
			Expect(out).To(ContainSubstring("3Gi"))
		})

		It("omits ALLOCATED RESOURCES section when all values are empty", func() {
			app := minimalApp()
			alloc := openapi.AIApplicationAllocatedResources{} // no fields set
			app.AllocatedResources.Set(&alloc)

			err := describe.RenderOutput(&buf, app, "table")
			Expect(err).To(BeNil())
			Expect(buf.String()).NotTo(ContainSubstring("ALLOCATED RESOURCES"))
		})

		It("renders RESOURCES (N) table when resources are present", func() {
			app := minimalApp()
			app.Resources = []openapi.AIApplicationResource{
				{Name: ptr.To("grafana-server"), Kind: ptr.To("Deployment"), Id: ptr.To("res-id-1")},
			}
			err := describe.RenderOutput(&buf, app, "table")
			Expect(err).To(BeNil())
			out := buf.String()
			Expect(out).To(ContainSubstring("RESOURCES (1)"))
			Expect(out).To(ContainSubstring("grafana-server"))
			Expect(out).To(ContainSubstring("Deployment"))
		})

		It("omits RESOURCES section when the list is empty", func() {
			err := describe.RenderOutput(&buf, minimalApp(), "table")
			Expect(err).To(BeNil())
			Expect(buf.String()).NotTo(ContainSubstring("RESOURCES"))
		})

		It("uses RFC3339 format (with timezone) for Created At", func() {
			err := describe.RenderOutput(&buf, minimalApp(), "table")
			Expect(err).To(BeNil())
			// RFC3339 timestamps contain a 'T' separator and timezone offset or 'Z'
			Expect(buf.String()).To(MatchRegexp(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`))
		})
	})

	Context("json format", func() {
		It("produces valid JSON with correct field values", func() {
			err := describe.RenderOutput(&buf, minimalApp(), "json")
			Expect(err).To(BeNil())
			var decoded map[string]interface{}
			Expect(json.Unmarshal(buf.Bytes(), &decoded)).To(Succeed())
			Expect(decoded["name"]).To(Equal("my-app"))
			Expect(decoded["type"]).To(Equal("inference"))
		})
	})

	Context("yaml format", func() {
		It("produces valid YAML with correct field values", func() {
			err := describe.RenderOutput(&buf, minimalApp(), "yaml")
			Expect(err).To(BeNil())
			var decoded map[string]interface{}
			Expect(yaml.Unmarshal(buf.Bytes(), &decoded)).To(Succeed())
			Expect(decoded["name"]).To(Equal("my-app"))
			Expect(decoded["type"]).To(Equal("inference"))
		})
	})

	Context("empty format string", func() {
		It("defaults to table output", func() {
			err := describe.RenderOutput(&buf, minimalApp(), "")
			Expect(err).To(BeNil())
			Expect(buf.String()).To(ContainSubstring("GENERAL"))
		})
	})

	Context("unsupported format", func() {
		It("returns an error for an unknown format", func() {
			err := describe.RenderOutput(&buf, minimalApp(), "xml")
			Expect(err).ToNot(BeNil())
			Expect(strings.ToLower(err.Error())).To(ContainSubstring("unsupported"))
		})
	})
})
