/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	controlplanedetailsdownloader "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/control-plane-details-downloader"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	advice_restart_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/advice-restart-controller"
	inference_workload_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/inference-workload-controller"
	job_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/job-controller"
	runaiv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	workload_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestApplyControllerTransitionDefaultsFreshInstall(t *testing.T) {
	config := &runaiv1.RunaiConfig{Spec: runaiv1.RunaiConfigSpec{
		WorkloadController:          &workload_controller.WorkloadController{Enabled: ptr.To(true)},
		InferenceWorkloadController: &inference_workload_controller.InferenceWorkloadController{Enabled: ptr.To(true)},
		JobController:               &job_controller.JobController{Enabled: ptr.To(true)},
	}}

	if err := applyControllerTransitionDefaults(config, controllerTransitionFreshInstall, ""); err != nil {
		t.Fatalf("applyControllerTransitionDefaults() error = %v", err)
	}

	assertControllerEnabled(t, "workload controller", config.Spec.WorkloadController.Enabled, false)
	assertControllerEnabled(t, "inference workload controller", config.Spec.InferenceWorkloadController.Enabled, false)
	assertControllerEnabled(t, "job controller", config.Spec.JobController.Enabled, false)
	assertLegacyProjectPermissions(t, config, false)
	if len(config.Annotations) != 0 {
		t.Fatalf("fresh install annotations = %#v, want no cleanup request", config.Annotations)
	}
	assertNoIndependentOperatorEnablement(t, config)
}

func TestApplyControllerTransitionDefaultsUpgradeDoesNotDependOnExistingRunaiConfig(t *testing.T) {
	config := disabledControllerConfig()

	if err := applyControllerTransitionDefaults(config, controllerTransitionUpgrade, "2.26.0"); err != nil {
		t.Fatalf("applyControllerTransitionDefaults() error = %v", err)
	}

	assertControllerEnabled(t, "workload controller", config.Spec.WorkloadController.Enabled, true)
	assertControllerEnabled(t, "inference workload controller", config.Spec.InferenceWorkloadController.Enabled, true)
	assertControllerEnabled(t, "job controller", config.Spec.JobController.Enabled, true)
	if got := config.Annotations[legacyCleanupSourceVersionAnnotation]; got != "2.26.0" {
		t.Fatalf("cleanup source annotation = %q, want 2.26.0", got)
	}
}

func TestApplyControllerTransitionDefaultsUpgrade(t *testing.T) {
	tests := []struct {
		name            string
		previousVersion string
		wantLegacy      bool
	}{
		{
			name:            "pre-3.0 upgrade keeps legacy controllers enabled",
			previousVersion: "2.26.0",
			wantLegacy:      true,
		},
		{
			name:            "3.x upgrade preserves disabled legacy controllers",
			previousVersion: "3.0.0",
			wantLegacy:      false,
		},
		{
			name:            "3.0 prerelease rollout preserves disabled legacy controllers",
			previousVersion: "3.0.0-rc.1",
			wantLegacy:      false,
		},
		{
			name:            "branch build rollout preserves disabled legacy controllers",
			previousVersion: "0.0.0-12345.v300-main",
			wantLegacy:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := disabledControllerConfig()
			if err := applyControllerTransitionDefaults(config, controllerTransitionUpgrade, tt.previousVersion); err != nil {
				t.Fatalf("applyControllerTransitionDefaults() error = %v", err)
			}

			assertControllerEnabled(t, "workload controller", config.Spec.WorkloadController.Enabled, tt.wantLegacy)
			assertControllerEnabled(t, "inference workload controller", config.Spec.InferenceWorkloadController.Enabled, tt.wantLegacy)
			assertControllerEnabled(t, "job controller", config.Spec.JobController.Enabled, tt.wantLegacy)
			if tt.wantLegacy && config.Annotations[legacyCleanupSourceVersionAnnotation] != tt.previousVersion {
				t.Fatalf("cleanup source annotation = %q, want %q",
					config.Annotations[legacyCleanupSourceVersionAnnotation], tt.previousVersion)
			}
		})
	}
}

