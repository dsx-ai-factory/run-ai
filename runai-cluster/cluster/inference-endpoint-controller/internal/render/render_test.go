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
	"strings"

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
	goldenStack = []string{"03-inferencepool.yaml", "04-epp.yaml", "05-httproute.yaml", "08-epp-networkpolicy.yaml"}
)

const (
	appLabel     = "app"
	llama3Name   = "llama3"
	llama3Ns     = "model-serving"
	llama3Labels = "vllm-llama3"
	configMapKey = "default-plugins.yaml"
	metricsNs    = "runai"
	// namespaceNameLabel is the label the API server sets on every Namespace.
	namespaceNameLabel = "kubernetes.io/metadata.name"

	pluginQueue   = "queue-scorer"
	pluginKVCache = "kv-cache-utilization-scorer"
	pluginPrefix  = "prefix-cache-scorer"
	keyPluginRef  = "pluginRef"
	keyType       = "type"
	keyName       = "name"
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
		MetricsNamespace: metricsNs,
		Hostnames:        []string{"llama3.test"},
		PathPrefix:       "/v1",
		EppImage:         "registry.k8s.io/gateway-api-inference-extension/epp:v1.5.0",
	}
}

var _ = Describe("Render", func() {
	DescribeTable("keeps YAML document markers inside user-provided strings", func(field string) {
		in := llama3Input()
		payload := "value\n---\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: injected\n  namespace: model-serving\n"
		expectedCount := 11
		switch field {
		case "hostname":
			in.Hostnames = []string{payload}
		case "image":
			in.EppImage = payload
		case "path prefix":
			in.PathPrefix = "/" + payload
		}
		objects, err := render.Render(in)
		Expect(err).NotTo(HaveOccurred())
		Expect(objects).To(HaveLen(expectedCount))
		for _, object := range objects {
			Expect(object.GetName()).NotTo(Equal("injected"))
			Expect(object.GetNamespace()).To(Equal(in.Namespace))
		}
	}, Entry("hostname", "hostname"), Entry("image", "image"), Entry("path prefix", "path prefix"))

	It("preserves scalar-looking strings in routes and images", func() {
		in := llama3Input()
		in.Hostnames = []string{"yes", "123"}
		in.GatewayName = "yes"
		in.EppImage = "123"
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

	})

	DescribeTable("DNSLabel",
		func(in, want string) { Expect(render.DNSLabel(in)).To(Equal(want)) },
		Entry("already a label", "llama3", "llama3"),
		Entry("upper case and separators", "meta/Llama-3.1_8B", "meta-llama-3-1-8b"),
		Entry("leading and trailing separators", "--chat--", "chat"),
		Entry("only separators", "///", ""),
		Entry("truncated to 63 without a trailing dash", strings.Repeat("a", 62)+"-bcd", strings.Repeat("a", 62)),
	)

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

	Context("path prefix", func() {
		routeRules := func(in render.Input) []interface{} {
			objects, err := render.Render(in)
			Expect(err).NotTo(HaveOccurred())
			rules, _, err := unstructured.NestedSlice(byKey(objects)["HTTPRoute/llama3-route"].Object, "spec", "rules")
			Expect(err).NotTo(HaveOccurred())
			return rules
		}

		It("renders one rule that sends the prefix to the InferencePool", func() {
			in := llama3Input()
			in.PathPrefix = "/openai/v1"
			rules := routeRules(in)
			Expect(rules).To(HaveLen(1))
			rule := rules[0].(map[string]interface{})
			match := rule["matches"].([]interface{})[0].(map[string]interface{})["path"].(map[string]interface{})
			Expect(match).To(HaveKeyWithValue("type", "PathPrefix"))
			Expect(match).To(HaveKeyWithValue("value", "/openai/v1"))
			backend := rule["backendRefs"].([]interface{})[0].(map[string]interface{})
			Expect(backend).To(HaveKeyWithValue("kind", "InferencePool"))
			Expect(backend).To(HaveKeyWithValue("name", "llama3-pool"))
		})

		DescribeTable("rejects an invalid prefix",
			func(prefix string) {
				in := llama3Input()
				in.PathPrefix = prefix
				_, err := render.Render(in)
				Expect(err).To(MatchError(ContainSubstring("pathPrefix")))
			},
			Entry("empty", ""),
			Entry("without a leading slash", "v1"),
		)
	})

	Context("for a different endpoint", func() {
		in := render.Input{
			Name:             "chat",
			Namespace:        "team-a",
			Selector:         map[string]string{appLabel: "vllm-chat", "tier": "gold"},
			Port:             9000,
			GatewayName:      "public-gw",
			GatewayNamespace: "gateways",
			MetricsNamespace: "observability",
			Hostnames:        []string{"chat.example.com", "chat.internal"},
			PathPrefix:       "/chat/v1",
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
			path := rules[0].(map[string]any)["matches"].([]any)[0].(map[string]any)["path"].(map[string]any)
			Expect(path["value"]).To(Equal(in.PathPrefix))
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
				"Role/chat-epp-non-sa", "Role/chat-epp-sa", "RoleBinding/chat-epp-non-sa", "RoleBinding/chat-epp-sa",
				"NetworkPolicy/chat-epp"} {
				Expect(rendered).To(HaveKey(k))
			}
		})

		It("admits ingress to the EPP only from the gateway and metrics namespaces", func() {
			policy := rendered["NetworkPolicy/chat-epp"]
			Expect(policy).NotTo(BeNil())
			podSelector, _, _ := unstructured.NestedStringMap(policy.Object, "spec", "podSelector", "matchLabels")
			dep := rendered["Deployment/chat-epp"]
			podLabels, _, _ := unstructured.NestedStringMap(dep.Object, "spec", "template", "metadata", "labels")
			Expect(podSelector).To(Equal(podLabels), "the policy must select exactly the EPP Pod")
			policyTypes, _, _ := unstructured.NestedStringSlice(policy.Object, "spec", "policyTypes")
			Expect(policyTypes).To(Equal([]string{"Ingress"}), "egress stays open until an allow-list exists")
			Expect(policy.Object["spec"]).NotTo(HaveKey("egress"))

			rules, _, _ := unstructured.NestedSlice(policy.Object, "spec", "ingress")
			Expect(rules).To(HaveLen(2))
			Expect(rules[0]).To(Equal(map[string]any{
				"from": []any{map[string]any{"namespaceSelector": map[string]any{"matchLabels": map[string]any{namespaceNameLabel: "gateways"}}}},
				"ports": []any{
					map[string]any{"protocol": "TCP", "port": int64(9002)},
					map[string]any{"protocol": "TCP", "port": int64(9003)},
				},
			}))
			Expect(rules[1]).To(Equal(map[string]any{
				"from":  []any{map[string]any{"namespaceSelector": map[string]any{"matchLabels": map[string]any{namespaceNameLabel: "observability"}}}},
				"ports": []any{map[string]any{"protocol": "TCP", "port": int64(9090)}},
			}))
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
		Entry("name too long for the EPP Service", func(in *render.Input) { in.Name = strings.Repeat("a", 60) }),
		Entry("name with a dot", func(in *render.Input) { in.Name = "llama3.1" }),
		Entry("no namespace", func(in *render.Input) { in.Namespace = "" }),
		Entry("no selector", func(in *render.Input) { in.Selector = nil }),
		Entry("no hostnames", func(in *render.Input) { in.Hostnames = nil }),
		Entry("no gateway", func(in *render.Input) { in.GatewayName = "" }),
		Entry("no image", func(in *render.Input) { in.EppImage = "" }),
		Entry("no metrics namespace", func(in *render.Input) { in.MetricsNamespace = "" }),
		Entry("metrics namespace that is not a DNS label", func(in *render.Input) { in.MetricsNamespace = "Monitoring_1" }),
		Entry("zero port", func(in *render.Input) { in.Port = 0 }),
		Entry("scorer without plugin", func(in *render.Input) { in.Scorers = []render.Scorer{{Plugin: "", Weight: 1}} }),
	)

	It("accepts the longest name whose EPP Service is still a DNS label", func() {
		in := llama3Input()
		in.Name = strings.Repeat("a", 59)
		_, err := render.Render(in)
		Expect(err).NotTo(HaveOccurred())
	})

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
