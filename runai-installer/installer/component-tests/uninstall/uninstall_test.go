package uninstall_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/uninstall"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUninstallBaseCfg() *config.UninstallConfig {
	return &config.UninstallConfig{
		ExecutionTimeout: 5 * time.Minute,
		StepTimeout:      2 * time.Minute,
	}
}

func assertNamespaceGone(t *testing.T, k8s *harness.FakeK8sWrapper, namespace, msg string) {
	t.Helper()
	exists, err := k8s.NamespaceExists(namespace)
	require.NoError(t, err)
	assert.False(t, exists, msg)
}

func hasRunaiCRDs(t *testing.T, k8s *harness.FakeK8sWrapper) bool {
	t.Helper()
	crds, err := k8s.ListCRDs()
	require.NoError(t, err)
	for _, crd := range crds {
		if strings.HasSuffix(strings.ToLower(crd), ".run.ai") {
			return true
		}
	}
	return false
}

func TestUninstall_NoReleasesReturnsSuccess(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := harness.NewFakeK8s(harness.FakeK8sEnv{})
	cfg := newUninstallBaseCfg()
	deps := uninstall.Deps{
		HelmWrapper: fakeHelm,
		K8sWrapper:  fakeK8s,
	}

	err := orchestrator.RunUninstall(cfg, deps)
	require.NoError(t, err)

	_, cpExists := fakeHelm.GetRelease(step.CPReleaseName)
	assert.False(t, cpExists, "CP release should not exist")

	_, clusterExists := fakeHelm.GetRelease(step.ClusterReleaseName)
	assert.False(t, clusterExists, "cluster release should not exist")

	assertNamespaceGone(t, fakeK8s, step.CPNamespace, "CP namespace should not exist")
	assertNamespaceGone(t, fakeK8s, step.ClusterNamespace, "cluster namespace should not exist")
	assertNamespaceGone(t, fakeK8s, step.ReservationNamespace, "runai-reservation namespace should not exist")
	assertNamespaceGone(t, fakeK8s, step.ScaleAdjustNamespace, "runai-scale-adjust namespace should not exist")

	assert.False(t, hasRunaiCRDs(t, fakeK8s), "no runai CRDs should exist")
}

func TestUninstall_HelmTimeoutIsNonFatal(t *testing.T) {
	// Simulate the OpenShift case: helm uninstall wait times out (resources not
	// deleted within the step budget), but the flow continues to ForceDeleteNamespace
	// which handles cleanup. The overall uninstall should succeed.
	timeoutErr := fmt.Errorf("helm uninstall timed out: %w", context.DeadlineExceeded)
	fakeHelm := harness.NewFakeHelmWithReleases(
		harness.ReleaseEntry{Name: step.CPReleaseName, Version: "1.0.0"},
		harness.ReleaseEntry{Name: step.ClusterReleaseName, Version: "1.0.0"},
	)
	fakeHelm.UninstallErrors = map[string]error{
		step.CPReleaseName:      timeoutErr,
		step.ClusterReleaseName: timeoutErr,
	}
	fakeK8s := harness.NewFakeK8s(harness.FakeK8sEnv{
		Namespaces: []string{
			step.CPNamespace,
			step.ClusterNamespace,
			step.ReservationNamespace,
			step.ScaleAdjustNamespace,
		},
	})
	cfg := newUninstallBaseCfg()
	deps := uninstall.Deps{
		HelmWrapper: fakeHelm,
		K8sWrapper:  fakeK8s,
	}

	err := orchestrator.RunUninstall(cfg, deps)
	require.NoError(t, err, "helm wait timeout should be non-fatal")

	_, cpExists := fakeHelm.GetRelease(step.CPReleaseName)
	assert.False(t, cpExists, "CP release should have been uninstalled")

	_, clusterExists := fakeHelm.GetRelease(step.ClusterReleaseName)
	assert.False(t, clusterExists, "cluster release should have been uninstalled")

	assertNamespaceGone(t, fakeK8s, step.CPNamespace, "CP namespace should have been deleted by ForceDeleteNamespace")
	assertNamespaceGone(t, fakeK8s, step.ClusterNamespace, "cluster namespace should have been deleted by ForceDeleteNamespace")
	assertNamespaceGone(t, fakeK8s, step.ReservationNamespace, "runai-reservation namespace should have been deleted")
	assertNamespaceGone(t, fakeK8s, step.ScaleAdjustNamespace, "runai-scale-adjust namespace should have been deleted")
}

