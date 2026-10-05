/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package render_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/inference-endpoint-controller/internal/render"
)

// goldenDir holds the hand-written Track A manifests. They are the oracle: the
// renderer must reproduce them exactly for the llama3 input.
const goldenDir = "../../test/testdata/golden"

var (
	goldenStack   = []string{"03-inferencepool.yaml", "04-epp.yaml", "05-httproute.yaml"}
	goldenRewrite = []string{"07-phase1-modelrewrite.yaml"}
)

const (
	appLabel     = "app"
	llama3Name   = "llama3"
	llama3Ns     = "model-serving"
	llama3Labels = "vllm-llama3"
	configMapKey = "default-plugins.yaml"

	pluginQueue   = "queue-scorer"
	pluginKVCache = "kv-cache-utilization-scorer"
	pluginPrefix  = "prefix-cache-scorer"
	keyPluginRef  = "pluginRef"
	keyWeight     = "weight"
)

// llama3Input is the Track A endpoint, the same values as config/samples.
// Scorers are left empty so the defaults (30/20/10) are exercised.
func llama3Input() render.Input {
	return render.Input{
		Name:             llama3Name,
		Namespace:        llama3Ns,
		Selector:         map[string]string{appLabel: llama3Labels},
		Port:             8000,
		GatewayName:      "inference-agw",
		GatewayNamespace: llama3Ns,
		Hostnames:        []string{"llama3.test"},
		EppImage:         "registry.k8s.io/gateway-api-inference-extension/epp:v1.5.0",
	}
}

// llama3Canary is Track A step 7: 90 % llama3, 10 % llama3-v2.
func llama3Canary() []render.Rewrite {
	return []render.Rewrite{{
		Model:   llama3Name,
		Targets: []render.RewriteTarget{{Model: llama3Name, Weight: 90}, {Model: "llama3-v2", Weight: 10}},
	}}
}

