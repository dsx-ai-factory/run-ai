/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

func TestWorkloadRunnerOwnershipRulesAreFeatureGated(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "runai", Name: "runai"}}
	runaiConfig.SetDefaultsWhereNeeded()

	assertWorkloadRunnerOwnershipRules(t, webhooksFile(runaiConfig)["webhooks.yaml"], false)
	runaiConfig.Spec.Global.EnableWorkloadOwnershipProtection = ptr.To(true)
	assertWorkloadRunnerOwnershipRules(t, webhooksFile(runaiConfig)["webhooks.yaml"], true)
}

func TestLegacyOwnershipRulesTargetAnyWorkloadController(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "runai", Name: "runai"}}
	runaiConfig.SetDefaultsWhereNeeded()
	runaiConfig.Spec.Global.EnableWorkloadOwnershipProtection = ptr.To(true)

	expected := map[string]struct {
		path       string
		operations []any
		resources  [][]any
	}{
		"runai-workload-ownership-protection-validation": {
			path: "/validate-workload-ownership-protection", operations: []any{"UPDATE", "DELETE"},
			resources: [][]any{{"trainingworkloads", "interactiveworkloads", "distributedworkloads"}},
		},
		"runai-inference-workload-ownership-protection-validation": {
			path: "/validate-workload-ownership-protection", operations: []any{"UPDATE", "DELETE"},
			resources: [][]any{{"inferenceworkloads"}, {"distributedinferenceworkloads"}},
		},
		"runai-workload-ownership-mutation": {
			path: "/mutate-workload-ownership", operations: []any{"UPDATE"},
			resources: [][]any{{"trainingworkloads", "interactiveworkloads", "distributedworkloads"}},
		},
		"runai-inference-workload-ownership-mutation": {
			path: "/mutate-workload-ownership", operations: []any{"UPDATE"},
			resources: [][]any{{"inferenceworkloads"}, {"distributedinferenceworkloads"}},
		},
	}

	for _, object := range decodeWebhookObjects(t, webhooksFile(runaiConfig)["webhooks.yaml"]) {
		want, tracked := expected[object.GetName()]
		if !tracked {
			continue
		}
		webhooks, found, err := unstructured.NestedSlice(object.Object, "webhooks")
		if err != nil || !found || len(webhooks) != 1 {
			t.Fatalf("%s webhooks have an invalid shape: found=%t count=%d err=%v", object.GetName(), found, len(webhooks), err)
		}
		webhook := webhooks[0].(map[string]any)
		serviceName, _, _ := unstructured.NestedString(webhook, "clientConfig", "service", "name")
		path, _, _ := unstructured.NestedString(webhook, "clientConfig", "service", "path")
		if serviceName != "anyworkload-controller" || path != want.path {
			t.Fatalf("%s targets service=%q path=%q, want anyworkload-controller %q", object.GetName(), serviceName, path, want.path)
		}
		rules, _, _ := unstructured.NestedSlice(webhook, "rules")
		if len(rules) != len(want.resources) {
			t.Fatalf("%s has %d rules, want %d", object.GetName(), len(rules), len(want.resources))
		}
		for index, ruleValue := range rules {
			rule := ruleValue.(map[string]any)
			operations, _, _ := unstructured.NestedSlice(rule, "operations")
			resources, _, _ := unstructured.NestedSlice(rule, "resources")
			if !reflect.DeepEqual(operations, want.operations) || !reflect.DeepEqual(resources, want.resources[index]) {
				t.Fatalf("%s rule %d operations/resources = %v/%v, want %v/%v",
					object.GetName(), index, operations, resources, want.operations, want.resources[index])
			}
		}
		delete(expected, object.GetName())
	}
	if len(expected) != 0 {
		t.Fatalf("missing legacy ownership configurations: %v", reflect.ValueOf(expected).MapKeys())
	}
}