func TestApplyControllerTransitionDefaultsStopsLegacyRestartRequestProduction(t *testing.T) {
	config := disabledControllerConfig()
	config.Spec.AdviceRestartController = &advice_restart_controller.AdviceRestartController{Enabled: ptr.To(true)}

	if err := applyControllerTransitionDefaults(config, controllerTransitionUpgrade, "2.26.0"); err != nil {
		t.Fatalf("applyControllerTransitionDefaults() error = %v", err)
	}
	if config.Spec.AdviceRestartController.Enabled == nil || *config.Spec.AdviceRestartController.Enabled {
		t.Fatal("advice-restart-controller must be disabled before legacy cleanup")
	}
}

func TestApplyControllerTransitionDefaultsInvalidVersionDoesNotMutate(t *testing.T) {
	config := disabledControllerConfig()
	before := config.DeepCopy()

	if err := applyControllerTransitionDefaults(config, controllerTransitionUpgrade, "not-a-version"); err == nil {
		t.Fatal("applyControllerTransitionDefaults() error = nil, want version parsing error")
	}
	if !reflect.DeepEqual(before.Spec, config.Spec) {
		t.Fatalf("config mutated on error: before=%#v after=%#v", before.Spec, config.Spec)
	}
}

func TestApplyControllerTransitionDefaultsUpgradeRequiresPreviousVersion(t *testing.T) {
	config := enabledLegacyControllerConfig()

	if err := applyControllerTransitionDefaults(config, controllerTransitionUpgrade, ""); err == nil {
		t.Fatal("applyControllerTransitionDefaults() error = nil, want missing previous version error")
	}
}

func TestCommitRunaiConfigAppliesTransitionAfterClusterConfigPatch(t *testing.T) {
	t.Run("fresh install force-disables only legacy fields", func(t *testing.T) {
		k8s := newRunaiConfigFakeClient(t)
		config := namedRunaiConfig()
		params := &hooks.ClusterHookParams{
			ClusterDetails: hooks.ClusterDetails{Version: "0.0.0-30002277.master-8edbfb4"},
			ClusterConfig: []byte(`{
				"workload-controller":{"enabled":true},
				"inference-workload-controller":{"enabled":true},
				"job-controller":{"enabled":true}
			}`),
		}

		if err := commitRunaiconfigToCluster(context.Background(), k8s, config, params, nil,
			false, "", controllerTransitionFreshInstall); err != nil {
			t.Fatalf("commitRunaiconfigToCluster() error = %v", err)
		}

		got := getRunaiConfig(t, k8s)
		assertControllerEnabled(t, "workload controller", got.Spec.WorkloadController.Enabled, false)
		assertControllerEnabled(t, "inference workload controller", got.Spec.InferenceWorkloadController.Enabled, false)
		assertControllerEnabled(t, "job controller", got.Spec.JobController.Enabled, false)
		assertLegacyProjectPermissions(t, got, false)
		assertNoIndependentOperatorEnablement(t, got)
	})

	t.Run("pre-3.0 upgrade force-enables legacy fields", func(t *testing.T) {
		k8s := newRunaiConfigFakeClient(t, namedRunaiConfig())
		desired := getRunaiConfig(t, k8s)
		params := &hooks.ClusterHookParams{
			ClusterDetails: hooks.ClusterDetails{Version: "3.0.0"},
			ClusterConfig: []byte(`{
				"workload-controller":{"enabled":false},
				"inference-workload-controller":{"enabled":false},
				"job-controller":{"enabled":false}
			}`),
		}

		if err := commitRunaiconfigToCluster(context.Background(), k8s, desired, params, headlessInstallationConfiguration(),
			true, "2.26.0", controllerTransitionUpgrade); err != nil {
			t.Fatalf("commitRunaiconfigToCluster() error = %v", err)
		}

		got := getRunaiConfig(t, k8s)
		assertControllerEnabled(t, "workload controller", got.Spec.WorkloadController.Enabled, true)
		assertControllerEnabled(t, "inference workload controller", got.Spec.InferenceWorkloadController.Enabled, true)
		assertControllerEnabled(t, "job controller", got.Spec.JobController.Enabled, true)
	})
}

