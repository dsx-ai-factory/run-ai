/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/distributedinference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/nim"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"k8s.io/utils/ptr"
)

var _ = Describe("Display", func() {
	const (
		labelName1  = "label-name-1"
		labelValue1 = "label-value-1"
		labelName2  = "label-name-2"
		labelValue2 = "label-value-2"

		masterLabelName1  = "master-name-1"
		masterLabelValue1 = "master-value-1"

		labelsSectionName       = "Labels"
		workerLabelsSectionName = "Worker Labels"
		masterLabelsSectionName = "Master Labels"
		issuesSectionName       = "Issues"
	)

	var (
		workload          *describe.Workload
		workloadDisplayer *describe.WorkloadDisplayer

		labels       []openapi.Label
		masterLabels []openapi.Label
	)

	BeforeEach(func() {
		workload = &describe.Workload{}

		workloadDisplayer = describe.NewWorkloadDisplayer(workload, displayer.DisplayOptions{})

		labels = []openapi.Label{
			{Name: *openapi.NewNullableString(ptr.To(labelName1)), Value: *openapi.NewNullableString(ptr.To(labelValue1))},
			{Name: *openapi.NewNullableString(ptr.To(labelName2)), Value: *openapi.NewNullableString(ptr.To(labelValue2))},
		}

		masterLabels = []openapi.Label{
			{Name: *openapi.NewNullableString(ptr.To(masterLabelName1)), Value: *openapi.NewNullableString(ptr.To(masterLabelValue1))},
		}
	})

	Context("pods section", func() {
		Context("labels", func() {
			It("labels are displayed", func() {
				workload.Spec = displayer.Spec{
					Labels: lo.ToSlicePtr(labels),
				}

				output := captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)
				Expect(output).To(ContainSubstring(labelsSectionName))
				validateLabelsOutput(labels, output)
			})

			It("master labels are displayed", func() {
				workload.Spec = displayer.Spec{
					Labels:       lo.ToSlicePtr(labels),
					MasterLabels: lo.ToSlicePtr(masterLabels),
				}

				output := captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)
				Expect(output).To(ContainSubstring(workerLabelsSectionName))
				validateLabelsOutput(labels, output)
				Expect(output).To(ContainSubstring(masterLabelsSectionName))
				validateLabelsOutput(masterLabels, output)
			})

			It("no labels", func() {
				workload.Spec = displayer.Spec{}

				output := captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)
				Expect(output).To(Not(ContainSubstring("Labels")))
			})

			It("more labels then max display count", func() {
				targetLabelCount := displayer.MaxDisplayLabelsCount + 1
				manyLabels := make([]openapi.Label, targetLabelCount)
				for i := 0; i < targetLabelCount; i++ {
					manyLabels[i] = openapi.Label{
						Name:  *openapi.NewNullableString(ptr.To(fmt.Sprintf("label-%d", i))),
						Value: *openapi.NewNullableString(ptr.To(fmt.Sprintf("value-%d", i))),
					}
				}

				workload.Spec = displayer.Spec{
					Labels: lo.ToSlicePtr(manyLabels),
				}

				output := captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)
				Expect(output).To(ContainSubstring(labelsSectionName))
				Expect(output).To(ContainSubstring(fmt.Sprintf("Displaying %d/%d labels", displayer.MaxDisplayLabelsCount, targetLabelCount)))
			})
		})
		Context("issues", func() {
			It("no issues - nil", func() {
				workload.Issues = nil
				output := captureOutputWithError(workloadDisplayer.AsTable)
				Expect(output).NotTo(ContainSubstring(issuesSectionName))
			})

			It("no issues - empty", func() {
				workload.Issues = []displayer.Issue{}
				output := captureOutputWithError(workloadDisplayer.AsTable)
				Expect(output).NotTo(ContainSubstring(issuesSectionName))
			})

			It("display issues", func() {
				computeIssue := displayer.Issue{
					Type:        "Compute",
					Kind:        constants.KindRunaiJob,
					Name:        "job1",
					Description: "abc",
				}
				workload.Issues = []displayer.Issue{computeIssue}

				output := captureOutputWithError(workloadDisplayer.AsTable)
				Expect(output).To(ContainSubstring(issuesSectionName))
				Expect(output).Should(MatchRegexp(`%s\s+%s\s+%s\s+%s`, computeIssue.Type, computeIssue.Kind, computeIssue.Name, computeIssue.Description))
			})
		})
	})
	Context("GPU compute mode", func() {
		BeforeEach(func() {
			workloadDisplayer = describe.NewWorkloadDisplayer(workload, displayer.DisplayOptions{ShowGeneral: true, ShowCompute: true})
		})

		It("omits the mode when it is not provided", func() {
			Expect(captureOutputWithError(workloadDisplayer.AsTable)).NotTo(ContainSubstring("GPU Compute Mode"))
			Expect(captureOutputWithError(workloadDisplayer.AsYAML)).NotTo(ContainSubstring("gpucomputemode"))
		})

		It("shows the mode under compute resources in table output and preserves JSON and YAML output", func() {
			workload.Compute.GPUComputeMode = ptr.To("sm-sharing")
			Expect(captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)).NotTo(ContainSubstring("GPU Compute Mode"))
			Expect(captureOutputWithError(workloadDisplayer.AsTable)).To(MatchRegexp(`Compute Resources\s+GPU Compute Mode:\s+sm-sharing`))
			Expect(captureOutputWithError(workloadDisplayer.AsJSON)).To(ContainSubstring(`"gpuComputeMode": "sm-sharing"`))
			Expect(captureOutputWithError(workloadDisplayer.AsYAML)).To(ContainSubstring("gpucomputemode: sm-sharing"))
		})

		It("hides the mode when compute resources are hidden", func() {
			workload.Compute.GPUComputeMode = ptr.To("sm-sharing")
			workloadDisplayer = describe.NewWorkloadDisplayer(workload, displayer.DisplayOptions{ShowGeneral: true})
			Expect(captureOutputWithError(workloadDisplayer.AsTable)).NotTo(ContainSubstring("GPU Compute Mode"))
		})

		It("shows separate master and worker modes under their compute resources", func() {
			workload.Compute.GPUComputeMode = ptr.To("time-slicing")
			workload.MasterConfig = &displayer.WorkloadRuntimeConfig{Compute: &displayer.Compute{GPUComputeMode: ptr.To("sm-sharing")}}
			output := captureOutputWithError(workloadDisplayer.AsTable)
			Expect(output).To(MatchRegexp(`(?m)^    Compute Resources\n    GPU Compute Mode:\s+sm-sharing`))
			Expect(output).To(MatchRegexp(`(?m)^Compute Resources\n\n  GPU Compute Mode:\s+time-slicing`))
			Expect(strings.Count(output, "GPU Compute Mode:")).To(Equal(2))
		})

		It("shows separate leader and worker modes under their compute resources", func() {
			workload.DistributedInference = &distributedinference.Fields{
				Leader: &displayer.WorkloadRuntimeConfig{Compute: &displayer.Compute{GPUComputeMode: ptr.To("time-slicing")}},
				Worker: &displayer.WorkloadRuntimeConfig{Compute: &displayer.Compute{GPUComputeMode: ptr.To("sm-sharing")}},
			}
			output := captureOutputWithError(workloadDisplayer.AsTable)
			Expect(output).To(MatchRegexp(`Leader Configuration\s+Compute Resources\n    GPU Compute Mode:\s+time-slicing`))
			Expect(output).To(MatchRegexp(`Worker Configuration\s+Compute Resources\n    GPU Compute Mode:\s+sm-sharing`))
			Expect(strings.Count(output, "GPU Compute Mode:")).To(Equal(2))
		})

		It("shows the NIM mode under compute resources using its existing configuration fields", func() {
			workload.Nim = &nim.Fields{GpuComputeMode: ptr.To("sm-sharing")}
			Expect(captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)).NotTo(ContainSubstring("GPU Compute Mode"))
			Expect(captureOutputWithError(workloadDisplayer.AsTable)).To(MatchRegexp(`Compute Resources\s+GPU Compute Mode:\s+sm-sharing`))
		})
	})

	Context("network topology section", func() {
		It("network topology is displayed when set", func() {
			workload.NetworkTopology = &displayer.NetworkTopology{
				Name:               ptr.To("my-topology"),
				RequiredPlacement:  ptr.To("rack"),
				PreferredPlacement: ptr.To("block"),
				ActualPlacement:    ptr.To("rack (rack-1)"),
			}

			output := captureOutputWithError(workloadDisplayer.AsTable)
			Expect(output).To(ContainSubstring("Network Topology"))
			Expect(output).To(ContainSubstring("my-topology"))
			Expect(output).To(ContainSubstring("rack"))
			Expect(output).To(ContainSubstring("block"))
			Expect(output).To(ContainSubstring("rack (rack-1)"))
		})

		It("network topology is not displayed when nil", func() {
			workload.NetworkTopology = nil

			output := captureOutputWithError(workloadDisplayer.AsTable)
			Expect(output).NotTo(ContainSubstring("Network Topology"))
		})
	})

	Context("annotations", func() {
		It("annotations are displayed", func() {
			annotations := []openapi.Annotation{
				{Name: *openapi.NewNullableString(ptr.To("annotation1")), Value: *openapi.NewNullableString(ptr.To("value1"))},
				{Name: *openapi.NewNullableString(ptr.To("annotation2")), Value: *openapi.NewNullableString(ptr.To("value2"))},
			}

			workload.Spec = displayer.Spec{
				Annotations: lo.ToSlicePtr(annotations),
			}

			output := captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)
			Expect(output).To(ContainSubstring("Annotations"))
			validateAnnotationsOutput(annotations, output)
		})

		It("runai-submit-command annotation is displayed", func() {
			annotations := []openapi.Annotation{
				{Name: *openapi.NewNullableString(ptr.To("runai-submit-command")), Value: *openapi.NewNullableString(ptr.To("runai training submit test --image pytorch"))},
			}

			workload.Spec = displayer.Spec{
				Annotations: lo.ToSlicePtr(annotations),
			}

			output := captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)
			Expect(output).To(ContainSubstring("Annotations"))
			Expect(output).To(ContainSubstring("runai-submit-command:runai training submit test --image pytorch"))
		})

		It("no annotations when nil", func() {
			workload.Spec = displayer.Spec{
				Annotations: nil,
			}

			output := captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)
			Expect(output).NotTo(ContainSubstring("Annotations"))
		})

		It("no annotations when empty", func() {
			workload.Spec = displayer.Spec{
				Annotations: []*openapi.Annotation{},
			}

			output := captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)
			Expect(output).NotTo(ContainSubstring("Annotations"))
		})

		It("labels and annotations are both displayed", func() {
			workload.Spec = displayer.Spec{
				Labels:      lo.ToSlicePtr(labels),
				Annotations: lo.ToSlicePtr([]openapi.Annotation{{Name: *openapi.NewNullableString(ptr.To("ann")), Value: *openapi.NewNullableString(ptr.To("val"))}}),
			}

			output := captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)
			Expect(output).To(ContainSubstring(labelsSectionName))
			Expect(output).To(ContainSubstring("Annotations"))
			validateLabelsOutput(labels, output)
			Expect(output).To(ContainSubstring("ann:val"))
		})

		It("more annotations than max display count", func() {
			targetAnnotationCount := displayer.MaxDisplayLabelsCount + 1
			manyAnnotations := make([]openapi.Annotation, targetAnnotationCount)
			for i := 0; i < targetAnnotationCount; i++ {
				manyAnnotations[i] = openapi.Annotation{
					Name:  *openapi.NewNullableString(ptr.To(fmt.Sprintf("annotation-%d", i))),
					Value: *openapi.NewNullableString(ptr.To(fmt.Sprintf("value-%d", i))),
				}
			}

			workload.Spec = displayer.Spec{
				Annotations: lo.ToSlicePtr(manyAnnotations),
			}

			output := captureOutput(workloadDisplayer.DisplayGeneralAndRuntime)
			Expect(output).To(ContainSubstring("Annotations"))
			Expect(output).To(ContainSubstring(fmt.Sprintf("Displaying %d/%d labels", displayer.MaxDisplayLabelsCount, targetAnnotationCount)))
		})
	})
})

func captureOutput(f func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()
	w.Close()
	out, _ := io.ReadAll(r)

	os.Stdout = orig
	return string(out)
}

func captureOutputWithError(f func() error) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = f()
	w.Close()
	out, _ := io.ReadAll(r)

	os.Stdout = orig
	return string(out)
}

func validateLabelsOutput(expected []openapi.Label, output string) {
	for _, expectedLabel := range expected {
		Expect(output).To(ContainSubstring(fmt.Sprintf("%s:%s", expectedLabel.GetName(), expectedLabel.GetValue())))
	}
}

func validateAnnotationsOutput(expected []openapi.Annotation, output string) {
	for _, expectedAnnotation := range expected {
		Expect(output).To(ContainSubstring(fmt.Sprintf("%s:%s", expectedAnnotation.GetName(), expectedAnnotation.GetValue())))
	}
}
