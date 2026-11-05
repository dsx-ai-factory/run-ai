/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package render_test

import (
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/inference-endpoint-controller/internal/render"
)

const (
	balancedProfile     = "balanced"
	queueFocusedProfile = "queue-focused"
	keyType             = "type"
	keyName             = "name"
	pluginMaxScore      = "max-score-picker"
	missingProfile      = "missing"
)

func profilesInput() render.Input {
	in := llama3Input()
	in.RuntimeProfiles = true
	in.DefaultProfile, in.ActiveProfile = balancedProfile, balancedProfile
	in.Profiles = []render.SchedulingProfile{
		{Name: balancedProfile, Scorers: []render.Scorer{{Plugin: pluginQueue, Weight: 10}, {Plugin: pluginKVCache, Weight: 30}}},
		{Name: queueFocusedProfile, Scorers: []render.Scorer{{Plugin: pluginQueue, Weight: 60}, {Plugin: pluginKVCache, Weight: 30}}},
	}
	return in
}

var _ = Describe("Runtime profile rendering", func() {
	It("preloads both weight policies, one shared plugin set, and the header handler", func() {
		objects := byKey(must(render.Render(profilesInput())))
		cfg := eppConfig(objects["ConfigMap/llama3-epp"])
		Expect(cfg["apiVersion"]).To(Equal("llm-d.ai/v1alpha1"))
		Expect(cfg["plugins"]).To(ConsistOf(
			map[string]any{keyType: "header-profile-handler", "parameters": map[string]any{"headerName": "EPP-Profile", "defaultProfile": balancedProfile}},
			map[string]any{keyType: pluginMaxScore},
			map[string]any{keyType: pluginQueue}, map[string]any{keyType: pluginKVCache},
			map[string]any{keyType: "core-metrics-extractor"},
			map[string]any{keyType: "metrics-data-source", "parameters": map[string]any{"scheme": "http", "path": "/metrics", "insecureSkipVerify": true}},
		))
		Expect(cfg["schedulingProfiles"]).To(ConsistOf(
			map[string]any{keyName: balancedProfile, "plugins": []any{
				map[string]any{keyPluginRef: pluginKVCache, keyWeight: float64(30)},
				map[string]any{keyPluginRef: pluginQueue, keyWeight: float64(10)},
				map[string]any{keyPluginRef: pluginMaxScore},
			}},
			map[string]any{keyName: queueFocusedProfile, "plugins": []any{
				map[string]any{keyPluginRef: pluginKVCache, keyWeight: float64(30)},
				map[string]any{keyPluginRef: pluginQueue, keyWeight: float64(60)},
				map[string]any{keyPluginRef: pluginMaxScore},
			}},
		))
		containers, _, _ := unstructured.NestedSlice(objects["Deployment/llama3-epp"].Object, "spec", "template", "spec", "containers")
		Expect(containers[0].(map[string]any)["args"]).To(ContainElement("--allow-experimental-plugins"))
		failureMode, _, _ := unstructured.NestedString(objects["InferencePool/llama3-pool"].Object, "spec", "endpointPickerRef", "failureMode")
		Expect(failureMode).To(Equal("FailClose"))
	})

	It("changes only the route when selecting another preloaded profile", func() {
		in := profilesInput()
		before := byKey(must(render.Render(in)))
		in.ActiveProfile = queueFocusedProfile
		after := byKey(must(render.Render(in)))
		for key, original := range before {
			if key == "HTTPRoute/llama3-route" {
				Expect(after[key].Object).NotTo(Equal(original.Object))
			} else {
				Expect(after[key].Object).To(Equal(original.Object), key)
			}
		}
	})

	It("keeps the config hash stable when map-like lists are reordered", func() {
		in := profilesInput()
		before := byKey(must(render.Render(in)))
		slices.Reverse(in.Profiles)
		slices.Reverse(in.Profiles[0].Scorers)
		after := byKey(must(render.Render(in)))
		Expect(after["ConfigMap/llama3-epp"].Object).To(Equal(before["ConfigMap/llama3-epp"].Object))
		Expect(configHash(after["Deployment/llama3-epp"])).To(Equal(configHash(before["Deployment/llama3-epp"])))
		Expect(in.Profiles[0].Name).To(Equal(queueFocusedProfile), "renderer must not reorder caller-owned slices")
	})

	DescribeTable("rolls the EPP for static configuration changes", func(change func(*render.Input)) {
		in := profilesInput()
		before := byKey(must(render.Render(in)))
		change(&in)
		after := byKey(must(render.Render(in)))
		Expect(configHash(after["Deployment/llama3-epp"])).NotTo(Equal(configHash(before["Deployment/llama3-epp"])))
	},
		Entry("weights", func(in *render.Input) { in.Profiles[0].Scorers[0].Weight = 20 }),
		Entry("default profile", func(in *render.Input) { in.DefaultProfile = queueFocusedProfile }),
		Entry("plugin selection", func(in *render.Input) { in.Profiles[0].Scorers[0].Plugin = pluginPrefix }),
	)

	DescribeTable("rejects invalid catalogs even without API admission", func(change func(*render.Input)) {
		in := profilesInput()
		change(&in)
		_, err := render.Render(in)
		Expect(err).To(HaveOccurred())
	},
		Entry("unsupported implementation", func(in *render.Input) { in.RuntimeProfiles = false }),
		Entry("missing catalog", func(in *render.Input) { in.Profiles = nil }),
		Entry("flat scorers list", func(in *render.Input) { in.Scorers = render.DefaultScorers() }),
		Entry("unknown active profile", func(in *render.Input) { in.ActiveProfile = missingProfile }),
		Entry("unknown default profile", func(in *render.Input) { in.DefaultProfile = missingProfile }),
		Entry("duplicate names", func(in *render.Input) { in.Profiles[1].Name = balancedProfile }),
		Entry("invalid name", func(in *render.Input) { in.Profiles[1].Name = "bad\nname" }),
		Entry("missing scorers", func(in *render.Input) { in.Profiles[0].Scorers = nil }),
		Entry("unsupported scorer", func(in *render.Input) { in.Profiles[0].Scorers[0].Plugin = missingProfile }),
		Entry("duplicate scorer", func(in *render.Input) { in.Profiles[0].Scorers[0].Plugin = pluginKVCache }),
		Entry("bad weight", func(in *render.Input) { in.Profiles[0].Scorers[0].Weight = 101 }),
		Entry("all zero weights", func(in *render.Input) {
			for i := range in.Profiles[0].Scorers {
				in.Profiles[0].Scorers[i].Weight = 0
			}
		}),
	)
})