func TestOpenShiftOwnershipWebhooksRequestServiceCAInjection(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "runai", Name: "runai"}}
	runaiConfig.SetDefaultsWhereNeeded()
	runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = runv1internal.KubernetesFlavorOpenShift
	runaiConfig.Spec.Global.EnableWorkloadOwnershipProtection = ptr.To(true)

	want := map[string]bool{
		"runai-anyworkload-validation":                             true,
		"runai-anyworkload-mutation":                               true,
		"runai-workload-ownership-protection-validation":           true,
		"runai-inference-workload-ownership-protection-validation": true,
		"runai-workload-ownership-mutation":                        true,
		"runai-inference-workload-ownership-mutation":              true,
	}
	for _, object := range decodeWebhookObjects(t, webhooksFile(runaiConfig)["webhooks.yaml"]) {
		if !want[object.GetName()] {
			continue
		}
		if object.GetAnnotations()["service.beta.openshift.io/inject-cabundle"] != "true" {
			t.Errorf("%s does not request OpenShift service CA injection", object.GetName())
		}
		delete(want, object.GetName())
	}
	if len(want) != 0 {
		t.Fatalf("missing ownership webhook configurations: %v", reflect.ValueOf(want).MapKeys())
	}
}

func assertWorkloadRunnerOwnershipRules(t *testing.T, content string, enabled bool) {
	t.Helper()
	for _, object := range decodeWebhookObjects(t, content) {
		if object.GetName() != "runai-anyworkload-validation" && object.GetName() != "runai-anyworkload-mutation" {
			continue
		}
		webhooks, found, err := unstructured.NestedSlice(object.Object, "webhooks")
		if err != nil || !found {
			t.Fatalf("%s webhooks has an invalid shape: found=%t err=%v", object.GetName(), found, err)
		}
		wantCount := 1
		if enabled {
			wantCount = 2
		}
		if len(webhooks) != wantCount {
			t.Fatalf("%s webhook count = %d, want %d", object.GetName(), len(webhooks), wantCount)
		}
		if !enabled {
			continue
		}
		ownershipWebhook, ok := webhooks[1].(map[string]any)
		if !ok {
			t.Fatalf("%s ownership webhook has type %T, want map[string]any", object.GetName(), webhooks[1])
		}
		path, found, err := unstructured.NestedString(ownershipWebhook, "clientConfig", "service", "path")
		if err != nil || !found {
			t.Fatalf("%s ownership service path has an invalid shape: found=%t err=%v", object.GetName(), found, err)
		}
		if path != "/validate-workload-ownership-protection" && path != "/mutate-workload-ownership" {
			t.Fatalf("unexpected ownership path %v", path)
		}
		rules, found, err := unstructured.NestedSlice(ownershipWebhook, "rules")
		if err != nil || !found || len(rules) == 0 {
			t.Fatalf("%s ownership rules have an invalid shape: found=%t count=%d err=%v", object.GetName(), found, len(rules), err)
		}
		firstRule, ok := rules[0].(map[string]any)
		if !ok {
			t.Fatalf("%s first ownership rule has type %T, want map[string]any", object.GetName(), rules[0])
		}
		operations, found, err := unstructured.NestedSlice(firstRule, "operations")
		if err != nil || !found {
			t.Fatalf("%s ownership operations have an invalid shape: found=%t err=%v", object.GetName(), found, err)
		}
		if object.GetName() == "runai-anyworkload-validation" && !containsOperations(operations, "CREATE", "UPDATE", "DELETE") {
			t.Fatalf("validation ownership operations = %v", operations)
		}
		if object.GetName() == "runai-anyworkload-mutation" && !containsOperations(operations, "UPDATE") {
			t.Fatalf("mutation ownership operations = %v", operations)
		}
	}
}

func containsOperations(actual []any, expected ...string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for index := range expected {
		if actual[index] != expected[index] {
			return false
		}
	}
	return true
}

func decodeWebhookObjects(t *testing.T, content string) []*unstructured.Unstructured {
	t.Helper()
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewBufferString(content), 4096)
	var objects []*unstructured.Unstructured
	for {
		object := &unstructured.Unstructured{}
		if err := decoder.Decode(object); err == io.EOF {
			return objects
		} else if err != nil {
			t.Fatalf("decode generated webhook YAML: %v", err)
		}
		if object.GetName() != "" {
			objects = append(objects, object)
		}
	}
}