func TestUninstall_HelmGenuineFailureIsFatal(t *testing.T) {
	fakeHelm := harness.NewFakeHelmWithReleases(
		harness.ReleaseEntry{Name: step.CPReleaseName, Version: "1.0.0"},
		harness.ReleaseEntry{Name: step.ClusterReleaseName, Version: "1.0.0"},
	)
	fakeHelm.UninstallErrors = map[string]error{
		step.ClusterReleaseName: fmt.Errorf("failed to delete release: permission denied"),
	}
	fakeK8s := harness.NewFakeK8s(harness.FakeK8sEnv{
		Namespaces: []string{step.CPNamespace, step.ClusterNamespace},
	})
	cfg := newUninstallBaseCfg()
	deps := uninstall.Deps{
		HelmWrapper: fakeHelm,
		K8sWrapper:  fakeK8s,
	}

	err := orchestrator.RunUninstall(cfg, deps)
	require.Error(t, err, "genuine helm failure should fail the uninstall")
}

func TestUninstall_RemovesBothCPAndCluster(t *testing.T) {
	fakeHelm := harness.NewFakeHelmWithReleases(
		harness.ReleaseEntry{Name: step.CPReleaseName, Version: "1.0.0"},
		harness.ReleaseEntry{Name: step.ClusterReleaseName, Version: "1.0.0"},
	)
	fakeK8s := harness.NewFakeK8s(harness.FakeK8sEnv{
		Namespaces: []string{
			step.CPNamespace,
			step.ClusterNamespace,
			step.ReservationNamespace,
			step.ScaleAdjustNamespace,
		},
		CRDs: []string{
			"runaiconfigs.run.ai",
			"trainingworkloads.run.ai",
			"runaijobs.run.ai",
			// ResourceInterface is a CR defined by the NVIDIA GPU operator under
			// optimization.nvidia.com.  Including its CRD here causes the fake to
			// behave like a cluster where the GPU operator is installed, so that
			// ListResourceInterfaces returns the seeded instances rather than nil.
			"resourceinterfaces.optimization.nvidia.com",
		},
		MutatingWebhookConfigurations: map[string]map[string]string{
			"test-mutating-webhook-1": {"app": "runai"},
			"test-mutating-webhook-2": {"app": "runai"},
			// Left behind by an older install: no ownership label.
			"test-mutating-webhook-unlabeled": {},
		},
		ValidatingWebhookConfigurations: map[string]map[string]string{
			"test-validating-webhook-1":         {"app": "runai"},
			"test-validating-webhook-2":         {"app": "runai"},
			"test-validating-webhook-unlabeled": {},
			// Another product's webhook, must survive the uninstall.
			"third-party-webhook": {"app": "cert-manager"},
		},
		WebhookServiceNamespaces: map[string][]string{
			"test-mutating-webhook-unlabeled":   {step.ClusterNamespace},
			"test-validating-webhook-unlabeled": {step.CPNamespace},
			"third-party-webhook":               {"cert-manager"},
		},
		ResourceInterfaces: map[string]map[string]string{
			"test-resource-interface-1": {"run.ai/resource-interface-group": "batch"},
			"test-resource-interface-2": {"run.ai/resource-interface-group": "apps"},
		},
		ClusterRoles: map[string]map[string]string{
			"test-cluster-role-1": {"run.ai/additional-workload-gvks-crud": "true"},
			"test-cluster-role-2": {"run.ai/additional-workload-gvks-read": "true"},
		},
		PriorityClasses: map[string]map[string]string{
			"test-priority-class-high": {"run.ai/priority-class-id": "abc-123"},
			"test-priority-class-low":  {"run.ai/priority-class-id": "def-456"},
		},
	})
	cfg := newUninstallBaseCfg()
	deps := uninstall.Deps{
		HelmWrapper: fakeHelm,
		K8sWrapper:  fakeK8s,
	}

	err := orchestrator.RunUninstall(cfg, deps)
	require.NoError(t, err)

	_, cpExists := fakeHelm.GetRelease(step.CPReleaseName)
	assert.False(t, cpExists, "CP release should have been uninstalled")

	_, clusterExists := fakeHelm.GetRelease(step.ClusterReleaseName)
	assert.False(t, clusterExists, "cluster release should have been uninstalled")

	assertNamespaceGone(t, fakeK8s, step.CPNamespace, "CP namespace should have been deleted")
	assertNamespaceGone(t, fakeK8s, step.ClusterNamespace, "cluster namespace should have been deleted")
	assertNamespaceGone(t, fakeK8s, step.ReservationNamespace, "runai-reservation namespace should have been deleted")
	assertNamespaceGone(t, fakeK8s, step.ScaleAdjustNamespace, "runai-scale-adjust namespace should have been deleted")

	assert.False(t, hasRunaiCRDs(t, fakeK8s), "all runai CRDs should have been deleted")

	runaiNamespaces := []string{step.ClusterNamespace, step.CPNamespace}

	remaining, err := fakeK8s.ListMutatingWebhookConfigurations("app=runai", runaiNamespaces)
	require.NoError(t, err)
	assert.Empty(t, remaining, "all RunAI MutatingWebhookConfigurations should have been deleted")

	remaining, err = fakeK8s.ListValidatingWebhookConfigurations("app=runai", runaiNamespaces)
	require.NoError(t, err)
	assert.Empty(t, remaining, "all RunAI ValidatingWebhookConfigurations should have been deleted")

	assert.False(t, fakeK8s.MutatingWebhookConfigurationExists("test-mutating-webhook-unlabeled"),
		"an unlabeled MutatingWebhookConfiguration pointing at the cluster namespace should have been deleted")
	assert.True(t, fakeK8s.ValidatingWebhookConfigurationExists("third-party-webhook"),
		"a webhook owned by another product should have been left alone")

	riRemaining, err := fakeK8s.ListResourceInterfaces("run.ai/")
	require.NoError(t, err)
	assert.Empty(t, riRemaining, "all RunAI ResourceInterfaces should have been deleted")

	crRemaining, err := fakeK8s.ListClusterRoles("run.ai/")
	require.NoError(t, err)
	assert.Empty(t, crRemaining, "all RunAI ClusterRoles should have been deleted")

	pcRemaining, err := fakeK8s.ListPriorityClasses("run.ai/")
	require.NoError(t, err)
	assert.Empty(t, pcRemaining, "all RunAI PriorityClasses should have been deleted")
}