func TestCommitRunaiConfigUpdateFailureLeavesPersistedConfigUnchanged(t *testing.T) {
	existing := namedRunaiConfig()
	setLegacyControllersEnabled(existing, true)
	failingUpdate := interceptor.Funcs{
		Update: func(context.Context, client.WithWatch, client.Object, ...client.UpdateOption) error {
			return errors.New("update failed")
		},
	}
	k8s := newRunaiConfigFakeClientWithInterceptor(t, failingUpdate, existing)
	desired := getRunaiConfig(t, k8s)
	before := desired.DeepCopy()
	params := &hooks.ClusterHookParams{
		ClusterDetails: hooks.ClusterDetails{Version: "3.0.0"},
		ClusterConfig:  []byte(`{"workload-controller":{"enabled":false}}`),
	}

	if err := commitRunaiconfigToCluster(context.Background(), k8s, desired, params, headlessInstallationConfiguration(),
		true, "2.26.0", controllerTransitionUpgrade); err == nil {
		t.Fatal("commitRunaiconfigToCluster() error = nil, want update failure")
	}

	got := getRunaiConfig(t, k8s)
	if !reflect.DeepEqual(before.Spec, got.Spec) {
		t.Fatalf("persisted config changed on failed update: before=%#v after=%#v", before.Spec, got.Spec)
	}
}

func disabledControllerConfig() *runaiv1.RunaiConfig {
	return &runaiv1.RunaiConfig{Spec: runaiv1.RunaiConfigSpec{
		WorkloadController:          &workload_controller.WorkloadController{Enabled: ptr.To(false)},
		InferenceWorkloadController: &inference_workload_controller.InferenceWorkloadController{Enabled: ptr.To(false)},
		JobController:               &job_controller.JobController{Enabled: ptr.To(false)},
	}}
}

func enabledLegacyControllerConfig() *runaiv1.RunaiConfig {
	config := disabledControllerConfig()
	setLegacyControllersEnabled(config, true)
	return config
}

func namedRunaiConfig() *runaiv1.RunaiConfig {
	return &runaiv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{
		Name:      runaiv1.DefaultName,
		Namespace: runaiv1.DefaultNamespace,
	}}
}

func headlessInstallationConfiguration() *controlplanedetailsdownloader.InstallationConfigurationData {
	return &controlplanedetailsdownloader.InstallationConfigurationData{Internal: runaiv1_internal.Internal{
		Development: &runaiv1_internal.Development{Headless: true},
	}}
}

func newRunaiConfigFakeClient(t *testing.T, objects ...client.Object) client.Client {
	t.Helper()
	return newRunaiConfigFakeClientWithInterceptor(t, interceptor.Funcs{}, objects...)
}

func newRunaiConfigFakeClientWithInterceptor(t *testing.T, funcs interceptor.Funcs, objects ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := runaiv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add RunaiConfig to scheme: %v", err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).WithInterceptorFuncs(funcs).Build()
}

func getRunaiConfig(t *testing.T, k8s client.Client) *runaiv1.RunaiConfig {
	t.Helper()
	config := &runaiv1.RunaiConfig{}
	key := types.NamespacedName{Name: runaiv1.DefaultName, Namespace: runaiv1.DefaultNamespace}
	if err := k8s.Get(context.Background(), key, config); err != nil {
		t.Fatalf("get RunaiConfig: %v", err)
	}
	return config
}

func assertControllerEnabled(t *testing.T, name string, got *bool, want bool) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s enabled = nil, want %t", name, want)
	}
	if *got != want {
		t.Errorf("%s enabled = %t, want %t", name, *got, want)
	}
}

func assertLegacyProjectPermissions(t *testing.T, config *runaiv1.RunaiConfig, want bool) {
	t.Helper()
	if config.Spec.AccessRuleController == nil || config.Spec.AccessRuleController.Args == nil ||
		config.Spec.AccessRuleController.Args.LegacyProjectPermissions == nil {
		t.Fatalf("legacyProjectPermissions = nil, want %t", want)
	}
	if got := *config.Spec.AccessRuleController.Args.LegacyProjectPermissions; got != want {
		t.Errorf("legacyProjectPermissions = %t, want %t", got, want)
	}
}

func assertNoIndependentOperatorEnablement(t *testing.T, config *runaiv1.RunaiConfig) {
	t.Helper()
	serialized, err := json.Marshal(config.Spec)
	if err != nil {
		t.Fatalf("marshal RunaiConfig spec: %v", err)
	}
	var spec map[string]map[string]any
	if err := json.Unmarshal(serialized, &spec); err != nil {
		t.Fatalf("unmarshal RunaiConfig spec: %v", err)
	}
	if mpi, found := spec["mpi-operator"]; found {
		if _, enabled := mpi["enabled"]; enabled {
			t.Fatal("transition persisted MPI operator enablement")
		}
	}
	if _, found := spec["training-operator"]; found {
		t.Fatal("transition persisted training operator configuration")
	}
}