var _ = Describe("Render", func() {
	DescribeTable("keeps YAML document markers inside user-provided strings", func(field string) {
		in := llama3Input()
		payload := "value\n---\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: injected\n  namespace: model-serving\n"
		expectedCount := 10
		switch field {
		case "hostname":
			in.Hostnames = []string{payload}
		case "image":
			in.EppImage = payload
		case "rewrite":
			in.Rewrites = []render.Rewrite{{Model: payload, Targets: []render.RewriteTarget{{Model: payload, Weight: 1}}}}
			expectedCount++
		}
		objects, err := render.Render(in)
		Expect(err).NotTo(HaveOccurred())
		Expect(objects).To(HaveLen(expectedCount))
		for _, object := range objects {
			Expect(object.GetName()).NotTo(Equal("injected"))
			Expect(object.GetNamespace()).To(Equal(in.Namespace))
		}
	}, Entry("hostname", "hostname"), Entry("image", "image"), Entry("rewrite models", "rewrite"))

	It("preserves scalar-looking strings in routes, images, and rewrites", func() {
		in := llama3Input()
		in.Hostnames = []string{"yes", "123"}
		in.GatewayName = "yes"
		in.EppImage = "123"
		in.Rewrites = []render.Rewrite{{Model: "true", Targets: []render.RewriteTarget{{Model: "123", Weight: 1}}}}
		objects, err := render.Render(in)
		Expect(err).NotTo(HaveOccurred())
		indexed := byKey(objects)
		route := indexed["HTTPRoute/llama3-route"]
		hostnames, _, err := unstructured.NestedStringSlice(route.Object, "spec", "hostnames")
		Expect(err).NotTo(HaveOccurred())
		Expect(hostnames).To(Equal(in.Hostnames))
		parents, _, err := unstructured.NestedSlice(route.Object, "spec", "parentRefs")
		Expect(err).NotTo(HaveOccurred())
		Expect(parents[0].(map[string]interface{})["name"]).To(Equal("yes"))
		containers, _, err := unstructured.NestedSlice(indexed["Deployment/llama3-epp"].Object, "spec", "template", "spec", "containers")
		Expect(err).NotTo(HaveOccurred())
		Expect(containers[0].(map[string]interface{})["image"]).To(Equal("123"))
		rules, _, err := unstructured.NestedSlice(indexed["InferenceModelRewrite/llama3-rewrite"].Object, "spec", "rules")
		Expect(err).NotTo(HaveOccurred())
		rule := rules[0].(map[string]interface{})
		Expect(rule["matches"].([]interface{})[0].(map[string]interface{})["model"].(map[string]interface{})["value"]).To(Equal("true"))
		Expect(rule["targets"].([]interface{})[0].(map[string]interface{})["modelRewrite"]).To(Equal("123"))
	})

	Context("for the Track A llama3 endpoint", func() {
		var rendered map[string]*unstructured.Unstructured

		BeforeEach(func() {
			objs, err := render.Render(llama3Input())
			Expect(err).NotTo(HaveOccurred())
			rendered = byKey(objs)
		})

		It("reproduces every namespaced golden object exactly", func() {
			golden := loadGolden(goldenStack...)
			Expect(rendered).To(HaveLen(len(golden)), "rendered object count must equal the golden namespaced object count")

			for k, want := range golden {
				got, ok := rendered[k]
				Expect(ok).To(BeTrue(), "missing rendered object %s", k)
				normalizeForGolden(want, got)
				Expect(got.Object).To(Equal(want.Object), "%s differs from golden", k)
			}
			for k := range rendered {
				Expect(golden).To(HaveKey(k), "unexpected rendered object %s", k)
			}
		})

		It("labels every object with the endpoint and keeps it in the endpoint namespace", func() {
			for k, obj := range rendered {
				Expect(obj.GetNamespace()).To(Equal(llama3Ns), k)
				Expect(obj.GetLabels()).To(HaveKeyWithValue(render.EndpointLabel, llama3Name), k)
			}
		})

		It("stamps the EPP Deployment pod template with a hash of the ConfigMap", func() {
			hash := configHash(rendered["Deployment/llama3-epp"])
			Expect(hash).To(HaveLen(64), "sha256 hex")

			again, err := render.Render(llama3Input())
			Expect(err).NotTo(HaveOccurred())
			Expect(configHash(byKey(again)["Deployment/llama3-epp"])).To(Equal(hash), "hash is deterministic")
		})

		It("renders no InferenceModelRewrite when there are no rewrites", func() {
			Expect(rendered).NotTo(HaveKey("InferenceModelRewrite/llama3-rewrite"))
		})
	})

	Context("scorer weights", func() {
		It("writes the scorers and weights into the EPP config and changes the hash", func() {
			in := llama3Input()
			in.Scorers = []render.Scorer{
				{Plugin: pluginQueue, Weight: 0},
				{Plugin: pluginPrefix, Weight: 100},
			}
			objs, err := render.Render(in)
			Expect(err).NotTo(HaveOccurred())
			rendered := byKey(objs)

			cfg := eppConfig(rendered["ConfigMap/llama3-epp"])
			plugins := cfg["plugins"].([]any)
			Expect(plugins[0]).To(Equal(map[string]any{keyType: pluginQueue}))
			Expect(plugins[1]).To(Equal(map[string]any{keyType: pluginPrefix}))
			Expect(plugins).To(HaveLen(4), "2 scorers + metrics-data-source + core-metrics-extractor")

			profile := cfg["schedulingProfiles"].([]any)[0].(map[string]any)
			Expect(profile["plugins"]).To(Equal([]any{
				map[string]any{keyPluginRef: pluginQueue, keyWeight: float64(0)},
				map[string]any{keyPluginRef: pluginPrefix, keyWeight: float64(100)},
			}))

			defaultHash := configHash(byKey(must(render.Render(llama3Input())))["Deployment/llama3-epp"])
			Expect(configHash(rendered["Deployment/llama3-epp"])).NotTo(Equal(defaultHash), "a weight change must roll the Deployment")
		})

		It("uses queue 30, kvCache 20, prefixCache 10 when no scorers are given", func() {
			cfg := eppConfig(byKey(must(render.Render(llama3Input())))["ConfigMap/llama3-epp"])
			profile := cfg["schedulingProfiles"].([]any)[0].(map[string]any)
			Expect(profile["plugins"]).To(Equal([]any{
				map[string]any{keyPluginRef: pluginQueue, keyWeight: float64(30)},
				map[string]any{keyPluginRef: pluginKVCache, keyWeight: float64(20)},
				map[string]any{keyPluginRef: pluginPrefix, keyWeight: float64(10)},
			}))
		})
	})

	Context("rewrites", func() {
		It("reproduces the Track A canary rewrite golden exactly", func() {
			in := llama3Input()
			in.Rewrites = llama3Canary()
			rendered := byKey(must(render.Render(in)))
			Expect(rendered).To(HaveLen(11), "10 stack objects + 1 InferenceModelRewrite")

			golden := loadGolden(goldenRewrite...)
			want := golden["InferenceModelRewrite/llama3-rewrite"]
			Expect(want).NotTo(BeNil())
			got := rendered["InferenceModelRewrite/llama3-rewrite"]
			Expect(got).NotTo(BeNil())
			normalizeForGolden(want, got)
			Expect(got.Object).To(Equal(want.Object))
		})

		It("puts several rewrites into one object as separate rules", func() {
			in := llama3Input()
			in.Rewrites = append(llama3Canary(), render.Rewrite{
				Model: "gpt-4", Targets: []render.RewriteTarget{{Model: llama3Name, Weight: 1}},
			})
			rw := byKey(must(render.Render(in)))["InferenceModelRewrite/llama3-rewrite"]
			rules, _, _ := unstructured.NestedSlice(rw.Object, "spec", "rules")
			Expect(rules).To(HaveLen(2))
			match := rules[1].(map[string]any)["matches"].([]any)[0].(map[string]any)["model"].(map[string]any)
			Expect(match["value"]).To(Equal("gpt-4"))
		})

		It("does not change the EPP config hash", func() {
			in := llama3Input()
			in.Rewrites = llama3Canary()
			withRewrite := configHash(byKey(must(render.Render(in)))["Deployment/llama3-epp"])
			without := configHash(byKey(must(render.Render(llama3Input())))["Deployment/llama3-epp"])
			Expect(withRewrite).To(Equal(without), "rewrites are live objects and must not restart the EPP")
		})
	})

	Context("for a different endpoint", func() {
		in := render.Input{
			Name:             "chat",
			Namespace:        "team-a",
			Selector:         map[string]string{appLabel: "vllm-chat", "tier": "gold"},
			Port:             9000,
			GatewayName:      "public-gw",
			GatewayNamespace: "gateways",
			Hostnames:        []string{"chat.example.com", "chat.internal"},
			EppImage:         "ghcr.io/llm-d/llm-d-inference-scheduler:v0.10.0",
		}
		var rendered map[string]*unstructured.Unstructured

		BeforeEach(func() {
			objs, err := render.Render(in)
			Expect(err).NotTo(HaveOccurred())
			rendered = byKey(objs)
		})

		It("substitutes the selector, port and EPP reference into the pool", func() {
			pool := rendered["InferencePool/chat-pool"]
			Expect(pool).NotTo(BeNil())
			selector, _, _ := unstructured.NestedStringMap(pool.Object, "spec", "selector", "matchLabels")
			Expect(selector).To(Equal(in.Selector))
			ports, _, _ := unstructured.NestedSlice(pool.Object, "spec", "targetPorts")
			Expect(ports).To(ConsistOf(map[string]any{"number": int64(9000)}))
			ref, _, _ := unstructured.NestedString(pool.Object, "spec", "endpointPickerRef", keyName)
			Expect(ref).To(Equal("chat-epp"))
		})

		It("substitutes the gateway, hostnames and pool reference into the route", func() {
			route := rendered["HTTPRoute/chat-route"]
			Expect(route).NotTo(BeNil())
			hosts, _, _ := unstructured.NestedStringSlice(route.Object, "spec", "hostnames")
			Expect(hosts).To(Equal(in.Hostnames))
			parents, _, _ := unstructured.NestedSlice(route.Object, "spec", "parentRefs")
			Expect(parents).To(ConsistOf(map[string]any{keyName: "public-gw", "namespace": "gateways"}))
			rules, _, _ := unstructured.NestedSlice(route.Object, "spec", "rules")
			Expect(rules).To(HaveLen(1))
			backends := rules[0].(map[string]any)["backendRefs"].([]any)
			Expect(backends[0].(map[string]any)[keyName]).To(Equal("chat-pool"))
		})

		It("substitutes the image, pool flags and service account into the EPP deployment", func() {
			dep := rendered["Deployment/chat-epp"]
			Expect(dep).NotTo(BeNil())
			containers, _, _ := unstructured.NestedSlice(dep.Object, "spec", "template", "spec", "containers")
			c := containers[0].(map[string]any)
			Expect(c["image"]).To(Equal(in.EppImage))
			args := c["args"].([]any)
			Expect(args).To(ContainElements("--pool-name", "chat-pool", "--pool-namespace", "team-a"))
			sa, _, _ := unstructured.NestedString(dep.Object, "spec", "template", "spec", "serviceAccountName")
			Expect(sa).To(Equal("chat-epp"))
		})

		It("names every EPP stack object after the endpoint", func() {
			for _, k := range []string{"ServiceAccount/chat-epp", "ConfigMap/chat-epp", "Service/chat-epp",
				"Role/chat-epp-non-sa", "Role/chat-epp-sa", "RoleBinding/chat-epp-non-sa", "RoleBinding/chat-epp-sa"} {
				Expect(rendered).To(HaveKey(k))
			}
		})
	})

	DescribeTable("preserves selector labels as strings",
		func(key, value string) {
			in := llama3Input()
			in.Selector[key] = value
			objs, err := render.Render(in)
			Expect(err).NotTo(HaveOccurred())
			pool := byKey(objs)["InferencePool/llama3-pool"]
			Expect(pool).NotTo(BeNil())
			selector, found, err := unstructured.NestedStringMap(pool.Object, "spec", "selector", "matchLabels")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(selector).To(Equal(in.Selector))
		},
		Entry("boolean true", "selected", "true"),
		Entry("boolean false", "selected", "false"),
		Entry("YAML boolean alias", "selected", "yes"),
		Entry("null", "selected", "null"),
		Entry("integer", "selected", "123"),
		Entry("leading zeros", "selected", "001"),
		Entry("scientific notation", "selected", "1e3"),
		Entry("empty value", "selected", ""),
		Entry("numeric key with leading zeros", "001", "target"),
	)

	DescribeTable("rejects incomplete input",
		func(mutate func(*render.Input)) {
			in := llama3Input()
			mutate(&in)
			_, err := render.Render(in)
			Expect(err).To(HaveOccurred())
		},
		Entry("no name", func(in *render.Input) { in.Name = "" }),
		Entry("no namespace", func(in *render.Input) { in.Namespace = "" }),
		Entry("no selector", func(in *render.Input) { in.Selector = nil }),
		Entry("no hostnames", func(in *render.Input) { in.Hostnames = nil }),
		Entry("no gateway", func(in *render.Input) { in.GatewayName = "" }),
		Entry("no image", func(in *render.Input) { in.EppImage = "" }),
		Entry("zero port", func(in *render.Input) { in.Port = 0 }),
		Entry("scorer without plugin", func(in *render.Input) { in.Scorers = []render.Scorer{{Plugin: "", Weight: 1}} }),
		Entry("rewrite without targets", func(in *render.Input) { in.Rewrites = []render.Rewrite{{Model: "x"}} }),
	)

	It("defaults the gateway namespace to the endpoint namespace", func() {
		in := llama3Input()
		in.GatewayNamespace = ""
		route := byKey(must(render.Render(in)))["HTTPRoute/llama3-route"]
		parents, _, _ := unstructured.NestedSlice(route.Object, "spec", "parentRefs")
		Expect(parents[0].(map[string]any)["namespace"]).To(Equal(llama3Ns))
	})
})