func TestUninstall_StripsRunAICRFinalizers(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := harness.NewFakeK8s(harness.FakeK8sEnv{
		CRDs: []string{"nodepools.run.ai"},
		RunAICRFinalizersByCRD: map[string][]string{
			"nodepools.run.ai": {"default", "gpu-pool"},
		},
	})
	cfg := newUninstallBaseCfg()
	deps := uninstall.Deps{
		HelmWrapper: fakeHelm,
		K8sWrapper:  fakeK8s,
	}

	err := orchestrator.RunUninstall(cfg, deps)
	require.NoError(t, err)

	assert.Empty(t, fakeK8s.CRsWithFinalizers("nodepools.run.ai"),
		"finalizers should have been stripped from all NodePool CRs")
}

func TestUninstall_RemovesAccessRuleCRDWithFinalizers(t *testing.T) {
	fakeK8s := harness.NewFakeK8s(harness.FakeK8sEnv{
		CRDs: []string{"accessrules.run.ai"},
		RunAICRFinalizersByCRD: map[string][]string{
			"accessrules.run.ai": {"rule"},
		},
	})
	cfg := newUninstallBaseCfg()
	deps := uninstall.Deps{
		HelmWrapper: &harness.FakeHelmInstaller{},
		K8sWrapper:  fakeK8s,
	}

	err := orchestrator.RunUninstall(cfg, deps)
	require.NoError(t, err)

	assert.Empty(t, fakeK8s.CRsWithFinalizers("accessrules.run.ai"))
	assert.False(t, hasRunaiCRDs(t, fakeK8s))
}

