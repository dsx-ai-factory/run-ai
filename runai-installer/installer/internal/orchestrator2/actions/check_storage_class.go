package actions

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

// CheckDefaultStorageClass is a read-only preflight check that warns when the
// cluster has no default StorageClass. Run:ai's control plane and several
// prerequisites provision PVCs without an explicit storageClassName, which stay
// Pending without a default. It never fails the install: a missing default is a
// warning, since PVs may be pre-provisioned or storageClassName set explicitly.
type CheckDefaultStorageClass struct{}

func (a CheckDefaultStorageClass) Kind() string { return "preflight.default_storage_class" }

func (a CheckDefaultStorageClass) Args() map[string]any { return map[string]any{} }

func (a CheckDefaultStorageClass) Describe() string {
	return "verify the cluster has a default StorageClass for dynamic provisioning (warning only)"
}

func (a CheckDefaultStorageClass) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	has, err := deps.ClusterProbe.HasDefaultStorageClass(ctx)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("could not list StorageClasses: %v", err),
			Remediation: "grant the installer permission to list storage.k8s.io/storageclasses so storage provisioning can be validated",
		}, nil
	}
	if !has {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      "no default StorageClass found",
			Remediation: "mark a StorageClass default (storageclass.kubernetes.io/is-default-class=true) or set storageClassName explicitly in the component values",
		}, nil
	}
	return orchestrator2.ActionResult{}, nil
}
