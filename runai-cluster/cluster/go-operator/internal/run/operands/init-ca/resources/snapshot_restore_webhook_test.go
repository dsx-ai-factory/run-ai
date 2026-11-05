/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

func snapshotTestConfig() *runv1.RunaiConfig {
	runaiConfig := &runv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "runai", Name: "runai"},
	}
	runaiConfig.SetDefaultsWhereNeeded()
	return runaiConfig
}

// restoreWebhook returns the restore configuration, or nil when it was not
// registered. It decodes rather than greps, because the acceptance criterion
// rests on fields surviving the YAML round-trip.
func restoreWebhook(t *testing.T, snapshotInstalled bool) *admissionv1.MutatingWebhookConfiguration {
	t.Helper()

	for _, doc := range strings.Split(webhooksFile(snapshotTestConfig(), snapshotInstalled)["webhooks.yaml"], "---") {
		if !strings.Contains(doc, snapshotRestoreWebhookConfigurationName) {
			continue
		}
		configuration := &admissionv1.MutatingWebhookConfiguration{}
		if err := yaml.Unmarshal([]byte(doc), configuration); err != nil {
			t.Fatalf("decode webhook configuration: %v", err)
		}
		if configuration.Name == snapshotRestoreWebhookConfigurationName {
			return configuration
		}
	}
	return nil
}

// A cluster without Snapshot must carry no configuration at all, so an opt-in
// label routes nothing.
func TestSnapshotRestoreWebhookIsRegisteredOnlyWhenSnapshotIsInstalled(t *testing.T) {
	if got := restoreWebhook(t, false); got != nil {
		t.Fatal("restore webhook registered on a cluster without Snapshot")
	}
	if got := restoreWebhook(t, true); got == nil {
		t.Fatal("restore webhook not registered although Snapshot is installed")
	}
}

// Sorting first would drop runai-pod's additions from the key, and captures
// would silently stop matching their replicas. Asserted so a rename fails loudly.
func TestSnapshotRestoreWebhookSortsAfterRunaiPod(t *testing.T) {
	if snapshotRestoreWebhookConfigurationName <= "runai-pod" {
		t.Fatalf("%q must sort after %q", snapshotRestoreWebhookConfigurationName, "runai-pod")
	}
}

func TestSnapshotRestoreWebhookShape(t *testing.T) {
	configuration := restoreWebhook(t, true)
	if configuration == nil {
		t.Fatal("restore webhook not registered")
	}

	// init-ca garbage-collects by this label; without it the configuration
	// outlives an uninstall.
	if configuration.Labels["app"] != "runai" {
		t.Errorf("labels = %v, want app=runai so init-ca can garbage-collect it", configuration.Labels)
	}

	if len(configuration.Webhooks) != 1 {
		t.Fatalf("webhooks = %d, want 1", len(configuration.Webhooks))
	}
	hook := configuration.Webhooks[0]

	// An outage must mean cold starts, not refused pods.
	if hook.FailurePolicy == nil || *hook.FailurePolicy != admissionv1.Ignore {
		t.Errorf("failurePolicy = %v, want Ignore", hook.FailurePolicy)
	}

	// The acceptance criterion.
	if hook.ObjectSelector == nil {
		t.Fatal("objectSelector is nil, so every pod in a project namespace would be sent to the webhook")
	}
	if got := hook.ObjectSelector.MatchLabels[constants.SnapshotOptInLabel]; got != "true" {
		t.Errorf("objectSelector[%s] = %q, want %q", constants.SnapshotOptInLabel, got, "true")
	}

	if hook.NamespaceSelector == nil || len(hook.NamespaceSelector.MatchExpressions) != 1 ||
		hook.NamespaceSelector.MatchExpressions[0].Key != "runai/queue" {
		t.Errorf("namespaceSelector = %v, want runai/queue Exists", hook.NamespaceSelector)
	}

	// Reinvocation would void the ordering the name buys.
	if hook.ReinvocationPolicy != nil && *hook.ReinvocationPolicy != admissionv1.NeverReinvocationPolicy {
		t.Errorf("reinvocationPolicy = %v, want unset or Never", *hook.ReinvocationPolicy)
	}

	if len(hook.Rules) != 1 {
		t.Fatalf("rules = %d, want 1", len(hook.Rules))
	}
	rule := hook.Rules[0]
	if len(rule.Operations) != 1 || rule.Operations[0] != admissionv1.Create {
		t.Errorf("operations = %v, want [CREATE] only", rule.Operations)
	}
	if len(rule.Resources) != 1 || rule.Resources[0] != "pods" {
		t.Errorf("resources = %v, want [pods]", rule.Resources)
	}

	if hook.ClientConfig.Service == nil {
		t.Fatal("clientConfig.Service is nil")
	}
	if got := hook.ClientConfig.Service.Name; got != "anyworkload-controller" {
		t.Errorf("service = %q, want anyworkload-controller", got)
	}
	if got := ptr.Deref(hook.ClientConfig.Service.Path, ""); got != snapshotRestoreWebhookPath {
		t.Errorf("path = %q, want %q", got, snapshotRestoreWebhookPath)
	}
}

// On OpenShift service-CA also manages the anyworkload-controller serving
// secret. Without this annotation the caBundle goes stale on re-issue and the
// handshake fails — which failurePolicy: Ignore then hides.
func TestSnapshotRestoreWebhookCarriesOpenShiftCABundleAnnotation(t *testing.T) {
	runaiConfig := snapshotTestConfig()
	runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = runv1internal.KubernetesFlavorOpenShift

	var found *admissionv1.MutatingWebhookConfiguration
	for _, doc := range strings.Split(webhooksFile(runaiConfig, true)["webhooks.yaml"], "---") {
		if !strings.Contains(doc, snapshotRestoreWebhookConfigurationName) {
			continue
		}
		configuration := &admissionv1.MutatingWebhookConfiguration{}
		if err := yaml.Unmarshal([]byte(doc), configuration); err != nil {
			t.Fatalf("decode webhook configuration: %v", err)
		}
		if configuration.Name == snapshotRestoreWebhookConfigurationName {
			found = configuration
		}
	}
	if found == nil {
		t.Fatal("restore webhook not registered on OpenShift")
	}
	if got := found.Annotations["service.beta.openshift.io/inject-cabundle"]; got != "true" {
		t.Errorf("inject-cabundle = %q, want %q", got, "true")
	}
}