func TestUninstall_StripsFinalizersBeforeDeletingCRDs(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := harness.NewFakeK8s(harness.FakeK8sEnv{
		CRDs: []string{"nodepools.run.ai", "runaiconfigs.run.ai"},
		RunAICRFinalizersByCRD: map[string][]string{
			"nodepools.run.ai": {"default"},
		},
	})
	cfg := newUninstallBaseCfg()
	deps := uninstall.Deps{
		HelmWrapper: fakeHelm,
		K8sWrapper:  fakeK8s,
	}

	err := orchestrator.RunUninstall(cfg, deps)
	require.NoError(t, err)

	stripIdx := fakeK8s.FirstCallIndex("StripRunAICRFinalizers")
	firstDeleteCRDIdx := fakeK8s.FirstCallIndex("DeleteCRD:")
	require.NotEqual(t, -1, stripIdx, "StripRunAICRFinalizers should have been called")
	require.NotEqual(t, -1, firstDeleteCRDIdx, "DeleteCRD should have been called")
	assert.Less(t, stripIdx, firstDeleteCRDIdx, "StripRunAICRFinalizers must be called before any DeleteCRD")
}

func TestUninstall_NoCRsWithFinalizers_StripIsNoOp(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := harness.NewFakeK8s(harness.FakeK8sEnv{
		CRDs: []string{"nodepools.run.ai"},
	})
	cfg := newUninstallBaseCfg()
	deps := uninstall.Deps{
		HelmWrapper: fakeHelm,
		K8sWrapper:  fakeK8s,
	}

	err := orchestrator.RunUninstall(cfg, deps)
	require.NoError(t, err)

	assert.NotEqual(t, -1, fakeK8s.FirstCallIndex("StripRunAICRFinalizers"),
		"StripRunAICRFinalizers should be called even when no CRs carry finalizers")
	assert.False(t, hasRunaiCRDs(t, fakeK8s), "CRDs should still be deleted")
}

func TestUninstall_PendingHelmReleaseIsFixedAndUninstalledSuccessfully(t *testing.T) {
	// Simulate a cluster where the CP chart was left in pending-install state by a
	// previously interrupted helm install. Without the fix, UninstallRelease would
	// return an error and the uninstall would fail. With FixPendingRelease called
	// before UninstallRelease the pending state is cleared and uninstall succeeds.
	fakeHelm := harness.NewFakeHelmWithReleases(
		harness.ReleaseEntry{Name: step.CPReleaseName, Version: "1.0.0"},
		harness.ReleaseEntry{Name: step.ClusterReleaseName, Version: "1.0.0"},
	)
	fakeHelm.PendingReleases = map[string]bool{
		step.CPReleaseName: true,
	}
	fakeK8s := harness.NewFakeK8s(harness.FakeK8sEnv{
		Namespaces: []string{
			step.CPNamespace,
			step.ClusterNamespace,
			step.ReservationNamespace,
			step.ScaleAdjustNamespace,
		},
	})
	cfg := newUninstallBaseCfg()
	deps := uninstall.Deps{
		HelmWrapper: fakeHelm,
		K8sWrapper:  fakeK8s,
	}

	err := orchestrator.RunUninstall(cfg, deps)
	require.NoError(t, err, "pending helm release should be fixed before uninstall and not block the flow")

	_, cpExists := fakeHelm.GetRelease(step.CPReleaseName)
	assert.False(t, cpExists, "pending CP release should have been uninstalled")

	_, clusterExists := fakeHelm.GetRelease(step.ClusterReleaseName)
	assert.False(t, clusterExists, "cluster release should have been uninstalled")

	assertNamespaceGone(t, fakeK8s, step.CPNamespace, "CP namespace should be gone")
	assertNamespaceGone(t, fakeK8s, step.ClusterNamespace, "cluster namespace should be gone")
}