// loadGolden reads the namespaced objects out of the given Track A files. The
// three cluster-scoped RBAC objects in 04-epp.yaml are metrics-auth helpers the
// controller does not generate in v0, so they are skipped.
func loadGolden(files ...string) map[string]*unstructured.Unstructured {
	out := map[string]*unstructured.Unstructured{}
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(goldenDir, name))
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "read golden %s", name)
		dec := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
		for {
			obj := &unstructured.Unstructured{}
			if err := dec.Decode(obj); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				Fail("decode golden " + name + ": " + err.Error())
			}
			if obj.Object == nil || obj.GetKind() == "" {
				continue
			}
			if obj.GetNamespace() == "" {
				continue // cluster-scoped, out of v0 scope
			}
			out[key(obj)] = obj
		}
	}
	return out
}

// normalizeForGolden removes what legitimately differs between a hand-written
// golden object and a rendered one: the lab's tracking label, our endpoint
// label, our config-hash annotation, and comments inside the EPP config
// document (compared semantically instead).
func normalizeForGolden(want, got *unstructured.Unstructured) {
	stripLabel(want, "lab.track")
	stripLabel(got, render.EndpointLabel)
	if got.GetKind() == "Deployment" {
		stripPodTemplateAnnotation(got, render.ConfigHashAnnotation)
	}
	if got.GetKind() == "ConfigMap" {
		wantCfg, gotCfg := eppConfig(want), eppConfig(got)
		ExpectWithOffset(1, gotCfg).To(Equal(wantCfg), "EPP config document differs from golden")
		unstructured.RemoveNestedField(want.Object, "data", configMapKey)
		unstructured.RemoveNestedField(got.Object, "data", configMapKey)
	}
}