const baseYAMLRelPath = "../../../../../../runai-chart/templates/clusterroles/base.yaml"

func TestWebhookNamesInClusterRoleResourceNames(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "runai",
			Name:      "runai",
		},
	}
	runaiConfig.SetDefaultsWhereNeeded()

	runaiConfig.Spec.AdviceRestartController.Enabled = ptr.To(true)
	runaiConfig.Spec.Global.EnableWorkloadOwnershipProtection = ptr.To(true)
	runaiConfig.Spec.PodGroupAssigner.Enabled = ptr.To(true)
	runaiConfig.Spec.Internal.Development.Headless = false

	data := webhooksFile(runaiConfig)
	webhooksYAML := data["webhooks.yaml"]

	webhookNames := extractWebhookNames(t, webhooksYAML)
	if len(webhookNames) == 0 {
		t.Fatal("no webhook names found in generated webhooks.yaml")
	}

	baseYAML, err := os.ReadFile(baseYAMLRelPath)
	if err != nil {
		t.Fatalf("failed to read base.yaml: %v", err)
	}

	clusterRoles := []string{"runai-init-ca"}
	for _, roleName := range clusterRoles {
		resourceNames := extractAdmissionResourceNames(t, string(baseYAML), roleName)
		if len(resourceNames) == 0 {
			t.Fatalf("no resourceNames found for ClusterRole %q", roleName)
		}
		nameSet := make(map[string]bool, len(resourceNames))
		for _, n := range resourceNames {
			nameSet[n] = true
		}
		for _, wh := range webhookNames {
			if !nameSet[wh] {
				t.Errorf("webhook %q (from init-ca ConfigMap) is missing from ClusterRole %q resourceNames in base.yaml", wh, roleName)
			}
		}
	}
}

func TestLegacyWebhookNamesInClusterRoleResourceNames(t *testing.T) {
	baseYAML, err := os.ReadFile(baseYAMLRelPath)
	if err != nil {
		t.Fatalf("failed to read base.yaml: %v", err)
	}

	resourceNames := extractAdmissionResourceNames(t, string(baseYAML), "runai-init-ca")
	nameSet := make(map[string]bool, len(resourceNames))
	for _, name := range resourceNames {
		nameSet[name] = true
	}

	for _, legacyName := range []string{
		"assets-sync",
		"runai-assets-sync-validation",
		"runai-interactive-policy-validation",
		"runai-interactive-workload-validation",
		"runai-interactive-workload-mutation",
		"runai-training-policy-validation",
		"runai-training-workload-validation",
		"runai-training-workload-mutation",
		"runai-distributed-policy-validation",
		"runai-distributed-workload-validation",
		"runai-distributed-workload-mutation",
		"runai-inference-policy-validation",
		"runai-inference-policy-validation-old",
		"runai-inference-workload-validation",
		"runai-inference-workload-validation-old",
		"runai-inference-workload-mutation",
		"runai-inference-workload-mutation-old",
		"runai-distributed-inference-workload-validation",
		"runai-distributed-inference-workload-mutation",
		"runai-workload-ownership-protection-validation",
		"runai-workload-ownership-mutation",
		"runai-inference-workload-ownership-protection-validation",
		"runai-inference-workload-ownership-mutation",
		"runai-restartrequest-validation",
		"runai-advice-restart-dw-mutation",
	} {
		if !nameSet[legacyName] {
			t.Errorf("legacy webhook %q is missing from ClusterRole resourceNames; init-ca needs it to clean up upgrades", legacyName)
		}
	}
}

