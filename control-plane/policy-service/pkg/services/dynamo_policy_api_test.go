/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/tree"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gvk_compliance"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Dynamo v1beta1 policy API", Serial, func() {
	router := CreateRouter()
	const fixtureDir = "../system_mapping/testing/inputs/nvidia_com_DynamoGraphDeployment_v1beta1"
	gvk := api.Gvk{Group: convutils.Ref("nvidia.com"), Kind: "DynamoGraphDeployment", Version: "v1beta1"}

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()
		tree.Restart()
		backend := NewBackendMockClient(2, 1)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backend))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backend))
		authorization.SetMock(NewPoliciesAuthClient(backend))
		cluster.SetMock(cluster.NewSimpleMockClient().WithClusters(backend.ClusterIdPerTenant))
		assets.SetMock(assets.SimpleMockClientFactory{SimpleMockClient: assets.NewSimpleMockClient()})
	})

	apply := func(rules []api.PolicyRule, workload map[string]any, expectedError ...string) map[string]any {
		created := createGvkPolicyForTest(router, 1, api.CreateGvkPolicyReq{
			WorkloadType: gvk, Scope: api.SCOPETYPE_PROJECT, ProjectId: convutils.Ref("1111"), Rules: rules,
		})
		req, err := http.NewRequest(http.MethodGet, gvkPolicyURL(created.Id)+"?breakdown=true", nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header = *headerForAdmin(1)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req.WithContext(tenantContextForTest(1)))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		var policy policyApi.GetGvkPolicyResp
		decoder := json.NewDecoder(bytes.NewReader(response.Body.Bytes()))
		decoder.UseNumber()
		Expect(decoder.Decode(&policy)).To(Succeed())
		Expect(policy.Breakdown.Get()).NotTo(BeNil())
		var broken []string
		gvk_compliance.BrokenRuleHook = func(_ *policyApi.EffectiveRulesEffectiveRulesInner, message string) { broken = append(broken, message) }
		defer func() { gvk_compliance.BrokenRuleHook = nil }()
		resolved := gvk_compliance.ResolveRules(&policy)
		Expect(broken).To(BeEmpty())
		Expect(resolved).To(HaveLen(len(rules)))
		errs := gvk_compliance.RunRulesWithDependencies(workload, resolved, *policy.Breakdown.Get().SpecSelectors, 20, gvk_compliance.AllCycles)
		if len(expectedError) > 0 {
			Expect(errs).NotTo(BeEmpty())
			Expect(errs[0].Message).To(ContainSubstring(expectedError[0]))
		} else {
			Expect(errs).To(BeEmpty())
		}
		Expect(broken).To(BeEmpty())
		return workload
	}

	for _, caseName := range []string{"workload", "component", "pod", "container", "named-containers", "defaults", "parameterized-dependency", "unknown-component"} {
		It("applies the "+caseName+" policy returned by the API", func() {
			var fixture struct {
				EffectiveRules []api.PolicyRule `json:"effectiveRules"`
			}
			readDynamoAPIFixture(filepath.Join(fixtureDir, "cases", caseName, "policy.yaml"), &fixture)
			var workload, golden map[string]any
			readDynamoAPIFixture(filepath.Join(fixtureDir, "skeleton.yaml"), &workload)
			readDynamoAPIFixture(filepath.Join(fixtureDir, "cases", caseName, "golden.yaml"), &golden)
			actual := apply(fixture.EffectiveRules, workload)
			Expect(normalizeDynamoAPIValue(actual)).To(Equal(normalizeDynamoAPIValue(golden)))
		})
	}

	It("rejects a main-container policy when the pod template is omitted", func() {
		var fixture struct {
			EffectiveRules []api.PolicyRule `json:"effectiveRules"`
		}
		readDynamoAPIFixture(filepath.Join(fixtureDir, "cases", "defaults", "policy.yaml"), &fixture)
		var imageRule api.PolicyRule
		for _, rule := range fixture.EffectiveRules {
			if rule.Field == "image" {
				imageRule = rule
			}
		}
		imageRule.SpecSelector = "allMainContainers"
		workload := map[string]any{"apiVersion": "nvidia.com/v1beta1", "kind": "DynamoGraphDeployment", "metadata": map[string]any{"name": "test"}, "spec": map[string]any{"components": []any{map[string]any{"name": "worker", "type": "worker"}}}}
		actual := apply([]api.PolicyRule{imageRule}, workload, "Dynamo container policies require an explicit main container")
		component := actual["spec"].(map[string]any)["components"].([]any)[0].(map[string]any)
		Expect(component).NotTo(HaveKey("podTemplate"))
	})

	makeRule := func(field, selector string, enforcement map[string]any, params map[string]any) api.PolicyRule {
		data, err := json.Marshal(map[string]any{"field": field, "specSelector": selector, "enforcements": []any{enforcement}, "selectorParams": params})
		Expect(err).NotTo(HaveOccurred())
		var rule api.PolicyRule
		Expect(json.Unmarshal(data, &rule)).To(Succeed())
		return rule
	}
	loadWorkload := func() map[string]any {
		var workload map[string]any
		readDynamoAPIFixture(filepath.Join(fixtureDir, "skeleton.yaml"), &workload)
		return workload
	}
	getComponents := func(workload map[string]any) []any { return workload["spec"].(map[string]any)["components"].([]any) }
	createPolicy := func(rule api.PolicyRule, version ...string) *httptest.ResponseRecorder {
		workloadType := gvk
		if len(version) > 0 {
			workloadType.Version = version[0]
		}
		req, err := http.NewRequest(http.MethodPost, urlGvkPolicies, makePayload(api.CreateGvkPolicyReq{WorkloadType: workloadType, Scope: api.SCOPETYPE_PROJECT, ProjectId: convutils.Ref("1111"), Rules: []api.PolicyRule{rule}}))
		Expect(err).NotTo(HaveOccurred())
		req.Header = *headerForAdmin(1)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req.WithContext(tenantContextForTest(1)))
		return response
	}

	for _, version := range []string{"v1alpha1", "v1beta1"} {
		backends := []string{"vllm", "sglang", "trtllm"}
		if version == "v1alpha1" {
			backends = append(backends, "tensorrt-llm", "nixl", "dynamo")
		}
		for _, backend := range backends {
			for _, enforcement := range []string{"set", "default"} {
				It("accepts "+backend+" in "+version+" "+enforcement+" policies", func() {
					response := createPolicy(makeRule("backendFramework", "workloadSpec", map[string]any{enforcement: map[string]any{"value": backend}}, nil), version)
					Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
				})
			}
		}
	}

	for _, tc := range []struct {
		name, field, selector string
		value                 any
		params                map[string]any
	}{
		{"backend nixl", "backendFramework", "workloadSpec", "nixl", nil},
		{"backend dynamo", "backendFramework", "workloadSpec", "dynamo", nil},
		{"backend tensorrt-llm", "backendFramework", "workloadSpec", "tensorrt-llm", nil},
		{"unknown component type", "type", "allComponentSpecs", "not-a-role", nil},
		{"negative replicas", "replicas", "allComponentSpecs", -1, nil},
		{"replicas exceeding int32", "replicas", "allComponentSpecs", int64(2147483648), nil},
		{"multinode nodeCount one", "multinode-nodeCount", "componentSpec", 1, map[string]any{"componentName": "component-a"}},
		{"fractional replicas", "replicas", "allComponentSpecs", 1.5, nil},
		{"missing componentName", "replicas", "componentSpec", 2, nil},
		{"missing containerName", "image", "appContainerByName", "example/test:latest", nil},
		{"numeric componentName", "replicas", "componentSpec", 2, map[string]any{"componentName": 123}},
		{"missing value", "replicas", "allComponentSpecs", nil, nil},
		{"null component type", "type", "allComponentSpecs", nil, nil},
		{"null backend", "backendFramework", "workloadSpec", nil, nil},
		{"null boolean", "globalDynamoNamespace", "allComponentSpecs", nil, nil},
		{"fractional replica string", "replicas", "allComponentSpecs", "1.5", nil},
		{"negative replica string", "replicas", "allComponentSpecs", "-1", nil},
		{"node count exceeding int32", "multinode-nodeCount", "allComponentSpecs", "2147483648", nil},
		{"numeric boolean", "globalDynamoNamespace", "allComponentSpecs", 0, nil},
		{"uppercase boolean string", "globalDynamoNamespace", "allComponentSpecs", "FALSE", nil},
		{"null componentName", "replicas", "componentSpec", 2, map[string]any{"componentName": nil}},
		{"boolean containerName", "image", "appContainerByName", "image", map[string]any{"containerName": false}},
		{"missing componentName for named container", "image", "componentAppContainerByName", "image", map[string]any{"containerName": "main"}},
	} {
		for _, enforcement := range []string{"set", "default"} {
			It("rejects "+tc.name+" in "+enforcement, func() {
				value := map[string]any{"value": tc.value}
				if tc.name == "missing value" {
					value = map[string]any{}
				}
				response := createPolicy(makeRule(tc.field, tc.selector, map[string]any{enforcement: value}, tc.params))
				Expect(response.Code).To(Equal(http.StatusBadRequest), response.Body.String())
			})
		}
	}

	for _, tc := range []struct {
		field           string
		input, expected any
	}{
		{"replicas", "2", json.Number("2")},
		{"multinode-nodeCount", "2", json.Number("2")},
		{"globalDynamoNamespace", "false", false},
		{"globalDynamoNamespace", "true", true},
		{"replicas", 2147483647, json.Number("2147483647")},
		{"multinode-nodeCount", 2147483647, json.Number("2147483647")},
	} {
		It(fmt.Sprintf("applies %s value %v", tc.field, tc.input), func() {
			workload := loadWorkload()
			rule := makeRule(tc.field, "componentSpec", map[string]any{"set": map[string]any{"value": tc.input}}, map[string]any{"componentName": "component-a"})
			actual := apply([]api.PolicyRule{rule}, workload)
			component := getComponents(actual)[0].(map[string]any)
			value := component[tc.field]
			if tc.field == "multinode-nodeCount" {
				value = component["multinode"].(map[string]any)["nodeCount"]
			}
			Expect(value).To(Equal(tc.expected))
		})
	}

	It("sets replicas to zero through the API", func() {
		actual := apply([]api.PolicyRule{makeRule("replicas", "allComponentSpecs", map[string]any{"set": map[string]any{"value": 0}}, nil)}, loadWorkload())
		for _, component := range getComponents(actual) {
			Expect(component.(map[string]any)["replicas"]).To(Equal(json.Number("0")))
		}
	})
	It("preserves explicit zero replicas under DEFAULT", func() {
		workload := loadWorkload()
		getComponents(workload)[0].(map[string]any)["replicas"] = json.Number("0")
		actual := apply([]api.PolicyRule{makeRule("replicas", "allComponentSpecs", map[string]any{"default": map[string]any{"value": 3}}, nil)}, workload)
		Expect(getComponents(actual)[0].(map[string]any)["replicas"]).To(Equal(json.Number("0")))
	})
	It("preserves false under DEFAULT", func() {
		workload := loadWorkload()
		getComponents(workload)[0].(map[string]any)["globalDynamoNamespace"] = false
		actual := apply([]api.PolicyRule{makeRule("globalDynamoNamespace", "allComponentSpecs", map[string]any{"default": map[string]any{"value": true}}, nil)}, workload)
		Expect(getComponents(actual)[0].(map[string]any)["globalDynamoNamespace"]).To(BeFalse())
	})
	It("accepts false as present for REQUIRED", func() {
		workload := loadWorkload()
		getComponents(workload)[0].(map[string]any)["globalDynamoNamespace"] = false
		apply([]api.PolicyRule{makeRule("globalDynamoNamespace", "componentSpec", map[string]any{"required": true}, map[string]any{"componentName": "component-a"})}, workload)
	})
	It("preserves a zero GPU quantity under DEFAULT", func() {
		workload := loadWorkload()
		pod := getComponents(workload)[0].(map[string]any)["podTemplate"].(map[string]any)["spec"].(map[string]any)
		pod["containers"].([]any)[0].(map[string]any)["resources"] = map[string]any{"limits": map[string]any{"nvidia.com/gpu": "0"}}
		actual := apply([]api.PolicyRule{makeRule("resourceLimits", "allMainContainers", map[string]any{"default": map[string]any{"key": "nvidia.com/gpu", "value": "1"}}, nil)}, workload)
		first := getComponents(actual)[0].(map[string]any)["podTemplate"].(map[string]any)["spec"].(map[string]any)["containers"].([]any)[0].(map[string]any)
		Expect(first["resources"].(map[string]any)["limits"].(map[string]any)["nvidia.com/gpu"]).To(Equal("0"))
	})
	It("preserves an empty annotation value under DEFAULT", func() {
		workload := loadWorkload()
		workload["metadata"].(map[string]any)["annotations"] = map[string]any{"example.com/note": ""}
		actual := apply([]api.PolicyRule{makeRule("annotations", "workloadMetadata", map[string]any{"default": map[string]any{"key": "example.com/note", "value": "fallback"}}, nil)}, workload)
		Expect(actual["metadata"].(map[string]any)["annotations"].(map[string]any)["example.com/note"]).To(Equal(""))
	})
	It("preserves dots and slashes in annotation keys", func() {
		actual := apply([]api.PolicyRule{makeRule("annotations", "workloadMetadata", map[string]any{"set": map[string]any{"key": "example.com/team.name", "value": "test"}}, nil)}, loadWorkload())
		Expect(actual["metadata"].(map[string]any)["annotations"].(map[string]any)).To(HaveKeyWithValue("example.com/team.name", "test"))
	})
	It("initializes metadata for multiple selected components independently", func() {
		workload := loadWorkload()
		rules := []api.PolicyRule{
			makeRule("labels", "componentPodMetadata", map[string]any{"set": map[string]any{"key": "team", "value": "a"}}, map[string]any{"componentName": "component-a"}),
			makeRule("labels", "componentPodMetadata", map[string]any{"set": map[string]any{"key": "team", "value": "c"}}, map[string]any{"componentName": "component-c"}),
		}
		actual := apply(rules, workload)
		components := getComponents(actual)
		for _, i := range []int{0, 2} {
			Expect(components[i].(map[string]any)["podTemplate"].(map[string]any)["metadata"].(map[string]any)["labels"]).To(HaveKey("team"))
		}
		Expect(components[1].(map[string]any)["podTemplate"].(map[string]any)).NotTo(HaveKey("metadata"))
	})
	It("targets by component name after reordering", func() {
		workload := loadWorkload()
		components := getComponents(workload)
		components[0], components[2] = components[2], components[0]
		actual := apply([]api.PolicyRule{makeRule("serviceAccountName", "componentPodSpec", map[string]any{"set": map[string]any{"value": "target-only"}}, map[string]any{"componentName": "component-c"})}, workload)
		rendered := getComponents(actual)
		Expect(rendered[0].(map[string]any)["podTemplate"].(map[string]any)["spec"]).To(HaveKeyWithValue("serviceAccountName", "target-only"))
		Expect(rendered[2].(map[string]any)["podTemplate"].(map[string]any)["spec"]).NotTo(HaveKey("serviceAccountName"))
	})
	It("keeps selector parameters separate from JQ code", func() {
		workload := loadWorkload()
		before, err := json.Marshal(workload)
		Expect(err).NotTo(HaveOccurred())
		actual := apply([]api.PolicyRule{makeRule("replicas", "componentSpec", map[string]any{"set": map[string]any{"value": 8}}, map[string]any{"componentName": "component-a\" | .spec.components[] | #"})}, workload)
		after, err := json.Marshal(actual)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
	})
	It("changes a named sidecar without requiring explicit main", func() {
		workload := loadWorkload()
		pod := getComponents(workload)[0].(map[string]any)["podTemplate"].(map[string]any)["spec"].(map[string]any)
		pod["containers"] = []any{map[string]any{"name": "sidecar", "image": "before"}}
		actual := apply([]api.PolicyRule{makeRule("image", "componentAppContainerByName", map[string]any{"set": map[string]any{"value": "after"}}, map[string]any{"componentName": "component-a", "containerName": "sidecar"})}, workload)
		containers := getComponents(actual)[0].(map[string]any)["podTemplate"].(map[string]any)["spec"].(map[string]any)["containers"].([]any)
		Expect(containers).To(HaveLen(1))
		Expect(containers[0]).To(HaveKeyWithValue("image", "after"))
	})
	It("keeps app and init containers with matching names isolated", func() {
		workload := loadWorkload()
		pod := getComponents(workload)[0].(map[string]any)["podTemplate"].(map[string]any)["spec"].(map[string]any)
		pod["containers"].([]any)[1].(map[string]any)["name"] = "init"
		pod["initContainers"].([]any)[0].(map[string]any)["name"] = "bootstrap"
		actual := apply([]api.PolicyRule{makeRule("image", "appContainerByName", map[string]any{"set": map[string]any{"value": "app-only"}}, map[string]any{"containerName": "init"})}, workload)
		other := getComponents(actual)[1].(map[string]any)["podTemplate"].(map[string]any)["spec"].(map[string]any)
		Expect(other["initContainers"].([]any)[0]).To(HaveKeyWithValue("image", "placeholder"))
		Expect(pod["containers"].([]any)[1]).To(HaveKeyWithValue("image", "app-only"))
	})
	It("reports fields exposed by the live mapping API", func() {
		req, err := http.NewRequest(http.MethodPost, "/api/v1/policies/mappings/minimal", makePayload(map[string]any{"workloadType": gvk}))
		Expect(err).NotTo(HaveOccurred())
		req.Header = *headerForAdmin(1)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req.WithContext(tenantContextForTest(1)))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		var minimal map[string]any
		Expect(json.Unmarshal(response.Body.Bytes(), &minimal)).To(Succeed())
		fields := minimal["fields"].(map[string]any)
		for _, field := range []string{"replicas", "type", "image", "resourceLimits", "labels", "annotations"} {
			Expect(fields).To(HaveKey(field))
		}
	})

})

func readDynamoAPIFixture(path string, out any) {
	raw, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	raw, err = yaml.YAMLToJSON(raw)
	Expect(err).NotTo(HaveOccurred())
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	Expect(decoder.Decode(out)).To(Succeed())
}

func normalizeDynamoAPIValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		result := map[string]any{}
		for key, child := range value {
			if child != nil {
				result[key] = normalizeDynamoAPIValue(child)
			}
		}
		return result
	case []any:
		result := make([]any, len(value))
		for i, child := range value {
			result[i] = normalizeDynamoAPIValue(child)
		}
		return result
	default:
		return value
	}
}