// eppConfig parses the EndpointPickerConfig document out of the EPP ConfigMap.
func eppConfig(cm *unstructured.Unstructured) map[string]any {
	text, _, _ := unstructured.NestedString(cm.Object, "data", configMapKey)
	ExpectWithOffset(1, text).NotTo(BeEmpty())
	var out map[string]any
	ExpectWithOffset(1, yaml.Unmarshal([]byte(text), &out)).To(Succeed())
	return out
}

// configHash reads the config-hash annotation off the EPP Deployment's pod template.
func configHash(dep *unstructured.Unstructured) string {
	ExpectWithOffset(1, dep).NotTo(BeNil())
	annotations, _, _ := unstructured.NestedStringMap(dep.Object, "spec", "template", "metadata", "annotations")
	return annotations[render.ConfigHashAnnotation]
}

func stripPodTemplateAnnotation(dep *unstructured.Unstructured, name string) {
	annotations, _, _ := unstructured.NestedStringMap(dep.Object, "spec", "template", "metadata", "annotations")
	delete(annotations, name)
	if len(annotations) == 0 {
		unstructured.RemoveNestedField(dep.Object, "spec", "template", "metadata", "annotations")
		return
	}
	_ = unstructured.SetNestedStringMap(dep.Object, annotations, "spec", "template", "metadata", "annotations")
}

func must(objs []*unstructured.Unstructured, err error) []*unstructured.Unstructured {
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return objs
}

func key(obj *unstructured.Unstructured) string { return obj.GetKind() + "/" + obj.GetName() }

func byKey(objs []*unstructured.Unstructured) map[string]*unstructured.Unstructured {
	out := map[string]*unstructured.Unstructured{}
	for _, obj := range objs {
		out[key(obj)] = obj
	}
	return out
}

func stripLabel(obj *unstructured.Unstructured, label string) {
	labels := obj.GetLabels()
	delete(labels, label)
	if len(labels) == 0 {
		unstructured.RemoveNestedField(obj.Object, "metadata", "labels")
		return
	}
	obj.SetLabels(labels)
}