func TestWebhooksFileContainsOnlySurvivingConfigurations(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "runai", Name: "runai"},
	}
	runaiConfig.SetDefaultsWhereNeeded()
	runaiConfig.Spec.AdviceRestartController.Enabled = ptr.To(true)
	runaiConfig.Spec.Global.EnableWorkloadOwnershipProtection = ptr.To(true)

	webhookNames := extractWebhookNames(t, webhooksFile(runaiConfig)["webhooks.yaml"])
	sort.Strings(webhookNames)
	expected := []string{
		"application-validation",
		"runai-accessrule-validation",
		"runai-anyworkload-mutation",
		"runai-anyworkload-validation",
		"runai-inference-workload-ownership-mutation",
		"runai-inference-workload-ownership-protection-validation",
		"runai-pod",
		"runai-pod-controller-mutation",
		"runai-project-validation",
		"runai-restartrequest-validation",
		"runai-secret-pvc-validation",
		"runai-workload-ownership-mutation",
		"runai-workload-ownership-protection-validation",
	}
	sort.Strings(expected)

	if !reflect.DeepEqual(webhookNames, expected) {
		t.Fatalf("generated webhook names = %v, want %v", webhookNames, expected)
	}
}

func TestPodControllerMutationExcludesRunaiJobs(t *testing.T) {
	runaiConfig := &runv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "runai", Name: "runai"},
	}
	runaiConfig.SetDefaultsWhereNeeded()

	docs := strings.Split(webhooksFile(runaiConfig)["webhooks.yaml"], "---")
	for _, doc := range docs {
		if strings.TrimSpace(doc) == "" {
			continue
		}
		var webhook admissionv1.MutatingWebhookConfiguration
		if err := yaml.Unmarshal([]byte(doc), &webhook); err != nil {
			t.Fatalf("failed to decode webhook document: %v", err)
		}
		if webhook.Kind != "MutatingWebhookConfiguration" || webhook.Name != "runai-pod-controller-mutation" {
			continue
		}
		if len(webhook.Webhooks) == 0 || len(webhook.Webhooks[0].Rules) == 0 {
			t.Fatal("runai-pod-controller-mutation has no webhook rules")
		}

		got := webhook.Webhooks[0].Rules[0].Resources
		want := []string{"statefulsets", "jobs", "deployments"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("pod-controller mutation resources = %v, want %v", got, want)
		}
		return
	}

	t.Fatal("runai-pod-controller-mutation was not generated")
}

type webhookMeta struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
}

func extractWebhookNames(t *testing.T, yamlContent string) []string {
	t.Helper()
	docs := strings.Split(yamlContent, "---")
	var names []string
	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}
		var meta webhookMeta
		if err := yaml.Unmarshal([]byte(doc), &meta); err != nil {
			t.Fatalf("failed to unmarshal webhook document: %v\ndoc:\n%s", err, doc)
		}
		if meta.Metadata.Name != "" {
			names = append(names, meta.Metadata.Name)
		}
	}
	return names
}

// extractAdmissionResourceNames finds the ClusterRole section for the given name in
// base.yaml and extracts the resourceNames list from the admissionregistration.k8s.io
// rule that restricts update/patch/delete.
func extractAdmissionResourceNames(t *testing.T, fileContent, clusterRoleName string) []string {
	t.Helper()

	sections := strings.Split(fileContent, "---")
	for _, section := range sections {
		if !strings.Contains(section, "kind: ClusterRole") {
			continue
		}
		if !strings.Contains(section, "name: "+clusterRoleName) {
			continue
		}
		if !strings.Contains(section, "admissionregistration.k8s.io") {
			continue
		}

		re := regexp.MustCompile(`resourceNames:\s*(\[.*\])`)
		matches := re.FindAllStringSubmatch(section, -1)
		for _, match := range matches {
			jsonArray := match[1]
			var names []string
			if err := json.Unmarshal([]byte(jsonArray), &names); err != nil {
				t.Fatalf("failed to parse resourceNames JSON for ClusterRole %q: %v\nraw: %s", clusterRoleName, err, jsonArray)
			}
			if containsAdmissionWebhookName(names) {
				return names
			}
		}
	}

	return nil
}

func containsAdmissionWebhookName(names []string) bool {
	for _, n := range names {
		if strings.Contains(n, "runai-pod") || strings.Contains(n, "validation") || strings.Contains(n, "mutation") {
			return true
		}
	}
	return false
}
