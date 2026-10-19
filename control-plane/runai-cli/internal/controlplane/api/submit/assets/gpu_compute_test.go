/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package assets

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	workspaceclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workspace/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("GPU overrides with asset and template references", func() {
	DescribeTable("preserves references and sends only the explicit compute overrides",
		func(source, args, expected string) {
			gs := spec.NewWorkspaceSpec()
			command := &cobra.Command{Use: "workspace"}
			project := "project"
			gs.SetFlags(command, &project)
			Expect(command.ParseFlags(strings.Fields("--dry-run " + args))).To(Succeed())
			if source == "asset" {
				gs.SetComputeAsset("compute-id")
			} else {
				gs.SetTemplateId("template-id")
			}
			client, err := workspaceclient.NewWorkspaceClient("http://localhost", "")
			Expect(err).NotTo(HaveOccurred())
			body, err := client.Submit(context.Background(), "workload", project, "cluster", gs)
			Expect(err).NotTo(HaveOccurred())
			var request struct {
				TemplateID string `json:"templateId"`
				Assets     struct {
					Compute string `json:"compute"`
				} `json:"assets"`
				Spec struct {
					Compute json.RawMessage `json:"compute"`
				} `json:"spec"`
			}
			Expect(json.Unmarshal([]byte(body), &request)).To(Succeed())
			if source == "asset" {
				Expect(request.Assets.Compute).To(Equal("compute-id"))
			} else {
				Expect(request.TemplateID).To(Equal("template-id"))
			}
			if expected == "" {
				Expect(request.Spec.Compute).To(BeEmpty())
			} else {
				Expect(request.Spec.Compute).To(MatchJSON(expected))
			}
		},
		Entry("asset without overrides", "asset", "", ""),
		Entry("template without overrides", "template", "", ""),
		Entry("asset with new memory names", "asset", "--gpu-memory-portion-request 0.5 --gpu-memory-portion-limit 0.75", `{"gpuMemoryPortionRequest":0.5,"gpuMemoryPortionLimit":0.75}`),
		Entry("asset with old memory names", "asset", "--gpu-portion-request 0.5 --gpu-portion-limit 0.75", `{"gpuPortionRequest":0.5,"gpuPortionLimit":0.75}`),
		Entry("template with new memory names", "template", "--gpu-memory-portion-request 0.5", `{"gpuMemoryPortionRequest":0.5}`),
		Entry("template with old memory names", "template", "--gpu-portion-request 0.5", `{"gpuPortionRequest":0.5}`),
		Entry("explicit mode overriding a template", "template", "--gpu-compute-mode sm-sharing", `{"gpuComputeMode":"sm-sharing"}`),
	)
})
